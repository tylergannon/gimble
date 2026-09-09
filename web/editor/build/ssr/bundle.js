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
	//#region build/.goja/assets/editor.remote-B_MFZRZC.js
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
	var require__skgo_skgo_missing = /* @__PURE__ */ __commonJSMin((() => {
		throw new Error("skgo: node:async_hooks is unavailable in the SSR engine");
	}));
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
	function stringify$1$1(value, reducers, options) {
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
	function noop$1() {}
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
	var als$1;
	Promise.resolve().then(() => /* @__PURE__ */ __toESM(require__skgo_skgo_missing(), 1)).then((hooks) => als$1 = new hooks.AsyncLocalStorage()).catch(() => {});
	function get_request_store() {
		const result = try_get_request_store();
		if (!result) {
			let message = "Could not get the request store.";
			if (als$1) message += " This is an internal error.";
			else message += " In environments without `AsyncLocalStorage`, the request store (used by e.g. remote functions) must be accessed synchronously, not after an `await`. If it was accessed synchronously then this is an internal error.";
			throw new Error(message);
		}
		return result;
	}
	function try_get_request_store() {
		return sync_store ?? als$1?.getStore() ?? null;
	}
	/**
	* @template T
	* @param {RequestStore | null} store
	* @param {() => T} fn
	*/
	function with_request_store(store, fn) {
		try {
			sync_store = store;
			return als$1 ? als$1.run(store, fn) : fn();
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
	var text_encoder$1 = new TextEncoder();
	new TextDecoder();
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
		const stringify = (value) => stringify$1$1(value, all_reducers);
		return all_reducers;
	}
	/**
	* Stringifies the argument (if any) for a remote function in such a way that
	* it is both a valid URL and a valid file name (necessary for prerendering).
	* @param {any} value
	*/
	function stringify_remote_arg(value) {
		if (value === void 0) return "";
		return url_friendly_base64_encode(stringify$1$1(value, create_remote_arg_reducers(true)));
	}
	/**
	* Base64-encodes `string` in such a way that the result is safe to use
	* as both a URI component and a filename
	* @param {string} string
	*/
	function url_friendly_base64_encode(string) {
		return base64_encode$1(text_encoder$1.encode(string)).replaceAll("=", "").replaceAll("+", "-").replaceAll("/", "_");
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
			if (__.id && state.is_in_render) get_promise().catch(noop$1);
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
			if (__.id && state.is_in_render) get_promise().catch(noop$1);
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
				return noop$1;
			}
			const generator = get_generator();
			let aborted = false;
			const close = () => {
				aborted = true;
				generator.return().catch(noop$1);
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
	var editor_remote_exports = /* @__PURE__ */ __exportAll({
		getDoc: () => getDoc$1,
		saveDoc: () => saveDoc,
		watchDoc: () => watchDoc
	});
	var unimplemented = () => {
		throw new Error("skgo: implemented in Go");
	};
	var getDoc$1 = query(() => unimplemented());
	var saveDoc = command("unchecked", (_arg) => unimplemented());
	var watchDoc = query.live(() => unimplemented());
	init_remote_functions(editor_remote_exports, "src/routes/editor.remote.ts", "1t7vxpr");
	for (const [name, fn] of Object.entries(editor_remote_exports)) {
		fn.__.id = "1t7vxpr/" + name;
		fn.__.name = name;
	}
	//#endregion
	//#region build/.goja/bundle.js
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
	var object_prototype = Object.prototype;
	Array.prototype;
	var get_prototype_of = Object.getPrototypeOf;
	var has_own_property = Object.prototype.hasOwnProperty;
	var noop = () => {};
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
		if (value == null || !value && is_boolean) return "";
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
	/**
	* 'Transparent' effects do not create a transition boundary.
	* This is on a block effect 99% of the time but may also be on a branch effect if its parent block effect was pruned
	*/
	var EFFECT_TRANSPARENT = 65536;
	var EFFECT_PRESERVED = 1 << 19;
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
	/** @import { Snapshot } from './types' */
	/**
	* In dev, we keep track of which properties could not be cloned. In prod
	* we don't bother, but we keep a dummy array around so that the
	* signature stays the same
	* @type {string[]}
	*/
	var empty = [];
	/**
	* @template T
	* @param {T} value
	* @param {boolean} [skip_warning]
	* @param {boolean} [no_tojson]
	* @returns {Snapshot<T>}
	*/
	function snapshot(value, skip_warning = false, no_tojson = false) {
		return clone(value, /* @__PURE__ */ new Map(), "", empty, null, no_tojson);
	}
	/**
	* @template T
	* @param {T} value
	* @param {Map<T, Snapshot<T>>} cloned
	* @param {string} path
	* @param {string[]} paths
	* @param {null | T} [original] The original value, if `value` was produced from a `toJSON` call
	* @param {boolean} [no_tojson]
	* @returns {Snapshot<T>}
	*/
	function clone(value, cloned, path, paths, original = null, no_tojson = false) {
		if (typeof value === "object" && value !== null) {
			var unwrapped = cloned.get(value);
			if (unwrapped !== void 0) return unwrapped;
			if (value instanceof Map) return new Map(value);
			if (value instanceof Set) return new Set(value);
			if (is_array(value)) {
				var copy = Array(value.length);
				cloned.set(value, copy);
				if (original !== null) cloned.set(original, copy);
				for (var i = 0; i < value.length; i += 1) {
					var element = value[i];
					if (i in value) copy[i] = clone(element, cloned, path, paths, null, no_tojson);
				}
				return copy;
			}
			if (get_prototype_of(value) === object_prototype) {
				/** @type {Snapshot<any>} */
				copy = {};
				cloned.set(value, copy);
				if (original !== null) cloned.set(original, copy);
				for (var key of Object.keys(value)) copy[key] = clone(value[key], cloned, path, paths, null, no_tojson);
				return copy;
			}
			if (value instanceof Date) return structuredClone(value);
			if (typeof value.toJSON === "function" && !no_tojson) return clone(
				/** @type {T & { toJSON(): any } } */
				value.toJSON(),
				cloned,
				path,
				paths,
				value
			);
		}
		if (value instanceof EventTarget) return value;
		try {
			return structuredClone(value);
		} catch (e) {
			return value;
		}
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
	EFFECT_TRANSPARENT | EFFECT_PRESERVED;
	var BLOCK_OPEN = `<!--[-->`;
	var BLOCK_CLOSE = `<!--]-->`;
	var EMPTY_COMMENT = `<!---->`;
	var VOID_ELEMENT_NAMES = [
		"area",
		"base",
		"br",
		"col",
		"command",
		"embed",
		"hr",
		"img",
		"input",
		"keygen",
		"link",
		"meta",
		"param",
		"source",
		"track",
		"wbr"
	];
	/**
	* Returns `true` if `name` is of a void element
	* @param {string} name
	*/
	function is_void(name) {
		return VOID_ELEMENT_NAMES.includes(name) || name.toLowerCase() === "!doctype";
	}
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
	/** List of elements that require raw contents and should not have SSR comments put in them */
	var RAW_TEXT_ELEMENTS = [
		"textarea",
		"script",
		"style",
		"title"
	];
	/** @param {string} name */
	function is_raw_text_element(name) {
		return RAW_TEXT_ELEMENTS.includes(name);
	}
	var REGEX_VALID_TAG_NAME = /^[a-zA-Z][a-zA-Z0-9]*(-[a-zA-Z0-9.\-_\u00B7\u00C0-\u00D6\u00D8-\u00F6\u00F8-\u037D\u037F-\u1FFF\u200C-\u200D\u203F-\u2040\u2070-\u218F\u2C00-\u2FEF\u3001-\uD7FF\uF900-\uFDCF\uFDF0-\uFFFD\u{10000}-\u{EFFFF}]*)?$/u;
	/** @type {AbortController | null} */
	var controller = null;
	function abort() {
		controller?.abort(STALE_REACTION);
		controller = null;
	}
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
	* `<svelte:element this="%tag%">` is not a valid element name — the element will not be rendered
	* @param {string} tag
	* @returns {never}
	*/
	function dynamic_element_invalid_tag(tag) {
		const error = /* @__PURE__ */ new Error(`dynamic_element_invalid_tag\n\`<svelte:element this="${tag}">\` is not a valid element name — the element will not be rendered\nhttps://svelte.dev/e/dynamic_element_invalid_tag`);
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
	/** @import { AsyncLocalStorage } from 'node:async_hooks' */
	/** @import { RenderContext } from '#server' */
	/** @type {Promise<void> | null} */
	var current_render = null;
	/** @type {RenderContext | null} */
	var context$1 = null;
	/** @returns {RenderContext} */
	function get_render_context() {
		const store = context$1 ?? als?.getStore();
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
			if (als === null) async_local_storage_unavailable();
			return als.run(context$1, fn);
		} finally {
			context$1 = null;
		}
	}
	/** @type {AsyncLocalStorage<RenderContext | null> | null} */
	var als = null;
	/** @type {Promise<void> | null} */
	var als_import = null;
	/**
	*
	* @returns {Promise<void>}
	*/
	function init_render_context() {
		als_import ??= Promise.resolve().then(() => /* @__PURE__ */ __toESM(require__skgo_skgo_missing(), 1)).then((hooks) => {
			als = new hooks.AsyncLocalStorage();
		}).then(noop, noop);
		return als_import;
	}
	function in_webcontainer() {
		return !!globalThis.process?.versions?.webcontainer;
	}
	var text_encoder;
	var crypto;
	/** @param {string} module_name */
	var obfuscated_import = (module_name) => Promise.reject(/* @__PURE__ */ new Error("skgo: no dynamic import in the SSR engine"));
	/** @param {string} data */
	async function sha256(data) {
		text_encoder ??= new TextEncoder();
		crypto ??= globalThis.crypto?.subtle?.digest ? globalThis.crypto : (await obfuscated_import("node:crypto")).webcrypto;
		return base64_encode(await crypto.subtle.digest("SHA-256", text_encoder.encode(data)));
	}
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
	/** @import { Component } from 'svelte' */
	/** @import { Csp, HydratableContext, RenderOutput, SSRContext, SyncRenderOutput, Sha256Source } from './types.js' */
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
		* @type {{ select_value: string | undefined }}
		*/
		local;
		/**
		* @param {SSRState} global
		* @param {Renderer | undefined} [parent]
		*/
		constructor(global, parent) {
			this.#parent = parent;
			this.global = global;
			this.local = parent ? { ...parent.local } : { select_value: void 0 };
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
			promise.catch(noop);
			this.promise = promise;
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
				result.catch(noop);
				result.finally(() => set_ssr_context(null)).catch(noop);
				if (child.global.mode === "sync") await_invalid();
				child.promise = result;
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
					result.catch(noop);
					child.promise = result;
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
					child.promise = result.then((transformed) => {
						set_ssr_context(parent_context);
						child.#out.push(Renderer.#serialize_failed_boundary(transformed));
						failed_snippet(child, transformed, noop);
						child.#out.push(BLOCK_CLOSE);
					});
					child.promise.catch(noop);
				} else {
					child.#out.push(Renderer.#serialize_failed_boundary(result));
					failed_snippet(child, result, noop);
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
			const child = this.child(fn);
			child.#is_component_body = true;
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
			const { value, ...select_attrs } = attrs;
			this.push(`<select${attributes(select_attrs, css_hash, classes, styles, flags)}>`);
			this.child((renderer) => {
				renderer.local.select_value = value;
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
				if (value === this.local.select_value) renderer.#out.push(" selected=\"\"");
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
		* Render a component. Throws if any of the children are performing asynchronous work.
		*
		* @template {Record<string, any>} Props
		* @param {Component<Props>} component
		* @param {{ props?: Omit<Props, '$$slots' | '$$events'>; context?: Map<any, any>; idPrefix?: string }} options
		* @returns {AccumulatedContent}
		*/
		static #render(component, options) {
			var previous_context = ssr_context;
			try {
				const renderer = Renderer.#open_render("sync", component, options);
				const content = renderer.#collect_content();
				return Renderer.#close_render(content, renderer);
			} finally {
				abort();
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
			try {
				const renderer = Renderer.#open_render("async", component, options);
				const content = await renderer.#collect_content_async();
				const hydratables = await renderer.#collect_hydratables();
				if (hydratables !== null) content.head = hydratables + content.head;
				return Renderer.#close_render(content, renderer);
			} finally {
				set_ssr_context(previous_context);
				abort();
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
						failed(failed_renderer, transformed, noop);
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
		* @param {'sync' | 'async'} mode
		* @param {import('svelte').Component<Props>} component
		* @param {{ props?: Omit<Props, '$$slots' | '$$events'>; context?: Map<any, any>; idPrefix?: string; csp?: Csp; transformError?: (error: unknown) => unknown }} options
		* @returns {Renderer}
		*/
		static #open_render(mode, component, options) {
			if (options.idPrefix?.includes("--")) invalid_id_prefix();
			var previous_context = ssr_context;
			try {
				const renderer = new Renderer(new SSRState(mode, options.idPrefix ? options.idPrefix + "-" : "", options.csp, options.transformError));
				set_ssr_context({
					p: null,
					c: options.context ?? null,
					r: renderer
				});
				renderer.push(BLOCK_OPEN);
				component(renderer, options.props ?? {});
				renderer.push(BLOCK_CLOSE);
				return renderer;
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
			for (const cleanup of renderer.#collect_on_destroy()) cleanup();
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
	* @param {Renderer} renderer
	* @param {string} tag
	* @param {() => void} attributes_fn
	* @param {() => void} children_fn
	* @returns {void}
	*/
	function element(renderer, tag, attributes_fn = noop, children_fn = noop) {
		renderer.push("<!---->");
		if (tag) {
			if (!REGEX_VALID_TAG_NAME.test(tag)) dynamic_element_invalid_tag(tag);
			renderer.push(`<${tag}`);
			attributes_fn();
			renderer.push(`>`);
			if (!is_void(tag)) {
				children_fn();
				if (!is_raw_text_element(tag)) renderer.push(EMPTY_COMMENT);
				renderer.push(`</${tag}>`);
			}
		}
		renderer.push("<!---->");
	}
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
	* @param {string} hash
	* @param {Renderer} renderer
	* @param {(renderer: Renderer) => Promise<void> | void} fn
	* @returns {void}
	*/
	function head(hash, renderer, fn) {
		renderer.head((renderer) => {
			renderer.push(`<!--${hash}-->`);
			renderer.child(fn);
			renderer.push(EMPTY_COMMENT);
		});
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
	* @param {Record<string, unknown>[]} props
	* @returns {Record<string, unknown>}
	*/
	function spread_props(props) {
		/** @type {Record<string, unknown>} */
		const merged_props = {};
		let key;
		for (let i = 0; i < props.length; i++) {
			const obj = props[i];
			if (obj == null) continue;
			for (key of Object.keys(obj)) {
				const desc = Object.getOwnPropertyDescriptor(obj, key);
				if (desc) Object.defineProperty(merged_props, key, desc);
				else merged_props[key] = obj[key];
			}
		}
		return merged_props;
	}
	/**
	* @param {unknown} value
	* @returns {string}
	*/
	function stringify$1(value) {
		return typeof value === "string" ? value : value == null ? "" : value + "";
	}
	/**
	* @param {any} value
	* @param {string | undefined} [hash]
	* @param {Record<string, boolean>} [directives]
	*/
	function attr_class(value, hash, directives) {
		var result = to_class(value, hash, directives);
		return result ? ` class="${escape_html(result, true)}"` : "";
	}
	/**
	* @param {any} value
	* @param {Record<string,any>|[Record<string,any>,Record<string,any>]} [directives]
	*/
	function attr_style(value, directives) {
		var result = to_style(value, directives);
		return result ? ` style="${escape_html(result, true)}"` : "";
	}
	/** @param {any} array_like_or_iterator */
	function ensure_array_like(array_like_or_iterator) {
		if (array_like_or_iterator) return array_like_or_iterator.length !== void 0 ? array_like_or_iterator : Array.from(array_like_or_iterator);
		return [];
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
	var afterNavigate = noop$1;
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
				$$renderer.push("<!--[0-->");
				$$renderer.push(`<div id="svelte-announcer" aria-live="assertive" aria-atomic="true" style="position: absolute; left: 0; top: 0; clip: rect(0 0 0 0); clip-path: inset(50%); overflow: hidden; white-space: nowrap; width: 1px; height: 1px">`);
				if (navigated) {
					$$renderer.push("<!--[0-->");
					$$renderer.push(`${escape_html(title)}`);
				} else $$renderer.push("<!--[-1-->");
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
		constructor({ page, tree, form, error, onerror = noop$1 }) {
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
	var favicon_default = "data:image/svg+xml,%3csvg%20xmlns='http://www.w3.org/2000/svg'%20viewBox='0%200%20128%20128'%20role='img'%20aria-label='Gimble'%3e%3crect%20width='128'%20height='128'%20rx='28'%20fill='%23102D35'/%3e%3cg%20fill='none'%20stroke='%23E0A13A'%20stroke-width='9'%20stroke-linecap='round'%3e%3ccircle%20cx='64'%20cy='64'%20r='43'/%3e%3cellipse%20cx='64'%20cy='64'%20rx='27'%20ry='48'%20transform='rotate(36%2064%2064)'/%3e%3c/g%3e%3ccircle%20cx='64'%20cy='64'%20r='24'%20fill='%23FFF4DF'%20stroke='%23245B67'%20stroke-width='6'/%3e%3cpath%20d='m64%2042%2012%2027-12-6-12%206z'%20fill='%23D55B3A'/%3e%3cpath%20d='m64%2086-12-27%2012%206%2012-6z'%20fill='%232B7781'/%3e%3c/svg%3e";
	function _layout($$renderer, $$props) {
		let { children } = $$props;
		head("12qhfyh", $$renderer, ($$renderer) => {
			$$renderer.push(`<link rel="icon"${attr("href", favicon_default)}/>`);
		});
		children($$renderer);
		$$renderer.push(`<!---->`);
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
	function Error$1($$renderer, $$props) {
		$$renderer.component(($$renderer) => {
			$$renderer.push(`<h1>${escape_html(page.status)}</h1> <p>${escape_html(page.error?.message)}</p>`);
		});
	}
	/**
	* @file
	* @license @lucide/svelte v1.41.0 - ISC
	*
	* This source code is licensed under the ISC license.
	* See the LICENSE file in the root directory of this source tree.
	*/
	var defaultAttributes = {
		xmlns: "http://www.w3.org/2000/svg",
		width: 24,
		height: 24,
		viewBox: "0 0 24 24",
		fill: "none",
		stroke: "currentColor",
		"stroke-width": 2,
		"stroke-linecap": "round",
		"stroke-linejoin": "round"
	};
	/**
	* @file
	* @license @lucide/svelte v1.41.0 - ISC
	*
	* This source code is licensed under the ISC license.
	* See the LICENSE file in the root directory of this source tree.
	*/
	/**
	* Check if a component has an accessibility prop
	*
	* @param {object} props
	* @returns {boolean} Whether the component has an accessibility prop
	*/
	var hasA11yProp = (props) => {
		for (const prop in props) if (prop.startsWith("aria-") || prop === "role" || prop === "title") return true;
		return false;
	};
	/**
	* @file
	* @license @lucide/svelte v1.41.0 - ISC
	*
	* This source code is licensed under the ISC license.
	* See the LICENSE file in the root directory of this source tree.
	*/
	var LucideContext = Symbol("lucide-context");
	var getLucideContext = () => getContext(LucideContext);
	function Icon($$renderer, $$props) {
		$$renderer.component(($$renderer) => {
			const globalProps = getLucideContext() ?? {};
			const { name, color = globalProps.color ?? "currentColor", size = globalProps.size ?? 24, strokeWidth = globalProps.strokeWidth ?? 2, absoluteStrokeWidth = globalProps.absoluteStrokeWidth ?? false, iconNode = [], children, $$slots, $$events, ...props } = $$props;
			const calculatedStrokeWidth = derived(() => absoluteStrokeWidth ? Number(strokeWidth) * 24 / Number(size) : strokeWidth);
			$$renderer.push(`<svg${attributes({
				...defaultAttributes,
				...!children && !hasA11yProp(props) && { "aria-hidden": "true" },
				...props,
				width: size,
				height: size,
				stroke: color,
				"stroke-width": calculatedStrokeWidth(),
				class: clsx([
					"lucide-icon lucide",
					globalProps.class,
					name && `lucide-${name}`,
					props.class
				])
			}, void 0, void 0, void 0, 3)}><!--[-->`);
			const each_array = ensure_array_like(iconNode);
			for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
				let [tag, attrs] = each_array[$$index];
				element($$renderer, tag, () => {
					$$renderer.push(`${attributes({ ...attrs }, void 0, void 0, void 0, 3)}`);
				});
			}
			$$renderer.push(`<!--]-->`);
			children?.($$renderer);
			$$renderer.push(`<!----></svg>`);
		});
	}
	function Chevron_down($$renderer, $$props) {
		let { $$slots, $$events, ...props } = $$props;
		Icon($$renderer, spread_props([
			{ name: "chevron-down" },
			props,
			{ iconNode: [["path", { "d": "m6 9 6 6 6-6" }]] }
		]));
	}
	function Minus($$renderer, $$props) {
		let { $$slots, $$events, ...props } = $$props;
		Icon($$renderer, spread_props([
			{ name: "minus" },
			props,
			{ iconNode: [["path", { "d": "M5 12h14" }]] }
		]));
	}
	function Moon($$renderer, $$props) {
		let { $$slots, $$events, ...props } = $$props;
		Icon($$renderer, spread_props([
			{ name: "moon" },
			props,
			{ iconNode: [["path", { "d": "M20.985 12.486a9 9 0 1 1-9.473-9.472c.405-.022.617.46.402.803a6 6 0 0 0 8.268 8.268c.344-.215.825-.004.803.401" }]] }
		]));
	}
	function Plus($$renderer, $$props) {
		let { $$slots, $$events, ...props } = $$props;
		Icon($$renderer, spread_props([
			{ name: "plus" },
			props,
			{ iconNode: [["path", { "d": "M5 12h14" }], ["path", { "d": "M12 5v14" }]] }
		]));
	}
	function Sun($$renderer, $$props) {
		let { $$slots, $$events, ...props } = $$props;
		Icon($$renderer, spread_props([
			{ name: "sun" },
			props,
			{ iconNode: [
				["circle", {
					"cx": "12",
					"cy": "12",
					"r": "4"
				}],
				["path", { "d": "M12 2v2" }],
				["path", { "d": "M12 20v2" }],
				["path", { "d": "m4.93 4.93 1.41 1.41" }],
				["path", { "d": "m17.66 17.66 1.41 1.41" }],
				["path", { "d": "M2 12h2" }],
				["path", { "d": "M20 12h2" }],
				["path", { "d": "m6.34 17.66-1.41 1.41" }],
				["path", { "d": "m19.07 4.93-1.41 1.41" }]
			] }
		]));
	}
	function Trash($$renderer, $$props) {
		let { $$slots, $$events, ...props } = $$props;
		Icon($$renderer, spread_props([
			{ name: "trash" },
			props,
			{ iconNode: [
				["path", { "d": "M10 11v6" }],
				["path", { "d": "M14 11v6" }],
				["path", { "d": "M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6" }],
				["path", { "d": "M3 6h18" }],
				["path", { "d": "M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" }]
			] }
		]));
	}
	function Triangle_alert($$renderer, $$props) {
		let { $$slots, $$events, ...props } = $$props;
		Icon($$renderer, spread_props([
			{ name: "triangle-alert" },
			props,
			{ iconNode: [
				["path", { "d": "m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3" }],
				["path", { "d": "M12 9v4" }],
				["path", { "d": "M12 17h.01" }]
			] }
		]));
	}
	function X($$renderer, $$props) {
		let { $$slots, $$events, ...props } = $$props;
		Icon($$renderer, spread_props([
			{ name: "x" },
			props,
			{ iconNode: [["path", { "d": "M18 6 6 18" }], ["path", { "d": "m6 6 12 12" }]] }
		]));
	}
	var ID_RE = /^[A-Za-z_][A-Za-z0-9_]*$/;
	var DUR_RE = /^[0-9]+(ms|s|m|h|d)$/;
	var EFFORT = [
		"low",
		"medium",
		"high"
	];
	var FIDELITY = [
		"full",
		"compacted",
		"none"
	];
	var WORKSPACE = ["isolated", "shared"];
	var TERMINAL = ["success", "failure"];
	var NODE_TYPES = [
		"agent",
		"fan_out",
		"fan_in",
		"command",
		"supervisor",
		"loop"
	];
	var TYPE_DESC = {
		agent: "Runs an LLM task.",
		fan_out: "Concurrently walks each branch.",
		fan_in: "Evaluates branch evidence with an LLM turn.",
		command: "Executes one shell command.",
		supervisor: "Observes nodes and coaches them outside the walk.",
		loop: "Iterates a checklist file until done."
	};
	var MODEL = [
		"model.name",
		"model.version",
		"model.effort"
	];
	var LLM = [...MODEL, "fidelity"];
	var LIMITS = [
		"timeout",
		"max_retries",
		"max_visits",
		"thread_id"
	];
	var FIELD_SECTIONS = {
		agent: [
			["Task", ["label", "prompt"]],
			["Model", LLM],
			["Limits", LIMITS]
		],
		fan_in: [
			["Task", ["label", "prompt"]],
			["Model", LLM],
			["Limits", LIMITS]
		],
		fan_out: [
			["Task", [
				"label",
				"prompt",
				"workspace",
				"max_parallel"
			]],
			["Model", LLM],
			["Limits", LIMITS]
		],
		command: [["Command", ["label", "command"]], ["Limits", ["timeout", "max_visits"]]],
		supervisor: [
			["Coaching", [
				"label",
				"prompt",
				"interval"
			]],
			["Model", MODEL],
			["Limits", ["timeout"]]
		],
		loop: [
			["Loop", ["label", "checklist"]],
			["Item judge", [
				"item_judge.model.name",
				"item_judge.model.version",
				"item_judge.model.effort"
			]],
			["Goal evaluator", [
				"goal_evaluator.model.name",
				"goal_evaluator.model.version",
				"goal_evaluator.model.effort"
			]],
			["Limits", ["timeout", "max_visits"]]
		]
	};
	var FIELD_META = {
		label: { label: "Label" },
		prompt: {
			label: "Prompt",
			kind: "textarea",
			rows: 8
		},
		command: {
			label: "Command",
			kind: "textarea",
			rows: 3,
			mono: true,
			placeholder: "go test ./..."
		},
		checklist: {
			label: "Checklist path",
			mono: true,
			placeholder: "docs/CHECKLIST.md"
		},
		workspace: {
			label: "Workspace",
			kind: "enum",
			options: WORKSPACE
		},
		max_parallel: {
			label: "Max parallel",
			kind: "int"
		},
		"model.name": {
			label: "Model name",
			inherit: true,
			mono: true
		},
		"model.version": {
			label: "Version (optional)",
			inherit: true,
			mono: true
		},
		"model.effort": {
			label: "Reasoning effort",
			kind: "enum",
			options: EFFORT,
			inherit: true
		},
		fidelity: {
			label: "Fidelity",
			kind: "enum",
			options: FIDELITY,
			inherit: true
		},
		"item_judge.model.name": {
			label: "Model name",
			mono: true,
			placeholder: "flash"
		},
		"item_judge.model.version": {
			label: "Version (optional)",
			mono: true
		},
		"item_judge.model.effort": {
			label: "Reasoning effort",
			kind: "enum",
			options: EFFORT
		},
		"goal_evaluator.model.name": {
			label: "Model name",
			inherit: true,
			mono: true
		},
		"goal_evaluator.model.version": {
			label: "Version (optional)",
			inherit: true,
			mono: true
		},
		"goal_evaluator.model.effort": {
			label: "Reasoning effort",
			kind: "enum",
			options: EFFORT,
			inherit: true
		},
		timeout: {
			label: "Timeout",
			kind: "duration",
			inherit: true,
			placeholder: "30m"
		},
		interval: {
			label: "Interval",
			kind: "duration",
			placeholder: "5m"
		},
		max_retries: {
			label: "Max retries",
			kind: "int",
			inherit: true
		},
		max_visits: {
			label: "Max visits",
			kind: "int"
		},
		thread_id: {
			label: "Thread id",
			mono: true
		}
	};
	var REQUIRED = {
		command: ["command", "edges.success"],
		supervisor: ["prompt", "supervises"],
		loop: ["edges.loop", "edges.exit"],
		fan_out: ["branches"]
	};
	var DEFAULT_KEYS = [
		"model.name",
		"model.version",
		"model.effort",
		"fidelity",
		"timeout",
		"max_retries"
	];
	function isTerminal(id) {
		return id === "success" || id === "failure";
	}
	function getPath(obj, key) {
		let cur = obj;
		for (const part of key.split(".")) {
			if (cur === null || typeof cur !== "object") return void 0;
			cur = cur[part];
		}
		return cur;
	}
	function routeEdges(n) {
		const e = n.edges;
		return e && !Array.isArray(e) && typeof e === "object" ? e : {};
	}
	function listEdges(n) {
		const e = n.type === "fan_out" ? n.branch_edges : n.edges;
		return Array.isArray(e) ? e.filter((x) => x && typeof x === "object") : [];
	}
	function edgesKey(n) {
		return n.type === "fan_out" ? "branch_edges" : "edges";
	}
	function linksOf(n) {
		switch (n.type) {
			case "command": {
				const e = routeEdges(n);
				return [{
					to: e.success ?? "",
					label: "success"
				}, {
					to: e.error ?? "",
					label: "error"
				}];
			}
			case "loop": {
				const e = routeEdges(n);
				return [{
					to: e.loop ?? "",
					label: "loop"
				}, {
					to: e.exit ?? "",
					label: "exit"
				}];
			}
			case "supervisor": return (n.supervises ?? []).map((to) => ({
				to: String(to),
				label: "",
				dashed: true
			}));
			default: return listEdges(n).map((e) => ({
				to: String(e.to ?? ""),
				label: e.condition ?? ""
			}));
		}
	}
	function normalizeBranch(b) {
		return typeof b === "string" ? {
			id: b,
			artifacts: []
		} : b;
	}
	function validate(g) {
		const ids = g.nodes.map((n) => n.id);
		const errs = {};
		const push = (id, m) => (errs[id] = errs[id] || []).push(m);
		const known = (x) => ids.includes(x) || isTerminal(x);
		g.nodes.forEach((n) => {
			const id = String(n.id ?? "");
			if (!ID_RE.test(id)) push(id, "id must match ^[A-Za-z_][A-Za-z0-9_]*$");
			if (isTerminal(id)) push(id, "id may not be success or failure");
			if (ids.filter((x) => x === id).length > 1) push(id, "duplicate id");
			if (!NODE_TYPES.includes(n.type)) push(id, `unknown type "${n.type}"`);
			for (const k of ["timeout", "interval"]) {
				const v = n[k];
				if (v !== void 0 && v !== "" && !DUR_RE.test(String(v))) push(id, `${k} must be an integer followed by ms, s, m, h, or d`);
			}
			for (const k of REQUIRED[n.type] ?? []) {
				const v = getPath(n, k);
				if (v === void 0 || v === null || v === "" || Array.isArray(v) && !v.length) push(id, `${k} is required`);
			}
			linksOf(n).forEach((l) => {
				if (l.to && !known(l.to)) push(id, `target "${l.to}" does not exist`);
			});
			if (n.type === "supervisor") (n.supervises ?? []).forEach((x) => {
				if (x === id) push(id, "cannot supervise itself");
			});
			if (n.type === "fan_out") (n.branches ?? []).forEach((b) => {
				if (typeof b === "object" && (!b.id || !(b.artifacts ?? []).length)) push(id, `branch ${b.id || "(unnamed)"} needs an id and at least one artifact`);
			});
			const selections = [
				n.model,
				n.item_judge?.model,
				n.goal_evaluator?.model
			];
			for (const selection of selections) {
				if (!selection) continue;
				if (!selection.name?.trim()) push(id, "model.name is required when a model selection is present");
				if (selection.version !== void 0 && !selection.version.trim()) push(id, "model.version must be nonblank when present");
				if (selection.effort !== void 0 && !EFFORT.includes(selection.effort)) push(id, "model.effort must be low, medium, or high");
			}
		});
		return errs;
	}
	var X0 = 60;
	var Y0 = 80;
	var COL = 340;
	var ROW = 170;
	function isBox(l) {
		return !!l && typeof l.w === "number" && typeof l.h === "number";
	}
	function edgeKey(from, index) {
		return `e:${from}>${index}`;
	}
	function autoLayout(g, saved) {
		const out = { ...saved };
		const nodes = g.nodes.filter((n) => n.id);
		const byId = new Map(nodes.map((n) => [n.id, n]));
		const walkers = nodes.filter((n) => n.type !== "supervisor");
		const depth = /* @__PURE__ */ new Map();
		if (g.start && byId.has(g.start) && byId.get(g.start).type !== "supervisor") {
			depth.set(g.start, 0);
			const queue = [g.start];
			while (queue.length) {
				const id = queue.shift();
				const d = depth.get(id);
				for (const l of linksOf(byId.get(id))) {
					if (!l.to || isTerminal(l.to) || !byId.has(l.to) || depth.has(l.to)) continue;
					if (byId.get(l.to).type === "supervisor") continue;
					depth.set(l.to, d + 1);
					queue.push(l.to);
				}
			}
		}
		let maxDepth = -1;
		for (const d of depth.values()) maxDepth = Math.max(maxDepth, d);
		walkers.filter((n) => !depth.has(n.id)).forEach((n) => depth.set(n.id, maxDepth + 1));
		const rows = /* @__PURE__ */ new Map();
		let bottom = Y0;
		for (const n of walkers) {
			const col = depth.get(n.id) ?? 0;
			const row = rows.get(col) ?? 0;
			rows.set(col, row + 1);
			const y = Y0 + row * ROW;
			bottom = Math.max(bottom, y + 120);
			if (!isBox(out[n.id])) out[n.id] = {
				x: X0 + col * COL,
				y,
				w: 240,
				h: 120
			};
		}
		let supIndex = 0;
		for (const n of nodes) {
			if (n.type !== "supervisor") continue;
			if (!isBox(out[n.id])) out[n.id] = {
				x: X0 + supIndex * COL,
				y: bottom + ROW - 120,
				w: 240,
				h: 120
			};
			supIndex++;
		}
		if (!isBox(out.success) || !isBox(out.failure)) {
			const boxes = nodes.map((n) => out[n.id]).filter(isBox);
			const mx = boxes.length ? Math.max(...boxes.map((l) => l.x + l.w)) : X0;
			const my = boxes.length ? Math.min(...boxes.map((l) => l.y)) : Y0;
			if (!isBox(out.success)) out.success = {
				x: mx + 160,
				y: my + 36,
				w: 150,
				h: 56
			};
			if (!isBox(out.failure)) out.failure = {
				x: mx + 160,
				y: my + 136,
				w: 150,
				h: 56
			};
		}
		return out;
	}
	function boundsOf(layout) {
		const ls = Object.values(layout).filter(isBox);
		if (!ls.length) return {
			x: 0,
			y: 0,
			w: 1e3,
			h: 600
		};
		const x0 = Math.min(...ls.map((l) => l.x)) - 200;
		const y0 = Math.min(...ls.map((l) => l.y)) - 200;
		return {
			x: x0,
			y: y0,
			w: Math.max(...ls.map((l) => l.x + l.w)) + 200 - x0,
			h: Math.max(...ls.map((l) => l.y + l.h)) + 200 - y0
		};
	}
	var FG = "var(--foreground)";
	var MUTED = "var(--muted-foreground)";
	var SUP = "var(--chart-2)";
	var IN = "var(--chart-3)";
	function anchor(b, p) {
		const cx = b.x + b.w / 2;
		const cy = b.y + b.h / 2;
		if (p.x < b.x) return [b.x, Math.max(b.y + 12, Math.min(b.y + b.h - 12, p.y))];
		if (p.x > b.x + b.w) return [b.x + b.w, Math.max(b.y + 12, Math.min(b.y + b.h - 12, p.y))];
		return p.y < cy ? [cx, b.y] : [cx, b.y + b.h];
	}
	function quad(a, via, b) {
		const cx = 2 * via.x - (a[0] + b[0]) / 2;
		const cy = 2 * via.y - (a[1] + b[1]) / 2;
		return `M${a[0]} ${a[1]} Q${cx} ${cy} ${b[0]} ${b[1]}`;
	}
	function buildScene(g, layout, sel, errs, view, showInherited, models) {
		const d = g.defaults ?? {};
		const paths = [];
		const labels = [];
		const handles = [];
		g.nodes.forEach((n) => {
			const L = layout[n.id];
			if (!isBox(L)) return;
			const links = linksOf(n).filter((l) => l.to);
			links.forEach((l, i) => {
				const isOut = sel === n.id;
				const isIn = !isOut && sel === l.to;
				const selected = isOut || isIn;
				const T = layout[l.to];
				if (!isBox(T)) return;
				const key = edgeKey(n.id, i);
				const W = layout[key];
				let dpath;
				let mid;
				let handle;
				if (l.dashed) {
					const P0 = W ?? {
						x: (L.x + L.w / 2 + T.x + T.w / 2) / 2,
						y: Math.min(L.y, T.y + T.h) - 40
					};
					dpath = quad(anchor(L, P0), P0, anchor(T, P0));
					mid = null;
					handle = P0;
				} else if (W) {
					dpath = quad(anchor(L, W), W, anchor(T, W));
					mid = [W.x, W.y];
					handle = W;
				} else if (T.x < L.x + L.w - 20) {
					const P0 = {
						x: (L.x + L.w / 2 + T.x + T.w / 2) / 2,
						y: Math.max(L.y + L.h, T.y + T.h) + 70
					};
					dpath = quad([L.x + L.w / 2, L.y + L.h], P0, [T.x + T.w / 2, T.y + T.h]);
					mid = [P0.x, P0.y];
					handle = P0;
				} else {
					const sy = L.y + L.h / 2 + (i - (links.length - 1) / 2) * 16;
					const sx = L.x + L.w;
					let tx = T.x;
					let ty = T.y + T.h / 2;
					if (T.x < L.x + L.w && T.x + T.w > L.x) {
						tx = T.x + T.w / 2;
						ty = T.y > L.y ? T.y : T.y + T.h;
					}
					const dx = Math.max(50, Math.abs(tx - sx) / 2);
					dpath = `M${sx} ${sy} C${sx + dx} ${sy} ${tx - dx} ${ty} ${tx} ${ty}`;
					mid = [(sx + 3 * (sx + dx) + 3 * (tx - dx) + tx) / 8, (sy + 3 * sy + 3 * ty + ty) / 8 + (i - (links.length - 1) / 2) * 18];
					handle = {
						x: mid[0],
						y: mid[1]
					};
				}
				if (l.dashed) paths.push({
					key,
					d: dpath,
					stroke: SUP,
					width: selected ? 2 : 1.5,
					dash: "6 4",
					marker: "url(#arw-sup)"
				});
				else paths.push({
					key,
					d: dpath,
					stroke: isOut ? FG : isIn ? IN : MUTED,
					width: selected ? 1.75 : 1.25,
					dash: "none",
					marker: isOut ? "url(#arw-sel)" : isIn ? "url(#arw-in)" : "url(#arw)"
				});
				if (mid && l.label) labels.push({
					key,
					maxW: 150,
					x: mid[0],
					y: mid[1] - 13,
					text: l.label,
					full: l.label,
					color: isOut ? FG : isIn ? IN : MUTED,
					border: isIn ? IN : "var(--border)"
				});
				if (sel === n.id) handles.push({
					key,
					x: handle.x,
					y: handle.y,
					color: l.dashed ? SUP : FG,
					dashed: !!l.dashed
				});
			});
		});
		const S = g.start ? layout[g.start] : void 0;
		if (isBox(S)) {
			const y = S.y + S.h / 2;
			const x0 = S.x - 96;
			paths.push({
				key: "start",
				d: `M${x0} ${y} L${S.x} ${y}`,
				stroke: FG,
				width: 1.75,
				dash: "none",
				marker: "url(#arw-sel)"
			});
			labels.push({
				key: "start",
				maxW: 60,
				x: x0 + 8,
				y,
				text: "start",
				full: "Start node",
				color: FG,
				border: "var(--border)"
			});
		}
		const terminals = TERMINAL.map((t) => {
			const L = layout[t];
			const ok = t === "success";
			const selected = sel === t;
			const ring = selected ? "0 0 0 2px var(--foreground)" : ok ? "0 0 0 1px var(--border)" : "0 0 0 1px color-mix(in oklch, var(--destructive) 45%, transparent)";
			return {
				id: t,
				label: t,
				x: L.x,
				y: L.y,
				w: L.w,
				h: L.h,
				shadow: ring + ", 0 1px 2px rgba(0,0,0,.05)",
				color: ok ? FG : "var(--destructive)",
				dot: ok ? FG : "var(--destructive)",
				handleColor: selected ? MUTED : "transparent"
			};
		});
		const nodes = g.nodes.map((n) => {
			const L = isBox(layout[n.id]) ? layout[n.id] : {
				x: 0,
				y: 0,
				w: 240,
				h: 120
			};
			const selected = sel === n.id;
			const hasErrors = !!errs[n.id]?.length;
			const ring = hasErrors ? "0 0 0 2px var(--destructive)" : selected ? "0 0 0 2px var(--foreground)" : "0 0 0 1px color-mix(in oklch, var(--foreground) 10%, transparent)";
			const meta = [];
			const push = (k, v, inherited = false) => {
				if (v !== void 0 && v !== "" && v !== null && (showInherited || !inherited)) meta.push({
					k,
					v: String(v).split("\n")[0],
					color: inherited ? MUTED : FG
				});
			};
			const resolved = models.filter((model) => model.node_id === n.id);
			const pushModel = (label, role) => {
				const model = resolved.find((candidate) => !role || candidate.role === role);
				if (model) push(label, `${model.native_model} · ${model.effective_effort}`, !model.source.startsWith("node "));
			};
			if (n.type === "command") {
				push("cmd", n.command);
				push("timeout", n.timeout ?? d.timeout, n.timeout === void 0);
			} else if (n.type === "loop") {
				push("checklist", n.checklist);
				pushModel("judge", "item_judge");
				pushModel("evaluator", "goal_evaluator");
				push("visits", n.max_visits);
			} else if (n.type === "supervisor") {
				push("every", n.interval);
				push("watching", (n.supervises ?? []).length);
				pushModel("model", "supervisor");
			} else {
				pushModel("model");
				push("timeout", n.timeout ?? d.timeout, n.timeout === void 0);
				push("retries", n.max_retries ?? d.max_retries, n.max_retries === void 0);
				if (n.type === "fan_out") {
					push("branches", (n.branches ?? []).map(normalizeBranch).length);
					push("ws", n.workspace);
				}
			}
			const isSup = n.type === "supervisor";
			return {
				id: n.id,
				type: n.type,
				label: n.label || n.id,
				isStart: g.start === n.id,
				hasErrors,
				errorText: (errs[n.id] ?? []).join("\n"),
				x: L.x,
				y: L.y,
				w: L.w,
				h: L.h,
				shadow: ring + ", 0 1px 2px rgba(0,0,0,.06)",
				radius: isSup ? "4px" : "var(--radius-xl)",
				bg: isSup ? "color-mix(in oklch, var(--chart-2) 7%, var(--card))" : "var(--card)",
				border: isSup ? "1.5px dashed var(--chart-2)" : "1px solid transparent",
				badgeColor: isSup ? IN : MUTED,
				badgeBorder: isSup ? "color-mix(in oklch, var(--chart-2) 50%, transparent)" : "var(--border)",
				meta,
				handleColor: selected ? MUTED : "var(--border)"
			};
		});
		const B = boundsOf(layout);
		const miniNodes = [...g.nodes.map((n) => n.id), ...TERMINAL].map((id) => {
			const L = layout[id];
			if (!isBox(L)) return null;
			return {
				id,
				x: L.x,
				y: L.y,
				w: L.w,
				h: L.h,
				rx: isTerminal(id) ? 28 : 12,
				fill: sel === id ? FG : errs[id]?.length ? "var(--destructive)" : isTerminal(id) ? "var(--border)" : "var(--ring)"
			};
		}).filter((x) => x !== null);
		const { pan, zoom, vw, vh } = view;
		return {
			paths,
			labels,
			handles,
			terminals,
			nodes,
			mini: {
				box: `${B.x} ${B.y} ${B.w} ${B.h}`,
				nodes: miniNodes,
				view: {
					x: -pan.x / zoom,
					y: -pan.y / zoom,
					w: vw / zoom,
					h: vh / zoom,
					sw: Math.max(B.w, B.h) / 400
				},
				bounds: B,
				scale: Math.max(B.w / 180, B.h / 112)
			}
		};
	}
	function Canvas($$renderer, $$props) {
		$$renderer.component(($$renderer) => {
			let { editor } = $$props;
			const scene = derived(() => buildScene(editor.graph, editor.layout, editor.sel, editor.errors, {
				pan: editor.pan,
				zoom: editor.zoom,
				vw: editor.vw,
				vh: editor.vh
			}, editor.showInherited, editor.models));
			const xform = derived(() => `translate(${editor.pan.x}px, ${editor.pan.y}px) scale(${editor.zoom})`);
			const gridSize = derived(() => `${24 * editor.zoom}px ${24 * editor.zoom}px`);
			const gridRepeat = derived(() => editor.zoom < .4 ? "no-repeat" : "repeat");
			const gridPos = derived(() => `${editor.pan.x}px ${editor.pan.y}px`);
			const bgCursor = derived(() => editor.drag?.type === "pan" ? "grabbing" : "grab");
			$$renderer.push(`<div class="viewport svelte-o4ydsk" role="presentation"${attr_style("", {
				"background-size": gridSize(),
				"background-repeat": gridRepeat(),
				"background-position": gridPos(),
				cursor: bgCursor()
			})}><div class="world svelte-o4ydsk"${attr_style("", { transform: xform() })}><svg width="1" height="1" class="edges svelte-o4ydsk"><defs><marker id="arw" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="7" markerHeight="7" orient="auto-start-reverse"><path d="M0 0L10 5L0 10z" style="fill:var(--muted-foreground)"></path></marker><marker id="arw-sup" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="7" markerHeight="7" orient="auto-start-reverse"><path d="M0 0L10 5L0 10z" style="fill:var(--chart-2)"></path></marker><marker id="arw-in" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="7" markerHeight="7" orient="auto-start-reverse"><path d="M0 0L10 5L0 10z" style="fill:var(--chart-3)"></path></marker><marker id="arw-sel" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="7" markerHeight="7" orient="auto-start-reverse"><path d="M0 0L10 5L0 10z" style="fill:var(--foreground)"></path></marker></defs><!--[-->`);
			const each_array = ensure_array_like(scene().paths);
			for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
				let p = each_array[$$index];
				$$renderer.push(`<path${attr("d", p.d)} fill="none"${attr("stroke", p.stroke)}${attr("stroke-width", p.width)}${attr("stroke-dasharray", p.dash)}${attr("marker-end", p.marker)}></path>`);
			}
			$$renderer.push(`<!--]--></svg> <!--[-->`);
			const each_array_1 = ensure_array_like(scene().terminals);
			for (let $$index_1 = 0, $$length = each_array_1.length; $$index_1 < $$length; $$index_1++) {
				let t = each_array_1[$$index_1];
				$$renderer.push(`<div class="terminal svelte-o4ydsk" role="presentation"${attr_style("", {
					left: `${stringify$1(t.x)}px`,
					top: `${stringify$1(t.y)}px`,
					width: `${stringify$1(t.w)}px`,
					height: `${stringify$1(t.h)}px`,
					"box-shadow": t.shadow,
					color: t.color
				})}><span class="dot svelte-o4ydsk"${attr_style("", { background: t.dot })}></span>${escape_html(t.label)} <div class="resize terminal-resize svelte-o4ydsk" role="presentation"${attr_style("", { "border-color": t.handleColor })}></div></div>`);
			}
			$$renderer.push(`<!--]--> <!--[-->`);
			const each_array_2 = ensure_array_like(scene().nodes);
			for (let $$index_3 = 0, $$length = each_array_2.length; $$index_3 < $$length; $$index_3++) {
				let n = each_array_2[$$index_3];
				$$renderer.push(`<div class="node svelte-o4ydsk" role="presentation"${attr_style("", {
					left: `${stringify$1(n.x)}px`,
					top: `${stringify$1(n.y)}px`,
					width: `${stringify$1(n.w)}px`,
					height: `${stringify$1(n.h)}px`,
					background: n.bg,
					"border-radius": n.radius,
					border: n.border,
					"box-shadow": n.shadow
				})}><div class="node-head svelte-o4ydsk"><span class="badge svelte-o4ydsk"${attr_style("", {
					"border-color": n.badgeBorder,
					color: n.badgeColor
				})}>${escape_html(n.type)}</span> <div class="spacer svelte-o4ydsk"></div> `);
				if (n.hasErrors) {
					$$renderer.push("<!--[0-->");
					$$renderer.push(`<span class="warn tip svelte-o4ydsk"${attr("data-tip", n.errorText)}>`);
					Triangle_alert($$renderer, { size: 15 });
					$$renderer.push(`<!----></span>`);
				} else $$renderer.push("<!--[-1-->");
				$$renderer.push(`<!--]--></div> <div class="node-body svelte-o4ydsk"><div class="label svelte-o4ydsk">${escape_html(n.label)}</div> <div class="id svelte-o4ydsk">${escape_html(n.id)}</div></div> <div class="meta svelte-o4ydsk"><!--[-->`);
				const each_array_3 = ensure_array_like(n.meta);
				for (let $$index_2 = 0, $$length = each_array_3.length; $$index_2 < $$length; $$index_2++) {
					let m = each_array_3[$$index_2];
					$$renderer.push(`<span class="chip svelte-o4ydsk"><span class="k svelte-o4ydsk">${escape_html(m.k)}</span><span class="v svelte-o4ydsk"${attr_style("", { color: m.color })}>${escape_html(m.v)}</span></span>`);
				}
				$$renderer.push(`<!--]--></div> <div class="resize node-resize svelte-o4ydsk" role="presentation"${attr_style("", { "border-color": n.handleColor })}></div></div>`);
			}
			$$renderer.push(`<!--]--> <!--[-->`);
			const each_array_4 = ensure_array_like(scene().handles);
			for (let $$index_4 = 0, $$length = each_array_4.length; $$index_4 < $$length; $$index_4++) {
				let h = each_array_4[$$index_4];
				$$renderer.push(`<div class="handle svelte-o4ydsk" role="presentation" title="Drag to reshape edge"${attr_style("", {
					left: `${stringify$1(h.x)}px`,
					top: `${stringify$1(h.y)}px`,
					"border-color": h.color
				})}></div>`);
			}
			$$renderer.push(`<!--]--> <!--[-->`);
			const each_array_5 = ensure_array_like(scene().labels);
			for (let $$index_5 = 0, $$length = each_array_5.length; $$index_5 < $$length; $$index_5++) {
				let l = each_array_5[$$index_5];
				$$renderer.push(`<div class="edge-label svelte-o4ydsk"${attr("title", l.full)}${attr_style("", {
					left: `${stringify$1(l.x)}px`,
					top: `${stringify$1(l.y)}px`,
					"max-width": `${stringify$1(l.maxW)}px`,
					"border-color": l.border,
					color: l.color
				})}>${escape_html(l.text)}</div>`);
			}
			$$renderer.push(`<!--]--></div> <div class="banners svelte-o4ydsk">`);
			if (editor.syntaxError) {
				$$renderer.push("<!--[0-->");
				const e = editor.syntaxError;
				$$renderer.push(`<div class="banner cn-alert cn-alert-variant-destructive svelte-o4ydsk" role="alert">`);
				Triangle_alert($$renderer, { size: 15 });
				$$renderer.push(`<!----> <span>The file has a YAML syntax error${escape_html(e.line ? ` at line ${e.line}, column ${e.col ?? 1}` : "")}: ${escape_html(e.message)}. Editing is
					disabled until it is fixed on disk.</span></div>`);
			} else $$renderer.push("<!--[-1-->");
			$$renderer.push(`<!--]--> `);
			if (editor.banner) {
				$$renderer.push("<!--[0-->");
				$$renderer.push(`<div class="banner cn-alert cn-alert-variant-default svelte-o4ydsk" role="status"><span>${escape_html(editor.banner)}</span> <button class="cn-button cn-button-variant-ghost cn-button-size-icon-xs" title="Dismiss">`);
				X($$renderer, { size: 14 });
				$$renderer.push(`<!----></button></div>`);
			} else $$renderer.push("<!--[-1-->");
			$$renderer.push(`<!--]--></div> <div class="minimap svelte-o4ydsk" role="presentation"><svg width="180" height="112"${attr("viewBox", scene().mini.box)} preserveAspectRatio="xMidYMid meet" class="svelte-o4ydsk"><!--[-->`);
			const each_array_6 = ensure_array_like(scene().mini.nodes);
			for (let $$index_6 = 0, $$length = each_array_6.length; $$index_6 < $$length; $$index_6++) {
				let r = each_array_6[$$index_6];
				$$renderer.push(`<rect${attr("x", r.x)}${attr("y", r.y)}${attr("width", r.w)}${attr("height", r.h)}${attr("rx", r.rx)}${attr_style("", { fill: r.fill })}></rect>`);
			}
			$$renderer.push(`<!--]--><rect class="mini-view svelte-o4ydsk"${attr("x", scene().mini.view.x)}${attr("y", scene().mini.view.y)}${attr("width", scene().mini.view.w)}${attr("height", scene().mini.view.h)}${attr_style("", { "stroke-width": scene().mini.view.sw })}></rect></svg></div> <div class="hint svelte-o4ydsk"><span class="svelte-o4ydsk">Drag canvas to pan · scroll to pan · ⌘/Ctrl+scroll to zoom</span> <label class="snap svelte-o4ydsk"><button type="button" class="cn-checkbox" role="checkbox"${attr("aria-checked", editor.snap)}${attr("data-checked", editor.snap ? "" : void 0)}>`);
			if (editor.snap) {
				$$renderer.push("<!--[0-->");
				$$renderer.push(`<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><path d="M20 6 9 17l-5-5"></path></svg>`);
			} else $$renderer.push("<!--[-1-->");
			$$renderer.push(`<!--]--></button> <span>Snap to grid</span></label></div></div>`);
		});
	}
	function NativeSelect($$renderer, $$props) {
		$$renderer.component(($$renderer) => {
			let { value = "", options, size = "default", title = "", disabled = false, onchange } = $$props;
			$$renderer.push(`<span class="cn-native-select-wrapper">`);
			$$renderer.select({
				class: "cn-native-select",
				title,
				value,
				disabled,
				onchange: (e) => onchange(e.currentTarget.value, e.currentTarget)
			}, ($$renderer) => {
				$$renderer.push(`<!--[-->`);
				const each_array = ensure_array_like(options);
				for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
					let o = each_array[$$index];
					$$renderer.option({
						value: o.v,
						disabled: o.disabled ?? false
					}, ($$renderer) => {
						$$renderer.push(`${escape_html(o.l)}`);
					});
				}
				$$renderer.push(`<!--]-->`);
			}, void 0, { "cn-native-select-size-sm": size === "sm" });
			$$renderer.push(` `);
			Chevron_down($$renderer, { class: "cn-native-select-icon" });
			$$renderer.push(`<!----></span>`);
		});
	}
	function Header($$renderer, $$props) {
		$$renderer.component(($$renderer) => {
			let { editor } = $$props;
			const graphName = derived(() => editor.graph.name || "Untitled pipeline");
			const nodeCount = derived(() => editor.graph.nodes.length);
			const jumpOptions = derived(() => [{
				v: "",
				l: "Jump to node…"
			}, ...editor.graph.nodes.map((n) => ({
				v: n.id,
				l: n.label ? `${n.id} · ${n.label}` : n.id
			}))]);
			const jumpValue = derived(() => editor.selNode ? editor.selNode.id : "");
			const addOptions = [{
				v: "",
				l: "Add node…",
				disabled: true
			}, ...NODE_TYPES.map((t) => ({
				v: t,
				l: t
			}))];
			const zoomPct = derived(() => Math.round(editor.zoom * 100) + "%");
			$$renderer.push(`<header class="svelte-1elxaub"><span class="wordmark svelte-1elxaub">gimble</span> <span class="divider svelte-1elxaub"></span> <span class="name svelte-1elxaub"${attr("title", editor.path)}>${escape_html(graphName())}</span> <span class="count svelte-1elxaub">${escape_html(nodeCount())} nodes</span> <span${attr_class("cn-badge", void 0, {
				"cn-badge-variant-destructive": editor.errorCount > 0,
				"cn-badge-variant-outline": editor.errorCount === 0
			})}${attr("title", editor.errorCount ? "Schema and lint issues" : "No issues")}>${escape_html(editor.errorCount)} ${escape_html(editor.errorCount === 1 ? "issue" : "issues")}</span> <div class="spacer svelte-1elxaub"></div> <div class="jump svelte-1elxaub">`);
			NativeSelect($$renderer, {
				size: "sm",
				value: jumpValue(),
				options: jumpOptions(),
				onchange: (v) => v && editor.jumpTo(v)
			});
			$$renderer.push(`<!----></div> <div class="add svelte-1elxaub">`);
			NativeSelect($$renderer, {
				size: "sm",
				value: "",
				disabled: editor.readOnly,
				options: addOptions,
				onchange: (v, el) => {
					if (v) editor.addNode(v);
					el.value = "";
				}
			});
			$$renderer.push(`<!----></div> <div class="zoom svelte-1elxaub"><button class="cn-button cn-button-variant-ghost cn-button-size-icon-xs" title="Zoom out">`);
			Minus($$renderer, { size: 14 });
			$$renderer.push(`<!----></button> <span class="pct svelte-1elxaub">${escape_html(zoomPct())}</span> <button class="cn-button cn-button-variant-ghost cn-button-size-icon-xs" title="Zoom in">`);
			Plus($$renderer, { size: 14 });
			$$renderer.push(`<!----></button></div> <button class="cn-button cn-button-variant-outline cn-button-size-sm">Fit</button> <button class="cn-button cn-button-variant-outline cn-button-size-sm">Graph</button> <button class="cn-button cn-button-variant-ghost cn-button-size-icon-sm"${attr("title", editor.dark ? "Light mode" : "Dark mode")}>`);
			if (editor.dark) {
				$$renderer.push("<!--[0-->");
				Sun($$renderer, { size: 16 });
			} else {
				$$renderer.push("<!--[-1-->");
				Moon($$renderer, { size: 16 });
			}
			$$renderer.push(`<!--]--></button></header>`);
		});
	}
	function Field($$renderer, $$props) {
		$$renderer.component(($$renderer) => {
			let { field, showRequired = true, onvalue } = $$props;
			$$renderer.push(`<div class="field svelte-175sa6o"><div class="head svelte-175sa6o"><label class="cn-label"${attr("for", `field-${field.key}`)}>${escape_html(field.label)}</label> `);
			if (showRequired && field.required) {
				$$renderer.push("<!--[0-->");
				$$renderer.push(`<span class="hint svelte-175sa6o">required</span>`);
			} else $$renderer.push("<!--[-1-->");
			$$renderer.push(`<!--]--> `);
			if (field.inherited) {
				$$renderer.push("<!--[0-->");
				$$renderer.push(`<span class="hint inherited svelte-175sa6o">default: ${escape_html(field.inherited)}</span>`);
			} else $$renderer.push("<!--[-1-->");
			$$renderer.push(`<!--]--></div> `);
			if (field.kind === "text") {
				$$renderer.push("<!--[0-->");
				$$renderer.push(`<input${attr("id", `field-${field.key}`)}${attr_class("cn-input", void 0, { "mono": field.mono })}${attr("value", field.value)}${attr("placeholder", field.placeholder)}/>`);
			} else if (field.kind === "textarea") {
				$$renderer.push("<!--[1-->");
				$$renderer.push(`<textarea${attr("id", `field-${field.key}`)}${attr_class("cn-textarea svelte-175sa6o", void 0, { "mono": field.mono })}${attr("rows", field.rows)}${attr("placeholder", field.placeholder)}>`);
				const $$body = escape_html(field.value);
				if ($$body) $$renderer.push(`${$$body}`);
				$$renderer.push(`</textarea>`);
			} else {
				$$renderer.push("<!--[-1-->");
				NativeSelect($$renderer, {
					value: field.value,
					options: field.options,
					onchange: (v) => onvalue(v)
				});
			}
			$$renderer.push(`<!--]--> `);
			if (field.error) {
				$$renderer.push("<!--[0-->");
				$$renderer.push(`<div class="field-error">${escape_html(field.error)}</div>`);
			} else $$renderer.push("<!--[-1-->");
			$$renderer.push(`<!--]--></div>`);
		});
	}
	function Inspector($$renderer, $$props) {
		$$renderer.component(($$renderer) => {
			let { editor } = $$props;
			const graph = derived(() => editor.graph);
			const defaults = derived(() => graph().defaults ?? {});
			const ids = derived(() => graph().nodes.map((n) => n.id));
			const sel = derived(() => editor.selNode);
			const selErrors = derived(() => sel() ? editor.errors[sel().id] ?? [] : []);
			const selModels = derived(() => {
				if (!sel()) return [];
				const ids = /* @__PURE__ */ new Set([sel().id]);
				if (sel().type === "fan_out") for (const branch of sel().branches ?? []) ids.add(normalizeBranch(branch).id);
				return editor.models.filter((model) => ids.has(model.node_id));
			});
			const selIdError = derived(() => selErrors().find((e) => /^id |^duplicate/.test(e)) ?? "");
			const known = (x) => !x || ids().includes(x) || x === "success" || x === "failure";
			function parseValue(kind, v) {
				if (v === "") return void 0;
				if (kind === "int") {
					const n = parseInt(v, 10);
					return Number.isNaN(n) ? void 0 : n;
				}
				return v;
			}
			function field(key, raw, required, inheritFrom) {
				const m = FIELD_META[key] ?? { label: key };
				const kind = m.kind ?? "text";
				const value = raw === void 0 || raw === null ? "" : String(raw);
				const inheritedKey = key.startsWith("goal_evaluator.") ? key.slice(15) : key;
				const inherited = m.inherit && inheritFrom && getPath(inheritFrom, inheritedKey) !== void 0 && raw === void 0 ? String(getPath(inheritFrom, inheritedKey)) : "";
				let error = "";
				if (kind === "duration" && value && !DUR_RE.test(value)) error = "Integer followed by ms, s, m, h, or d";
				const options = kind === "enum" ? [{
					v: "",
					l: inherited ? `inherit (${inherited})` : "—"
				}, ...(m.options ?? []).map((o) => ({
					v: o,
					l: o
				}))] : [];
				return {
					key,
					label: m.label,
					value,
					kind: kind === "textarea" ? "textarea" : kind === "enum" ? "enum" : "text",
					options,
					rows: m.rows ?? 3,
					placeholder: inherited || m.placeholder || "",
					required,
					inherited,
					error,
					mono: !!(m.mono || kind === "duration" || kind === "int")
				};
			}
			const sections = derived(() => {
				if (!sel()) return [];
				const req = REQUIRED[sel().type] ?? [];
				return (FIELD_SECTIONS[sel().type] ?? []).map(([title, keys]) => ({
					title,
					fields: keys.map((k) => {
						let inheritFrom = defaults();
						if (k.startsWith("model.") && sel().model !== void 0) inheritFrom = null;
						if (k.startsWith("goal_evaluator.model.") && sel().goal_evaluator?.model !== void 0) inheritFrom = null;
						return {
							vm: field(k, getPath(sel(), k), req.includes(k), inheritFrom),
							kind: FIELD_META[k]?.kind ?? "text"
						};
					})
				}));
			});
			const targetOptions = derived(() => [
				{
					v: "",
					l: "— choose target —"
				},
				...graph().nodes.filter((n) => n.type !== "supervisor").map((n) => ({
					v: n.id,
					l: `${n.id}${n.label ? " · " + n.label : ""}`
				})),
				{
					v: "success",
					l: "success"
				},
				{
					v: "failure",
					l: "failure"
				}
			]);
			const transitions = derived(() => {
				if (!sel() || sel().type !== "command" && sel().type !== "loop") return [];
				return (sel().type === "command" ? [[
					"edges.success",
					"success",
					true
				], [
					"edges.error",
					"error",
					false
				]] : [[
					"edges.loop",
					"loop",
					true
				], [
					"edges.exit",
					"exit",
					true
				]]).map(([key, label, required]) => {
					const raw = getPath(sel(), key);
					const to = raw === void 0 || raw === null ? "" : String(raw);
					return {
						key,
						label,
						to,
						required,
						error: !known(to) ? "Unknown target" : required && !to ? "Required" : ""
					};
				});
			});
			const hasTransitions = derived(() => !!sel() && (sel().type === "command" || sel().type === "loop"));
			const hasEdgeList = derived(() => !!sel() && !hasTransitions() && sel().type !== "supervisor");
			const linksHint = derived(() => !sel() ? "" : sel().type === "command" ? "Exit code decides the edge; no conditions needed." : sel().type === "loop" ? "Loop is the entry node of one lap; exit is where the evaluator sends the walk." : sel().type === "fan_out" ? "Branch edges leave the fan-out once every branch has finished. Leave the condition blank for an unconditional edge." : "The agent picks one edge by its condition. Leave the condition blank for an unconditional edge.");
			const edges = derived(() => sel() && hasEdgeList() ? listEdges(sel()) : []);
			const superviseOptions = derived(() => sel() && sel().type === "supervisor" ? graph().nodes.filter((n) => n.id !== sel().id).map((n) => ({
				id: n.id,
				label: n.label ?? "",
				checked: (sel().supervises ?? []).includes(n.id)
			})) : []);
			const branches = derived(() => sel() && sel().type === "fan_out" ? (sel().branches ?? []).map(normalizeBranch) : []);
			const startOptions = derived(() => [{
				v: "",
				l: "—"
			}, ...graph().nodes.filter((n) => n.type !== "supervisor").map((n) => ({
				v: n.id,
				l: n.id
			}))]);
			const startError = derived(() => graph().start && !ids().includes(graph().start) ? "Start node does not exist" : !graph().start ? "Start is required" : "");
			const defaultFields = derived(() => DEFAULT_KEYS.map((k) => {
				const f = field(k, getPath(defaults(), k), false, null);
				if (f.kind === "enum") f.options[0] = {
					v: "",
					l: "—"
				};
				return {
					vm: f,
					kind: FIELD_META[k]?.kind ?? "text"
				};
			}));
			const nameField = derived(() => ({
				key: "name",
				label: "Name",
				value: graph().name ?? "",
				kind: "text",
				options: [],
				rows: 1,
				placeholder: "",
				required: false,
				inherited: "",
				error: "",
				mono: false
			}));
			const goalField = derived(() => ({
				key: "goal",
				label: "Goal",
				value: graph().goal ?? "",
				kind: "textarea",
				options: [],
				rows: 4,
				placeholder: "Objective exposed to prompt expansion",
				required: false,
				inherited: "",
				error: "",
				mono: false
			}));
			$$renderer.push(`<aside class="svelte-gu3dqx">`);
			if (!sel()) {
				$$renderer.push("<!--[0-->");
				$$renderer.push(`<fieldset class="panel svelte-gu3dqx"${attr("disabled", editor.readOnly, true)}><div><div class="title svelte-gu3dqx">Graph settings</div> <div class="subtitle svelte-gu3dqx">Pipeline metadata and file-level defaults.</div></div> `);
				if (editor.graphIssues.length) {
					$$renderer.push("<!--[0-->");
					$$renderer.push(`<div class="cn-alert cn-alert-variant-destructive">`);
					Triangle_alert($$renderer, {});
					$$renderer.push(`<!----> <div class="cn-alert-title">Pipeline issues</div> <div class="cn-alert-description svelte-gu3dqx"><!--[-->`);
					const each_array = ensure_array_like(editor.graphIssues);
					for (let i = 0, $$length = each_array.length; i < $$length; i++) {
						let e = each_array[i];
						$$renderer.push(`<div class="svelte-gu3dqx">${escape_html(e)}</div>`);
					}
					$$renderer.push(`<!--]--></div></div>`);
				} else $$renderer.push("<!--[-1-->");
				$$renderer.push(`<!--]--> <div class="section svelte-gu3dqx"><div class="section-title">Pipeline</div> `);
				Field($$renderer, {
					field: nameField(),
					onvalue: (v) => editor.setGraphField("name", v)
				});
				$$renderer.push(`<!----> `);
				Field($$renderer, {
					field: goalField(),
					onvalue: (v) => editor.setGraphField("goal", v)
				});
				$$renderer.push(`<!----> <div class="field svelte-gu3dqx"><label class="cn-label" for="field-start">Start node</label> `);
				NativeSelect($$renderer, {
					value: graph().start ?? "",
					options: startOptions(),
					onchange: (v) => editor.setGraphField("start", v)
				});
				$$renderer.push(`<!----> `);
				if (startError()) {
					$$renderer.push("<!--[0-->");
					$$renderer.push(`<div class="field-error">${escape_html(startError())}</div>`);
				} else $$renderer.push("<!--[-1-->");
				$$renderer.push(`<!--]--></div></div> <div class="section svelte-gu3dqx"><div class="section-title">Defaults</div> <!--[-->`);
				const each_array_1 = ensure_array_like(defaultFields());
				for (let $$index_1 = 0, $$length = each_array_1.length; $$index_1 < $$length; $$index_1++) {
					let f = each_array_1[$$index_1];
					Field($$renderer, {
						field: f.vm,
						showRequired: false,
						onvalue: (v) => editor.setDefault(f.vm.key, parseValue(f.kind, v))
					});
				}
				$$renderer.push(`<!--]--></div></fieldset>`);
			} else {
				$$renderer.push("<!--[-1-->");
				const node = sel();
				$$renderer.push(`<fieldset class="panel svelte-gu3dqx"${attr("disabled", editor.readOnly, true)}><div class="head svelte-gu3dqx"><div class="head-main svelte-gu3dqx"><div class="type-row svelte-gu3dqx"><span class="type-badge">${escape_html(node.type)}</span> <span class="type-desc svelte-gu3dqx">${escape_html(TYPE_DESC[node.type] ?? "")}</span></div> <div class="field id-field svelte-gu3dqx"><label class="cn-label" for="field-id">id</label> <input id="field-id" class="cn-input mono"${attr("value", node.id)}/> `);
				if (selIdError()) {
					$$renderer.push("<!--[0-->");
					$$renderer.push(`<div class="field-error">${escape_html(selIdError())}</div>`);
				} else $$renderer.push("<!--[-1-->");
				$$renderer.push(`<!--]--></div></div> <button class="cn-button cn-button-variant-ghost cn-button-size-icon-sm" title="Delete node">`);
				Trash($$renderer, { size: 16 });
				$$renderer.push(`<!----></button></div> `);
				if (selErrors().length) {
					$$renderer.push("<!--[0-->");
					$$renderer.push(`<div class="cn-alert cn-alert-variant-destructive">`);
					Triangle_alert($$renderer, {});
					$$renderer.push(`<!----> <div class="cn-alert-title">Schema issues</div> <div class="cn-alert-description svelte-gu3dqx"><!--[-->`);
					const each_array_2 = ensure_array_like(selErrors());
					for (let i = 0, $$length = each_array_2.length; i < $$length; i++) {
						let e = each_array_2[i];
						$$renderer.push(`<div class="svelte-gu3dqx">${escape_html(e)}</div>`);
					}
					$$renderer.push(`<!--]--></div></div>`);
				} else $$renderer.push("<!--[-1-->");
				$$renderer.push(`<!--]--> `);
				if (selModels().length) {
					$$renderer.push("<!--[0-->");
					$$renderer.push(`<div class="section effective-models svelte-gu3dqx"><div class="section-title">Effective selections</div> <!--[-->`);
					const each_array_3 = ensure_array_like(selModels());
					for (let $$index_3 = 0, $$length = each_array_3.length; $$index_3 < $$length; $$index_3++) {
						let model = each_array_3[$$index_3];
						$$renderer.push(`<div class="model-summary svelte-gu3dqx"><div><strong class="svelte-gu3dqx">${escape_html(model.node_id)}/${escape_html(model.role)}</strong> · ${escape_html(model.harness)}/${escape_html(model.provider)}</div> <div class="mono-sm svelte-gu3dqx">${escape_html(model.native_model)} · ${escape_html(model.effective_effort)}</div> <div>${escape_html(model.source)}${escape_html(model.authored_version ? ` · version ${model.authored_version}` : "")}</div></div>`);
					}
					$$renderer.push(`<!--]--></div>`);
				} else $$renderer.push("<!--[-1-->");
				$$renderer.push(`<!--]--> <div class="start-row svelte-gu3dqx"><button type="button" class="cn-switch" role="switch" aria-label="Start node"${attr("aria-checked", graph().start === node.id)}${attr("data-checked", graph().start === node.id ? "" : void 0)}><span class="cn-switch-thumb svelte-gu3dqx"></span></button> <span class="svelte-gu3dqx">Start node</span></div> <!--[-->`);
				const each_array_4 = ensure_array_like(sections());
				for (let $$index_5 = 0, $$length = each_array_4.length; $$index_5 < $$length; $$index_5++) {
					let s = each_array_4[$$index_5];
					$$renderer.push(`<div class="section svelte-gu3dqx"><div class="section-title">${escape_html(s.title)}</div> <!--[-->`);
					const each_array_5 = ensure_array_like(s.fields);
					for (let $$index_4 = 0, $$length = each_array_5.length; $$index_4 < $$length; $$index_4++) {
						let f = each_array_5[$$index_4];
						Field($$renderer, {
							field: f.vm,
							onvalue: (v) => editor.setNodeField(node.id, f.vm.key, parseValue(f.kind, v))
						});
					}
					$$renderer.push(`<!--]--></div>`);
				}
				$$renderer.push(`<!--]--> `);
				if (hasTransitions()) {
					$$renderer.push("<!--[0-->");
					$$renderer.push(`<div class="section svelte-gu3dqx"><div class="section-head"><span class="section-title">Transitions</span></div> <div class="links-hint svelte-gu3dqx">${escape_html(linksHint())}</div> <!--[-->`);
					const each_array_6 = ensure_array_like(transitions());
					for (let $$index_6 = 0, $$length = each_array_6.length; $$index_6 < $$length; $$index_6++) {
						let t = each_array_6[$$index_6];
						$$renderer.push(`<div class="link-card svelte-gu3dqx"><div class="link-main svelte-gu3dqx"><div class="link-row svelte-gu3dqx"><span class="link-label svelte-gu3dqx">${escape_html(t.label)}</span> `);
						NativeSelect($$renderer, {
							size: "sm",
							value: t.to,
							options: targetOptions(),
							onchange: (v) => editor.setNodeField(node.id, t.key, v || (t.required ? "" : void 0))
						});
						$$renderer.push(`<!----></div> `);
						if (t.error) {
							$$renderer.push("<!--[0-->");
							$$renderer.push(`<div class="field-error">${escape_html(t.error)}</div>`);
						} else $$renderer.push("<!--[-1-->");
						$$renderer.push(`<!--]--></div></div>`);
					}
					$$renderer.push(`<!--]--></div>`);
				} else $$renderer.push("<!--[-1-->");
				$$renderer.push(`<!--]--> `);
				if (hasEdgeList()) {
					$$renderer.push("<!--[0-->");
					$$renderer.push(`<div class="section svelte-gu3dqx"><div class="section-head"><span class="section-title">Outgoing edges</span> <button class="cn-button cn-button-variant-outline cn-button-size-xs">`);
					Plus($$renderer, { size: 12 });
					$$renderer.push(`<!---->Add edge</button></div> <div class="links-hint svelte-gu3dqx">${escape_html(linksHint())}</div> <!--[-->`);
					const each_array_7 = ensure_array_like(edges());
					for (let i = 0, $$length = each_array_7.length; i < $$length; i++) {
						let e = each_array_7[i];
						$$renderer.push(`<div class="link-card svelte-gu3dqx"><div class="link-main svelte-gu3dqx"><div class="link-row svelte-gu3dqx"><span class="link-label svelte-gu3dqx">edge ${escape_html(i + 1)}</span> `);
						NativeSelect($$renderer, {
							size: "sm",
							value: e.to ?? "",
							options: targetOptions(),
							onchange: (v) => editor.setEdge(node.id, i, { to: v })
						});
						$$renderer.push(`<!----></div> <input class="cn-input cn-input-size-sm"${attr("value", e.condition ?? "")} placeholder="Condition — why the agent should take this edge"/> `);
						if (!known(e.to ?? "")) {
							$$renderer.push("<!--[0-->");
							$$renderer.push(`<div class="field-error">Unknown target</div>`);
						} else $$renderer.push("<!--[-1-->");
						$$renderer.push(`<!--]--></div> <button class="cn-button cn-button-variant-ghost cn-button-size-icon-sm" title="Remove edge">`);
						X($$renderer, { size: 14 });
						$$renderer.push(`<!----></button></div>`);
					}
					$$renderer.push(`<!--]--></div>`);
				} else $$renderer.push("<!--[-1-->");
				$$renderer.push(`<!--]--> `);
				if (node.type === "supervisor") {
					$$renderer.push("<!--[0-->");
					$$renderer.push(`<div class="section supervises svelte-gu3dqx"><div class="section-title">Supervises</div> <!--[-->`);
					const each_array_8 = ensure_array_like(superviseOptions());
					for (let $$index_8 = 0, $$length = each_array_8.length; $$index_8 < $$length; $$index_8++) {
						let o = each_array_8[$$index_8];
						$$renderer.push(`<label class="check-row svelte-gu3dqx"><button type="button" class="cn-checkbox" role="checkbox"${attr("aria-label", `Supervise ${o.id}`)}${attr("aria-checked", o.checked)}${attr("data-checked", o.checked ? "" : void 0)}>`);
						if (o.checked) {
							$$renderer.push("<!--[0-->");
							$$renderer.push(`<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><path d="M20 6 9 17l-5-5"></path></svg>`);
						} else $$renderer.push("<!--[-1-->");
						$$renderer.push(`<!--]--></button> <span class="check-id svelte-gu3dqx">${escape_html(o.id)}</span><span class="check-label svelte-gu3dqx">${escape_html(o.label)}</span></label>`);
					}
					$$renderer.push(`<!--]--></div>`);
				} else $$renderer.push("<!--[-1-->");
				$$renderer.push(`<!--]--> `);
				if (node.type === "fan_out") {
					$$renderer.push("<!--[0-->");
					$$renderer.push(`<div class="section svelte-gu3dqx"><div class="section-head"><span class="section-title">Branches</span> <button class="cn-button cn-button-variant-outline cn-button-size-xs">`);
					Plus($$renderer, { size: 12 });
					$$renderer.push(`<!---->Add branch</button></div> <!--[-->`);
					const each_array_9 = ensure_array_like(branches());
					for (let i = 0, $$length = each_array_9.length; i < $$length; i++) {
						let b = each_array_9[i];
						$$renderer.push(`<div class="link-card svelte-gu3dqx"><div class="link-main svelte-gu3dqx"><input class="cn-input cn-input-size-sm mono-sm svelte-gu3dqx"${attr("value", b.id ?? "")} placeholder="branch id"/> <input class="cn-input cn-input-size-sm mono-sm svelte-gu3dqx"${attr("value", (b.artifacts ?? []).join(", "))} placeholder="artifacts, comma separated"/> <textarea class="cn-textarea branch-prompt svelte-gu3dqx" rows="3" placeholder="Prompt override (optional)">`);
						const $$body = escape_html(b.agent?.prompt ?? "");
						if ($$body) $$renderer.push(`${$$body}`);
						$$renderer.push(`</textarea> <div class="branch-model-grid svelte-gu3dqx"><input class="cn-input cn-input-size-sm mono-sm svelte-gu3dqx"${attr("value", b.agent?.model?.name ?? "")} placeholder="model name (inherit)"/> <input class="cn-input cn-input-size-sm mono-sm svelte-gu3dqx"${attr("value", b.agent?.model?.version ?? "")} placeholder="version (optional)"/> `);
						NativeSelect($$renderer, {
							size: "sm",
							value: b.agent?.model?.effort ?? "",
							options: [{
								v: "",
								l: "effort (model default)"
							}, ...EFFORT.map((v) => ({
								v,
								l: v
							}))],
							onchange: (v) => editor.setBranchModel(node.id, i, "effort", v || void 0)
						});
						$$renderer.push(`<!----></div></div> <button class="cn-button cn-button-variant-ghost cn-button-size-icon-sm" title="Remove branch">`);
						X($$renderer, { size: 14 });
						$$renderer.push(`<!----></button></div>`);
					}
					$$renderer.push(`<!--]--></div>`);
				} else $$renderer.push("<!--[-1-->");
				$$renderer.push(`<!--]--></fieldset>`);
			}
			$$renderer.push(`<!--]--></aside>`);
		});
	}
	var ConflictError = class extends Error {
		current;
		constructor(current) {
			super("the file changed on disk");
			this.current = current;
		}
	};
	function fromWire(d) {
		let layout = null;
		if (d.has_layout) {
			layout = {};
			for (const p of d.layout) {
				const entry = {
					x: p.x,
					y: p.y
				};
				if (p.w > 0) entry.w = p.w;
				if (p.h > 0) entry.h = p.h;
				layout[p.id] = entry;
			}
		}
		return {
			path: d.path,
			yaml: d.yaml,
			layout,
			version: d.version,
			diagnostics: d.diagnostics ?? [],
			parse_error: d.parse_error ?? "",
			models: d.models ?? []
		};
	}
	function toWire(layout) {
		return Object.entries(layout).map(([id, e]) => ({
			id,
			x: e.x,
			y: e.y,
			w: e.w ?? 0,
			h: e.h ?? 0
		}));
	}
	async function getDoc() {
		const q = getDoc$1();
		await q.refresh();
		return fromWire(await q);
	}
	async function putDoc(body) {
		const result = await saveDoc({
			yaml: body.yaml,
			layout: body.layout ? toWire(body.layout) : [],
			write_layout: body.layout !== null,
			version: body.version
		});
		const current = fromWire(result.document);
		if (!result.saved) throw new ConflictError(current);
		return current;
	}
	function subscribe(onChange) {
		const iterator = watchDoc()[Symbol.asyncIterator]();
		let stopped = false;
		(async () => {
			try {
				for (;;) {
					const { done, value } = await iterator.next();
					if (done || stopped) return;
					onChange(value.version);
				}
			} catch (e) {
				console.error("editor: the change stream ended", e);
			}
		})();
		return () => {
			stopped = true;
			iterator.return?.();
		};
	}
	var ALIAS = Symbol.for("yaml.alias");
	var DOC = Symbol.for("yaml.document");
	var MAP = Symbol.for("yaml.map");
	var PAIR = Symbol.for("yaml.pair");
	var SCALAR$1 = Symbol.for("yaml.scalar");
	var SEQ = Symbol.for("yaml.seq");
	var NODE_TYPE = Symbol.for("yaml.node.type");
	var isAlias = (node) => !!node && typeof node === "object" && node[NODE_TYPE] === ALIAS;
	var isDocument = (node) => !!node && typeof node === "object" && node[NODE_TYPE] === DOC;
	var isMap = (node) => !!node && typeof node === "object" && node[NODE_TYPE] === MAP;
	var isPair = (node) => !!node && typeof node === "object" && node[NODE_TYPE] === PAIR;
	var isScalar = (node) => !!node && typeof node === "object" && node[NODE_TYPE] === SCALAR$1;
	var isSeq = (node) => !!node && typeof node === "object" && node[NODE_TYPE] === SEQ;
	function isCollection(node) {
		if (node && typeof node === "object") switch (node[NODE_TYPE]) {
			case MAP:
			case SEQ: return true;
		}
		return false;
	}
	function isNode(node) {
		if (node && typeof node === "object") switch (node[NODE_TYPE]) {
			case ALIAS:
			case MAP:
			case SCALAR$1:
			case SEQ: return true;
		}
		return false;
	}
	var hasAnchor = (node) => (isScalar(node) || isCollection(node)) && !!node.anchor;
	var BREAK$1 = Symbol("break visit");
	var SKIP$1 = Symbol("skip children");
	var REMOVE$1 = Symbol("remove node");
	/**
	* Apply a visitor to an AST node or document.
	*
	* Walks through the tree (depth-first) starting from `node`, calling a
	* `visitor` function with three arguments:
	*   - `key`: For sequence values and map `Pair`, the node's index in the
	*     collection. Within a `Pair`, `'key'` or `'value'`, correspondingly.
	*     `null` for the root node.
	*   - `node`: The current node.
	*   - `path`: The ancestry of the current node.
	*
	* The return value of the visitor may be used to control the traversal:
	*   - `undefined` (default): Do nothing and continue
	*   - `visit.SKIP`: Do not visit the children of this node, continue with next
	*     sibling
	*   - `visit.BREAK`: Terminate traversal completely
	*   - `visit.REMOVE`: Remove the current node, then continue with the next one
	*   - `Node`: Replace the current node, then continue by visiting it
	*   - `number`: While iterating the items of a sequence or map, set the index
	*     of the next step. This is useful especially if the index of the current
	*     node has changed.
	*
	* If `visitor` is a single function, it will be called with all values
	* encountered in the tree, including e.g. `null` values. Alternatively,
	* separate visitor functions may be defined for each `Map`, `Pair`, `Seq`,
	* `Alias` and `Scalar` node. To define the same visitor function for more than
	* one node type, use the `Collection` (map and seq), `Value` (map, seq & scalar)
	* and `Node` (alias, map, seq & scalar) targets. Of all these, only the most
	* specific defined one will be used for each node.
	*/
	function visit$1(node, visitor) {
		const visitor_ = initVisitor(visitor);
		if (isDocument(node)) {
			if (visit_(null, node.contents, visitor_, Object.freeze([node])) === REMOVE$1) node.contents = null;
		} else visit_(null, node, visitor_, Object.freeze([]));
	}
	/** Terminate visit traversal completely */
	visit$1.BREAK = BREAK$1;
	/** Do not visit the children of the current node */
	visit$1.SKIP = SKIP$1;
	/** Remove the current node */
	visit$1.REMOVE = REMOVE$1;
	function visit_(key, node, visitor, path) {
		const ctrl = callVisitor(key, node, visitor, path);
		if (isNode(ctrl) || isPair(ctrl)) {
			replaceNode(key, path, ctrl);
			return visit_(key, ctrl, visitor, path);
		}
		if (typeof ctrl !== "symbol") {
			if (isCollection(node)) {
				path = Object.freeze(path.concat(node));
				for (let i = 0; i < node.items.length; ++i) {
					const ci = visit_(i, node.items[i], visitor, path);
					if (typeof ci === "number") i = ci - 1;
					else if (ci === BREAK$1) return BREAK$1;
					else if (ci === REMOVE$1) {
						node.items.splice(i, 1);
						i -= 1;
					}
				}
			} else if (isPair(node)) {
				path = Object.freeze(path.concat(node));
				const ck = visit_("key", node.key, visitor, path);
				if (ck === BREAK$1) return BREAK$1;
				else if (ck === REMOVE$1) node.key = null;
				const cv = visit_("value", node.value, visitor, path);
				if (cv === BREAK$1) return BREAK$1;
				else if (cv === REMOVE$1) node.value = null;
			}
		}
		return ctrl;
	}
	/**
	* Apply an async visitor to an AST node or document.
	*
	* Walks through the tree (depth-first) starting from `node`, calling a
	* `visitor` function with three arguments:
	*   - `key`: For sequence values and map `Pair`, the node's index in the
	*     collection. Within a `Pair`, `'key'` or `'value'`, correspondingly.
	*     `null` for the root node.
	*   - `node`: The current node.
	*   - `path`: The ancestry of the current node.
	*
	* The return value of the visitor may be used to control the traversal:
	*   - `Promise`: Must resolve to one of the following values
	*   - `undefined` (default): Do nothing and continue
	*   - `visit.SKIP`: Do not visit the children of this node, continue with next
	*     sibling
	*   - `visit.BREAK`: Terminate traversal completely
	*   - `visit.REMOVE`: Remove the current node, then continue with the next one
	*   - `Node`: Replace the current node, then continue by visiting it
	*   - `number`: While iterating the items of a sequence or map, set the index
	*     of the next step. This is useful especially if the index of the current
	*     node has changed.
	*
	* If `visitor` is a single function, it will be called with all values
	* encountered in the tree, including e.g. `null` values. Alternatively,
	* separate visitor functions may be defined for each `Map`, `Pair`, `Seq`,
	* `Alias` and `Scalar` node. To define the same visitor function for more than
	* one node type, use the `Collection` (map and seq), `Value` (map, seq & scalar)
	* and `Node` (alias, map, seq & scalar) targets. Of all these, only the most
	* specific defined one will be used for each node.
	*/
	async function visitAsync(node, visitor) {
		const visitor_ = initVisitor(visitor);
		if (isDocument(node)) {
			if (await visitAsync_(null, node.contents, visitor_, Object.freeze([node])) === REMOVE$1) node.contents = null;
		} else await visitAsync_(null, node, visitor_, Object.freeze([]));
	}
	/** Terminate visit traversal completely */
	visitAsync.BREAK = BREAK$1;
	/** Do not visit the children of the current node */
	visitAsync.SKIP = SKIP$1;
	/** Remove the current node */
	visitAsync.REMOVE = REMOVE$1;
	async function visitAsync_(key, node, visitor, path) {
		const ctrl = await callVisitor(key, node, visitor, path);
		if (isNode(ctrl) || isPair(ctrl)) {
			replaceNode(key, path, ctrl);
			return visitAsync_(key, ctrl, visitor, path);
		}
		if (typeof ctrl !== "symbol") {
			if (isCollection(node)) {
				path = Object.freeze(path.concat(node));
				for (let i = 0; i < node.items.length; ++i) {
					const ci = await visitAsync_(i, node.items[i], visitor, path);
					if (typeof ci === "number") i = ci - 1;
					else if (ci === BREAK$1) return BREAK$1;
					else if (ci === REMOVE$1) {
						node.items.splice(i, 1);
						i -= 1;
					}
				}
			} else if (isPair(node)) {
				path = Object.freeze(path.concat(node));
				const ck = await visitAsync_("key", node.key, visitor, path);
				if (ck === BREAK$1) return BREAK$1;
				else if (ck === REMOVE$1) node.key = null;
				const cv = await visitAsync_("value", node.value, visitor, path);
				if (cv === BREAK$1) return BREAK$1;
				else if (cv === REMOVE$1) node.value = null;
			}
		}
		return ctrl;
	}
	function initVisitor(visitor) {
		if (typeof visitor === "object" && (visitor.Collection || visitor.Node || visitor.Value)) return Object.assign({
			Alias: visitor.Node,
			Map: visitor.Node,
			Scalar: visitor.Node,
			Seq: visitor.Node
		}, visitor.Value && {
			Map: visitor.Value,
			Scalar: visitor.Value,
			Seq: visitor.Value
		}, visitor.Collection && {
			Map: visitor.Collection,
			Seq: visitor.Collection
		}, visitor);
		return visitor;
	}
	function callVisitor(key, node, visitor, path) {
		if (typeof visitor === "function") return visitor(key, node, path);
		if (isMap(node)) return visitor.Map?.(key, node, path);
		if (isSeq(node)) return visitor.Seq?.(key, node, path);
		if (isPair(node)) return visitor.Pair?.(key, node, path);
		if (isScalar(node)) return visitor.Scalar?.(key, node, path);
		if (isAlias(node)) return visitor.Alias?.(key, node, path);
	}
	function replaceNode(key, path, node) {
		const parent = path[path.length - 1];
		if (isCollection(parent)) parent.items[key] = node;
		else if (isPair(parent)) {
			if (key === "key") parent.key = node;
			else parent.value = node;
		} else if (isDocument(parent)) parent.contents = node;
		else {
			const pt = isAlias(parent) ? "alias" : "scalar";
			throw new Error(`Cannot replace node with ${pt} parent`);
		}
	}
	var escapeChars = {
		"!": "%21",
		",": "%2C",
		"[": "%5B",
		"]": "%5D",
		"{": "%7B",
		"}": "%7D"
	};
	var escapeTagName = (tn) => tn.replace(/[!,[\]{}]/g, (ch) => escapeChars[ch]);
	var Directives = class Directives {
		constructor(yaml, tags) {
			/**
			* The directives-end/doc-start marker `---`. If `null`, a marker may still be
			* included in the document's stringified representation.
			*/
			this.docStart = null;
			/** The doc-end marker `...`.  */
			this.docEnd = false;
			this.yaml = Object.assign({}, Directives.defaultYaml, yaml);
			this.tags = Object.assign({}, Directives.defaultTags, tags);
		}
		clone() {
			const copy = new Directives(this.yaml, this.tags);
			copy.docStart = this.docStart;
			return copy;
		}
		/**
		* During parsing, get a Directives instance for the current document and
		* update the stream state according to the current version's spec.
		*/
		atDocument() {
			const res = new Directives(this.yaml, this.tags);
			switch (this.yaml.version) {
				case "1.1":
					this.atNextDocument = true;
					break;
				case "1.2":
					this.atNextDocument = false;
					this.yaml = {
						explicit: Directives.defaultYaml.explicit,
						version: "1.2"
					};
					this.tags = Object.assign({}, Directives.defaultTags);
			}
			return res;
		}
		/**
		* @param onError - May be called even if the action was successful
		* @returns `true` on success
		*/
		add(line, onError) {
			if (this.atNextDocument) {
				this.yaml = {
					explicit: Directives.defaultYaml.explicit,
					version: "1.1"
				};
				this.tags = Object.assign({}, Directives.defaultTags);
				this.atNextDocument = false;
			}
			const parts = line.trim().split(/[ \t]+/);
			const name = parts.shift();
			switch (name) {
				case "%TAG": {
					if (parts.length !== 2) {
						onError(0, "%TAG directive should contain exactly two parts");
						if (parts.length < 2) return false;
					}
					const [handle, prefix] = parts;
					this.tags[handle] = prefix;
					return true;
				}
				case "%YAML": {
					this.yaml.explicit = true;
					if (parts.length !== 1) {
						onError(0, "%YAML directive should contain exactly one part");
						return false;
					}
					const [version] = parts;
					if (version === "1.1" || version === "1.2") {
						this.yaml.version = version;
						return true;
					} else {
						const isValid = /^\d+\.\d+$/.test(version);
						onError(6, `Unsupported YAML version ${version}`, isValid);
						return false;
					}
				}
				default:
					onError(0, `Unknown directive ${name}`, true);
					return false;
			}
		}
		/**
		* Resolves a tag, matching handles to those defined in %TAG directives.
		*
		* @returns Resolved tag, which may also be the non-specific tag `'!'` or a
		*   `'!local'` tag, or `null` if unresolvable.
		*/
		tagName(source, onError) {
			if (source === "!") return "!";
			if (source[0] !== "!") {
				onError(`Not a valid tag: ${source}`);
				return null;
			}
			if (source[1] === "<") {
				const verbatim = source.slice(2, -1);
				if (verbatim === "!" || verbatim === "!!") {
					onError(`Verbatim tags aren't resolved, so ${source} is invalid.`);
					return null;
				}
				if (source[source.length - 1] !== ">") onError("Verbatim tags must end with a >");
				return verbatim;
			}
			const [, handle, suffix] = source.match(/^(.*!)([^!]*)$/s);
			if (!suffix) onError(`The ${source} tag has no suffix`);
			const prefix = this.tags[handle];
			if (prefix) try {
				return prefix + decodeURIComponent(suffix);
			} catch (error) {
				onError(String(error));
				return null;
			}
			if (handle === "!") return source;
			onError(`Could not resolve tag: ${source}`);
			return null;
		}
		/**
		* Given a fully resolved tag, returns its printable string form,
		* taking into account current tag prefixes and defaults.
		*/
		tagString(tag) {
			for (const [handle, prefix] of Object.entries(this.tags)) if (tag.startsWith(prefix)) return handle + escapeTagName(tag.substring(prefix.length));
			return tag[0] === "!" ? tag : `!<${tag}>`;
		}
		toString(doc) {
			const lines = this.yaml.explicit ? [`%YAML ${this.yaml.version || "1.2"}`] : [];
			const tagEntries = Object.entries(this.tags);
			let tagNames;
			if (doc && tagEntries.length > 0 && isNode(doc.contents)) {
				const tags = {};
				visit$1(doc.contents, (_key, node) => {
					if (isNode(node) && node.tag) tags[node.tag] = true;
				});
				tagNames = Object.keys(tags);
			} else tagNames = [];
			for (const [handle, prefix] of tagEntries) {
				if (handle === "!!" && prefix === "tag:yaml.org,2002:") continue;
				if (!doc || tagNames.some((tn) => tn.startsWith(prefix))) lines.push(`%TAG ${handle} ${prefix}`);
			}
			return lines.join("\n");
		}
	};
	Directives.defaultYaml = {
		explicit: false,
		version: "1.2"
	};
	Directives.defaultTags = { "!!": "tag:yaml.org,2002:" };
	/**
	* Verify that the input string is a valid anchor.
	*
	* Will throw on errors.
	*/
	function anchorIsValid(anchor) {
		if (/[\x00-\x19\s,[\]{}]/.test(anchor)) {
			const msg = `Anchor must not contain whitespace or control characters: ${JSON.stringify(anchor)}`;
			throw new Error(msg);
		}
		return true;
	}
	function anchorNames(root) {
		const anchors = /* @__PURE__ */ new Set();
		visit$1(root, { Value(_key, node) {
			if (node.anchor) anchors.add(node.anchor);
		} });
		return anchors;
	}
	/** Find a new anchor name with the given `prefix` and a one-indexed suffix. */
	function findNewAnchor(prefix, exclude) {
		for (let i = 1;; ++i) {
			const name = `${prefix}${i}`;
			if (!exclude.has(name)) return name;
		}
	}
	function createNodeAnchors(doc, prefix) {
		const aliasObjects = [];
		const sourceObjects = /* @__PURE__ */ new Map();
		let prevAnchors = null;
		return {
			onAnchor: (source) => {
				aliasObjects.push(source);
				prevAnchors ?? (prevAnchors = anchorNames(doc));
				const anchor = findNewAnchor(prefix, prevAnchors);
				prevAnchors.add(anchor);
				return anchor;
			},
			/**
			* With circular references, the source node is only resolved after all
			* of its child nodes are. This is why anchors are set only after all of
			* the nodes have been created.
			*/
			setAnchors: () => {
				for (const source of aliasObjects) {
					const ref = sourceObjects.get(source);
					if (typeof ref === "object" && ref.anchor && (isScalar(ref.node) || isCollection(ref.node))) ref.node.anchor = ref.anchor;
					else {
						const error = /* @__PURE__ */ new Error("Failed to resolve repeated object (this should not happen)");
						error.source = source;
						throw error;
					}
				}
			},
			sourceObjects
		};
	}
	/**
	* Applies the JSON.parse reviver algorithm as defined in the ECMA-262 spec,
	* in section 24.5.1.1 "Runtime Semantics: InternalizeJSONProperty" of the
	* 2021 edition: https://tc39.es/ecma262/#sec-json.parse
	*
	* Includes extensions for handling Map and Set objects.
	*/
	function applyReviver(reviver, obj, key, val) {
		if (val && typeof val === "object") {
			if (Array.isArray(val)) for (let i = 0, len = val.length; i < len; ++i) {
				const v0 = val[i];
				const v1 = applyReviver(reviver, val, String(i), v0);
				if (v1 === void 0) delete val[i];
				else if (v1 !== v0) val[i] = v1;
			}
			else if (val instanceof Map) for (const k of Array.from(val.keys())) {
				const v0 = val.get(k);
				const v1 = applyReviver(reviver, val, k, v0);
				if (v1 === void 0) val.delete(k);
				else if (v1 !== v0) val.set(k, v1);
			}
			else if (val instanceof Set) for (const v0 of Array.from(val)) {
				const v1 = applyReviver(reviver, val, v0, v0);
				if (v1 === void 0) val.delete(v0);
				else if (v1 !== v0) {
					val.delete(v0);
					val.add(v1);
				}
			}
			else for (const [k, v0] of Object.entries(val)) {
				const v1 = applyReviver(reviver, val, k, v0);
				if (v1 === void 0) delete val[k];
				else if (v1 !== v0) val[k] = v1;
			}
		}
		return reviver.call(obj, key, val);
	}
	/**
	* Recursively convert any node or its contents to native JavaScript
	*
	* @param value - The input value
	* @param arg - If `value` defines a `toJSON()` method, use this
	*   as its first argument
	* @param ctx - Conversion context, originally set in Document#toJS(). If
	*   `{ keep: true }` is not set, output should be suitable for JSON
	*   stringification.
	*/
	function toJS(value, arg, ctx) {
		if (Array.isArray(value)) return value.map((v, i) => toJS(v, String(i), ctx));
		if (value && typeof value.toJSON === "function") {
			if (!ctx || !hasAnchor(value)) return value.toJSON(arg, ctx);
			const data = {
				aliasCount: 0,
				count: 1,
				res: void 0
			};
			ctx.anchors.set(value, data);
			ctx.onCreate = (res) => {
				data.res = res;
				delete ctx.onCreate;
			};
			const res = value.toJSON(arg, ctx);
			if (ctx.onCreate) ctx.onCreate(res);
			return res;
		}
		if (typeof value === "bigint" && !ctx?.keep) return Number(value);
		return value;
	}
	var NodeBase = class {
		constructor(type) {
			Object.defineProperty(this, NODE_TYPE, { value: type });
		}
		/** Create a copy of this node.  */
		clone() {
			const copy = Object.create(Object.getPrototypeOf(this), Object.getOwnPropertyDescriptors(this));
			if (this.range) copy.range = this.range.slice();
			return copy;
		}
		/** A plain JavaScript representation of this node. */
		toJS(doc, { mapAsMap, maxAliasCount, onAnchor, reviver } = {}) {
			if (!isDocument(doc)) throw new TypeError("A document argument is required");
			const ctx = {
				anchors: /* @__PURE__ */ new Map(),
				doc,
				keep: true,
				mapAsMap: mapAsMap === true,
				mapKeyWarned: false,
				maxAliasCount: typeof maxAliasCount === "number" ? maxAliasCount : 100
			};
			const res = toJS(this, "", ctx);
			if (typeof onAnchor === "function") for (const { count, res } of ctx.anchors.values()) onAnchor(res, count);
			return typeof reviver === "function" ? applyReviver(reviver, { "": res }, "", res) : res;
		}
	};
	var Alias = class extends NodeBase {
		constructor(source) {
			super(ALIAS);
			this.source = source;
			Object.defineProperty(this, "tag", { set() {
				throw new Error("Alias nodes cannot have tags");
			} });
		}
		/**
		* Resolve the value of this alias within `doc`, finding the last
		* instance of the `source` anchor before this node.
		*/
		resolve(doc, ctx) {
			if (ctx?.maxAliasCount === 0) throw new ReferenceError("Alias resolution is disabled");
			let nodes;
			if (ctx?.aliasResolveCache) nodes = ctx.aliasResolveCache;
			else {
				nodes = [];
				visit$1(doc, { Node: (_key, node) => {
					if (isAlias(node) || hasAnchor(node)) nodes.push(node);
				} });
				if (ctx) ctx.aliasResolveCache = nodes;
			}
			let found = void 0;
			for (const node of nodes) {
				if (node === this) break;
				if (node.anchor === this.source) found = node;
			}
			return found;
		}
		toJSON(_arg, ctx) {
			if (!ctx) return { source: this.source };
			const { anchors, doc, maxAliasCount } = ctx;
			const source = this.resolve(doc, ctx);
			if (!source) {
				const msg = `Unresolved alias (the anchor must be set before the alias): ${this.source}`;
				throw new ReferenceError(msg);
			}
			let data = anchors.get(source);
			if (!data) {
				toJS(source, null, ctx);
				data = anchors.get(source);
			}
			/* istanbul ignore if */
			if (data?.res === void 0) throw new ReferenceError("This should not happen: Alias anchor was not resolved?");
			if (maxAliasCount >= 0) {
				data.count += 1;
				if (data.aliasCount === 0) data.aliasCount = getAliasCount(doc, source, anchors);
				if (data.count * data.aliasCount > maxAliasCount) throw new ReferenceError("Excessive alias count indicates a resource exhaustion attack");
			}
			return data.res;
		}
		toString(ctx, _onComment, _onChompKeep) {
			const src = `*${this.source}`;
			if (ctx) {
				anchorIsValid(this.source);
				if (ctx.options.verifyAliasOrder && !ctx.anchors.has(this.source)) {
					const msg = `Unresolved alias (the anchor must be set before the alias): ${this.source}`;
					throw new Error(msg);
				}
				if (ctx.implicitKey) return `${src} `;
			}
			return src;
		}
	};
	function getAliasCount(doc, node, anchors) {
		if (isAlias(node)) {
			const source = node.resolve(doc);
			const anchor = anchors && source && anchors.get(source);
			return anchor ? anchor.count * anchor.aliasCount : 0;
		} else if (isCollection(node)) {
			let count = 0;
			for (const item of node.items) {
				const c = getAliasCount(doc, item, anchors);
				if (c > count) count = c;
			}
			return count;
		} else if (isPair(node)) {
			const kc = getAliasCount(doc, node.key, anchors);
			const vc = getAliasCount(doc, node.value, anchors);
			return Math.max(kc, vc);
		}
		return 1;
	}
	var isScalarValue = (value) => !value || typeof value !== "function" && typeof value !== "object";
	var Scalar = class extends NodeBase {
		constructor(value) {
			super(SCALAR$1);
			this.value = value;
		}
		toJSON(arg, ctx) {
			return ctx?.keep ? this.value : toJS(this.value, arg, ctx);
		}
		toString() {
			return String(this.value);
		}
	};
	Scalar.BLOCK_FOLDED = "BLOCK_FOLDED";
	Scalar.BLOCK_LITERAL = "BLOCK_LITERAL";
	Scalar.PLAIN = "PLAIN";
	Scalar.QUOTE_DOUBLE = "QUOTE_DOUBLE";
	Scalar.QUOTE_SINGLE = "QUOTE_SINGLE";
	var defaultTagPrefix = "tag:yaml.org,2002:";
	function findTagObject(value, tagName, tags) {
		if (tagName) {
			const match = tags.filter((t) => t.tag === tagName);
			const tagObj = match.find((t) => !t.format) ?? match[0];
			if (!tagObj) throw new Error(`Tag ${tagName} not found`);
			return tagObj;
		}
		return tags.find((t) => t.identify?.(value) && !t.format);
	}
	function createNode(value, tagName, ctx) {
		if (isDocument(value)) value = value.contents;
		if (isNode(value)) return value;
		if (isPair(value)) {
			const map = ctx.schema[MAP].createNode?.(ctx.schema, null, ctx);
			map.items.push(value);
			return map;
		}
		if (value instanceof String || value instanceof Number || value instanceof Boolean || typeof BigInt !== "undefined" && value instanceof BigInt) value = value.valueOf();
		const { aliasDuplicateObjects, onAnchor, onTagObj, schema, sourceObjects } = ctx;
		let ref = void 0;
		if (aliasDuplicateObjects && value && typeof value === "object") {
			ref = sourceObjects.get(value);
			if (ref) {
				ref.anchor ?? (ref.anchor = onAnchor(value));
				return new Alias(ref.anchor);
			} else {
				ref = {
					anchor: null,
					node: null
				};
				sourceObjects.set(value, ref);
			}
		}
		if (tagName?.startsWith("!!")) tagName = defaultTagPrefix + tagName.slice(2);
		let tagObj = findTagObject(value, tagName, schema.tags);
		if (!tagObj) {
			if (value && typeof value.toJSON === "function") value = value.toJSON();
			if (!value || typeof value !== "object") {
				const node = new Scalar(value);
				if (ref) ref.node = node;
				return node;
			}
			tagObj = value instanceof Map ? schema[MAP] : Symbol.iterator in Object(value) ? schema[SEQ] : schema[MAP];
		}
		if (onTagObj) {
			onTagObj(tagObj);
			delete ctx.onTagObj;
		}
		const node = tagObj?.createNode ? tagObj.createNode(ctx.schema, value, ctx) : typeof tagObj?.nodeClass?.from === "function" ? tagObj.nodeClass.from(ctx.schema, value, ctx) : new Scalar(value);
		if (tagName) node.tag = tagName;
		else if (!tagObj.default) node.tag = tagObj.tag;
		if (ref) ref.node = node;
		return node;
	}
	function collectionFromPath(schema, path, value) {
		let v = value;
		for (let i = path.length - 1; i >= 0; --i) {
			const k = path[i];
			if (typeof k === "number" && Number.isInteger(k) && k >= 0) {
				const a = [];
				a[k] = v;
				v = a;
			} else v = /* @__PURE__ */ new Map([[k, v]]);
		}
		return createNode(v, void 0, {
			aliasDuplicateObjects: false,
			keepUndefined: false,
			onAnchor: () => {
				throw new Error("This should not happen, please report a bug.");
			},
			schema,
			sourceObjects: /* @__PURE__ */ new Map()
		});
	}
	var isEmptyPath = (path) => path == null || typeof path === "object" && !!path[Symbol.iterator]().next().done;
	var Collection = class extends NodeBase {
		constructor(type, schema) {
			super(type);
			Object.defineProperty(this, "schema", {
				value: schema,
				configurable: true,
				enumerable: false,
				writable: true
			});
		}
		/**
		* Create a copy of this collection.
		*
		* @param schema - If defined, overwrites the original's schema
		*/
		clone(schema) {
			const copy = Object.create(Object.getPrototypeOf(this), Object.getOwnPropertyDescriptors(this));
			if (schema) copy.schema = schema;
			copy.items = copy.items.map((it) => isNode(it) || isPair(it) ? it.clone(schema) : it);
			if (this.range) copy.range = this.range.slice();
			return copy;
		}
		/**
		* Adds a value to the collection. For `!!map` and `!!omap` the value must
		* be a Pair instance or a `{ key, value }` object, which may not have a key
		* that already exists in the map.
		*/
		addIn(path, value) {
			if (isEmptyPath(path)) this.add(value);
			else {
				const [key, ...rest] = path;
				const node = this.get(key, true);
				if (isCollection(node)) node.addIn(rest, value);
				else if (node === void 0 && this.schema) this.set(key, collectionFromPath(this.schema, rest, value));
				else throw new Error(`Expected YAML collection at ${key}. Remaining path: ${rest}`);
			}
		}
		/**
		* Removes a value from the collection.
		* @returns `true` if the item was found and removed.
		*/
		deleteIn(path) {
			const [key, ...rest] = path;
			if (rest.length === 0) return this.delete(key);
			const node = this.get(key, true);
			if (isCollection(node)) return node.deleteIn(rest);
			else throw new Error(`Expected YAML collection at ${key}. Remaining path: ${rest}`);
		}
		/**
		* Returns item at `key`, or `undefined` if not found. By default unwraps
		* scalar values from their surrounding node; to disable set `keepScalar` to
		* `true` (collections are always returned intact).
		*/
		getIn(path, keepScalar) {
			const [key, ...rest] = path;
			const node = this.get(key, true);
			if (rest.length === 0) return !keepScalar && isScalar(node) ? node.value : node;
			else return isCollection(node) ? node.getIn(rest, keepScalar) : void 0;
		}
		hasAllNullValues(allowScalar) {
			return this.items.every((node) => {
				if (!isPair(node)) return false;
				const n = node.value;
				return n == null || allowScalar && isScalar(n) && n.value == null && !n.commentBefore && !n.comment && !n.tag;
			});
		}
		/**
		* Checks if the collection includes a value with the key `key`.
		*/
		hasIn(path) {
			const [key, ...rest] = path;
			if (rest.length === 0) return this.has(key);
			const node = this.get(key, true);
			return isCollection(node) ? node.hasIn(rest) : false;
		}
		/**
		* Sets a value in this collection. For `!!set`, `value` needs to be a
		* boolean to add/remove the item from the set.
		*/
		setIn(path, value) {
			const [key, ...rest] = path;
			if (rest.length === 0) this.set(key, value);
			else {
				const node = this.get(key, true);
				if (isCollection(node)) node.setIn(rest, value);
				else if (node === void 0 && this.schema) this.set(key, collectionFromPath(this.schema, rest, value));
				else throw new Error(`Expected YAML collection at ${key}. Remaining path: ${rest}`);
			}
		}
	};
	/**
	* Stringifies a comment.
	*
	* Empty comment lines are left empty,
	* lines consisting of a single space are replaced by `#`,
	* and all other lines are prefixed with a `#`.
	*/
	var stringifyComment = (str) => str.replace(/^(?!$)(?: $)?/gm, "#");
	function indentComment(comment, indent) {
		if (/^\n+$/.test(comment)) return comment.substring(1);
		return indent ? comment.replace(/^(?! *$)/gm, indent) : comment;
	}
	var lineComment = (str, indent, comment) => str.endsWith("\n") ? indentComment(comment, indent) : comment.includes("\n") ? "\n" + indentComment(comment, indent) : (str.endsWith(" ") ? "" : " ") + comment;
	var FOLD_FLOW = "flow";
	var FOLD_BLOCK = "block";
	var FOLD_QUOTED = "quoted";
	/**
	* Tries to keep input at up to `lineWidth` characters, splitting only on spaces
	* not followed by newlines or spaces unless `mode` is `'quoted'`. Lines are
	* terminated with `\n` and started with `indent`.
	*/
	function foldFlowLines(text, indent, mode = "flow", { indentAtStart, lineWidth = 80, minContentWidth = 20, onFold, onOverflow } = {}) {
		if (!lineWidth || lineWidth < 0) return text;
		if (lineWidth < minContentWidth) minContentWidth = 0;
		const endStep = Math.max(1 + minContentWidth, 1 + lineWidth - indent.length);
		if (text.length <= endStep) return text;
		const folds = [];
		const escapedFolds = {};
		let end = lineWidth - indent.length;
		if (typeof indentAtStart === "number") {
			if (indentAtStart > lineWidth - Math.max(2, minContentWidth)) folds.push(0);
			else end = lineWidth - indentAtStart;
		}
		let split = void 0;
		let prev = void 0;
		let overflow = false;
		let i = -1;
		let escStart = -1;
		let escEnd = -1;
		if (mode === "block") {
			i = consumeMoreIndentedLines(text, i, indent.length);
			if (i !== -1) end = i + endStep;
		}
		for (let ch; ch = text[i += 1];) {
			if (mode === "quoted" && ch === "\\") {
				escStart = i;
				switch (text[i + 1]) {
					case "x":
						i += 3;
						break;
					case "u":
						i += 5;
						break;
					case "U":
						i += 9;
						break;
					default: i += 1;
				}
				escEnd = i;
			}
			if (ch === "\n") {
				if (mode === "block") i = consumeMoreIndentedLines(text, i, indent.length);
				end = i + indent.length + endStep;
				split = void 0;
			} else {
				if (ch === " " && prev && prev !== " " && prev !== "\n" && prev !== "	") {
					const next = text[i + 1];
					if (next && next !== " " && next !== "\n" && next !== "	") split = i;
				}
				if (i >= end) {
					if (split) {
						folds.push(split);
						end = split + endStep;
						split = void 0;
					} else if (mode === "quoted") {
						while (prev === " " || prev === "	") {
							prev = ch;
							ch = text[i += 1];
							overflow = true;
						}
						const j = i > escEnd + 1 ? i - 2 : escStart - 1;
						if (escapedFolds[j]) return text;
						folds.push(j);
						escapedFolds[j] = true;
						end = j + endStep;
						split = void 0;
					} else overflow = true;
				}
			}
			prev = ch;
		}
		if (overflow && onOverflow) onOverflow();
		if (folds.length === 0) return text;
		if (onFold) onFold();
		let res = text.slice(0, folds[0]);
		for (let i = 0; i < folds.length; ++i) {
			const fold = folds[i];
			const end = folds[i + 1] || text.length;
			if (fold === 0) res = `\n${indent}${text.slice(0, end)}`;
			else {
				if (mode === "quoted" && escapedFolds[fold]) res += `${text[fold]}\\`;
				res += `\n${indent}${text.slice(fold + 1, end)}`;
			}
		}
		return res;
	}
	/**
	* Presumes `i + 1` is at the start of a line
	* @returns index of last newline in more-indented block
	*/
	function consumeMoreIndentedLines(text, i, indent) {
		let end = i;
		let start = i + 1;
		let ch = text[start];
		while (ch === " " || ch === "	") if (i < start + indent) ch = text[++i];
		else {
			do
				ch = text[++i];
			while (ch && ch !== "\n");
			end = i;
			start = i + 1;
			ch = text[start];
		}
		return end;
	}
	var getFoldOptions = (ctx, isBlock) => ({
		indentAtStart: isBlock ? ctx.indent.length : ctx.indentAtStart,
		lineWidth: ctx.options.lineWidth,
		minContentWidth: ctx.options.minContentWidth
	});
	var containsDocumentMarker = (str) => /^(%|---|\.\.\.)/m.test(str);
	function lineLengthOverLimit(str, lineWidth, indentLength) {
		if (!lineWidth || lineWidth < 0) return false;
		const limit = lineWidth - indentLength;
		const strLen = str.length;
		if (strLen <= limit) return false;
		for (let i = 0, start = 0; i < strLen; ++i) if (str[i] === "\n") {
			if (i - start > limit) return true;
			start = i + 1;
			if (strLen - start <= limit) return false;
		}
		return true;
	}
	function doubleQuotedString(value, ctx) {
		const json = JSON.stringify(value);
		if (ctx.options.doubleQuotedAsJSON) return json;
		const { implicitKey } = ctx;
		const minMultiLineLength = ctx.options.doubleQuotedMinMultiLineLength;
		const indent = ctx.indent || (containsDocumentMarker(value) ? "  " : "");
		let str = "";
		let start = 0;
		for (let i = 0, ch = json[i]; ch; ch = json[++i]) {
			if (ch === " " && json[i + 1] === "\\" && json[i + 2] === "n") {
				str += json.slice(start, i) + "\\ ";
				i += 1;
				start = i;
				ch = "\\";
			}
			if (ch === "\\") switch (json[i + 1]) {
				case "u":
					{
						str += json.slice(start, i);
						const code = json.substr(i + 2, 4);
						switch (code) {
							case "0000":
								str += "\\0";
								break;
							case "0007":
								str += "\\a";
								break;
							case "000b":
								str += "\\v";
								break;
							case "001b":
								str += "\\e";
								break;
							case "0085":
								str += "\\N";
								break;
							case "00a0":
								str += "\\_";
								break;
							case "2028":
								str += "\\L";
								break;
							case "2029":
								str += "\\P";
								break;
							default: if (code.substr(0, 2) === "00") str += "\\x" + code.substr(2);
							else str += json.substr(i, 6);
						}
						i += 5;
						start = i + 1;
					}
					break;
				case "n":
					if (implicitKey || json[i + 2] === "\"" || json.length < minMultiLineLength) i += 1;
					else {
						str += json.slice(start, i) + "\n\n";
						while (json[i + 2] === "\\" && json[i + 3] === "n" && json[i + 4] !== "\"") {
							str += "\n";
							i += 2;
						}
						str += indent;
						if (json[i + 2] === " ") str += "\\";
						i += 1;
						start = i + 1;
					}
					break;
				default: i += 1;
			}
		}
		str = start ? str + json.slice(start) : json;
		return implicitKey ? str : foldFlowLines(str, indent, FOLD_QUOTED, getFoldOptions(ctx, false));
	}
	function singleQuotedString(value, ctx) {
		if (ctx.options.singleQuote === false || ctx.implicitKey && value.includes("\n") || /[ \t]\n|\n[ \t]/.test(value)) return doubleQuotedString(value, ctx);
		const indent = ctx.indent || (containsDocumentMarker(value) ? "  " : "");
		const res = "'" + value.replace(/'/g, "''").replace(/\n+/g, `$&\n${indent}`) + "'";
		return ctx.implicitKey ? res : foldFlowLines(res, indent, FOLD_FLOW, getFoldOptions(ctx, false));
	}
	function quotedString(value, ctx) {
		const { singleQuote } = ctx.options;
		let qs;
		if (singleQuote === false) qs = doubleQuotedString;
		else {
			const hasDouble = value.includes("\"");
			const hasSingle = value.includes("'");
			if (hasDouble && !hasSingle) qs = singleQuotedString;
			else if (hasSingle && !hasDouble) qs = doubleQuotedString;
			else qs = singleQuote ? singleQuotedString : doubleQuotedString;
		}
		return qs(value, ctx);
	}
	var blockEndNewlines;
	try {
		blockEndNewlines = /* @__PURE__ */ new RegExp("(^|(?<!\n))\n+(?!\n|$)", "g");
	} catch {
		blockEndNewlines = /\n+(?!\n|$)/g;
	}
	function blockString({ comment, type, value }, ctx, onComment, onChompKeep) {
		const { blockQuote, commentString, lineWidth } = ctx.options;
		if (!blockQuote || /\n[\t ]+$/.test(value)) return quotedString(value, ctx);
		const indent = ctx.indent || (ctx.forceBlockIndent || containsDocumentMarker(value) ? "  " : "");
		const literal = blockQuote === "literal" ? true : blockQuote === "folded" || type === Scalar.BLOCK_FOLDED ? false : type === Scalar.BLOCK_LITERAL ? true : !lineLengthOverLimit(value, lineWidth, indent.length);
		if (!value) return literal ? "|\n" : ">\n";
		let chomp;
		let endStart;
		for (endStart = value.length; endStart > 0; --endStart) {
			const ch = value[endStart - 1];
			if (ch !== "\n" && ch !== "	" && ch !== " ") break;
		}
		let end = value.substring(endStart);
		const endNlPos = end.indexOf("\n");
		if (endNlPos === -1) chomp = "-";
		else if (value === end || endNlPos !== end.length - 1) {
			chomp = "+";
			if (onChompKeep) onChompKeep();
		} else chomp = "";
		if (end) {
			value = value.slice(0, -end.length);
			if (end[end.length - 1] === "\n") end = end.slice(0, -1);
			end = end.replace(blockEndNewlines, `$&${indent}`);
		}
		let startWithSpace = false;
		let startEnd;
		let startNlPos = -1;
		for (startEnd = 0; startEnd < value.length; ++startEnd) {
			const ch = value[startEnd];
			if (ch === " ") startWithSpace = true;
			else if (ch === "\n") startNlPos = startEnd;
			else break;
		}
		let start = value.substring(0, startNlPos < startEnd ? startNlPos + 1 : startEnd);
		if (start) {
			value = value.substring(start.length);
			start = start.replace(/\n+/g, `$&${indent}`);
		}
		let header = (startWithSpace ? indent ? "2" : "1" : "") + chomp;
		if (comment) {
			header += " " + commentString(comment.replace(/ ?[\r\n]+/g, " "));
			if (onComment) onComment();
		}
		if (!literal) {
			const foldedValue = value.replace(/\n+/g, "\n$&").replace(/(?:^|\n)([\t ].*)(?:([\n\t ]*)\n(?![\n\t ]))?/g, "$1$2").replace(/\n+/g, `$&${indent}`);
			let literalFallback = false;
			const foldOptions = getFoldOptions(ctx, true);
			if (blockQuote !== "folded" && type !== Scalar.BLOCK_FOLDED) foldOptions.onOverflow = () => {
				literalFallback = true;
			};
			const body = foldFlowLines(`${start}${foldedValue}${end}`, indent, FOLD_BLOCK, foldOptions);
			if (!literalFallback) return `>${header}\n${indent}${body}`;
		}
		value = value.replace(/\n+/g, `$&${indent}`);
		return `|${header}\n${indent}${start}${value}${end}`;
	}
	function plainString(item, ctx, onComment, onChompKeep) {
		const { type, value } = item;
		const { actualString, implicitKey, indent, indentStep, inFlow } = ctx;
		if (implicitKey && value.includes("\n") || inFlow && /[[\]{},]/.test(value)) return quotedString(value, ctx);
		if (/^[\n\t ,[\]{}#&*!|>'"%@`]|^[?-]$|^[?-][ \t]|[\n:][ \t]|[ \t]\n|[\n\t ]#|[\n\t :]$/.test(value)) return implicitKey || inFlow || !value.includes("\n") ? quotedString(value, ctx) : blockString(item, ctx, onComment, onChompKeep);
		if (!implicitKey && !inFlow && type !== Scalar.PLAIN && value.includes("\n")) return blockString(item, ctx, onComment, onChompKeep);
		if (containsDocumentMarker(value)) {
			if (indent === "") {
				ctx.forceBlockIndent = true;
				return blockString(item, ctx, onComment, onChompKeep);
			} else if (implicitKey && indent === indentStep) return quotedString(value, ctx);
		}
		const str = value.replace(/\n+/g, `$&\n${indent}`);
		if (actualString) {
			const test = (tag) => tag.default && tag.tag !== "tag:yaml.org,2002:str" && tag.test?.test(str);
			const { compat, tags } = ctx.doc.schema;
			if (tags.some(test) || compat?.some(test)) return quotedString(value, ctx);
		}
		return implicitKey ? str : foldFlowLines(str, indent, FOLD_FLOW, getFoldOptions(ctx, false));
	}
	function stringifyString(item, ctx, onComment, onChompKeep) {
		const { implicitKey, inFlow } = ctx;
		const ss = typeof item.value === "string" ? item : Object.assign({}, item, { value: String(item.value) });
		let { type } = item;
		if (type !== Scalar.QUOTE_DOUBLE) {
			if (/[\x00-\x08\x0b-\x1f\x7f-\x9f\u{D800}-\u{DFFF}]/u.test(ss.value)) type = Scalar.QUOTE_DOUBLE;
		}
		const _stringify = (_type) => {
			switch (_type) {
				case Scalar.BLOCK_FOLDED:
				case Scalar.BLOCK_LITERAL: return implicitKey || inFlow ? quotedString(ss.value, ctx) : blockString(ss, ctx, onComment, onChompKeep);
				case Scalar.QUOTE_DOUBLE: return doubleQuotedString(ss.value, ctx);
				case Scalar.QUOTE_SINGLE: return singleQuotedString(ss.value, ctx);
				case Scalar.PLAIN: return plainString(ss, ctx, onComment, onChompKeep);
				default: return null;
			}
		};
		let res = _stringify(type);
		if (res === null) {
			const { defaultKeyType, defaultStringType } = ctx.options;
			const t = implicitKey && defaultKeyType || defaultStringType;
			res = _stringify(t);
			if (res === null) throw new Error(`Unsupported default string type ${t}`);
		}
		return res;
	}
	function createStringifyContext(doc, options) {
		const opt = Object.assign({
			blockQuote: true,
			commentString: stringifyComment,
			defaultKeyType: null,
			defaultStringType: "PLAIN",
			directives: null,
			doubleQuotedAsJSON: false,
			doubleQuotedMinMultiLineLength: 40,
			falseStr: "false",
			flowCollectionPadding: true,
			indentSeq: true,
			lineWidth: 80,
			minContentWidth: 20,
			nullStr: "null",
			simpleKeys: false,
			singleQuote: null,
			trailingComma: false,
			trueStr: "true",
			verifyAliasOrder: true
		}, doc.schema.toStringOptions, options);
		let inFlow;
		switch (opt.collectionStyle) {
			case "block":
				inFlow = false;
				break;
			case "flow":
				inFlow = true;
				break;
			default: inFlow = null;
		}
		return {
			anchors: /* @__PURE__ */ new Set(),
			doc,
			flowCollectionPadding: opt.flowCollectionPadding ? " " : "",
			indent: "",
			indentStep: typeof opt.indent === "number" ? " ".repeat(opt.indent) : "  ",
			inFlow,
			options: opt
		};
	}
	function getTagObject(tags, item) {
		if (item.tag) {
			const match = tags.filter((t) => t.tag === item.tag);
			if (match.length > 0) return match.find((t) => t.format === item.format) ?? match[0];
		}
		let tagObj = void 0;
		let obj;
		if (isScalar(item)) {
			obj = item.value;
			let match = tags.filter((t) => t.identify?.(obj));
			if (match.length > 1) {
				const testMatch = match.filter((t) => t.test);
				if (testMatch.length > 0) match = testMatch;
			}
			tagObj = match.find((t) => t.format === item.format) ?? match.find((t) => !t.format);
		} else {
			obj = item;
			tagObj = tags.find((t) => t.nodeClass && obj instanceof t.nodeClass);
		}
		if (!tagObj) {
			const name = obj?.constructor?.name ?? (obj === null ? "null" : typeof obj);
			throw new Error(`Tag not resolved for ${name} value`);
		}
		return tagObj;
	}
	function stringifyProps(node, tagObj, { anchors, doc }) {
		if (!doc.directives) return "";
		const props = [];
		const anchor = (isScalar(node) || isCollection(node)) && node.anchor;
		if (anchor && anchorIsValid(anchor)) {
			anchors.add(anchor);
			props.push(`&${anchor}`);
		}
		const tag = node.tag ?? (tagObj.default ? null : tagObj.tag);
		if (tag) props.push(doc.directives.tagString(tag));
		return props.join(" ");
	}
	function stringify(item, ctx, onComment, onChompKeep) {
		if (isPair(item)) return item.toString(ctx, onComment, onChompKeep);
		if (isAlias(item)) {
			if (ctx.doc.directives) return item.toString(ctx);
			if (ctx.resolvedAliases?.has(item)) throw new TypeError(`Cannot stringify circular structure without alias nodes`);
			else {
				if (ctx.resolvedAliases) ctx.resolvedAliases.add(item);
				else ctx.resolvedAliases = /* @__PURE__ */ new Set([item]);
				item = item.resolve(ctx.doc);
			}
		}
		let tagObj = void 0;
		const node = isNode(item) ? item : ctx.doc.createNode(item, { onTagObj: (o) => tagObj = o });
		tagObj ?? (tagObj = getTagObject(ctx.doc.schema.tags, node));
		const props = stringifyProps(node, tagObj, ctx);
		if (props.length > 0) ctx.indentAtStart = (ctx.indentAtStart ?? 0) + props.length + 1;
		const str = typeof tagObj.stringify === "function" ? tagObj.stringify(node, ctx, onComment, onChompKeep) : isScalar(node) ? stringifyString(node, ctx, onComment, onChompKeep) : node.toString(ctx, onComment, onChompKeep);
		if (!props) return str;
		return isScalar(node) || str[0] === "{" || str[0] === "[" ? `${props} ${str}` : `${props}\n${ctx.indent}${str}`;
	}
	function stringifyPair({ key, value }, ctx, onComment, onChompKeep) {
		const { allNullValues, doc, indent, indentStep, options: { commentString, indentSeq, simpleKeys } } = ctx;
		let keyComment = isNode(key) && key.comment || null;
		if (simpleKeys) {
			if (keyComment) throw new Error("With simple keys, key nodes cannot have comments");
			if (isCollection(key) || !isNode(key) && typeof key === "object") throw new Error("With simple keys, collection cannot be used as a key value");
		}
		let explicitKey = !simpleKeys && (!key || keyComment && value == null && !ctx.inFlow || isCollection(key) || (isScalar(key) ? key.type === Scalar.BLOCK_FOLDED || key.type === Scalar.BLOCK_LITERAL : typeof key === "object"));
		ctx = Object.assign({}, ctx, {
			allNullValues: false,
			implicitKey: !explicitKey && (simpleKeys || !allNullValues),
			indent: indent + indentStep
		});
		let keyCommentDone = false;
		let chompKeep = false;
		let str = stringify(key, ctx, () => keyCommentDone = true, () => chompKeep = true);
		if (!explicitKey && !ctx.inFlow && str.length > 1024) {
			if (simpleKeys) throw new Error("With simple keys, single line scalar must not span more than 1024 characters");
			explicitKey = true;
		}
		if (ctx.inFlow) {
			if (allNullValues || value == null) {
				if (keyCommentDone && onComment) onComment();
				return str === "" ? "?" : explicitKey ? `? ${str}` : str;
			}
		} else if (allNullValues && !simpleKeys || value == null && explicitKey) {
			str = `? ${str}`;
			if (keyComment && !keyCommentDone) str += lineComment(str, ctx.indent, commentString(keyComment));
			else if (chompKeep && onChompKeep) onChompKeep();
			return str;
		}
		if (keyCommentDone) keyComment = null;
		if (explicitKey) {
			if (keyComment) str += lineComment(str, ctx.indent, commentString(keyComment));
			str = `? ${str}\n${indent}:`;
		} else {
			str = `${str}:`;
			if (keyComment) str += lineComment(str, ctx.indent, commentString(keyComment));
		}
		let vsb, vcb, valueComment;
		if (isNode(value)) {
			vsb = !!value.spaceBefore;
			vcb = value.commentBefore;
			valueComment = value.comment;
		} else {
			vsb = false;
			vcb = null;
			valueComment = null;
			if (value && typeof value === "object") value = doc.createNode(value);
		}
		ctx.implicitKey = false;
		if (!explicitKey && !keyComment && isScalar(value)) ctx.indentAtStart = str.length + 1;
		chompKeep = false;
		if (!indentSeq && indentStep.length >= 2 && !ctx.inFlow && !explicitKey && isSeq(value) && !value.flow && !value.tag && !value.anchor) ctx.indent = ctx.indent.substring(2);
		let valueCommentDone = false;
		const valueStr = stringify(value, ctx, () => valueCommentDone = true, () => chompKeep = true);
		let ws = " ";
		if (keyComment || vsb || vcb) {
			ws = vsb ? "\n" : "";
			if (vcb) {
				const cs = commentString(vcb);
				ws += `\n${indentComment(cs, ctx.indent)}`;
			}
			if (valueStr === "" && !ctx.inFlow) {
				if (ws === "\n" && valueComment) ws = "\n\n";
			} else ws += `\n${ctx.indent}`;
		} else if (!explicitKey && isCollection(value)) {
			const vs0 = valueStr[0];
			const nl0 = valueStr.indexOf("\n");
			const hasNewline = nl0 !== -1;
			const flow = ctx.inFlow ?? value.flow ?? value.items.length === 0;
			if (hasNewline || !flow) {
				let hasPropsLine = false;
				if (hasNewline && (vs0 === "&" || vs0 === "!")) {
					let sp0 = valueStr.indexOf(" ");
					if (vs0 === "&" && sp0 !== -1 && sp0 < nl0 && valueStr[sp0 + 1] === "!") sp0 = valueStr.indexOf(" ", sp0 + 1);
					if (sp0 === -1 || nl0 < sp0) hasPropsLine = true;
				}
				if (!hasPropsLine) ws = `\n${ctx.indent}`;
			}
		} else if (valueStr === "" || valueStr[0] === "\n") ws = "";
		str += ws + valueStr;
		if (ctx.inFlow) {
			if (valueCommentDone && onComment) onComment();
		} else if (valueComment && !valueCommentDone) str += lineComment(str, ctx.indent, commentString(valueComment));
		else if (chompKeep && onChompKeep) onChompKeep();
		return str;
	}
	function warn(logLevel, warning) {
		if (logLevel === "debug" || logLevel === "warn") console.warn(warning);
	}
	var MERGE_KEY = "<<";
	var merge = {
		identify: (value) => value === MERGE_KEY || typeof value === "symbol" && value.description === MERGE_KEY,
		default: "key",
		tag: "tag:yaml.org,2002:merge",
		test: /^<<$/,
		resolve: () => Object.assign(new Scalar(Symbol(MERGE_KEY)), { addToJSMap: addMergeToJSMap }),
		stringify: () => MERGE_KEY
	};
	var isMergeKey = (ctx, key) => (merge.identify(key) || isScalar(key) && (!key.type || key.type === Scalar.PLAIN) && merge.identify(key.value)) && ctx?.doc.schema.tags.some((tag) => tag.tag === merge.tag && tag.default);
	function addMergeToJSMap(ctx, map, value) {
		const source = resolveAliasValue(ctx, value);
		if (isSeq(source)) for (const it of source.items) mergeValue(ctx, map, it);
		else if (Array.isArray(source)) for (const it of source) mergeValue(ctx, map, it);
		else mergeValue(ctx, map, source);
	}
	function mergeValue(ctx, map, value) {
		const source = resolveAliasValue(ctx, value);
		if (!isMap(source)) throw new Error("Merge sources must be maps or map aliases");
		const srcMap = source.toJSON(null, ctx, Map);
		for (const [key, value] of srcMap) if (map instanceof Map) {
			if (!map.has(key)) map.set(key, value);
		} else if (map instanceof Set) map.add(key);
		else if (!Object.prototype.hasOwnProperty.call(map, key)) Object.defineProperty(map, key, {
			value,
			writable: true,
			enumerable: true,
			configurable: true
		});
		return map;
	}
	function resolveAliasValue(ctx, value) {
		return ctx && isAlias(value) ? value.resolve(ctx.doc, ctx) : value;
	}
	function addPairToJSMap(ctx, map, { key, value }) {
		if (isNode(key) && key.addToJSMap) key.addToJSMap(ctx, map, value);
		else if (isMergeKey(ctx, key)) addMergeToJSMap(ctx, map, value);
		else {
			const jsKey = toJS(key, "", ctx);
			if (map instanceof Map) map.set(jsKey, toJS(value, jsKey, ctx));
			else if (map instanceof Set) map.add(jsKey);
			else {
				const stringKey = stringifyKey(key, jsKey, ctx);
				const jsValue = toJS(value, stringKey, ctx);
				if (stringKey in map) Object.defineProperty(map, stringKey, {
					value: jsValue,
					writable: true,
					enumerable: true,
					configurable: true
				});
				else map[stringKey] = jsValue;
			}
		}
		return map;
	}
	function stringifyKey(key, jsKey, ctx) {
		if (jsKey === null) return "";
		if (typeof jsKey !== "object") return String(jsKey);
		if (isNode(key) && ctx?.doc) {
			const strCtx = createStringifyContext(ctx.doc, {});
			strCtx.anchors = /* @__PURE__ */ new Set();
			for (const node of ctx.anchors.keys()) strCtx.anchors.add(node.anchor);
			strCtx.inFlow = true;
			strCtx.inStringifyKey = true;
			const strKey = key.toString(strCtx);
			if (!ctx.mapKeyWarned) {
				let jsonStr = JSON.stringify(strKey);
				if (jsonStr.length > 40) jsonStr = jsonStr.substring(0, 36) + "...\"";
				warn(ctx.doc.options.logLevel, `Keys with collection values will be stringified due to JS Object restrictions: ${jsonStr}. Set mapAsMap: true to use object keys.`);
				ctx.mapKeyWarned = true;
			}
			return strKey;
		}
		return JSON.stringify(jsKey);
	}
	function createPair(key, value, ctx) {
		return new Pair(createNode(key, void 0, ctx), createNode(value, void 0, ctx));
	}
	var Pair = class Pair {
		constructor(key, value = null) {
			Object.defineProperty(this, NODE_TYPE, { value: PAIR });
			this.key = key;
			this.value = value;
		}
		clone(schema) {
			let { key, value } = this;
			if (isNode(key)) key = key.clone(schema);
			if (isNode(value)) value = value.clone(schema);
			return new Pair(key, value);
		}
		toJSON(_, ctx) {
			return addPairToJSMap(ctx, ctx?.mapAsMap ? /* @__PURE__ */ new Map() : {}, this);
		}
		toString(ctx, onComment, onChompKeep) {
			return ctx?.doc ? stringifyPair(this, ctx, onComment, onChompKeep) : JSON.stringify(this);
		}
	};
	function stringifyCollection(collection, ctx, options) {
		return (ctx.inFlow ?? collection.flow ? stringifyFlowCollection : stringifyBlockCollection)(collection, ctx, options);
	}
	function stringifyBlockCollection({ comment, items }, ctx, { blockItemPrefix, flowChars, itemIndent, onChompKeep, onComment }) {
		const { indent, options: { commentString } } = ctx;
		const itemCtx = Object.assign({}, ctx, {
			indent: itemIndent,
			type: null
		});
		let chompKeep = false;
		const lines = [];
		for (let i = 0; i < items.length; ++i) {
			const item = items[i];
			let comment = null;
			if (isNode(item)) {
				if (!chompKeep && item.spaceBefore) lines.push("");
				addCommentBefore(ctx, lines, item.commentBefore, chompKeep);
				if (item.comment) comment = item.comment;
			} else if (isPair(item)) {
				const ik = isNode(item.key) ? item.key : null;
				if (ik) {
					if (!chompKeep && ik.spaceBefore) lines.push("");
					addCommentBefore(ctx, lines, ik.commentBefore, chompKeep);
				}
			}
			chompKeep = false;
			let str = stringify(item, itemCtx, () => comment = null, () => chompKeep = true);
			if (comment) str += lineComment(str, itemIndent, commentString(comment));
			if (chompKeep && comment) chompKeep = false;
			lines.push(blockItemPrefix + str);
		}
		let str;
		if (lines.length === 0) str = flowChars.start + flowChars.end;
		else {
			str = lines[0];
			for (let i = 1; i < lines.length; ++i) {
				const line = lines[i];
				str += line ? `\n${indent}${line}` : "\n";
			}
		}
		if (comment) {
			str += "\n" + indentComment(commentString(comment), indent);
			if (onComment) onComment();
		} else if (chompKeep && onChompKeep) onChompKeep();
		return str;
	}
	function stringifyFlowCollection({ items }, ctx, { flowChars, itemIndent }) {
		const { indent, indentStep, flowCollectionPadding: fcPadding, options: { commentString } } = ctx;
		itemIndent += indentStep;
		const itemCtx = Object.assign({}, ctx, {
			indent: itemIndent,
			inFlow: true,
			type: null
		});
		let reqNewline = false;
		let linesAtValue = 0;
		const lines = [];
		for (let i = 0; i < items.length; ++i) {
			const item = items[i];
			let comment = null;
			if (isNode(item)) {
				if (item.spaceBefore) lines.push("");
				addCommentBefore(ctx, lines, item.commentBefore, false);
				if (item.comment) comment = item.comment;
			} else if (isPair(item)) {
				const ik = isNode(item.key) ? item.key : null;
				if (ik) {
					if (ik.spaceBefore) lines.push("");
					addCommentBefore(ctx, lines, ik.commentBefore, false);
					if (ik.comment) reqNewline = true;
				}
				const iv = isNode(item.value) ? item.value : null;
				if (iv) {
					if (iv.comment) comment = iv.comment;
					if (iv.commentBefore) reqNewline = true;
				} else if (item.value == null && ik?.comment) comment = ik.comment;
			}
			if (comment) reqNewline = true;
			let str = stringify(item, itemCtx, () => comment = null);
			reqNewline || (reqNewline = lines.length > linesAtValue || str.includes("\n"));
			if (i < items.length - 1) str += ",";
			else if (ctx.options.trailingComma) {
				if (ctx.options.lineWidth > 0) reqNewline || (reqNewline = lines.reduce((sum, line) => sum + line.length + 2, 2) + (str.length + 2) > ctx.options.lineWidth);
				if (reqNewline) str += ",";
			}
			if (comment) str += lineComment(str, itemIndent, commentString(comment));
			lines.push(str);
			linesAtValue = lines.length;
		}
		const { start, end } = flowChars;
		if (lines.length === 0) return start + end;
		else {
			if (!reqNewline) {
				const len = lines.reduce((sum, line) => sum + line.length + 2, 2);
				reqNewline = ctx.options.lineWidth > 0 && len > ctx.options.lineWidth;
			}
			if (reqNewline) {
				let str = start;
				for (const line of lines) str += line ? `\n${indentStep}${indent}${line}` : "\n";
				return `${str}\n${indent}${end}`;
			} else return `${start}${fcPadding}${lines.join(" ")}${fcPadding}${end}`;
		}
	}
	function addCommentBefore({ indent, options: { commentString } }, lines, comment, chompKeep) {
		if (comment && chompKeep) comment = comment.replace(/^\n+/, "");
		if (comment) {
			const ic = indentComment(commentString(comment), indent);
			lines.push(ic.trimStart());
		}
	}
	function findPair(items, key) {
		const k = isScalar(key) ? key.value : key;
		for (const it of items) if (isPair(it)) {
			if (it.key === key || it.key === k) return it;
			if (isScalar(it.key) && it.key.value === k) return it;
		}
	}
	var YAMLMap = class extends Collection {
		static get tagName() {
			return "tag:yaml.org,2002:map";
		}
		constructor(schema) {
			super(MAP, schema);
			this.items = [];
		}
		/**
		* A generic collection parsing method that can be extended
		* to other node classes that inherit from YAMLMap
		*/
		static from(schema, obj, ctx) {
			const { keepUndefined, replacer } = ctx;
			const map = new this(schema);
			const add = (key, value) => {
				if (typeof replacer === "function") value = replacer.call(obj, key, value);
				else if (Array.isArray(replacer) && !replacer.includes(key)) return;
				if (value !== void 0 || keepUndefined) map.items.push(createPair(key, value, ctx));
			};
			if (obj instanceof Map) for (const [key, value] of obj) add(key, value);
			else if (obj && typeof obj === "object") for (const key of Object.keys(obj)) add(key, obj[key]);
			if (typeof schema.sortMapEntries === "function") map.items.sort(schema.sortMapEntries);
			return map;
		}
		/**
		* Adds a value to the collection.
		*
		* @param overwrite - If not set `true`, using a key that is already in the
		*   collection will throw. Otherwise, overwrites the previous value.
		*/
		add(pair, overwrite) {
			let _pair;
			if (isPair(pair)) _pair = pair;
			else if (!pair || typeof pair !== "object" || !("key" in pair)) _pair = new Pair(pair, pair?.value);
			else _pair = new Pair(pair.key, pair.value);
			const prev = findPair(this.items, _pair.key);
			const sortEntries = this.schema?.sortMapEntries;
			if (prev) {
				if (!overwrite) throw new Error(`Key ${_pair.key} already set`);
				if (isScalar(prev.value) && isScalarValue(_pair.value)) prev.value.value = _pair.value;
				else prev.value = _pair.value;
			} else if (sortEntries) {
				const i = this.items.findIndex((item) => sortEntries(_pair, item) < 0);
				if (i === -1) this.items.push(_pair);
				else this.items.splice(i, 0, _pair);
			} else this.items.push(_pair);
		}
		delete(key) {
			const it = findPair(this.items, key);
			if (!it) return false;
			return this.items.splice(this.items.indexOf(it), 1).length > 0;
		}
		get(key, keepScalar) {
			const node = findPair(this.items, key)?.value;
			return (!keepScalar && isScalar(node) ? node.value : node) ?? void 0;
		}
		has(key) {
			return !!findPair(this.items, key);
		}
		set(key, value) {
			this.add(new Pair(key, value), true);
		}
		/**
		* @param ctx - Conversion context, originally set in Document#toJS()
		* @param {Class} Type - If set, forces the returned collection type
		* @returns Instance of Type, Map, or Object
		*/
		toJSON(_, ctx, Type) {
			const map = Type ? new Type() : ctx?.mapAsMap ? /* @__PURE__ */ new Map() : {};
			if (ctx?.onCreate) ctx.onCreate(map);
			for (const item of this.items) addPairToJSMap(ctx, map, item);
			return map;
		}
		toString(ctx, onComment, onChompKeep) {
			if (!ctx) return JSON.stringify(this);
			for (const item of this.items) if (!isPair(item)) throw new Error(`Map items must all be pairs; found ${JSON.stringify(item)} instead`);
			if (!ctx.allNullValues && this.hasAllNullValues(false)) ctx = Object.assign({}, ctx, { allNullValues: true });
			return stringifyCollection(this, ctx, {
				blockItemPrefix: "",
				flowChars: {
					start: "{",
					end: "}"
				},
				itemIndent: ctx.indent || "",
				onChompKeep,
				onComment
			});
		}
	};
	var map = {
		collection: "map",
		default: true,
		nodeClass: YAMLMap,
		tag: "tag:yaml.org,2002:map",
		resolve(map, onError) {
			if (!isMap(map)) onError("Expected a mapping for this tag");
			return map;
		},
		createNode: (schema, obj, ctx) => YAMLMap.from(schema, obj, ctx)
	};
	var YAMLSeq = class extends Collection {
		static get tagName() {
			return "tag:yaml.org,2002:seq";
		}
		constructor(schema) {
			super(SEQ, schema);
			this.items = [];
		}
		add(value) {
			this.items.push(value);
		}
		/**
		* Removes a value from the collection.
		*
		* `key` must contain a representation of an integer for this to succeed.
		* It may be wrapped in a `Scalar`.
		*
		* @returns `true` if the item was found and removed.
		*/
		delete(key) {
			const idx = asItemIndex(key);
			if (typeof idx !== "number") return false;
			return this.items.splice(idx, 1).length > 0;
		}
		get(key, keepScalar) {
			const idx = asItemIndex(key);
			if (typeof idx !== "number") return void 0;
			const it = this.items[idx];
			return !keepScalar && isScalar(it) ? it.value : it;
		}
		/**
		* Checks if the collection includes a value with the key `key`.
		*
		* `key` must contain a representation of an integer for this to succeed.
		* It may be wrapped in a `Scalar`.
		*/
		has(key) {
			const idx = asItemIndex(key);
			return typeof idx === "number" && idx < this.items.length;
		}
		/**
		* Sets a value in this collection. For `!!set`, `value` needs to be a
		* boolean to add/remove the item from the set.
		*
		* If `key` does not contain a representation of an integer, this will throw.
		* It may be wrapped in a `Scalar`.
		*/
		set(key, value) {
			const idx = asItemIndex(key);
			if (typeof idx !== "number") throw new Error(`Expected a valid index, not ${key}.`);
			const prev = this.items[idx];
			if (isScalar(prev) && isScalarValue(value)) prev.value = value;
			else this.items[idx] = value;
		}
		toJSON(_, ctx) {
			const seq = [];
			if (ctx?.onCreate) ctx.onCreate(seq);
			let i = 0;
			for (const item of this.items) seq.push(toJS(item, String(i++), ctx));
			return seq;
		}
		toString(ctx, onComment, onChompKeep) {
			if (!ctx) return JSON.stringify(this);
			return stringifyCollection(this, ctx, {
				blockItemPrefix: "- ",
				flowChars: {
					start: "[",
					end: "]"
				},
				itemIndent: (ctx.indent || "") + "  ",
				onChompKeep,
				onComment
			});
		}
		static from(schema, obj, ctx) {
			const { replacer } = ctx;
			const seq = new this(schema);
			if (obj && Symbol.iterator in Object(obj)) {
				let i = 0;
				for (let it of obj) {
					if (typeof replacer === "function") {
						const key = obj instanceof Set ? it : String(i++);
						it = replacer.call(obj, key, it);
					}
					seq.items.push(createNode(it, void 0, ctx));
				}
			}
			return seq;
		}
	};
	function asItemIndex(key) {
		let idx = isScalar(key) ? key.value : key;
		if (idx && typeof idx === "string") idx = Number(idx);
		return typeof idx === "number" && Number.isInteger(idx) && idx >= 0 ? idx : null;
	}
	var seq = {
		collection: "seq",
		default: true,
		nodeClass: YAMLSeq,
		tag: "tag:yaml.org,2002:seq",
		resolve(seq, onError) {
			if (!isSeq(seq)) onError("Expected a sequence for this tag");
			return seq;
		},
		createNode: (schema, obj, ctx) => YAMLSeq.from(schema, obj, ctx)
	};
	var string = {
		identify: (value) => typeof value === "string",
		default: true,
		tag: "tag:yaml.org,2002:str",
		resolve: (str) => str,
		stringify(item, ctx, onComment, onChompKeep) {
			ctx = Object.assign({ actualString: true }, ctx);
			return stringifyString(item, ctx, onComment, onChompKeep);
		}
	};
	var nullTag = {
		identify: (value) => value == null,
		createNode: () => new Scalar(null),
		default: true,
		tag: "tag:yaml.org,2002:null",
		test: /^(?:~|[Nn]ull|NULL)?$/,
		resolve: () => new Scalar(null),
		stringify: ({ source }, ctx) => typeof source === "string" && nullTag.test.test(source) ? source : ctx.options.nullStr
	};
	var boolTag = {
		identify: (value) => typeof value === "boolean",
		default: true,
		tag: "tag:yaml.org,2002:bool",
		test: /^(?:[Tt]rue|TRUE|[Ff]alse|FALSE)$/,
		resolve: (str) => new Scalar(str[0] === "t" || str[0] === "T"),
		stringify({ source, value }, ctx) {
			if (source && boolTag.test.test(source)) {
				if (value === (source[0] === "t" || source[0] === "T")) return source;
			}
			return value ? ctx.options.trueStr : ctx.options.falseStr;
		}
	};
	function stringifyNumber({ format, minFractionDigits, tag, value }) {
		if (typeof value === "bigint") return String(value);
		const num = typeof value === "number" ? value : Number(value);
		if (!isFinite(num)) return isNaN(num) ? ".nan" : num < 0 ? "-.inf" : ".inf";
		let n = Object.is(value, -0) ? "-0" : JSON.stringify(value);
		if (!format && minFractionDigits && (!tag || tag === "tag:yaml.org,2002:float") && /^-?\d/.test(n) && !n.includes("e")) {
			let i = n.indexOf(".");
			if (i < 0) {
				i = n.length;
				n += ".";
			}
			let d = minFractionDigits - (n.length - i - 1);
			while (d-- > 0) n += "0";
		}
		return n;
	}
	var floatNaN$1 = {
		identify: (value) => typeof value === "number",
		default: true,
		tag: "tag:yaml.org,2002:float",
		test: /^(?:[-+]?\.(?:inf|Inf|INF)|\.nan|\.NaN|\.NAN)$/,
		resolve: (str) => str.slice(-3).toLowerCase() === "nan" ? NaN : str[0] === "-" ? Number.NEGATIVE_INFINITY : Number.POSITIVE_INFINITY,
		stringify: stringifyNumber
	};
	var floatExp$1 = {
		identify: (value) => typeof value === "number",
		default: true,
		tag: "tag:yaml.org,2002:float",
		format: "EXP",
		test: /^[-+]?(?:\.[0-9]+|[0-9]+(?:\.[0-9]*)?)[eE][-+]?[0-9]+$/,
		resolve: (str) => parseFloat(str),
		stringify(node) {
			const num = Number(node.value);
			return isFinite(num) ? num.toExponential() : stringifyNumber(node);
		}
	};
	var float$1 = {
		identify: (value) => typeof value === "number",
		default: true,
		tag: "tag:yaml.org,2002:float",
		test: /^[-+]?(?:\.[0-9]+|[0-9]+\.[0-9]*)$/,
		resolve(str) {
			const node = new Scalar(parseFloat(str));
			const dot = str.indexOf(".");
			if (dot !== -1 && str[str.length - 1] === "0") node.minFractionDigits = str.length - dot - 1;
			return node;
		},
		stringify: stringifyNumber
	};
	var intIdentify$2 = (value) => typeof value === "bigint" || Number.isInteger(value);
	var intResolve$1 = (str, offset, radix, { intAsBigInt }) => intAsBigInt ? BigInt(str) : parseInt(str.substring(offset), radix);
	function intStringify$1(node, radix, prefix) {
		const { value } = node;
		if (intIdentify$2(value) && value >= 0) return prefix + value.toString(radix);
		return stringifyNumber(node);
	}
	var intOct$1 = {
		identify: (value) => intIdentify$2(value) && value >= 0,
		default: true,
		tag: "tag:yaml.org,2002:int",
		format: "OCT",
		test: /^0o[0-7]+$/,
		resolve: (str, _onError, opt) => intResolve$1(str, 2, 8, opt),
		stringify: (node) => intStringify$1(node, 8, "0o")
	};
	var int$1 = {
		identify: intIdentify$2,
		default: true,
		tag: "tag:yaml.org,2002:int",
		test: /^[-+]?[0-9]+$/,
		resolve: (str, _onError, opt) => intResolve$1(str, 0, 10, opt),
		stringify: stringifyNumber
	};
	var intHex$1 = {
		identify: (value) => intIdentify$2(value) && value >= 0,
		default: true,
		tag: "tag:yaml.org,2002:int",
		format: "HEX",
		test: /^0x[0-9a-fA-F]+$/,
		resolve: (str, _onError, opt) => intResolve$1(str, 2, 16, opt),
		stringify: (node) => intStringify$1(node, 16, "0x")
	};
	var schema$2 = [
		map,
		seq,
		string,
		nullTag,
		boolTag,
		intOct$1,
		int$1,
		intHex$1,
		floatNaN$1,
		floatExp$1,
		float$1
	];
	function intIdentify$1(value) {
		return typeof value === "bigint" || Number.isInteger(value);
	}
	var stringifyJSON = ({ value }) => JSON.stringify(value);
	var jsonScalars = [
		{
			identify: (value) => typeof value === "string",
			default: true,
			tag: "tag:yaml.org,2002:str",
			resolve: (str) => str,
			stringify: stringifyJSON
		},
		{
			identify: (value) => value == null,
			createNode: () => new Scalar(null),
			default: true,
			tag: "tag:yaml.org,2002:null",
			test: /^null$/,
			resolve: () => null,
			stringify: stringifyJSON
		},
		{
			identify: (value) => typeof value === "boolean",
			default: true,
			tag: "tag:yaml.org,2002:bool",
			test: /^true$|^false$/,
			resolve: (str) => str === "true",
			stringify: stringifyJSON
		},
		{
			identify: intIdentify$1,
			default: true,
			tag: "tag:yaml.org,2002:int",
			test: /^-?(?:0|[1-9][0-9]*)$/,
			resolve: (str, _onError, { intAsBigInt }) => intAsBigInt ? BigInt(str) : parseInt(str, 10),
			stringify: ({ value }) => intIdentify$1(value) ? value.toString() : JSON.stringify(value)
		},
		{
			identify: (value) => typeof value === "number",
			default: true,
			tag: "tag:yaml.org,2002:float",
			test: /^-?(?:0|[1-9][0-9]*)(?:\.[0-9]*)?(?:[eE][-+]?[0-9]+)?$/,
			resolve: (str) => parseFloat(str),
			stringify: stringifyJSON
		}
	];
	var schema$1 = [map, seq].concat(jsonScalars, {
		default: true,
		tag: "",
		test: /^/,
		resolve(str, onError) {
			onError(`Unresolved plain scalar ${JSON.stringify(str)}`);
			return str;
		}
	});
	var binary = {
		identify: (value) => value instanceof Uint8Array,
		default: false,
		tag: "tag:yaml.org,2002:binary",
		/**
		* Returns a Buffer in node and an Uint8Array in browsers
		*
		* To use the resulting buffer as an image, you'll want to do something like:
		*
		*   const blob = new Blob([buffer], { type: 'image/jpeg' })
		*   document.querySelector('#photo').src = URL.createObjectURL(blob)
		*/
		resolve(src, onError) {
			if (typeof atob === "function") {
				const str = atob(src.replace(/[\n\r]/g, ""));
				const buffer = new Uint8Array(str.length);
				for (let i = 0; i < str.length; ++i) buffer[i] = str.charCodeAt(i);
				return buffer;
			} else {
				onError("This environment does not support reading binary tags; either Buffer or atob is required");
				return src;
			}
		},
		stringify({ comment, type, value }, ctx, onComment, onChompKeep) {
			if (!value) return "";
			const buf = value;
			let str;
			if (typeof btoa === "function") {
				let s = "";
				for (let i = 0; i < buf.length; ++i) s += String.fromCharCode(buf[i]);
				str = btoa(s);
			} else throw new Error("This environment does not support writing binary tags; either Buffer or btoa is required");
			type ?? (type = Scalar.BLOCK_LITERAL);
			if (type !== Scalar.QUOTE_DOUBLE) {
				const lineWidth = Math.max(ctx.options.lineWidth - ctx.indent.length, ctx.options.minContentWidth);
				const n = Math.ceil(str.length / lineWidth);
				const lines = new Array(n);
				for (let i = 0, o = 0; i < n; ++i, o += lineWidth) lines[i] = str.substr(o, lineWidth);
				str = lines.join(type === Scalar.BLOCK_LITERAL ? "\n" : " ");
			}
			return stringifyString({
				comment,
				type,
				value: str
			}, ctx, onComment, onChompKeep);
		}
	};
	function resolvePairs(seq, onError) {
		if (isSeq(seq)) for (let i = 0; i < seq.items.length; ++i) {
			let item = seq.items[i];
			if (isPair(item)) continue;
			else if (isMap(item)) {
				if (item.items.length > 1) onError("Each pair must have its own sequence indicator");
				const pair = item.items[0] || new Pair(new Scalar(null));
				if (item.commentBefore) pair.key.commentBefore = pair.key.commentBefore ? `${item.commentBefore}\n${pair.key.commentBefore}` : item.commentBefore;
				if (item.comment) {
					const cn = pair.value ?? pair.key;
					cn.comment = cn.comment ? `${item.comment}\n${cn.comment}` : item.comment;
				}
				item = pair;
			}
			seq.items[i] = isPair(item) ? item : new Pair(item);
		}
		else onError("Expected a sequence for this tag");
		return seq;
	}
	function createPairs(schema, iterable, ctx) {
		const { replacer } = ctx;
		const pairs = new YAMLSeq(schema);
		pairs.tag = "tag:yaml.org,2002:pairs";
		let i = 0;
		if (iterable && Symbol.iterator in Object(iterable)) for (let it of iterable) {
			if (typeof replacer === "function") it = replacer.call(iterable, String(i++), it);
			let key, value;
			if (Array.isArray(it)) {
				if (it.length === 2) {
					key = it[0];
					value = it[1];
				} else throw new TypeError(`Expected [key, value] tuple: ${it}`);
			} else if (it && it instanceof Object) {
				const keys = Object.keys(it);
				if (keys.length === 1) {
					key = keys[0];
					value = it[key];
				} else throw new TypeError(`Expected tuple with one key, not ${keys.length} keys`);
			} else key = it;
			pairs.items.push(createPair(key, value, ctx));
		}
		return pairs;
	}
	var pairs = {
		collection: "seq",
		default: false,
		tag: "tag:yaml.org,2002:pairs",
		resolve: resolvePairs,
		createNode: createPairs
	};
	var YAMLOMap = class YAMLOMap extends YAMLSeq {
		constructor() {
			super();
			this.add = YAMLMap.prototype.add.bind(this);
			this.delete = YAMLMap.prototype.delete.bind(this);
			this.get = YAMLMap.prototype.get.bind(this);
			this.has = YAMLMap.prototype.has.bind(this);
			this.set = YAMLMap.prototype.set.bind(this);
			this.tag = YAMLOMap.tag;
		}
		/**
		* If `ctx` is given, the return type is actually `Map<unknown, unknown>`,
		* but TypeScript won't allow widening the signature of a child method.
		*/
		toJSON(_, ctx) {
			if (!ctx) return super.toJSON(_);
			const map = /* @__PURE__ */ new Map();
			if (ctx?.onCreate) ctx.onCreate(map);
			for (const pair of this.items) {
				let key, value;
				if (isPair(pair)) {
					key = toJS(pair.key, "", ctx);
					value = toJS(pair.value, key, ctx);
				} else key = toJS(pair, "", ctx);
				if (map.has(key)) throw new Error("Ordered maps must not include duplicate keys");
				map.set(key, value);
			}
			return map;
		}
		static from(schema, iterable, ctx) {
			const pairs = createPairs(schema, iterable, ctx);
			const omap = new this();
			omap.items = pairs.items;
			return omap;
		}
	};
	YAMLOMap.tag = "tag:yaml.org,2002:omap";
	var omap = {
		collection: "seq",
		identify: (value) => value instanceof Map,
		nodeClass: YAMLOMap,
		default: false,
		tag: "tag:yaml.org,2002:omap",
		resolve(seq, onError) {
			const pairs = resolvePairs(seq, onError);
			const seenKeys = [];
			for (const { key } of pairs.items) if (isScalar(key)) {
				if (seenKeys.includes(key.value)) onError(`Ordered maps must not include duplicate keys: ${key.value}`);
				else seenKeys.push(key.value);
			}
			return Object.assign(new YAMLOMap(), pairs);
		},
		createNode: (schema, iterable, ctx) => YAMLOMap.from(schema, iterable, ctx)
	};
	function boolStringify({ value, source }, ctx) {
		if (source && (value ? trueTag : falseTag).test.test(source)) return source;
		return value ? ctx.options.trueStr : ctx.options.falseStr;
	}
	var trueTag = {
		identify: (value) => value === true,
		default: true,
		tag: "tag:yaml.org,2002:bool",
		test: /^(?:Y|y|[Yy]es|YES|[Tt]rue|TRUE|[Oo]n|ON)$/,
		resolve: () => new Scalar(true),
		stringify: boolStringify
	};
	var falseTag = {
		identify: (value) => value === false,
		default: true,
		tag: "tag:yaml.org,2002:bool",
		test: /^(?:N|n|[Nn]o|NO|[Ff]alse|FALSE|[Oo]ff|OFF)$/,
		resolve: () => new Scalar(false),
		stringify: boolStringify
	};
	var floatNaN = {
		identify: (value) => typeof value === "number",
		default: true,
		tag: "tag:yaml.org,2002:float",
		test: /^(?:[-+]?\.(?:inf|Inf|INF)|\.nan|\.NaN|\.NAN)$/,
		resolve: (str) => str.slice(-3).toLowerCase() === "nan" ? NaN : str[0] === "-" ? Number.NEGATIVE_INFINITY : Number.POSITIVE_INFINITY,
		stringify: stringifyNumber
	};
	var floatExp = {
		identify: (value) => typeof value === "number",
		default: true,
		tag: "tag:yaml.org,2002:float",
		format: "EXP",
		test: /^[-+]?(?:[0-9][0-9_]*)?(?:\.[0-9_]*)?[eE][-+]?[0-9]+$/,
		resolve: (str) => parseFloat(str.replace(/_/g, "")),
		stringify(node) {
			const num = Number(node.value);
			return isFinite(num) ? num.toExponential() : stringifyNumber(node);
		}
	};
	var float = {
		identify: (value) => typeof value === "number",
		default: true,
		tag: "tag:yaml.org,2002:float",
		test: /^[-+]?(?:[0-9][0-9_]*)?\.[0-9_]*$/,
		resolve(str) {
			const node = new Scalar(parseFloat(str.replace(/_/g, "")));
			const dot = str.indexOf(".");
			if (dot !== -1) {
				const f = str.substring(dot + 1).replace(/_/g, "");
				if (f[f.length - 1] === "0") node.minFractionDigits = f.length;
			}
			return node;
		},
		stringify: stringifyNumber
	};
	var intIdentify = (value) => typeof value === "bigint" || Number.isInteger(value);
	function intResolve(str, offset, radix, { intAsBigInt }) {
		const sign = str[0];
		if (sign === "-" || sign === "+") offset += 1;
		str = str.substring(offset).replace(/_/g, "");
		if (intAsBigInt) {
			switch (radix) {
				case 2:
					str = `0b${str}`;
					break;
				case 8:
					str = `0o${str}`;
					break;
				case 16: str = `0x${str}`;
			}
			const n = BigInt(str);
			return sign === "-" ? BigInt(-1) * n : n;
		}
		const n = parseInt(str, radix);
		return sign === "-" ? -1 * n : n;
	}
	function intStringify(node, radix, prefix) {
		const { value } = node;
		if (intIdentify(value)) {
			const str = value.toString(radix);
			return value < 0 ? "-" + prefix + str.substr(1) : prefix + str;
		}
		return stringifyNumber(node);
	}
	var intBin = {
		identify: intIdentify,
		default: true,
		tag: "tag:yaml.org,2002:int",
		format: "BIN",
		test: /^[-+]?0b[0-1_]+$/,
		resolve: (str, _onError, opt) => intResolve(str, 2, 2, opt),
		stringify: (node) => intStringify(node, 2, "0b")
	};
	var intOct = {
		identify: intIdentify,
		default: true,
		tag: "tag:yaml.org,2002:int",
		format: "OCT",
		test: /^[-+]?0[0-7_]+$/,
		resolve: (str, _onError, opt) => intResolve(str, 1, 8, opt),
		stringify: (node) => intStringify(node, 8, "0")
	};
	var int = {
		identify: intIdentify,
		default: true,
		tag: "tag:yaml.org,2002:int",
		test: /^[-+]?[0-9][0-9_]*$/,
		resolve: (str, _onError, opt) => intResolve(str, 0, 10, opt),
		stringify: stringifyNumber
	};
	var intHex = {
		identify: intIdentify,
		default: true,
		tag: "tag:yaml.org,2002:int",
		format: "HEX",
		test: /^[-+]?0x[0-9a-fA-F_]+$/,
		resolve: (str, _onError, opt) => intResolve(str, 2, 16, opt),
		stringify: (node) => intStringify(node, 16, "0x")
	};
	var YAMLSet = class YAMLSet extends YAMLMap {
		constructor(schema) {
			super(schema);
			this.tag = YAMLSet.tag;
		}
		add(key) {
			let pair;
			if (isPair(key)) pair = key;
			else if (key && typeof key === "object" && "key" in key && "value" in key && key.value === null) pair = new Pair(key.key, null);
			else pair = new Pair(key, null);
			if (!findPair(this.items, pair.key)) this.items.push(pair);
		}
		/**
		* If `keepPair` is `true`, returns the Pair matching `key`.
		* Otherwise, returns the value of that Pair's key.
		*/
		get(key, keepPair) {
			const pair = findPair(this.items, key);
			return !keepPair && isPair(pair) ? isScalar(pair.key) ? pair.key.value : pair.key : pair;
		}
		set(key, value) {
			if (typeof value !== "boolean") throw new Error(`Expected boolean value for set(key, value) in a YAML set, not ${typeof value}`);
			const prev = findPair(this.items, key);
			if (prev && !value) this.items.splice(this.items.indexOf(prev), 1);
			else if (!prev && value) this.items.push(new Pair(key));
		}
		toJSON(_, ctx) {
			return super.toJSON(_, ctx, Set);
		}
		toString(ctx, onComment, onChompKeep) {
			if (!ctx) return JSON.stringify(this);
			if (this.hasAllNullValues(true)) return super.toString(Object.assign({}, ctx, { allNullValues: true }), onComment, onChompKeep);
			else throw new Error("Set items must all have null values");
		}
		static from(schema, iterable, ctx) {
			const { replacer } = ctx;
			const set = new this(schema);
			if (iterable && Symbol.iterator in Object(iterable)) for (let value of iterable) {
				if (typeof replacer === "function") value = replacer.call(iterable, value, value);
				set.items.push(createPair(value, null, ctx));
			}
			return set;
		}
	};
	YAMLSet.tag = "tag:yaml.org,2002:set";
	var set = {
		collection: "map",
		identify: (value) => value instanceof Set,
		nodeClass: YAMLSet,
		default: false,
		tag: "tag:yaml.org,2002:set",
		createNode: (schema, iterable, ctx) => YAMLSet.from(schema, iterable, ctx),
		resolve(map, onError) {
			if (isMap(map)) {
				if (map.hasAllNullValues(true)) return Object.assign(new YAMLSet(), map);
				else onError("Set items must all have null values");
			} else onError("Expected a mapping for this tag");
			return map;
		}
	};
	/** Internal types handle bigint as number, because TS can't figure it out. */
	function parseSexagesimal(str, asBigInt) {
		const sign = str[0];
		const parts = sign === "-" || sign === "+" ? str.substring(1) : str;
		const num = (n) => asBigInt ? BigInt(n) : Number(n);
		const res = parts.replace(/_/g, "").split(":").reduce((res, p) => res * num(60) + num(p), num(0));
		return sign === "-" ? num(-1) * res : res;
	}
	/**
	* hhhh:mm:ss.sss
	*
	* Internal types handle bigint as number, because TS can't figure it out.
	*/
	function stringifySexagesimal(node) {
		let { value } = node;
		let num = (n) => n;
		if (typeof value === "bigint") num = (n) => BigInt(n);
		else if (isNaN(value) || !isFinite(value)) return stringifyNumber(node);
		let sign = "";
		if (value < 0) {
			sign = "-";
			value *= num(-1);
		}
		const _60 = num(60);
		const parts = [value % _60];
		if (value < 60) parts.unshift(0);
		else {
			value = (value - parts[0]) / _60;
			parts.unshift(value % _60);
			if (value >= 60) {
				value = (value - parts[0]) / _60;
				parts.unshift(value);
			}
		}
		return sign + parts.map((n) => String(n).padStart(2, "0")).join(":").replace(/000000\d*$/, "");
	}
	var intTime = {
		identify: (value) => typeof value === "bigint" || Number.isInteger(value),
		default: true,
		tag: "tag:yaml.org,2002:int",
		format: "TIME",
		test: /^[-+]?[0-9][0-9_]*(?::[0-5]?[0-9])+$/,
		resolve: (str, _onError, { intAsBigInt }) => parseSexagesimal(str, intAsBigInt),
		stringify: stringifySexagesimal
	};
	var floatTime = {
		identify: (value) => typeof value === "number",
		default: true,
		tag: "tag:yaml.org,2002:float",
		format: "TIME",
		test: /^[-+]?[0-9][0-9_]*(?::[0-5]?[0-9])+\.[0-9_]*$/,
		resolve: (str) => parseSexagesimal(str, false),
		stringify: stringifySexagesimal
	};
	var timestamp = {
		identify: (value) => value instanceof Date,
		default: true,
		tag: "tag:yaml.org,2002:timestamp",
		test: RegExp("^([0-9]{4})-([0-9]{1,2})-([0-9]{1,2})(?:(?:t|T|[ \\t]+)([0-9]{1,2}):([0-9]{1,2}):([0-9]{1,2}(\\.[0-9]+)?)(?:[ \\t]*(Z|[-+][012]?[0-9](?::[0-9]{2})?))?)?$"),
		resolve(str) {
			const match = str.match(timestamp.test);
			if (!match) throw new Error("!!timestamp expects a date, starting with yyyy-mm-dd");
			const [, year, month, day, hour, minute, second] = match.map(Number);
			const millisec = match[7] ? Number((match[7] + "00").substr(1, 3)) : 0;
			let date = Date.UTC(year, month - 1, day, hour || 0, minute || 0, second || 0, millisec);
			const tz = match[8];
			if (tz && tz !== "Z") {
				let d = parseSexagesimal(tz, false);
				if (Math.abs(d) < 30) d *= 60;
				date -= 6e4 * d;
			}
			return new Date(date);
		},
		stringify: ({ value }) => value?.toISOString().replace(/(T00:00:00)?\.000Z$/, "") ?? ""
	};
	var schema = [
		map,
		seq,
		string,
		nullTag,
		trueTag,
		falseTag,
		intBin,
		intOct,
		int,
		intHex,
		floatNaN,
		floatExp,
		float,
		binary,
		merge,
		omap,
		pairs,
		set,
		intTime,
		floatTime,
		timestamp
	];
	var schemas = /* @__PURE__ */ new Map([
		["core", schema$2],
		["failsafe", [
			map,
			seq,
			string
		]],
		["json", schema$1],
		["yaml11", schema],
		["yaml-1.1", schema]
	]);
	var tagsByName = {
		binary,
		bool: boolTag,
		float: float$1,
		floatExp: floatExp$1,
		floatNaN: floatNaN$1,
		floatTime,
		int: int$1,
		intHex: intHex$1,
		intOct: intOct$1,
		intTime,
		map,
		merge,
		null: nullTag,
		omap,
		pairs,
		seq,
		set,
		timestamp
	};
	var coreKnownTags = {
		"tag:yaml.org,2002:binary": binary,
		"tag:yaml.org,2002:merge": merge,
		"tag:yaml.org,2002:omap": omap,
		"tag:yaml.org,2002:pairs": pairs,
		"tag:yaml.org,2002:set": set,
		"tag:yaml.org,2002:timestamp": timestamp
	};
	function getTags(customTags, schemaName, addMergeTag) {
		const schemaTags = schemas.get(schemaName);
		if (schemaTags && !customTags) return addMergeTag && !schemaTags.includes(merge) ? schemaTags.concat(merge) : schemaTags.slice();
		let tags = schemaTags;
		if (!tags) {
			if (Array.isArray(customTags)) tags = [];
			else {
				const keys = Array.from(schemas.keys()).filter((key) => key !== "yaml11").map((key) => JSON.stringify(key)).join(", ");
				throw new Error(`Unknown schema "${schemaName}"; use one of ${keys} or define customTags array`);
			}
		}
		if (Array.isArray(customTags)) for (const tag of customTags) tags = tags.concat(tag);
		else if (typeof customTags === "function") tags = customTags(tags.slice());
		if (addMergeTag) tags = tags.concat(merge);
		return tags.reduce((tags, tag) => {
			const tagObj = typeof tag === "string" ? tagsByName[tag] : tag;
			if (!tagObj) {
				const tagName = JSON.stringify(tag);
				const keys = Object.keys(tagsByName).map((key) => JSON.stringify(key)).join(", ");
				throw new Error(`Unknown custom tag ${tagName}; use one of ${keys}`);
			}
			if (!tags.includes(tagObj)) tags.push(tagObj);
			return tags;
		}, []);
	}
	var sortMapEntriesByKey = (a, b) => a.key < b.key ? -1 : a.key > b.key ? 1 : 0;
	var Schema = class Schema {
		constructor({ compat, customTags, merge, resolveKnownTags, schema, sortMapEntries, toStringDefaults }) {
			this.compat = Array.isArray(compat) ? getTags(compat, "compat") : compat ? getTags(null, compat) : null;
			this.name = typeof schema === "string" && schema || "core";
			this.knownTags = resolveKnownTags ? coreKnownTags : {};
			this.tags = getTags(customTags, this.name, merge);
			this.toStringOptions = toStringDefaults ?? null;
			Object.defineProperty(this, MAP, { value: map });
			Object.defineProperty(this, SCALAR$1, { value: string });
			Object.defineProperty(this, SEQ, { value: seq });
			this.sortMapEntries = typeof sortMapEntries === "function" ? sortMapEntries : sortMapEntries === true ? sortMapEntriesByKey : null;
		}
		clone() {
			const copy = Object.create(Schema.prototype, Object.getOwnPropertyDescriptors(this));
			copy.tags = this.tags.slice();
			return copy;
		}
	};
	function stringifyDocument(doc, options) {
		const lines = [];
		let hasDirectives = options.directives === true;
		if (options.directives !== false && doc.directives) {
			const dir = doc.directives.toString(doc);
			if (dir) {
				lines.push(dir);
				hasDirectives = true;
			} else if (doc.directives.docStart) hasDirectives = true;
		}
		if (hasDirectives) lines.push("---");
		const ctx = createStringifyContext(doc, options);
		const { commentString } = ctx.options;
		if (doc.commentBefore) {
			if (lines.length !== 1) lines.unshift("");
			const cs = commentString(doc.commentBefore);
			lines.unshift(indentComment(cs, ""));
		}
		let chompKeep = false;
		let contentComment = null;
		if (doc.contents) {
			if (isNode(doc.contents)) {
				if (doc.contents.spaceBefore && hasDirectives) lines.push("");
				if (doc.contents.commentBefore) {
					const cs = commentString(doc.contents.commentBefore);
					lines.push(indentComment(cs, ""));
				}
				ctx.forceBlockIndent = !!doc.comment;
				contentComment = doc.contents.comment;
			}
			const onChompKeep = contentComment ? void 0 : () => chompKeep = true;
			let body = stringify(doc.contents, ctx, () => contentComment = null, onChompKeep);
			if (contentComment) body += lineComment(body, "", commentString(contentComment));
			if ((body[0] === "|" || body[0] === ">") && lines[lines.length - 1] === "---") lines[lines.length - 1] = `--- ${body}`;
			else lines.push(body);
		} else lines.push(stringify(doc.contents, ctx));
		if (doc.directives?.docEnd) {
			if (doc.comment) {
				const cs = commentString(doc.comment);
				if (cs.includes("\n")) {
					lines.push("...");
					lines.push(indentComment(cs, ""));
				} else lines.push(`... ${cs}`);
			} else lines.push("...");
		} else {
			let dc = doc.comment;
			if (dc && chompKeep) dc = dc.replace(/^\n+/, "");
			if (dc) {
				if ((!chompKeep || contentComment) && lines[lines.length - 1] !== "") lines.push("");
				lines.push(indentComment(commentString(dc), ""));
			}
		}
		return lines.join("\n") + "\n";
	}
	var Document = class Document {
		constructor(value, replacer, options) {
			/** A comment before this Document */
			this.commentBefore = null;
			/** A comment immediately after this Document */
			this.comment = null;
			/** Errors encountered during parsing. */
			this.errors = [];
			/** Warnings encountered during parsing. */
			this.warnings = [];
			Object.defineProperty(this, NODE_TYPE, { value: DOC });
			let _replacer = null;
			if (typeof replacer === "function" || Array.isArray(replacer)) _replacer = replacer;
			else if (options === void 0 && replacer) {
				options = replacer;
				replacer = void 0;
			}
			const opt = Object.assign({
				intAsBigInt: false,
				keepSourceTokens: false,
				logLevel: "warn",
				prettyErrors: true,
				strict: true,
				stringKeys: false,
				uniqueKeys: true,
				version: "1.2"
			}, options);
			this.options = opt;
			let { version } = opt;
			if (options?._directives) {
				this.directives = options._directives.atDocument();
				if (this.directives.yaml.explicit) version = this.directives.yaml.version;
			} else this.directives = new Directives({ version });
			this.setSchema(version, options);
			this.contents = value === void 0 ? null : this.createNode(value, _replacer, options);
		}
		/**
		* Create a deep copy of this Document and its contents.
		*
		* Custom Node values that inherit from `Object` still refer to their original instances.
		*/
		clone() {
			const copy = Object.create(Document.prototype, { [NODE_TYPE]: { value: DOC } });
			copy.commentBefore = this.commentBefore;
			copy.comment = this.comment;
			copy.errors = this.errors.slice();
			copy.warnings = this.warnings.slice();
			copy.options = Object.assign({}, this.options);
			if (this.directives) copy.directives = this.directives.clone();
			copy.schema = this.schema.clone();
			copy.contents = isNode(this.contents) ? this.contents.clone(copy.schema) : this.contents;
			if (this.range) copy.range = this.range.slice();
			return copy;
		}
		/** Adds a value to the document. */
		add(value) {
			if (assertCollection(this.contents)) this.contents.add(value);
		}
		/** Adds a value to the document. */
		addIn(path, value) {
			if (assertCollection(this.contents)) this.contents.addIn(path, value);
		}
		/**
		* Create a new `Alias` node, ensuring that the target `node` has the required anchor.
		*
		* If `node` already has an anchor, `name` is ignored.
		* Otherwise, the `node.anchor` value will be set to `name`,
		* or if an anchor with that name is already present in the document,
		* `name` will be used as a prefix for a new unique anchor.
		* If `name` is undefined, the generated anchor will use 'a' as a prefix.
		*/
		createAlias(node, name) {
			if (!node.anchor) {
				const prev = anchorNames(this);
				node.anchor = !name || prev.has(name) ? findNewAnchor(name || "a", prev) : name;
			}
			return new Alias(node.anchor);
		}
		createNode(value, replacer, options) {
			let _replacer = void 0;
			if (typeof replacer === "function") {
				value = replacer.call({ "": value }, "", value);
				_replacer = replacer;
			} else if (Array.isArray(replacer)) {
				const keyToStr = (v) => typeof v === "number" || v instanceof String || v instanceof Number;
				const asStr = replacer.filter(keyToStr).map(String);
				if (asStr.length > 0) replacer = replacer.concat(asStr);
				_replacer = replacer;
			} else if (options === void 0 && replacer) {
				options = replacer;
				replacer = void 0;
			}
			const { aliasDuplicateObjects, anchorPrefix, flow, keepUndefined, onTagObj, tag } = options ?? {};
			const { onAnchor, setAnchors, sourceObjects } = createNodeAnchors(this, anchorPrefix || "a");
			const ctx = {
				aliasDuplicateObjects: aliasDuplicateObjects ?? true,
				keepUndefined: keepUndefined ?? false,
				onAnchor,
				onTagObj,
				replacer: _replacer,
				schema: this.schema,
				sourceObjects
			};
			const node = createNode(value, tag, ctx);
			if (flow && isCollection(node)) node.flow = true;
			setAnchors();
			return node;
		}
		/**
		* Convert a key and a value into a `Pair` using the current schema,
		* recursively wrapping all values as `Scalar` or `Collection` nodes.
		*/
		createPair(key, value, options = {}) {
			return new Pair(this.createNode(key, null, options), this.createNode(value, null, options));
		}
		/**
		* Removes a value from the document.
		* @returns `true` if the item was found and removed.
		*/
		delete(key) {
			return assertCollection(this.contents) ? this.contents.delete(key) : false;
		}
		/**
		* Removes a value from the document.
		* @returns `true` if the item was found and removed.
		*/
		deleteIn(path) {
			if (isEmptyPath(path)) {
				if (this.contents == null) return false;
				this.contents = null;
				return true;
			}
			return assertCollection(this.contents) ? this.contents.deleteIn(path) : false;
		}
		/**
		* Returns item at `key`, or `undefined` if not found. By default unwraps
		* scalar values from their surrounding node; to disable set `keepScalar` to
		* `true` (collections are always returned intact).
		*/
		get(key, keepScalar) {
			return isCollection(this.contents) ? this.contents.get(key, keepScalar) : void 0;
		}
		/**
		* Returns item at `path`, or `undefined` if not found. By default unwraps
		* scalar values from their surrounding node; to disable set `keepScalar` to
		* `true` (collections are always returned intact).
		*/
		getIn(path, keepScalar) {
			if (isEmptyPath(path)) return !keepScalar && isScalar(this.contents) ? this.contents.value : this.contents;
			return isCollection(this.contents) ? this.contents.getIn(path, keepScalar) : void 0;
		}
		/**
		* Checks if the document includes a value with the key `key`.
		*/
		has(key) {
			return isCollection(this.contents) ? this.contents.has(key) : false;
		}
		/**
		* Checks if the document includes a value at `path`.
		*/
		hasIn(path) {
			if (isEmptyPath(path)) return this.contents !== void 0;
			return isCollection(this.contents) ? this.contents.hasIn(path) : false;
		}
		/**
		* Sets a value in this document. For `!!set`, `value` needs to be a
		* boolean to add/remove the item from the set.
		*/
		set(key, value) {
			if (this.contents == null) this.contents = collectionFromPath(this.schema, [key], value);
			else if (assertCollection(this.contents)) this.contents.set(key, value);
		}
		/**
		* Sets a value in this document. For `!!set`, `value` needs to be a
		* boolean to add/remove the item from the set.
		*/
		setIn(path, value) {
			if (isEmptyPath(path)) this.contents = value;
			else if (this.contents == null) this.contents = collectionFromPath(this.schema, Array.from(path), value);
			else if (assertCollection(this.contents)) this.contents.setIn(path, value);
		}
		/**
		* Change the YAML version and schema used by the document.
		* A `null` version disables support for directives, explicit tags, anchors, and aliases.
		* It also requires the `schema` option to be given as a `Schema` instance value.
		*
		* Overrides all previously set schema options.
		*/
		setSchema(version, options = {}) {
			if (typeof version === "number") version = String(version);
			let opt;
			switch (version) {
				case "1.1":
					if (this.directives) this.directives.yaml.version = "1.1";
					else this.directives = new Directives({ version: "1.1" });
					opt = {
						resolveKnownTags: false,
						schema: "yaml-1.1"
					};
					break;
				case "1.2":
				case "next":
					if (this.directives) this.directives.yaml.version = version;
					else this.directives = new Directives({ version });
					opt = {
						resolveKnownTags: true,
						schema: "core"
					};
					break;
				case null:
					if (this.directives) delete this.directives;
					opt = null;
					break;
				default: {
					const sv = JSON.stringify(version);
					throw new Error(`Expected '1.1', '1.2' or null as first argument, but found: ${sv}`);
				}
			}
			if (options.schema instanceof Object) this.schema = options.schema;
			else if (opt) this.schema = new Schema(Object.assign(opt, options));
			else throw new Error(`With a null YAML version, the { schema: Schema } option is required`);
		}
		toJS({ json, jsonArg, mapAsMap, maxAliasCount, onAnchor, reviver } = {}) {
			const ctx = {
				anchors: /* @__PURE__ */ new Map(),
				doc: this,
				keep: !json,
				mapAsMap: mapAsMap === true,
				mapKeyWarned: false,
				maxAliasCount: typeof maxAliasCount === "number" ? maxAliasCount : 100
			};
			const res = toJS(this.contents, jsonArg ?? "", ctx);
			if (typeof onAnchor === "function") for (const { count, res } of ctx.anchors.values()) onAnchor(res, count);
			return typeof reviver === "function" ? applyReviver(reviver, { "": res }, "", res) : res;
		}
		/**
		* A JSON representation of the document `contents`.
		*
		* @param jsonArg Used by `JSON.stringify` to indicate the array index or
		*   property name.
		*/
		toJSON(jsonArg, onAnchor) {
			return this.toJS({
				json: true,
				jsonArg,
				mapAsMap: false,
				onAnchor
			});
		}
		/** A YAML representation of the document. */
		toString(options = {}) {
			if (this.errors.length > 0) throw new Error("Document with errors cannot be stringified");
			if ("indent" in options && (!Number.isInteger(options.indent) || Number(options.indent) <= 0)) {
				const s = JSON.stringify(options.indent);
				throw new Error(`"indent" option must be a positive integer, not ${s}`);
			}
			return stringifyDocument(this, options);
		}
	};
	function assertCollection(contents) {
		if (isCollection(contents)) return true;
		throw new Error("Expected a YAML collection as document contents");
	}
	var YAMLError = class extends Error {
		constructor(name, pos, code, message) {
			super();
			this.name = name;
			this.code = code;
			this.message = message;
			this.pos = pos;
		}
	};
	var YAMLParseError = class extends YAMLError {
		constructor(pos, code, message) {
			super("YAMLParseError", pos, code, message);
		}
	};
	var YAMLWarning = class extends YAMLError {
		constructor(pos, code, message) {
			super("YAMLWarning", pos, code, message);
		}
	};
	var prettifyError = (src, lc) => (error) => {
		if (error.pos[0] === -1) return;
		error.linePos = error.pos.map((pos) => lc.linePos(pos));
		const { line, col } = error.linePos[0];
		error.message += ` at line ${line}, column ${col}`;
		let ci = col - 1;
		let lineStr = src.substring(lc.lineStarts[line - 1], lc.lineStarts[line]).replace(/[\n\r]+$/, "");
		if (ci >= 60 && lineStr.length > 80) {
			const trimStart = Math.min(ci - 39, lineStr.length - 79);
			lineStr = "…" + lineStr.substring(trimStart);
			ci -= trimStart - 1;
		}
		if (lineStr.length > 80) lineStr = lineStr.substring(0, 79) + "…";
		if (line > 1 && /^ *$/.test(lineStr.substring(0, ci))) {
			let prev = src.substring(lc.lineStarts[line - 2], lc.lineStarts[line - 1]);
			if (prev.length > 80) prev = prev.substring(0, 79) + "…\n";
			lineStr = prev + lineStr;
		}
		if (/[^ ]/.test(lineStr)) {
			let count = 1;
			const end = error.linePos[1];
			if (end?.line === line && end.col > col) count = Math.max(1, Math.min(end.col - col, 80 - ci));
			const pointer = " ".repeat(ci) + "^".repeat(count);
			error.message += `:\n\n${lineStr}\n${pointer}\n`;
		}
	};
	function resolveProps(tokens, { flow, indicator, next, offset, onError, parentIndent, startOnNewline }) {
		let spaceBefore = false;
		let atNewline = startOnNewline;
		let hasSpace = startOnNewline;
		let comment = "";
		let commentSep = "";
		let hasNewline = false;
		let reqSpace = false;
		let tab = null;
		let anchor = null;
		let tag = null;
		let newlineAfterProp = null;
		let comma = null;
		let found = null;
		let start = null;
		for (const token of tokens) {
			if (reqSpace) {
				if (token.type !== "space" && token.type !== "newline" && token.type !== "comma") onError(token.offset, "MISSING_CHAR", "Tags and anchors must be separated from the next token by white space");
				reqSpace = false;
			}
			if (tab) {
				if (atNewline && token.type !== "comment" && token.type !== "newline") onError(tab, "TAB_AS_INDENT", "Tabs are not allowed as indentation");
				tab = null;
			}
			switch (token.type) {
				case "space":
					if (!flow && (indicator !== "doc-start" || next?.type !== "flow-collection") && token.source.includes("	")) tab = token;
					hasSpace = true;
					break;
				case "comment": {
					if (!hasSpace) onError(token, "MISSING_CHAR", "Comments must be separated from other tokens by white space characters");
					const cb = token.source.substring(1) || " ";
					if (!comment) comment = cb;
					else comment += commentSep + cb;
					commentSep = "";
					atNewline = false;
					break;
				}
				case "newline":
					if (atNewline) {
						if (comment) comment += token.source;
						else if (!found || indicator !== "seq-item-ind") spaceBefore = true;
					} else commentSep += token.source;
					atNewline = true;
					hasNewline = true;
					if (anchor || tag) newlineAfterProp = token;
					hasSpace = true;
					break;
				case "anchor":
					if (anchor) onError(token, "MULTIPLE_ANCHORS", "A node can have at most one anchor");
					if (token.source.endsWith(":")) onError(token.offset + token.source.length - 1, "BAD_ALIAS", "Anchor ending in : is ambiguous", true);
					anchor = token;
					start ?? (start = token.offset);
					atNewline = false;
					hasSpace = false;
					reqSpace = true;
					break;
				case "tag":
					if (tag) onError(token, "MULTIPLE_TAGS", "A node can have at most one tag");
					tag = token;
					start ?? (start = token.offset);
					atNewline = false;
					hasSpace = false;
					reqSpace = true;
					break;
				case indicator:
					if (anchor || tag) onError(token, "BAD_PROP_ORDER", `Anchors and tags must be after the ${token.source} indicator`);
					if (found) onError(token, "UNEXPECTED_TOKEN", `Unexpected ${token.source} in ${flow ?? "collection"}`);
					found = token;
					atNewline = indicator === "seq-item-ind" || indicator === "explicit-key-ind";
					hasSpace = false;
					break;
				case "comma": if (flow) {
					if (comma) onError(token, "UNEXPECTED_TOKEN", `Unexpected , in ${flow}`);
					comma = token;
					atNewline = false;
					hasSpace = false;
					break;
				}
				default:
					onError(token, "UNEXPECTED_TOKEN", `Unexpected ${token.type} token`);
					atNewline = false;
					hasSpace = false;
			}
		}
		const last = tokens[tokens.length - 1];
		const end = last ? last.offset + last.source.length : offset;
		if (reqSpace && next && next.type !== "space" && next.type !== "newline" && next.type !== "comma" && (next.type !== "scalar" || next.source !== "")) onError(next.offset, "MISSING_CHAR", "Tags and anchors must be separated from the next token by white space");
		if (tab && (atNewline && tab.indent <= parentIndent || next?.type === "block-map" || next?.type === "block-seq")) onError(tab, "TAB_AS_INDENT", "Tabs are not allowed as indentation");
		return {
			comma,
			found,
			spaceBefore,
			comment,
			hasNewline,
			anchor,
			tag,
			newlineAfterProp,
			end,
			start: start ?? end
		};
	}
	function containsNewline(key) {
		if (!key) return null;
		switch (key.type) {
			case "alias":
			case "scalar":
			case "double-quoted-scalar":
			case "single-quoted-scalar":
				if (key.source.includes("\n")) return true;
				if (key.end) {
					for (const st of key.end) if (st.type === "newline") return true;
				}
				return false;
			case "flow-collection":
				for (const it of key.items) {
					for (const st of it.start) if (st.type === "newline") return true;
					if (it.sep) {
						for (const st of it.sep) if (st.type === "newline") return true;
					}
					if (containsNewline(it.key) || containsNewline(it.value)) return true;
				}
				return false;
			default: return true;
		}
	}
	function flowIndentCheck(indent, fc, onError) {
		if (fc?.type === "flow-collection") {
			const end = fc.end[0];
			if (end.indent === indent && (end.source === "]" || end.source === "}") && containsNewline(fc)) onError(end, "BAD_INDENT", "Flow end indicator should be more indented than parent", true);
		}
	}
	function mapIncludes(ctx, items, search) {
		const { uniqueKeys } = ctx.options;
		if (uniqueKeys === false) return false;
		const isEqual = typeof uniqueKeys === "function" ? uniqueKeys : (a, b) => a === b || isScalar(a) && isScalar(b) && a.value === b.value;
		return items.some((pair) => isEqual(pair.key, search));
	}
	var startColMsg = "All mapping items must start at the same column";
	function resolveBlockMap({ composeNode, composeEmptyNode }, ctx, bm, onError, tag) {
		const map = new ((tag?.nodeClass) ?? YAMLMap)(ctx.schema);
		if (ctx.atRoot) ctx.atRoot = false;
		let offset = bm.offset;
		let commentEnd = null;
		for (const collItem of bm.items) {
			const { start, key, sep, value } = collItem;
			const keyProps = resolveProps(start, {
				indicator: "explicit-key-ind",
				next: key ?? sep?.[0],
				offset,
				onError,
				parentIndent: bm.indent,
				startOnNewline: true
			});
			const implicitKey = !keyProps.found;
			if (implicitKey) {
				if (key) {
					if (key.type === "block-seq") onError(offset, "BLOCK_AS_IMPLICIT_KEY", "A block sequence may not be used as an implicit map key");
					else if ("indent" in key && key.indent !== bm.indent) onError(offset, "BAD_INDENT", startColMsg);
				}
				if (!keyProps.anchor && !keyProps.tag && !sep) {
					commentEnd = keyProps.end;
					if (keyProps.comment) {
						if (map.comment) map.comment += "\n" + keyProps.comment;
						else map.comment = keyProps.comment;
					}
					continue;
				}
				if (keyProps.newlineAfterProp || containsNewline(key)) onError(key ?? start[start.length - 1], "MULTILINE_IMPLICIT_KEY", "Implicit keys need to be on a single line");
			} else if (keyProps.found?.indent !== bm.indent) onError(offset, "BAD_INDENT", startColMsg);
			ctx.atKey = true;
			const keyStart = keyProps.end;
			const keyNode = key ? composeNode(ctx, key, keyProps, onError) : composeEmptyNode(ctx, keyStart, start, null, keyProps, onError);
			if (ctx.schema.compat) flowIndentCheck(bm.indent, key, onError);
			ctx.atKey = false;
			if (mapIncludes(ctx, map.items, keyNode)) onError(keyStart, "DUPLICATE_KEY", "Map keys must be unique");
			const valueProps = resolveProps(sep ?? [], {
				indicator: "map-value-ind",
				next: value,
				offset: keyNode.range[2],
				onError,
				parentIndent: bm.indent,
				startOnNewline: !key || key.type === "block-scalar"
			});
			offset = valueProps.end;
			if (valueProps.found) {
				if (implicitKey) {
					if (value?.type === "block-map" && !valueProps.hasNewline) onError(offset, "BLOCK_AS_IMPLICIT_KEY", "Nested mappings are not allowed in compact mappings");
					if (ctx.options.strict && keyProps.start < valueProps.found.offset - 1024) onError(keyNode.range, "KEY_OVER_1024_CHARS", "The : indicator must be at most 1024 chars after the start of an implicit block mapping key");
				}
				const valueNode = value ? composeNode(ctx, value, valueProps, onError) : composeEmptyNode(ctx, offset, sep, null, valueProps, onError);
				if (ctx.schema.compat) flowIndentCheck(bm.indent, value, onError);
				offset = valueNode.range[2];
				const pair = new Pair(keyNode, valueNode);
				if (ctx.options.keepSourceTokens) pair.srcToken = collItem;
				map.items.push(pair);
			} else {
				if (implicitKey) onError(keyNode.range, "MISSING_CHAR", "Implicit map keys need to be followed by map values");
				if (valueProps.comment) {
					if (keyNode.comment) keyNode.comment += "\n" + valueProps.comment;
					else keyNode.comment = valueProps.comment;
				}
				const pair = new Pair(keyNode);
				if (ctx.options.keepSourceTokens) pair.srcToken = collItem;
				map.items.push(pair);
			}
		}
		if (commentEnd && commentEnd < offset) onError(commentEnd, "IMPOSSIBLE", "Map comment with trailing content");
		map.range = [
			bm.offset,
			offset,
			commentEnd ?? offset
		];
		return map;
	}
	function resolveBlockSeq({ composeNode, composeEmptyNode }, ctx, bs, onError, tag) {
		const seq = new ((tag?.nodeClass) ?? YAMLSeq)(ctx.schema);
		if (ctx.atRoot) ctx.atRoot = false;
		if (ctx.atKey) ctx.atKey = false;
		let offset = bs.offset;
		let commentEnd = null;
		for (const { start, value } of bs.items) {
			const props = resolveProps(start, {
				indicator: "seq-item-ind",
				next: value,
				offset,
				onError,
				parentIndent: bs.indent,
				startOnNewline: true
			});
			if (!props.found) {
				if (props.anchor || props.tag || value) {
					if (value?.type === "block-seq") onError(props.end, "BAD_INDENT", "All sequence items must start at the same column");
					else onError(offset, "MISSING_CHAR", "Sequence item without - indicator");
				} else {
					commentEnd = props.end;
					if (props.comment) seq.comment = props.comment;
					continue;
				}
			}
			const node = value ? composeNode(ctx, value, props, onError) : composeEmptyNode(ctx, props.end, start, null, props, onError);
			if (ctx.schema.compat) flowIndentCheck(bs.indent, value, onError);
			offset = node.range[2];
			seq.items.push(node);
		}
		seq.range = [
			bs.offset,
			offset,
			commentEnd ?? offset
		];
		return seq;
	}
	function resolveEnd(end, offset, reqSpace, onError) {
		let comment = "";
		if (end) {
			let hasSpace = false;
			let sep = "";
			for (const token of end) {
				const { source, type } = token;
				switch (type) {
					case "space":
						hasSpace = true;
						break;
					case "comment": {
						if (reqSpace && !hasSpace) onError(token, "MISSING_CHAR", "Comments must be separated from other tokens by white space characters");
						const cb = source.substring(1) || " ";
						if (!comment) comment = cb;
						else comment += sep + cb;
						sep = "";
						break;
					}
					case "newline":
						if (comment) sep += source;
						hasSpace = true;
						break;
					default: onError(token, "UNEXPECTED_TOKEN", `Unexpected ${type} at node end`);
				}
				offset += source.length;
			}
		}
		return {
			comment,
			offset
		};
	}
	var blockMsg = "Block collections are not allowed within flow collections";
	var isBlock = (token) => token && (token.type === "block-map" || token.type === "block-seq");
	function resolveFlowCollection({ composeNode, composeEmptyNode }, ctx, fc, onError, tag) {
		const isMap = fc.start.source === "{";
		const fcName = isMap ? "flow map" : "flow sequence";
		const coll = new ((tag?.nodeClass) ?? (isMap ? YAMLMap : YAMLSeq))(ctx.schema);
		coll.flow = true;
		const atRoot = ctx.atRoot;
		if (atRoot) ctx.atRoot = false;
		if (ctx.atKey) ctx.atKey = false;
		let offset = fc.offset + fc.start.source.length;
		for (let i = 0; i < fc.items.length; ++i) {
			const collItem = fc.items[i];
			const { start, key, sep, value } = collItem;
			const props = resolveProps(start, {
				flow: fcName,
				indicator: "explicit-key-ind",
				next: key ?? sep?.[0],
				offset,
				onError,
				parentIndent: fc.indent,
				startOnNewline: false
			});
			if (!props.found) {
				if (!props.anchor && !props.tag && !sep && !value) {
					if (i === 0 && props.comma) onError(props.comma, "UNEXPECTED_TOKEN", `Unexpected , in ${fcName}`);
					else if (i < fc.items.length - 1) onError(props.start, "UNEXPECTED_TOKEN", `Unexpected empty item in ${fcName}`);
					if (props.comment) {
						if (coll.comment) coll.comment += "\n" + props.comment;
						else coll.comment = props.comment;
					}
					offset = props.end;
					continue;
				}
				if (!isMap && ctx.options.strict && containsNewline(key)) onError(key, "MULTILINE_IMPLICIT_KEY", "Implicit keys of flow sequence pairs need to be on a single line");
			}
			if (i === 0) {
				if (props.comma) onError(props.comma, "UNEXPECTED_TOKEN", `Unexpected , in ${fcName}`);
			} else {
				if (!props.comma) onError(props.start, "MISSING_CHAR", `Missing , between ${fcName} items`);
				if (props.comment) {
					let prevItemComment = "";
					loop: for (const st of start) switch (st.type) {
						case "comma":
						case "space": break;
						case "comment":
							prevItemComment = st.source.substring(1);
							break loop;
						default: break loop;
					}
					if (prevItemComment) {
						let prev = coll.items[coll.items.length - 1];
						if (isPair(prev)) prev = prev.value ?? prev.key;
						if (prev.comment) prev.comment += "\n" + prevItemComment;
						else prev.comment = prevItemComment;
						props.comment = props.comment.substring(prevItemComment.length + 1);
					}
				}
			}
			if (!isMap && !sep && !props.found) {
				const valueNode = value ? composeNode(ctx, value, props, onError) : composeEmptyNode(ctx, props.end, sep, null, props, onError);
				coll.items.push(valueNode);
				offset = valueNode.range[2];
				if (isBlock(value)) onError(valueNode.range, "BLOCK_IN_FLOW", blockMsg);
			} else {
				ctx.atKey = true;
				const keyStart = props.end;
				const keyNode = key ? composeNode(ctx, key, props, onError) : composeEmptyNode(ctx, keyStart, start, null, props, onError);
				if (isBlock(key)) onError(keyNode.range, "BLOCK_IN_FLOW", blockMsg);
				ctx.atKey = false;
				const valueProps = resolveProps(sep ?? [], {
					flow: fcName,
					indicator: "map-value-ind",
					next: value,
					offset: keyNode.range[2],
					onError,
					parentIndent: fc.indent,
					startOnNewline: false
				});
				if (valueProps.found) {
					if (!isMap && !props.found && ctx.options.strict) {
						if (sep) for (const st of sep) {
							if (st === valueProps.found) break;
							if (st.type === "newline") {
								onError(st, "MULTILINE_IMPLICIT_KEY", "Implicit keys of flow sequence pairs need to be on a single line");
								break;
							}
						}
						if (props.start < valueProps.found.offset - 1024) onError(valueProps.found, "KEY_OVER_1024_CHARS", "The : indicator must be at most 1024 chars after the start of an implicit flow sequence key");
					}
				} else if (value) {
					if ("source" in value && value.source?.[0] === ":") onError(value, "MISSING_CHAR", `Missing space after : in ${fcName}`);
					else onError(valueProps.start, "MISSING_CHAR", `Missing , or : between ${fcName} items`);
				}
				const valueNode = value ? composeNode(ctx, value, valueProps, onError) : valueProps.found ? composeEmptyNode(ctx, valueProps.end, sep, null, valueProps, onError) : null;
				if (valueNode) {
					if (isBlock(value)) onError(valueNode.range, "BLOCK_IN_FLOW", blockMsg);
				} else if (valueProps.comment) {
					if (keyNode.comment) keyNode.comment += "\n" + valueProps.comment;
					else keyNode.comment = valueProps.comment;
				}
				const pair = new Pair(keyNode, valueNode);
				if (ctx.options.keepSourceTokens) pair.srcToken = collItem;
				if (isMap) {
					const map = coll;
					if (mapIncludes(ctx, map.items, keyNode)) onError(keyStart, "DUPLICATE_KEY", "Map keys must be unique");
					map.items.push(pair);
				} else {
					const map = new YAMLMap(ctx.schema);
					map.flow = true;
					map.items.push(pair);
					const endRange = (valueNode ?? keyNode).range;
					map.range = [
						keyNode.range[0],
						endRange[1],
						endRange[2]
					];
					coll.items.push(map);
				}
				offset = valueNode ? valueNode.range[2] : valueProps.end;
			}
		}
		const expectedEnd = isMap ? "}" : "]";
		const [ce, ...ee] = fc.end;
		let cePos = offset;
		if (ce?.source === expectedEnd) cePos = ce.offset + ce.source.length;
		else {
			const name = fcName[0].toUpperCase() + fcName.substring(1);
			const msg = atRoot ? `${name} must end with a ${expectedEnd}` : `${name} in block collection must be sufficiently indented and end with a ${expectedEnd}`;
			onError(offset, atRoot ? "MISSING_CHAR" : "BAD_INDENT", msg);
			if (ce && ce.source.length !== 1) ee.unshift(ce);
		}
		if (ee.length > 0) {
			const end = resolveEnd(ee, cePos, ctx.options.strict, onError);
			if (end.comment) {
				if (coll.comment) coll.comment += "\n" + end.comment;
				else coll.comment = end.comment;
			}
			coll.range = [
				fc.offset,
				cePos,
				end.offset
			];
		} else coll.range = [
			fc.offset,
			cePos,
			cePos
		];
		return coll;
	}
	function resolveCollection(CN, ctx, token, onError, tagName, tag) {
		const coll = token.type === "block-map" ? resolveBlockMap(CN, ctx, token, onError, tag) : token.type === "block-seq" ? resolveBlockSeq(CN, ctx, token, onError, tag) : resolveFlowCollection(CN, ctx, token, onError, tag);
		const Coll = coll.constructor;
		if (tagName === "!" || tagName === Coll.tagName) {
			coll.tag = Coll.tagName;
			return coll;
		}
		if (tagName) coll.tag = tagName;
		return coll;
	}
	function composeCollection(CN, ctx, token, props, onError) {
		const tagToken = props.tag;
		const tagName = !tagToken ? null : ctx.directives.tagName(tagToken.source, (msg) => onError(tagToken, "TAG_RESOLVE_FAILED", msg));
		if (token.type === "block-seq") {
			const { anchor, newlineAfterProp: nl } = props;
			const lastProp = anchor && tagToken ? anchor.offset > tagToken.offset ? anchor : tagToken : anchor ?? tagToken;
			if (lastProp && (!nl || nl.offset < lastProp.offset)) onError(lastProp, "MISSING_CHAR", "Missing newline after block sequence props");
		}
		const expType = token.type === "block-map" ? "map" : token.type === "block-seq" ? "seq" : token.start.source === "{" ? "map" : "seq";
		if (!tagToken || !tagName || tagName === "!" || tagName === YAMLMap.tagName && expType === "map" || tagName === YAMLSeq.tagName && expType === "seq") return resolveCollection(CN, ctx, token, onError, tagName);
		let tag = ctx.schema.tags.find((t) => t.tag === tagName && t.collection === expType);
		if (!tag) {
			const kt = ctx.schema.knownTags[tagName];
			if (kt?.collection === expType) {
				ctx.schema.tags.push(Object.assign({}, kt, { default: false }));
				tag = kt;
			} else {
				if (kt) onError(tagToken, "BAD_COLLECTION_TYPE", `${kt.tag} used for ${expType} collection, but expects ${kt.collection ?? "scalar"}`, true);
				else onError(tagToken, "TAG_RESOLVE_FAILED", `Unresolved tag: ${tagName}`, true);
				return resolveCollection(CN, ctx, token, onError, tagName);
			}
		}
		const coll = resolveCollection(CN, ctx, token, onError, tagName, tag);
		const res = tag.resolve?.(coll, (msg) => onError(tagToken, "TAG_RESOLVE_FAILED", msg), ctx.options) ?? coll;
		const node = isNode(res) ? res : new Scalar(res);
		node.range = coll.range;
		node.tag = tagName;
		if (tag?.format) node.format = tag.format;
		return node;
	}
	function resolveBlockScalar(ctx, scalar, onError) {
		const start = scalar.offset;
		const header = parseBlockScalarHeader(scalar, ctx.options.strict, onError);
		if (!header) return {
			value: "",
			type: null,
			comment: "",
			range: [
				start,
				start,
				start
			]
		};
		const type = header.mode === ">" ? Scalar.BLOCK_FOLDED : Scalar.BLOCK_LITERAL;
		const lines = scalar.source ? splitLines(scalar.source) : [];
		let chompStart = lines.length;
		for (let i = lines.length - 1; i >= 0; --i) {
			const content = lines[i][1];
			if (content === "" || content === "\r") chompStart = i;
			else break;
		}
		if (chompStart === 0) {
			const value = header.chomp === "+" && lines.length > 0 ? "\n".repeat(Math.max(1, lines.length - 1)) : "";
			let end = start + header.length;
			if (scalar.source) end += scalar.source.length;
			return {
				value,
				type,
				comment: header.comment,
				range: [
					start,
					end,
					end
				]
			};
		}
		let trimIndent = scalar.indent + header.indent;
		let offset = scalar.offset + header.length;
		let contentStart = 0;
		for (let i = 0; i < chompStart; ++i) {
			const [indent, content] = lines[i];
			if (content === "" || content === "\r") {
				if (header.indent === 0 && indent.length > trimIndent) trimIndent = indent.length;
			} else {
				if (indent.length < trimIndent) onError(offset + indent.length, "MISSING_CHAR", "Block scalars with more-indented leading empty lines must use an explicit indentation indicator");
				if (header.indent === 0) trimIndent = indent.length;
				contentStart = i;
				if (trimIndent === 0 && !ctx.atRoot) onError(offset, "BAD_INDENT", "Block scalar values in collections must be indented");
				break;
			}
			offset += indent.length + content.length + 1;
		}
		for (let i = lines.length - 1; i >= chompStart; --i) if (lines[i][0].length > trimIndent) chompStart = i + 1;
		let value = "";
		let sep = "";
		let prevMoreIndented = false;
		for (let i = 0; i < contentStart; ++i) value += lines[i][0].slice(trimIndent) + "\n";
		for (let i = contentStart; i < chompStart; ++i) {
			let [indent, content] = lines[i];
			offset += indent.length + content.length + 1;
			const crlf = content[content.length - 1] === "\r";
			if (crlf) content = content.slice(0, -1);
			/* istanbul ignore if already caught in lexer */
			if (content && indent.length < trimIndent) {
				const message = `Block scalar lines must not be less indented than their ${header.indent ? "explicit indentation indicator" : "first line"}`;
				onError(offset - content.length - (crlf ? 2 : 1), "BAD_INDENT", message);
				indent = "";
			}
			if (type === Scalar.BLOCK_LITERAL) {
				value += sep + indent.slice(trimIndent) + content;
				sep = "\n";
			} else if (indent.length > trimIndent || content[0] === "	") {
				if (sep === " ") sep = "\n";
				else if (!prevMoreIndented && sep === "\n") sep = "\n\n";
				value += sep + indent.slice(trimIndent) + content;
				sep = "\n";
				prevMoreIndented = true;
			} else if (content === "") {
				if (sep === "\n") value += "\n";
				else sep = "\n";
			} else {
				value += sep + content;
				sep = " ";
				prevMoreIndented = false;
			}
		}
		switch (header.chomp) {
			case "-": break;
			case "+":
				for (let i = chompStart; i < lines.length; ++i) value += "\n" + lines[i][0].slice(trimIndent);
				if (value[value.length - 1] !== "\n") value += "\n";
				break;
			default: value += "\n";
		}
		const end = start + header.length + scalar.source.length;
		return {
			value,
			type,
			comment: header.comment,
			range: [
				start,
				end,
				end
			]
		};
	}
	function parseBlockScalarHeader({ offset, props }, strict, onError) {
		/* istanbul ignore if should not happen */
		if (props[0].type !== "block-scalar-header") {
			onError(props[0], "IMPOSSIBLE", "Block scalar header not found");
			return null;
		}
		const { source } = props[0];
		const mode = source[0];
		let indent = 0;
		let chomp = "";
		let error = -1;
		for (let i = 1; i < source.length; ++i) {
			const ch = source[i];
			if (!chomp && (ch === "-" || ch === "+")) chomp = ch;
			else {
				const n = Number(ch);
				if (!indent && n) indent = n;
				else if (error === -1) error = offset + i;
			}
		}
		if (error !== -1) onError(error, "UNEXPECTED_TOKEN", `Block scalar header includes extra characters: ${source}`);
		let hasSpace = false;
		let comment = "";
		let length = source.length;
		for (let i = 1; i < props.length; ++i) {
			const token = props[i];
			switch (token.type) {
				case "space": hasSpace = true;
				case "newline":
					length += token.source.length;
					break;
				case "comment":
					if (strict && !hasSpace) onError(token, "MISSING_CHAR", "Comments must be separated from other tokens by white space characters");
					length += token.source.length;
					comment = token.source.substring(1);
					break;
				case "error":
					onError(token, "UNEXPECTED_TOKEN", token.message);
					length += token.source.length;
					break;
				/* istanbul ignore next should not happen */
				default: {
					onError(token, "UNEXPECTED_TOKEN", `Unexpected token in block scalar header: ${token.type}`);
					const ts = token.source;
					if (ts && typeof ts === "string") length += ts.length;
				}
			}
		}
		return {
			mode,
			indent,
			chomp,
			comment,
			length
		};
	}
	/** @returns Array of lines split up as `[indent, content]` */
	function splitLines(source) {
		const split = source.split(/\n( *)/);
		const first = split[0];
		const m = first.match(/^( *)/);
		const lines = [m?.[1] ? [m[1], first.slice(m[1].length)] : ["", first]];
		for (let i = 1; i < split.length; i += 2) lines.push([split[i], split[i + 1]]);
		return lines;
	}
	function resolveFlowScalar(scalar, strict, onError) {
		const { offset, type, source, end } = scalar;
		let _type;
		let value;
		const _onError = (rel, code, msg) => onError(offset + rel, code, msg);
		switch (type) {
			case "scalar":
				_type = Scalar.PLAIN;
				value = plainValue(source, _onError);
				break;
			case "single-quoted-scalar":
				_type = Scalar.QUOTE_SINGLE;
				value = singleQuotedValue(source, _onError);
				break;
			case "double-quoted-scalar":
				_type = Scalar.QUOTE_DOUBLE;
				value = doubleQuotedValue(source, _onError);
				break;
			/* istanbul ignore next should not happen */
			default:
				onError(scalar, "UNEXPECTED_TOKEN", `Expected a flow scalar value, but found: ${type}`);
				return {
					value: "",
					type: null,
					comment: "",
					range: [
						offset,
						offset + source.length,
						offset + source.length
					]
				};
		}
		const valueEnd = offset + source.length;
		const re = resolveEnd(end, valueEnd, strict, onError);
		return {
			value,
			type: _type,
			comment: re.comment,
			range: [
				offset,
				valueEnd,
				re.offset
			]
		};
	}
	function plainValue(source, onError) {
		let badChar = "";
		switch (source[0]) {
			/* istanbul ignore next should not happen */
			case "	":
				badChar = "a tab character";
				break;
			case ",":
				badChar = "flow indicator character ,";
				break;
			case "%":
				badChar = "directive indicator character %";
				break;
			case "|":
			case ">":
				badChar = `block scalar indicator ${source[0]}`;
				break;
			case "@":
			case "`": badChar = `reserved character ${source[0]}`;
		}
		if (badChar) onError(0, "BAD_SCALAR_START", `Plain value cannot start with ${badChar}`);
		return foldLines(source);
	}
	function singleQuotedValue(source, onError) {
		if (source[source.length - 1] !== "'" || source.length === 1) onError(source.length, "MISSING_CHAR", "Missing closing 'quote");
		return foldLines(source.slice(1, -1)).replace(/''/g, "'");
	}
	function foldLines(source) {
		/**
		* The negative lookbehind here and in the `re` RegExp is to
		* prevent causing a polynomial search time in certain cases.
		*
		* The try-catch is for Safari, which doesn't support this yet:
		* https://caniuse.com/js-regexp-lookbehind
		*/
		let first, line;
		try {
			first = /* @__PURE__ */ new RegExp("(.*?)(?<![ 	])[ 	]*\r?\n", "sy");
			line = /* @__PURE__ */ new RegExp("[ 	]*(.*?)(?:(?<![ 	])[ 	]*)?\r?\n", "sy");
		} catch {
			first = /(.*?)[ \t]*\r?\n/sy;
			line = /[ \t]*(.*?)[ \t]*\r?\n/sy;
		}
		let match = first.exec(source);
		if (!match) return source;
		let res = match[1];
		let sep = " ";
		let pos = first.lastIndex;
		line.lastIndex = pos;
		while (match = line.exec(source)) {
			if (match[1] === "") {
				if (sep === "\n") res += sep;
				else sep = "\n";
			} else {
				res += sep + match[1];
				sep = " ";
			}
			pos = line.lastIndex;
		}
		const last = /[ \t]*(.*)/sy;
		last.lastIndex = pos;
		match = last.exec(source);
		return res + sep + (match?.[1] ?? "");
	}
	function doubleQuotedValue(source, onError) {
		let res = "";
		for (let i = 1; i < source.length - 1; ++i) {
			const ch = source[i];
			if (ch === "\r" && source[i + 1] === "\n") continue;
			if (ch === "\n") {
				const { fold, offset } = foldNewline(source, i);
				res += fold;
				i = offset;
			} else if (ch === "\\") {
				let next = source[++i];
				const cc = escapeCodes[next];
				if (cc) res += cc;
				else if (next === "\n") {
					next = source[i + 1];
					while (next === " " || next === "	") next = source[++i + 1];
				} else if (next === "\r" && source[i + 1] === "\n") {
					next = source[++i + 1];
					while (next === " " || next === "	") next = source[++i + 1];
				} else if (next === "x" || next === "u" || next === "U") {
					const length = next === "x" ? 2 : next === "u" ? 4 : 8;
					res += parseCharCode(source, i + 1, length, onError);
					i += length;
				} else {
					const raw = source.substr(i - 1, 2);
					onError(i - 1, "BAD_DQ_ESCAPE", `Invalid escape sequence ${raw}`);
					res += raw;
				}
			} else if (ch === " " || ch === "	") {
				const wsStart = i;
				let next = source[i + 1];
				while (next === " " || next === "	") next = source[++i + 1];
				if (next !== "\n" && !(next === "\r" && source[i + 2] === "\n")) res += i > wsStart ? source.slice(wsStart, i + 1) : ch;
			} else res += ch;
		}
		if (source[source.length - 1] !== "\"" || source.length === 1) onError(source.length, "MISSING_CHAR", "Missing closing \"quote");
		return res;
	}
	/**
	* Fold a single newline into a space, multiple newlines to N - 1 newlines.
	* Presumes `source[offset] === '\n'`
	*/
	function foldNewline(source, offset) {
		let fold = "";
		let ch = source[offset + 1];
		while (ch === " " || ch === "	" || ch === "\n" || ch === "\r") {
			if (ch === "\r" && source[offset + 2] !== "\n") break;
			if (ch === "\n") fold += "\n";
			offset += 1;
			ch = source[offset + 1];
		}
		if (!fold) fold = " ";
		return {
			fold,
			offset
		};
	}
	var escapeCodes = {
		"0": "\0",
		a: "\x07",
		b: "\b",
		e: "\x1B",
		f: "\f",
		n: "\n",
		r: "\r",
		t: "	",
		v: "\v",
		N: "",
		_: "\xA0",
		L: "\u2028",
		P: "\u2029",
		" ": " ",
		"\"": "\"",
		"/": "/",
		"\\": "\\",
		"	": "	"
	};
	function parseCharCode(source, offset, length, onError) {
		const cc = source.substr(offset, length);
		const code = cc.length === length && /^[0-9a-fA-F]+$/.test(cc) ? parseInt(cc, 16) : NaN;
		try {
			return String.fromCodePoint(code);
		} catch {
			const raw = source.substr(offset - 2, length + 2);
			onError(offset - 2, "BAD_DQ_ESCAPE", `Invalid escape sequence ${raw}`);
			return raw;
		}
	}
	function composeScalar(ctx, token, tagToken, onError) {
		const { value, type, comment, range } = token.type === "block-scalar" ? resolveBlockScalar(ctx, token, onError) : resolveFlowScalar(token, ctx.options.strict, onError);
		const tagName = tagToken ? ctx.directives.tagName(tagToken.source, (msg) => onError(tagToken, "TAG_RESOLVE_FAILED", msg)) : null;
		let tag;
		if (ctx.options.stringKeys && ctx.atKey) tag = ctx.schema[SCALAR$1];
		else if (tagName) tag = findScalarTagByName(ctx.schema, value, tagName, tagToken, onError);
		else if (token.type === "scalar") tag = findScalarTagByTest(ctx, value, token, onError);
		else tag = ctx.schema[SCALAR$1];
		let scalar;
		try {
			const res = tag.resolve(value, (msg) => onError(tagToken ?? token, "TAG_RESOLVE_FAILED", msg), ctx.options);
			scalar = isScalar(res) ? res : new Scalar(res);
		} catch (error) {
			const msg = error instanceof Error ? error.message : String(error);
			onError(tagToken ?? token, "TAG_RESOLVE_FAILED", msg);
			scalar = new Scalar(value);
		}
		scalar.range = range;
		scalar.source = value;
		if (type) scalar.type = type;
		if (tagName) scalar.tag = tagName;
		if (tag.format) scalar.format = tag.format;
		if (comment) scalar.comment = comment;
		return scalar;
	}
	function findScalarTagByName(schema, value, tagName, tagToken, onError) {
		if (tagName === "!") return schema[SCALAR$1];
		const matchWithTest = [];
		for (const tag of schema.tags) if (!tag.collection && tag.tag === tagName) {
			if (tag.default && tag.test) matchWithTest.push(tag);
			else return tag;
		}
		for (const tag of matchWithTest) if (tag.test?.test(value)) return tag;
		const kt = schema.knownTags[tagName];
		if (kt && !kt.collection) {
			schema.tags.push(Object.assign({}, kt, {
				default: false,
				test: void 0
			}));
			return kt;
		}
		onError(tagToken, "TAG_RESOLVE_FAILED", `Unresolved tag: ${tagName}`, tagName !== "tag:yaml.org,2002:str");
		return schema[SCALAR$1];
	}
	function findScalarTagByTest({ atKey, directives, schema }, value, token, onError) {
		const tag = schema.tags.find((tag) => (tag.default === true || atKey && tag.default === "key") && tag.test?.test(value)) || schema[SCALAR$1];
		if (schema.compat) {
			const compat = schema.compat.find((tag) => tag.default && tag.test?.test(value)) ?? schema[SCALAR$1];
			if (tag.tag !== compat.tag) onError(token, "TAG_RESOLVE_FAILED", `Value may be parsed as either ${directives.tagString(tag.tag)} or ${directives.tagString(compat.tag)}`, true);
		}
		return tag;
	}
	function emptyScalarPosition(offset, before, pos) {
		if (before) {
			pos ?? (pos = before.length);
			for (let i = pos - 1; i >= 0; --i) {
				let st = before[i];
				switch (st.type) {
					case "space":
					case "comment":
					case "newline":
						offset -= st.source.length;
						continue;
				}
				st = before[++i];
				while (st?.type === "space") {
					offset += st.source.length;
					st = before[++i];
				}
				break;
			}
		}
		return offset;
	}
	var CN = {
		composeNode,
		composeEmptyNode
	};
	function composeNode(ctx, token, props, onError) {
		const atKey = ctx.atKey;
		const { spaceBefore, comment, anchor, tag } = props;
		let node;
		let isSrcToken = true;
		switch (token.type) {
			case "alias":
				node = composeAlias(ctx, token, onError);
				if (anchor || tag) onError(token, "ALIAS_PROPS", "An alias node must not specify any properties");
				break;
			case "scalar":
			case "single-quoted-scalar":
			case "double-quoted-scalar":
			case "block-scalar":
				node = composeScalar(ctx, token, tag, onError);
				if (anchor) node.anchor = anchor.source.substring(1);
				break;
			case "block-map":
			case "block-seq":
			case "flow-collection":
				try {
					node = composeCollection(CN, ctx, token, props, onError);
					if (anchor) node.anchor = anchor.source.substring(1);
				} catch (error) {
					onError(token, "RESOURCE_EXHAUSTION", error instanceof Error ? error.message : String(error));
				}
				break;
			default:
				onError(token, "UNEXPECTED_TOKEN", token.type === "error" ? token.message : `Unsupported token (type: ${token.type})`);
				isSrcToken = false;
		}
		node ?? (node = composeEmptyNode(ctx, token.offset, void 0, null, props, onError));
		if (anchor && node.anchor === "") onError(anchor, "BAD_ALIAS", "Anchor cannot be an empty string");
		if (atKey && ctx.options.stringKeys && (!isScalar(node) || typeof node.value !== "string" || node.tag && node.tag !== "tag:yaml.org,2002:str")) onError(tag ?? token, "NON_STRING_KEY", "With stringKeys, all keys must be strings");
		if (spaceBefore) node.spaceBefore = true;
		if (comment) {
			if (token.type === "scalar" && token.source === "") node.comment = comment;
			else node.commentBefore = comment;
		}
		if (ctx.options.keepSourceTokens && isSrcToken) node.srcToken = token;
		return node;
	}
	function composeEmptyNode(ctx, offset, before, pos, { spaceBefore, comment, anchor, tag, end }, onError) {
		const node = composeScalar(ctx, {
			type: "scalar",
			offset: emptyScalarPosition(offset, before, pos),
			indent: -1,
			source: ""
		}, tag, onError);
		if (anchor) {
			node.anchor = anchor.source.substring(1);
			if (node.anchor === "") onError(anchor, "BAD_ALIAS", "Anchor cannot be an empty string");
		}
		if (spaceBefore) node.spaceBefore = true;
		if (comment) {
			node.comment = comment;
			node.range[2] = end;
		}
		return node;
	}
	function composeAlias({ options }, { offset, source, end }, onError) {
		const alias = new Alias(source.substring(1));
		if (alias.source === "") onError(offset, "BAD_ALIAS", "Alias cannot be an empty string");
		if (alias.source.endsWith(":")) onError(offset + source.length - 1, "BAD_ALIAS", "Alias ending in : is ambiguous", true);
		const valueEnd = offset + source.length;
		const re = resolveEnd(end, valueEnd, options.strict, onError);
		alias.range = [
			offset,
			valueEnd,
			re.offset
		];
		if (re.comment) alias.comment = re.comment;
		return alias;
	}
	function composeDoc(options, directives, { offset, start, value, end }, onError) {
		const doc = new Document(void 0, Object.assign({ _directives: directives }, options));
		const ctx = {
			atKey: false,
			atRoot: true,
			directives: doc.directives,
			options: doc.options,
			schema: doc.schema
		};
		const props = resolveProps(start, {
			indicator: "doc-start",
			next: value ?? end?.[0],
			offset,
			onError,
			parentIndent: 0,
			startOnNewline: true
		});
		if (props.found) {
			doc.directives.docStart = true;
			if (value && (value.type === "block-map" || value.type === "block-seq") && !props.hasNewline) onError(props.end, "MISSING_CHAR", "Block collection cannot start on same line with directives-end marker");
		}
		doc.contents = value ? composeNode(ctx, value, props, onError) : composeEmptyNode(ctx, props.end, start, null, props, onError);
		const contentEnd = doc.contents.range[2];
		const re = resolveEnd(end, contentEnd, false, onError);
		if (re.comment) doc.comment = re.comment;
		doc.range = [
			offset,
			contentEnd,
			re.offset
		];
		return doc;
	}
	function getErrorPos(src) {
		if (typeof src === "number") return [src, src + 1];
		if (Array.isArray(src)) return src.length === 2 ? src : [src[0], src[1]];
		const { offset, source } = src;
		return [offset, offset + (typeof source === "string" ? source.length : 1)];
	}
	function parsePrelude(prelude) {
		let comment = "";
		let atComment = false;
		let afterEmptyLine = false;
		for (let i = 0; i < prelude.length; ++i) {
			const source = prelude[i];
			switch (source[0]) {
				case "#":
					comment += (comment === "" ? "" : afterEmptyLine ? "\n\n" : "\n") + (source.substring(1) || " ");
					atComment = true;
					afterEmptyLine = false;
					break;
				case "%":
					if (prelude[i + 1]?.[0] !== "#") i += 1;
					atComment = false;
					break;
				default:
					if (!atComment) afterEmptyLine = true;
					atComment = false;
			}
		}
		return {
			comment,
			afterEmptyLine
		};
	}
	/**
	* Compose a stream of CST nodes into a stream of YAML Documents.
	*
	* ```ts
	* import { Composer, Parser } from 'yaml'
	*
	* const src: string = ...
	* const tokens = new Parser().parse(src)
	* const docs = new Composer().compose(tokens)
	* ```
	*/
	var Composer = class {
		constructor(options = {}) {
			this.doc = null;
			this.atDirectives = false;
			this.prelude = [];
			this.errors = [];
			this.warnings = [];
			this.onError = (source, code, message, warning) => {
				const pos = getErrorPos(source);
				if (warning) this.warnings.push(new YAMLWarning(pos, code, message));
				else this.errors.push(new YAMLParseError(pos, code, message));
			};
			this.directives = new Directives({ version: options.version || "1.2" });
			this.options = options;
		}
		decorate(doc, afterDoc) {
			const { comment, afterEmptyLine } = parsePrelude(this.prelude);
			if (comment) {
				const dc = doc.contents;
				if (afterDoc) doc.comment = doc.comment ? `${doc.comment}\n${comment}` : comment;
				else if (afterEmptyLine || doc.directives.docStart || !dc) doc.commentBefore = comment;
				else if (isCollection(dc) && !dc.flow && dc.items.length > 0) {
					let it = dc.items[0];
					if (isPair(it)) it = it.key;
					const cb = it.commentBefore;
					it.commentBefore = cb ? `${comment}\n${cb}` : comment;
				} else {
					const cb = dc.commentBefore;
					dc.commentBefore = cb ? `${comment}\n${cb}` : comment;
				}
			}
			if (afterDoc) {
				for (let i = 0; i < this.errors.length; ++i) doc.errors.push(this.errors[i]);
				for (let i = 0; i < this.warnings.length; ++i) doc.warnings.push(this.warnings[i]);
			} else {
				doc.errors = this.errors;
				doc.warnings = this.warnings;
			}
			this.prelude = [];
			this.errors = [];
			this.warnings = [];
		}
		/**
		* Current stream status information.
		*
		* Mostly useful at the end of input for an empty stream.
		*/
		streamInfo() {
			return {
				comment: parsePrelude(this.prelude).comment,
				directives: this.directives,
				errors: this.errors,
				warnings: this.warnings
			};
		}
		/**
		* Compose tokens into documents.
		*
		* @param forceDoc - If the stream contains no document, still emit a final document including any comments and directives that would be applied to a subsequent document.
		* @param endOffset - Should be set if `forceDoc` is also set, to set the document range end and to indicate errors correctly.
		*/
		*compose(tokens, forceDoc = false, endOffset = -1) {
			for (const token of tokens) yield* this.next(token);
			yield* this.end(forceDoc, endOffset);
		}
		/** Advance the composer by one CST token. */
		*next(token) {
			switch (token.type) {
				case "directive":
					this.directives.add(token.source, (offset, message, warning) => {
						const pos = getErrorPos(token);
						pos[0] += offset;
						this.onError(pos, "BAD_DIRECTIVE", message, warning);
					});
					this.prelude.push(token.source);
					this.atDirectives = true;
					break;
				case "document": {
					const doc = composeDoc(this.options, this.directives, token, this.onError);
					if (this.atDirectives && !doc.directives.docStart) this.onError(token, "MISSING_CHAR", "Missing directives-end/doc-start indicator line");
					this.decorate(doc, false);
					if (this.doc) yield this.doc;
					this.doc = doc;
					this.atDirectives = false;
					break;
				}
				case "byte-order-mark":
				case "space": break;
				case "comment":
				case "newline":
					this.prelude.push(token.source);
					break;
				case "error": {
					const msg = token.source ? `${token.message}: ${JSON.stringify(token.source)}` : token.message;
					const error = new YAMLParseError(getErrorPos(token), "UNEXPECTED_TOKEN", msg);
					if (this.atDirectives || !this.doc) this.errors.push(error);
					else this.doc.errors.push(error);
					break;
				}
				case "doc-end": {
					if (!this.doc) {
						this.errors.push(new YAMLParseError(getErrorPos(token), "UNEXPECTED_TOKEN", "Unexpected doc-end without preceding document"));
						break;
					}
					this.doc.directives.docEnd = true;
					const end = resolveEnd(token.end, token.offset + token.source.length, this.doc.options.strict, this.onError);
					this.decorate(this.doc, true);
					if (end.comment) {
						const dc = this.doc.comment;
						this.doc.comment = dc ? `${dc}\n${end.comment}` : end.comment;
					}
					this.doc.range[2] = end.offset;
					break;
				}
				default: this.errors.push(new YAMLParseError(getErrorPos(token), "UNEXPECTED_TOKEN", `Unsupported token ${token.type}`));
			}
		}
		/**
		* Call at end of input to yield any remaining document.
		*
		* @param forceDoc - If the stream contains no document, still emit a final document including any comments and directives that would be applied to a subsequent document.
		* @param endOffset - Should be set if `forceDoc` is also set, to set the document range end and to indicate errors correctly.
		*/
		*end(forceDoc = false, endOffset = -1) {
			if (this.doc) {
				this.decorate(this.doc, true);
				yield this.doc;
				this.doc = null;
			} else if (forceDoc) {
				const doc = new Document(void 0, Object.assign({ _directives: this.directives }, this.options));
				if (this.atDirectives) this.onError(endOffset, "MISSING_CHAR", "Missing directives-end indicator line");
				doc.range = [
					0,
					endOffset,
					endOffset
				];
				this.decorate(doc, false);
				yield doc;
			}
		}
	};
	var BREAK = Symbol("break visit");
	var SKIP = Symbol("skip children");
	var REMOVE = Symbol("remove item");
	/**
	* Apply a visitor to a CST document or item.
	*
	* Walks through the tree (depth-first) starting from the root, calling a
	* `visitor` function with two arguments when entering each item:
	*   - `item`: The current item, which included the following members:
	*     - `start: SourceToken[]` – Source tokens before the key or value,
	*       possibly including its anchor or tag.
	*     - `key?: Token | null` – Set for pair values. May then be `null`, if
	*       the key before the `:` separator is empty.
	*     - `sep?: SourceToken[]` – Source tokens between the key and the value,
	*       which should include the `:` map value indicator if `value` is set.
	*     - `value?: Token` – The value of a sequence item, or of a map pair.
	*   - `path`: The steps from the root to the current node, as an array of
	*     `['key' | 'value', number]` tuples.
	*
	* The return value of the visitor may be used to control the traversal:
	*   - `undefined` (default): Do nothing and continue
	*   - `visit.SKIP`: Do not visit the children of this token, continue with
	*      next sibling
	*   - `visit.BREAK`: Terminate traversal completely
	*   - `visit.REMOVE`: Remove the current item, then continue with the next one
	*   - `number`: Set the index of the next step. This is useful especially if
	*     the index of the current token has changed.
	*   - `function`: Define the next visitor for this item. After the original
	*     visitor is called on item entry, next visitors are called after handling
	*     a non-empty `key` and when exiting the item.
	*/
	function visit(cst, visitor) {
		if ("type" in cst && cst.type === "document") cst = {
			start: cst.start,
			value: cst.value
		};
		_visit(Object.freeze([]), cst, visitor);
	}
	/** Terminate visit traversal completely */
	visit.BREAK = BREAK;
	/** Do not visit the children of the current item */
	visit.SKIP = SKIP;
	/** Remove the current item */
	visit.REMOVE = REMOVE;
	/** Find the item at `path` from `cst` as the root */
	visit.itemAtPath = (cst, path) => {
		let item = cst;
		for (const [field, index] of path) {
			const tok = item?.[field];
			if (tok && "items" in tok) item = tok.items[index];
			else return void 0;
		}
		return item;
	};
	/**
	* Get the immediate parent collection of the item at `path` from `cst` as the root.
	*
	* Throws an error if the collection is not found, which should never happen if the item itself exists.
	*/
	visit.parentCollection = (cst, path) => {
		const parent = visit.itemAtPath(cst, path.slice(0, -1));
		const field = path[path.length - 1][0];
		const coll = parent?.[field];
		if (coll && "items" in coll) return coll;
		throw new Error("Parent collection not found");
	};
	function _visit(path, item, visitor) {
		let ctrl = visitor(item, path);
		if (typeof ctrl === "symbol") return ctrl;
		for (const field of ["key", "value"]) {
			const token = item[field];
			if (token && "items" in token) {
				for (let i = 0; i < token.items.length; ++i) {
					const ci = _visit(Object.freeze(path.concat([[field, i]])), token.items[i], visitor);
					if (typeof ci === "number") i = ci - 1;
					else if (ci === BREAK) return BREAK;
					else if (ci === REMOVE) {
						token.items.splice(i, 1);
						i -= 1;
					}
				}
				if (typeof ctrl === "function" && field === "key") ctrl = ctrl(item, path);
			}
		}
		return typeof ctrl === "function" ? ctrl(item, path) : ctrl;
	}
	/** Identify the type of a lexer token. May return `null` for unknown tokens. */
	function tokenType(source) {
		switch (source) {
			case "﻿": return "byte-order-mark";
			case "": return "doc-mode";
			case "": return "flow-error-end";
			case "": return "scalar";
			case "---": return "doc-start";
			case "...": return "doc-end";
			case "":
			case "\n":
			case "\r\n": return "newline";
			case "-": return "seq-item-ind";
			case "?": return "explicit-key-ind";
			case ":": return "map-value-ind";
			case "{": return "flow-map-start";
			case "}": return "flow-map-end";
			case "[": return "flow-seq-start";
			case "]": return "flow-seq-end";
			case ",": return "comma";
		}
		switch (source[0]) {
			case " ":
			case "	": return "space";
			case "#": return "comment";
			case "%": return "directive-line";
			case "*": return "alias";
			case "&": return "anchor";
			case "!": return "tag";
			case "'": return "single-quoted-scalar";
			case "\"": return "double-quoted-scalar";
			case "|":
			case ">": return "block-scalar-header";
		}
		return null;
	}
	function isEmpty(ch) {
		switch (ch) {
			case void 0:
			case " ":
			case "\n":
			case "\r":
			case "	": return true;
			default: return false;
		}
	}
	var hexDigits = /* @__PURE__ */ new Set("0123456789ABCDEFabcdef");
	var tagChars = /* @__PURE__ */ new Set("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-#;/?:@&=+$_.!~*'()");
	var flowIndicatorChars = /* @__PURE__ */ new Set(",[]{}");
	var invalidAnchorChars = /* @__PURE__ */ new Set(" ,[]{}\n\r	");
	var isNotAnchorChar = (ch) => !ch || invalidAnchorChars.has(ch);
	/**
	* Splits an input string into lexical tokens, i.e. smaller strings that are
	* easily identifiable by `tokens.tokenType()`.
	*
	* Lexing starts always in a "stream" context. Incomplete input may be buffered
	* until a complete token can be emitted.
	*
	* In addition to slices of the original input, the following control characters
	* may also be emitted:
	*
	* - `\x02` (Start of Text): A document starts with the next token
	* - `\x18` (Cancel): Unexpected end of flow-mode (indicates an error)
	* - `\x1f` (Unit Separator): Next token is a scalar value
	* - `\u{FEFF}` (Byte order mark): Emitted separately outside documents
	*/
	var Lexer = class {
		constructor() {
			/**
			* Flag indicating whether the end of the current buffer marks the end of
			* all input
			*/
			this.atEnd = false;
			/**
			* Explicit indent set in block scalar header, as an offset from the current
			* minimum indent, so e.g. set to 1 from a header `|2+`. Set to -1 if not
			* explicitly set.
			*/
			this.blockScalarIndent = -1;
			/**
			* Block scalars that include a + (keep) chomping indicator in their header
			* include trailing empty lines, which are otherwise excluded from the
			* scalar's contents.
			*/
			this.blockScalarKeep = false;
			/** Current input */
			this.buffer = "";
			/**
			* Flag noting whether the map value indicator : can immediately follow this
			* node within a flow context.
			*/
			this.flowKey = false;
			/** Count of surrounding flow collection levels. */
			this.flowLevel = 0;
			/**
			* Minimum level of indentation required for next lines to be parsed as a
			* part of the current scalar value.
			*/
			this.indentNext = 0;
			/** Indentation level of the current line. */
			this.indentValue = 0;
			/** Position of the next \n character. */
			this.lineEndPos = null;
			/** Stores the state of the lexer if reaching the end of incpomplete input */
			this.next = null;
			/** A pointer to `buffer`; the current position of the lexer. */
			this.pos = 0;
		}
		/**
		* Generate YAML tokens from the `source` string. If `incomplete`,
		* a part of the last line may be left as a buffer for the next call.
		*
		* @returns A generator of lexical tokens
		*/
		*lex(source, incomplete = false) {
			if (source) {
				if (typeof source !== "string") throw TypeError("source is not a string");
				this.buffer = this.buffer ? this.buffer + source : source;
				this.lineEndPos = null;
			}
			this.atEnd = !incomplete;
			let next = this.next ?? "stream";
			while (next && (incomplete || this.hasChars(1))) next = yield* this.parseNext(next);
		}
		atLineEnd() {
			let i = this.pos;
			let ch = this.buffer[i];
			while (ch === " " || ch === "	") ch = this.buffer[++i];
			if (!ch || ch === "#" || ch === "\n") return true;
			if (ch === "\r") return this.buffer[i + 1] === "\n";
			return false;
		}
		charAt(n) {
			return this.buffer[this.pos + n];
		}
		continueScalar(offset) {
			let ch = this.buffer[offset];
			if (this.indentNext > 0) {
				let indent = 0;
				while (ch === " ") ch = this.buffer[++indent + offset];
				if (ch === "\r") {
					const next = this.buffer[indent + offset + 1];
					if (next === "\n" || !next && !this.atEnd) return offset + indent + 1;
				}
				return ch === "\n" || indent >= this.indentNext || !ch && !this.atEnd ? offset + indent : -1;
			}
			if (ch === "-" || ch === ".") {
				const dt = this.buffer.substr(offset, 3);
				if ((dt === "---" || dt === "...") && isEmpty(this.buffer[offset + 3])) return -1;
			}
			return offset;
		}
		getLine() {
			let end = this.lineEndPos;
			if (typeof end !== "number" || end !== -1 && end < this.pos) {
				end = this.buffer.indexOf("\n", this.pos);
				this.lineEndPos = end;
			}
			if (end === -1) return this.atEnd ? this.buffer.substring(this.pos) : null;
			if (this.buffer[end - 1] === "\r") end -= 1;
			return this.buffer.substring(this.pos, end);
		}
		hasChars(n) {
			return this.pos + n <= this.buffer.length;
		}
		setNext(state) {
			this.buffer = this.buffer.substring(this.pos);
			this.pos = 0;
			this.lineEndPos = null;
			this.next = state;
			return null;
		}
		peek(n) {
			return this.buffer.substr(this.pos, n);
		}
		*parseNext(next) {
			switch (next) {
				case "stream": return yield* this.parseStream();
				case "line-start": return yield* this.parseLineStart();
				case "block-start": return yield* this.parseBlockStart();
				case "doc": return yield* this.parseDocument();
				case "flow": return yield* this.parseFlowCollection();
				case "quoted-scalar": return yield* this.parseQuotedScalar();
				case "block-scalar": return yield* this.parseBlockScalar();
				case "plain-scalar": return yield* this.parsePlainScalar();
			}
		}
		*parseStream() {
			let line = this.getLine();
			if (line === null) return this.setNext("stream");
			if (line[0] === "﻿") {
				yield* this.pushCount(1);
				line = line.substring(1);
			}
			if (line[0] === "%") {
				let dirEnd = line.length;
				let cs = line.indexOf("#");
				while (cs !== -1) {
					const ch = line[cs - 1];
					if (ch === " " || ch === "	") {
						dirEnd = cs - 1;
						break;
					} else cs = line.indexOf("#", cs + 1);
				}
				while (true) {
					const ch = line[dirEnd - 1];
					if (ch === " " || ch === "	") dirEnd -= 1;
					else break;
				}
				const n = (yield* this.pushCount(dirEnd)) + (yield* this.pushSpaces(true));
				yield* this.pushCount(line.length - n);
				this.pushNewline();
				return "stream";
			}
			if (this.atLineEnd()) {
				const sp = yield* this.pushSpaces(true);
				yield* this.pushCount(line.length - sp);
				yield* this.pushNewline();
				return "stream";
			}
			yield "";
			return yield* this.parseLineStart();
		}
		*parseLineStart() {
			const ch = this.charAt(0);
			if (!ch && !this.atEnd) return this.setNext("line-start");
			if (ch === "-" || ch === ".") {
				if (!this.atEnd && !this.hasChars(4)) return this.setNext("line-start");
				const s = this.peek(3);
				if ((s === "---" || s === "...") && isEmpty(this.charAt(3))) {
					yield* this.pushCount(3);
					this.indentValue = 0;
					this.indentNext = 0;
					return s === "---" ? "doc" : "stream";
				}
			}
			this.indentValue = yield* this.pushSpaces(false);
			if (this.indentNext > this.indentValue && !isEmpty(this.charAt(1))) this.indentNext = this.indentValue;
			return yield* this.parseBlockStart();
		}
		*parseBlockStart() {
			const [ch0, ch1] = this.peek(2);
			if (!ch1 && !this.atEnd) return this.setNext("block-start");
			if ((ch0 === "-" || ch0 === "?" || ch0 === ":") && isEmpty(ch1)) {
				const n = (yield* this.pushCount(1)) + (yield* this.pushSpaces(true));
				this.indentNext = this.indentValue + 1;
				this.indentValue += n;
				return "block-start";
			}
			return "doc";
		}
		*parseDocument() {
			yield* this.pushSpaces(true);
			const line = this.getLine();
			if (line === null) return this.setNext("doc");
			let n = yield* this.pushIndicators();
			switch (line[n]) {
				case "#": yield* this.pushCount(line.length - n);
				case void 0:
					yield* this.pushNewline();
					return yield* this.parseLineStart();
				case "{":
				case "[":
					yield* this.pushCount(1);
					this.flowKey = false;
					this.flowLevel = 1;
					return "flow";
				case "}":
				case "]":
					yield* this.pushCount(1);
					return "doc";
				case "*":
					yield* this.pushUntil(isNotAnchorChar);
					return "doc";
				case "\"":
				case "'": return yield* this.parseQuotedScalar();
				case "|":
				case ">":
					n += yield* this.parseBlockScalarHeader();
					n += yield* this.pushSpaces(true);
					yield* this.pushCount(line.length - n);
					yield* this.pushNewline();
					return yield* this.parseBlockScalar();
				default: return yield* this.parsePlainScalar();
			}
		}
		*parseFlowCollection() {
			let nl, sp;
			let indent = -1;
			do {
				nl = yield* this.pushNewline();
				if (nl > 0) {
					sp = yield* this.pushSpaces(false);
					this.indentValue = indent = sp;
				} else sp = 0;
				sp += yield* this.pushSpaces(true);
			} while (nl + sp > 0);
			const line = this.getLine();
			if (line === null) return this.setNext("flow");
			if (indent !== -1 && indent < this.indentNext && line[0] !== "#" || indent === 0 && (line.startsWith("---") || line.startsWith("...")) && isEmpty(line[3])) {
				if (!(indent === this.indentNext - 1 && this.flowLevel === 1 && (line[0] === "]" || line[0] === "}"))) {
					this.flowLevel = 0;
					yield "";
					return yield* this.parseLineStart();
				}
			}
			let n = 0;
			while (line[n] === ",") {
				n += yield* this.pushCount(1);
				n += yield* this.pushSpaces(true);
				this.flowKey = false;
			}
			n += yield* this.pushIndicators();
			switch (line[n]) {
				case void 0: return "flow";
				case "#":
					yield* this.pushCount(line.length - n);
					return "flow";
				case "{":
				case "[":
					yield* this.pushCount(1);
					this.flowKey = false;
					this.flowLevel += 1;
					return "flow";
				case "}":
				case "]":
					yield* this.pushCount(1);
					this.flowKey = true;
					this.flowLevel -= 1;
					return this.flowLevel ? "flow" : "doc";
				case "*":
					yield* this.pushUntil(isNotAnchorChar);
					return "flow";
				case "\"":
				case "'":
					this.flowKey = true;
					return yield* this.parseQuotedScalar();
				case ":": {
					const next = this.charAt(1);
					if (this.flowKey || isEmpty(next) || next === ",") {
						this.flowKey = false;
						yield* this.pushCount(1);
						yield* this.pushSpaces(true);
						return "flow";
					}
				}
				default:
					this.flowKey = false;
					return yield* this.parsePlainScalar();
			}
		}
		*parseQuotedScalar() {
			const quote = this.charAt(0);
			let end = this.buffer.indexOf(quote, this.pos + 1);
			if (quote === "'") while (end !== -1 && this.buffer[end + 1] === "'") end = this.buffer.indexOf("'", end + 2);
			else while (end !== -1) {
				let n = 0;
				while (this.buffer[end - 1 - n] === "\\") n += 1;
				if (n % 2 === 0) break;
				end = this.buffer.indexOf("\"", end + 1);
			}
			const qb = this.buffer.substring(0, end);
			let nl = qb.indexOf("\n", this.pos);
			if (nl !== -1) {
				while (nl !== -1) {
					const cs = this.continueScalar(nl + 1);
					if (cs === -1) break;
					nl = qb.indexOf("\n", cs);
				}
				if (nl !== -1) end = nl - (qb[nl - 1] === "\r" ? 2 : 1);
			}
			if (end === -1) {
				if (!this.atEnd) return this.setNext("quoted-scalar");
				end = this.buffer.length;
			}
			yield* this.pushToIndex(end + 1, false);
			return this.flowLevel ? "flow" : "doc";
		}
		*parseBlockScalarHeader() {
			this.blockScalarIndent = -1;
			this.blockScalarKeep = false;
			let i = this.pos;
			while (true) {
				const ch = this.buffer[++i];
				if (ch === "+") this.blockScalarKeep = true;
				else if (ch > "0" && ch <= "9") this.blockScalarIndent = Number(ch) - 1;
				else if (ch !== "-") break;
			}
			return yield* this.pushUntil((ch) => isEmpty(ch) || ch === "#");
		}
		*parseBlockScalar() {
			let nl = this.pos - 1;
			let indent = 0;
			let ch;
			loop: for (let i = this.pos; ch = this.buffer[i]; ++i) switch (ch) {
				case " ":
					indent += 1;
					break;
				case "\n":
					nl = i;
					indent = 0;
					break;
				case "\r": {
					const next = this.buffer[i + 1];
					if (!next && !this.atEnd) return this.setNext("block-scalar");
					if (next === "\n") break;
				}
				default: break loop;
			}
			if (!ch && !this.atEnd) return this.setNext("block-scalar");
			if (indent >= this.indentNext) {
				if (this.blockScalarIndent === -1) this.indentNext = indent;
				else this.indentNext = this.blockScalarIndent + (this.indentNext === 0 ? 1 : this.indentNext);
				do {
					const cs = this.continueScalar(nl + 1);
					if (cs === -1) break;
					nl = this.buffer.indexOf("\n", cs);
				} while (nl !== -1);
				if (nl === -1) {
					if (!this.atEnd) return this.setNext("block-scalar");
					nl = this.buffer.length;
				}
			}
			let i = nl + 1;
			ch = this.buffer[i];
			while (ch === " ") ch = this.buffer[++i];
			if (ch === "	") {
				while (ch === "	" || ch === " " || ch === "\r" || ch === "\n") ch = this.buffer[++i];
				nl = i - 1;
			} else if (!this.blockScalarKeep) do {
				let i = nl - 1;
				let ch = this.buffer[i];
				if (ch === "\r") ch = this.buffer[--i];
				const lastChar = i;
				while (ch === " ") ch = this.buffer[--i];
				if (ch === "\n" && i >= this.pos && i + 1 + indent > lastChar) nl = i;
				else break;
			} while (true);
			yield "";
			yield* this.pushToIndex(nl + 1, true);
			return yield* this.parseLineStart();
		}
		*parsePlainScalar() {
			const inFlow = this.flowLevel > 0;
			let end = this.pos - 1;
			let i = this.pos - 1;
			let ch;
			while (ch = this.buffer[++i]) if (ch === ":") {
				const next = this.buffer[i + 1];
				if (isEmpty(next) || inFlow && flowIndicatorChars.has(next)) break;
				end = i;
			} else if (isEmpty(ch)) {
				let next = this.buffer[i + 1];
				if (ch === "\r") {
					if (next === "\n") {
						i += 1;
						ch = "\n";
						next = this.buffer[i + 1];
					} else end = i;
				}
				if (next === "#" || inFlow && flowIndicatorChars.has(next)) break;
				if (ch === "\n") {
					const cs = this.continueScalar(i + 1);
					if (cs === -1) break;
					i = Math.max(i, cs - 2);
				}
			} else {
				if (inFlow && flowIndicatorChars.has(ch)) break;
				end = i;
			}
			if (!ch && !this.atEnd) return this.setNext("plain-scalar");
			yield "";
			yield* this.pushToIndex(end + 1, true);
			return inFlow ? "flow" : "doc";
		}
		*pushCount(n) {
			if (n > 0) {
				yield this.buffer.substr(this.pos, n);
				this.pos += n;
				return n;
			}
			return 0;
		}
		*pushToIndex(i, allowEmpty) {
			const s = this.buffer.slice(this.pos, i);
			if (s) {
				yield s;
				this.pos += s.length;
				return s.length;
			} else if (allowEmpty) yield "";
			return 0;
		}
		*pushIndicators() {
			let n = 0;
			loop: while (true) {
				switch (this.charAt(0)) {
					case "!":
						n += yield* this.pushTag();
						n += yield* this.pushSpaces(true);
						continue loop;
					case "&":
						n += yield* this.pushUntil(isNotAnchorChar);
						n += yield* this.pushSpaces(true);
						continue loop;
					case "-":
					case "?":
					case ":": {
						const inFlow = this.flowLevel > 0;
						const ch1 = this.charAt(1);
						if (isEmpty(ch1) || inFlow && flowIndicatorChars.has(ch1)) {
							if (!inFlow) this.indentNext = this.indentValue + 1;
							else if (this.flowKey) this.flowKey = false;
							n += yield* this.pushCount(1);
							n += yield* this.pushSpaces(true);
							continue loop;
						}
					}
				}
				break loop;
			}
			return n;
		}
		*pushTag() {
			if (this.charAt(1) === "<") {
				let i = this.pos + 2;
				let ch = this.buffer[i];
				while (!isEmpty(ch) && ch !== ">") ch = this.buffer[++i];
				return yield* this.pushToIndex(ch === ">" ? i + 1 : i, false);
			} else {
				let i = this.pos + 1;
				let ch = this.buffer[i];
				while (ch) if (tagChars.has(ch)) ch = this.buffer[++i];
				else if (ch === "%" && hexDigits.has(this.buffer[i + 1]) && hexDigits.has(this.buffer[i + 2])) ch = this.buffer[i += 3];
				else break;
				return yield* this.pushToIndex(i, false);
			}
		}
		*pushNewline() {
			const ch = this.buffer[this.pos];
			if (ch === "\n") return yield* this.pushCount(1);
			else if (ch === "\r" && this.charAt(1) === "\n") return yield* this.pushCount(2);
			else return 0;
		}
		*pushSpaces(allowTabs) {
			let i = this.pos - 1;
			let ch;
			do
				ch = this.buffer[++i];
			while (ch === " " || allowTabs && ch === "	");
			const n = i - this.pos;
			if (n > 0) {
				yield this.buffer.substr(this.pos, n);
				this.pos = i;
			}
			return n;
		}
		*pushUntil(test) {
			let i = this.pos;
			let ch = this.buffer[i];
			while (!test(ch)) ch = this.buffer[++i];
			return yield* this.pushToIndex(i, false);
		}
	};
	/**
	* Tracks newlines during parsing in order to provide an efficient API for
	* determining the one-indexed `{ line, col }` position for any offset
	* within the input.
	*/
	var LineCounter = class {
		constructor() {
			this.lineStarts = [];
			/**
			* Should be called in ascending order. Otherwise, call
			* `lineCounter.lineStarts.sort()` before calling `linePos()`.
			*/
			this.addNewLine = (offset) => this.lineStarts.push(offset);
			/**
			* Performs a binary search and returns the 1-indexed { line, col }
			* position of `offset`. If `line === 0`, `addNewLine` has never been
			* called or `offset` is before the first known newline.
			*/
			this.linePos = (offset) => {
				let low = 0;
				let high = this.lineStarts.length;
				while (low < high) {
					const mid = low + high >> 1;
					if (this.lineStarts[mid] < offset) low = mid + 1;
					else high = mid;
				}
				if (this.lineStarts[low] === offset) return {
					line: low + 1,
					col: 1
				};
				if (low === 0) return {
					line: 0,
					col: offset
				};
				const start = this.lineStarts[low - 1];
				return {
					line: low,
					col: offset - start + 1
				};
			};
		}
	};
	function includesToken(list, type) {
		for (let i = 0; i < list.length; ++i) if (list[i].type === type) return true;
		return false;
	}
	function findNonEmptyIndex(list) {
		for (let i = 0; i < list.length; ++i) switch (list[i].type) {
			case "space":
			case "comment":
			case "newline": break;
			default: return i;
		}
		return -1;
	}
	function isFlowToken(token) {
		switch (token?.type) {
			case "alias":
			case "scalar":
			case "single-quoted-scalar":
			case "double-quoted-scalar":
			case "flow-collection": return true;
			default: return false;
		}
	}
	function getPrevProps(parent) {
		switch (parent.type) {
			case "document": return parent.start;
			case "block-map": {
				const it = parent.items[parent.items.length - 1];
				return it.sep ?? it.start;
			}
			case "block-seq": return parent.items[parent.items.length - 1].start;
			/* istanbul ignore next should not happen */
			default: return [];
		}
	}
	/** Note: May modify input array */
	function getFirstKeyStartProps(prev) {
		if (prev.length === 0) return [];
		let i = prev.length;
		loop: while (--i >= 0) switch (prev[i].type) {
			case "doc-start":
			case "explicit-key-ind":
			case "map-value-ind":
			case "seq-item-ind":
			case "newline": break loop;
		}
		while (prev[++i]?.type === "space");
		return prev.splice(i, prev.length);
	}
	function arrayPushArray(target, source) {
		if (source.length < 1e5) Array.prototype.push.apply(target, source);
		else for (let i = 0; i < source.length; ++i) target.push(source[i]);
	}
	function fixFlowSeqItems(fc) {
		if (fc.start.type === "flow-seq-start") {
			for (const it of fc.items) if (it.sep && !it.value && !includesToken(it.start, "explicit-key-ind") && !includesToken(it.sep, "map-value-ind")) {
				if (it.key) it.value = it.key;
				delete it.key;
				if (isFlowToken(it.value)) {
					if (it.value.end) arrayPushArray(it.value.end, it.sep);
					else it.value.end = it.sep;
				} else arrayPushArray(it.start, it.sep);
				delete it.sep;
			}
		}
	}
	/**
	* A YAML concrete syntax tree (CST) parser
	*
	* ```ts
	* const src: string = ...
	* for (const token of new Parser().parse(src)) {
	*   // token: Token
	* }
	* ```
	*
	* To use the parser with a user-provided lexer:
	*
	* ```ts
	* function* parse(source: string, lexer: Lexer) {
	*   const parser = new Parser()
	*   for (const lexeme of lexer.lex(source))
	*     yield* parser.next(lexeme)
	*   yield* parser.end()
	* }
	*
	* const src: string = ...
	* const lexer = new Lexer()
	* for (const token of parse(src, lexer)) {
	*   // token: Token
	* }
	* ```
	*/
	var Parser = class {
		/**
		* @param onNewLine - If defined, called separately with the start position of
		*   each new line (in `parse()`, including the start of input).
		*/
		constructor(onNewLine) {
			/** If true, space and sequence indicators count as indentation */
			this.atNewLine = true;
			/** If true, next token is a scalar value */
			this.atScalar = false;
			/** Current indentation level */
			this.indent = 0;
			/** Current offset since the start of parsing */
			this.offset = 0;
			/** On the same line with a block map key */
			this.onKeyLine = false;
			/** Top indicates the node that's currently being built */
			this.stack = [];
			/** The source of the current token, set in parse() */
			this.source = "";
			/** The type of the current token, set in parse() */
			this.type = "";
			this.lexer = new Lexer();
			this.onNewLine = onNewLine;
		}
		/**
		* Parse `source` as a YAML stream.
		* If `incomplete`, a part of the last line may be left as a buffer for the next call.
		*
		* Errors are not thrown, but yielded as `{ type: 'error', message }` tokens.
		*
		* @returns A generator of tokens representing each directive, document, and other structure.
		*/
		*parse(source, incomplete = false) {
			if (this.onNewLine && this.offset === 0) this.onNewLine(0);
			for (const lexeme of this.lexer.lex(source, incomplete)) yield* this.next(lexeme);
			if (!incomplete) yield* this.end();
		}
		/**
		* Advance the parser by the `source` of one lexical token.
		*/
		*next(source) {
			this.source = source;
			if (this.atScalar) {
				this.atScalar = false;
				yield* this.step();
				this.offset += source.length;
				return;
			}
			const type = tokenType(source);
			if (!type) {
				const message = `Not a YAML token: ${source}`;
				yield* this.pop({
					type: "error",
					offset: this.offset,
					message,
					source
				});
				this.offset += source.length;
			} else if (type === "scalar") {
				this.atNewLine = false;
				this.atScalar = true;
				this.type = "scalar";
			} else {
				this.type = type;
				yield* this.step();
				switch (type) {
					case "newline":
						this.atNewLine = true;
						this.indent = 0;
						if (this.onNewLine) this.onNewLine(this.offset + source.length);
						break;
					case "space":
						if (this.atNewLine && source[0] === " ") this.indent += source.length;
						break;
					case "explicit-key-ind":
					case "map-value-ind":
					case "seq-item-ind":
						if (this.atNewLine) this.indent += source.length;
						break;
					case "doc-mode":
					case "flow-error-end": return;
					default: this.atNewLine = false;
				}
				this.offset += source.length;
			}
		}
		/** Call at end of input to push out any remaining constructions */
		*end() {
			while (this.stack.length > 0) yield* this.pop();
		}
		get sourceToken() {
			return {
				type: this.type,
				offset: this.offset,
				indent: this.indent,
				source: this.source
			};
		}
		*step() {
			const top = this.peek(1);
			if (this.type === "doc-end" && top?.type !== "doc-end") {
				while (this.stack.length > 0) yield* this.pop();
				this.stack.push({
					type: "doc-end",
					offset: this.offset,
					source: this.source
				});
				return;
			}
			if (!top) return yield* this.stream();
			switch (top.type) {
				case "document": return yield* this.document(top);
				case "alias":
				case "scalar":
				case "single-quoted-scalar":
				case "double-quoted-scalar": return yield* this.scalar(top);
				case "block-scalar": return yield* this.blockScalar(top);
				case "block-map": return yield* this.blockMap(top);
				case "block-seq": return yield* this.blockSequence(top);
				case "flow-collection": return yield* this.flowCollection(top);
				case "doc-end": return yield* this.documentEnd(top);
			}
			/* istanbul ignore next should not happen */
			yield* this.pop();
		}
		peek(n) {
			return this.stack[this.stack.length - n];
		}
		*pop(error) {
			const token = error ?? this.stack.pop();
			/* istanbul ignore if should not happen */
			if (!token) yield {
				type: "error",
				offset: this.offset,
				source: "",
				message: "Tried to pop an empty stack"
			};
			else if (this.stack.length === 0) yield token;
			else {
				const top = this.peek(1);
				if (token.type === "block-scalar") token.indent = "indent" in top ? top.indent : 0;
				else if (token.type === "flow-collection" && top.type === "document") token.indent = 0;
				if (token.type === "flow-collection") fixFlowSeqItems(token);
				switch (top.type) {
					case "document":
						top.value = token;
						break;
					case "block-scalar":
						top.props.push(token);
						break;
					case "block-map": {
						const it = top.items[top.items.length - 1];
						if (it.value) {
							top.items.push({
								start: [],
								key: token,
								sep: []
							});
							this.onKeyLine = true;
							return;
						} else if (it.sep) it.value = token;
						else {
							Object.assign(it, {
								key: token,
								sep: []
							});
							this.onKeyLine = !it.explicitKey;
							return;
						}
						break;
					}
					case "block-seq": {
						const it = top.items[top.items.length - 1];
						if (it.value) top.items.push({
							start: [],
							value: token
						});
						else it.value = token;
						break;
					}
					case "flow-collection": {
						const it = top.items[top.items.length - 1];
						if (!it || it.value) top.items.push({
							start: [],
							key: token,
							sep: []
						});
						else if (it.sep) it.value = token;
						else Object.assign(it, {
							key: token,
							sep: []
						});
						return;
					}
					/* istanbul ignore next should not happen */
					default:
						yield* this.pop();
						yield* this.pop(token);
				}
				if ((top.type === "document" || top.type === "block-map" || top.type === "block-seq") && (token.type === "block-map" || token.type === "block-seq")) {
					const last = token.items[token.items.length - 1];
					if (last && !last.sep && !last.value && last.start.length > 0 && findNonEmptyIndex(last.start) === -1 && (token.indent === 0 || last.start.every((st) => st.type !== "comment" || st.indent < token.indent))) {
						if (top.type === "document") top.end = last.start;
						else top.items.push({ start: last.start });
						token.items.splice(-1, 1);
					}
				}
			}
		}
		*stream() {
			switch (this.type) {
				case "directive-line":
					yield {
						type: "directive",
						offset: this.offset,
						source: this.source
					};
					return;
				case "byte-order-mark":
				case "space":
				case "comment":
				case "newline":
					yield this.sourceToken;
					return;
				case "doc-mode":
				case "doc-start": {
					const doc = {
						type: "document",
						offset: this.offset,
						start: []
					};
					if (this.type === "doc-start") doc.start.push(this.sourceToken);
					this.stack.push(doc);
					return;
				}
			}
			yield {
				type: "error",
				offset: this.offset,
				message: `Unexpected ${this.type} token in YAML stream`,
				source: this.source
			};
		}
		*document(doc) {
			if (doc.value) return yield* this.lineEnd(doc);
			switch (this.type) {
				case "doc-start":
					if (findNonEmptyIndex(doc.start) !== -1) {
						yield* this.pop();
						yield* this.step();
					} else doc.start.push(this.sourceToken);
					return;
				case "anchor":
				case "tag":
				case "space":
				case "comment":
				case "newline":
					doc.start.push(this.sourceToken);
					return;
			}
			const bv = this.startBlockValue(doc);
			if (bv) this.stack.push(bv);
			else yield {
				type: "error",
				offset: this.offset,
				message: `Unexpected ${this.type} token in YAML document`,
				source: this.source
			};
		}
		*scalar(scalar) {
			if (this.type === "map-value-ind") {
				const start = getFirstKeyStartProps(getPrevProps(this.peek(2)));
				let sep;
				if (scalar.end) {
					sep = scalar.end;
					sep.push(this.sourceToken);
					delete scalar.end;
				} else sep = [this.sourceToken];
				const map = {
					type: "block-map",
					offset: scalar.offset,
					indent: scalar.indent,
					items: [{
						start,
						key: scalar,
						sep
					}]
				};
				this.onKeyLine = true;
				this.stack[this.stack.length - 1] = map;
			} else yield* this.lineEnd(scalar);
		}
		*blockScalar(scalar) {
			switch (this.type) {
				case "space":
				case "comment":
				case "newline":
					scalar.props.push(this.sourceToken);
					return;
				case "scalar":
					scalar.source = this.source;
					this.atNewLine = true;
					this.indent = 0;
					if (this.onNewLine) {
						let nl = this.source.indexOf("\n") + 1;
						while (nl !== 0) {
							this.onNewLine(this.offset + nl);
							nl = this.source.indexOf("\n", nl) + 1;
						}
					}
					yield* this.pop();
					break;
				/* istanbul ignore next should not happen */
				default:
					yield* this.pop();
					yield* this.step();
			}
		}
		*blockMap(map) {
			const it = map.items[map.items.length - 1];
			switch (this.type) {
				case "newline":
					this.onKeyLine = false;
					if (it.value) {
						const end = "end" in it.value ? it.value.end : void 0;
						if ((Array.isArray(end) ? end[end.length - 1] : void 0)?.type === "comment") end?.push(this.sourceToken);
						else map.items.push({ start: [this.sourceToken] });
					} else if (it.sep) it.sep.push(this.sourceToken);
					else it.start.push(this.sourceToken);
					return;
				case "space":
				case "comment":
					if (it.value) map.items.push({ start: [this.sourceToken] });
					else if (it.sep) it.sep.push(this.sourceToken);
					else {
						if (this.atIndentedComment(it.start, map.indent)) {
							const end = map.items[map.items.length - 2]?.value?.end;
							if (Array.isArray(end)) {
								arrayPushArray(end, it.start);
								end.push(this.sourceToken);
								map.items.pop();
								return;
							}
						}
						it.start.push(this.sourceToken);
					}
					return;
			}
			if (this.indent >= map.indent) {
				const atMapIndent = !this.onKeyLine && this.indent === map.indent;
				const atNextItem = atMapIndent && (it.sep || it.explicitKey) && this.type !== "seq-item-ind";
				let start = [];
				if (atNextItem && it.sep && !it.value) {
					const nl = [];
					for (let i = 0; i < it.sep.length; ++i) {
						const st = it.sep[i];
						switch (st.type) {
							case "newline":
								nl.push(i);
								break;
							case "space": break;
							case "comment":
								if (st.indent > map.indent) nl.length = 0;
								break;
							default: nl.length = 0;
						}
					}
					if (nl.length >= 2) start = it.sep.splice(nl[1]);
				}
				switch (this.type) {
					case "anchor":
					case "tag":
						if (atNextItem || it.value) {
							start.push(this.sourceToken);
							map.items.push({ start });
							this.onKeyLine = true;
						} else if (it.sep) it.sep.push(this.sourceToken);
						else it.start.push(this.sourceToken);
						return;
					case "explicit-key-ind":
						if (!it.sep && !it.explicitKey) {
							it.start.push(this.sourceToken);
							it.explicitKey = true;
						} else if (atNextItem || it.value) {
							start.push(this.sourceToken);
							map.items.push({
								start,
								explicitKey: true
							});
						} else this.stack.push({
							type: "block-map",
							offset: this.offset,
							indent: this.indent,
							items: [{
								start: [this.sourceToken],
								explicitKey: true
							}]
						});
						this.onKeyLine = true;
						return;
					case "map-value-ind":
						if (it.explicitKey) {
							if (!it.sep) {
								if (includesToken(it.start, "newline")) Object.assign(it, {
									key: null,
									sep: [this.sourceToken]
								});
								else {
									const start = getFirstKeyStartProps(it.start);
									this.stack.push({
										type: "block-map",
										offset: this.offset,
										indent: this.indent,
										items: [{
											start,
											key: null,
											sep: [this.sourceToken]
										}]
									});
								}
							} else if (it.value) map.items.push({
								start: [],
								key: null,
								sep: [this.sourceToken]
							});
							else if (includesToken(it.sep, "map-value-ind")) this.stack.push({
								type: "block-map",
								offset: this.offset,
								indent: this.indent,
								items: [{
									start,
									key: null,
									sep: [this.sourceToken]
								}]
							});
							else if (isFlowToken(it.key) && !includesToken(it.sep, "newline")) {
								const start = getFirstKeyStartProps(it.start);
								const key = it.key;
								const sep = it.sep;
								sep.push(this.sourceToken);
								delete it.key;
								delete it.sep;
								this.stack.push({
									type: "block-map",
									offset: this.offset,
									indent: this.indent,
									items: [{
										start,
										key,
										sep
									}]
								});
							} else if (start.length > 0) it.sep = it.sep.concat(start, this.sourceToken);
							else it.sep.push(this.sourceToken);
						} else if (!it.sep) Object.assign(it, {
							key: null,
							sep: [this.sourceToken]
						});
						else if (it.value || atNextItem) map.items.push({
							start,
							key: null,
							sep: [this.sourceToken]
						});
						else if (includesToken(it.sep, "map-value-ind")) this.stack.push({
							type: "block-map",
							offset: this.offset,
							indent: this.indent,
							items: [{
								start: [],
								key: null,
								sep: [this.sourceToken]
							}]
						});
						else it.sep.push(this.sourceToken);
						this.onKeyLine = true;
						return;
					case "alias":
					case "scalar":
					case "single-quoted-scalar":
					case "double-quoted-scalar": {
						const fs = this.flowScalar(this.type);
						if (atNextItem || it.value) {
							map.items.push({
								start,
								key: fs,
								sep: []
							});
							this.onKeyLine = true;
						} else if (it.sep) this.stack.push(fs);
						else {
							Object.assign(it, {
								key: fs,
								sep: []
							});
							this.onKeyLine = true;
						}
						return;
					}
					default: {
						const bv = this.startBlockValue(map);
						if (bv) {
							if (bv.type === "block-seq") {
								if (!it.explicitKey && it.sep && !includesToken(it.sep, "newline")) {
									yield* this.pop({
										type: "error",
										offset: this.offset,
										message: "Unexpected block-seq-ind on same line with key",
										source: this.source
									});
									return;
								}
							} else if (atMapIndent) map.items.push({ start });
							this.stack.push(bv);
							return;
						}
					}
				}
			}
			yield* this.pop();
			yield* this.step();
		}
		*blockSequence(seq) {
			const it = seq.items[seq.items.length - 1];
			switch (this.type) {
				case "newline":
					if (it.value) {
						const end = "end" in it.value ? it.value.end : void 0;
						if ((Array.isArray(end) ? end[end.length - 1] : void 0)?.type === "comment") end?.push(this.sourceToken);
						else seq.items.push({ start: [this.sourceToken] });
					} else it.start.push(this.sourceToken);
					return;
				case "space":
				case "comment":
					if (it.value) seq.items.push({ start: [this.sourceToken] });
					else {
						if (this.atIndentedComment(it.start, seq.indent)) {
							const end = seq.items[seq.items.length - 2]?.value?.end;
							if (Array.isArray(end)) {
								arrayPushArray(end, it.start);
								end.push(this.sourceToken);
								seq.items.pop();
								return;
							}
						}
						it.start.push(this.sourceToken);
					}
					return;
				case "anchor":
				case "tag":
					if (it.value || this.indent <= seq.indent) break;
					it.start.push(this.sourceToken);
					return;
				case "seq-item-ind":
					if (this.indent !== seq.indent) break;
					if (it.value || includesToken(it.start, "seq-item-ind")) seq.items.push({ start: [this.sourceToken] });
					else it.start.push(this.sourceToken);
					return;
			}
			if (this.indent > seq.indent) {
				const bv = this.startBlockValue(seq);
				if (bv) {
					this.stack.push(bv);
					return;
				}
			}
			yield* this.pop();
			yield* this.step();
		}
		*flowCollection(fc) {
			const it = fc.items[fc.items.length - 1];
			if (this.type === "flow-error-end") {
				let top;
				do {
					yield* this.pop();
					top = this.peek(1);
				} while (top?.type === "flow-collection");
			} else if (fc.end.length === 0) {
				switch (this.type) {
					case "comma":
					case "explicit-key-ind":
						if (!it || it.sep) fc.items.push({ start: [this.sourceToken] });
						else it.start.push(this.sourceToken);
						return;
					case "map-value-ind":
						if (!it || it.value) fc.items.push({
							start: [],
							key: null,
							sep: [this.sourceToken]
						});
						else if (it.sep) it.sep.push(this.sourceToken);
						else Object.assign(it, {
							key: null,
							sep: [this.sourceToken]
						});
						return;
					case "space":
					case "comment":
					case "newline":
					case "anchor":
					case "tag":
						if (!it || it.value) fc.items.push({ start: [this.sourceToken] });
						else if (it.sep) it.sep.push(this.sourceToken);
						else it.start.push(this.sourceToken);
						return;
					case "alias":
					case "scalar":
					case "single-quoted-scalar":
					case "double-quoted-scalar": {
						const fs = this.flowScalar(this.type);
						if (!it || it.value) fc.items.push({
							start: [],
							key: fs,
							sep: []
						});
						else if (it.sep) this.stack.push(fs);
						else Object.assign(it, {
							key: fs,
							sep: []
						});
						return;
					}
					case "flow-map-end":
					case "flow-seq-end":
						fc.end.push(this.sourceToken);
						return;
				}
				const bv = this.startBlockValue(fc);
				/* istanbul ignore else should not happen */
				if (bv) this.stack.push(bv);
				else {
					yield* this.pop();
					yield* this.step();
				}
			} else {
				const parent = this.peek(2);
				if (parent.type === "block-map" && (this.type === "map-value-ind" && parent.indent === fc.indent || this.type === "newline" && !parent.items[parent.items.length - 1].sep)) {
					yield* this.pop();
					yield* this.step();
				} else if (this.type === "map-value-ind" && parent.type !== "flow-collection") {
					const start = getFirstKeyStartProps(getPrevProps(parent));
					fixFlowSeqItems(fc);
					const sep = fc.end.splice(1, fc.end.length);
					sep.push(this.sourceToken);
					const map = {
						type: "block-map",
						offset: fc.offset,
						indent: fc.indent,
						items: [{
							start,
							key: fc,
							sep
						}]
					};
					this.onKeyLine = true;
					this.stack[this.stack.length - 1] = map;
				} else yield* this.lineEnd(fc);
			}
		}
		flowScalar(type) {
			if (this.onNewLine) {
				let nl = this.source.indexOf("\n") + 1;
				while (nl !== 0) {
					this.onNewLine(this.offset + nl);
					nl = this.source.indexOf("\n", nl) + 1;
				}
			}
			return {
				type,
				offset: this.offset,
				indent: this.indent,
				source: this.source
			};
		}
		startBlockValue(parent) {
			switch (this.type) {
				case "alias":
				case "scalar":
				case "single-quoted-scalar":
				case "double-quoted-scalar": return this.flowScalar(this.type);
				case "block-scalar-header": return {
					type: "block-scalar",
					offset: this.offset,
					indent: this.indent,
					props: [this.sourceToken],
					source: ""
				};
				case "flow-map-start":
				case "flow-seq-start": return {
					type: "flow-collection",
					offset: this.offset,
					indent: this.indent,
					start: this.sourceToken,
					items: [],
					end: []
				};
				case "seq-item-ind": return {
					type: "block-seq",
					offset: this.offset,
					indent: this.indent,
					items: [{ start: [this.sourceToken] }]
				};
				case "explicit-key-ind": {
					this.onKeyLine = true;
					const start = getFirstKeyStartProps(getPrevProps(parent));
					start.push(this.sourceToken);
					return {
						type: "block-map",
						offset: this.offset,
						indent: this.indent,
						items: [{
							start,
							explicitKey: true
						}]
					};
				}
				case "map-value-ind": {
					this.onKeyLine = true;
					const start = getFirstKeyStartProps(getPrevProps(parent));
					return {
						type: "block-map",
						offset: this.offset,
						indent: this.indent,
						items: [{
							start,
							key: null,
							sep: [this.sourceToken]
						}]
					};
				}
			}
			return null;
		}
		atIndentedComment(start, indent) {
			if (this.type !== "comment") return false;
			if (this.indent <= indent) return false;
			return start.every((st) => st.type === "newline" || st.type === "space");
		}
		*documentEnd(docEnd) {
			if (this.type !== "doc-mode") {
				if (docEnd.end) docEnd.end.push(this.sourceToken);
				else docEnd.end = [this.sourceToken];
				if (this.type === "newline") yield* this.pop();
			}
		}
		*lineEnd(token) {
			switch (this.type) {
				case "comma":
				case "doc-start":
				case "doc-end":
				case "flow-seq-end":
				case "flow-map-end":
				case "map-value-ind":
					yield* this.pop();
					yield* this.step();
					break;
				case "newline": this.onKeyLine = false;
				default:
					if (token.end) token.end.push(this.sourceToken);
					else token.end = [this.sourceToken];
					if (this.type === "newline") yield* this.pop();
			}
		}
	};
	function parseOptions(options) {
		const prettyErrors = options.prettyErrors !== false;
		return {
			lineCounter: options.lineCounter || prettyErrors && new LineCounter() || null,
			prettyErrors
		};
	}
	/** Parse an input string into a single YAML.Document */
	function parseDocument(source, options = {}) {
		const { lineCounter, prettyErrors } = parseOptions(options);
		const parser = new Parser(lineCounter?.addNewLine);
		const composer = new Composer(options);
		let doc = null;
		for (const _doc of composer.compose(parser.parse(source), true, source.length)) if (!doc) doc = _doc;
		else if (doc.options.logLevel !== "silent") {
			doc.errors.push(new YAMLParseError(_doc.range.slice(0, 2), "MULTIPLE_DOCS", "Source contains multiple documents; please use YAML.parseAllDocuments()"));
			break;
		}
		if (prettyErrors && lineCounter) {
			doc.errors.forEach(prettifyError(source, lineCounter));
			doc.warnings.forEach(prettifyError(source, lineCounter));
		}
		return doc;
	}
	var STR_TAG = "tag:yaml.org,2002:str";
	var ORIGINAL = Symbol("original value");
	function sourceOf(item, ctx) {
		const tok = item.srcToken;
		if (!tok || item[ORIGINAL] !== item.value || item.comment || ctx.implicitKey) return null;
		if (tok.type === "block-scalar") {
			if (tok.props.some((p) => p.type !== "block-scalar-header" && p.type !== "newline")) return null;
			const header = tok.props.find((p) => p.type === "block-scalar-header")?.source ?? "";
			if (!header || /[0-9+]/.test(header)) return null;
			const lines = tok.source.replace(/\n+$/, "").split("\n");
			const indents = lines.filter((l) => l.trim()).map((l) => l.match(/^ */)[0].length);
			if (!indents.length) return null;
			const min = Math.min(...indents);
			const body = lines.map((l) => l.trim() ? ctx.indent + l.slice(min) : "").join("\n");
			return header + "\n" + body;
		}
		if (tok.type === "scalar" || tok.type === "single-quoted-scalar" || tok.type === "double-quoted-scalar") {
			if (tok.source.includes("\n")) return null;
			if (tok.type === "scalar" && item.type !== "PLAIN") return null;
			return tok.source;
		}
		return null;
	}
	function preserveSources(doc) {
		visit$1(doc, { Scalar(_, node) {
			node[ORIGINAL] = node.value;
		} });
		doc.schema.tags = doc.schema.tags.map((tag) => {
			if (tag.tag !== STR_TAG || !tag.stringify) return tag;
			const base = tag.stringify;
			return {
				...tag,
				stringify: (item, ctx, onComment, onChompKeep) => (isScalar(item) ? sourceOf(item, ctx) : null) ?? base(item, ctx, onComment, onChompKeep)
			};
		});
	}
	function detectIndent(text) {
		const m = text.match(/^( +)\S/m);
		return m ? m[1].length : 2;
	}
	var PipelineDoc = class PipelineDoc {
		doc;
		source;
		options;
		constructor(doc, source, options) {
			this.doc = doc;
			this.source = source;
			this.options = options;
		}
		static parse(text) {
			const doc = parseDocument(text, { keepSourceTokens: true });
			preserveSources(doc);
			return new PipelineDoc(doc, text, {
				flowCollectionPadding: false,
				indent: detectIndent(text)
			});
		}
		get errors() {
			return this.doc.errors.map((e) => e.message.split("\n")[0]);
		}
		get syntaxError() {
			const e = this.doc.errors[0];
			if (!e) return null;
			const pos = e.linePos?.[0];
			return {
				message: e.message.split("\n")[0].replace(/ at line \d+, column \d+:?$/, ""),
				line: pos?.line,
				col: pos?.col
			};
		}
		toString() {
			if (this.doc.contents === null) return this.source;
			return this.doc.toString(this.options);
		}
		toGraph() {
			const js = this.doc.toJS();
			const root = js && typeof js === "object" && !Array.isArray(js) ? js : {};
			const nodes = (Array.isArray(root.nodes) ? root.nodes : []).filter((n) => !!n && typeof n === "object" && !Array.isArray(n)).map((n) => ({
				...n,
				id: n.id === void 0 || n.id === null ? "" : String(n.id)
			}));
			const defaults = root.defaults && typeof root.defaults === "object" && !Array.isArray(root.defaults) ? root.defaults : void 0;
			const str = (v) => v === void 0 || v === null ? void 0 : String(v);
			return {
				name: str(root.name),
				goal: str(root.goal),
				start: str(root.start),
				defaults,
				nodes
			};
		}
		get(path) {
			return this.doc.getIn(path);
		}
		has(path) {
			return this.doc.hasIn(path);
		}
		set(path, value) {
			if (value === void 0) {
				this.doc.deleteIn(path);
				return;
			}
			const existing = this.doc.getIn(path, true);
			if (isScalar(existing) && (typeof value === "string" || typeof value === "number" || typeof value === "boolean")) {
				if (typeof existing.value === typeof value) {
					existing.value = value;
					return;
				}
				const created = this.doc.createNode(value);
				created.comment = existing.comment;
				created.commentBefore = existing.commentBefore;
				created.spaceBefore = existing.spaceBefore;
				this.doc.setIn(path, created);
				return;
			}
			const flow = isSeq(existing) && existing.flow;
			this.doc.setIn(path, this.doc.createNode(value));
			if (flow && Array.isArray(value)) {
				const created = this.doc.getIn(path, true);
				if (isSeq(created)) created.flow = true;
			}
		}
		delete(path) {
			this.doc.deleteIn(path);
		}
		push(path, value) {
			if (isCollection(this.doc.getIn(path, true))) this.doc.addIn(path, this.doc.createNode(value));
			else this.doc.setIn(path, this.doc.createNode([value]));
		}
		deleteAndPrune(path) {
			this.doc.deleteIn(path);
			const parent = path.slice(0, -1);
			if (!parent.length) return;
			const node = this.doc.getIn(parent, true);
			if (isCollection(node) && node.items.length === 0) this.doc.deleteIn(parent);
		}
	};
	var SAVE_DELAY = 400;
	var RETRY_DELAY = 3e3;
	var PLACE_STEP = 40;
	var DARK_KEY = "gimble-editor-dark";
	var SNAP_KEY = "gimble-editor-snap";
	var ROUTE_KEYS = [
		"success",
		"error",
		"loop",
		"exit"
	];
	function readFlag(key) {
		try {
			return localStorage.getItem(key) === "1";
		} catch {
			return false;
		}
	}
	function writeFlag(key, value) {
		try {
			localStorage.setItem(key, value ? "1" : "0");
		} catch {}
	}
	var Editor = class {
		path = "";
		version = "";
		serverDiagnostics = [];
		models = [];
		parseError = "";
		banner = "";
		loading = true;
		saving = false;
		needsFit = false;
		doc = PipelineDoc.parse("");
		graphValue = this.doc.toGraph();
		yamlErrorsValue = this.doc.errors;
		syntaxErrorValue = this.doc.syntaxError;
		get graph() {
			return this.graphValue;
		}
		get yamlErrors() {
			return this.yamlErrorsValue;
		}
		get syntaxError() {
			return this.syntaxErrorValue;
		}
		get readOnly() {
			return this.syntaxError !== null;
		}
		savedLayout = {};
		layoutFromDisk = false;
		layoutTouched = false;
		get layout() {
			return autoLayout(this.graph, this.savedLayout);
		}
		sel = null;
		pan = {
			x: 40,
			y: 60
		};
		zoom = .85;
		vw = 1e3;
		vh = 700;
		drag = null;
		dark = readFlag(DARK_KEY);
		snap = readFlag(SNAP_KEY);
		showInherited = true;
		get clientErrors() {
			return validate(this.graph);
		}
		diagnosticOwner(d) {
			const id = d.node_id || d.edge?.[0];
			if (!id) return null;
			const g = this.graph;
			if (g.nodes.some((n) => n.id === id)) return id;
			const owner = g.nodes.find((n) => n.type === "fan_out" && (n.branches ?? []).some((b) => normalizeBranch(b).id === id));
			return owner ? owner.id : null;
		}
		get errors() {
			const out = {};
			for (const [id, list] of Object.entries(this.clientErrors)) out[id] = [...list];
			for (const d of this.serverDiagnostics) {
				const id = this.diagnosticOwner(d);
				if (!id) continue;
				const msg = d.severity === "error" ? d.message : `${d.severity}: ${d.message}`;
				const list = out[id] ??= [];
				if (!list.includes(msg)) list.push(msg);
			}
			return out;
		}
		get graphIssues() {
			const list = [];
			if (this.parseError) list.push(this.parseError);
			for (const e of this.yamlErrors) list.push(e);
			const g = this.graph;
			const ids = g.nodes.map((n) => n.id);
			if (!g.start) list.push("Start is required");
			else if (!ids.includes(g.start)) list.push("Start node does not exist");
			for (const d of this.serverDiagnostics) {
				if (this.diagnosticOwner(d)) continue;
				const where = d.node_id || d.edge?.[0];
				const text = where ? `${where}: ${d.message}` : d.message;
				const msg = d.severity === "error" ? text : `${d.severity}: ${text}`;
				if (!list.includes(msg)) list.push(msg);
			}
			return list;
		}
		get errorCount() {
			return Object.values(this.errors).reduce((a, b) => a + b.length, 0) + this.graphIssues.length;
		}
		get selNode() {
			return this.sel ? this.graph.nodes.find((n) => n.id === this.sel) ?? null : null;
		}
		timer = null;
		inflight = false;
		dirty = false;
		saveFailed = false;
		edits = 0;
		announced = "";
		loadInflight = false;
		changePending = false;
		constructor(initial) {
			this.adopt(initial);
			this.needsFit = true;
		}
		get pending() {
			return this.timer !== null || this.inflight || this.dirty || this.saveFailed;
		}
		start() {
			const stop = subscribe((version) => this.onChange(version));
			return () => {
				stop();
				if (this.timer) clearTimeout(this.timer);
			};
		}
		onChange(version) {
			this.announced = version;
			if (version === this.version) return;
			if (this.loadInflight || this.inflight) {
				this.changePending = true;
				return;
			}
			if (!this.pending) this.load();
		}
		settle() {
			if (!this.changePending || this.pending) return;
			this.changePending = false;
			if (this.announced !== this.version) this.load();
		}
		async load(recheck = true) {
			const edits = this.edits;
			this.loadInflight = true;
			try {
				const d = await getDoc();
				if (this.edits !== edits) return;
				const first = this.loading;
				this.adopt(d);
				if (first) this.needsFit = true;
			} catch (e) {
				this.banner = `Could not load the document: ${e instanceof Error ? e.message : String(e)}`;
				this.loading = false;
			} finally {
				this.loadInflight = false;
				if (recheck && this.announced && this.announced !== this.version && !this.pending) {
					this.changePending = false;
					this.load(false);
				} else this.settle();
			}
		}
		adopt(d) {
			this.doc = PipelineDoc.parse(d.yaml);
			this.refreshDocument();
			this.savedLayout = d.layout ?? {};
			this.layoutFromDisk = d.layout !== null;
			this.layoutTouched = false;
			this.path = d.path;
			this.version = d.version;
			this.serverDiagnostics = d.diagnostics ?? [];
			this.parseError = d.parse_error ?? "";
			this.models = d.models ?? [];
			if (this.sel && !isTerminal(this.sel) && !this.graph.nodes.some((n) => n.id === this.sel)) this.sel = null;
			this.loading = false;
		}
		touch() {
			this.refreshDocument();
			this.edits++;
			this.scheduleSave();
		}
		refreshDocument() {
			this.graphValue = this.doc.toGraph();
			this.yamlErrorsValue = this.doc.errors;
			this.syntaxErrorValue = this.doc.syntaxError;
		}
		scheduleSave(delay = SAVE_DELAY) {
			if (this.timer) clearTimeout(this.timer);
			this.timer = setTimeout(() => {
				this.timer = null;
				this.flush();
			}, delay);
		}
		async flush() {
			if (this.inflight) {
				this.dirty = true;
				return;
			}
			this.inflight = true;
			this.saving = true;
			const layout = this.layoutFromDisk || this.layoutTouched ? snapshot(this.savedLayout) : null;
			let retry = false;
			try {
				const d = await putDoc({
					yaml: this.doc.toString(),
					layout,
					version: this.version
				});
				this.path = d.path;
				this.version = d.version;
				this.serverDiagnostics = d.diagnostics ?? [];
				this.parseError = d.parse_error ?? "";
				this.models = d.models ?? [];
				if (d.layout !== null) this.layoutFromDisk = true;
				if (this.saveFailed) {
					this.saveFailed = false;
					this.banner = "";
				}
			} catch (e) {
				if (e instanceof ConflictError) {
					if (this.timer) clearTimeout(this.timer);
					this.timer = null;
					this.dirty = false;
					this.saveFailed = false;
					this.adopt(e.current);
					this.banner = "The file changed on disk; your unsaved edits were dropped.";
				} else {
					this.saveFailed = true;
					retry = true;
					this.banner = `Save failed: ${e instanceof Error ? e.message : String(e)}`;
				}
			} finally {
				this.inflight = false;
				this.saving = false;
				if (this.dirty) {
					this.dirty = false;
					this.scheduleSave();
				} else if (retry) this.scheduleSave(RETRY_DELAY);
				this.settle();
			}
		}
		index(id) {
			return this.graph.nodes.findIndex((n) => n.id === id);
		}
		setGraphField(key, value) {
			if (this.readOnly) return;
			if (key === "start") this.doc.set(["start"], value);
			else this.doc.set([key], value === "" ? void 0 : value);
			this.touch();
		}
		setDefault(key, value) {
			if (this.readOnly) return;
			const path = ["defaults", ...key.split(".")];
			if (value === void 0 && key === "model.name") this.doc.deleteAndPrune(["defaults", "model"]);
			else if (value === void 0) this.doc.deleteAndPrune(path);
			else this.doc.set(path, value);
			this.touch();
		}
		setNodeField(id, key, value) {
			if (this.readOnly) return;
			const i = this.index(id);
			if (i < 0) return;
			const parts = key.split(".");
			if (value === void 0 && parts.at(-1) === "name" && parts.at(-2) === "model") this.doc.deleteAndPrune([
				"nodes",
				i,
				...parts.slice(0, -1)
			]);
			else this.doc.set([
				"nodes",
				i,
				...parts
			], value);
			this.touch();
		}
		setEdge(id, index, patch) {
			if (this.readOnly) return;
			const i = this.index(id);
			const n = this.graph.nodes[i];
			if (!n) return;
			const key = edgesKey(n);
			if ("to" in patch) this.doc.set([
				"nodes",
				i,
				key,
				index,
				"to"
			], patch.to ?? "");
			if ("condition" in patch) this.doc.set([
				"nodes",
				i,
				key,
				index,
				"condition"
			], patch.condition || void 0);
			this.touch();
		}
		addEdge(id) {
			if (this.readOnly) return;
			const i = this.index(id);
			const n = this.graph.nodes[i];
			if (!n) return;
			this.doc.push([
				"nodes",
				i,
				edgesKey(n)
			], { to: "" });
			this.touch();
		}
		removeEdge(id, index) {
			if (this.readOnly) return;
			const i = this.index(id);
			const n = this.graph.nodes[i];
			if (!n) return;
			this.doc.delete([
				"nodes",
				i,
				edgesKey(n),
				index
			]);
			this.touch();
		}
		setSupervises(id, list) {
			this.setNodeField(id, "supervises", list);
		}
		branchesOf(id) {
			return (this.graph.nodes.find((x) => x.id === id)?.branches ?? []).map(normalizeBranch);
		}
		setBranch(id, index, patch) {
			if (this.readOnly) return;
			const i = this.index(id);
			const n = this.graph.nodes[i];
			if (!n) return;
			const branches = n.branches ?? [];
			const base = [
				"nodes",
				i,
				"branches",
				index
			];
			if (typeof branches[index] === "string") this.doc.set(base, {
				id: branches[index],
				artifacts: []
			});
			if (patch.id !== void 0) this.doc.set([...base, "id"], patch.id);
			if (patch.artifacts !== void 0) this.doc.set([...base, "artifacts"], patch.artifacts);
			if (patch.prompt !== void 0) {
				if (patch.prompt === "") this.doc.deleteAndPrune([
					...base,
					"agent",
					"prompt"
				]);
				else this.doc.set([
					...base,
					"agent",
					"prompt"
				], patch.prompt);
			}
			this.touch();
		}
		setBranchModel(id, index, key, value) {
			if (this.readOnly) return;
			const i = this.index(id);
			if (i < 0) return;
			const base = [
				"nodes",
				i,
				"branches",
				index
			];
			const branch = this.graph.nodes[i]?.branches?.[index];
			if (typeof branch === "string") this.doc.set(base, {
				id: branch,
				artifacts: []
			});
			if (value === void 0 && key === "name") this.doc.deleteAndPrune([
				...base,
				"agent",
				"model"
			]);
			else if (value === void 0) this.doc.deleteAndPrune([
				...base,
				"agent",
				"model",
				key
			]);
			else this.doc.set([
				...base,
				"agent",
				"model",
				key
			], value);
			this.touch();
		}
		addBranch(id) {
			if (this.readOnly) return;
			const i = this.index(id);
			if (i < 0) return;
			this.doc.push([
				"nodes",
				i,
				"branches"
			], {
				id: "",
				artifacts: []
			});
			this.touch();
		}
		removeBranch(id, index) {
			if (this.readOnly) return;
			const i = this.index(id);
			if (i < 0) return;
			this.doc.delete([
				"nodes",
				i,
				"branches",
				index
			]);
			this.touch();
		}
		setStart(id, on) {
			if (this.readOnly) return;
			const g = this.graph;
			if (on) this.doc.set(["start"], id);
			else if (g.start === id) this.doc.set(["start"], "");
			this.touch();
		}
		renameNode(oldId, newId) {
			if (this.readOnly) return;
			newId = newId.trim();
			if (!newId || newId === oldId) return;
			const g = this.graph;
			const i = this.index(oldId);
			if (i < 0) return;
			this.doc.set([
				"nodes",
				i,
				"id"
			], newId);
			g.nodes.forEach((n, j) => {
				if (n.type === "command" || n.type === "loop") {
					const e = routeEdges(n);
					for (const k of ROUTE_KEYS) if (e[k] === oldId) this.doc.set([
						"nodes",
						j,
						"edges",
						k
					], newId);
				} else if (n.type !== "supervisor") listEdges(n).forEach((e, k) => {
					if (e.to === oldId) this.doc.set([
						"nodes",
						j,
						edgesKey(n),
						k,
						"to"
					], newId);
				});
				if (n.type === "supervisor") (n.supervises ?? []).forEach((s, k) => {
					if (s === oldId) this.doc.set([
						"nodes",
						j,
						"supervises",
						k
					], newId);
				});
			});
			if (g.start === oldId) this.doc.set(["start"], newId);
			const layout = {};
			const prefix = `e:${oldId}>`;
			for (const [k, v] of Object.entries(this.savedLayout)) {
				const nk = k === oldId ? newId : k.startsWith(prefix) ? `e:${newId}>${k.slice(prefix.length)}` : k;
				layout[nk] = v;
			}
			this.savedLayout = layout;
			this.layoutTouched = true;
			if (this.sel === oldId) this.sel = newId;
			this.touch();
		}
		deleteNode(id) {
			if (this.readOnly) return;
			if (!id || isTerminal(id)) return;
			const g = this.graph;
			const i = this.index(id);
			if (i < 0) return;
			g.nodes.forEach((n, j) => {
				if (j === i) return;
				if (n.type === "command" || n.type === "loop") {
					const e = routeEdges(n);
					for (const k of ROUTE_KEYS) if (e[k] === id) this.doc.set([
						"nodes",
						j,
						"edges",
						k
					], k === "error" ? void 0 : "");
				} else if (n.type !== "supervisor") {
					const edges = listEdges(n);
					if (edges.some((e) => e.to === id)) this.doc.set([
						"nodes",
						j,
						edgesKey(n)
					], edges.filter((e) => e.to !== id));
				}
				if (n.type === "supervisor" && (n.supervises ?? []).includes(id)) this.doc.set([
					"nodes",
					j,
					"supervises"
				], (n.supervises ?? []).filter((x) => x !== id));
			});
			if (g.start === id) this.doc.set(["start"], "");
			this.doc.delete(["nodes", i]);
			const layout = {};
			const prefix = `e:${id}>`;
			for (const [k, v] of Object.entries(this.savedLayout)) if (k !== id && !k.startsWith(prefix)) layout[k] = v;
			this.savedLayout = layout;
			this.layoutTouched = true;
			if (this.sel === id) this.sel = null;
			this.touch();
		}
		addNode(type) {
			if (this.readOnly) return;
			const ids = this.graph.nodes.map((n) => n.id);
			let i = 1;
			let id = type;
			while (ids.includes(id)) id = `${type}_${++i}`;
			const n = {
				id,
				type
			};
			switch (type) {
				case "command":
					Object.assign(n, {
						command: "",
						edges: { success: "" }
					});
					break;
				case "supervisor":
					Object.assign(n, {
						prompt: "",
						supervises: []
					});
					break;
				case "loop":
					Object.assign(n, { edges: {
						loop: "",
						exit: ""
					} });
					break;
				case "fan_out":
					Object.assign(n, {
						branches: [],
						branch_edges: []
					});
					break;
				default: Object.assign(n, { edges: [] });
			}
			this.doc.push(["nodes"], n);
			const { x, y } = this.freeSlot((this.vw / 2 - this.pan.x) / this.zoom - 120, (this.vh / 2 - this.pan.y) / this.zoom - 60);
			this.savedLayout[id] = {
				x,
				y,
				w: 240,
				h: 120
			};
			this.layoutTouched = true;
			this.sel = id;
			this.touch();
		}
		freeSlot(x0, y0) {
			const taken = Object.values(this.layout).filter(isBox);
			const clear = (b) => !taken.some((l) => b.x < l.x + l.w && b.x + b.w > l.x && b.y < l.y + l.h && b.y + b.h > l.y);
			const cols = Math.max(1, Math.ceil(this.vw / this.zoom / 2 / PLACE_STEP));
			for (let row = 0; row < 1e3; row++) for (let col = 0; col < cols; col++) {
				const b = {
					x: x0 + col * PLACE_STEP,
					y: y0 + row * PLACE_STEP,
					w: 240,
					h: 120
				};
				if (clear(b)) return b;
			}
			return {
				x: x0,
				y: y0
			};
		}
		setLayout(key, entry) {
			if (this.readOnly) return;
			this.savedLayout[key] = entry;
			this.layoutTouched = true;
			this.edits++;
			this.scheduleSave();
		}
		fitView() {
			const ls = Object.values(this.layout).filter(isBox);
			if (!ls.length) return;
			const x0 = Math.min(...ls.map((l) => l.x)) - 140;
			const y0 = Math.min(...ls.map((l) => l.y)) - 40;
			const x1 = Math.max(...ls.map((l) => l.x + l.w)) + 140;
			const y1 = Math.max(...ls.map((l) => l.y + l.h)) + 40;
			const zoom = Math.min(1.5, this.vw / (x1 - x0), this.vh / (y1 - y0));
			this.zoom = zoom;
			this.pan = {
				x: (this.vw - (x1 - x0) * zoom) / 2 - x0 * zoom,
				y: (this.vh - (y1 - y0) * zoom) / 2 - y0 * zoom
			};
		}
		zoomBy(f) {
			const nz = Math.min(2.5, Math.max(.2, this.zoom * f));
			this.pan = {
				x: this.vw / 2 - (this.vw / 2 - this.pan.x) * nz / this.zoom,
				y: this.vh / 2 - (this.vh / 2 - this.pan.y) * nz / this.zoom
			};
			this.zoom = nz;
		}
		zoomAt(mx, my, nz) {
			nz = Math.min(2.5, Math.max(.2, nz));
			this.pan = {
				x: mx - (mx - this.pan.x) * nz / this.zoom,
				y: my - (my - this.pan.y) * nz / this.zoom
			};
			this.zoom = nz;
		}
		jumpTo(id) {
			const L = this.layout[id];
			if (!isBox(L)) return;
			this.sel = id;
			this.pan = {
				x: this.vw / 2 - (L.x + L.w / 2) * this.zoom,
				y: this.vh / 2 - (L.y + L.h / 2) * this.zoom
			};
		}
		toggleDark() {
			this.dark = !this.dark;
			writeFlag(DARK_KEY, this.dark);
		}
		setSnap(on) {
			this.snap = on;
			writeFlag(SNAP_KEY, on);
		}
		terminalIds() {
			return TERMINAL;
		}
	};
	function _page($$renderer, $$props) {
		$$renderer.component(($$renderer) => {
			let { data } = $$props;
			const initialDocument = () => fromWire(data);
			const editor = new Editor(initialDocument());
			$$renderer.push(`<div class="root svelte-1uha8ag">`);
			Header($$renderer, { editor });
			$$renderer.push(`<!----> <div class="body svelte-1uha8ag">`);
			Canvas($$renderer, { editor });
			$$renderer.push(`<!----> `);
			Inspector($$renderer, { editor });
			$$renderer.push(`<!----></div></div>`);
		});
	}
	var components = [
		_layout,
		Error$1,
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
