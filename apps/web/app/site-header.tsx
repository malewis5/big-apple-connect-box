import Link from "next/link";
import { getContent } from "./content";
import { SITE_NAME } from "./site";
import { TrackedLink } from "./tracked-link";

// The top of every page. It is drawn once, in layout.tsx, so it stays put
// when you move from one page to another. It holds the site's two doors:
// "Notify me" for residents (the main one, so it's the only button) and
// "For partners" for the city, advocates, funders and researchers.
export async function SiteHeader() {
  const { header } = await getContent();
  return (
    <>
      {/* The first stop for the Tab key: jump past the header. Hidden until focused. */}
      <a
        href="#main"
        className="sr-only focus:not-sr-only focus:fixed focus:top-4 focus:left-4 focus:z-50 focus:rounded-full focus:bg-zinc-900 focus:px-4 focus:py-2 focus:text-sm focus:font-medium focus:text-white"
      >
        {header.skip}
      </a>
      {/* One row at every width. On a phone there's only room for the name and
          the button; partners find their section by scrolling or in the footer. */}
      <header className="mx-auto flex w-full max-w-6xl items-center justify-between gap-3 px-6 py-3 text-zinc-900">
        <Link href="/" className="flex min-h-11 min-w-0 items-center gap-2.5 text-sm font-semibold tracking-tight">
          <span className="size-2.5 shrink-0 rounded-full bg-emerald-500 shadow-[0_0_0_4px_rgb(16_185_129/0.15)]" />
          {SITE_NAME}
        </Link>
        <nav className="flex shrink-0 items-center gap-2 sm:gap-6">
          <TrackedLink
            href="/#partners"
            event="Partner clicked"
            place="header"
            className="hidden min-h-11 items-center text-sm font-medium text-zinc-600 transition hover:text-zinc-900 sm:flex"
          >
            {header.partners}
          </TrackedLink>
          <TrackedLink
            href="/#updates"
            event="Get updates clicked"
            place="header"
            className="flex min-h-11 items-center rounded-full bg-zinc-900 px-4 py-2 text-sm font-medium whitespace-nowrap text-white transition hover:bg-zinc-700"
          >
            {header.notify}
          </TrackedLink>
        </nav>
      </header>
    </>
  );
}
