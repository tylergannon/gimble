import { mkdirSync, readdirSync, readFileSync, rmSync, statSync, writeFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { pathToFileURL } from 'node:url';

/**
 * The skgo adapter. It emits everything the Go binary embeds and nothing else:
 * the client bundle, kit's own SPA boot document, and a manifest describing the
 * routes Go must answer with that document. No server bundle is copied, so no
 * JavaScript runs in production.
 *
 * It also carries the remote-function ids and the server-load module paths
 * forward. `skgo generate` writes `skgo.remotes.json` beside this file when it
 * emits the `.remote.ts` modules and the `+*.server.ts` stubs; the adapter
 * checks that kit really compiled those modules and copies the lists into the
 * build, so the Go binary can refuse to serve a frontend that was built from a
 * different set of Go functions than it answers.
 *
 * @param {{ out?: string, precompress?: boolean }} [options]
 * @returns {import('@sveltejs/kit').Adapter}
 */
export default function skgo({ out = 'build', precompress = true } = {}) {
	return {
		name: 'skgo',
		async adapt(builder) {
			rmSync(out, { force: true, recursive: true });

			const generated = readGenerated();
			const { manifest: kit, source } = await readKitManifest(builder);
			const hashes = checkRemoteHashes(kit, generated.remotes);

			const nodes = readNodes(builder, source, kit);
			checkServerLoads(nodes, generated.loads);

			const endpoints = checkEndpoints(builder, generated.endpoints);

			builder.writeClient(`${out}/client`);
			checkRemoteIds(`${out}/client`, generated.remotes, hashes);

			builder.writePrerendered(`${out}/prerendered`);
			const prerendered = readPrerendered(builder);
			await builder.generateFallback(`${out}/index.html`);

			// Kit's own `builder.compress` writes a `.br` and a `.gz` beside every
			// file whose extension it compresses, and Go chooses one per request
			// from Accept-Encoding. Same contract adapter-node ships with its
			// `precompress: true` default.
			if (precompress) {
				await builder.compress(`${out}/client`);
				await builder.compress(`${out}/prerendered`);
			}

			// `builder` has no writeJson in kit 3.0.0-next.25 (it existed on the
			// kit 2 line); write the manifest ourselves.
			write(
				`${out}/skgo.manifest.json`,
				JSON.stringify(
					{
						appDir: builder.config.appDir,
						base: builder.config.paths.base,
						version: builder.config.version.name,
						// One entry per node, positionally: the vite-root-relative
						// path of its `+*.server.ts`, or "" for a node that has
						// none. It is how Go finds the load that answers a slot of
						// a route's branch, and it is the same key kit itself
						// records in the node module it builds.
						nodes,
						routes: kit._.routes.map((/** @type {any} */ route) => ({
							id: route.id,
							pattern: route.pattern.source,
							params: route.params,
							// The methods kit's build says the compiled `+server.ts`
							// exports. Without it a route that is an endpoint and
							// nothing else is indistinguishable from a page route, and
							// Go answered it with the boot document at HTTP 200.
							endpoint: endpoints.has(route.id)
								? { methods: endpoints.get(route.id) }
								: null,
							page: route.page
								? {
										// `[...layouts, leaf]` is the branch the
										// client's `x-sveltekit-invalidated` string
										// is positional over. A hole in `layouts` is
										// a slot no layout fills; it travels as -1
										// because JSON has no holes.
										layouts: Array.from(
											{ length: route.page.layouts.length },
											(_, i) => route.page.layouts[i] ?? -1
										),
										leaf: route.page.leaf
									}
								: null
						})),
						remotes: generated.remotes,
						prerendered,
						precompressed: precompress
					},
					null,
					'\t'
				)
			);

			builder.log.minor(`skgo: wrote ${out}/`);
		}
	};
}

/**
 * Reads what `skgo generate` emitted.
 *
 * A missing or shapeless file is a failure, not an empty list: an adapter that
 * quietly compares nothing to nothing reports success for an app whose two
 * halves were never checked against each other.
 *
 * @returns {{ remotes: string[], loads: string[], endpoints: Record<string, string[]> }}
 */
function readGenerated() {
	let raw;
	try {
		raw = readFileSync('skgo.remotes.json', 'utf-8');
	} catch {
		throw new Error(
			'skgo: skgo.remotes.json is missing. Run `go generate ./...` before building the frontend.'
		);
	}
	const parsed = JSON.parse(raw);
	if (!Array.isArray(parsed.remotes)) {
		throw new Error(
			'skgo: skgo.remotes.json has no `remotes` array. Run `go generate ./...` before building the frontend.'
		);
	}
	for (const id of parsed.remotes) {
		if (typeof id !== 'string' || !/^[^/]+\/[^/]+$/.test(id)) {
			throw new Error(
				`skgo: skgo.remotes.json lists ${JSON.stringify(id)}, which is not a <hash>/<name> id.`
			);
		}
	}
	if (!Array.isArray(parsed.loads)) {
		throw new Error(
			'skgo: skgo.remotes.json has no `loads` array. Run `go generate ./...` before building the frontend.'
		);
	}
	for (const module of parsed.loads) {
		if (typeof module !== 'string' || !/\/\+(page|layout)\.server\.ts$/.test(module)) {
			throw new Error(
				`skgo: skgo.remotes.json lists ${JSON.stringify(module)}, which is not a +page.server.ts or +layout.server.ts path.`
			);
		}
	}
	if (typeof parsed.endpoints !== 'object' || parsed.endpoints === null || Array.isArray(parsed.endpoints)) {
		throw new Error(
			'skgo: skgo.remotes.json has no `endpoints` object. Run `go generate ./...` before building the frontend.'
		);
	}
	for (const [id, methods] of Object.entries(parsed.endpoints)) {
		if (!id.startsWith('/') || !Array.isArray(methods) || methods.length === 0) {
			throw new Error(
				`skgo: skgo.remotes.json maps ${JSON.stringify(id)} to ${JSON.stringify(methods)}, which is not a route id and its methods.`
			);
		}
	}
	return { remotes: parsed.remotes, loads: parsed.loads, endpoints: parsed.endpoints };
}

/**
 * The same check for server routes, against the one place kit reports what it
 * compiled: `builder.routes[].api.methods`, which kit derives by importing each
 * built `+server.js` and reading its exports
 * (packages/kit/src/core/postbuild/analyse.js, `analyse_endpoint`). A `fallback`
 * export travels there as `'*'`, and `skgo generate` writes the same spelling,
 * so the two lists are compared literally.
 *
 * This is the check that makes a hand-written `+server.ts` fail the build rather
 * than 404 in the browser: kit would compile it, Go would never have been told
 * about it, and the route would answer nothing.
 *
 * @param {import('@sveltejs/kit').Builder} builder
 * @param {Record<string, string[]>} declared
 * @returns {Map<string, string[]>} the methods kit compiled, per route id
 */
function checkEndpoints(builder, declared) {
	/** @type {Map<string, string[]>} */
	const built = new Map();
	for (const route of builder.routes) {
		if (route.api.methods.length > 0) built.set(route.id, [...route.api.methods].sort());
	}

	/** @type {string[]} */
	const problems = [];
	for (const [id, methods] of Object.entries(declared)) {
		const compiled = built.get(id);
		if (!compiled) {
			problems.push(`  generated but not compiled: ${methods.join(', ')} ${id}`);
			continue;
		}
		const want = [...methods].sort().join(', ');
		const got = compiled.join(', ');
		if (want !== got) {
			problems.push(`  ${id}: Go answers ${want}, the built +server.ts exports ${got}`);
		}
	}
	for (const [id, methods] of built) {
		if (!(id in declared)) {
			problems.push(`  compiled but not generated: ${methods.join(', ')} ${id}`);
		}
	}

	if (problems.length) {
		throw new Error(
			'skgo: skgo.remotes.json does not describe the server routes kit just compiled.\n' +
				problems.join('\n') +
				'\n  Every server route is written in Go. Run `go generate ./...`.'
		);
	}
	return built;
}

/**
 * The pathnames the build wrote a file for, exactly as kit records them:
 * percent-decoded and carrying the configured base. Go serves the prerendered
 * tree off this list and issues the trailing-slash 308 off it, which is what
 * adapter-node does with the same array.
 *
 * A prerendered *redirect* is refused rather than dropped. Kit records one when
 * a page it rendered redirected somewhere, and skgo has nothing that would
 * replay it — an app that produced one would silently lose it.
 *
 * @param {import('@sveltejs/kit').Builder} builder
 * @returns {string[]}
 */
function readPrerendered(builder) {
	if (builder.prerendered.redirects.size > 0) {
		throw new Error(
			'skgo: the build prerendered a redirect, which skgo does not serve:\n' +
				[...builder.prerendered.redirects]
					.map(([from, { status, location }]) => `  ${from} -> ${status} ${location}`)
					.join('\n')
		);
	}
	return [...builder.prerendered.paths].sort();
}

/**
 * Kit's own server manifest, as an object *and* as text. `generateManifest`
 * returns the source of a module whose only imports sit inside lazy thunks, so
 * importing it resolves nothing and runs no application code — but the source
 * is worth keeping, because the thunks carry information the imported object
 * has already hidden inside closures. See readNodes.
 *
 * @param {import('@sveltejs/kit').Builder} builder
 * @returns {Promise<{ manifest: any, source: string }>}
 */
async function readKitManifest(builder) {
	const dir = builder.getBuildDirectory('skgo');
	const file = join(dir, 'kit-manifest.js');
	const source = builder.generateManifest({ relativePath: '.' });
	mkdirSync(dir, { recursive: true });
	writeFileSync(file, `export const manifest = ${source};\n`);
	try {
		// The cache buster matters: `vp build` can run twice in one process.
		return { manifest: (await import(`${pathToFileURL(file).href}?t=${Date.now()}`)).manifest, source };
	} finally {
		rmSync(file, { force: true });
	}
}

/**
 * The vite-root-relative path of every node's `+*.server.ts`, in the positions
 * the manifest's own route branches point at. Kit records the path as
 * `server_id` in the node modules it builds, and that is the only place the
 * mapping from a branch slot back to an authored file survives the build.
 *
 * The positions are the trap. Kit *renumbers* nodes when it writes a manifest:
 * `generate_manifest` collects the nodes the surviving routes use and hands each
 * one a fresh consecutive index, so a route's `leaf` is a position in the
 * manifest's node array and not the number the node's own module declares. The
 * two agree only while every node is used — and prerendering a single page drops
 * that page's node and shifts every later one down by one.
 *
 * It shows up as a page whose Go load never runs while the layout above it works
 * perfectly, which reads like a bug in the load rather than in the manifest.
 *
 * The renumbering is recoverable because kit writes the original index into each
 * node's import path (`__memo(() => import('./nodes/6.js'))`), in the new order.
 *
 * @param {import('@sveltejs/kit').Builder} builder
 * @param {string} source kit's generated manifest, as text
 * @param {any} kit the same manifest, imported
 * @returns {string[]}
 */
function readNodes(builder, source, kit) {
	const dir = join(builder.getServerDirectory(), 'nodes');
	/** @type {Map<number, string>} */
	const serverIds = new Map();
	for (const name of readdirSync(dir)) {
		if (!name.endsWith('.js')) continue;
		const module = readFileSync(join(dir, name), 'utf-8');
		const index = Number(module.match(/^export const index = (\d+);$/m)?.[1]);
		if (!Number.isInteger(index)) {
			throw new Error(`skgo: ${join(dir, name)} does not declare a node index`);
		}
		serverIds.set(index, module.match(/^export const server_id = "([^"]*)";$/m)?.[1] ?? '');
	}

	const order = [...source.matchAll(/import\('[^']*\/nodes\/(\d+)\.js'\)/g)].map((m) =>
		Number(m[1])
	);
	if (order.length !== kit._.nodes.length) {
		throw new Error(
			`skgo: kit's manifest holds ${kit._.nodes.length} node(s) but ${order.length} node import(s) could be read out of it. ` +
				'The manifest no longer says which node module each branch slot points at, and skgo would silently run the wrong load.'
		);
	}

	return order.map((index) => {
		if (!serverIds.has(index)) {
			throw new Error(`skgo: kit's manifest imports nodes/${index}.js, which the build did not write`);
		}
		return serverIds.get(index) ?? '';
	});
}

/**
 * Kit's own manifest lists the remote *modules* it compiled, keyed by the hash
 * it derived from each module's path. If that set is not the one `skgo
 * generate` wrote, the two halves of the app were generated from different
 * sources and the build must not succeed.
 *
 * Modules are as far as this check can reach. `generate_manifest` in kit's
 * packages/kit/src/core/generate_manifest/index.js emits
 * `'<hash>': __memo(() => import('./chunks/remote-<hash>.js'))` and nothing
 * else; kit resolves the function half of an id at *runtime*, by looking the
 * name up on the imported module namespace
 * (packages/kit/src/runtime/server/remote-functions.js). There is no
 * per-function list in the manifest to compare against. checkRemoteIds below
 * reads the other build artifact that does carry the names.
 *
 * @param {any} kit kit's own server manifest
 * @param {string[]} remotes
 * @returns {Set<string>} the module hashes kit compiled
 */
function checkRemoteHashes(kit, remotes) {
	if (!kit._ || typeof kit._.remotes !== 'object' || kit._.remotes === null) {
		throw new Error(
			"skgo: kit's generated manifest has no `remotes` map. The adapter cannot tell which remote modules were compiled, so it cannot check them; this build of SvelteKit is not one skgo has been taught to read."
		);
	}
	const built = new Set(Object.keys(kit._.remotes));
	const declared = new Set(remotes.map((id) => id.split('/')[0]));

	const missing = [...declared].filter((hash) => !built.has(hash));
	const extra = [...built].filter((hash) => !declared.has(hash));
	if (missing.length || extra.length) {
		throw new Error(
			'skgo: skgo.remotes.json does not describe the remote modules kit just compiled.\n' +
				(missing.length ? `  generated but not compiled: ${missing.join(', ')}\n` : '') +
				(extra.length ? `  compiled but not generated: ${extra.join(', ')}\n` : '') +
				'  A generated .remote.ts only reaches the build once app code imports it. Run `go generate ./...`.'
		);
	}
	return built;
}

/**
 * Checks the remote functions the shipped bundle can actually call.
 *
 * The module-level check above is blind to an extra export added by hand to an
 * already-generated `.remote.ts`: the module's hash does not change, so both
 * sides still agree, the bundle calls `<hash>/<newName>`, and the Go binary —
 * which was never told about it — answers 404 in the browser.
 *
 * Kit's client carries the full id as a literal. Its vite plugin rewrites every
 * remote module's client half into one `export const <name> =
 * __remote.<type>('<hash>/<name>')` per export
 * (packages/kit/src/exports/vite/index.js), and the client runtime pastes that
 * id straight into `${base}/${app_dir}/remote/${id}`
 * (packages/kit/src/runtime/client/remote-functions/query/index.js). So the ids
 * survive bundling and minification as string literals, and the set of them in
 * the client we are about to ship is exactly the set of remote calls the
 * browser can make.
 *
 * Every one of those must be a function `skgo generate` declared. The reverse
 * is not required: a declared function no component imports is treeshaken out
 * of the client, and Go serving a function nobody calls yet breaks nothing.
 *
 * @param {string} clientDir the client bundle this build will ship
 * @param {string[]} remotes
 * @param {Set<string>} hashes the module hashes kit compiled
 */
function checkRemoteIds(clientDir, remotes, hashes) {
	const declared = new Set(remotes);
	const called = new Map();

	for (const file of walk(clientDir)) {
		if (!file.endsWith('.js')) continue;
		const source = readFileSync(file, 'utf-8');
		for (const [, hash, name] of source.matchAll(
			/['"`]([A-Za-z0-9]{1,16})\/([A-Za-z_$][A-Za-z0-9_$]*)['"`]/g
		)) {
			if (!hashes.has(hash)) continue;
			called.set(`${hash}/${name}`, file);
		}
	}

	if (called.size === 0 && hashes.size > 0) {
		throw new Error(
			`skgo: kit compiled ${hashes.size} remote module(s) but no remote-function id appears in ${clientDir}. Either nothing imports them, or the ids no longer survive bundling as literals and this check has stopped meaning anything.`
		);
	}

	const undeclared = [...called].filter(([id]) => !declared.has(id));
	if (undeclared.length) {
		throw new Error(
			'skgo: the built frontend calls remote functions that `skgo generate` did not write.\n' +
				undeclared.map(([id, file]) => `  ${id} (in ${file})`).join('\n') +
				'\n  Go answers only the generated ids, so every one of these would 404 in the browser.\n' +
				'  A remote function written by hand in a .remote.ts is not a supported mode: write it in the matching .remote.go and run `go generate ./...`.'
		);
	}
}

/**
 * @param {string} dir
 * @returns {Generator<string>}
 */
function* walk(dir) {
	for (const entry of readdirSync(dir)) {
		const path = join(dir, entry);
		if (statSync(path).isDirectory()) yield* walk(path);
		else yield path;
	}
}

/**
 * The same check for server loads. Kit decides whether its client asks for
 * `__data.json` at all by looking for a `load` export in the *built*
 * `+*.server.js`, so a stub that did not reach the build is a page whose Go load
 * is never called, and a compiled stub with no Go load behind it is a page whose
 * data nobody answers.
 *
 * @param {string[]} nodes
 * @param {string[]} loads
 */
function checkServerLoads(nodes, loads) {
	const built = new Set(nodes.filter(Boolean));
	const declared = new Set(loads);

	const missing = [...declared].filter((id) => !built.has(id));
	const extra = [...built].filter((id) => !declared.has(id));
	if (missing.length === 0 && extra.length === 0) return;

	throw new Error(
		'skgo: skgo.remotes.json does not describe the server loads kit just compiled.\n' +
			(missing.length ? `  generated but not compiled: ${missing.join(', ')}\n` : '') +
			(extra.length ? `  compiled but not generated: ${extra.join(', ')}\n` : '') +
			'  Every server load is written in Go. Run `go generate ./...`.'
	);
}

/**
 * @param {string} file
 * @param {string} contents
 */
function write(file, contents) {
	mkdirSync(dirname(file), { recursive: true });
	writeFileSync(file, contents);
}
