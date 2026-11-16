# The website

The site at [bigappleconnect.com](https://www.bigappleconnect.com), in
[Next.js](https://nextjs.org). The home page makes the case for checking
NYCHA internet at home, shows the University of Chicago precedent, and asks
partners to get in touch and residents to sign up for updates.
For the project as a whole, start with the [main README](../../README.md).

From the repository root:

```sh
pnpm install
pnpm dev      # run the website at http://localhost:3000
```

## Words

Everything people read is in [`messages/en.json`](messages/en.json), loaded
through [next-intl](https://next-intl.dev) by `app/content.ts`. The site has
one language and no locale in its addresses, so every page stays static and
no cookie is set. [`AGENTS.md`](AGENTS.md) has the rules for writing them, and
`i18n/request.ts` explains how to add a language.

## Email updates

The home page's signup uses double opt-in through [Resend](https://resend.com):
the form emails a link to `/confirm#<code>`, and the address is added to
Resend only when the person opens it and presses "Confirm my email". Nothing
is stored before that; the code carries the address, signed. The code is in
`app/signup/`.

Set these in Vercel (Production and Preview):

| Variable | What it is |
|---|---|
| `RESEND_API_KEY` | A Resend API key with sending and contacts access |
| `SIGNUP_SECRET` | Any long random string, such as `openssl rand -base64 32`. Changing it breaks links already sent |
| `RESEND_SEGMENT_ID` | Optional. The segment new contacts join |
| `SIGNUP_FROM` | Optional. The From line. Defaults to `Big Apple Connect Box <hello@bigappleconnect.com>` |

Under `pnpm dev` without `RESEND_API_KEY`, nothing is sent: the confirmation
link is printed in the terminal, and confirming prints what would be saved.
In production without it, the form says signups aren't working and points to
the email address.

In Resend, once:

1. Add and verify the domain `bigappleconnect.com`. Resend gives a DKIM
   record (`resend._domainkey`) and MX and SPF records on the `send`
   subdomain; add them in Vercel DNS. They don't clash with Google's.
2. Under Audience, Properties, create a property named `role` of type string.
   It holds the answer to "I am a...": `resident`, `advocate`, `government`,
   `press` or `other`.
3. Optionally create a segment, such as "Updates", and set `RESEND_SEGMENT_ID`.

Send updates as Resend Broadcasts; they add the unsubscribe link themselves.

## Analytics events

Besides page views, the site records these custom events in Vercel Web
Analytics (Pro or Enterprise plans only; on Hobby they record nothing):
`Get updates clicked`, `Watch clicked`, `Partner clicked`, `Partner email clicked` and
`GitHub clicked` (each with a `place`), `Video played`, and from the server,
`Signup requested` and `Signup confirmed` (each with a `role`).

## The 3D model and frames

The box's 3D model is made by a script, not by hand. With
[Blender](https://www.blender.org) installed, rebuild it and compress it from
the repository root:

```sh
/Applications/Blender.app/Contents/MacOS/Blender -b -P design/box/build_box.py
npx @gltf-transform/cli meshopt apps/web/public/box/box.glb apps/web/public/box/box.glb
```

The website doesn't run the 3D scene for visitors. It plays image frames
rendered from it, so the story scrolls smoothly on any phone and connection.
The scrolling story (`app/box-story.tsx`) is off the home page until the box
can be had; it's kept so it can come back.
After changing the model or the scene in `app/capture/`, render the frames
again (needs Google Chrome):

```sh
pnpm --filter web build:frames
```
