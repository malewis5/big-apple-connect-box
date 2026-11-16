"use client";

import { track } from "@vercel/analytics";
import Link from "next/link";
import type { ComponentProps } from "react";

// The custom events the site records, as named in Vercel Web Analytics. They
// show which calls to action people use. Custom events need a Pro or
// Enterprise plan; on Hobby, track() records nothing and nothing breaks.
export type TrackedEvent = "Get updates clicked" | "Watch clicked" | "Partner clicked" | "Partner email clicked" | "GitHub clicked";

// A link that records a custom event when pressed. `place` says which copy of
// the link it was (the header, the hero, the partner section and so on).
// Addresses on this site go through Next's Link; others are plain links.
export function TrackedLink({
  event,
  place,
  href,
  onClick,
  ...props
}: Omit<ComponentProps<"a">, "href"> & { href: string; event: TrackedEvent; place: string }) {
  const handleClick: ComponentProps<"a">["onClick"] = (e) => {
    track(event, { place });
    onClick?.(e);
  };
  if (href.startsWith("/")) return <Link href={href} onClick={handleClick} {...props} />;
  return <a href={href} onClick={handleClick} {...props} />;
}
