/* @ds-bundle: {"format":4,"namespace":"ShadcnSvelteDesignSystem_ce4a1c","components":[{"name":"Badge","sourcePath":"components/actions/Badge.jsx"},{"name":"Button","sourcePath":"components/actions/Button.jsx"},{"name":"Kbd","sourcePath":"components/actions/Kbd.jsx"},{"name":"KbdGroup","sourcePath":"components/actions/Kbd.jsx"},{"name":"Spinner","sourcePath":"components/actions/Spinner.jsx"},{"name":"Toggle","sourcePath":"components/actions/Toggle.jsx"},{"name":"Avatar","sourcePath":"components/data-display/Avatar.jsx"},{"name":"Card","sourcePath":"components/data-display/Card.jsx"},{"name":"CardHeader","sourcePath":"components/data-display/Card.jsx"},{"name":"CardTitle","sourcePath":"components/data-display/Card.jsx"},{"name":"CardDescription","sourcePath":"components/data-display/Card.jsx"},{"name":"CardContent","sourcePath":"components/data-display/Card.jsx"},{"name":"CardFooter","sourcePath":"components/data-display/Card.jsx"},{"name":"Progress","sourcePath":"components/data-display/Progress.jsx"},{"name":"Separator","sourcePath":"components/data-display/Separator.jsx"},{"name":"Skeleton","sourcePath":"components/data-display/Skeleton.jsx"},{"name":"Table","sourcePath":"components/data-display/Table.jsx"},{"name":"TableHeader","sourcePath":"components/data-display/Table.jsx"},{"name":"TableBody","sourcePath":"components/data-display/Table.jsx"},{"name":"TableFooter","sourcePath":"components/data-display/Table.jsx"},{"name":"TableRow","sourcePath":"components/data-display/Table.jsx"},{"name":"TableHead","sourcePath":"components/data-display/Table.jsx"},{"name":"TableCell","sourcePath":"components/data-display/Table.jsx"},{"name":"TableCaption","sourcePath":"components/data-display/Table.jsx"},{"name":"Alert","sourcePath":"components/feedback/Alert.jsx"},{"name":"AlertTitle","sourcePath":"components/feedback/Alert.jsx"},{"name":"AlertDescription","sourcePath":"components/feedback/Alert.jsx"},{"name":"Dialog","sourcePath":"components/feedback/Dialog.jsx"},{"name":"DialogHeader","sourcePath":"components/feedback/Dialog.jsx"},{"name":"DialogTitle","sourcePath":"components/feedback/Dialog.jsx"},{"name":"DialogDescription","sourcePath":"components/feedback/Dialog.jsx"},{"name":"DialogFooter","sourcePath":"components/feedback/Dialog.jsx"},{"name":"Tooltip","sourcePath":"components/feedback/Tooltip.jsx"},{"name":"Checkbox","sourcePath":"components/forms/Checkbox.jsx"},{"name":"Input","sourcePath":"components/forms/Input.jsx"},{"name":"Label","sourcePath":"components/forms/Label.jsx"},{"name":"RadioGroup","sourcePath":"components/forms/RadioGroup.jsx"},{"name":"RadioGroupItem","sourcePath":"components/forms/RadioGroup.jsx"},{"name":"Select","sourcePath":"components/forms/Select.jsx"},{"name":"Slider","sourcePath":"components/forms/Slider.jsx"},{"name":"Switch","sourcePath":"components/forms/Switch.jsx"},{"name":"Textarea","sourcePath":"components/forms/Textarea.jsx"},{"name":"ICON_PATHS","sourcePath":"components/icon/Icon.jsx"},{"name":"Icon","sourcePath":"components/icon/Icon.jsx"},{"name":"Accordion","sourcePath":"components/navigation/Accordion.jsx"},{"name":"AccordionItem","sourcePath":"components/navigation/Accordion.jsx"},{"name":"AccordionTrigger","sourcePath":"components/navigation/Accordion.jsx"},{"name":"AccordionContent","sourcePath":"components/navigation/Accordion.jsx"},{"name":"Breadcrumb","sourcePath":"components/navigation/Breadcrumb.jsx"},{"name":"BreadcrumbList","sourcePath":"components/navigation/Breadcrumb.jsx"},{"name":"BreadcrumbItem","sourcePath":"components/navigation/Breadcrumb.jsx"},{"name":"BreadcrumbLink","sourcePath":"components/navigation/Breadcrumb.jsx"},{"name":"BreadcrumbPage","sourcePath":"components/navigation/Breadcrumb.jsx"},{"name":"BreadcrumbSeparator","sourcePath":"components/navigation/Breadcrumb.jsx"},{"name":"DropdownMenu","sourcePath":"components/navigation/DropdownMenu.jsx"},{"name":"DropdownMenuTrigger","sourcePath":"components/navigation/DropdownMenu.jsx"},{"name":"DropdownMenuContent","sourcePath":"components/navigation/DropdownMenu.jsx"},{"name":"DropdownMenuItem","sourcePath":"components/navigation/DropdownMenu.jsx"},{"name":"DropdownMenuLabel","sourcePath":"components/navigation/DropdownMenu.jsx"},{"name":"DropdownMenuSeparator","sourcePath":"components/navigation/DropdownMenu.jsx"},{"name":"DropdownMenuShortcut","sourcePath":"components/navigation/DropdownMenu.jsx"},{"name":"Tabs","sourcePath":"components/navigation/Tabs.jsx"},{"name":"TabsList","sourcePath":"components/navigation/Tabs.jsx"},{"name":"TabsTrigger","sourcePath":"components/navigation/Tabs.jsx"},{"name":"TabsContent","sourcePath":"components/navigation/Tabs.jsx"}],"sourceHashes":{"components/actions/Badge.jsx":"b2ecdebe71d5","components/actions/Button.jsx":"27bb355d5a0b","components/actions/Kbd.jsx":"b1404ea86c1c","components/actions/Spinner.jsx":"3451b14db612","components/actions/Toggle.jsx":"d819df885076","components/data-display/Avatar.jsx":"ab6de0cd5f3b","components/data-display/Card.jsx":"c9cec82c88fb","components/data-display/Progress.jsx":"a9184399ad92","components/data-display/Separator.jsx":"2dfd4af52765","components/data-display/Skeleton.jsx":"c986fdd5058c","components/data-display/Table.jsx":"3d154b9bb5d7","components/feedback/Alert.jsx":"2110c3895157","components/feedback/Dialog.jsx":"5790ca89d959","components/feedback/Tooltip.jsx":"901aaa5da9e4","components/forms/Checkbox.jsx":"fee551084d47","components/forms/Input.jsx":"9b73ce4f91dc","components/forms/Label.jsx":"9591adf60739","components/forms/RadioGroup.jsx":"f969f6499ae6","components/forms/Select.jsx":"d7e4dfc853a8","components/forms/Slider.jsx":"65e2d4f7b204","components/forms/Switch.jsx":"096e4da24fb7","components/forms/Textarea.jsx":"221646b97d83","components/icon/Icon.jsx":"6b719c222137","components/navigation/Accordion.jsx":"9b2f12eaf26e","components/navigation/Breadcrumb.jsx":"81253159cd11","components/navigation/DropdownMenu.jsx":"2829e6e99b3f","components/navigation/Tabs.jsx":"123be2c40b8a","ui_kits/dashboard/DashboardScreen.jsx":"592462444c40","ui_kits/dashboard/LoginScreen.jsx":"032c434da40a"},"inlinedExternals":[],"unexposedExports":[]} */

