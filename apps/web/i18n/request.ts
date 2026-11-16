import { getRequestConfig } from "next-intl/server";
import { DEFAULT_LOCALE } from "../app/site";

// next-intl's request configuration: which language a page is in, and its words.
//
// The site has one language and no locale in its addresses, so every page is
// English and nothing here reads the request. That keeps every page static,
// and no cookie is set. Callers pass the locale explicitly (content.ts does),
// which next-intl hands over as `locale`; `requestLocale` is never read, so
// next-intl never looks at the request headers.
//
// To add a language: add messages/<locale>.json with the same keys, reviewed
// by a native speaker, then add locale routing (a [locale] segment, hreflang
// alternates, a sitemap entry per language and Markdown per language in
// proxy.ts). If you use next-intl's middleware, set `localeCookie: false`:
// the Privacy page promises no cookies. See
// https://next-intl.dev/docs/routing and https://vercel.shop/docs/skills/enable-i18n.
const catalogs: Record<string, () => Promise<{ default: Record<string, unknown> }>> = {
  en: () => import("../messages/en.json"),
};

export default getRequestConfig(async ({ locale }) => {
  const resolved = locale && locale in catalogs ? locale : DEFAULT_LOCALE;
  return {
    locale: resolved,
    messages: (await catalogs[resolved]()).default,
    timeZone: "America/New_York",
  };
});
