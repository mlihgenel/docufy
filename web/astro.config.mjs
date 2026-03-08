import { defineConfig } from "astro/config";
import sitemap from "@astrojs/sitemap";

const site = process.env.SITE_URL ?? "https://mlihgenel.github.io";
const base = process.env.BASE_PATH ?? "/";

export default defineConfig({
  site,
  base,
  output: "static",
  trailingSlash: "always",
  integrations: [sitemap()]
});
