# site/

Static landing page + schema host for `mclip.dev`. Built with Astro 5.

## Local development

```sh
cd site
npm install
npm run dev          # http://localhost:4321
npm run build        # outputs to site/dist/
npm run preview      # serve the built output
```

## Deployment

Hosted on Vercel. Setup notes:

- The Vercel project is linked at the **repo root**, not the `site/` subdirectory.
- The Vercel project's **Root Directory** setting is `site/` — Vercel auto-detects Astro and uses `npm run build` → `dist/`.
- `vercel deploy` runs from the repo root.

## What lives here

- `src/pages/index.astro` — the landing page.
- `src/layouts/Base.astro` — HTML shell, shared head/footer.
- `src/styles/global.css` — single global stylesheet.
- `public/schemas/config/v0.json` — JSON Schema (Draft 2020-12) for the MCLIP config file, cited from `profile-v0.md` §13.3 as `https://mclip.dev/schemas/config/v0.json`.
- `public/favicon.svg` — favicon.

## Schema-update workflow

When `profile-v0.md` §13.3 changes, hand-update `public/schemas/config/v0.json` to match. A future improvement is to auto-generate the schema from the normative spec; not in scope for v0 of this site.
