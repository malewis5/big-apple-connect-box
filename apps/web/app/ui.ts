// Shared layout for the home page's sections, so each one lines up with the
// header and the next: the same width, the same left edge, the same heading
// size and the same space between sections. The cover is centered; every
// section after it reads from the same left edge as the header.

// The page's column: the header's width and padding, so text starts where
// the site name does.
export const CONTAINER = "mx-auto w-full max-w-6xl px-6";

// Every section after "How it works" has the same frame: a thin rule above it
// and the same space around it. No cards or background bands. Tighter on a
// phone, where a desktop-sized gap is most of a screen of nothing.
export const SECTION = "border-t border-zinc-200 py-16 sm:py-24";

// A section's heading and the paragraph under it.
export const SECTION_TITLE = "text-4xl font-semibold tracking-tight text-balance sm:text-5xl";
export const SECTION_LEAD = "mt-4 max-w-2xl text-lg leading-relaxed text-pretty text-zinc-600 sm:text-xl";

// The frame around the demo video.
export const CARD =
  "rounded-3xl bg-white shadow-[0_1px_2px_rgb(0_0_0/0.04),0_16px_40px_-20px_rgb(0_0_0/0.18)] ring-1 ring-zinc-900/5";