(() => {

const __ds_ns = (window.ShadcnSvelteDesignSystem_ce4a1c = window.ShadcnSvelteDesignSystem_ce4a1c || {});

const __ds_scope = {};

(__ds_ns.__errors = __ds_ns.__errors || []);

// components/actions/Badge.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
const cx = (...c) => c.filter(Boolean).join(" ");

/** Badge — a small status / label pill. */
function Badge({
  variant = "default",
  className = "",
  as,
  children,
  ...props
}) {
  const Tag = as || (props.href ? "a" : "span");
  return /*#__PURE__*/React.createElement(Tag, _extends({
    "data-slot": "badge",
    className: cx("cn-badge", `cn-badge-variant-${variant}`, className)
  }, props), children);
}
Object.assign(__ds_scope, { Badge });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/actions/Badge.jsx", error: String((e && e.message) || e) }); }

// components/actions/Button.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
const cx = (...c) => c.filter(Boolean).join(" ");

/**
 * Button — the primary action element. Ports the shadcn-svelte `buttonVariants`
 * (default | outline | secondary | ghost | destructive | link) and every size.
 */
function Button({
  variant = "default",
  size = "default",
  className = "",
  as,
  children,
  ...props
}) {
  const Tag = as || (props.href ? "a" : "button");
  const extra = {};
  if (Tag === "button" && !props.type) extra.type = "button";
  return /*#__PURE__*/React.createElement(Tag, _extends({
    "data-slot": "button",
    className: cx("cn-button", `cn-button-variant-${variant}`, `cn-button-size-${size}`, className)
  }, extra, props), children);
}
Object.assign(__ds_scope, { Button });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/actions/Button.jsx", error: String((e && e.message) || e) }); }

// components/actions/Kbd.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
/** Kbd — a keyboard-key hint. Group several inside a KbdGroup for shortcuts. */
function Kbd({
  className = "",
  children,
  ...props
}) {
  return /*#__PURE__*/React.createElement("kbd", _extends({
    "data-slot": "kbd",
    className: ["cn-kbd", className].filter(Boolean).join(" ")
  }, props), children);
}
function KbdGroup({
  className = "",
  children,
  ...props
}) {
  return /*#__PURE__*/React.createElement("span", _extends({
    "data-slot": "kbd-group",
    className: className,
    style: {
      display: "inline-flex",
      alignItems: "center",
      gap: "0.25rem"
    }
  }, props), children);
}
Object.assign(__ds_scope, { Kbd, KbdGroup });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/actions/Kbd.jsx", error: String((e && e.message) || e) }); }

// components/actions/Toggle.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
const cx = (...c) => c.filter(Boolean).join(" ");

/** Toggle — a two-state pressable button (aria-pressed drives the on state). */
function Toggle({
  variant = "default",
  size = "default",
  pressed,
  onPressedChange,
  className = "",
  children,
  ...props
}) {
  const [internal, setInternal] = React.useState(false);
  const isPressed = pressed !== undefined ? pressed : internal;
  return /*#__PURE__*/React.createElement("button", _extends({
    type: "button",
    "data-slot": "toggle",
    "aria-pressed": isPressed,
    onClick: () => {
      if (pressed === undefined) setInternal(v => !v);
      onPressedChange && onPressedChange(!isPressed);
    },
    className: cx("cn-toggle", `cn-toggle-variant-${variant}`, `cn-toggle-size-${size}`, className)
  }, props), children);
}
Object.assign(__ds_scope, { Toggle });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/actions/Toggle.jsx", error: String((e && e.message) || e) }); }

// components/data-display/Avatar.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
const cx = (...c) => c.filter(Boolean).join(" ");

/** Avatar — circular user image with a text/icon fallback. */
function Avatar({
  src,
  alt = "",
  fallback,
  size = "default",
  className = "",
  children,
  ...props
}) {
  const [errored, setErrored] = React.useState(false);
  return /*#__PURE__*/React.createElement("span", _extends({
    "data-slot": "avatar",
    "data-size": size,
    className: cx("cn-avatar", className)
  }, props), src && !errored ? /*#__PURE__*/React.createElement("img", {
    className: "cn-avatar-image",
    src: src,
    alt: alt,
    onError: () => setErrored(true)
  }) : /*#__PURE__*/React.createElement("span", {
    className: "cn-avatar-fallback"
  }, fallback || children));
}
Object.assign(__ds_scope, { Avatar });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/data-display/Avatar.jsx", error: String((e && e.message) || e) }); }

// components/data-display/Card.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
const cx = (...c) => c.filter(Boolean).join(" ");

/** Card — elevated surface container. Compose with the sub-parts below. */
function Card({
  className = "",
  children,
  ...props
}) {
  return /*#__PURE__*/React.createElement("div", _extends({
    "data-slot": "card",
    className: cx("cn-card", className)
  }, props), children);
}
function CardHeader({
  className = "",
  children,
  ...props
}) {
  return /*#__PURE__*/React.createElement("div", _extends({
    "data-slot": "card-header",
    className: cx("cn-card-header", className)
  }, props), children);
}
function CardTitle({
  className = "",
  children,
  ...props
}) {
  return /*#__PURE__*/React.createElement("div", _extends({
    "data-slot": "card-title",
    className: cx("cn-card-title", className)
  }, props), children);
}
function CardDescription({
  className = "",
  children,
  ...props
}) {
  return /*#__PURE__*/React.createElement("div", _extends({
    "data-slot": "card-description",
    className: cx("cn-card-description", className)
  }, props), children);
}
function CardContent({
  className = "",
  children,
  ...props
}) {
  return /*#__PURE__*/React.createElement("div", _extends({
    "data-slot": "card-content",
    className: cx("cn-card-content", className)
  }, props), children);
}
function CardFooter({
  className = "",
  children,
  ...props
}) {
  return /*#__PURE__*/React.createElement("div", _extends({
    "data-slot": "card-footer",
    className: cx("cn-card-footer", className)
  }, props), children);
}
Object.assign(__ds_scope, { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/data-display/Card.jsx", error: String((e && e.message) || e) }); }

// components/data-display/Progress.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
const cx = (...c) => c.filter(Boolean).join(" ");
/** Progress — determinate progress bar (0–100). */
function Progress({
  value = 0,
  className = "",
  ...props
}) {
  const v = Math.min(100, Math.max(0, value));
  return /*#__PURE__*/React.createElement("div", _extends({
    "data-slot": "progress",
    role: "progressbar",
    "aria-valuenow": v,
    "aria-valuemin": 0,
    "aria-valuemax": 100,
    className: cx("cn-progress", className)
  }, props), /*#__PURE__*/React.createElement("div", {
    className: "cn-progress-indicator",
    style: {
      width: `${v}%`
    }
  }));
}
Object.assign(__ds_scope, { Progress });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/data-display/Progress.jsx", error: String((e && e.message) || e) }); }

// components/data-display/Separator.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
const cx = (...c) => c.filter(Boolean).join(" ");

/** Separator — a hairline divider, horizontal (default) or vertical. */
function Separator({
  orientation = "horizontal",
  className = "",
  ...props
}) {
  return /*#__PURE__*/React.createElement("div", _extends({
    "data-slot": "separator",
    role: "separator",
    "aria-orientation": orientation,
    className: cx("cn-separator", `cn-separator-${orientation}`, className)
  }, props));
}
Object.assign(__ds_scope, { Separator });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/data-display/Separator.jsx", error: String((e && e.message) || e) }); }

// components/data-display/Skeleton.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
const cx = (...c) => c.filter(Boolean).join(" ");
/** Skeleton — animated loading placeholder. Size it with width/height. */
function Skeleton({
  className = "",
  style,
  ...props
}) {
  return /*#__PURE__*/React.createElement("div", _extends({
    "data-slot": "skeleton",
    className: cx("cn-skeleton", className),
    style: style
  }, props));
}
Object.assign(__ds_scope, { Skeleton });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/data-display/Skeleton.jsx", error: String((e && e.message) || e) }); }

// components/data-display/Table.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
const cx = (...c) => c.filter(Boolean).join(" ");

/** Table primitives — thin wrappers over native table elements. */
function Table({
  className = "",
  children,
  ...props
}) {
  return /*#__PURE__*/React.createElement("div", {
    "data-slot": "table-container",
    className: "cn-table-container"
  }, /*#__PURE__*/React.createElement("table", _extends({
    "data-slot": "table",
    className: cx("cn-table", className)
  }, props), children));
}
function TableHeader({
  className = "",
  children,
  ...props
}) {
  return /*#__PURE__*/React.createElement("thead", _extends({
    "data-slot": "table-header",
    className: cx("cn-table-header", className)
  }, props), children);
}
function TableBody({
  className = "",
  children,
  ...props
}) {
  return /*#__PURE__*/React.createElement("tbody", _extends({
    "data-slot": "table-body",
    className: cx("cn-table-body", className)
  }, props), children);
}
function TableFooter({
  className = "",
  children,
  ...props
}) {
  return /*#__PURE__*/React.createElement("tfoot", _extends({
    "data-slot": "table-footer",
    className: cx("cn-table-footer", className)
  }, props), children);
}
function TableRow({
  className = "",
  children,
  ...props
}) {
  return /*#__PURE__*/React.createElement("tr", _extends({
    "data-slot": "table-row",
    className: cx("cn-table-row", className)
  }, props), children);
}
function TableHead({
  className = "",
  children,
  ...props
}) {
  return /*#__PURE__*/React.createElement("th", _extends({
    "data-slot": "table-head",
    className: cx("cn-table-head", className)
  }, props), children);
}
function TableCell({
  className = "",
  children,
  ...props
}) {
  return /*#__PURE__*/React.createElement("td", _extends({
    "data-slot": "table-cell",
    className: cx("cn-table-cell", className)
  }, props), children);
}
function TableCaption({
  className = "",
  children,
  ...props
}) {
  return /*#__PURE__*/React.createElement("caption", _extends({
    "data-slot": "table-caption",
    className: cx("cn-table-caption", className)
  }, props), children);
}
Object.assign(__ds_scope, { Table, TableHeader, TableBody, TableFooter, TableRow, TableHead, TableCell, TableCaption });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/data-display/Table.jsx", error: String((e && e.message) || e) }); }

// components/feedback/Alert.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
const cx = (...c) => c.filter(Boolean).join(" ");

/** Alert — a callout box. Optional leading <Icon>, plus Title + Description. */
function Alert({
  variant = "default",
  className = "",
  children,
  ...props
}) {
  return /*#__PURE__*/React.createElement("div", _extends({
    "data-slot": "alert",
    role: "alert",
    className: cx("cn-alert", `cn-alert-variant-${variant}`, className)
  }, props), children);
}
function AlertTitle({
  className = "",
  children,
  ...props
}) {
  return /*#__PURE__*/React.createElement("div", _extends({
    "data-slot": "alert-title",
    className: cx("cn-alert-title", className)
  }, props), children);
}
function AlertDescription({
  className = "",
  children,
  ...props
}) {
  return /*#__PURE__*/React.createElement("div", _extends({
    "data-slot": "alert-description",
    className: cx("cn-alert-description", className)
  }, props), children);
}
Object.assign(__ds_scope, { Alert, AlertTitle, AlertDescription });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/feedback/Alert.jsx", error: String((e && e.message) || e) }); }

// components/feedback/Tooltip.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
const cx = (...c) => c.filter(Boolean).join(" ");

/** Tooltip — hover/focus popup. Wraps a trigger; `content` is the bubble text. */
function Tooltip({
  content,
  side = "top",
  className = "",
  children,
  ...props
}) {
  const [open, setOpen] = React.useState(false);
  const pos = {
    top: {
      bottom: "calc(100% + 8px)",
      left: "50%",
      transform: "translateX(-50%)"
    },
    bottom: {
      top: "calc(100% + 8px)",
      left: "50%",
      transform: "translateX(-50%)"
    },
    left: {
      right: "calc(100% + 8px)",
      top: "50%",
      transform: "translateY(-50%)"
    },
    right: {
      left: "calc(100% + 8px)",
      top: "50%",
      transform: "translateY(-50%)"
    }
  }[side];
  return /*#__PURE__*/React.createElement("span", _extends({
    "data-slot": "tooltip",
    style: {
      position: "relative",
      display: "inline-flex"
    },
    onMouseEnter: () => setOpen(true),
    onMouseLeave: () => setOpen(false),
    onFocus: () => setOpen(true),
    onBlur: () => setOpen(false)
  }, props), children, open ? /*#__PURE__*/React.createElement("span", {
    role: "tooltip",
    className: cx("cn-tooltip-content", className),
    style: {
      position: "absolute",
      whiteSpace: "nowrap",
      zIndex: 50,
      ...pos
    }
  }, content) : null);
}
Object.assign(__ds_scope, { Tooltip });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/feedback/Tooltip.jsx", error: String((e && e.message) || e) }); }

// components/forms/Input.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
const cx = (...c) => c.filter(Boolean).join(" ");
/** Input — single-line text field. */
function Input({
  className = "",
  type = "text",
  ...props
}) {
  return /*#__PURE__*/React.createElement("input", _extends({
    "data-slot": "input",
    type: type,
    className: cx("cn-input", className)
  }, props));
}
Object.assign(__ds_scope, { Input });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/forms/Input.jsx", error: String((e && e.message) || e) }); }

// components/forms/Label.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
const cx = (...c) => c.filter(Boolean).join(" ");
/** Label — form label; associate with a control via htmlFor. */
function Label({
  className = "",
  children,
  ...props
}) {
  return /*#__PURE__*/React.createElement("label", _extends({
    "data-slot": "label",
    className: cx("cn-label", className)
  }, props), children);
}
Object.assign(__ds_scope, { Label });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/forms/Label.jsx", error: String((e && e.message) || e) }); }

// components/forms/RadioGroup.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
const cx = (...c) => c.filter(Boolean).join(" ");
const RadioCtx = React.createContext(null);

/** RadioGroup — single-select group. Controlled via `value`/`onValueChange`. */
function RadioGroup({
  value,
  defaultValue,
  onValueChange,
  className = "",
  children,
  ...props
}) {
  const [internal, setInternal] = React.useState(defaultValue);
  const current = value !== undefined ? value : internal;
  const set = v => {
    if (value === undefined) setInternal(v);
    onValueChange && onValueChange(v);
  };
  return /*#__PURE__*/React.createElement("div", _extends({
    role: "radiogroup",
    "data-slot": "radio-group",
    className: cx("cn-radio-group", className)
  }, props), /*#__PURE__*/React.createElement(RadioCtx.Provider, {
    value: {
      current,
      set
    }
  }, children));
}

/** RadioGroupItem — one option; give each a unique `value`. */
function RadioGroupItem({
  value,
  className = "",
  ...props
}) {
  const ctx = React.useContext(RadioCtx);
  const checked = ctx && ctx.current === value;
  return /*#__PURE__*/React.createElement("button", _extends({
    type: "button",
    role: "radio",
    "aria-checked": !!checked,
    "data-slot": "radio-group-item",
    "data-checked": checked ? "" : undefined,
    onClick: () => ctx && ctx.set(value),
    className: cx("cn-radio-group-item", className)
  }, props), checked ? /*#__PURE__*/React.createElement("span", {
    className: "cn-radio-group-indicator-icon"
  }) : null);
}
Object.assign(__ds_scope, { RadioGroup, RadioGroupItem });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/forms/RadioGroup.jsx", error: String((e && e.message) || e) }); }

// components/forms/Slider.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
const cx = (...c) => c.filter(Boolean).join(" ");

/** Slider — single-thumb range slider. Controlled via `value`/`onValueChange`. */
function Slider({
  value,
  defaultValue = 50,
  onValueChange,
  min = 0,
  max = 100,
  step = 1,
  className = "",
  ...props
}) {
  const [internal, setInternal] = React.useState(defaultValue);
  const val = value !== undefined ? value : internal;
  const pct = (val - min) / (max - min) * 100;
  const trackRef = React.useRef(null);
  const setFromClientX = clientX => {
    const el = trackRef.current;
    if (!el) return;
    const rect = el.getBoundingClientRect();
    const ratio = Math.min(1, Math.max(0, (clientX - rect.left) / rect.width));
    let next = min + ratio * (max - min);
    next = Math.round(next / step) * step;
    next = Math.min(max, Math.max(min, next));
    if (value === undefined) setInternal(next);
    onValueChange && onValueChange(next);
  };
  const onDown = e => {
    setFromClientX(e.clientX);
    const move = ev => setFromClientX(ev.clientX);
    const up = () => {
      window.removeEventListener("pointermove", move);
      window.removeEventListener("pointerup", up);
    };
    window.addEventListener("pointermove", move);
    window.addEventListener("pointerup", up);
  };
  return /*#__PURE__*/React.createElement("span", _extends({
    "data-slot": "slider",
    role: "slider",
    "aria-valuenow": val,
    "aria-valuemin": min,
    "aria-valuemax": max,
    className: cx("cn-slider", className),
    onPointerDown: onDown
  }, props), /*#__PURE__*/React.createElement("span", {
    ref: trackRef,
    className: "cn-slider-track",
    "data-slot": "slider-track"
  }, /*#__PURE__*/React.createElement("span", {
    className: "cn-slider-range",
    style: {
      width: `${pct}%`
    }
  })), /*#__PURE__*/React.createElement("span", {
    className: "cn-slider-thumb",
    style: {
      left: `${pct}%`
    }
  }));
}
Object.assign(__ds_scope, { Slider });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/forms/Slider.jsx", error: String((e && e.message) || e) }); }

// components/forms/Switch.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
const cx = (...c) => c.filter(Boolean).join(" ");

/** Switch — on/off toggle. Controlled via `checked`, or uncontrolled. */
function Switch({
  checked,
  defaultChecked = false,
  onCheckedChange,
  className = "",
  disabled,
  ...props
}) {
  const [internal, setInternal] = React.useState(defaultChecked);
  const isChecked = checked !== undefined ? checked : internal;
  return /*#__PURE__*/React.createElement("button", _extends({
    type: "button",
    role: "switch",
    "aria-checked": isChecked,
    "data-slot": "switch",
    "data-checked": isChecked ? "" : undefined,
    disabled: disabled,
    onClick: () => {
      if (checked === undefined) setInternal(v => !v);
      onCheckedChange && onCheckedChange(!isChecked);
    },
    className: cx("cn-switch", className),
    style: disabled ? {
      opacity: 0.5,
      pointerEvents: "none"
    } : undefined
  }, props), /*#__PURE__*/React.createElement("span", {
    className: "cn-switch-thumb",
    "data-slot": "switch-thumb"
  }));
}
Object.assign(__ds_scope, { Switch });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/forms/Switch.jsx", error: String((e && e.message) || e) }); }

// components/forms/Textarea.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
const cx = (...c) => c.filter(Boolean).join(" ");
/** Textarea — multi-line text field. */
function Textarea({
  className = "",
  rows = 3,
  ...props
}) {
  return /*#__PURE__*/React.createElement("textarea", _extends({
    "data-slot": "textarea",
    rows: rows,
    className: cx("cn-textarea", className)
  }, props));
}
Object.assign(__ds_scope, { Textarea });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/forms/Textarea.jsx", error: String((e && e.message) || e) }); }

// components/icon/Icon.jsx
try { (() => {
/**
 * Lucide icon paths (MIT © Lucide contributors) — the canonical shadcn-svelte
 * icon set. currentColor stroke, 24×24 viewBox, 2px stroke, round caps/joins.
 */
const ICON_PATHS = {
  check: '<path d="M20 6 9 17l-5-5"/>',
  "chevron-down": '<path d="m6 9 6 6 6-6"/>',
  "chevron-right": '<path d="m9 18 6-6-6-6"/>',
  "chevron-left": '<path d="m15 18-6-6 6-6"/>',
  "chevron-up": '<path d="m18 15-6-6-6 6"/>',
  "chevrons-up-down": '<path d="m7 15 5 5 5-5"/><path d="m7 9 5-5 5 5"/>',
  x: '<path d="M18 6 6 18"/><path d="m6 6 12 12"/>',
  search: '<circle cx="11" cy="11" r="8"/><path d="m21 21-4.3-4.3"/>',
  plus: '<path d="M5 12h14"/><path d="M12 5v14"/>',
  minus: '<path d="M5 12h14"/>',
  settings: '<path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z"/><circle cx="12" cy="12" r="3"/>',
  user: '<path d="M19 21v-2a4 4 0 0 0-4-4H9a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/>',
  bell: '<path d="M10.268 21a2 2 0 0 0 3.464 0"/><path d="M3.262 15.326A1 1 0 0 0 4 17h16a1 1 0 0 0 .74-1.673C19.41 13.956 18 12.499 18 8A6 6 0 0 0 6 8c0 4.499-1.411 5.956-2.738 7.326"/>',
  calendar: '<path d="M8 2v4"/><path d="M16 2v4"/><rect width="18" height="18" x="3" y="4" rx="2"/><path d="M3 10h18"/>',
  "credit-card": '<rect width="20" height="14" x="2" y="5" rx="2"/><line x1="2" x2="22" y1="10" y2="10"/>',
  mail: '<rect width="20" height="16" x="2" y="4" rx="2"/><path d="m22 7-8.97 5.7a1.94 1.94 0 0 1-2.06 0L2 7"/>',
  "arrow-right": '<path d="M5 12h14"/><path d="m12 5 7 7-7 7"/>',
  "circle-check": '<circle cx="12" cy="12" r="10"/><path d="m9 12 2 2 4-4"/>',
  "triangle-alert": '<path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3"/><path d="M12 9v4"/><path d="M12 17h.01"/>',
  info: '<circle cx="12" cy="12" r="10"/><path d="M12 16v-4"/><path d="M12 8h.01"/>',
  "loader-circle": '<path d="M21 12a9 9 0 1 1-6.219-8.56"/>',
  "more-horizontal": '<circle cx="12" cy="12" r="1"/><circle cx="19" cy="12" r="1"/><circle cx="5" cy="12" r="1"/>',
  "trash-2": '<path d="M3 6h18"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/><line x1="10" x2="10" y1="11" y2="17"/><line x1="14" x2="14" y1="11" y2="17"/>',
  copy: '<rect width="14" height="14" x="8" y="8" rx="2" ry="2"/><path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2"/>',
  sun: '<circle cx="12" cy="12" r="4"/><path d="M12 2v2"/><path d="M12 20v2"/><path d="m4.93 4.93 1.41 1.41"/><path d="m17.66 17.66 1.41 1.41"/><path d="M2 12h2"/><path d="M20 12h2"/><path d="m6.34 17.66-1.41 1.41"/><path d="m19.07 4.93-1.41 1.41"/>',
  moon: '<path d="M12 3a6 6 0 0 0 9 9 9 9 0 1 1-9-9Z"/>',
  house: '<path d="M15 21v-8a1 1 0 0 0-1-1h-4a1 1 0 0 0-1 1v8"/><path d="M3 10a2 2 0 0 1 .709-1.528l7-5.999a2 2 0 0 1 2.582 0l7 5.999A2 2 0 0 1 21 10v9a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/>',
  star: '<path d="M11.525 2.295a.53.53 0 0 1 .95 0l2.31 4.679a2.123 2.123 0 0 0 1.595 1.16l5.166.756a.53.53 0 0 1 .294.904l-3.736 3.638a2.123 2.123 0 0 0-.611 1.878l.882 5.14a.53.53 0 0 1-.771.56l-4.618-2.428a2.122 2.122 0 0 0-1.973 0L6.396 21.01a.53.53 0 0 1-.77-.56l.881-5.139a2.122 2.122 0 0 0-.611-1.879L2.16 9.795a.53.53 0 0 1 .294-.906l5.165-.755a2.122 2.122 0 0 0 1.597-1.16z"/>',
  "panel-left": '<rect width="18" height="18" x="3" y="3" rx="2"/><path d="M9 3v18"/>',
  "log-out": '<path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/><polyline points="16 17 21 12 16 7"/><line x1="21" x2="9" y1="12" y2="12"/>',
  "credit-card-2": '<rect width="20" height="14" x="2" y="5" rx="2"/><line x1="2" x2="22" y1="10" y2="10"/>',
  folder: '<path d="M20 20a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.9a2 2 0 0 1-1.69-.9L9.6 3.9A2 2 0 0 0 7.93 3H4a2 2 0 0 0-2 2v13a2 2 0 0 0 2 2Z"/>',
  file: '<path d="M15 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7Z"/><path d="M14 2v4a2 2 0 0 0 2 2h4"/>',
  heart: '<path d="M19 14c1.49-1.46 3-3.21 3-5.5A5.5 5.5 0 0 0 16.5 3c-1.76 0-3 .5-4.5 2-1.5-1.5-2.74-2-4.5-2A5.5 5.5 0 0 0 2 8.5c0 2.3 1.5 4.05 3 5.5l7 7Z"/>',
  "circle-help": '<circle cx="12" cy="12" r="10"/><path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3"/><path d="M12 17h.01"/>'
};

/**
 * A Lucide icon. Renders inline SVG so it inherits color via currentColor
 * and scales with font-size or the `size` prop.
 */
function Icon({
  name,
  size = 16,
  strokeWidth = 2,
  className = "",
  style,
  ...rest
}) {
  const inner = ICON_PATHS[name];
  return React.createElement("svg", {
    xmlns: "http://www.w3.org/2000/svg",
    width: size,
    height: size,
    viewBox: "0 0 24 24",
    fill: "none",
    stroke: "currentColor",
    strokeWidth,
    strokeLinecap: "round",
    strokeLinejoin: "round",
    className,
    "aria-hidden": "true",
    style,
    dangerouslySetInnerHTML: {
      __html: inner || ""
    },
    ...rest
  });
}
Object.assign(__ds_scope, { ICON_PATHS, Icon });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/icon/Icon.jsx", error: String((e && e.message) || e) }); }

// components/actions/Spinner.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
/** Spinner — a spinning loader-circle glyph. Inherits currentColor & size. */
function Spinner({
  size = 16,
  className = "",
  ...props
}) {
  return /*#__PURE__*/React.createElement(__ds_scope.Icon, _extends({
    name: "loader-circle",
    size: size,
    className: ["cn-spinner", className].filter(Boolean).join(" ")
  }, props));
}
Object.assign(__ds_scope, { Spinner });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/actions/Spinner.jsx", error: String((e && e.message) || e) }); }

// components/feedback/Dialog.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
const cx = (...c) => c.filter(Boolean).join(" ");

/** Dialog — modal overlay. Controlled via `open`/`onOpenChange`. */
function Dialog({
  open,
  onOpenChange,
  className = "",
  children,
  ...props
}) {
  if (!open) return null;
  const close = () => onOpenChange && onOpenChange(false);
  return /*#__PURE__*/React.createElement("div", null, /*#__PURE__*/React.createElement("div", {
    className: "cn-dialog-overlay",
    onClick: close
  }), /*#__PURE__*/React.createElement("div", _extends({
    role: "dialog",
    "aria-modal": "true",
    "data-slot": "dialog-content",
    className: cx("cn-dialog-content", className),
    style: {
      position: "fixed",
      top: "50%",
      left: "50%",
      transform: "translate(-50%, -50%)",
      zIndex: 50
    }
  }, props), /*#__PURE__*/React.createElement("button", {
    type: "button",
    onClick: close,
    "aria-label": "Close",
    className: "cn-button cn-button-variant-ghost cn-button-size-icon-sm",
    style: {
      position: "absolute",
      top: "1rem",
      right: "1rem"
    }
  }, /*#__PURE__*/React.createElement(__ds_scope.Icon, {
    name: "x"
  })), children));
}
function DialogHeader({
  className = "",
  children,
  ...props
}) {
  return /*#__PURE__*/React.createElement("div", _extends({
    "data-slot": "dialog-header",
    className: cx("cn-dialog-header", className)
  }, props), children);
}
function DialogTitle({
  className = "",
  children,
  ...props
}) {
  return /*#__PURE__*/React.createElement("div", _extends({
    "data-slot": "dialog-title",
    className: cx("cn-dialog-title", className)
  }, props), children);
}
function DialogDescription({
  className = "",
  children,
  ...props
}) {
  return /*#__PURE__*/React.createElement("div", _extends({
    "data-slot": "dialog-description",
    className: cx("cn-dialog-description", className)
  }, props), children);
}
function DialogFooter({
  className = "",
  children,
  ...props
}) {
  return /*#__PURE__*/React.createElement("div", _extends({
    "data-slot": "dialog-footer",
    className: cx("cn-dialog-footer", className)
  }, props), children);
}
Object.assign(__ds_scope, { Dialog, DialogHeader, DialogTitle, DialogDescription, DialogFooter });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/feedback/Dialog.jsx", error: String((e && e.message) || e) }); }

// components/forms/Checkbox.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
const cx = (...c) => c.filter(Boolean).join(" ");

/** Checkbox — controlled or uncontrolled boolean control. */
function Checkbox({
  checked,
  defaultChecked = false,
  onCheckedChange,
  className = "",
  disabled,
  ...props
}) {
  const [internal, setInternal] = React.useState(defaultChecked);
  const isChecked = checked !== undefined ? checked : internal;
  return /*#__PURE__*/React.createElement("button", _extends({
    type: "button",
    role: "checkbox",
    "aria-checked": isChecked,
    "data-slot": "checkbox",
    "data-checked": isChecked ? "" : undefined,
    disabled: disabled,
    onClick: () => {
      if (checked === undefined) setInternal(v => !v);
      onCheckedChange && onCheckedChange(!isChecked);
    },
    className: cx("cn-checkbox", className),
    style: disabled ? {
      opacity: 0.5,
      pointerEvents: "none"
    } : undefined
  }, props), isChecked ? /*#__PURE__*/React.createElement(__ds_scope.Icon, {
    name: "check",
    size: 14,
    strokeWidth: 3
  }) : null);
}
Object.assign(__ds_scope, { Checkbox });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/forms/Checkbox.jsx", error: String((e && e.message) || e) }); }

// components/forms/Select.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
const cx = (...c) => c.filter(Boolean).join(" ");

/** Select — styled native select with a chevron. Pass <option>s as children. */
function Select({
  className = "",
  children,
  size = "default",
  ...props
}) {
  return /*#__PURE__*/React.createElement("span", {
    className: "cn-native-select-wrapper"
  }, /*#__PURE__*/React.createElement("select", _extends({
    "data-slot": "native-select",
    "data-size": size,
    className: cx("cn-native-select", className),
    style: size === "sm" ? {
      height: "2rem"
    } : undefined
  }, props), children), /*#__PURE__*/React.createElement(__ds_scope.Icon, {
    name: "chevron-down",
    className: "cn-native-select-icon"
  }));
}
Object.assign(__ds_scope, { Select });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/forms/Select.jsx", error: String((e && e.message) || e) }); }

// components/navigation/Accordion.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
const cx = (...c) => c.filter(Boolean).join(" ");
const AccCtx = React.createContext(null);

/** Accordion — collapsible sections. type "single" (default) or "multiple". */
function Accordion({
  type = "single",
  defaultValue,
  className = "",
  children,
  ...props
}) {
  const [open, setOpen] = React.useState(defaultValue == null ? [] : Array.isArray(defaultValue) ? defaultValue : [defaultValue]);
  const toggle = val => {
    setOpen(prev => {
      const has = prev.includes(val);
      if (type === "multiple") return has ? prev.filter(v => v !== val) : [...prev, val];
      return has ? [] : [val];
    });
  };
  return /*#__PURE__*/React.createElement("div", _extends({
    "data-slot": "accordion",
    className: cx("cn-accordion", className)
  }, props), /*#__PURE__*/React.createElement(AccCtx.Provider, {
    value: {
      open,
      toggle
    }
  }, children));
}
function AccordionItem({
  value,
  className = "",
  children,
  ...props
}) {
  return /*#__PURE__*/React.createElement("div", _extends({
    "data-slot": "accordion-item",
    "data-value": value,
    className: cx("cn-accordion-item", className)
  }, props), React.Children.map(children, c => React.isValidElement(c) ? React.cloneElement(c, {
    __value: value
  }) : c));
}
function AccordionTrigger({
  __value,
  className = "",
  children,
  ...props
}) {
  const ctx = React.useContext(AccCtx);
  const isOpen = ctx && ctx.open.includes(__value);
  return /*#__PURE__*/React.createElement("button", _extends({
    type: "button",
    "data-slot": "accordion-trigger",
    "data-open": isOpen ? "" : undefined,
    "aria-expanded": !!isOpen,
    onClick: () => ctx && ctx.toggle(__value),
    className: cx("cn-accordion-trigger", className)
  }, props), children, /*#__PURE__*/React.createElement(__ds_scope.Icon, {
    name: "chevron-down"
  }));
}
function AccordionContent({
  __value,
  className = "",
  children,
  ...props
}) {
  const ctx = React.useContext(AccCtx);
  const isOpen = ctx && ctx.open.includes(__value);
  if (!isOpen) return null;
  return /*#__PURE__*/React.createElement("div", _extends({
    "data-slot": "accordion-content",
    className: cx("cn-accordion-content", className)
  }, props), children);
}
Object.assign(__ds_scope, { Accordion, AccordionItem, AccordionTrigger, AccordionContent });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/navigation/Accordion.jsx", error: String((e && e.message) || e) }); }

// components/navigation/Breadcrumb.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
const cx = (...c) => c.filter(Boolean).join(" ");

/** Breadcrumb — hierarchical path. Compose the parts below. */
function Breadcrumb({
  className = "",
  children,
  ...props
}) {
  return /*#__PURE__*/React.createElement("nav", _extends({
    "aria-label": "breadcrumb",
    "data-slot": "breadcrumb",
    className: cx("cn-breadcrumb", className)
  }, props), children);
}
function BreadcrumbList({
  className = "",
  children,
  ...props
}) {
  return /*#__PURE__*/React.createElement("ol", _extends({
    "data-slot": "breadcrumb-list",
    className: cx("cn-breadcrumb-list", className),
    style: {
      listStyle: "none",
      margin: 0,
      padding: 0
    }
  }, props), children);
}
function BreadcrumbItem({
  className = "",
  children,
  ...props
}) {
  return /*#__PURE__*/React.createElement("li", _extends({
    "data-slot": "breadcrumb-item",
    className: cx("cn-breadcrumb-item", className)
  }, props), children);
}
function BreadcrumbLink({
  className = "",
  children,
  ...props
}) {
  return /*#__PURE__*/React.createElement("a", _extends({
    "data-slot": "breadcrumb-link",
    className: cx("cn-breadcrumb-link", className)
  }, props), children);
}
function BreadcrumbPage({
  className = "",
  children,
  ...props
}) {
  return /*#__PURE__*/React.createElement("span", _extends({
    role: "link",
    "aria-current": "page",
    "data-slot": "breadcrumb-page",
    className: cx("cn-breadcrumb-page", className)
  }, props), children);
}
function BreadcrumbSeparator({
  className = "",
  children,
  ...props
}) {
  return /*#__PURE__*/React.createElement("li", _extends({
    role: "presentation",
    "aria-hidden": "true",
    "data-slot": "breadcrumb-separator",
    className: cx("cn-breadcrumb-separator", className)
  }, props), children || /*#__PURE__*/React.createElement(__ds_scope.Icon, {
    name: "chevron-right",
    size: 14
  }));
}
Object.assign(__ds_scope, { Breadcrumb, BreadcrumbList, BreadcrumbItem, BreadcrumbLink, BreadcrumbPage, BreadcrumbSeparator });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/navigation/Breadcrumb.jsx", error: String((e && e.message) || e) }); }

// components/navigation/DropdownMenu.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
const cx = (...c) => c.filter(Boolean).join(" ");
const DdCtx = React.createContext(null);

/** DropdownMenu — click-triggered menu. Wrap a Trigger + Content. */
function DropdownMenu({
  className = "",
  children,
  ...props
}) {
  const [open, setOpen] = React.useState(false);
  const ref = React.useRef(null);
  React.useEffect(() => {
    if (!open) return;
    const onDoc = e => {
      if (ref.current && !ref.current.contains(e.target)) setOpen(false);
    };
    document.addEventListener("mousedown", onDoc);
    return () => document.removeEventListener("mousedown", onDoc);
  }, [open]);
  return /*#__PURE__*/React.createElement("div", _extends({
    ref: ref,
    "data-slot": "dropdown-menu",
    style: {
      position: "relative",
      display: "inline-block"
    },
    className: className
  }, props), /*#__PURE__*/React.createElement(DdCtx.Provider, {
    value: {
      open,
      setOpen
    }
  }, children));
}
function DropdownMenuTrigger({
  children
}) {
  const ctx = React.useContext(DdCtx);
  const child = React.Children.only(children);
  return React.cloneElement(child, {
    "aria-expanded": ctx.open,
    onClick: e => {
      child.props.onClick && child.props.onClick(e);
      ctx.setOpen(v => !v);
    }
  });
}
function DropdownMenuContent({
  align = "start",
  className = "",
  children,
  ...props
}) {
  const ctx = React.useContext(DdCtx);
  if (!ctx.open) return null;
  const alignStyle = align === "end" ? {
    right: 0
  } : {
    left: 0
  };
  return /*#__PURE__*/React.createElement("div", _extends({
    role: "menu",
    "data-slot": "dropdown-menu-content",
    className: cx("cn-dropdown-menu-content", className),
    style: {
      position: "absolute",
      top: "calc(100% + 4px)",
      zIndex: 50,
      ...alignStyle
    }
  }, props), children);
}
function DropdownMenuItem({
  variant,
  className = "",
  onClick,
  children,
  ...props
}) {
  const ctx = React.useContext(DdCtx);
  return /*#__PURE__*/React.createElement("div", _extends({
    role: "menuitem",
    "data-slot": "dropdown-menu-item",
    "data-variant": variant,
    className: cx("cn-dropdown-menu-item", className),
    onClick: e => {
      onClick && onClick(e);
      ctx && ctx.setOpen(false);
    }
  }, props), children);
}
function DropdownMenuLabel({
  className = "",
  children,
  ...props
}) {
  return /*#__PURE__*/React.createElement("div", _extends({
    "data-slot": "dropdown-menu-label",
    className: cx("cn-dropdown-menu-label", className)
  }, props), children);
}
function DropdownMenuSeparator({
  className = "",
  ...props
}) {
  return /*#__PURE__*/React.createElement("div", _extends({
    "data-slot": "dropdown-menu-separator",
    className: cx("cn-dropdown-menu-separator", className)
  }, props));
}
function DropdownMenuShortcut({
  className = "",
  children,
  ...props
}) {
  return /*#__PURE__*/React.createElement("span", _extends({
    "data-slot": "dropdown-menu-shortcut",
    className: cx("cn-dropdown-menu-shortcut", className)
  }, props), children);
}
Object.assign(__ds_scope, { DropdownMenu, DropdownMenuTrigger, DropdownMenuContent, DropdownMenuItem, DropdownMenuLabel, DropdownMenuSeparator, DropdownMenuShortcut });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/navigation/DropdownMenu.jsx", error: String((e && e.message) || e) }); }

// components/navigation/Tabs.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
const cx = (...c) => c.filter(Boolean).join(" ");
const TabsCtx = React.createContext(null);

/** Tabs — segmented tab switcher. Controlled via `value`/`onValueChange`. */
function Tabs({
  value,
  defaultValue,
  onValueChange,
  className = "",
  children,
  ...props
}) {
  const [internal, setInternal] = React.useState(defaultValue);
  const current = value !== undefined ? value : internal;
  const set = v => {
    if (value === undefined) setInternal(v);
    onValueChange && onValueChange(v);
  };
  return /*#__PURE__*/React.createElement("div", _extends({
    "data-slot": "tabs",
    className: cx("cn-tabs", className)
  }, props), /*#__PURE__*/React.createElement(TabsCtx.Provider, {
    value: {
      current,
      set
    }
  }, children));
}
function TabsList({
  className = "",
  children,
  ...props
}) {
  return /*#__PURE__*/React.createElement("div", _extends({
    role: "tablist",
    "data-slot": "tabs-list",
    className: cx("cn-tabs-list", className)
  }, props), children);
}
function TabsTrigger({
  value,
  className = "",
  children,
  ...props
}) {
  const ctx = React.useContext(TabsCtx);
  const active = ctx && ctx.current === value;
  return /*#__PURE__*/React.createElement("button", _extends({
    type: "button",
    role: "tab",
    "aria-selected": !!active,
    "data-slot": "tabs-trigger",
    "data-active": active ? "" : undefined,
    onClick: () => ctx && ctx.set(value),
    className: cx("cn-tabs-trigger", className)
  }, props), children);
}
function TabsContent({
  value,
  className = "",
  children,
  ...props
}) {
  const ctx = React.useContext(TabsCtx);
  if (!ctx || ctx.current !== value) return null;
  return /*#__PURE__*/React.createElement("div", _extends({
    role: "tabpanel",
    "data-slot": "tabs-content",
    className: cx("cn-tabs-content", className)
  }, props), children);
}
Object.assign(__ds_scope, { Tabs, TabsList, TabsTrigger, TabsContent });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/navigation/Tabs.jsx", error: String((e && e.message) || e) }); }

// ui_kits/dashboard/DashboardScreen.jsx
try { (() => {
// Dashboard screen — top nav + stat cards + data table, composing the primitives.
const SD = window.ShadcnSvelteDesignSystem_ce4a1c;
function StatCard({
  title,
  value,
  delta,
  icon
}) {
  const {
    Card,
    CardHeader,
    CardTitle,
    CardContent,
    Icon
  } = SD;
  return /*#__PURE__*/React.createElement(Card, null, /*#__PURE__*/React.createElement(CardHeader, {
    style: {
      flexDirection: "row",
      alignItems: "center",
      justifyContent: "space-between"
    }
  }, /*#__PURE__*/React.createElement(CardTitle, {
    style: {
      fontSize: "var(--text-sm)",
      color: "var(--muted-foreground)",
      fontWeight: 500
    }
  }, title), /*#__PURE__*/React.createElement(Icon, {
    name: icon,
    size: 16,
    style: {
      color: "var(--muted-foreground)"
    }
  })), /*#__PURE__*/React.createElement(CardContent, null, /*#__PURE__*/React.createElement("div", {
    style: {
      fontSize: "var(--text-2xl)",
      fontWeight: 600,
      letterSpacing: "-0.015em"
    }
  }, value), /*#__PURE__*/React.createElement("div", {
    style: {
      fontSize: "var(--text-xs)",
      color: "var(--muted-foreground)",
      marginTop: 4
    }
  }, delta)));
}
function DashboardScreen({
  user,
  onLogout
}) {
  const {
    Button,
    Badge,
    Avatar,
    Icon,
    Input,
    Tabs,
    TabsList,
    TabsTrigger,
    TabsContent,
    Table,
    TableHeader,
    TableBody,
    TableRow,
    TableHead,
    TableCell,
    DropdownMenu,
    DropdownMenuTrigger,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuSeparator,
    Progress
  } = SD;
  const rows = [{
    id: "#3210",
    customer: "Olivia Martin",
    email: "olivia@email.com",
    status: "Fulfilled",
    amount: "$1,999.00"
  }, {
    id: "#3209",
    customer: "Jackson Lee",
    email: "jackson@email.com",
    status: "Processing",
    amount: "$39.00"
  }, {
    id: "#3208",
    customer: "Isabella Nguyen",
    email: "isabella@email.com",
    status: "Fulfilled",
    amount: "$299.00"
  }, {
    id: "#3207",
    customer: "William Kim",
    email: "will@email.com",
    status: "Declined",
    amount: "$99.00"
  }, {
    id: "#3206",
    customer: "Sofia Davis",
    email: "sofia@email.com",
    status: "Processing",
    amount: "$599.00"
  }];
  const statusVariant = {
    Fulfilled: "secondary",
    Processing: "outline",
    Declined: "destructive"
  };
  return /*#__PURE__*/React.createElement("div", {
    style: {
      minHeight: "100%",
      background: "var(--background)",
      color: "var(--foreground)",
      fontFamily: "var(--font-sans)"
    }
  }, /*#__PURE__*/React.createElement("header", {
    style: {
      display: "flex",
      alignItems: "center",
      gap: 16,
      height: 56,
      padding: "0 24px",
      borderBottom: "1px solid var(--border)"
    }
  }, /*#__PURE__*/React.createElement("div", {
    style: {
      display: "flex",
      alignItems: "center",
      gap: 8,
      fontWeight: 600,
      fontFamily: "var(--font-mono)",
      fontSize: 14
    }
  }, /*#__PURE__*/React.createElement("div", {
    style: {
      width: 24,
      height: 24,
      borderRadius: 6,
      background: "var(--primary)",
      display: "grid",
      placeItems: "center"
    }
  }, /*#__PURE__*/React.createElement("img", {
    src: "../../assets/favicon-32x32.png",
    width: "15",
    height: "15",
    alt: ""
  })), "Acme Inc"), /*#__PURE__*/React.createElement("nav", {
    style: {
      display: "flex",
      gap: 20,
      marginLeft: 12,
      fontSize: 14
    }
  }, /*#__PURE__*/React.createElement("a", {
    href: "#",
    style: {
      color: "var(--foreground)",
      textDecoration: "none",
      fontWeight: 500
    }
  }, "Overview"), /*#__PURE__*/React.createElement("a", {
    href: "#",
    style: {
      color: "var(--muted-foreground)",
      textDecoration: "none"
    }
  }, "Customers"), /*#__PURE__*/React.createElement("a", {
    href: "#",
    style: {
      color: "var(--muted-foreground)",
      textDecoration: "none"
    }
  }, "Products"), /*#__PURE__*/React.createElement("a", {
    href: "#",
    style: {
      color: "var(--muted-foreground)",
      textDecoration: "none"
    }
  }, "Settings")), /*#__PURE__*/React.createElement("div", {
    style: {
      marginLeft: "auto",
      display: "flex",
      alignItems: "center",
      gap: 12
    }
  }, /*#__PURE__*/React.createElement("div", {
    style: {
      position: "relative",
      width: 200
    }
  }, /*#__PURE__*/React.createElement(Icon, {
    name: "search",
    size: 15,
    style: {
      position: "absolute",
      left: 10,
      top: "50%",
      transform: "translateY(-50%)",
      color: "var(--muted-foreground)"
    }
  }), /*#__PURE__*/React.createElement(Input, {
    placeholder: "Search\u2026",
    style: {
      paddingLeft: 30,
      height: 32
    }
  })), /*#__PURE__*/React.createElement(Button, {
    variant: "ghost",
    size: "icon"
  }, /*#__PURE__*/React.createElement(Icon, {
    name: "bell"
  })), /*#__PURE__*/React.createElement(DropdownMenu, null, /*#__PURE__*/React.createElement(DropdownMenuTrigger, null, /*#__PURE__*/React.createElement("button", {
    style: {
      border: 0,
      background: "transparent",
      padding: 0,
      cursor: "pointer",
      borderRadius: 9999
    }
  }, /*#__PURE__*/React.createElement(Avatar, {
    src: "../../assets/avatars/01.png",
    fallback: "CN"
  }))), /*#__PURE__*/React.createElement(DropdownMenuContent, {
    align: "end"
  }, /*#__PURE__*/React.createElement(DropdownMenuLabel, null, user), /*#__PURE__*/React.createElement(DropdownMenuSeparator, null), /*#__PURE__*/React.createElement(DropdownMenuItem, null, /*#__PURE__*/React.createElement(Icon, {
    name: "user"
  }), " Profile"), /*#__PURE__*/React.createElement(DropdownMenuItem, null, /*#__PURE__*/React.createElement(Icon, {
    name: "settings"
  }), " Settings"), /*#__PURE__*/React.createElement(DropdownMenuSeparator, null), /*#__PURE__*/React.createElement(DropdownMenuItem, {
    variant: "destructive",
    onClick: onLogout
  }, /*#__PURE__*/React.createElement(Icon, {
    name: "log-out"
  }), " Log out"))))), /*#__PURE__*/React.createElement("main", {
    style: {
      padding: 24,
      display: "flex",
      flexDirection: "column",
      gap: 24,
      maxWidth: 1100,
      margin: "0 auto"
    }
  }, /*#__PURE__*/React.createElement("div", {
    style: {
      display: "flex",
      alignItems: "center",
      justifyContent: "space-between"
    }
  }, /*#__PURE__*/React.createElement("div", null, /*#__PURE__*/React.createElement("h1", {
    style: {
      fontSize: "var(--text-2xl)",
      fontWeight: 600,
      margin: 0,
      letterSpacing: "-0.015em"
    }
  }, "Dashboard"), /*#__PURE__*/React.createElement("p", {
    style: {
      color: "var(--muted-foreground)",
      fontSize: "var(--text-sm)",
      margin: "4px 0 0"
    }
  }, "Welcome back \u2014 here's what's happening.")), /*#__PURE__*/React.createElement(Button, null, /*#__PURE__*/React.createElement(Icon, {
    name: "plus"
  }), " Add product")), /*#__PURE__*/React.createElement("div", {
    style: {
      display: "grid",
      gridTemplateColumns: "repeat(4, 1fr)",
      gap: 16
    }
  }, /*#__PURE__*/React.createElement(StatCard, {
    title: "Total Revenue",
    value: "$45,231.89",
    delta: "+20.1% from last month",
    icon: "credit-card"
  }), /*#__PURE__*/React.createElement(StatCard, {
    title: "Subscriptions",
    value: "+2,350",
    delta: "+180.1% from last month",
    icon: "user"
  }), /*#__PURE__*/React.createElement(StatCard, {
    title: "Sales",
    value: "+12,234",
    delta: "+19% from last month",
    icon: "star"
  }), /*#__PURE__*/React.createElement(StatCard, {
    title: "Active Now",
    value: "+573",
    delta: "+201 since last hour",
    icon: "bell"
  })), /*#__PURE__*/React.createElement(Tabs, {
    defaultValue: "orders"
  }, /*#__PURE__*/React.createElement(TabsList, null, /*#__PURE__*/React.createElement(TabsTrigger, {
    value: "orders"
  }, "Recent orders"), /*#__PURE__*/React.createElement(TabsTrigger, {
    value: "analytics"
  }, "Analytics")), /*#__PURE__*/React.createElement(TabsContent, {
    value: "orders"
  }, /*#__PURE__*/React.createElement("div", {
    style: {
      border: "1px solid var(--border)",
      borderRadius: "var(--radius-xl)",
      overflow: "hidden",
      marginTop: 12
    }
  }, /*#__PURE__*/React.createElement(Table, null, /*#__PURE__*/React.createElement(TableHeader, null, /*#__PURE__*/React.createElement(TableRow, null, /*#__PURE__*/React.createElement(TableHead, null, "Order"), /*#__PURE__*/React.createElement(TableHead, null, "Customer"), /*#__PURE__*/React.createElement(TableHead, null, "Status"), /*#__PURE__*/React.createElement(TableHead, {
    style: {
      textAlign: "right"
    }
  }, "Amount"))), /*#__PURE__*/React.createElement(TableBody, null, rows.map(r => /*#__PURE__*/React.createElement(TableRow, {
    key: r.id
  }, /*#__PURE__*/React.createElement(TableCell, {
    style: {
      fontFamily: "var(--font-mono)"
    }
  }, r.id), /*#__PURE__*/React.createElement(TableCell, null, /*#__PURE__*/React.createElement("div", {
    style: {
      fontWeight: 500
    }
  }, r.customer), /*#__PURE__*/React.createElement("div", {
    style: {
      color: "var(--muted-foreground)",
      fontSize: "var(--text-xs)"
    }
  }, r.email)), /*#__PURE__*/React.createElement(TableCell, null, /*#__PURE__*/React.createElement(Badge, {
    variant: statusVariant[r.status]
  }, r.status)), /*#__PURE__*/React.createElement(TableCell, {
    style: {
      textAlign: "right",
      fontWeight: 500
    }
  }, r.amount))))))), /*#__PURE__*/React.createElement(TabsContent, {
    value: "analytics"
  }, /*#__PURE__*/React.createElement("div", {
    style: {
      marginTop: 12,
      display: "grid",
      gap: 16
    }
  }, [["Storage used", 72], ["Monthly quota", 41], ["Team seats", 90]].map(([label, val]) => /*#__PURE__*/React.createElement("div", {
    key: label,
    style: {
      display: "grid",
      gap: 6
    }
  }, /*#__PURE__*/React.createElement("div", {
    style: {
      display: "flex",
      justifyContent: "space-between",
      fontSize: "var(--text-sm)"
    }
  }, /*#__PURE__*/React.createElement("span", null, label), /*#__PURE__*/React.createElement("span", {
    style: {
      color: "var(--muted-foreground)"
    }
  }, val, "%")), /*#__PURE__*/React.createElement(Progress, {
    value: val
  }))))))));
}
window.DashboardScreen = DashboardScreen;
})(); } catch (e) { __ds_ns.__errors.push({ path: "ui_kits/dashboard/DashboardScreen.jsx", error: String((e && e.message) || e) }); }

// ui_kits/dashboard/LoginScreen.jsx
try { (() => {
// Login screen — faithful to shadcn-svelte login-02 block.
const S = window.ShadcnSvelteDesignSystem_ce4a1c;
function LoginScreen({
  onLogin
}) {
  const {
    Button,
    Input,
    Label,
    Separator,
    Icon
  } = S;
  const [email, setEmail] = React.useState("");
  return /*#__PURE__*/React.createElement("div", {
    style: {
      minHeight: "100%",
      display: "grid",
      placeItems: "center",
      background: "var(--surface)",
      padding: 24,
      fontFamily: "var(--font-sans)",
      color: "var(--foreground)"
    }
  }, /*#__PURE__*/React.createElement("form", {
    onSubmit: e => {
      e.preventDefault();
      onLogin(email || "m@example.com");
    },
    style: {
      display: "flex",
      flexDirection: "column",
      gap: 24,
      width: 360
    }
  }, /*#__PURE__*/React.createElement("div", {
    style: {
      display: "flex",
      flexDirection: "column",
      alignItems: "center",
      gap: 4,
      textAlign: "center"
    }
  }, /*#__PURE__*/React.createElement("div", {
    style: {
      width: 40,
      height: 40,
      borderRadius: 10,
      background: "var(--primary)",
      display: "grid",
      placeItems: "center",
      marginBottom: 8
    }
  }, /*#__PURE__*/React.createElement("img", {
    src: "../../assets/favicon-32x32.png",
    width: "24",
    height: "24",
    alt: ""
  })), /*#__PURE__*/React.createElement("h1", {
    style: {
      fontSize: "var(--text-2xl)",
      fontWeight: 700,
      margin: 0,
      letterSpacing: "-0.015em"
    }
  }, "Login to your account"), /*#__PURE__*/React.createElement("p", {
    style: {
      color: "var(--muted-foreground)",
      fontSize: "var(--text-sm)",
      margin: 0
    }
  }, "Enter your email below to login to your account")), /*#__PURE__*/React.createElement("div", {
    style: {
      display: "grid",
      gap: 6
    }
  }, /*#__PURE__*/React.createElement(Label, {
    htmlFor: "email"
  }, "Email"), /*#__PURE__*/React.createElement(Input, {
    id: "email",
    type: "email",
    placeholder: "m@example.com",
    value: email,
    onChange: e => setEmail(e.target.value),
    required: true
  })), /*#__PURE__*/React.createElement("div", {
    style: {
      display: "grid",
      gap: 6
    }
  }, /*#__PURE__*/React.createElement("div", {
    style: {
      display: "flex",
      alignItems: "center"
    }
  }, /*#__PURE__*/React.createElement(Label, {
    htmlFor: "password"
  }, "Password"), /*#__PURE__*/React.createElement("a", {
    href: "#",
    style: {
      marginLeft: "auto",
      fontSize: "var(--text-sm)",
      color: "var(--foreground)",
      textUnderlineOffset: 4
    }
  }, "Forgot your password?")), /*#__PURE__*/React.createElement(Input, {
    id: "password",
    type: "password",
    defaultValue: "password",
    required: true
  })), /*#__PURE__*/React.createElement(Button, {
    type: "submit"
  }, "Login"), /*#__PURE__*/React.createElement("div", {
    style: {
      display: "flex",
      alignItems: "center",
      gap: 12,
      color: "var(--muted-foreground)",
      fontSize: "var(--text-sm)"
    }
  }, /*#__PURE__*/React.createElement(Separator, {
    style: {
      flex: 1
    }
  }), " Or continue with ", /*#__PURE__*/React.createElement(Separator, {
    style: {
      flex: 1
    }
  })), /*#__PURE__*/React.createElement(Button, {
    variant: "outline",
    type: "button",
    onClick: () => onLogin("octocat@github.com")
  }, /*#__PURE__*/React.createElement("svg", {
    width: "16",
    height: "16",
    viewBox: "0 0 24 24",
    fill: "currentColor",
    "aria-hidden": "true"
  }, /*#__PURE__*/React.createElement("path", {
    d: "M12 .297c-6.63 0-12 5.373-12 12 0 5.303 3.438 9.8 8.205 11.385.6.113.82-.258.82-.577 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61C4.422 18.07 3.633 17.7 3.633 17.7c-1.087-.744.084-.729.084-.729 1.205.084 1.838 1.236 1.838 1.236 1.07 1.835 2.809 1.305 3.495.998.108-.776.417-1.305.76-1.605-2.665-.3-5.466-1.332-5.466-5.93 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23.96-.267 1.98-.399 3-.405 1.02.006 2.04.138 3 .405 2.28-1.552 3.285-1.23 3.285-1.23.645 1.653.24 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81 2.22 0 1.606-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 22.092 24 17.592 24 12.297c0-6.627-5.373-12-12-12"
  })), "Login with GitHub"), /*#__PURE__*/React.createElement("p", {
    style: {
      textAlign: "center",
      fontSize: "var(--text-sm)",
      color: "var(--muted-foreground)",
      margin: 0
    }
  }, "Don't have an account? ", /*#__PURE__*/React.createElement("a", {
    href: "#",
    style: {
      color: "var(--foreground)",
      textUnderlineOffset: 4,
      textDecoration: "underline"
    }
  }, "Sign up"))));
}
window.LoginScreen = LoginScreen;
})(); } catch (e) { __ds_ns.__errors.push({ path: "ui_kits/dashboard/LoginScreen.jsx", error: String((e && e.message) || e) }); }

__ds_ns.Badge = __ds_scope.Badge;

__ds_ns.Button = __ds_scope.Button;

__ds_ns.Kbd = __ds_scope.Kbd;

__ds_ns.KbdGroup = __ds_scope.KbdGroup;

__ds_ns.Spinner = __ds_scope.Spinner;

__ds_ns.Toggle = __ds_scope.Toggle;

__ds_ns.Avatar = __ds_scope.Avatar;

__ds_ns.Card = __ds_scope.Card;

__ds_ns.CardHeader = __ds_scope.CardHeader;

__ds_ns.CardTitle = __ds_scope.CardTitle;

__ds_ns.CardDescription = __ds_scope.CardDescription;

__ds_ns.CardContent = __ds_scope.CardContent;

__ds_ns.CardFooter = __ds_scope.CardFooter;

__ds_ns.Progress = __ds_scope.Progress;

__ds_ns.Separator = __ds_scope.Separator;

__ds_ns.Skeleton = __ds_scope.Skeleton;

__ds_ns.Table = __ds_scope.Table;

__ds_ns.TableHeader = __ds_scope.TableHeader;

__ds_ns.TableBody = __ds_scope.TableBody;

__ds_ns.TableFooter = __ds_scope.TableFooter;

__ds_ns.TableRow = __ds_scope.TableRow;

__ds_ns.TableHead = __ds_scope.TableHead;

__ds_ns.TableCell = __ds_scope.TableCell;

__ds_ns.TableCaption = __ds_scope.TableCaption;

__ds_ns.Alert = __ds_scope.Alert;

__ds_ns.AlertTitle = __ds_scope.AlertTitle;

__ds_ns.AlertDescription = __ds_scope.AlertDescription;

__ds_ns.Dialog = __ds_scope.Dialog;

__ds_ns.DialogHeader = __ds_scope.DialogHeader;

__ds_ns.DialogTitle = __ds_scope.DialogTitle;

__ds_ns.DialogDescription = __ds_scope.DialogDescription;

__ds_ns.DialogFooter = __ds_scope.DialogFooter;

__ds_ns.Tooltip = __ds_scope.Tooltip;

__ds_ns.Checkbox = __ds_scope.Checkbox;

__ds_ns.Input = __ds_scope.Input;

__ds_ns.Label = __ds_scope.Label;

__ds_ns.RadioGroup = __ds_scope.RadioGroup;

__ds_ns.RadioGroupItem = __ds_scope.RadioGroupItem;

__ds_ns.Select = __ds_scope.Select;

__ds_ns.Slider = __ds_scope.Slider;

__ds_ns.Switch = __ds_scope.Switch;

__ds_ns.Textarea = __ds_scope.Textarea;

__ds_ns.ICON_PATHS = __ds_scope.ICON_PATHS;

__ds_ns.Icon = __ds_scope.Icon;

__ds_ns.Accordion = __ds_scope.Accordion;

__ds_ns.AccordionItem = __ds_scope.AccordionItem;

__ds_ns.AccordionTrigger = __ds_scope.AccordionTrigger;

__ds_ns.AccordionContent = __ds_scope.AccordionContent;

__ds_ns.Breadcrumb = __ds_scope.Breadcrumb;

__ds_ns.BreadcrumbList = __ds_scope.BreadcrumbList;

__ds_ns.BreadcrumbItem = __ds_scope.BreadcrumbItem;

__ds_ns.BreadcrumbLink = __ds_scope.BreadcrumbLink;

__ds_ns.BreadcrumbPage = __ds_scope.BreadcrumbPage;

__ds_ns.BreadcrumbSeparator = __ds_scope.BreadcrumbSeparator;

__ds_ns.DropdownMenu = __ds_scope.DropdownMenu;

__ds_ns.DropdownMenuTrigger = __ds_scope.DropdownMenuTrigger;

__ds_ns.DropdownMenuContent = __ds_scope.DropdownMenuContent;

__ds_ns.DropdownMenuItem = __ds_scope.DropdownMenuItem;

__ds_ns.DropdownMenuLabel = __ds_scope.DropdownMenuLabel;

__ds_ns.DropdownMenuSeparator = __ds_scope.DropdownMenuSeparator;

__ds_ns.DropdownMenuShortcut = __ds_scope.DropdownMenuShortcut;

__ds_ns.Tabs = __ds_scope.Tabs;

__ds_ns.TabsList = __ds_scope.TabsList;

__ds_ns.TabsTrigger = __ds_scope.TabsTrigger;

__ds_ns.TabsContent = __ds_scope.TabsContent;

})();
