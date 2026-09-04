# shadcn-svelte Design System

A faithful clone of the design tokens and design sensibility of **[shadcn-svelte](https://www.shadcn-svelte.com)** — the community Svelte port of shadcn/ui. This system reproduces the default **"Vega" (Neutral)** theme exactly: the same OKLCH token values, the same Geist type, the same radii, and the same restrained, hairline-first component styling.

> **This is not a component library. It is how you build your component library.** shadcn-svelte is *skinnable* — components receive their look from a single set of CSS custom properties. This design system preserves that property: **retheme the entire system by editing `tokens/colors.css` (and friends) in one place.** Everything else — 27 component families, cards, the UI kit — reads from those tokens.

## Sources

Built by reading the official codebase (read-only, mounted locally):
- **Codebase:** `shadcn-svelte-main/` — the monorepo. Tokens + component styles were lifted verbatim from `docs/src/app.css` (`:root` / `.dark`) and `docs/src/lib/registry/styles/style-vega.css` (the default preset). Component APIs from `docs/src/lib/registry/ui/*`.
- **Site:** https://www.shadcn-svelte.com · **Repo:** https://github.com/huntabyte/shadcn-svelte
- **Upstream:** shadcn/ui (https://ui.shadcn.com), Bits UI (headless behavior), Lucide (icons).

Nothing here is invented. Where the source uses Tailwind `@apply` on semantic `.cn-*` classes, those classes were hand-translated into plain, token-driven CSS in `components.css` — the same architecture, minus the Tailwind build step.

---

## CONTENT FUNDAMENTALS

**Voice.** Confident, technical, plainspoken. Documentation addresses the developer directly as **"you"** and speaks of the system as **"we"** ("If a component does not exist, we bring it in…"). Short declaratives, occasional bolded thesis lines ("**This is not a component library.**"), and italicized asides that reframe a point (*"You are not learning different APIs for every new component."*).

**Casing.** Sentence case everywhere — headings, buttons, menu items, card titles ("Create project", "Login to your account", not Title Case). The brand name is always lowercase and hyphenated: **shadcn-svelte**.

**UI copy.** Terse and literal. Buttons are verbs ("Deploy", "Login", "Add product", "Continue"). Descriptions are one calm sentence ("Deploy your new project in one-click.", "This action cannot be undone."). Placeholders use `m@example.com`. Empty/among sample data the canonical demo names recur: Olivia Martin, Jackson Lee, Isabella Nguyen, William Kim, Sofia Davis.

**Emoji.** None. The brand never uses emoji in UI or docs. Icons do that job.

**Vibe.** Engineer-to-engineer. Calm, precise, a little opinionated about craft ("Beautiful Defaults", "Open Code"). Never salesy, never playful-cute.

---

## VISUAL FOUNDATIONS

**Overall feel.** Quiet, neutral, high-contrast, generous whitespace. The design gets out of the way of content. Pure black-and-white core with color reserved almost entirely for data (charts) and destructive states.

**Color.** A pure achromatic (chroma = 0) grayscale ramp in **OKLCH**. Light mode: white background, near-black foreground/primary, `oklch(0.97)` secondary/muted/accent, `oklch(0.922)` borders. The **only** chromatic tokens are `--destructive` (a red, `oklch(0.577 0.245 27)`) and the `--chart-1…5` ramp (Tailwind blues). Dark mode inverts to `oklch(0.145)` background with light primary. There is no brand accent hue — neutrality is the brand.

**Type.** **Geist** (variable sans) for everything UI + headings; **Geist Mono** for code, shortcuts, and the wordmark. Body workhorse is **14px (text-sm)**. Weights: 400 body, **500 medium** for nearly all UI chrome (buttons, labels, badges, titles, table headers), 600–700 for display headings. Tight tracking (-0.015em) on large headings.

**Spacing.** Tailwind 4px base unit. Compact by default — 6–10px gaps dominate inside controls; cards pad 24px. Control heights: 36px (default), 32px (sm), 40px (lg).

**Radius.** Base **0.625rem / 10px**, with a derived scale: sm 6 · md 8 · lg 10 · xl 14 · 2xl 18 · 4xl (pill). Buttons/inputs use md (8), cards & dialogs use xl (14), badges are full-pill.

**Borders & elevation.** Hairline-first. 1px borders in `--border`. Cards use a **1px ring (`foreground/10`) + `shadow-xs`** rather than a heavy drop shadow. Popovers/menus: `shadow-md` + ring. Dialogs: `shadow-lg`. Shadows are soft, low-opacity black. No neumorphism, no glow.

**Backgrounds.** Flat solid fills — `--background` (page), `--surface`/`--card` for panels. **No gradients** on chrome (a barely-there `section-soft` background gradient exists on marketing surfaces only). No textures, no illustrations, no full-bleed hero imagery in the system itself.

**Transparency & blur.** Used sparingly and purposefully: modal/sheet overlays are `black/10` with an optional `backdrop-blur-xs`; hover states use color-mix alpha (e.g. `primary/80`); a "translucent menu" variant uses `popover/70` + heavy backdrop blur.

**Animation.** Subtle and fast. Enter/exit ~100–200ms (`fade-in`, `zoom-in-95`, directional `slide-in-2`). Accordions animate height. No bounce, no springy overshoot, no infinite decorative motion. Reduced-motion friendly.

**Hover / press.** Hover = a small opacity/background shift (solid buttons → `primary/80`; ghost/outline → `muted` fill). **Press** = the whole button nudges down 1px (`active:translate-y-px`) and on touch drops to `opacity-60`. Focus = a 3px `ring/50` halo plus a `ring`-colored border.

**Cards.** Rounded-xl (14px), `--card` fill, hairline ring + `shadow-xs`, 24px vertical padding, 24px gap between header/content/footer. Clean and unadorned.

**Imagery.** The system ships almost no imagery — just user avatars (warm, natural photos) and neutral placeholders. Product screenshots in the wild are clean, light, UI-forward.

---

## ICONOGRAPHY

- **Icon set:** **[Lucide](https://lucide.dev)** is the canonical shadcn-svelte icon system (`@lucide/svelte`). The docs site also pulls Tabler / Phosphor / Hugeicons for specific examples, but **Lucide is the default** and what components use.
- **Style:** 24×24 viewBox, **currentColor stroke, ~2px stroke width, round caps & joins, no fill.** Icons inherit text color and size (commonly 16px inside controls, matching `[&_svg]:size-4`).
- **Delivery here:** the `Icon` component (`components/icon/`) bundles a curated Lucide subset as inline SVG so glyphs inherit `currentColor` and scale via a `size` prop. For any glyph outside the subset, copy the raw SVG from lucide.dev — do not hand-draw. Names available are listed in `Icon.d.ts`.
- **Emoji / unicode as icons:** never. Icons are always Lucide SVGs. The single exception in this system is the inline GitHub mark on the login button (a brand glyph, copied verbatim from the source block).
- **Brand mark:** shadcn-svelte ships **no dedicated logotype**. The only mark is the favicon — a rounded black square — copied to `assets/favicon-32x32.png` and `assets/apple-touch-icon.png`. Everywhere a logo would go, set the name **shadcn-svelte** in Geist Mono. No logo was reconstructed or invented.

---

## Components

All components live under `components/<group>/` as `<Name>.jsx` + `<Name>.d.ts`, exported on the runtime namespace `window.ShadcnSvelteDesignSystem_ce4a1c`. They emit semantic `.cn-*` classes styled by `components.css`.

- **icon/** — `Icon` (Lucide glyph, inline SVG), plus the `ICON_PATHS` map.
- **actions/** — `Button`, `Badge`, `Toggle`, `Kbd`, `KbdGroup`, `Spinner`.
- **forms/** — `Input`, `Textarea`, `Label`, `Checkbox`, `RadioGroup` (+ `RadioGroupItem`), `Switch`, `Select`, `Slider`.
- **data-display/** — `Card` (+ `CardHeader`, `CardTitle`, `CardDescription`, `CardContent`, `CardFooter`), `Avatar`, `Table` (+ `TableHeader`, `TableBody`, `TableFooter`, `TableRow`, `TableHead`, `TableCell`, `TableCaption`), `Separator`, `Skeleton`, `Progress`.
- **feedback/** — `Alert` (+ `AlertTitle`, `AlertDescription`), `Tooltip`, `Dialog` (+ `DialogHeader`, `DialogTitle`, `DialogDescription`, `DialogFooter`).
- **navigation/** — `Tabs` (+ `TabsList`, `TabsTrigger`, `TabsContent`), `Breadcrumb` (+ `BreadcrumbList`, `BreadcrumbItem`, `BreadcrumbLink`, `BreadcrumbPage`, `BreadcrumbSeparator`), `Accordion` (+ `AccordionItem`, `AccordionTrigger`, `AccordionContent`), `DropdownMenu` (+ `DropdownMenuTrigger`, `DropdownMenuContent`, `DropdownMenuItem`, `DropdownMenuLabel`, `DropdownMenuSeparator`, `DropdownMenuShortcut`).

Each component directory has a `*.prompt.md` (usage) and a `@dsCard`-tagged card HTML shown in the Design System tab.

### Scope vs. the full shadcn-svelte registry

The source registry defines **~57 component families**. This system implements the **27 most-used** primitives above with high fidelity (exact token values, exact `.cn-*` styling). The remaining families are **not yet built** and are the natural next iteration:

> **Not yet built:** AlertDialog, AspectRatio, ButtonGroup, Calendar / RangeCalendar, Carousel, Chart, Collapsible, Command, ContextMenu, DataTable, Drawer, Empty, Field, Form, HoverCard, InputGroup, InputOTP, Item, Menubar, NavigationMenu, Pagination, Popover, Resizable, ScrollArea, Select (rich/listbox variant), Sheet, Sidebar, Sonner (toast), Tabs (line variant), ToggleGroup.

The tokens and `components.css` architecture already cover these — adding them is mostly wiring markup to existing `.cn-*` classes (the full class definitions for every family are in the source `style-vega.css`).

---

## UI kits

- **ui_kits/dashboard/** — an interactive admin product: a login screen (ported from the `login-02` block) that flows into a dashboard (top nav, stat cards, tabbed orders table + analytics). Composes the primitives only; nothing re-implemented.

---

## Root manifest / index

- **`styles.css`** — the single entry point consumers link. `@import`s only.
- **`tokens/`** — `colors.css` (OKLCH palette + `.dark`), `typography.css` (Geist stacks + size scale), `radius.css` (radius/spacing/shadow), `fonts.css` (Geist webfonts).
- **`components.css`** — token-driven `.cn-*` slot styles (translated from `style-vega.css`).
- **`components/`** — the 27 React component families (see above).
- **`guidelines/`** — foundation specimen cards (Colors, Type, Spacing, Brand).
- **`ui_kits/dashboard/`** — the interactive product recreation.
- **`assets/`** — favicon/apple-touch-icon (the only brand mark), demo avatars, placeholders.
- **`SKILL.md`** — Agent-Skills-compatible entry point.

## Substitutions & caveats

- **Fonts:** Geist + Geist Mono are loaded from **Google Fonts** (`tokens/fonts.css`) rather than the source's `@fontsource` variable packages. **Same typefaces, not a visual substitution** — but if you want the exact variable binaries self-hosted, drop them in and swap the `@import`.
- **Coverage:** 27 of ~57 families built (see Scope). This is the deliberate first cut, not the ceiling.
