// Facts about the site that aren't words to translate: addresses, dates,
// files and lists of keys. The words people read are in messages/en.json and
// reach the pages through content.ts.

// bigappleconnect.com redirects to the www host, so www is canonical.
export const SITE_URL = "https://www.bigappleconnect.com";
export const SITE_NAME = "Big Apple Connect Box";
export const REPO = "https://github.com/malewis5/big-apple-connect-box";

// Where partners, press and people with private questions write. Google
// Workspace delivers it to the maintainer; see README.md, "Email".
export const CONTACT_EMAIL = "hello@bigappleconnect.com";

// The language of the words on every page. The site has one for now; see
// i18n/request.ts before adding another.
export const DEFAULT_LOCALE = "en";

// The day each page's words last changed, as YYYY-MM-DD. The sitemap tells
// search engines this date, so change it whenever that page's words in
// messages/en.json change.
export const HOME_UPDATED = "2026-11-16";
export const PAGE_UPDATED = {
  about: "2026-11-16",
  contact: "2026-11-16",
  privacy: "2026-11-16",
} as const;

// The About, Contact and Privacy pages, in the order the footer lists them.
export const INFO_PAGES = ["about", "contact", "privacy"] as const;
export type InfoSlug = (typeof INFO_PAGES)[number];

// The answers to the signup form's optional "I am a..." question. Each one
// has a label in messages/en.json under Home.signup.roles, and is saved in
// Resend as the contact's `role` property.
export const SIGNUP_ROLES = ["resident", "advocate", "government", "press", "other"] as const;
export type SignupRole = (typeof SIGNUP_ROLES)[number];

// The demo video. Its captions are in public/demo/demo-wide.vtt, made by
// design/demo/build_captions.py; the transcript in messages/en.json repeats
// them for the Markdown page and structured data, so keep the two in step.
export const DEMO_VIDEO = {
  // For search engines: ISO 8601 length and upload time (with time zone).
  duration: "PT45S",
  seconds: 45,
  uploaded: "2026-11-13T12:00:00-05:00",
  // Files in public/demo. The page plays the AV1 file where it can and the
  // H.264 file elsewhere; search engines get the H.264 file and the JPEG.
  files: {
    webm: "/demo/demo-wide.webm",
    mp4: "/demo/demo-wide-720.mp4",
    captions: "/demo/demo-wide.vtt",
    poster: "/demo/demo-wide-poster.webp",
    thumbnail: "/demo/demo-wide-poster.jpg",
  },
} as const;
