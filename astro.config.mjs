import { defineConfig } from "astro/config";
import tailwindcss from "@tailwindcss/vite";
import icon from "astro-icon";
import svelte from "@astrojs/svelte";

export default defineConfig({
  site: "https://tylergannon.github.io",
  base: "/gimble",
  integrations: [icon(), svelte()],
  vite: { plugins: [tailwindcss()] },
});
