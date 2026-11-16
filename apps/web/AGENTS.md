<!-- BEGIN:nextjs-agent-rules -->

## This is NOT the Next.js you know

This version has breaking changes — APIs, conventions, and file structure may all differ from your training data. Read the relevant guide in `node_modules/next/dist/docs/` (resolved from this file's directory; in monorepos the `next` package may not be visible from the repo root) before writing any code. Heed deprecation notices.

This block is written and re-added by `next dev` — verify at `node_modules/next/dist/server/lib/generate-agent-files.js`. Removing it from a diff only re-creates the uncommitted change; committing it with your work keeps the tree clean.

<!-- END:nextjs-agent-rules -->

## Writing the site's words

The words people read are in `messages/en.json`, read through next-intl by
`app/content.ts`. Facts that aren't words (addresses, dates, files, keys) are
in `app/site.ts`. When you write or change the words:

- Set that page's date in `app/site.ts` to today: `HOME_UPDATED` for the home
  page, `PAGE_UPDATED` for About, Contact and Privacy. The sitemap reports it
  to search engines.
- Lists (steps, questions, page sections) are objects keyed by a stable name;
  their order in the file is their order on the page. Add a key, don't renumber.
- Don't hard-code words in components. Pass resolved strings to client
  components as props; never send the whole catalog to the browser.
- Every Chicago figure must link to the University of Chicago source that
  states it, and the page must keep saying the project isn't part of that work.
- A new language needs its own `messages/<locale>.json` with the same keys,
  reviewed by a native speaker, and the routing steps in `i18n/request.ts`.
- Don't name a thing and then describe it after a comma: "Big Apple Connect,
  the city's program" or "Netrics, the open-source software used in...". The
  other order is out too: "its maker, Matt Lewis". Use the name alone, or say
  what it is in a sentence of its own: "It follows Netrics. Netrics is the
  software used in...".
