import { defineConfig } from "astro/config";

export default defineConfig({
  site: "https://mclip.dev",
  trailingSlash: "never",
  build: {
    format: "file",
  },
});
