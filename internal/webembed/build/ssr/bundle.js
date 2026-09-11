// The globals kit's runtime and Svelte's renderer reach that a bare
// ECMAScript engine does not have and Go does not bind.
//
// This file is the SSR bundle's banner rather than a module of it. The fold
// that makes the bundle evaluates kit's shared chunk before the entry, so a
// polyfill imported as a module would arrive after the code that needs it —
// which the single-pass build only got away with by module order.
//
// URL, URLSearchParams, TextEncoder, TextDecoder, btoa and atob are not here:
// Go binds those natively when it creates a runtime, before this text runs.
// What is left is what a Go-backed constructor could not stand in for: the
// classes kit compares against with `instanceof` (Headers, Blob, File) and the
// two a render's own `event.fetch` builds and hands back (Request, Response),
// which are string and Map bookkeeping over a value that never leaves the
// engine except through `__skgo_fetch`.
// Svelte's no-AsyncLocalStorage fallback is gated on exactly this check
// (svelte/src/internal/server/render-context.js), and kit reads the same flag
// (kit/src/constants.js) to stop nulling its synchronous request store. It is
// Svelte's own supported path; the cost is one render per runtime at a time,
// which the pool in Go pays.
globalThis.process = { versions: { webcontainer: "skgo" } };
// Headers is reached by kit's own Redirect constructor
// (exports/internal/shared.js), which builds one to reject a location that
// could not survive an HTTP header — and does it inside a try/catch, so an
// absent Headers would be reported as an invalid location rather than as a
// missing global. Only the construction and the validation it performs are
// needed; nothing in the engine sends a request.
if (typeof globalThis.Headers === "undefined") {
	// RFC 9110 field-name and field-value rules, which is all the real
	// constructor checks that matters here.
	const NAME = /^[A-Za-z0-9!#$%&'*+.^_|~-]+$/;
	const BAD_VALUE = /[\u0000\r\n]/;
	globalThis.Headers = class Headers {
		constructor(init) {
			this._ = /* @__PURE__ */ new Map();
			if (init instanceof Headers) for (const [k, v] of init._) this._.set(k, v);
			else if (Array.isArray(init)) for (const [k, v] of init) this.set(k, v);
			else if (init && typeof init === "object") for (const k of Object.keys(init)) this.set(k, init[k]);
		}
		set(name, value) {
			const n = String(name);
			const v = String(value).trim();
			if (!NAME.test(n)) throw new TypeError("Invalid header name: " + n);
			if (BAD_VALUE.test(v)) throw new TypeError("Invalid header value");
			this._.set(n.toLowerCase(), v);
		}
		append(name, value) {
			const existing = this.get(name);
			this.set(name, existing === null ? value : existing + ", " + value);
		}
		get(name) {
			const k = String(name).toLowerCase();
			return this._.has(k) ? this._.get(k) : null;
		}
		has(name) {
			return this._.has(String(name).toLowerCase());
		}
		delete(name) {
			this._.delete(String(name).toLowerCase());
		}
		forEach(fn, thisArg) {
			for (const [k, v] of this._) fn.call(thisArg, v, k, this);
		}
		keys() {
			return this._.keys();
		}
		values() {
			return this._.values();
		}
		entries() {
			return this._.entries();
		}
		[Symbol.iterator]() {
			return this._.entries();
		}
	};
}
// Request and Response back a render-time `event.fetch`. Kit's own
// `normalize_fetch_input` (runtime/server/fetch.js) turns whatever a
// component passed into a real Request before deciding what to do with it, and
// the answer a render's fetch gets back has to be a real Response — `await
// (await event.fetch(...)).json()` is what a page actually writes. Nothing
// here sends bytes anywhere: building one is string and Map bookkeeping, and
// the one call that leaves the engine is `__skgo_fetch` itself.
if (typeof globalThis.Request === "undefined") globalThis.Request = class Request {
	constructor(input, init = {}) {
		if (input instanceof Request) {
			this.url = input.url;
			this.method = (init.method ?? input.method ?? "GET").toUpperCase();
			this.headers = init.headers ? new Headers(init.headers) : new Headers(input.headers);
			this._body = init.body !== void 0 ? init.body : input._body;
			this.credentials = init.credentials ?? input.credentials ?? "same-origin";
			this.mode = init.mode ?? input.mode ?? "cors";
		} else {
			this.url = String(input);
			this.method = (init.method ?? "GET").toUpperCase();
			this.headers = new Headers(init.headers);
			this._body = init.body;
			this.credentials = init.credentials ?? "same-origin";
			this.mode = init.mode ?? "cors";
		}
	}
	async text() {
		return this._body ?? "";
	}
	async json() {
		return JSON.parse(this._body ?? "null");
	}
};
if (typeof globalThis.Response === "undefined") globalThis.Response = class Response {
	constructor(body, init = {}) {
		this._body = body ?? "";
		this.status = init.status ?? 200;
		this.statusText = init.statusText ?? "";
		this.headers = init.headers instanceof Headers ? init.headers : new Headers(init.headers);
		this.ok = this.status >= 200 && this.status < 300;
	}
	async text() {
		return this._body;
	}
	async json() {
		return JSON.parse(this._body);
	}
	async arrayBuffer() {
		return new TextEncoder().encode(this._body).buffer;
	}
	clone() {
		return new Response(this._body, {
			status: this.status,
			statusText: this.statusText,
			headers: this.headers
		});
	}
};
// Blob and File are named by kit's form-field proxy, which asks whether a
// field's value is a File before it decides how to describe it to the markup.
// Nothing here ever holds one — an uploaded file's bytes are Go's, and they
// never enter the engine — so these exist to be compared against.
if (typeof globalThis.Blob === "undefined") globalThis.Blob = class Blob {
	constructor(parts = [], options = {}) {
		this._parts = parts;
		this.type = options.type ?? "";
		this.size = 0;
	}
};
if (typeof globalThis.File === "undefined") globalThis.File = class File extends globalThis.Blob {
	constructor(parts = [], name = "", options = {}) {
		super(parts, options);
		this.name = String(name);
		this.lastModified = options.lastModified ?? 0;
	}
};
(function() {
	//#region build/.goja/bundle.js
	var __create = Object.create;
	var __defProp = Object.defineProperty;
	var __getOwnPropDesc = Object.getOwnPropertyDescriptor;
	var __getOwnPropNames = Object.getOwnPropertyNames;
	var __getProtoOf = Object.getPrototypeOf;
	var __hasOwnProp = Object.prototype.hasOwnProperty;
	var __commonJSMin = (cb, mod) => () => (mod || (cb((mod = { exports: {} }).exports, mod), cb = null), mod.exports);
	var __exportAll = (all, no_symbols) => {
		let target = {};
		for (var name in all) __defProp(target, name, {
			get: all[name],
			enumerable: true
		});
		if (!no_symbols) __defProp(target, Symbol.toStringTag, { value: "Module" });
		return target;
	};
	var __copyProps = (to, from, except, desc) => {
		if (from && typeof from === "object" || typeof from === "function") for (var keys = __getOwnPropNames(from), i = 0, n = keys.length, key; i < n; i++) {
			key = keys[i];
			if (!__hasOwnProp.call(to, key) && key !== except) __defProp(to, key, {
				get: ((k) => from[k]).bind(null, key),
				enumerable: !(desc = __getOwnPropDesc(from, key)) || desc.enumerable
			});
		}
		return to;
	};
	var __toESM = (mod, isNodeMode, target) => (target = mod != null ? __create(__getProtoOf(mod)) : {}, __copyProps(isNodeMode || !mod || !mod.__esModule || !__hasOwnProp.call(mod, "default") ? __defProp(target, "default", {
		value: mod,
		enumerable: true
	}) : target, mod));
	var UNINITIALIZED = Symbol("uninitialized");
	var ATTR_REGEX = /[&"<]/g;
	var CONTENT_REGEX = /[&<]/g;
	/**
	* @template V
	* @param {V} value
	* @param {boolean} [is_attr]
	*/
	function escape_html(value, is_attr) {
		const str = String(value ?? "");
		const pattern = is_attr ? ATTR_REGEX : CONTENT_REGEX;
		pattern.lastIndex = 0;
		let escaped = "";
		let last = 0;
		while (pattern.test(str)) {
			const i = pattern.lastIndex - 1;
			const ch = str[i];
			escaped += str.substring(last, i) + (ch === "&" ? "&amp;" : ch === "\"" ? "&quot;" : "&lt;");
			last = i + 1;
		}
		return escaped + str.substring(last);
	}
	function r(e) {
		var t, f, n = "";
		if ("string" == typeof e || "number" == typeof e) n += e;
		else if ("object" == typeof e) if (Array.isArray(e)) {
			var o = e.length;
			for (t = 0; t < o; t++) e[t] && (f = r(e[t])) && (n && (n += " "), n += f);
		} else for (f in e) e[f] && (n && (n += " "), n += f);
		return n;
	}
	function clsx$1() {
		for (var e, t, f = 0, n = "", o = arguments.length; f < o; f++) (e = arguments[f]) && (t = r(e)) && (n && (n += " "), n += t);
		return n;
	}
	var is_array = Array.isArray;
	Array.prototype.indexOf;
	Array.prototype.includes;
	Array.from;
	Array.prototype;
	var has_own_property = Object.prototype.hasOwnProperty;
	var noop$1 = () => {};
	/**
	* TODO replace with Promise.withResolvers once supported widely enough
	* @template [T=void]
	*/
	function deferred() {
		/** @type {(value: T) => void} */
		var resolve;
		/** @type {(reason: any) => void} */
		var reject;
		return {
			promise: new Promise((res, rej) => {
				resolve = res;
				reject = rej;
			}),
			resolve,
			reject
		};
	}
	/**
	* `<div translate={false}>` should be rendered as `<div translate="no">` and _not_
	* `<div translate="false">`, which is equivalent to `<div translate="yes">`. There
	* may be other odd cases that need to be added to this list in future
	* @type {Record<string, Map<any, string>>}
	*/
	var replacements = { translate: /* @__PURE__ */ new Map([[true, "yes"], [false, "no"]]) };
	/**
	* @template V
	* @param {string} name
	* @param {V} value
	* @param {boolean} [is_boolean]
	* @returns {string}
	*/
	function attr(name, value, is_boolean = false) {
		if (name === "hidden" && value !== "until-found") is_boolean = true;
		if (value == null || is_boolean && !value && value !== "") return "";
		const normalized = has_own_property.call(replacements, name) && replacements[name].get(value) || value;
		return ` ${name}${is_boolean ? `=""` : `="${escape_html(normalized, true)}"`}`;
	}
	/**
	* Small wrapper around clsx to preserve Svelte's (weird) handling of falsy values.
	* TODO Svelte 6 revisit this, and likely turn all falsy values into the empty string (what clsx also does)
	* @param  {any} value
	*/
	function clsx(value) {
		if (typeof value === "object") return clsx$1(value);
		else return value ?? "";
	}
	var whitespace = [..." 	\n\r\f\xA0\v﻿"];
	/**
	* @param {any} value
	* @param {string | null} [hash]
	* @param {Record<string, boolean>} [directives]
	* @returns {string | null}
	*/
	function to_class(value, hash, directives) {
		var classname = value == null ? "" : "" + value;
		if (hash) classname = classname ? classname + " " + hash : hash;
		if (directives) {
			for (var key of Object.keys(directives)) if (directives[key]) classname = classname ? classname + " " + key : key;
			else if (classname.length) {
				var len = key.length;
				var a = 0;
				while ((a = classname.indexOf(key, a)) >= 0) {
					var b = a + len;
					if ((a === 0 || whitespace.includes(classname[a - 1])) && (b === classname.length || whitespace.includes(classname[b]))) classname = (a === 0 ? "" : classname.substring(0, a)) + classname.substring(b + 1);
					else a = b;
				}
			}
		}
		return classname === "" ? null : classname;
	}
	/**
	*
	* @param {Record<string,any>} styles
	* @param {boolean} important
	*/
	function append_styles(styles, important = false) {
		var separator = important ? " !important;" : ";";
		var css = "";
		for (var key of Object.keys(styles)) {
			var value = styles[key];
			if (value != null && value !== "") css += " " + key + ": " + value + separator;
		}
		return css;
	}
	/**
	* @param {string} name
	* @returns {string}
	*/
	function to_css_name(name) {
		if (name[0] !== "-" || name[1] !== "-") return name.toLowerCase();
		return name;
	}
	/**
	* @param {any} value
	* @param {Record<string, any> | [Record<string, any>, Record<string, any>]} [styles]
	* @returns {string | null}
	*/
	function to_style(value, styles) {
		if (styles) {
			var new_style = "";
			/** @type {Record<string,any> | undefined} */
			var normal_styles;
			/** @type {Record<string,any> | undefined} */
			var important_styles;
			if (Array.isArray(styles)) {
				normal_styles = styles[0];
				important_styles = styles[1];
			} else normal_styles = styles;
			if (value) {
				value = String(value).replaceAll(/\/\*.*?\*\//g, "").trim();
				/** @type {boolean | '"' | "'"} */
				var in_str = false;
				var in_apo = 0;
				var in_comment = false;
				var reserved_names = [];
				if (normal_styles) reserved_names.push(...Object.keys(normal_styles).map(to_css_name));
				if (important_styles) reserved_names.push(...Object.keys(important_styles).map(to_css_name));
				var start_index = 0;
				var name_index = -1;
				const len = value.length;
				for (var i = 0; i < len; i++) {
					var c = value[i];
					if (in_comment) {
						if (c === "/" && value[i - 1] === "*") in_comment = false;
					} else if (in_str) {
						if (in_str === c) in_str = false;
					} else if (c === "/" && value[i + 1] === "*") in_comment = true;
					else if (c === "\"" || c === "'") in_str = c;
					else if (c === "(") in_apo++;
					else if (c === ")") in_apo--;
					if (!in_comment && in_str === false && in_apo === 0) {
						if (c === ":" && name_index === -1) name_index = i;
						else if (c === ";" || i === len - 1) {
							if (name_index !== -1) {
								var name = to_css_name(value.substring(start_index, name_index).trim());
								if (!reserved_names.includes(name)) {
									if (c !== ";") i++;
									var property = value.substring(start_index, i).trim();
									new_style += " " + property + ";";
								}
							}
							start_index = i + 1;
							name_index = -1;
						}
					}
				}
			}
			if (normal_styles) new_style += append_styles(normal_styles);
			if (important_styles) new_style += append_styles(important_styles, true);
			new_style = new_style.trim();
			return new_style === "" ? null : new_style;
		}
		return value == null ? null : String(value);
	}
	var CLEAN = 1024;
	var DIRTY = 2048;
	var MAYBE_DIRTY = 4096;
	/** allow users to ignore aborted signal errors if `reason.name === 'StaleReactionError` */
	var STALE_REACTION = new class StaleReactionError extends Error {
		name = "StaleReactionError";
		message = "The reaction that called `getAbortSignal()` was re-run or destroyed";
	}();
	globalThis.document?.contentType;
	/**
	* `%name%(...)` can only be used during component initialisation
	* @param {string} name
	* @returns {never}
	*/
	function lifecycle_outside_component(name) {
		throw new Error(`https://svelte.dev/e/lifecycle_outside_component`);
	}
	/** True if experimental.async=true */
	var async_mode_flag = false;
	function enable_async_mode_flag() {
		async_mode_flag = true;
	}
	/**
	* @typedef {{ p: Context | null, c: Map<unknown, unknown> | null }} Context
	*/
	/**
	* @param {Context} context
	* @returns {Map<unknown, unknown> | null}
	*/
	function get_parent_context(context) {
		let parent = context.p;
		while (parent !== null && parent.c === null) parent = parent.p;
		return parent?.c ?? null;
	}
	/**
	* @param {Context | null} context
	* @param {string} name
	* @returns {Map<unknown, unknown>}
	*/
	function get_or_init_context_map(context, name) {
		if (context === null) lifecycle_outside_component(name);
		return context.c ??= new Map(get_parent_context(context) || void 0);
	}
	~(DIRTY | MAYBE_DIRTY | CLEAN);
	function _OverloadYield(e, d) {
		this.v = e, this.k = d;
	}
	function _awaitAsyncGenerator(e) {
		return new _OverloadYield(e, 0);
	}
	function _wrapAsyncGenerator(e) {
		return function() {
			return new AsyncGenerator(e.apply(this, arguments));
		};
	}
	function AsyncGenerator(e) {
		var r, t;
		function resume(r, t) {
			try {
				var n = e[r](t), o = n.value, u = o instanceof _OverloadYield;
				Promise.resolve(u ? o.v : o).then(function(t) {
					if (u) {
						var i = "return" === r ? "return" : "next";
						if (!o.k || t.done) return resume(i, t);
						t = e[i](t).value;
					}
					settle(n.done ? "return" : "normal", t);
				}, function(e) {
					resume("throw", e);
				});
			} catch (e) {
				settle("throw", e);
			}
		}
		function settle(e, n) {
			switch (e) {
				case "return":
					r.resolve({
						value: n,
						done: !0
					});
					break;
				case "throw":
					r.reject(n);
					break;
				default: r.resolve({
					value: n,
					done: !1
				});
			}
			(r = r.next) ? resume(r.key, r.arg) : t = null;
		}
		this._invoke = function(e, n) {
			return new Promise(function(o, u) {
				var i = {
					key: e,
					arg: n,
					resolve: o,
					reject: u,
					next: null
				};
				t ? t = t.next = i : (r = t = i, resume(e, n));
			});
		}, "function" != typeof e["return"] && (this["return"] = void 0);
	}
	AsyncGenerator.prototype["function" == typeof Symbol && Symbol.asyncIterator || "@@asyncIterator"] = function() {
		return this;
	}, AsyncGenerator.prototype.next = function(e) {
		return this._invoke("next", e);
	}, AsyncGenerator.prototype["throw"] = function(e) {
		return this._invoke("throw", e);
	}, AsyncGenerator.prototype["return"] = function(e) {
		return this._invoke("return", e);
	};
	var BLOCK_OPEN = `<!--[-->`;
	var BLOCK_CLOSE = `<!--]-->`;
	/**
	* Attributes that are boolean, i.e. they are present or not present.
	*/
	var DOM_BOOLEAN_ATTRIBUTES = [
		"allowfullscreen",
		"async",
		"autofocus",
		"autoplay",
		"checked",
		"controls",
		"default",
		"disabled",
		"formnovalidate",
		"indeterminate",
		"inert",
		"ismap",
		"loop",
		"multiple",
		"muted",
		"nomodule",
		"novalidate",
		"open",
		"playsinline",
		"readonly",
		"required",
		"reversed",
		"seamless",
		"selected",
		"webkitdirectory",
		"defer",
		"disablepictureinpicture",
		"disableremoteplayback"
	];
	/**
	* Returns `true` if `name` is a boolean attribute
	* @param {string} name
	*/
	function is_boolean_attribute(name) {
		return DOM_BOOLEAN_ATTRIBUTES.includes(name);
	}
	[...DOM_BOOLEAN_ATTRIBUTES];
	/** @import { SSRContext } from '#server' */
	/** @type {SSRContext | null} */
	var ssr_context = null;
	/** @param {SSRContext | null} v */
	function set_ssr_context(v) {
		ssr_context = v;
	}
	/**
	* @template T
	* @param {any} key
	* @returns {T}
	*/
	function getContext(key) {
		return get_or_init_context_map(ssr_context, "getContext").get(key);
	}
	/**
	* @param {Function} [fn]
	*/
	function push(fn) {
		ssr_context = {
			p: ssr_context,
			c: null,
			r: null
		};
	}
	function pop() {
		ssr_context = ssr_context.p;
	}
	/**
	* Wraps an `await` expression in such a way that the component context that was
	* active before the expression evaluated can be reapplied afterwards —
	* `await a + b()` becomes `(await $.save(a))() + b()`, meaning `b()` will have access
	* to the context of its component.
	* @template T
	* @param {Promise<T>} promise
	* @returns {Promise<() => T>}
	*/
	async function save(promise) {
		var previous_context = ssr_context;
		var value = await promise;
		return () => {
			ssr_context = previous_context;
			return value;
		};
	}
	/**
	* The node API `AsyncLocalStorage` is not available, but is required to use async server rendering.
	* @returns {never}
	*/
	function async_local_storage_unavailable() {
		const error = /* @__PURE__ */ new Error(`async_local_storage_unavailable\nThe node API \`AsyncLocalStorage\` is not available, but is required to use async server rendering.\nhttps://svelte.dev/e/async_local_storage_unavailable`);
		error.name = "Svelte error";
		throw error;
	}
	/**
	* Encountered asynchronous work while rendering synchronously.
	* @returns {never}
	*/
	function await_invalid() {
		const error = /* @__PURE__ */ new Error(`await_invalid\nEncountered asynchronous work while rendering synchronously.\nhttps://svelte.dev/e/await_invalid`);
		error.name = "Svelte error";
		throw error;
	}
	/**
	* The `html` property of server render results has been deprecated. Use `body` instead.
	* @returns {never}
	*/
	function html_deprecated() {
		const error = /* @__PURE__ */ new Error(`html_deprecated\nThe \`html\` property of server render results has been deprecated. Use \`body\` instead.\nhttps://svelte.dev/e/html_deprecated`);
		error.name = "Svelte error";
		throw error;
	}
	/**
	* `csp.nonce` was set while `csp.hash` was `true`. These options cannot be used simultaneously.
	* @returns {never}
	*/
	function invalid_csp() {
		const error = /* @__PURE__ */ new Error(`invalid_csp\n\`csp.nonce\` was set while \`csp.hash\` was \`true\`. These options cannot be used simultaneously.\nhttps://svelte.dev/e/invalid_csp`);
		error.name = "Svelte error";
		throw error;
	}
	/**
	* The `idPrefix` option cannot include `--`.
	* @returns {never}
	*/
	function invalid_id_prefix() {
		const error = /* @__PURE__ */ new Error(`invalid_id_prefix\nThe \`idPrefix\` option cannot include \`--\`.\nhttps://svelte.dev/e/invalid_id_prefix`);
		error.name = "Svelte error";
		throw error;
	}
	/**
	* Could not resolve `render` context.
	* @returns {never}
	*/
	function server_context_required() {
		const error = /* @__PURE__ */ new Error(`server_context_required\nCould not resolve \`render\` context.\nhttps://svelte.dev/e/server_context_required`);
		error.name = "Svelte error";
		throw error;
	}
	/**
	* A `hydratable` value with key `%key%` was created, but at least part of it was not used during the render.
	* 
	* The `hydratable` was initialized in:
	* %stack%
	* @param {string} key
	* @param {string} stack
	*/
	function unresolved_hydratable(key, stack) {
		console.warn(`https://svelte.dev/e/unresolved_hydratable`);
	}
	var require__skgo_skgo_missing = /* @__PURE__ */ __commonJSMin((() => {
		throw new Error("skgo: node:async_hooks is unavailable in the SSR engine");
	}));
	/** @import { AsyncLocalStorage } from 'node:async_hooks' */
	/** @import { RenderContext } from '#server' */
	/** @type {Promise<void> | null} */
	var current_render = null;
	/** @type {RenderContext | null} */
	var context$1 = null;
	/** @returns {RenderContext} */
	function get_render_context() {
		const store = context$1 ?? als$1?.getStore();
		if (!store) server_context_required();
		return store;
	}
	/**
	* @template T
	* @param {() => Promise<T>} fn
	* @returns {Promise<T>}
	*/
	async function with_render_context(fn) {
		context$1 = { hydratable: {
			lookup: /* @__PURE__ */ new Map(),
			comparisons: [],
			unresolved_promises: /* @__PURE__ */ new Map()
		} };
		if (in_webcontainer()) {
			const { promise, resolve } = deferred();
			const previous_render = current_render;
			current_render = promise;
			await previous_render;
			return fn().finally(resolve);
		}
		try {
			if (als$1 === null) async_local_storage_unavailable();
			return als$1.run(context$1, fn);
		} finally {
			context$1 = null;
		}
	}
	/** @type {AsyncLocalStorage<RenderContext | null> | null} */
	var als$1 = null;
	/** @type {Promise<void> | null} */
	var als_import = null;
	/**
	*
	* @returns {Promise<void>}
	*/
	function init_render_context() {
		als_import ??= Promise.resolve().then(() => /* @__PURE__ */ __toESM(require__skgo_skgo_missing(), 1)).then((hooks) => {
			als$1 = new hooks.AsyncLocalStorage();
		}).then(noop$1, noop$1);
		return als_import;
	}
	function in_webcontainer() {
		return !!globalThis.process?.versions?.webcontainer;
	}
	var text_encoder$2;
	var crypto;
	/** @param {string} module_name */
	var obfuscated_import = (module_name) => Promise.reject(/* @__PURE__ */ new Error("skgo: no dynamic import in the SSR engine"));
	/** @param {string} data */
	async function sha256(data) {
		text_encoder$2 ??= new TextEncoder();
		crypto ??= globalThis.crypto?.subtle?.digest ? globalThis.crypto : (await obfuscated_import("node:crypto")).webcrypto;
		return base64_encode$1(await crypto.subtle.digest("SHA-256", text_encoder$2.encode(data)));
	}
	/**
	* @param {Uint8Array} bytes
	* @returns {string}
	*/
	function base64_encode$1(bytes) {
		if (globalThis.Buffer) return globalThis.Buffer.from(bytes).toString("base64");
		let binary = "";
		for (let i = 0; i < bytes.length; i++) binary += String.fromCharCode(bytes[i]);
		return btoa(binary);
	}
	var MAX_ARRAY_LEN = 2 ** 32 - 1;
	var MAX_ARRAY_INDEX = MAX_ARRAY_LEN - 1;
	/** @type {Record<string, string>} */
	var escaped = {
		"<": "\\u003C",
		"\\": "\\\\",
		"\b": "\\b",
		"\f": "\\f",
		"\n": "\\n",
		"\r": "\\r",
		"	": "\\t",
		"\u2028": "\\u2028",
		"\u2029": "\\u2029"
	};
	var DevalueError = class extends Error {
		/**
		* @param {string} message
		* @param {string[]} keys
		* @param {any} [value] - The value that failed to be serialized
		* @param {any} [root] - The root value being serialized
		*/
		constructor(message, keys, value, root) {
			super(message);
			this.name = "DevalueError";
			this.path = keys.join("");
			this.value = value;
			this.root = root;
		}
	};
	/** @param {any} thing */
	function is_primitive(thing) {
		return thing === null || typeof thing !== "object" && typeof thing !== "function";
	}
	var object_proto_names$1 = /* @__PURE__ */ Object.getOwnPropertyNames(Object.prototype).sort().join("\0");
	/** @param {any} thing */
	function is_plain_object$1(thing) {
		const proto = Object.getPrototypeOf(thing);
		return proto === Object.prototype || proto === null || Object.getPrototypeOf(proto) === null || Object.getOwnPropertyNames(proto).sort().join("\0") === object_proto_names$1;
	}
	/** @param {any} thing */
	function get_type(thing) {
		return Object.prototype.toString.call(thing).slice(8, -1);
	}
	/** @param {string} char */
	function get_escaped_char(char) {
		switch (char) {
			case "\"": return "\\\"";
			case "<": return "\\u003C";
			case "\\": return "\\\\";
			case "\n": return "\\n";
			case "\r": return "\\r";
			case "	": return "\\t";
			case "\b": return "\\b";
			case "\f": return "\\f";
			case "\u2028": return "\\u2028";
			case "\u2029": return "\\u2029";
			default: return char < " " ? `\\u${char.charCodeAt(0).toString(16).padStart(4, "0")}` : "";
		}
	}
	/** @param {string} str */
	function stringify_string(str) {
		let result = "";
		let last_pos = 0;
		const len = str.length;
		for (let i = 0; i < len; i += 1) {
			const char = str[i];
			const replacement = get_escaped_char(char);
			if (replacement) {
				result += str.slice(last_pos, i) + replacement;
				last_pos = i + 1;
			}
		}
		return `"${last_pos === 0 ? str : result + str.slice(last_pos)}"`;
	}
	/** @param {Record<string | symbol, any>} object */
	function enumerable_symbols(object) {
		return Object.getOwnPropertySymbols(object).filter((symbol) => Object.getOwnPropertyDescriptor(object, symbol).enumerable);
	}
	var is_identifier = /^[a-zA-Z_$][a-zA-Z_$0-9]*$/;
	/** @param {string} key */
	function stringify_key(key) {
		return is_identifier.test(key) ? "." + key : "[" + JSON.stringify(key) + "]";
	}
	/** @param {number} n */
	function is_valid_array_index(n) {
		if (!Number.isInteger(n)) return false;
		if (n < 0) return false;
		if (n > MAX_ARRAY_INDEX) return false;
		return true;
	}
	/** @param {number} n */
	function is_valid_array_len(n) {
		if (!Number.isInteger(n)) return false;
		if (n < 0) return false;
		if (n > MAX_ARRAY_LEN) return false;
		return true;
	}
	/** @param {string} s */
	function is_valid_array_index_string(s) {
		if (s.length === 0) return false;
		if (s.length > 1 && s.charCodeAt(0) === 48) return false;
		for (let i = 0; i < s.length; i++) {
			const c = s.charCodeAt(i);
			if (c < 48 || c > 57) return false;
		}
		return is_valid_array_index(+s);
	}
	/**
	* Returns the length of the leading run of valid array indices in `keys`.
	* @param {readonly string[]} keys
	*/
	function array_index_cut(keys) {
		for (var i = keys.length - 1; i >= 0; i--) if (is_valid_array_index_string(keys[i])) break;
		return i + 1;
	}
	/**
	* Finds the populated indices of an array.
	* @param {unknown[]} array
	*/
	function valid_array_indices(array) {
		const keys = Object.keys(array);
		keys.length = array_index_cut(keys);
		return keys;
	}
	var chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_$";
	var unsafe_chars = /[<\b\f\n\r\t\0\u2028\u2029]/g;
	var reserved = /^(?:do|if|in|for|int|let|new|try|var|byte|case|char|else|enum|goto|long|this|void|with|await|break|catch|class|const|final|float|short|super|throw|while|yield|delete|double|export|import|native|return|switch|throws|typeof|boolean|default|extends|finally|package|private|abstract|continue|debugger|function|volatile|interface|protected|transient|implements|instanceof|synchronized)$/;
	/**
	* Turn a value into the JavaScript that creates an equivalent value
	* @param {any} value
	* @param {(value: any, uneval: (value: any) => string) => string | void} [replacer]
	*/
	function uneval$1(value, replacer) {
		const counts = /* @__PURE__ */ new Map();
		/** @type {string[]} */
		const keys = [];
		const custom = /* @__PURE__ */ new Map();
		/** @param {any} thing */
		function walk(thing) {
			if (!is_primitive(thing)) {
				if (counts.has(thing)) {
					counts.set(thing, counts.get(thing) + 1);
					return;
				}
				counts.set(thing, 1);
				if (replacer) {
					const str = replacer(thing, (value) => uneval$1(value, replacer));
					if (typeof str === "string") {
						custom.set(thing, str);
						return;
					}
				}
				if (typeof thing === "function") throw new DevalueError(`Cannot stringify a function`, keys, thing, value);
				switch (get_type(thing)) {
					case "Number":
					case "BigInt":
					case "String":
					case "Boolean":
					case "Date":
					case "RegExp":
					case "URL":
					case "URLSearchParams": return;
					case "Array":
						/** @type {any[]} */ thing.forEach((value, i) => {
							keys.push(`[${i}]`);
							walk(value);
							keys.pop();
						});
						break;
					case "Set":
						Array.from(thing).forEach(walk);
						break;
					case "Map":
						for (const [key, value] of thing) {
							keys.push(`.get(${is_primitive(key) ? stringify_primitive$1(key) : "..."})`);
							walk(key);
							walk(value);
							keys.pop();
						}
						break;
					case "Int8Array":
					case "Uint8Array":
					case "Uint8ClampedArray":
					case "Int16Array":
					case "Uint16Array":
					case "Float16Array":
					case "Int32Array":
					case "Uint32Array":
					case "Float32Array":
					case "Float64Array":
					case "BigInt64Array":
					case "BigUint64Array":
					case "DataView":
						walk(thing.buffer);
						return;
					case "ArrayBuffer": return;
					case "Temporal.Duration":
					case "Temporal.Instant":
					case "Temporal.PlainDate":
					case "Temporal.PlainTime":
					case "Temporal.PlainDateTime":
					case "Temporal.PlainMonthDay":
					case "Temporal.PlainYearMonth":
					case "Temporal.ZonedDateTime": return;
					default:
						if (!is_plain_object$1(thing)) throw new DevalueError(`Cannot stringify arbitrary non-POJOs`, keys, thing, value);
						if (enumerable_symbols(thing).length > 0) throw new DevalueError(`Cannot stringify POJOs with symbolic keys`, keys, thing, value);
						for (const key of Object.keys(thing)) {
							if (key === "__proto__") throw new DevalueError(`Cannot stringify objects with __proto__ keys`, keys, thing, value);
							keys.push(stringify_key(key));
							walk(thing[key]);
							keys.pop();
						}
				}
			} else if (typeof thing === "symbol") throw new DevalueError(`Cannot stringify a Symbol primitive`, keys, thing, value);
		}
		walk(value);
		const names = /* @__PURE__ */ new Map();
		Array.from(counts).filter((entry) => entry[1] > 1).sort((a, b) => b[1] - a[1]).forEach((entry, i) => {
			names.set(entry[0], get_name(i));
		});
		/**
		* @param {any} thing
		* @returns {string}
		*/
		function stringify(thing) {
			if (names.has(thing)) return names.get(thing);
			if (is_primitive(thing)) return stringify_primitive$1(thing);
			if (custom.has(thing)) return custom.get(thing);
			const type = get_type(thing);
			switch (type) {
				case "Number":
				case "String":
				case "Boolean":
				case "BigInt": return `Object(${stringify(thing.valueOf())})`;
				case "RegExp":
					const { source, flags } = thing;
					return flags ? `new RegExp(${stringify_string(source)},"${flags}")` : `new RegExp(${stringify_string(source)})`;
				case "Date": return `new Date(${thing.getTime()})`;
				case "URL": return `new URL(${stringify_string(thing.toString())})`;
				case "URLSearchParams": return `new URLSearchParams(${stringify_string(thing.toString())})`;
				case "Array": {
					let has_holes = false;
					let result = "[";
					for (let i = 0; i < thing.length; i += 1) {
						if (i > 0) result += ",";
						if (Object.hasOwn(thing, i)) result += stringify(thing[i]);
						else if (!has_holes) {
							const populated_keys = valid_array_indices(thing);
							const population = populated_keys.length;
							const d = String(thing.length).length;
							if (thing.length + 2 > 25 + d + population * (d + 2)) {
								const entries = populated_keys.map((k) => `${k}:${stringify(thing[k])}`).join(",");
								return `Object.assign(Array(${thing.length}),{${entries}})`;
							}
							has_holes = true;
						}
					}
					const tail = thing.length === 0 || thing.length - 1 in thing ? "" : ",";
					return result + tail + "]";
				}
				case "Set":
				case "Map": return `new ${type}([${Array.from(thing).map(stringify).join(",")}])`;
				case "Int8Array":
				case "Uint8Array":
				case "Uint8ClampedArray":
				case "Int16Array":
				case "Uint16Array":
				case "Float16Array":
				case "Int32Array":
				case "Uint32Array":
				case "Float32Array":
				case "Float64Array":
				case "BigInt64Array":
				case "BigUint64Array": {
					let str = `new ${type}`;
					if (!names.has(thing.buffer)) str += `([${stringify_typed_array_elements(type, thing.buffer)}])`;
					else str += `(${stringify(thing.buffer)})`;
					if (thing.byteLength !== thing.buffer.byteLength) {
						const start = thing.byteOffset / thing.BYTES_PER_ELEMENT;
						const end = start + thing.length;
						str += `.subarray(${start},${end})`;
					}
					return str;
				}
				case "DataView": {
					let str = `new DataView`;
					if (!names.has(thing.buffer)) str += `(new Uint8Array([${new Uint8Array(thing.buffer)}]).buffer`;
					else str += `(${stringify(thing.buffer)}`;
					if (thing.byteLength !== thing.buffer.byteLength) str += `,${thing.byteOffset},${thing.byteLength}`;
					return str + ")";
				}
				case "ArrayBuffer": return `new Uint8Array([${new Uint8Array(thing).toString()}]).buffer`;
				case "Temporal.Duration":
				case "Temporal.Instant":
				case "Temporal.PlainDate":
				case "Temporal.PlainTime":
				case "Temporal.PlainDateTime":
				case "Temporal.PlainMonthDay":
				case "Temporal.PlainYearMonth":
				case "Temporal.ZonedDateTime": return `${type}.from(${stringify_string(thing.toString())})`;
				default:
					const keys = Object.keys(thing);
					const obj = keys.map((key) => `${safe_key(key)}:${stringify(thing[key])}`).join(",");
					if (Object.getPrototypeOf(thing) === null) return keys.length > 0 ? `{${obj},__proto__:null}` : `{__proto__:null}`;
					return `{${obj}}`;
			}
		}
		const str = stringify(value);
		if (names.size) {
			/** @type {string[]} */
			const params = [];
			/** @type {string[]} */
			const statements = [];
			/** @type {string[]} */
			const values = [];
			/** @type {string[]} */
			const reconstructions = [];
			names.forEach((name, thing) => {
				params.push(name);
				if (custom.has(thing)) {
					values.push(custom.get(thing));
					return;
				}
				if (is_primitive(thing)) {
					values.push(stringify_primitive$1(thing));
					return;
				}
				const type = get_type(thing);
				switch (type) {
					case "Number":
					case "String":
					case "Boolean":
					case "BigInt":
						values.push(`Object(${stringify(thing.valueOf())})`);
						break;
					case "RegExp":
						const { source, flags } = thing;
						const regexp = flags ? `new RegExp(${stringify_string(source)},"${flags}")` : `new RegExp(${stringify_string(source)})`;
						values.push(regexp);
						break;
					case "Date":
						values.push(`new Date(${thing.getTime()})`);
						break;
					case "URL":
						values.push(`new URL(${stringify_string(thing.toString())})`);
						break;
					case "URLSearchParams":
						values.push(`new URLSearchParams(${stringify_string(thing.toString())})`);
						break;
					case "Array":
						values.push(`Array(${thing.length})`);
						/** @type {any[]} */ thing.forEach((v, i) => {
							statements.push(`${name}[${i}]=${stringify(v)}`);
						});
						break;
					case "Set": {
						values.push(`new Set`);
						const adds = Array.from(thing).map((v) => `.add(${stringify(v)})`);
						if (adds.length > 0) statements.push(name + adds.join(""));
						break;
					}
					case "Map": {
						values.push(`new Map`);
						const sets = Array.from(thing).map(([k, v]) => `.set(${stringify(k)}, ${stringify(v)})`);
						if (sets.length > 0) statements.push(name + sets.join(""));
						break;
					}
					case "Int8Array":
					case "Uint8Array":
					case "Uint8ClampedArray":
					case "Int16Array":
					case "Uint16Array":
					case "Float16Array":
					case "Int32Array":
					case "Uint32Array":
					case "Float32Array":
					case "Float64Array":
					case "BigInt64Array":
					case "BigUint64Array": {
						let str = `new ${type}`;
						if (!names.has(thing.buffer)) str += `([${stringify_typed_array_elements(type, thing.buffer)}])`;
						else str += `(${stringify(thing.buffer)})`;
						if (thing.byteLength !== thing.buffer.byteLength) {
							const start = thing.byteOffset / thing.BYTES_PER_ELEMENT;
							const end = start + thing.length;
							str += `.subarray(${start},${end})`;
						}
						values.push(`{}`);
						reconstructions.push(`${name}=${str}`);
						break;
					}
					case "DataView": {
						let str = `new DataView`;
						if (!names.has(thing.buffer)) str += `(new Uint8Array([${new Uint8Array(thing.buffer)}]).buffer`;
						else str += `(${stringify(thing.buffer)}`;
						if (thing.byteLength !== thing.buffer.byteLength) str += `,${thing.byteOffset},${thing.byteLength}`;
						str += ")";
						values.push(`{}`);
						reconstructions.push(`${name}=${str}`);
						break;
					}
					case "ArrayBuffer":
						values.push(`new Uint8Array([${new Uint8Array(thing)}]).buffer`);
						break;
					case "Temporal.Duration":
					case "Temporal.Instant":
					case "Temporal.PlainDate":
					case "Temporal.PlainTime":
					case "Temporal.PlainDateTime":
					case "Temporal.PlainMonthDay":
					case "Temporal.PlainYearMonth":
					case "Temporal.ZonedDateTime":
						values.push(`${type}.from(${stringify_string(thing.toString())})`);
						break;
					default:
						values.push(Object.getPrototypeOf(thing) === null ? "Object.create(null)" : "{}");
						Object.keys(thing).forEach((key) => {
							statements.push(`${name}${safe_prop(key)}=${stringify(thing[key])}`);
						});
				}
			});
			statements.push(`return ${str}`);
			const body = [...reconstructions, ...statements].join(";");
			if (params.length > 65534) return `(function(){var[${params.join(",")}]=arguments[0];${body}}([${values.join(",")}]))`;
			return `(function(${params.join(",")}){${body}}(${values.join(",")}))`;
		} else return str;
	}
	/**
	* Serialize the elements of `buffer`, read as `type`, as a comma-separated list.
	* The view is created from `type` rather than from the serialized value's own
	* constructor, which may be a subclass like Node's `Buffer` whose `toString`
	* decodes the bytes instead of listing them.
	* `BigInt64Array`/`BigUint64Array` elements are bigints and must be written
	* with an `n` suffix, otherwise the emitted `new BigInt64Array([...])` throws.
	* @param {string} type
	* @param {ArrayBufferLike} buffer
	*/
	function stringify_typed_array_elements(type, buffer) {
		const array = new globalThis[type](buffer);
		if (type === "BigInt64Array" || type === "BigUint64Array") return Array.from(array, (element) => `${element}n`).join(",");
		if (array instanceof Float32Array || array instanceof Float64Array || typeof Float16Array !== "undefined" && array instanceof Float16Array) return Array.from(array, (element) => Object.is(element, -0) ? "-0" : `${element}`).join(",");
		return array.toString();
	}
	/** @param {number} num */
	function get_name(num) {
		let name = "";
		do {
			name = chars[num % 54] + name;
			num = ~~(num / 54) - 1;
		} while (num >= 0);
		return reserved.test(name) ? `${name}0` : name;
	}
	/** @param {string} c */
	function escape_unsafe_char(c) {
		return escaped[c] || c;
	}
	/** @param {string} str */
	function escape_unsafe_chars(str) {
		return str.replace(unsafe_chars, escape_unsafe_char);
	}
	/** @param {string} key */
	function safe_key(key) {
		return /^[_$a-zA-Z][_$a-zA-Z0-9]*$/.test(key) ? key : escape_unsafe_chars(JSON.stringify(key));
	}
	/** @param {string} key */
	function safe_prop(key) {
		return /^[_$a-zA-Z][_$a-zA-Z0-9]*$/.test(key) ? `.${key}` : `[${escape_unsafe_chars(JSON.stringify(key))}]`;
	}
	/** @param {any} thing */
	function stringify_primitive$1(thing) {
		const type = typeof thing;
		if (type === "string") return stringify_string(thing);
		if (thing === void 0) return "void 0";
		if (thing === 0 && 1 / thing < 0) return "-0";
		const str = String(thing);
		if (type === "number") return str.replace(/^(-)?0\./, "$1.");
		if (type === "bigint") return thing + "n";
		return str;
	}
	/**	@type {(array_buffer: ArrayBuffer) => string} */
	function encode_native(array_buffer) {
		return new Uint8Array(array_buffer).toBase64();
	}
	/**	@type {(base64: string) => ArrayBuffer} */
	function decode_native(base64) {
		return Uint8Array.fromBase64(base64).buffer;
	}
	/** @type {(array_buffer: ArrayBuffer) => string} */
	function encode_buffer(array_buffer) {
		return Buffer.from(array_buffer).toString("base64");
	}
	/**	@type {(base64: string) => ArrayBuffer} */
	function decode_buffer(base64) {
		return Uint8Array.from(Buffer.from(base64, "base64")).buffer;
	}
	/** @type {(array_buffer: ArrayBuffer) => string} */
	function encode_legacy(array_buffer) {
		const array = new Uint8Array(array_buffer);
		let binary = "";
		const chunk_size = 32768;
		for (let i = 0; i < array.length; i += chunk_size) {
			const chunk = array.subarray(i, i + chunk_size);
			binary += String.fromCharCode.apply(null, chunk);
		}
		return btoa(binary);
	}
	/**	@type {(base64: string) => ArrayBuffer} */
	function decode_legacy(base64) {
		const binary_string = atob(base64);
		const len = binary_string.length;
		const array = new Uint8Array(len);
		for (let i = 0; i < len; i++) array[i] = binary_string.charCodeAt(i);
		return array.buffer;
	}
	var native = typeof Uint8Array.fromBase64 === "function";
	var buffer = typeof process === "object" && process.versions?.node !== void 0;
	var encode64 = native ? encode_native : buffer ? encode_buffer : encode_legacy;
	var decode64 = native ? decode_native : buffer ? decode_buffer : decode_legacy;
	/**
	* Merges caller-provided operation overrides over the defaults. Iterating the
	* default keys (rather than the override's own keys) means nullish members
	* fall back to the default, and inherited members — e.g. from a class
	* instance — are picked up.
	*
	* @template {Record<string, any>} T
	* @param {T} defaults
	* @param {Partial<T> | undefined} overrides
	* @returns {T}
	*/
	function merge_operations(defaults, overrides) {
		if (!overrides) return defaults;
		const merged = {};
		for (const key of Object.keys(defaults)) merged[key] = overrides[key] ?? defaults[key];
		return merged;
	}
	/** @type {{ kind: 'not-plain' }} */
	var NOT_PLAIN = Object.freeze({ kind: "not-plain" });
	/** @type {{ kind: 'symbol-keys' }} */
	var SYMBOL_KEYS = Object.freeze({ kind: "symbol-keys" });
	var default_stringify_operations = Object.freeze({
		identify: (value) => value,
		typeOf: (value) => value === null ? "null" : typeof value,
		toPrimitive: (value) => value,
		tagOf: (value) => get_type(value),
		isThenable: (value) => typeof value.then === "function",
		toPromise: (thenable) => Promise.resolve(thenable),
		unbox: (boxed) => boxed.valueOf(),
		toISOString: (date) => isNaN(date.getDate()) ? "" : date.toISOString(),
		toStringValue: (value) => value.toString(),
		regExpInfo: (regexp) => ({
			source: regexp.source,
			flags: regexp.flags
		}),
		valuesOf: (set) => set,
		entriesOf: (map) => map,
		viewInfo: (view) => ({
			buffer: view.buffer,
			byteOffset: view.byteOffset,
			byteLength: view.byteLength,
			length: view.length,
			bufferByteLength: view.buffer.byteLength
		}),
		toArrayBuffer: (buffer) => buffer,
		lengthOf: (array) => array.length,
		hasOwn: (value, key) => Object.hasOwn(value, key),
		indicesOf: (array) => valid_array_indices(array),
		shapeOf: (value) => {
			if (!is_plain_object$1(value)) return NOT_PLAIN;
			if (enumerable_symbols(value).length > 0) return SYMBOL_KEYS;
			return {
				kind: Object.getPrototypeOf(value) === null ? "null-proto" : "plain",
				keys: Object.keys(value)
			};
		},
		get: (value, key) => value[key]
	});
	var default_parse_operations = Object.freeze({
		fromPrimitive: (primitive) => primitive,
		fromISOString: (iso) => new Date(iso),
		fromStringValue: (tag, text) => {
			if (tag === "URL") return new URL(text);
			if (tag === "URLSearchParams") return new URLSearchParams(text);
			return Temporal[tag.slice(9)].from(text);
		},
		fromArrayBuffer: (buffer) => buffer,
		fromRegExpInfo: (source, flags) => new RegExp(source, flags),
		fromViewInfo: (tag, buffer, byteOffset, length) => {
			const Constructor = globalThis[tag];
			return byteOffset !== void 0 ? new Constructor(buffer, byteOffset, length) : new Constructor(buffer);
		},
		box: (value) => Object(value),
		createArray: (length) => new Array(length),
		createSparseArray: (length) => {
			/** @type {any[]} */
			const array = [];
			array[MAX_ARRAY_INDEX] = void 0;
			delete array[MAX_ARRAY_INDEX];
			array.length = length;
			return array;
		},
		createObject: () => ({}),
		createNullPrototypeObject: () => Object.create(null),
		createSet: () => /* @__PURE__ */ new Set(),
		createMap: () => /* @__PURE__ */ new Map(),
		set: (target, key, value) => {
			target[key] = value;
		},
		addValue: (set, value) => {
			set.add(value);
		},
		addEntry: (map, key, value) => {
			map.set(key, value);
		}
	});
	/**
	* Revive a value serialized with `devalue.stringify`
	* @param {string} serialized
	* @param {Record<string, (value: any) => any>} [revivers]
	* @param {import('./types.js').ParseOptions} [options]
	*/
	function parse$1(serialized, revivers, options) {
		return unflatten(JSON.parse(serialized), revivers, options);
	}
	/**
	* Revive a value flattened with `devalue.stringify`
	* @param {number | any[]} parsed
	* @param {Record<string, (value: any) => any>} [revivers]
	* @param {import('./types.js').ParseOptions} [options]
	*/
	function unflatten(parsed, revivers, options) {
		/** @type {import('./types.js').ParseOperations} */
		const ops = merge_operations(default_parse_operations, options?.operations);
		if (typeof parsed === "number") return hydrate(parsed, true);
		if (!Array.isArray(parsed) || parsed.length === 0) throw new Error("Invalid input");
		const values = parsed;
		const hydrated = Array(values.length);
		/**
		* A set of values currently being hydrated with custom revivers,
		* used to detect invalid cyclical dependencies
		* @type {Set<number> | null}
		*/
		let hydrating = null;
		/**
		* @param {number} index
		* @returns {any}
		*/
		function hydrate(index, standalone = false) {
			if (index === -1) return ops.fromPrimitive(void 0);
			if (index === -3) return ops.fromPrimitive(NaN);
			if (index === -4) return ops.fromPrimitive(Infinity);
			if (index === -5) return ops.fromPrimitive(-Infinity);
			if (index === -6) return ops.fromPrimitive(-0);
			if (standalone || typeof index !== "number") throw new Error(`Invalid input`);
			if (index in hydrated) return hydrated[index];
			if (index >= values.length) throw new Error(`Invalid input`);
			const value = values[index];
			if (!value || typeof value !== "object") hydrated[index] = ops.fromPrimitive(value);
			else if (Array.isArray(value)) {
				if (typeof value[0] === "string") {
					const type = value[0];
					const reviver = revivers && Object.hasOwn(revivers, type) ? revivers[type] : void 0;
					if (reviver) {
						let i = value[1];
						if (typeof i !== "number") i = values.push(value[1]) - 1;
						if (Object.hasOwn(hydrated, i)) return hydrated[index] = reviver(hydrated[i]);
						hydrating ??= /* @__PURE__ */ new Set();
						if (hydrating.has(i)) throw new Error("Invalid circular reference");
						hydrating.add(i);
						hydrated[index] = reviver(hydrate(i));
						hydrating.delete(i);
						return hydrated[index];
					}
					switch (type) {
						case "Date":
							hydrated[index] = ops.fromISOString(value[1]);
							break;
						case "Set":
							const set = ops.createSet();
							hydrated[index] = set;
							for (let i = 1; i < value.length; i += 1) ops.addValue(set, hydrate(value[i]));
							break;
						case "Map":
							const map = ops.createMap();
							hydrated[index] = map;
							for (let i = 1; i < value.length; i += 2) ops.addEntry(map, hydrate(value[i]), hydrate(value[i + 1]));
							break;
						case "RegExp":
							hydrated[index] = ops.fromRegExpInfo(value[1], value[2]);
							break;
						case "Object": {
							const wrapped_index = value[1];
							if (typeof values[wrapped_index] === "object" && values[wrapped_index][0] !== "BigInt") throw new Error("Invalid input");
							hydrated[index] = ops.box(hydrate(wrapped_index));
							break;
						}
						case "BigInt":
							hydrated[index] = ops.fromPrimitive(BigInt(value[1]));
							break;
						case "null":
							const obj = ops.createNullPrototypeObject();
							hydrated[index] = obj;
							for (let i = 1; i < value.length; i += 2) {
								if (value[i] === "__proto__") throw new Error("Cannot parse an object with a `__proto__` property");
								ops.set(obj, value[i], hydrate(value[i + 1]));
							}
							break;
						case "Int8Array":
						case "Uint8Array":
						case "Uint8ClampedArray":
						case "Int16Array":
						case "Uint16Array":
						case "Float16Array":
						case "Int32Array":
						case "Uint32Array":
						case "Float32Array":
						case "Float64Array":
						case "BigInt64Array":
						case "BigUint64Array":
						case "DataView": {
							if (values[value[1]][0] !== "ArrayBuffer") throw new Error("Invalid data");
							const buffer = hydrate(value[1]);
							hydrated[index] = ops.fromViewInfo(type, buffer, value[2], value[3]);
							break;
						}
						case "ArrayBuffer": {
							const base64 = value[1];
							if (typeof base64 !== "string") throw new Error("Invalid ArrayBuffer encoding");
							hydrated[index] = ops.fromArrayBuffer(decode64(base64));
							break;
						}
						case "URL":
						case "URLSearchParams":
						case "Temporal.Duration":
						case "Temporal.Instant":
						case "Temporal.PlainDate":
						case "Temporal.PlainTime":
						case "Temporal.PlainDateTime":
						case "Temporal.PlainMonthDay":
						case "Temporal.PlainYearMonth":
						case "Temporal.ZonedDateTime":
							hydrated[index] = ops.fromStringValue(type, value[1]);
							break;
						default: throw new Error(`Unknown type ${type}`);
					}
				} else if (value[0] === -7) {
					const len = value[1];
					if (!is_valid_array_len(len)) throw new Error("Invalid input");
					const array = ops.createSparseArray(len);
					hydrated[index] = array;
					for (let i = 2; i < value.length; i += 2) {
						const idx = value[i];
						if (!is_valid_array_index(idx) || idx >= len) throw new Error("Invalid input");
						ops.set(array, idx, hydrate(value[i + 1]));
					}
				} else {
					const array = ops.createArray(value.length);
					hydrated[index] = array;
					for (let i = 0; i < value.length; i += 1) {
						const n = value[i];
						if (n === -2) continue;
						ops.set(array, i, hydrate(n));
					}
				}
			} else {
				const object = ops.createObject();
				hydrated[index] = object;
				for (const key of Object.keys(value)) {
					if (key === "__proto__") throw new Error("Cannot parse an object with a `__proto__` property");
					ops.set(object, key, hydrate(value[key]));
				}
			}
			return hydrated[index];
		}
		return hydrate(0);
	}
	/**
	* Turn a value into a JSON string that can be parsed with `devalue.parse`
	* @param {any} value
	* @param {Record<string, (value: any) => any>} [reducers]
	* @param {import('./types.js').StringifyOptions} [options]
	*/
	function stringify$1(value, reducers, options) {
		const stringified = run(false, value, reducers, options);
		return typeof stringified === "string" ? stringified : `[${stringified.join(",")}]`;
	}
	/**
	* @param {boolean} async
	* @param {any} value
	* @param {Record<string, (value: any) => any>} [reducers]
	* @param {import('./types.js').StringifyOptions} [options]
	*/
	function run(async, value, reducers, options) {
		const ops = merge_operations(default_stringify_operations, options === null || options === void 0 ? void 0 : options.operations);
		/** @type {any[]} */
		const stringified = [];
		/** @type {Map<any, number>} */
		const indexes = /* @__PURE__ */ new Map();
		/** @type {Array<{ key: string, fn: (value: any) => any }>} */
		const custom = [];
		if (reducers) for (const key of Object.getOwnPropertyNames(reducers)) custom.push({
			key,
			fn: reducers[key]
		});
		/** @type {string[]} */
		const keys = [];
		let p = 0;
		/**
		* @param {any} thing
		* @param {number} [index]
		*/
		function flatten(thing, index) {
			var _index;
			const type = ops.typeOf(thing);
			if (type === "undefined") return -1;
			/** @type {number | undefined} */
			let number;
			if (type === "number") {
				number = ops.toPrimitive(thing);
				if (Number.isNaN(number)) return -3;
				if (number === Infinity) return -4;
				if (number === -Infinity) return -5;
				if (number === 0 && 1 / number < 0) return -6;
			}
			const id = ops.identify(thing);
			if (indexes.has(id)) return indexes.get(id);
			(_index = index) !== null && _index !== void 0 || (index = p++);
			indexes.set(id, index);
			for (const { key, fn } of custom) {
				const value = fn(thing);
				if (value) {
					stringified[index] = `["${key}",${flatten(value)}]`;
					return index;
				}
			}
			if (type === "function") throw new DevalueError(`Cannot stringify a function`, keys, thing, value);
			else if (type === "symbol") throw new DevalueError(`Cannot stringify a Symbol primitive`, keys, thing, value);
			/** @type {string | Promise<any>} */
			let str = "";
			if (type !== "object") str = stringify_primitive(type === "number" ? number : ops.toPrimitive(thing));
			else if (ops.isThenable(thing)) {
				if (!async) throw new DevalueError(`Cannot stringify a Promise or thenable — use stringifyAsync instead`, keys, thing, value);
				str = ops.toPromise(thing).then((value) => {
					const i = flatten(value, index);
					if (i < 0) stringified[index] = i;
				});
			} else {
				const tag = ops.tagOf(thing);
				switch (tag) {
					case "Number":
					case "String":
					case "Boolean":
					case "BigInt":
						str = `["Object",${flatten(ops.unbox(thing))}]`;
						break;
					case "Date":
						str = `["Date","${ops.toISOString(thing)}"]`;
						break;
					case "URL":
						str = `["URL",${stringify_string(ops.toStringValue(thing))}]`;
						break;
					case "URLSearchParams":
						str = `["URLSearchParams",${stringify_string(ops.toStringValue(thing))}]`;
						break;
					case "RegExp":
						const { source, flags } = ops.regExpInfo(thing);
						str = flags ? `["RegExp",${stringify_string(source)},"${flags}"]` : `["RegExp",${stringify_string(source)}]`;
						break;
					case "Array": {
						let mostly_dense = false;
						const length = ops.lengthOf(thing);
						str = "[";
						for (let i = 0; i < length; i += 1) {
							if (i > 0) str += ",";
							if (ops.hasOwn(thing, i)) {
								keys.push(`[${i}]`);
								str += flatten(ops.get(thing, i));
								keys.pop();
							} else if (mostly_dense) str += -2;
							else {
								const populated_keys = ops.indicesOf(thing);
								const population = populated_keys.length;
								const d = String(length).length;
								if ((length - population) * 3 > 4 + d + population * (d + 1)) {
									str = "[-7," + length;
									for (let j = 0; j < populated_keys.length; j++) {
										const key = populated_keys[j];
										keys.push(`[${key}]`);
										str += "," + key + "," + flatten(ops.get(thing, key));
										keys.pop();
									}
									break;
								} else {
									mostly_dense = true;
									str += -2;
								}
							}
						}
						str += "]";
						break;
					}
					case "Set":
						str = "[\"Set\"";
						for (const value of ops.valuesOf(thing)) str += `,${flatten(value)}`;
						str += "]";
						break;
					case "Map":
						str = "[\"Map\"";
						for (const [key, value] of ops.entriesOf(thing)) {
							const key_type = ops.typeOf(key);
							const key_is_primitive = key_type !== "object" && key_type !== "function" && key_type !== "symbol";
							keys.push(`.get(${key_is_primitive ? stringify_primitive(ops.toPrimitive(key)) : "..."})`);
							str += `,${flatten(key)},${flatten(value)}`;
							keys.pop();
						}
						str += "]";
						break;
					case "Int8Array":
					case "Uint8Array":
					case "Uint8ClampedArray":
					case "Int16Array":
					case "Uint16Array":
					case "Float16Array":
					case "Int32Array":
					case "Uint32Array":
					case "Float32Array":
					case "Float64Array":
					case "BigInt64Array":
					case "BigUint64Array": {
						const info = ops.viewInfo(thing);
						str = "[\"" + tag + "\"," + flatten(info.buffer);
						if (info.byteLength !== info.bufferByteLength) str += `,${info.byteOffset},${info.length}`;
						str += "]";
						break;
					}
					case "DataView": {
						const info = ops.viewInfo(thing);
						str = "[\"" + tag + "\"," + flatten(info.buffer);
						if (info.byteLength !== info.bufferByteLength) str += `,${info.byteOffset},${info.byteLength}`;
						str += "]";
						break;
					}
					case "ArrayBuffer":
						str = `["ArrayBuffer","${encode64(ops.toArrayBuffer(thing))}"]`;
						break;
					case "Temporal.Duration":
					case "Temporal.Instant":
					case "Temporal.PlainDate":
					case "Temporal.PlainTime":
					case "Temporal.PlainDateTime":
					case "Temporal.PlainMonthDay":
					case "Temporal.PlainYearMonth":
					case "Temporal.ZonedDateTime":
						str = `["${tag}",${stringify_string(ops.toStringValue(thing))}]`;
						break;
					default: {
						const shape = ops.shapeOf(thing);
						if (shape.kind === "not-plain") throw new DevalueError(`Cannot stringify arbitrary non-POJOs`, keys, thing, value);
						if (shape.kind === "symbol-keys") throw new DevalueError(`Cannot stringify POJOs with symbolic keys`, keys, thing, value);
						if (shape.kind === "null-proto") {
							str = "[\"null\"";
							for (const key of shape.keys) {
								if (key === "__proto__") throw new DevalueError(`Cannot stringify objects with __proto__ keys`, keys, thing, value);
								keys.push(stringify_key(key));
								str += `,${stringify_string(key)},${flatten(ops.get(thing, key))}`;
								keys.pop();
							}
							str += "]";
						} else {
							str = "{";
							let started = false;
							for (const key of shape.keys) {
								if (key === "__proto__") throw new DevalueError(`Cannot stringify objects with __proto__ keys`, keys, thing, value);
								if (started) str += ",";
								started = true;
								keys.push(stringify_key(key));
								str += `${stringify_string(key)}:${flatten(ops.get(thing, key))}`;
								keys.pop();
							}
							str += "}";
						}
					}
				}
			}
			stringified[index] = str;
			return index;
		}
		const index = flatten(value);
		if (index < 0) return `${index}`;
		return stringified;
	}
	/**
	* @param {any} thing
	* @returns {string}
	*/
	function stringify_primitive(thing) {
		const type = typeof thing;
		if (type === "string") return stringify_string(thing);
		if (thing === void 0) return (-1).toString();
		if (thing === 0 && 1 / thing < 0) return (-6).toString();
		if (type === "bigint") return `["BigInt","${thing}"]`;
		return String(thing);
	}
	/** @import { Component } from 'svelte' */
	/** @import { HydratableContext, SSRContext } from './types.js' */
	/** @import { Csp, RenderOutput, SyncRenderOutput, Sha256Source } from '../../server/public.js' */
	/** @import { MaybePromise } from '#shared' */
	/** @typedef {'head' | 'body'} RendererType */
	/** @typedef {{ [key in RendererType]: string }} AccumulatedContent */
	/**
	* @typedef {string | Renderer} RendererItem
	*/
	/**
	* Renderers are basically a tree of `string | Renderer`s, where each `Renderer` in the tree represents
	* work that may or may not have completed. A renderer can be {@link collect}ed to aggregate the
	* content from itself and all of its children, but this will throw if any of the children are
	* performing asynchronous work. To asynchronously collect a renderer, just `await` it.
	*
	* The `string` values within a renderer are always associated with the {@link type} of that renderer. To switch types,
	* call {@link child} with a different `type` argument.
	*/
	var Renderer = class Renderer {
		/**
		* The contents of the renderer.
		* @type {RendererItem[]}
		*/
		#out = [];
		/**
		* Any `onDestroy` callbacks registered during execution of this renderer.
		* @type {(() => void)[] | undefined}
		*/
		#on_destroy = void 0;
		/**
		* Whether this renderer is a component body.
		* @type {boolean}
		*/
		#is_component_body = false;
		/**
		* If set, this renderer is an error boundary. When async collection
		* of the children fails, the failed snippet is rendered instead.
		* @type {{
		* 	failed: (renderer: Renderer, error: unknown, reset: () => void) => void;
		* 	transformError: (error: unknown) => unknown;
		* 	context: SSRContext | null;
		* } | null}
		*/
		#boundary = null;
		/**
		* The type of string content that this renderer is accumulating.
		* @type {RendererType}
		*/
		type;
		/** @type {Renderer | undefined} */
		#parent;
		/**
		* Asynchronous work associated with this renderer
		* @type {Promise<void> | undefined}
		*/
		promise = void 0;
		/**
		* State which is associated with the content tree as a whole.
		* It will be re-exposed, uncopied, on all children.
		* @type {SSRState}
		* @readonly
		*/
		global;
		/**
		* State that is local to the branch it is declared in.
		* It will be shallow-copied to all children.
		*
		* @type {{ select_value: any, multiple: boolean }}
		*/
		local;
		/**
		* @param {SSRState} global
		* @param {Renderer | undefined} [parent]
		*/
		constructor(global, parent) {
			this.#parent = parent;
			this.global = global;
			this.local = parent ? { ...parent.local } : {
				select_value: void 0,
				multiple: false
			};
			this.type = parent ? parent.type : "body";
		}
		/**
		* @param {(renderer: Renderer) => void} fn
		*/
		head(fn) {
			const head = new Renderer(this.global, this);
			head.type = "head";
			this.#out.push(head);
			head.child(fn);
		}
		/**
		* @param {Array<Promise<void>>} blockers
		* @param {(renderer: Renderer) => void} fn
		*/
		async_block(blockers, fn) {
			this.#out.push(BLOCK_OPEN);
			this.async(blockers, fn);
			this.#out.push(BLOCK_CLOSE);
		}
		/**
		* @param {Array<Promise<void>>} blockers
		* @param {(renderer: Renderer) => void} fn
		*/
		async(blockers, fn) {
			let callback = fn;
			if (blockers.length > 0) {
				const context = ssr_context;
				callback = (renderer) => {
					return Promise.all(blockers).then(() => {
						const previous_context = ssr_context;
						try {
							set_ssr_context(context);
							return fn(renderer);
						} finally {
							set_ssr_context(previous_context);
						}
					});
				};
			}
			this.child(callback);
		}
		/**
		* @param {Array<() => void>} thunks
		*/
		run(thunks) {
			const context = ssr_context;
			let promise = Promise.resolve(thunks[0]());
			const promises = [promise];
			for (const fn of thunks.slice(1)) {
				promise = promise.then(() => {
					const previous_context = ssr_context;
					set_ssr_context(context);
					try {
						return fn();
					} finally {
						set_ssr_context(previous_context);
					}
				});
				promises.push(promise);
			}
			promise.catch(noop$1);
			this.promise = this.global.track(promise);
			return promises;
		}
		/**
		* @param {(renderer: Renderer) => MaybePromise<void>} fn
		*/
		child_block(fn) {
			this.#out.push(BLOCK_OPEN);
			this.child(fn);
			this.#out.push(BLOCK_CLOSE);
		}
		/**
		* Create a child renderer. The child renderer inherits the state from the parent,
		* but has its own content.
		* @param {(renderer: Renderer) => MaybePromise<void>} fn
		*/
		child(fn) {
			const child = new Renderer(this.global, this);
			this.#out.push(child);
			const parent = ssr_context;
			set_ssr_context({
				...ssr_context,
				p: parent,
				c: null,
				r: child
			});
			const result = fn(child);
			set_ssr_context(parent);
			if (result instanceof Promise) {
				result.catch(noop$1);
				result.finally(() => set_ssr_context(null)).catch(noop$1);
				if (child.global.mode === "sync") await_invalid();
				child.promise = child.global.track(result);
			}
			return child;
		}
		/**
		* Render children inside an error boundary. If the children throw and the API-level
		* `transformError` transform handles the error (doesn't re-throw), the `failed` snippet is
		* rendered instead. Otherwise the error propagates.
		*
		* @param {{ failed?: (renderer: Renderer, error: unknown, reset: () => void) => void }} props
		* @param {(renderer: Renderer) => MaybePromise<void>} children_fn
		*/
		boundary(props, children_fn) {
			const child = new Renderer(this.global, this);
			this.#out.push(child);
			const parent_context = ssr_context;
			if (props.failed) child.#boundary = {
				failed: props.failed,
				transformError: this.global.transformError,
				context: parent_context
			};
			set_ssr_context({
				...ssr_context,
				p: parent_context,
				c: null,
				r: child
			});
			try {
				const result = children_fn(child);
				set_ssr_context(parent_context);
				if (result instanceof Promise) {
					if (child.global.mode === "sync") await_invalid();
					result.catch(noop$1);
					child.promise = child.global.track(result);
				}
			} catch (error) {
				set_ssr_context(parent_context);
				const failed_snippet = props.failed;
				if (!failed_snippet) throw error;
				const result = this.global.transformError(error);
				child.#out.length = 0;
				child.#boundary = null;
				if (result instanceof Promise) {
					if (this.global.mode === "sync") await_invalid();
					child.promise = child.global.track(
						/** @type {Promise<unknown>} */
						result.then((transformed) => {
							set_ssr_context(parent_context);
							child.#out.push(Renderer.#serialize_failed_boundary(transformed));
							failed_snippet(child, transformed, noop$1);
							child.#out.push(BLOCK_CLOSE);
						})
					);
					child.promise.catch(noop$1);
				} else {
					child.#out.push(Renderer.#serialize_failed_boundary(result));
					failed_snippet(child, result, noop$1);
					child.#out.push(BLOCK_CLOSE);
				}
			}
		}
		/**
		* Create a component renderer. The component renderer inherits the state from the parent,
		* but has its own content. It is treated as an ordering boundary for ondestroy callbacks.
		* @param {(renderer: Renderer) => MaybePromise<void>} fn
		* @param {Function} [component_fn]
		* @returns {void}
		*/
		component(fn, component_fn) {
			push(component_fn);
			this.child((renderer) => {
				renderer.#is_component_body = true;
				return fn(renderer);
			});
			pop();
		}
		/**
		* @param {Record<string, any>} attrs
		* @param {(renderer: Renderer) => void} fn
		* @param {string | undefined} [css_hash]
		* @param {Record<string, boolean> | undefined} [classes]
		* @param {Record<string, string> | undefined} [styles]
		* @param {number | undefined} [flags]
		* @param {boolean | undefined} [is_rich]
		* @returns {void}
		*/
		select(attrs, fn, css_hash, classes, styles, flags, is_rich) {
			const { value, defaultValue, ...select_attrs } = attrs;
			if (select_attrs.multiple === "") select_attrs.multiple = true;
			this.push(`<select${attributes(select_attrs, css_hash, classes, styles, flags)}>`);
			this.child((renderer) => {
				renderer.local.select_value = value === void 0 ? defaultValue : value;
				renderer.local.multiple = !!select_attrs.multiple;
				fn(renderer);
			});
			this.push(`${is_rich ? "<!>" : ""}</select>`);
		}
		/**
		* @param {Record<string, any>} attrs
		* @param {string | number | boolean | ((renderer: Renderer) => void)} body
		* @param {string | undefined} [css_hash]
		* @param {Record<string, boolean> | undefined} [classes]
		* @param {Record<string, string> | undefined} [styles]
		* @param {number | undefined} [flags]
		* @param {boolean | undefined} [is_rich]
		*/
		option(attrs, body, css_hash, classes, styles, flags, is_rich) {
			this.#out.push(`<option${attributes(attrs, css_hash, classes, styles, flags)}`);
			/**
			* @param {Renderer} renderer
			* @param {any} value
			* @param {{ head?: string, body: any }} content
			*/
			const close = (renderer, value, { head, body }) => {
				if (has_own_property.call(attrs, "value")) value = attrs.value;
				var select_value = this.local.select_value;
				if (this.local.multiple && is_array(select_value) ? select_value.includes(value) : value === select_value) renderer.#out.push(" selected=\"\"");
				renderer.#out.push(`>${body}${is_rich ? "<!>" : ""}</option>`);
				if (head) renderer.head((child) => child.push(head));
			};
			if (typeof body === "function") this.child((renderer) => {
				const r = new Renderer(this.global, this);
				body(r);
				if (this.global.mode === "async") return r.#collect_content_async().then((content) => {
					close(renderer, content.body.replaceAll("<!---->", ""), content);
				});
				else {
					const content = r.#collect_content();
					close(renderer, content.body.replaceAll("<!---->", ""), content);
				}
			});
			else close(this, body, { body: escape_html(body) });
		}
		/**
		* @param {(renderer: Renderer) => void} fn
		*/
		title(fn) {
			const path = this.get_path();
			/** @param {string} head */
			const close = (head) => {
				this.global.set_title(head, path);
			};
			this.child((renderer) => {
				const r = new Renderer(renderer.global, renderer);
				fn(r);
				if (renderer.global.mode === "async") return r.#collect_content_async().then((content) => {
					close(content.head);
				});
				else {
					const content = r.#collect_content();
					close(content.head);
				}
			});
		}
		/**
		* @param {string | (() => Promise<string>)} content
		*/
		push(content) {
			if (typeof content === "function") this.child(async (renderer) => renderer.push(await content()));
			else this.#out.push(content);
		}
		/**
		* @param {() => void} fn
		*/
		on_destroy(fn) {
			(this.#on_destroy ??= []).push(fn);
		}
		/**
		* @returns {number[]}
		*/
		get_path() {
			return this.#parent ? [...this.#parent.get_path(), this.#parent.#out.indexOf(this)] : [];
		}
		/**
		* @deprecated this is needed for legacy component bindings
		*/
		copy() {
			const copy = new Renderer(this.global, this.#parent);
			copy.type = this.type;
			copy.#out = this.#out.map((item) => item instanceof Renderer ? item.copy() : item);
			copy.promise = this.promise;
			return copy;
		}
		/**
		* @param {Renderer} other
		* @deprecated this is needed for legacy component bindings
		*/
		subsume(other) {
			if (this.global.mode !== other.global.mode) throw new Error("invariant: A renderer cannot switch modes. If you're seeing this, there's a compiler bug. File an issue!");
			this.local = other.local;
			this.#out = other.#out.map((item, i) => {
				const current = this.#out[i];
				if (current instanceof Renderer && item instanceof Renderer) {
					current.subsume(item);
					return current;
				}
				return item;
			});
			this.promise = other.promise;
			this.type = other.type;
		}
		get length() {
			return this.#out.length;
		}
		/**
		* Creates the hydration comment that marks the start of a failed boundary.
		* The error is JSON-serialized and embedded inside an HTML comment for the client
		* to parse during hydration. The JSON is escaped to prevent `-->` or `<!--` sequences
		* from breaking out of the comment (XSS). Uses unicode escapes which `JSON.parse()`
		* handles transparently.
		* @param {unknown} error
		* @returns {string}
		*/
		static #serialize_failed_boundary(error) {
			return `<!--[?${JSON.stringify(error).replace(/>/g, "\\u003e").replace(/</g, "\\u003c")}-->`;
		}
		/**
		* Only available on the server and when compiling with the `server` option.
		* Takes a component and returns an object with `body` and `head` properties on it, which you can use to populate the HTML when server-rendering your app.
		* @template {Record<string, any>} Props
		* @param {Component<Props>} component
		* @param {{ props?: Omit<Props, '$$slots' | '$$events'>; context?: Map<any, any>; idPrefix?: string; csp?: Csp }} [options]
		* @returns {RenderOutput}
		*/
		static render(component, options = {}) {
			/** @type {AccumulatedContent | undefined} */
			let sync;
			/** @type {Promise<AccumulatedContent & { hashes: { script: Sha256Source[] } }> | undefined} */
			let async;
			const result = {};
			Object.defineProperties(result, {
				html: { get: () => {
					return (sync ??= Renderer.#render(component, options)).body;
				} },
				head: { get: () => {
					return (sync ??= Renderer.#render(component, options)).head;
				} },
				body: { get: () => {
					return (sync ??= Renderer.#render(component, options)).body;
				} },
				hashes: { value: { script: "" } },
				then: { value: 
				/**
				* this is not type-safe, but honestly it's the best I can do right now, and it's a straightforward function.
				*
				* @template TResult1
				* @template [TResult2=never]
				* @param { (value: SyncRenderOutput) => TResult1 } onfulfilled
				* @param { (reason: unknown) => TResult2 } onrejected
				*/
				(onfulfilled, onrejected) => {
					if (!async_mode_flag) {
						const result = sync ??= Renderer.#render(component, options);
						const user_result = onfulfilled({
							head: result.head,
							body: result.body,
							html: result.body,
							hashes: { script: [] }
						});
						return Promise.resolve(user_result);
					}
					async ??= init_render_context().then(() => with_render_context(() => Renderer.#render_async(component, options)));
					return async.then((result) => {
						Object.defineProperty(result, "html", { get: () => {
							html_deprecated();
						} });
						return onfulfilled(result);
					}, onrejected);
				} }
			});
			return result;
		}
		/**
		* Collect all of the `onDestroy` callbacks registered during rendering. In an async context, this is only safe to call
		* after awaiting `collect_async`.
		*
		* Child renderers are "porous" and don't affect execution order, but component body renderers
		* create ordering boundaries. Within a renderer, callbacks run in order until hitting a component boundary.
		* @returns {Iterable<() => void>}
		*/
		*#collect_on_destroy() {
			for (const component of this.#traverse_components()) yield* component.#collect_ondestroy();
		}
		/**
		* Performs a depth-first search of renderers, yielding the deepest components first, then additional components as we backtrack up the tree.
		* @returns {Iterable<Renderer>}
		*/
		*#traverse_components() {
			for (const child of this.#out) if (typeof child !== "string") yield* child.#traverse_components();
			if (this.#is_component_body) yield this;
		}
		/**
		* @returns {Iterable<() => void>}
		*/
		*#collect_ondestroy() {
			if (this.#on_destroy) for (const fn of this.#on_destroy) yield fn;
			for (const child of this.#out) if (child instanceof Renderer && !child.#is_component_body) yield* child.#collect_ondestroy();
		}
		/**
		* Runs every `onDestroy` callback in this renderer tree. On a failed render,
		* cleanup errors are suppressed so they do not mask the render error.
		* @param {boolean} suppress_errors
		*/
		#run_on_destroy(suppress_errors) {
			let first_error;
			let has_error = false;
			for (const cleanup of this.#collect_on_destroy()) try {
				cleanup();
			} catch (error) {
				if (!suppress_errors && !has_error) {
					first_error = error;
					has_error = true;
				}
			}
			if (has_error) throw first_error;
		}
		/**
		* @param {'sync' | 'async'} mode
		* @param {{ idPrefix?: string; csp?: Csp; transformError?: (error: unknown) => unknown }} options
		* @returns {Renderer}
		*/
		static #create(mode, options) {
			if (options.idPrefix?.includes("--")) invalid_id_prefix();
			return new Renderer(new SSRState(mode, options.idPrefix ? options.idPrefix + "-" : "", options.csp, options.transformError));
		}
		/**
		* Render a component. Throws if any of the children are performing asynchronous work.
		*
		* @template {Record<string, any>} Props
		* @param {Component<Props>} component
		* @param {{ props?: Omit<Props, '$$slots' | '$$events'>; context?: Map<any, any>; idPrefix?: string }} options
		* @returns {AccumulatedContent}
		*/
		static #render(component, options) {
			var previous_context = ssr_context;
			const renderer = Renderer.#create("sync", options);
			/** @type {AccumulatedContent | undefined} */
			let result;
			let render_error;
			let failed = false;
			try {
				try {
					Renderer.#open_render(renderer, component, options);
					result = Renderer.#close_render(renderer.#collect_content(), renderer);
				} catch (error) {
					render_error = error;
					failed = true;
				}
				renderer.#run_on_destroy(failed);
				if (failed) throw render_error;
				return result;
			} finally {
				renderer.global.abort();
				set_ssr_context(previous_context);
			}
		}
		/**
		* Render a component.
		*
		* @template {Record<string, any>} Props
		* @param {Component<Props>} component
		* @param {{ props?: Omit<Props, '$$slots' | '$$events'>; context?: Map<any, any>; idPrefix?: string; csp?: Csp }} options
		* @returns {Promise<AccumulatedContent & { hashes: { script: Sha256Source[] } }>}
		*/
		static async #render_async(component, options) {
			const previous_context = ssr_context;
			const renderer = Renderer.#create("async", options);
			/** @type {(AccumulatedContent & { hashes: { script: Sha256Source[] } }) | undefined} */
			let result;
			let render_error;
			let failed = false;
			try {
				try {
					Renderer.#open_render(renderer, component, options);
					const content = await renderer.#collect_content_async();
					const hydratables = await renderer.#collect_hydratables();
					if (hydratables !== null) content.head = hydratables + content.head;
					result = Renderer.#close_render(content, renderer);
				} catch (error) {
					render_error = error;
					failed = true;
					renderer.global.abort();
					await renderer.global.settle();
				}
				renderer.#run_on_destroy(failed);
				if (failed) throw render_error;
				return result;
			} finally {
				set_ssr_context(previous_context);
				renderer.global.abort();
			}
		}
		/**
		* Collect all of the code from the `out` array and return it as a string, or a promise resolving to a string.
		* @param {AccumulatedContent} content
		* @returns {AccumulatedContent}
		*/
		#collect_content(content = {
			head: "",
			body: ""
		}) {
			for (const item of this.#out) if (typeof item === "string") content[this.type] += item;
			else if (item instanceof Renderer) item.#collect_content(content);
			return content;
		}
		/**
		* Collect all of the code from the `out` array and return it as a string.
		* @param {AccumulatedContent} content
		* @returns {Promise<AccumulatedContent>}
		*/
		async #collect_content_async(content = {
			head: "",
			body: ""
		}) {
			await this.promise;
			for (const item of this.#out) if (typeof item === "string") content[this.type] += item;
			else if (item instanceof Renderer) {
				if (item.#boundary) {
					/** @type {AccumulatedContent} */
					const boundary_content = {
						head: "",
						body: ""
					};
					try {
						await item.#collect_content_async(boundary_content);
						content.head += boundary_content.head;
						content.body += boundary_content.body;
					} catch (error) {
						const { context, failed, transformError } = item.#boundary;
						set_ssr_context(context);
						let promise = transformError(error);
						set_ssr_context(null);
						let transformed = await promise;
						set_ssr_context(context);
						const failed_renderer = new Renderer(item.global, item);
						failed_renderer.type = item.type;
						failed_renderer.#out.push(Renderer.#serialize_failed_boundary(transformed));
						failed(failed_renderer, transformed, noop$1);
						failed_renderer.#out.push(BLOCK_CLOSE);
						await failed_renderer.#collect_content_async(content);
					}
				} else await item.#collect_content_async(content);
			}
			return content;
		}
		async #collect_hydratables() {
			const ctx = get_render_context().hydratable;
			for (const [_, key] of ctx.unresolved_promises) unresolved_hydratable(key, ctx.lookup.get(key)?.stack ?? "<missing stack trace>");
			for (const comparison of ctx.comparisons) await comparison;
			return await this.#hydratable_block(ctx);
		}
		/**
		* @template {Record<string, any>} Props
		* @param {Renderer} renderer
		* @param {import('svelte').Component<Props>} component
		* @param {{ props?: Omit<Props, '$$slots' | '$$events'>; context?: Map<any, any>; idPrefix?: string; csp?: Csp; transformError?: (error: unknown) => unknown }} options
		* @returns {void}
		*/
		static #open_render(renderer, component, options) {
			var previous_context = ssr_context;
			try {
				set_ssr_context({
					p: null,
					c: options.context ?? null,
					r: renderer
				});
				renderer.push(BLOCK_OPEN);
				component(renderer, options.props ?? {});
				renderer.push(BLOCK_CLOSE);
			} finally {
				set_ssr_context(previous_context);
			}
		}
		/**
		* @param {AccumulatedContent} content
		* @param {Renderer} renderer
		* @returns {AccumulatedContent & { hashes: { script: Sha256Source[] } }}
		*/
		static #close_render(content, renderer) {
			let head = content.head + renderer.global.get_title();
			let body = content.body;
			for (const { hash, code } of renderer.global.css) head += `<style id="${hash}">${code}</style>`;
			return {
				head,
				body,
				hashes: { script: renderer.global.csp.script_hashes }
			};
		}
		/**
		* @param {HydratableContext} ctx
		*/
		async #hydratable_block(ctx) {
			if (ctx.lookup.size === 0) return null;
			let entries = [];
			let has_promises = false;
			for (const [k, v] of ctx.lookup) {
				if (v.promises) {
					has_promises = true;
					for (const p of v.promises) await p;
				}
				entries.push(`[${uneval$1(k)},${v.serialized}]`);
			}
			let prelude = `const h = (window.__svelte ??= {}).h ??= new Map();`;
			if (has_promises) prelude = `const r = (v) => Promise.resolve(v);
				${prelude}`;
			const body = `
			{
				${prelude}

				for (const [k, v] of [
					${entries.join(",\n					")}
				]) {
					h.set(k, v);
				}
			}
		`;
			let csp_attr = "";
			if (this.global.csp.nonce) csp_attr = ` nonce="${this.global.csp.nonce}"`;
			else if (this.global.csp.hash) {
				const hash = await sha256(body);
				this.global.csp.script_hashes.push(`sha256-${hash}`);
			}
			return `\n\t\t<script${csp_attr}>${body}<\/script>`;
		}
	};
	var SSRState = class {
		/** @readonly @type {Csp & { script_hashes: Sha256Source[] }} */
		csp;
		/** @readonly @type {'sync' | 'async'} */
		mode;
		/** @readonly @type {() => string} */
		uid;
		/** @readonly @type {Set<{ hash: string; code: string }>} */
		css = /* @__PURE__ */ new Set();
		/** @type {Set<Promise<unknown>>} */
		#pending = /* @__PURE__ */ new Set();
		/** @type {AbortController | null} */
		#controller = null;
		#aborted = false;
		/**
		* `transformError` passed to `render`. Called when an error boundary catches an error.
		* Throws by default if unset in `render`.
		* @type {(error: unknown) => unknown}
		*/
		transformError;
		/** @type {{ path: number[], value: string }} */
		#title = {
			path: [],
			value: ""
		};
		/**
		* @param {'sync' | 'async'} mode
		* @param {string} id_prefix
		* @param {Csp} csp
		* @param {((error: unknown) => unknown) | undefined} [transformError]
		*/
		constructor(mode, id_prefix = "", csp = { hash: false }, transformError) {
			this.mode = mode;
			this.csp = {
				...csp,
				script_hashes: []
			};
			this.transformError = transformError ?? ((error) => {
				throw error;
			});
			let uid = 1;
			this.uid = () => `${id_prefix}s${uid++}`;
		}
		/**
		* @template T
		* @param {Promise<T>} promise
		* @returns {Promise<T>}
		*/
		track(promise) {
			this.#pending.add(promise);
			promise.then(() => this.#pending.delete(promise), () => this.#pending.delete(promise));
			return promise;
		}
		async settle() {
			while (this.#pending.size > 0) await Promise.allSettled([...this.#pending]);
		}
		abort() {
			if (this.#aborted) return;
			this.#aborted = true;
			this.#controller?.abort(STALE_REACTION);
		}
		get_abort_signal() {
			const controller = this.#controller ??= new AbortController();
			if (this.#aborted) controller.abort(STALE_REACTION);
			return controller.signal;
		}
		get_title() {
			return this.#title.value;
		}
		/**
		* Performs a depth-first (lexicographic) comparison using the path. Rejects sets
		* from earlier than or equal to the current value.
		* @param {string} value
		* @param {number[]} path
		*/
		set_title(value, path) {
			const current = this.#title.path;
			let i = 0;
			let l = Math.min(path.length, current.length);
			while (i < l && path[i] === current[i]) i += 1;
			if (path[i] === void 0) return;
			if (current[i] === void 0 || path[i] > current[i]) {
				this.#title.path = path;
				this.#title.value = value;
			}
		}
	};
	var INVALID_ATTR_NAME_CHAR_REGEX = /[\s'">/=\u{FDD0}-\u{FDEF}\u{FFFE}\u{FFFF}\u{1FFFE}\u{1FFFF}\u{2FFFE}\u{2FFFF}\u{3FFFE}\u{3FFFF}\u{4FFFE}\u{4FFFF}\u{5FFFE}\u{5FFFF}\u{6FFFE}\u{6FFFF}\u{7FFFE}\u{7FFFF}\u{8FFFE}\u{8FFFF}\u{9FFFE}\u{9FFFF}\u{AFFFE}\u{AFFFF}\u{BFFFE}\u{BFFFF}\u{CFFFE}\u{CFFFF}\u{DFFFE}\u{DFFFF}\u{EFFFE}\u{EFFFF}\u{FFFFE}\u{FFFFF}\u{10FFFE}\u{10FFFF}]/u;
	/**
	* Only available on the server and when compiling with the `server` option.
	* Takes a component and returns an object with `body` and `head` properties on it, which you can use to populate the HTML when server-rendering your app.
	* @template {Record<string, any>} Props
	* @param {Component<Props> | ComponentType<SvelteComponent<Props>>} component
	* @param {{ props?: Omit<Props, '$$slots' | '$$events'>; context?: Map<any, any>; idPrefix?: string; csp?: Csp; transformError?: (error: unknown) => unknown }} [options]
	* @returns {RenderOutput}
	*/
	function render(component, options = {}) {
		if (options.csp?.hash && options.csp.nonce) invalid_csp();
		return Renderer.render(component, options);
	}
	/**
	* @param {Record<string, unknown>} attrs
	* @param {string} [css_hash]
	* @param {Record<string, boolean>} [classes]
	* @param {Record<string, string>} [styles]
	* @param {number} [flags]
	* @returns {string}
	*/
	function attributes(attrs, css_hash, classes, styles, flags = 0) {
		if (styles) attrs.style = to_style(attrs.style, styles);
		if (attrs.class) attrs.class = clsx(attrs.class);
		if (css_hash || classes) attrs.class = to_class(attrs.class, css_hash, classes);
		let attr_str = "";
		let name;
		const is_html = (flags & 1) === 0;
		const lowercase = (flags & 2) === 0;
		const is_input = (flags & 4) !== 0;
		for (name of Object.keys(attrs)) {
			if (typeof attrs[name] === "function") continue;
			if (name[0] === "$" && name[1] === "$") continue;
			if (name === "" || INVALID_ATTR_NAME_CHAR_REGEX.test(name)) continue;
			var value = attrs[name];
			var lower = name.toLowerCase();
			if (lowercase) name = lower;
			if (lower.length > 2 && lower.startsWith("on")) continue;
			if (is_input) {
				if (name === "defaultvalue" || name === "defaultchecked") {
					name = name === "defaultvalue" ? "value" : "checked";
					if (attrs[name]) continue;
				}
			}
			attr_str += attr(name, value, is_html && is_boolean_attribute(name));
		}
		return attr_str;
	}
	/**
	* @template V
	* @param {() => V} get_value
	*/
	function once(get_value) {
		let value = UNINITIALIZED;
		return () => {
			if (value === UNINITIALIZED) value = get_value();
			return value;
		};
	}
	/**
	* @template T
	* @param {()=>T} fn
	* @returns {(new_value?: T) => (T | void)}
	*/
	function derived(fn) {
		const get_value = ssr_context === null ? fn : once(fn);
		/** @type {T | undefined} */
		let updated_value;
		return function(new_value) {
			if (arguments.length === 0) return updated_value ?? get_value();
			updated_value = new_value;
			return updated_value;
		};
	}
	enable_async_mode_flag();
	function noop() {}
	var afterNavigate = noop;
	function Root($$renderer, $$props) {
		$$renderer.component(($$renderer) => {
			const { page, components, onerror, tree, form, error } = $$props;
			let mounted = false;
			let navigated = false;
			let title = "";
			afterNavigate(() => {
				if (mounted) {
					navigated = true;
					title = document.title || "untitled page";
				} else mounted = true;
			});
			function node($$renderer, n, depth) {
				const Component = derived(() => n.component);
				const Error = derived(() => n.error);
				const data = derived(() => n.data);
				function failed($$renderer, error) {
					if (Error()) {
						$$renderer.push("<!--[-->");
						Error()($$renderer, { error });
						$$renderer.push("<!--]-->");
					} else {
						$$renderer.push("<!--[!-->");
						$$renderer.push("<!--]-->");
					}
				}
				$$renderer.boundary({ failed: Error() ? failed : void 0 }, ($$renderer) => {
					$$renderer.push(`<!--[-->`);
					if (n.child) {
						$$renderer.push("<!--[0-->");
						if (Component()) {
							$$renderer.push("<!--[-->");
							Component()($$renderer, {
								data: data(),
								form,
								params: page.params,
								children: ($$renderer) => {
									node($$renderer, n.child, depth + 1);
								},
								$$slots: { default: true }
							});
							$$renderer.push("<!--]-->");
						} else {
							$$renderer.push("<!--[!-->");
							$$renderer.push("<!--]-->");
						}
					} else {
						$$renderer.push("<!--[-1-->");
						if (Component()) {
							$$renderer.push("<!--[-->");
							Component()($$renderer, {
								data: data(),
								form,
								params: page.params,
								error
							});
							$$renderer.push("<!--]-->");
						} else {
							$$renderer.push("<!--[!-->");
							$$renderer.push("<!--]-->");
						}
					}
					$$renderer.push(`<!--]-->`);
					$$renderer.push(`<!--]-->`);
				});
			}
			node($$renderer, tree, 0);
			$$renderer.push(`<!----> `);
			if (mounted) {
				$$renderer.push(`<!--[0--><div id="svelte-announcer" aria-live="assertive" aria-atomic="true" style="position: absolute; left: 0; top: 0; clip: rect(0 0 0 0); clip-path: inset(50%); overflow: hidden; white-space: nowrap; width: 1px; height: 1px">`);
				if (navigated) $$renderer.push(`<!--[0-->${escape_html(title)}`);
				else $$renderer.push("<!--[-1-->");
				$$renderer.push(`<!--]--></div>`);
			} else $$renderer.push("<!--[-1-->");
			$$renderer.push(`<!--]-->`);
		});
	}
	var Props = class {
		/** @type {Page} */
		page;
		/**
		* An array of the `+layout.svelte` and `+page.svelte` component instances
		* that currently live on the page — used for capturing and restoring snapshots.
		* It's updated/manipulated through `bind:this` in `Root.svelte`.
		* @type {Array<Record<string, any>>}
		* @deprecated only used for `export const snapshot` — TODO 4.0 get rid
		*/
		components = [];
		/** @type {any} */
		form;
		/** @type {App.Error | undefined} */
		error;
		/** @type {RenderNode} */
		tree;
		/** @type {(error: unknown, reset: () => void) => void} */
		onerror;
		/**
		* @param {{
		*   page: Page;
		*   tree: RenderNode;
		*   form: any;
		*   error: App.Error | undefined;
		*   onerror?: (error: unknown, reset: () => void) => void;
		* }} props
		*/
		constructor({ page, tree, form, error, onerror = noop }) {
			this.page = page;
			this.tree = tree;
			this.onerror = onerror;
			this.form = form;
			this.error = error;
		}
	};
	var RenderNode = class {
		/** @type {Component} */
		component;
		/** @type {Component | undefined} */
		error;
		/** @type {Record<string, any>} */
		data = {};
		/** @type {RenderNode | undefined} */
		child;
		/**
		*
		* @param {Component} component
		* @param {Component | undefined} error
		*/
		constructor(component, error) {
			this.component = component;
			this.error = error;
		}
	};
	var MUTATIVE_METHODS = [
		"POST",
		"PUT",
		"PATCH",
		"DELETE"
	];
	[...MUTATIVE_METHODS];
	({}).dirname;
	var IN_WEBCONTAINER = !!globalThis.process?.versions?.webcontainer;
	/** @import { RequestEvent } from '@sveltejs/kit' */
	/** @import { RequestStore } from 'types' */
	/** @import { AsyncLocalStorage } from 'node:async_hooks' */
	/** @type {RequestStore | null} */
	var sync_store = null;
	/** @type {AsyncLocalStorage<RequestStore | null> | null} */
	var als;
	Promise.resolve().then(() => /* @__PURE__ */ __toESM(require__skgo_skgo_missing(), 1)).then((hooks) => als = new hooks.AsyncLocalStorage()).catch(() => {});
	function get_request_store() {
		const result = try_get_request_store();
		if (!result) {
			let message = "Could not get the request store.";
			if (als) message += " This is an internal error.";
			else message += " In environments without `AsyncLocalStorage`, the request store (used by e.g. remote functions) must be accessed synchronously, not after an `await`. If it was accessed synchronously then this is an internal error.";
			throw new Error(message);
		}
		return result;
	}
	function try_get_request_store() {
		return sync_store ?? als?.getStore() ?? null;
	}
	/**
	* @template T
	* @param {RequestStore | null} store
	* @param {() => T} fn
	*/
	function with_request_store(store, fn) {
		try {
			sync_store = store;
			return als ? als.run(store, fn) : fn();
		} finally {
			if (!IN_WEBCONTAINER) sync_store = null;
		}
	}
	/** @import { RemoteInternals } from 'types' */
	/** @type {RemoteInternals['type'][]} */
	var types = [
		"command",
		"form",
		"prerender",
		"query",
		"query_batch",
		"query_live"
	];
	/**
	* @param {Record<string, any>} module
	* @param {string} file
	* @param {string} hash
	*/
	function init_remote_functions(module, file, hash) {
		if (module.default) throw new Error(`Cannot export \`default\` from a remote module (${file}) — please use named exports instead`);
		for (const [name, fn] of Object.entries(module)) {
			if (!types.includes(fn?.__?.type)) throw new Error(`\`${name}\` exported from ${file} is invalid — all exports from this file must be remote functions`);
			fn.__.id = `${hash}/${name}`;
			fn.__.name = name;
		}
	}
	/** @import { StandardSchemaV1 } from '@standard-schema/spec' */
	var HttpError = class {
		/**
		* @param {App.Error} error
		*/
		constructor(error) {
			this.status = error.status;
			this.body = error;
		}
		toString() {
			return JSON.stringify(this.body);
		}
	};
	/**
	* An `HttpError` whose body is already in its final, user-facing form — either produced by the
	* `handleError` hook on the server and reconstructed here from the response, or authored directly
	* by the client runtime. Unlike a plain `HttpError` (which represents a fresh `error(...)` call
	* that the hook has yet to see), `handleError` must not run on it.
	* @extends HttpError
	*/
	var HandledHttpError = class extends HttpError {};
	var Redirect = class {
		/**
		* @param {300 | 301 | 302 | 303 | 304 | 305 | 306 | 307 | 308} status
		* @param {string} location
		*/
		constructor(status, location) {
			try {
				new Headers({ location });
			} catch {
				throw new Error(`Invalid redirect location ${JSON.stringify(location)}: this string contains characters that cannot be used in HTTP headers`);
			}
			this.status = status;
			this.location = location;
		}
	};
	/**
	* An error that was thrown from within the SvelteKit runtime that is not fatal and doesn't result in a 500, such as a 404.
	* `SvelteKitError` goes through `handleError`.
	* @extends Error
	*/
	var SvelteKitError = class extends Error {
		/**
		* @param {number} status
		* @param {string} text
		* @param {string} message
		*/
		constructor(status, text, message) {
			super(message);
			this.status = status;
			this.text = text;
		}
	};
	/**
	* Error thrown when form validation fails imperatively
	*/
	var ValidationError = class extends Error {
		/**
		* @param {StandardSchemaV1.Issue[]} issues
		*/
		constructor(issues) {
			super("Validation failed");
			this.name = "ValidationError";
			this.issues = issues;
		}
	};
	/** @type {(data: string) => any} */
	var parse = () => {
		throw new Error("");
	};
	/** @type {Record<string, (data: any) => any>} */
	var encoders = {};
	/** @type {Record<string, (data: any) => any>} */
	var decoders = {};
	/**
	*
	* @param {Transport} transport
	*/
	function init_transport(transport) {
		const transporters = Object.entries(transport);
		transporters.length;
		encoders = Object.fromEntries(transporters.map(([k, v]) => [k, v.encode]));
		decoders = Object.fromEntries(transporters.map(([k, v]) => [k, v.decode]));
		parse = (data) => parse$1(data, decoders);
	}
	function _layout($$renderer, $$props) {
		let { children } = $$props;
		$$renderer.push(`<nav class="svelte-12qhfyh"><a href="/" class="svelte-12qhfyh">Home</a> <a href="/about" class="svelte-12qhfyh">About</a></nav> <main class="svelte-12qhfyh">`);
		children($$renderer);
		$$renderer.push(`<!----></main>`);
	}
	function context() {
		return getContext("__request__");
	}
	var page = {
		get data() {
			return context().page.data;
		},
		get error() {
			return context().page.error;
		},
		get form() {
			return context().page.form;
		},
		get params() {
			return context().page.params;
		},
		get route() {
			return context().page.route;
		},
		get shallow() {
			return context().page.shallow;
		},
		get state() {
			return context().page.state;
		},
		get status() {
			return context().page.status;
		},
		get url() {
			return context().page.url;
		}
	};
	function _error($$renderer, $$props) {
		$$renderer.component(($$renderer) => {
			$$renderer.push(`<h1 data-testid="title">Error ${escape_html(page.status)}</h1> <p data-testid="error-message">${escape_html(page.error?.message)}</p>`);
		});
	}
	new URL("a://");
	new TextEncoder();
	/**
	* Throws an error with a HTTP status code and an optional message.
	* When called during request handling, this will cause SvelteKit to
	* return an error response; the error will be passed to `handleError` as an _expected_ error.
	* Make sure you're not catching the thrown error, which would prevent SvelteKit from handling it.
	* @param {number} status The [HTTP status code](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status#client_error_responses). Must be in the range 400-599.
	* @param {string} [message] The error message.
	* @overload
	* @param {{ status: number; message: string } extends App.Error ? number : never} status
	* @param {{ status: number; message: string } extends App.Error ? string : never} [message]
	* @return {never}
	* @throws {import('./public.js').HttpError} This error instructs SvelteKit to initiate HTTP error handling.
	* @throws {Error} If the provided status is invalid (not between 400 and 599).
	*/
	/**
	* Throws an error with a HTTP status code and an optional message.
	* When called during request handling, this will cause SvelteKit to
	* return an error response; the error will be passed to `handleError` as an _expected_ error.
	* Make sure you're not catching the thrown error, which would prevent SvelteKit from handling it.
	* @param {number} status The [HTTP status code](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status#client_error_responses). Must be in the range 400-599.
	* @param {string} message The error message.
	* @param {keyof Omit<App.Error, 'status' | 'message'> extends never ? never : Omit<App.Error, 'status' | 'message'>} properties Additional properties of the App.Error type.
	* @overload
	* @param {number} status
	* @param {string} message
	* @param {keyof Omit<App.Error, 'status' | 'message'> extends never ? never : Omit<App.Error, 'status' | 'message'>} properties
	* @return {never}
	* @throws {import('./public.js').HttpError} This error instructs SvelteKit to initiate HTTP error handling.
	* @throws {Error} If the provided status is invalid (not between 400 and 599).
	*/
	/**
	* Throws an error with a HTTP status code and an optional message.
	* When called during request handling, this will cause SvelteKit to
	* return an error response; the error will be passed to `handleError` as an _expected_ error.
	* Make sure you're not catching the thrown error, which would prevent SvelteKit from handling it.
	* @deprecated Passing an `App.Error` body as the second argument is deprecated — pass the `message` as the second argument, and any additional properties as the third
	* @param {number} status The [HTTP status code](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status#client_error_responses). Must be in the range 400-599.
	* @param {Omit<App.Error, 'status'> & { status?: App.Error['status'] }} body An object that conforms to the App.Error type. If a string is passed, it will be used as the message property.
	* @overload
	* @param {number} status
	* @param {Omit<App.Error, 'status'> & { status?: App.Error['status'] }} properties
	* @return {never}
	* @throws {import('./public.js').HttpError} This error instructs SvelteKit to initiate HTTP error handling.
	* @throws {Error} If the provided status is invalid (not between 400 and 599).
	*/
	/**
	* Throws an error with a HTTP status code and an optional message.
	* When called during request handling, this will cause SvelteKit to
	* return an error response; the error will be passed to `handleError` as an _expected_ error.
	* Make sure you're not catching the thrown error, which would prevent SvelteKit from handling it.
	* @param {any} status The [HTTP status code](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status#client_error_responses). Must be in the range 400-599.
	* @param {any} [message] A string, or (deprecated) a partial App.Error object
	* @param {any} [properties] Additional properties of the App.Error type when passing a string message.
	* @return {never}
	* @throws {import('./public.js').HttpError} This error instructs SvelteKit to initiate HTTP error handling.
	* @throws {Error} If the provided status is invalid (not between 400 and 599).
	*/
	function error(status, message, properties) {
		if (isNaN(status) || status < 400 || status > 599) throw new Error(`HTTP error status codes must be between 400 and 599 — ${status} is invalid`);
		if (message !== void 0 && typeof message !== "string") ({message, ...properties} = message);
		throw new HttpError({
			...properties,
			status,
			message: message ?? `Error: ${status}`
		});
	}
	function _typeof(o) {
		"@babel/helpers - typeof";
		return _typeof = "function" == typeof Symbol && "symbol" == typeof Symbol.iterator ? function(o) {
			return typeof o;
		} : function(o) {
			return o && "function" == typeof Symbol && o.constructor === Symbol && o !== Symbol.prototype ? "symbol" : typeof o;
		}, _typeof(o);
	}
	function toPrimitive(t, r) {
		if ("object" != _typeof(t) || !t) return t;
		var e = t[Symbol.toPrimitive];
		if (void 0 !== e) {
			var i = e.call(t, r || "default");
			if ("object" != _typeof(i)) return i;
			throw new TypeError("@@toPrimitive must return a primitive value.");
		}
		return ("string" === r ? String : Number)(t);
	}
	function toPropertyKey(t) {
		var i = toPrimitive(t, "string");
		return "symbol" == _typeof(i) ? i : i + "";
	}
	function _defineProperty(e, r, t) {
		return (r = toPropertyKey(r)) in e ? Object.defineProperty(e, r, {
			value: t,
			enumerable: !0,
			configurable: !0,
			writable: !0
		}) : e[r] = t, e;
	}
	function ownKeys(e, r) {
		var t = Object.keys(e);
		if (Object.getOwnPropertySymbols) {
			var o = Object.getOwnPropertySymbols(e);
			r && (o = o.filter(function(r) {
				return Object.getOwnPropertyDescriptor(e, r).enumerable;
			})), t.push.apply(t, o);
		}
		return t;
	}
	function _objectSpread2(e) {
		for (var r = 1; r < arguments.length; r++) {
			var t = null != arguments[r] ? arguments[r] : {};
			r % 2 ? ownKeys(Object(t), !0).forEach(function(r) {
				_defineProperty(e, r, t[r]);
			}) : Object.getOwnPropertyDescriptors ? Object.defineProperties(e, Object.getOwnPropertyDescriptors(t)) : ownKeys(Object(t)).forEach(function(r) {
				Object.defineProperty(e, r, Object.getOwnPropertyDescriptor(t, r));
			});
		}
		return e;
	}
	/** @import { RequestEvent } from '@sveltejs/kit' */
	/** @import { MaybePromise, RequestState, RemoteInternals, RequestStore, RemoteLiveQueryUserFunctionReturnType } from 'types' */
	/**
	* @param {any} validate_or_fn
	* @param {((arg?: any) => any) | undefined} [maybe_fn]
	* @returns {(arg?: any) => MaybePromise<any>}
	*/
	function create_validator(validate_or_fn, maybe_fn) {
		if (!maybe_fn) return (arg) => {
			if (arg !== void 0) error(400, "Bad Request");
		};
		if (validate_or_fn === "unchecked") return (arg) => arg;
		if ("~standard" in validate_or_fn) return async (arg) => {
			const result = await validate_or_fn["~standard"].validate(arg);
			if (result.issues) throw new ValidationError(result.issues);
			return result.value;
		};
		throw new Error("Invalid validator passed to remote function. Expected \"unchecked\" or a Standard Schema (https://standardschema.dev)");
	}
	/**
	* In case of a single remote function call, just returns the result.
	*
	* In case of a full page reload, returns the response for a remote function call,
	* either from the cache or by invoking the function.
	* Also saves an uneval'ed version of the result for later HTML inlining for hydration.
	*
	* @template {MaybePromise<any>} T
	* @param {RemoteInternals} internals
	* @param {string} payload — the stringified raw argument (i.e. the cache key the client will use)
	* @param {RequestState} state
	* @param {() => Promise<T>} get_result
	* @returns {Promise<T>}
	*/
	async function get_response(internals, payload, state, get_result) {
		var _cache$payload;
		await 0;
		const cache = get_cache(internals, state);
		if (!state.is_in_remote_query) get_implicit_lookup(internals, state)[payload] = get_result;
		return (_cache$payload = cache[payload]) !== null && _cache$payload !== void 0 ? _cache$payload : cache[payload] = get_result();
	}
	/**
	* @param {RequestEvent} event
	* @param {RequestState} state
	* @param {boolean} allow_cookies
	* @returns {RequestStore}
	*/
	function derive_remote_function_event(event, state, allow_cookies) {
		/** @type {RequestEvent} */
		const derived = _objectSpread2(_objectSpread2({}, event), {}, {
			setHeaders: () => {
				throw new Error("setHeaders is not allowed in remote functions");
			},
			cookies: _objectSpread2(_objectSpread2({}, event.cookies), {}, {
				set: (name, value, opts) => {
					if (!allow_cookies) throw new Error("Cannot set cookies in `query` or `prerender` functions");
					if (opts.path && !opts.path.startsWith("/")) throw new Error("Cookies set in remote functions must have an absolute path");
					return event.cookies.set(name, value, opts);
				},
				delete: (name, opts) => {
					if (!allow_cookies) throw new Error("Cannot delete cookies in `query` or `prerender` functions");
					if (opts.path && !opts.path.startsWith("/")) throw new Error("Cookies deleted in remote functions must have an absolute path");
					return event.cookies.delete(name, opts);
				}
			})
		});
		if (state.is_in_remote_query) for (const property of [
			"url",
			"params",
			"route"
		]) Object.defineProperty(derived, property, {
			enumerable: false,
			get() {
				throw new Error(`Cannot access event.${property} in a query. Pass the value as an argument to the query instead`);
			}
		});
		return {
			event: derived,
			state: _objectSpread2(_objectSpread2({}, state), {}, { is_in_remote_function: true })
		};
	}
	/**
	* Like `with_event` but removes things from `event` you cannot see/call in remote functions, such as `setHeaders`.
	* @template T
	* @param {RequestEvent} event
	* @param {RequestState} state
	* @param {boolean} allow_cookies
	* @param {() => any} get_input
	* @param {(arg?: any) => T} fn
	*/
	async function run_remote_function(event, state, allow_cookies, get_input, fn) {
		const store = derive_remote_function_event(event, state, allow_cookies);
		const input = await with_request_store(store, get_input);
		return with_request_store(store, () => fn(input));
	}
	/**
	* Like `with_event` but removes things from `event` you cannot see/call in remote functions, such as `setHeaders`.
	* @template T
	* @param {RequestEvent} event
	* @param {RequestState} state
	* @param {boolean} allow_cookies
	* @param {() => any} get_input
	* @param {(arg?: any) => RemoteLiveQueryUserFunctionReturnType<T>} fn
	* @param {string} name
	*/
	function run_remote_generator(_x, _x2, _x3, _x4, _x5, _x6) {
		return _run_remote_generator.apply(this, arguments);
	}
	function _run_remote_generator() {
		_run_remote_generator = _wrapAsyncGenerator(function* (event, state, allow_cookies, get_input, fn, name) {
			const store = derive_remote_function_event(event, state, allow_cookies);
			const input = yield _awaitAsyncGenerator(with_request_store(store, get_input));
			const iterator = to_iterator(yield _awaitAsyncGenerator(with_request_store(store, () => fn(input))), name);
			let done = false;
			try {
				while (true) {
					const result = yield _awaitAsyncGenerator(with_request_store(store, () => iterator.next()));
					if (result.done) {
						done = true;
						return result.value;
					}
					yield result.value;
				}
			} finally {
				if (!done && typeof iterator.return === "function") yield _awaitAsyncGenerator(with_request_store(store, () => {
					var _iterator$return;
					return (_iterator$return = iterator.return) === null || _iterator$return === void 0 ? void 0 : _iterator$return.call(iterator, void 0);
				}));
			}
		});
		return _run_remote_generator.apply(this, arguments);
	}
	/**
	* @template T
	* @param {Awaited<RemoteLiveQueryUserFunctionReturnType<T>>} source
	* @param {string} name
	* @returns {Iterator<T> | AsyncIterator<T>}
	*/
	function to_iterator(source, name) {
		if ("next" in source && typeof source.next === "function") return source;
		if (Symbol.asyncIterator in source && typeof source[Symbol.asyncIterator] === "function") return source[Symbol.asyncIterator]();
		if (Symbol.iterator in source && typeof source[Symbol.iterator] === "function") return source[Symbol.iterator]();
		throw new Error(`query.live '${name}' must return an Iterator, Iterable, AsyncIterator or AsyncIterable`);
	}
	/**
	* Note that `state` is deliberately not optional: resources that capture the request
	* state at creation must pass it explicitly, because reading it from the request store
	* at call time is only equivalent on runtimes with `AsyncLocalStorage` support.
	* Callers without a captured state (such as the module-level `form` instance getters)
	* should pass `get_request_store().state` themselves.
	* @param {RemoteInternals} internals
	* @param {RequestState} state
	*/
	function get_cache(internals, state) {
		var _state$remote$data;
		let cache = (_state$remote$data = state.remote.data) === null || _state$remote$data === void 0 ? void 0 : _state$remote$data.get(internals);
		if (cache === void 0) {
			var _state$remote, _state$remote$data2;
			cache = {};
			((_state$remote$data2 = (_state$remote = state.remote).data) !== null && _state$remote$data2 !== void 0 ? _state$remote$data2 : _state$remote.data = /* @__PURE__ */ new Map()).set(internals, cache);
		}
		return cache;
	}
	/**
	* @param {RemoteInternals} internals
	* @param {RequestState} state
	*/
	function get_implicit_lookup(internals, state) {
		var _state$remote$implici;
		let cache = (_state$remote$implici = state.remote.implicit) === null || _state$remote$implici === void 0 ? void 0 : _state$remote$implici.get(internals);
		if (cache === void 0) {
			var _state$remote2, _state$remote2$implic;
			cache = {};
			((_state$remote2$implic = (_state$remote2 = state.remote).implicit) !== null && _state$remote2$implic !== void 0 ? _state$remote2$implic : _state$remote2.implicit = /* @__PURE__ */ new Map()).set(internals, cache);
		}
		return cache;
	}
	/** @import { RemoteCommand } from '$app/server' */
	/** @import { MaybePromise, RemoteCommandInternals } from 'types' */
	/** @import { StandardSchemaV1 } from '@standard-schema/spec' */
	/**
	* Creates a remote command. When called from the browser, the function will be invoked on the server via a `fetch` call.
	*
	* See [Remote functions](https://svelte.dev/docs/kit/remote-functions#command) for full documentation.
	*
	* @template Output
	* @overload
	* @param {() => MaybePromise<Output>} fn
	* @returns {RemoteCommand<void, Output>}
	* @since 2.27
	*/
	/**
	* Creates a remote command. When called from the browser, the function will be invoked on the server via a `fetch` call.
	*
	* See [Remote functions](https://svelte.dev/docs/kit/remote-functions#command) for full documentation.
	*
	* @template Input
	* @template Output
	* @overload
	* @param {'unchecked'} validate
	* @param {(arg: Input) => MaybePromise<Output>} fn
	* @returns {RemoteCommand<Input, Output>}
	* @since 2.27
	*/
	/**
	* Creates a remote command. When called from the browser, the function will be invoked on the server via a `fetch` call.
	*
	* See [Remote functions](https://svelte.dev/docs/kit/remote-functions#command) for full documentation.
	*
	* @template {StandardSchemaV1} Schema
	* @template Output
	* @overload
	* @param {Schema} validate
	* @param {(arg: StandardSchemaV1.InferOutput<Schema>) => MaybePromise<Output>} fn
	* @returns {RemoteCommand<StandardSchemaV1.InferInput<Schema>, Output>}
	* @since 2.27
	*/
	/**
	* @template Input
	* @template Output
	* @param {any} validate_or_fn
	* @param {(arg?: Input) => MaybePromise<Output>} [maybe_fn]
	* @returns {RemoteCommand<Input, Output>}
	* @since 2.27
	*/
	/*@__NO_SIDE_EFFECTS__*/
	function command$1(validate_or_fn, maybe_fn) {
		/** @type {(arg?: Input) => MaybePromise<Output>} */
		const fn = maybe_fn ?? validate_or_fn;
		/** @type {(arg?: any) => MaybePromise<Input>} */
		const validate = create_validator(validate_or_fn, maybe_fn);
		/** @type {RemoteCommandInternals} */
		const __ = {
			type: "command",
			id: "",
			name: ""
		};
		/** @type {RemoteCommand<Input, Output> & { __: RemoteCommandInternals }} */
		const wrapper = (arg) => {
			const { event, state } = get_request_store();
			if (!MUTATIVE_METHODS.includes(event.request.method) || state.is_in_remote_query || state.is_in_remote_prerender) {
				const violation = state.is_in_remote_query || state.is_in_remote_prerender ? `inside a query or prerender function` : `from a ${event.request.method} handler`;
				throw new Error(`Cannot call a command (${__.name}) ${violation}`);
			}
			if (state.is_in_render) throw new Error(`Cannot call a command (${__.name}) during server-side rendering`);
			const promise = Promise.resolve(run_remote_function(event, state, true, () => validate(arg), fn));
			promise.updates = () => {
				throw new Error(`Cannot call '${__.name}(...).updates(...)' on the server`);
			};
			return promise;
		};
		Object.defineProperty(wrapper, "__", { value: __ });
		Object.defineProperty(wrapper, "pending", { get: () => 0 });
		return wrapper;
	}
	function _asyncIterator(r) {
		var n, t, o, e = 2;
		for ("undefined" != typeof Symbol && (t = Symbol.asyncIterator, o = Symbol.iterator); e--;) {
			if (t && null != (n = r[t])) return n.call(r);
			if (o && null != (n = r[o])) return new AsyncFromSyncIterator(n.call(r));
			t = "@@asyncIterator", o = "@@iterator";
		}
		throw new TypeError("Object is not async iterable");
	}
	function AsyncFromSyncIterator(r) {
		function AsyncFromSyncIteratorContinuation(r) {
			if (Object(r) !== r) return Promise.reject(/* @__PURE__ */ new TypeError(r + " is not an object."));
			var n = r.done;
			return Promise.resolve(r.value).then(function(r) {
				return {
					value: r,
					done: n
				};
			});
		}
		return AsyncFromSyncIterator = function AsyncFromSyncIterator(r) {
			this.s = r, this.n = r.next;
		}, AsyncFromSyncIterator.prototype = {
			s: null,
			n: null,
			next: function next() {
				return AsyncFromSyncIteratorContinuation(this.n.apply(this.s, arguments));
			},
			"return": function _return(r) {
				var n = this.s["return"];
				return void 0 === n ? Promise.resolve({
					value: r,
					done: !0
				}) : AsyncFromSyncIteratorContinuation(n.apply(this.s, arguments));
			},
			"throw": function _throw(r) {
				var n = this.s["return"];
				return void 0 === n ? Promise.reject(r) : AsyncFromSyncIteratorContinuation(n.apply(this.s, arguments));
			}
		}, new AsyncFromSyncIterator(r);
	}
	var text_encoder = new TextEncoder();
	new TextDecoder();
	/**
	* @param {Uint8Array} bytes
	* @returns {string}
	*/
	function base64_encode(bytes) {
		if (globalThis.Buffer) return globalThis.Buffer.from(bytes).toString("base64");
		let binary = "";
		for (let i = 0; i < bytes.length; i++) binary += String.fromCharCode(bytes[i]);
		return btoa(binary);
	}
	function _checkPrivateRedeclaration(e, t) {
		if (t.has(e)) throw new TypeError("Cannot initialize the same private elements twice on an object");
	}
	function _classPrivateFieldInitSpec(e, t, a) {
		_checkPrivateRedeclaration(e, t), t.set(e, a);
	}
	function _assertClassBrand(e, t, n) {
		if ("function" == typeof e ? e === t : e.has(t)) return arguments.length < 3 ? t : n;
		throw new TypeError("Private element is not present on this object");
	}
	function _classPrivateFieldSet2(s, a, r) {
		return s.set(_assertClassBrand(s, a), r), r;
	}
	function _classPrivateFieldGet2(s, a) {
		return s.get(_assertClassBrand(s, a));
	}
	var object_proto_names = /* @__PURE__ */ Object.getOwnPropertyNames(Object.prototype).sort().join("\0");
	/**
	* @param {unknown} thing
	* @returns {thing is Record<PropertyKey, unknown>}
	*/
	function is_plain_object(thing) {
		if (typeof thing !== "object" || thing === null) return false;
		const proto = Object.getPrototypeOf(thing);
		return proto === Object.prototype || proto === null || Object.getPrototypeOf(proto) === null || Object.getOwnPropertyNames(proto).sort().join("\0") === object_proto_names;
	}
	/**
	* @param {Record<string, any>} value
	* @param {Map<object, any>} clones
	*/
	function to_sorted(value, clones) {
		const clone = Object.getPrototypeOf(value) === null ? Object.create(null) : {};
		clones.set(value, clone);
		Object.defineProperty(clone, remote_arg_marker, { value: true });
		for (const key of Object.keys(value).sort()) {
			const property = value[key];
			Object.defineProperty(clone, key, {
				value: clones.get(property) ?? property,
				enumerable: true,
				configurable: true,
				writable: true
			});
		}
		return clone;
	}
	var remote_object = "__skrao";
	var remote_map = "__skram";
	var remote_set = "__skras";
	var remote_regex_guard = "__skrag";
	var remote_arg_marker = Symbol(remote_object);
	/**
	* @param {boolean} sort
	*/
	function create_remote_arg_reducers(sort) {
		/** @type {Record<string, (value: unknown) => unknown>} */
		const remote_fns_reducers = { 
		/** @param {unknown} value */
[remote_regex_guard]: (value) => {
			if (value instanceof RegExp) throw new Error("Regular expressions are not valid remote function arguments");
		} };
		if (sort) {
			const clones = /* @__PURE__ */ new Map();
			/** @type {(value: unknown) => Array<[unknown, unknown]> | undefined} */
			remote_fns_reducers[remote_map] = (value) => {
				if (!(value instanceof Map)) return;
				/** @type {Array<[string, string]>} */
				const entries = [];
				for (const [key, val] of value) entries.push([stringify(key), stringify(val)]);
				return entries.sort(([a1, a2], [b1, b2]) => {
					if (a1 < b1) return -1;
					if (a1 > b1) return 1;
					if (a2 < b2) return -1;
					if (a2 > b2) return 1;
					return 0;
				});
			};
			/** @type {(value: unknown) => unknown[] | undefined} */
			remote_fns_reducers[remote_set] = (value) => {
				if (!(value instanceof Set)) return;
				/** @type {string[]} */
				const items = [];
				for (const item of value) items.push(stringify(item));
				items.sort();
				return items;
			};
			/** @type {(value: unknown) => Record<PropertyKey, unknown> | undefined} */
			remote_fns_reducers[remote_object] = (value) => {
				if (!is_plain_object(value)) return;
				if (Object.hasOwn(value, remote_arg_marker)) return;
				if (clones.has(value)) return clones.get(value);
				return to_sorted(value, clones);
			};
		}
		const all_reducers = {
			...encoders,
			...remote_fns_reducers
		};
		/** @type {(value: unknown) => string} */
		const stringify = (value) => stringify$1(value, all_reducers);
		return all_reducers;
	}
	/**
	* Stringifies the argument (if any) for a remote function in such a way that
	* it is both a valid URL and a valid file name (necessary for prerendering).
	* @param {any} value
	*/
	function stringify_remote_arg(value) {
		if (value === void 0) return "";
		return url_friendly_base64_encode(stringify$1(value, create_remote_arg_reducers(true)));
	}
	/**
	* Base64-encodes `string` in such a way that the result is safe to use
	* as both a URI component and a filename
	* @param {string} string
	*/
	function url_friendly_base64_encode(string) {
		return base64_encode(text_encoder.encode(string)).replaceAll("=", "").replaceAll("+", "-").replaceAll("/", "_");
	}
	/**
	* @param {string} id
	* @param {string} payload
	*/
	function create_remote_key(id, payload) {
		return id + "/" + payload;
	}
	new RegExp(`${/\[([ux])\+([^\]]+)\]/.source}|${/\[(\[)?(\.\.\.)?([\w-]+?)(?:=([\w-]+))?\]\]?/g.source}`, "g");
	var symbol = Symbol.for("sveltekit.global_state");
	globalThis[symbol] ??= {};
	/**
	* @param {unknown} err
	* @return {Error}
	*/
	function coalesce_to_error(err) {
		return err instanceof Error || err && err.name && err.message ? err : new Error(JSON.stringify(err));
	}
	/**
	* When inside a double-quoted attribute value, only `&` and `"` hold special meaning.
	* @see https://html.spec.whatwg.org/multipage/parsing.html#attribute-value-(double-quoted)-state
	* @type {Record<string, string>}
	*/
	var escape_html_attr_dict = {
		"&": "&amp;",
		"\"": "&quot;"
	};
	/**
	* @type {Record<string, string>}
	*/
	var escape_html_dict = {
		"&": "&amp;",
		"<": "&lt;"
	};
	/** @param {Record<string, string>} dict */
	var escape_regex = (dict) => new RegExp(`[${Object.keys(dict).join("")}]|\\p{Surrogate}`, "gu");
	escape_regex(escape_html_attr_dict);
	escape_regex(escape_html_dict);
	/**
	* @param {import('@sveltejs/kit').RequestEvent} event
	* @param {import('types').RequestState} state
	* @param {any} error
	* @returns {App.Error | Promise<App.Error>}
	*/
	function handle_error_and_jsonify(event, state, error) {
		if (error instanceof HandledHttpError) return error.body;
		/** @type {import('@sveltejs/kit/hooks').CaughtError} */
		let caught;
		if (error instanceof HttpError) caught = {
			kind: "app",
			error: error.body
		};
		else if (error instanceof SvelteKitError) caught = {
			kind: "framework",
			error: {
				status: error.status,
				message: error.text
			}
		};
		else if (error instanceof ValidationError) caught = {
			kind: "validation",
			error: {
				status: 400,
				message: "Bad Request"
			},
			issues: error.issues
		};
		else {
			caught = {
				kind: "unknown",
				error
			};
			let e = error;
			while (e instanceof Error) {
				fix_stack_trace(e);
				e = e.cause;
			}
		}
		const fallback = caught.kind === "unknown" ? {
			status: 500,
			message: "Internal Error"
		} : caught.error;
		/**
		* The hook returns only the properties it wants to override; anything it omits
		* (including by returning nothing at all) is inherited from the caught error.
		* @param {Awaited<ReturnType<import('@sveltejs/kit/hooks').HandleServerError>>} body
		* @returns {App.Error}
		*/
		function merge(body) {
			return {
				...fallback,
				...body
			};
		}
		let result;
		try {
			const input = {
				...caught,
				event
			};
			result = with_request_store({
				event,
				state
			}, () => hooks.handleError(input));
		} catch (hook_error) {
			log_handle_error_hook_failure(error, hook_error);
			return {
				status: fallback.status,
				message: "Internal Error"
			};
		}
		if (result instanceof Promise) return result.then(merge, (hook_error) => {
			log_handle_error_hook_failure(error, hook_error);
			return {
				status: fallback.status,
				message: "Internal Error"
			};
		});
		return merge(result);
	}
	/**
	* @param {unknown} error
	* @param {unknown} hook_error
	*/
	function log_handle_error_hook_failure(error, hook_error) {
		const failure = new Error("The `handleError` hook failed", { cause: coalesce_to_error(hook_error) });
		failure.stack = failure.message;
		console.error(failure);
		if (error instanceof SvelteKitError) console.error(`Original error: ${error.status} ${error.text}: ${error.message}`);
		else console.error("Original error:", error);
	}
	var fs = globalThis.process?.getBuiltinModule?.("node:fs");
	var url = globalThis.process?.getBuiltinModule?.("node:url");
	var path = globalThis.process?.getBuiltinModule?.("node:path");
	var module = globalThis.process?.getBuiltinModule?.("node:module");
	var cwd = globalThis.process?.cwd?.();
	/** @type {(file: string) => string} */
	var relative = cwd ? (file) => path.relative(cwd, file) : (file) => file;
	/**
	* Applies sourcemaps, makes paths relative to the cwd, and truncates
	* non-user code from the bottom of the stack
	* @param {Error} error
	* @returns void
	*/
	var fix_stack_trace = (error) => {
		if (!error.stack || !fs) return;
		let end = 0;
		error.stack = error.stack.split("\n").map((line, i) => {
			const match = line.match(/^ {4}at.+(file:\/\/\/.*):(\d+):(\d+)(\)?)$/);
			if (!match) {
				if (!line.includes("node:internal/")) end = i + 1;
				return line;
			}
			const file = url.fileURLToPath(match[1]);
			const traced = trace(file, Number(match[2]) - 1, Number(match[3]) - 1);
			if (!/[\\/]node_modules[\\/]/.test(traced?.file ?? file)) end = i + 1;
			if (traced?.line) {
				const location = `${match[1]}:${match[2]}:${match[3]}`;
				const original = `${relative(traced.file)}:${traced.line}:${traced.column}`;
				return line.replace(location, original);
			}
			if (traced) return `${line.replace(match[1], relative(file))} [${traced.file}]`;
			return line;
		}).slice(0, end).join("\n");
	};
	/** @type {Map<string, { map: import('node:module').SourceMap; directory: string } | null>} */
	var source_maps = /* @__PURE__ */ new Map();
	/** @type {Map<string, Array<string | undefined>>} */
	var source_regions = /* @__PURE__ */ new Map();
	/** @param {string} file */
	function get_source_map(file) {
		if (source_maps.has(file)) return source_maps.get(file);
		try {
			let source;
			let directory = path.dirname(file);
			const code = fs.readFileSync(file, "utf8");
			const url = Array.from(code.matchAll(/\/\/[#@]\s*sourceMappingURL=(\S+)/g)).at(-1)?.[1];
			if (url?.startsWith("data:")) {
				const comma = url.indexOf(",");
				const metadata = url.slice(5, comma);
				const data = url.slice(comma + 1);
				source = metadata.endsWith(";base64") ? Buffer.from(data, "base64").toString() : decodeURIComponent(data);
			} else {
				const map_file = url ? path.resolve(path.dirname(file), decodeURIComponent(url)) : `${file}.map`;
				if (fs.existsSync(map_file)) {
					directory = path.dirname(map_file);
					source = fs.readFileSync(map_file, "utf8");
				}
			}
			if (source) {
				const source_map = {
					map: new module.SourceMap(JSON.parse(source)),
					directory
				};
				source_maps.set(file, source_map);
				return source_map;
			}
		} catch {}
		source_maps.set(file, null);
		return null;
	}
	/**
	*
	* @param {string} file
	* @param {number} line
	* @param {number} column
	* @returns {null | { file: string, line?: number, column?: number }}
	*/
	function trace(file, line, column) {
		const source_map = get_source_map(file);
		if (!source_map) return null;
		const entry = source_map.map.findEntry(line, column);
		if (entry && "originalSource" in entry && entry.originalSource && typeof entry.originalLine === "number" && typeof entry.originalColumn === "number") {
			const traced = {
				file: entry.originalSource.startsWith("file:") ? url.fileURLToPath(entry.originalSource) : path.resolve(source_map.directory, entry.originalSource),
				line: entry.originalLine + 1,
				column: entry.originalColumn + 1
			};
			return trace(traced.file, traced.line - 1, traced.column - 1) ?? traced;
		}
		let regions = source_regions.get(file);
		if (!regions) {
			/** @type {string | undefined} */
			let source;
			regions = fs.readFileSync(file, "utf8").split("\n").map((line) => {
				const start = line.match(/^\/\/#region (.+)$/);
				if (start) source = start[1];
				if (line === "//#endregion") source = void 0;
				return source;
			});
			source_regions.set(file, regions);
		}
		const source = regions[line];
		if (source) return { file: source };
		return null;
	}
	globalThis.process?.getBuiltinModule?.("node:util")?.styleText;
	var hooks = false;
	var prerendering = false;
	var _subscribers = /* @__PURE__ */ new WeakMap();
	var _start = /* @__PURE__ */ new WeakMap();
	var _stop = /* @__PURE__ */ new WeakMap();
	var _closed = /* @__PURE__ */ new WeakMap();
	var _terminal_error = /* @__PURE__ */ new WeakMap();
	/**
	* A pull-style async iterator that fans out a single stream of values to
	* multiple `for await (...)` consumers. Each subscriber gets its own
	* `AsyncGenerator` whose `.next()` resolves whenever a value is pushed via
	* `push(value)`. Multiple consumers see the same values without each one
	* driving an independent underlying source.
	*
	* Backpressure is **latest-wins**: if values arrive faster than a particular
	* consumer drains its iterator, only the most-recently-pushed value is kept
	* pending for that subscriber. Earlier undrained values are dropped. This is
	* appropriate for live data streams (reactive state replication), not for
	* event logs where every value must be delivered.
	*
	* Lifecycle hooks are exposed via the constructor:
	*
	*   - `start()` is called when the subscriber count transitions
	*     from 0 to 1 (e.g. to start a pump pulling from a real source).
	*   - `stop()` is called when the subscriber count transitions
	*     from non-zero back to 0 (e.g. to tear down that pump).
	*
	* Either hook may be omitted.
	*
	* The owner is responsible for calling `push(value)` to broadcast values,
	* `done()` to signal natural completion to all subscribers, and `fail(error)`
	* to broadcast a terminal error. After `done()` or `fail()`, the iterator
	* rejects further `subscribe()` calls with the terminal state appropriately.
	*
	* @template T
	*/
	var SharedIterator = class {
		/** Whether `done()` or `fail()` has been broadcast. */
		get closed() {
			return _classPrivateFieldGet2(_closed, this);
		}
		/**
		* @param {(instance: SharedIterator<T>) => (() => void)} [start]
		*/
		constructor(start) {
			_classPrivateFieldInitSpec(this, _subscribers, /* @__PURE__ */ new Set());
			_classPrivateFieldInitSpec(this, _start, void 0);
			_classPrivateFieldInitSpec(this, _stop, void 0);
			_classPrivateFieldInitSpec(this, _closed, false);
			_classPrivateFieldInitSpec(this, _terminal_error, void 0);
			_classPrivateFieldSet2(_start, this, start);
		}
		/** @param {T} value */
		push(value) {
			if (_classPrivateFieldGet2(_closed, this)) return;
			for (const subscriber of _classPrivateFieldGet2(_subscribers, this)) if (subscriber.waiting_resolve) {
				const resolve = subscriber.waiting_resolve;
				subscriber.waiting_resolve = null;
				subscriber.waiting_reject = null;
				resolve({
					value,
					done: false
				});
			} else subscriber.pending = { value };
		}
		/**
		* Signal natural completion to all current subscribers, and to any future
		* subscriber (which will receive an immediately-done iterator).
		*/
		done() {
			if (_classPrivateFieldGet2(_closed, this)) return;
			_classPrivateFieldSet2(_closed, this, true);
			for (const subscriber of _classPrivateFieldGet2(_subscribers, this)) {
				subscriber.finished = true;
				if (subscriber.waiting_resolve) {
					const resolve = subscriber.waiting_resolve;
					subscriber.waiting_resolve = null;
					subscriber.waiting_reject = null;
					resolve({
						value: void 0,
						done: true
					});
				}
			}
			_classPrivateFieldGet2(_subscribers, this).clear();
		}
		/**
		* Broadcast a terminal error. All current subscribers will reject their
		* next `.next()` call with `error`. Future subscribers will also reject
		* their first `.next()`.
		*
		* @param {unknown} error
		*/
		fail(error) {
			if (_classPrivateFieldGet2(_closed, this)) return;
			_classPrivateFieldSet2(_closed, this, true);
			_classPrivateFieldSet2(_terminal_error, this, error);
			for (const subscriber of _classPrivateFieldGet2(_subscribers, this)) {
				subscriber.finished = true;
				if (subscriber.waiting_reject) {
					const reject = subscriber.waiting_reject;
					subscriber.waiting_resolve = null;
					subscriber.waiting_reject = null;
					reject(error);
				} else subscriber.pending_error = { error };
			}
			_classPrivateFieldGet2(_subscribers, this).clear();
		}
		/**
		* Subscribe to the shared stream. Returns an `AsyncGenerator<T>` that
		* yields every value pushed after this call (and, if `initial_value` is
		* provided, that value as the first yield).
		*
		* @param {{ initial_value?: { value: T } }} [options]
		*   `initial_value` lets the caller seed the iterator with a synchronously-
		*   available current value before any new pushes arrive (e.g. the
		*   "last-seen value" of a reactive resource). Pass it wrapped in an
		*   object so `undefined` can be distinguished from "no initial value".
		* @returns {AsyncGenerator<T, void, void>}
		*/
		subscribe(options) {
			/** @type {Subscriber} */
			const subscriber = {
				pending: (options === null || options === void 0 ? void 0 : options.initial_value) ? { value: options.initial_value.value } : null,
				pending_error: _classPrivateFieldGet2(_closed, this) && _classPrivateFieldGet2(_terminal_error, this) !== void 0 ? { error: _classPrivateFieldGet2(_terminal_error, this) } : null,
				finished: _classPrivateFieldGet2(_closed, this) && _classPrivateFieldGet2(_terminal_error, this) === void 0,
				waiting_resolve: null,
				waiting_reject: null
			};
			if (!subscriber.finished && subscriber.pending_error === null) _classPrivateFieldGet2(_subscribers, this).add(subscriber);
			if (!_classPrivateFieldGet2(_closed, this)) {
				var _classPrivateFieldGet2$1, _classPrivateFieldGet3;
				(_classPrivateFieldGet2$1 = _classPrivateFieldGet2(_stop, this)) !== null && _classPrivateFieldGet2$1 !== void 0 || _classPrivateFieldSet2(_stop, this, (_classPrivateFieldGet3 = _classPrivateFieldGet2(_start, this)) === null || _classPrivateFieldGet3 === void 0 ? void 0 : _classPrivateFieldGet3.call(this, this));
			}
			const unsubscribe = () => {
				subscriber.finished = true;
				if (_classPrivateFieldGet2(_subscribers, this).delete(subscriber) && _classPrivateFieldGet2(_subscribers, this).size === 0) {
					var _classPrivateFieldGet4;
					(_classPrivateFieldGet4 = _classPrivateFieldGet2(_stop, this)) === null || _classPrivateFieldGet4 === void 0 || _classPrivateFieldGet4.call(this);
				}
			};
			/** @type {AsyncGenerator<T, void, void>} */
			const iterator = {
				next() {
					if (subscriber.pending_error) {
						const { error } = subscriber.pending_error;
						subscriber.pending_error = null;
						unsubscribe();
						return Promise.reject(error);
					}
					if (subscriber.pending) {
						const { value } = subscriber.pending;
						subscriber.pending = null;
						return Promise.resolve({
							value,
							done: false
						});
					}
					if (subscriber.finished) return Promise.resolve({
						value: void 0,
						done: true
					});
					return new Promise((resolve, reject) => {
						subscriber.waiting_resolve = resolve;
						subscriber.waiting_reject = reject;
					});
				},
				return(value) {
					unsubscribe();
					if (subscriber.waiting_resolve) {
						const resolve = subscriber.waiting_resolve;
						subscriber.waiting_resolve = null;
						subscriber.waiting_reject = null;
						resolve({
							value: void 0,
							done: true
						});
					}
					return Promise.resolve({
						value,
						done: true
					});
				},
				throw(error) {
					unsubscribe();
					if (subscriber.waiting_reject) {
						const reject = subscriber.waiting_reject;
						subscriber.waiting_resolve = null;
						subscriber.waiting_reject = null;
						reject(error);
					}
					return Promise.reject(error);
				},
				[Symbol.asyncIterator]() {
					return iterator;
				},
				async [Symbol.asyncDispose]() {}
			};
			return iterator;
		}
	};
	/** @import { RemoteLiveQuery, RemoteLiveQueryFunction, RemoteQuery, RemoteQueryFunction } from '$app/server' */
	/** @import { RequestEvent } from '@sveltejs/kit' */
	/** @import { RemoteInternals, MaybePromise, RequestState, RemoteQueryLiveInternals, RemoteQueryBatchInternals, RemoteQueryInternals, RemoteLiveQueryUserFunctionReturnType } from 'types' */
	/** @import { StandardSchemaV1 } from '@standard-schema/spec' */
	/**
	* Creates a remote query. When called from the browser, the function will be invoked on the server via a `fetch` call.
	*
	* See [Remote functions](https://svelte.dev/docs/kit/remote-functions#query) for full documentation.
	*
	* @template Output
	* @overload
	* @param {() => MaybePromise<Output>} fn
	* @returns {RemoteQueryFunction<void, Output>}
	* @since 2.27
	*/
	/**
	* Creates a remote query. When called from the browser, the function will be invoked on the server via a `fetch` call.
	*
	* See [Remote functions](https://svelte.dev/docs/kit/remote-functions#query) for full documentation.
	*
	* @template Input
	* @template Output
	* @overload
	* @param {'unchecked'} validate
	* @param {(arg: Input) => MaybePromise<Output>} fn
	* @returns {RemoteQueryFunction<Input, Output>}
	* @since 2.27
	*/
	/**
	* Creates a remote query. When called from the browser, the function will be invoked on the server via a `fetch` call.
	*
	* See [Remote functions](https://svelte.dev/docs/kit/remote-functions#query) for full documentation.
	*
	* @template {StandardSchemaV1} Schema
	* @template Output
	* @overload
	* @param {Schema} schema
	* @param {(arg: StandardSchemaV1.InferOutput<Schema>) => MaybePromise<Output>} fn
	* @returns {RemoteQueryFunction<StandardSchemaV1.InferInput<Schema>, Output, StandardSchemaV1.InferOutput<Schema>>}
	* @since 2.27
	*/
	/**
	* @template Input
	* @template Output
	* @param {any} validate_or_fn
	* @param {(args?: Input) => MaybePromise<Output>} [maybe_fn]
	* @returns {RemoteQueryFunction<Input, Output>}
	* @since 2.27
	*/
	/*@__NO_SIDE_EFFECTS__*/
	function query$1(validate_or_fn, maybe_fn) {
		/** @type {(arg?: Input) => Output} */
		const fn = maybe_fn !== null && maybe_fn !== void 0 ? maybe_fn : validate_or_fn;
		/** @type {(arg?: any) => MaybePromise<Input>} */
		const validate = create_validator(validate_or_fn, maybe_fn);
		/** @type {RemoteQueryInternals} */
		const __ = {
			type: "query",
			id: "",
			name: "",
			validate,
			bind(payload, validated_arg) {
				const { event, state } = get_request_store();
				return create_query_resource(__, payload, event, state, () => run_remote_function(event, _objectSpread2(_objectSpread2({}, state), {}, { is_in_remote_query: true }), false, () => validated_arg, fn));
			}
		};
		/** @type {RemoteQueryFunction<Input, Output> & { __: RemoteQueryInternals }} */
		const wrapper = (arg) => {
			if (prerendering) throw new Error(`Cannot call query '${__.name}' while prerendering, as prerendered pages need static data. Use 'prerender' from $app/server instead`);
			const { event, state } = get_request_store();
			const payload = stringify_remote_arg(arg);
			return create_query_resource(__, payload, event, state, () => run_remote_function(event, _objectSpread2(_objectSpread2({}, state), {}, { is_in_remote_query: true }), false, () => validate(arg), fn));
		};
		Object.defineProperty(wrapper, "__", { value: __ });
		return wrapper;
	}
	/**
	* Creates a live remote query. When called from the browser, the function will be invoked on the server via a streaming `fetch` call.
	*
	* See [Remote functions](https://svelte.dev/docs/kit/remote-functions#query.live) for full documentation.
	*
	* @template Output
	* @overload
	* @param {(arg: void) => RemoteLiveQueryUserFunctionReturnType<Output>} fn
	* @returns {RemoteLiveQueryFunction<void, Output>}
	*/
	/**
	* @template Input
	* @template Output
	* @overload
	* @param {'unchecked'} validate
	* @param {(arg: Input) => RemoteLiveQueryUserFunctionReturnType<Output>} fn
	* @returns {RemoteLiveQueryFunction<Input, Output>}
	*/
	/**
	* @template {StandardSchemaV1} Schema
	* @template Output
	* @overload
	* @param {Schema} schema
	* @param {(arg: StandardSchemaV1.InferOutput<Schema>) => RemoteLiveQueryUserFunctionReturnType<Output>} fn
	* @returns {RemoteLiveQueryFunction<StandardSchemaV1.InferInput<Schema>, Output, StandardSchemaV1.InferOutput<Schema>>}
	*/
	/**
	* @template Input
	* @template Output
	* @param {any} validate_or_fn
	* @param {(args: Input) => RemoteLiveQueryUserFunctionReturnType<Output>} [maybe_fn]
	* @returns {RemoteLiveQueryFunction<Input, Output>}
	*/
	/*@__NO_SIDE_EFFECTS__*/
	function live(validate_or_fn, maybe_fn) {
		/** @type {(arg: Input) => RemoteLiveQueryUserFunctionReturnType<Output>} */
		const fn = maybe_fn !== null && maybe_fn !== void 0 ? maybe_fn : validate_or_fn;
		/** @type {(arg?: any) => MaybePromise<Input>} */
		const validate = create_validator(validate_or_fn, maybe_fn);
		/**
		* @param {any} event
		* @param {any} state
		* @param {any} get_input
		*/
		const run = (event, state, get_input) => run_remote_generator(event, _objectSpread2(_objectSpread2({}, state), {}, { is_in_remote_query: true }), false, get_input, fn, __.name);
		/** @type {RemoteQueryLiveInternals} */
		const __ = {
			type: "query_live",
			id: "",
			name: "",
			run: (event, state, arg) => run(event, state, () => validate(arg)),
			validate,
			bind(payload, validated_arg) {
				const { event, state } = get_request_store();
				return create_live_query_resource(__, payload, event, state, () => run(event, state, () => validated_arg));
			}
		};
		/** @type {RemoteLiveQueryFunction<Input, Output> & { __: RemoteQueryLiveInternals }} */
		const wrapper = (arg) => {
			if (prerendering) throw new Error(`Cannot call query.live '${__.name}' while prerendering, as prerendered pages need static data. Use 'prerender' from $app/server instead`);
			const { event, state } = get_request_store();
			const payload = stringify_remote_arg(arg);
			return create_live_query_resource(__, payload, event, state, () => run(event, state, () => validate(arg)));
		};
		Object.defineProperty(wrapper, "__", { value: __ });
		return wrapper;
	}
	/**
	* Creates a batch query function that collects multiple calls and executes them in a single request
	*
	* See [Remote functions](https://svelte.dev/docs/kit/remote-functions#query.batch) for full documentation.
	*
	* @template Input
	* @template Output
	* @overload
	* @param {'unchecked'} validate
	* @param {(args: Input[]) => MaybePromise<(arg: Input, idx: number) => Output>} fn
	* @returns {RemoteQueryFunction<Input, Output>}
	* @since 2.35
	*/
	/**
	* Creates a batch query function that collects multiple calls and executes them in a single request
	*
	* See [Remote functions](https://svelte.dev/docs/kit/remote-functions#query.batch) for full documentation.
	*
	* @template {StandardSchemaV1} Schema
	* @template Output
	* @overload
	* @param {Schema} schema
	* @param {(args: StandardSchemaV1.InferOutput<Schema>[]) => MaybePromise<(arg: StandardSchemaV1.InferOutput<Schema>, idx: number) => Output>} fn
	* @returns {RemoteQueryFunction<StandardSchemaV1.InferInput<Schema>, Output, StandardSchemaV1.InferOutput<Schema>>}
	* @since 2.35
	*/
	/**
	* @template Input
	* @template Output
	* @param {any} validate_or_fn
	* @param {(args?: Input[]) => MaybePromise<(arg: Input, idx: number) => Output>} [maybe_fn]
	* @returns {RemoteQueryFunction<Input, Output>}
	* @since 2.35
	*/
	/*@__NO_SIDE_EFFECTS__*/
	function batch(validate_or_fn, maybe_fn) {
		/** @type {(args?: Input[]) => MaybePromise<(arg: Input, idx: number) => Output>} */
		const fn = maybe_fn !== null && maybe_fn !== void 0 ? maybe_fn : validate_or_fn;
		/** @type {(arg?: any) => MaybePromise<Input>} */
		const validate = create_validator(validate_or_fn, maybe_fn);
		/**
		* Enqueues a single call into the current batch (creating one if necessary)
		* and returns a promise that resolves with the result for this entry.
		*
		* @param {string} payload — the stringified raw argument (cache key)
		* @param {() => MaybePromise<any>} get_validated — produces the validated argument for this entry
		* @returns {Promise<any>}
		*/
		const enqueue = (payload, get_validated) => {
			const { event, state } = get_request_store();
			return new Promise((resolve, reject) => {
				var _state$remote, _state$remote$batches;
				const batches = (_state$remote$batches = (_state$remote = state.remote).batches) !== null && _state$remote$batches !== void 0 ? _state$remote$batches : _state$remote.batches = /* @__PURE__ */ new Map();
				let batched = batches.get(__.id);
				if (!batched) {
					batched = /* @__PURE__ */ new Map();
					batches.set(__.id, batched);
				}
				const entry = batched.get(payload);
				if (entry) {
					entry.resolvers.push({
						resolve,
						reject
					});
					return;
				}
				batched.set(payload, {
					get_validated,
					resolvers: [{
						resolve,
						reject
					}]
				});
				if (batched.size > 1) return;
				setTimeout(async () => {
					batches.delete(__.id);
					const entries = Array.from(batched.values());
					try {
						return await run_remote_function(event, _objectSpread2(_objectSpread2({}, state), {}, { is_in_remote_query: true }), false, async () => Promise.all(entries.map((entry) => entry.get_validated())), async (input) => {
							const get_result = await fn(input);
							for (let i = 0; i < entries.length; i++) try {
								const result = get_result(input[i], i);
								for (const resolver of entries[i].resolvers) resolver.resolve(result);
							} catch (error) {
								for (const resolver of entries[i].resolvers) resolver.reject(error);
							}
						});
					} catch (error) {
						for (const entry of batched.values()) for (const resolver of entry.resolvers) resolver.reject(error);
					}
				}, 0);
			});
		};
		/** @type {RemoteQueryBatchInternals} */
		const __ = {
			type: "query_batch",
			id: "",
			name: "",
			validate,
			run: async (args) => {
				const { event, state } = get_request_store();
				return run_remote_function(event, _objectSpread2(_objectSpread2({}, state), {}, { is_in_remote_query: true }), false, async () => Promise.all(args.map(validate)), async (input) => {
					const get_result = await fn(input);
					return Promise.all(input.map(async (arg, i) => {
						try {
							return {
								type: "result",
								data: get_result(arg, i)
							};
						} catch (error) {
							return {
								type: "error",
								error: await handle_error_and_jsonify(event, state, error)
							};
						}
					}));
				});
			},
			bind(payload, validated_arg) {
				const { event, state } = get_request_store();
				return create_query_resource(__, payload, event, state, () => enqueue(payload, () => validated_arg));
			}
		};
		/** @type {RemoteQueryFunction<Input, Output> & { __: RemoteQueryBatchInternals }} */
		const wrapper = (arg) => {
			if (prerendering) throw new Error(`Cannot call query.batch '${__.name}' while prerendering, as prerendered pages need static data. Use 'prerender' from $app/server instead`);
			const { event, state } = get_request_store();
			const payload = stringify_remote_arg(arg);
			return create_query_resource(__, payload, event, state, () => enqueue(payload, () => validate(arg)));
		};
		Object.defineProperty(wrapper, "__", { value: __ });
		return wrapper;
	}
	/**
	* Include this value in the returned payload...
	* @param {RequestEvent} event
	* @param {RequestState} state
	* @param {RemoteInternals} internals
	* @param {string} payload
	* @param {() => Promise<any>} fn
	*/
	function refresh(event, state, internals, payload, fn) {
		var _state$remote2, _state$remote2$explic;
		if (!internals.id) return;
		if (!event.isRemoteRequest && state.is_in_remote_form_or_command) return;
		const key = create_remote_key(internals.id, payload);
		((_state$remote2$explic = (_state$remote2 = state.remote).explicit) !== null && _state$remote2$explic !== void 0 ? _state$remote2$explic : _state$remote2.explicit = /* @__PURE__ */ new Map()).set(key, {
			internals,
			fn
		});
	}
	/**
	* @param {RemoteInternals} __
	* @param {string} payload — the stringified raw argument (i.e. the cache key the client will use)
	* @param {RequestEvent} event
	* @param {RequestState} state
	* @param {() => Promise<any>} fn
	* @returns {RemoteQuery<any>}
	*/
	function create_query_resource(__, payload, event, state, fn) {
		/** @type {Promise<any> | null} */
		let promise = null;
		const get_promise = () => {
			var _promise;
			return (_promise = promise) !== null && _promise !== void 0 ? _promise : promise = get_response(__, payload, state, fn);
		};
		const populate_hydratable = () => {
			if (__.id && state.is_in_render) get_promise().catch(noop);
		};
		return {
			/** @type {Promise<any>['catch']} */
			catch(onrejected) {
				return get_promise().catch(onrejected);
			},
			get current() {
				populate_hydratable();
			},
			get error() {
				populate_hydratable();
			},
			/** @type {Promise<any>['finally']} */
			finally(onfinally) {
				return get_promise().finally(onfinally);
			},
			get loading() {
				populate_hydratable();
				return true;
			},
			get ready() {
				populate_hydratable();
				return false;
			},
			refresh() {
				promise = null;
				delete get_cache(__, state)[payload];
				refresh(event, state, __, payload, get_promise);
				return Promise.resolve();
			},
			/** @param {any} value */
			set(value) {
				const p = promise = Promise.resolve(value);
				get_cache(__, state)[payload] = p;
				refresh(event, state, __, payload, () => p);
			},
			/** @type {Promise<any>['then']} */
			then(onfulfilled, onrejected) {
				return get_promise().then(onfulfilled, onrejected);
			},
			withOverride() {
				throw new Error(`Cannot call '${__.name}.withOverride()' on the server`);
			},
			get [Symbol.toStringTag]() {
				return "QueryResource";
			}
		};
	}
	/**
	* @param {RemoteQueryLiveInternals} __
	* @param {string} payload — the stringified raw argument (i.e. the cache key the client will use)
	* @param {RequestEvent} event
	* @param {RequestState} state
	* @param {() => AsyncGenerator<any, void, void>} get_generator
	* @returns {RemoteLiveQuery<any>}
	*/
	function create_live_query_resource(__, payload, event, state, get_generator) {
		/** @type {Promise<any> | null} */
		let promise = null;
		const get_first_value = async () => {
			var _iteratorAbruptCompletion = false;
			var _didIteratorError = false;
			var _iteratorError;
			try {
				for (var _iterator = _asyncIterator(get_generator()), _step; _iteratorAbruptCompletion = !(_step = await _iterator.next()).done; _iteratorAbruptCompletion = false) return _step.value;
			} catch (err) {
				_didIteratorError = true;
				_iteratorError = err;
			} finally {
				try {
					if (_iteratorAbruptCompletion && _iterator.return != null) await _iterator.return();
				} finally {
					if (_didIteratorError) throw _iteratorError;
				}
			}
			throw new Error(`query.live '${__.name}' did not yield a value`);
		};
		const get_promise = () => {
			var _promise2;
			return (_promise2 = promise) !== null && _promise2 !== void 0 ? _promise2 : promise = get_response(__, payload, state, get_first_value);
		};
		const populate_hydratable = () => {
			if (__.id && state.is_in_render) get_promise().catch(noop);
		};
		return {
			/** @type {Promise<any>['catch']} */
			catch(onrejected) {
				return get_promise().catch(onrejected);
			},
			get current() {
				populate_hydratable();
			},
			get error() {
				populate_hydratable();
			},
			/** @type {Promise<any>['finally']} */
			finally(onfinally) {
				return get_promise().finally(onfinally);
			},
			get done() {
				populate_hydratable();
				return false;
			},
			get loading() {
				populate_hydratable();
				return true;
			},
			get ready() {
				populate_hydratable();
				return false;
			},
			get connected() {
				populate_hydratable();
				return false;
			},
			reconnect() {
				promise = null;
				delete get_cache(__, state)[payload];
				refresh(event, state, __, payload, get_promise);
				return Promise.resolve();
			},
			/** @type {Promise<any>['then']} */
			then(onfulfilled, onrejected) {
				return get_promise().then(onfulfilled, onrejected);
			},
			[Symbol.asyncIterator]() {
				var _state$remote3, _state$remote3$live_i;
				const key = create_remote_key(__.id, payload);
				const cache = (_state$remote3$live_i = (_state$remote3 = state.remote).live_iterators) !== null && _state$remote3$live_i !== void 0 ? _state$remote3$live_i : _state$remote3.live_iterators = /* @__PURE__ */ new Map();
				let cached = cache.get(key);
				if (!cached) {
					cached = create_shared_live_iterator(event.request.signal, get_generator);
					cache.set(key, cached);
				}
				return cached.subscribe();
			},
			get [Symbol.toStringTag]() {
				return "LiveQueryResource";
			}
		};
	}
	/**
	* Wraps a lazily-created live-query generator so that multiple `for await`
	* consumers within the same request share one underlying iteration. The first
	* subscriber starts the generator; values are broadcast to all subscribers
	* via a `SharedIterator`. When the last subscriber unsubscribes, the generator
	* is closed via `generator.return(undefined)`.
	*
	* If `signal` aborts (typically because the client has disconnected), the
	* pump is torn down and any in-flight `next()` calls on consumer iterators
	* resolve with `{ done: true }`, so suspended `for await` loops unwind
	* cleanly rather than leaking.
	*
	* @param {AbortSignal} signal
	* @param {() => AsyncGenerator<any, void, void>} get_generator
	*/
	function create_shared_live_iterator(signal, get_generator) {
		return new SharedIterator((instance) => {
			if (signal.aborted) {
				instance.done();
				return noop;
			}
			const generator = get_generator();
			let aborted = false;
			const close = () => {
				aborted = true;
				generator.return().catch(noop);
			};
			signal.addEventListener("abort", () => (close(), instance.done()), { once: true });
			(async () => {
				try {
					while (true) {
						const result = await generator.next();
						if (result.done) {
							instance.done();
							return;
						}
						instance.push(result.value);
					}
				} catch (error) {
					if (!aborted) instance.fail(error);
				} finally {
					close();
				}
			})();
			return close;
		});
	}
	Object.defineProperty(query$1, "batch", {
		value: batch,
		enumerable: true
	});
	Object.defineProperty(query$1, "live", {
		value: live,
		enumerable: true
	});
	/**
	* Calls Go asynchronously. Its I/O runs outside the runtime while other
	* independent queries start; the runtime resumes when Go settles the promise.
	*
	* A refusal comes back as one of kit's own control objects rather than a bare
	* Error, because everything downstream classifies by type: handle_error
	* keeps an HttpError's body and replaces anything else with Internal Error,
	* and transformError rethrows a Redirect so that the whole document becomes
	* the 3xx kit answers with.
	*
	* The answer is devalue's flat form, the same bytes Go would have sent the
	* browser from /_app/remote/..., and it is read back with kit's own parse —
	* the app's transport decoders. So a query answering with a custom type hands
	* the component an instance of the app's class, and a method call on it during
	* a render works for the same reason it works in the browser.
	*/
	async function host(id, payload) {
		const raw = await globalThis.__skgo_remote(id, payload);
		const res = JSON.parse(raw);
		if (res.r) throw new Redirect(res.r.status, res.r.location);
		if (res.e) throw new HttpError({
			status: res.e.status ?? 500,
			message: res.e.message
		});
		return parse(res.v);
	}
	function query(validate_or_fn, maybe_fn) {
		const fn = (arg) => host(wrapper.__.id, stringify_remote_arg(arg));
		const wrapper = maybe_fn ? /* @__PURE__ */ query$1(validate_or_fn, fn) : /* @__PURE__ */ query$1(fn);
		return wrapper;
	}
	/**
	* Calls Go once for a whole batch. Kit's own enqueue has already collected
	* every call made to this function in one macrotask and deduplicated them by
	* payload, so args is the batch — and the app's Go function is invoked once
	* for all of it, which is the only thing that makes a batch query different
	* from a query.
	*
	* The payloads travel as a JSON array in the place a single payload would
	* occupy, and the answer is {n: [...]} — one envelope per payload, in the
	* order they were sent, which is the order kit's client matches results to
	* arguments by.
	*/
	async function host_batch(id, args) {
		const payloads = args.map((arg) => stringify_remote_arg(arg));
		const raw = await globalThis.__skgo_remote(id, JSON.stringify(payloads));
		const res = JSON.parse(raw);
		if (res.r) throw new Redirect(res.r.status, res.r.location);
		if (res.e) throw new HttpError({
			status: res.e.status ?? 500,
			message: res.e.message
		});
		const nodes = res.n ?? [];
		return (_arg, i) => {
			const node = nodes[i];
			if (!node) throw new HttpError({
				status: 500,
				message: "Internal Error"
			});
			if (node.e) throw new HttpError({
				status: node.e.status ?? 500,
				message: node.e.message
			});
			return parse(node.v);
		};
	}
	query.batch = (validate_or_fn, maybe_fn) => {
		const fn = (args) => host_batch(wrapper.__.id, args);
		const wrapper = maybe_fn ? query$1.batch(validate_or_fn, fn) : query$1.batch(fn);
		return wrapper;
	};
	query.live = (validate_or_fn, maybe_fn) => {
		const fn = (arg) => {
			const payload = stringify_remote_arg(arg);
			let taken = false;
			return {
				next: async () => {
					if (taken) return {
						value: void 0,
						done: true
					};
					taken = true;
					return {
						value: await host(wrapper.__.id, payload),
						done: false
					};
				},
				return: () => ({
					value: void 0,
					done: true
				})
			};
		};
		const wrapper = maybe_fn ? query$1.live(validate_or_fn, fn) : query$1.live(fn);
		return wrapper;
	};
	function command(validate_or_fn, maybe_fn) {
		const fn = (arg) => host(wrapper.__.id, stringify_remote_arg(arg));
		const wrapper = maybe_fn ? /* @__PURE__ */ command$1(validate_or_fn, fn) : /* @__PURE__ */ command$1(fn);
		return wrapper;
	}
	var hello_remote_exports = /* @__PURE__ */ __exportAll({
		greet: () => greet,
		status: () => status
	});
	var unimplemented = () => {
		throw new Error("skgo: implemented in Go");
	};
	var greet = command("unchecked", (_arg) => unimplemented());
	var status = query(() => unimplemented());
	init_remote_functions(hello_remote_exports, "src/routes/hello.remote.ts", "1086sy4");
	for (const [name, fn] of Object.entries(hello_remote_exports)) {
		fn.__.id = "1086sy4/" + name;
		fn.__.name = name;
	}
	function _page$1($$renderer, $$props) {
		$$renderer.component(($$renderer) => {
			let name = "world";
			let busy = false;
			{
				function failed($$renderer, error) {
					$$renderer.push(`<p data-testid="status-failed">${escape_html(error.message)}</p>`);
				}
				$$renderer.boundary({ failed }, ($$renderer) => {
					$$renderer.push(`<!--[-->`);
					{
						let s;
						var promises = $$renderer.run([async () => s = (await save(status()))()]);
						$$renderer.push(`<h1 data-testid="title">`);
						$$renderer.async([promises[0]], ($$renderer) => $$renderer.push(() => escape_html(s.name)));
						$$renderer.push(`</h1> <p data-testid="answered-by">Served by `);
						$$renderer.async([promises[0]], ($$renderer) => $$renderer.push(() => escape_html(s.goVersion)));
						$$renderer.push(`.</p> <p>Greetings so far: <strong data-testid="greetings">`);
						$$renderer.async([promises[0]], ($$renderer) => $$renderer.push(() => escape_html(s.greetings)));
						$$renderer.push(`</strong></p> <p data-testid="last-greeting">Last greeting: `);
						$$renderer.async([promises[0]], ($$renderer) => $$renderer.push(() => escape_html(s.lastGreeting || "(none yet)")));
						$$renderer.push(`</p>`);
					}
					$$renderer.push(`<!--]-->`);
				});
			}
			$$renderer.push(` <form><input data-testid="name"${attr("value", name)}/> <button data-testid="greet" type="submit"${attr("disabled", busy, true)}>Greet</button></form>`);
		});
	}
	function _page($$renderer) {
		$$renderer.push(`<h1 data-testid="title">About</h1> <p>This page is plain SvelteKit. It calls nothing, and Go serves it out of the
	binary like every other route.</p>`);
	}
	var components = [
		_layout,
		_error,
		_page$1,
		_page
	];
	/**
	* The app's transport hook, installed the way kit installs it
	* (runtime/server/index.js: init_transport(module.transport ?? {})).
	*
	* Kit does it per request because it loads the hooks with a dynamic import;
	* the hooks are static and this bundle has no top-level await, so it happens
	* once per runtime instead. From here on, parse() is the app's own decoders
	* and a value Go serialized under a transport key comes back as an instance
	* of the class src/hooks.ts declares.
	*/
	init_transport({});
	/**
	* kit's handle_error_and_jsonify (runtime/server/errors.js) with no
	* handleError hook: an HttpError keeps the body the app chose, a framework
	* error keeps its status and text, a validation failure is a Bad Request, and
	* anything else is Internal Error with the real cause left on the server.
	* skgo has no handleError hook, so there is nothing here to await or merge.
	*/
	function handle_error(error) {
		if (error instanceof HttpError) return error.body;
		if (error instanceof SvelteKitError) return {
			status: error.status,
			message: error.text
		};
		if (error instanceof ValidationError) return {
			status: 400,
			message: "Bad Request"
		};
		return {
			status: 500,
			message: "Internal Error"
		};
	}
	/**
	* A RequestState (packages/kit/src/types/internal.d.ts). Only the fields
	* kit's remote wrappers read during a render are populated. is_in_render is
	* what makes kit refuse a command.
	*/
	function make_state() {
		return {
			getClientAddress: () => "127.0.0.1",
			error: false,
			rerouted_url: null,
			depth: 0,
			remote: {
				data: null,
				implicit: null,
				explicit: null,
				forms: null,
				requested: null,
				batches: null,
				live_iterators: null
			},
			is_in_remote_function: false,
			is_in_render: true
		};
	}
	/**
	* event.fetch during a render. Kit's own (runtime/server/fetch.js) resolves a
	* relative URL against the page's own origin, forwards the incoming cookies
	* on a same-origin request unless credentials is "omit", and otherwise hands
	* the request to a real fetch. This engine has no socket to hand a
	* cross-origin request to — no application I/O executes in JavaScript here —
	* so a same-origin URL is a call back into Go's own handler for it, exactly
	* the shape a remote function's call is, and a cross-origin one is refused: a
	* thrown TypeError, before anything is dispatched. (Kit's own version also
	* forwards the page's own Authorization header on a same-origin request; this
	* one does not, because the render's own request never reaches the engine
	* with its headers — only its cookies, in req.cookies.)
	*
	* @param {any} req
	* @param {URL} url
	*/
	function create_fetch(req, url) {
		return async function(input, init) {
			const request = input instanceof Request ? input : new Request(input, init);
			const target = new URL(request.url, url);
			if (!same_origin(target, url)) throw new TypeError("skgo: event.fetch only reaches this app's own routes during server-side rendering (" + origin_of(target) + " is not " + origin_of(url) + ")");
			if ((init?.credentials ?? request.credentials ?? "same-origin") !== "omit") {
				const cookie = Object.entries(req.cookies ?? {}).map(([name, value]) => name + "=" + value).join("; ");
				if (cookie) request.headers.set("cookie", cookie);
			}
			const headers = {};
			request.headers.forEach((value, name) => {
				headers[name] = value;
			});
			const envelope = {
				method: request.method || "GET",
				url: target.href,
				headers
			};
			if (typeof request._body === "string") envelope.body = request._body;
			const raw = await globalThis.__skgo_fetch(JSON.stringify(envelope));
			const answer = JSON.parse(raw);
			if (answer.error) throw new TypeError(answer.error);
			return Promise.resolve(new Response(answer.response.body ?? "", {
				status: answer.response.status,
				headers: answer.response.headers
			}));
		};
	}
	/**
	* Same-origin, spelled out rather than taken from `origin`. The engine's URL is
	* Go's (goja_nodejs, over net/url) and its `origin` is scheme and hostname with
	* the port left out, which would make a page on :8080 and a URL on :9999 look
	* like the same site — and this is the one comparison standing between a
	* cross-origin fetch and Go's own handler for the path. Protocol, hostname and
	* port are each exactly what the standard compares.
	*
	* @param {URL} a
	* @param {URL} b
	*/
	function same_origin(a, b) {
		return a.protocol === b.protocol && a.hostname === b.hostname && a.port === b.port;
	}
	/**
	* What `url.origin` would say if this engine's URL reported it the way the
	* standard does. Only the refusal message above needs it.
	*
	* @param {URL} url
	*/
	function origin_of(url) {
		return url.host ? url.protocol + "//" + url.host : "null";
	}
	/**
	* A RequestEvent stand-in. run_remote_function spreads it and derives the
	* event a query actually sees, which is where kit makes url, params and
	* route throw. Go owns the real request; the only I/O anything here performs
	* is `fetch`'s in-process call back into Go for one of the app's own routes.
	*/
	function make_event(req, url) {
		return {
			cookies: {
				get: (name) => (req.cookies ?? {})[name],
				getAll: () => Object.entries(req.cookies ?? {}).map(([name, value]) => ({
					name,
					value
				})),
				set: () => {},
				delete: () => {},
				serialize: () => ""
			},
			fetch: create_fetch(req, url),
			getClientAddress: () => req.client_address ?? "127.0.0.1",
			locals: {},
			params: req.params ?? {},
			platform: void 0,
			request: {
				headers: { get: () => null },
				method: "GET"
			},
			route: { id: req.route_id ?? null },
			setHeaders: () => {},
			url,
			isDataRequest: false,
			isSubRequest: false,
			isRemoteRequest: false,
			tracing: { enabled: false }
		};
	}
	/**
	* One node's load result. Go sends devalue's flat form — the same bytes it
	* sends the client in __data.json, produced by the same encoders — and the
	* app's decoders read it back, so the component renders against the instance
	* the browser is about to hold rather than the object its fields travelled in.
	*
	* A value the load promised is in those bytes as kit's own placeholder, and it
	* is read back the way kit's client reads it (process_stream in client.js): a
	* Promise reviver alongside the app's decoders. So a promise is found wherever
	* the load left one, at any depth, which is where devalue's reducer put it.
	*
	* The promise it becomes never settles. Svelte's server renderer does not await
	* an await block — it pushes the block marker and renders the pending branch
	* (svelte/src/internal/server/index.js, await_block) — so the document leaves Go
	* with the loading state already in it, and the value follows it down as a chunk
	* Go appends. That is exactly what kit does, which hands its renderer the
	* promise itself.
	*/
	function node_data(node) {
		if (!node.data) return null;
		return parse$1(node.data, {
			...decoders,
			Promise: () => new Promise(() => {})
		});
	}
	function build_props(req, url) {
		const page = {
			error: req.error ?? null,
			params: req.params ?? {},
			route: { id: req.route_id ?? null },
			status: req.status ?? 200,
			url,
			data: {},
			form: req.form ?? null,
			shallow: null,
			state: {}
		};
		const branch = req.branch ?? [];
		const error_components = (req.error_components ?? []).map((i) => i == null ? void 0 : components[i]);
		const props = new Props({
			page,
			tree: new RenderNode(components[branch[0].node], void 0),
			form: req.form ?? null,
			error: req.error ?? void 0
		});
		let current_node = props.tree;
		let data = props.page.data;
		for (let i = 0; i < branch.length; i += 1) {
			data = {
				...data,
				...node_data(branch[i])
			};
			current_node.data = data;
			if (i < branch.length - 1) current_node = current_node.child = new RenderNode(components[branch[i + 1].node], error_components[i + 1]);
		}
		props.page.data = data;
		return props;
	}
	/**
	* Puts a non-enhanced submission's outcome where kit's form instance reads
	* it: the request's remote cache, under the instance's own internals object
	* and the empty-string key.
	*
	* That is the last thing kit's form wrapper does after it runs a submission
	* (runtime/app/server/remote/form.js: get_cache(__, state)[''] ??= output),
	* and it is what makes myForm.result, myForm.fields.x.issues() and the value
	* of every control render the submission. Nothing here runs a form body — Go
	* already ran it — so the stubs still throw and a rendered result is still
	* proof that Go answered.
	*
	* The output arrives in devalue's flat form and is read back with the app's
	* own decoders, for the same reason a load's data is: a result carrying a
	* transported type has to reach the component as an instance of its class.
	*
	* A keyed instance, one created by calling for(key) on a form, needs one more
	* step than an unkeyed one. Go's req.form_action.id is kit's own composite
	* id: the base hash/name, or that plus a slash and the key's JSON text — the
	* same string kit's server files a form's output under in the page's remote
	* data — and only the base half is registered anywhere: __skgo_forms holds
	* the module-level instance kit's form factory created, with no key at all.
	* That instance's own for method is kit's own code
	* (runtime/app/server/remote/form.js) for turning a key into the actual
	* per-key instance the page's own call to for(key) will return — it caches
	* what it creates in the request's form cache, keyed by the base id and the
	* key's JSON text together, so calling it here, before the page component
	* runs, makes the page's later call resolve to the very instance seeded below
	* rather than a fresh, empty one. Calling for reaches into the request
	* store, which is why this function now has to run inside with_request_store
	* rather than before it.
	*/
	function seed_form(req, state) {
		const seed = req.form_action;
		if (!seed) return;
		const first_slash = seed.id.indexOf("/");
		const second_slash = seed.id.indexOf("/", first_slash + 1);
		const base_id = second_slash === -1 ? seed.id : seed.id.slice(0, second_slash);
		const key_json = second_slash === -1 ? void 0 : seed.id.slice(second_slash + 1);
		for (const instance of globalThis.__skgo_forms ?? []) {
			if (!instance.__ || instance.__.id !== base_id) continue;
			const target = key_json === void 0 ? instance : instance.for(JSON.parse(key_json));
			(state.remote.data ??= /* @__PURE__ */ new Map()).set(target.__, { "": parse(seed.output) });
			return;
		}
		throw new Error("skgo: no form is registered as " + seed.id);
	}
	/**
	* Renders one page. The result object is filled in as the promise chain
	* settles; the host drains the job queue when this call returns, so done is
	* true by then or the render never finished — which is a Go error, not a
	* partial document.
	*/
	globalThis.__skgo_render = function(req_json) {
		const result = {
			done: false,
			failure: "",
			redirect: null,
			status: 200,
			error: null,
			head: "",
			body: ""
		};
		try {
			const req = JSON.parse(req_json);
			const url = new URL(req.url);
			const props = build_props(req, url);
			const state = make_state();
			const event = make_event(req, url);
			result.status = props.page.status;
			result.error = req.error ?? null;
			const options = {
				context: /* @__PURE__ */ new Map([["__request__", { page: props.page }]]),
				csp: req.csp.nonce ? { nonce: req.csp.nonce } : { hash: !!req.csp.hash },
				transformError: (e) => {
					if (e instanceof Redirect) throw e;
					const handled = handle_error(e);
					result.error = handled;
					result.status = handled.status;
					props.page.error = handled;
					props.page.status = handled.status;
					return handled;
				}
			};
			const promise = with_request_store({
				event,
				state
			}, () => {
				seed_form(req, state);
				return render(Root, {
					...options,
					props
				});
			});
			Promise.resolve(promise).then((rendered) => {
				result.head = rendered.head;
				result.body = rendered.body;
				result.done = true;
			}, (err) => {
				if (err instanceof Redirect) result.redirect = {
					status: err.status,
					location: err.location
				};
				else result.failure = err && (err.stack || err.message) || String(err);
				result.done = true;
			});
		} catch (err) {
			if (err instanceof Redirect) result.redirect = {
				status: err.status,
				location: err.location
			};
			else result.failure = err && (err.stack || err.message) || String(err);
			result.done = true;
		}
		return result;
	};
	globalThis.__skgo_ping = function() {
		return "ok";
	};
	//#endregion
})();
