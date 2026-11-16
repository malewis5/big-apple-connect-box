import type { ReactNode } from "react";
import { cache } from "react";
import { getTranslations } from "next-intl/server";
import {
  CONTACT_EMAIL,
  DEFAULT_LOCALE,
  INFO_PAGES,
  PAGE_UPDATED,
  SIGNUP_ROLES,
  type InfoSlug,
} from "./site";

// The site's words, read from messages/<locale>.json through next-intl and
// shaped for the pages, the Markdown versions, llms.txt, the sitemap and the
// structured data, so they all say the same thing.
//
// Lists (steps, questions, page sections) are objects in the catalog, keyed
// by a stable name. Their order in the file is the order on the page, and
// every string goes through next-intl, so a translation only has to keep the
// same keys.

type Translator = {
  (key: string, values?: Record<string, string | number>): string;
  rich(key: string, values: Record<string, (chunks: ReactNode) => ReactNode>): ReactNode;
  raw(key: string): unknown;
  has(key: string): boolean;
};

export type InfoPage = {
  slug: InfoSlug;
  title: string;
  description: string;
  // The day this page's words last changed, as YYYY-MM-DD, for the sitemap.
  updated: string;
  sections: { title: string; paragraphs: string[]; links: { label: string; href: string }[] }[];
};

async function load(locale: string) {
  // Passing the locale keeps next-intl from reading the request (i18n/request.ts).
  const t = (await getTranslations({ locale })) as unknown as Translator;
  // Some lists are optional, like a page section's links.
  const keys = (path: string) => {
    if (!t.has(path)) return [];
    const value = t.raw(path);
    return value && typeof value === "object" ? Object.keys(value) : [];
  };
  const list = <T>(path: string, item: (key: string, at: string) => T) =>
    keys(path).map((key) => item(key, `${path}.${key}`));

  const infoPage = (slug: InfoSlug): InfoPage => {
    const at = `Pages.${slug}`;
    return {
      slug,
      title: t(`${at}.title`),
      description: t(`${at}.description`),
      updated: PAGE_UPDATED[slug],
      sections: list(`${at}.sections`, (_, s) => ({
        title: t(`${s}.title`),
        paragraphs: list(`${s}.paragraphs`, (_, p) => t(p)),
        links: list(`${s}.links`, (_, l) => ({ label: t(`${l}.label`), href: t(`${l}.href`) })),
      })),
    };
  };

  return {
    locale,
    site: {
      title: t("Site.title"),
      description: t("Site.description"),
      disclaimer: t("Site.disclaimer"),
    },
    header: {
      skip: t("Header.skip"),
      notify: t("Header.notify"),
      partners: t("Header.partners"),
    },
    footer: {
      project: t("Footer.project"),
      site: t("Footer.site"),
      contact: t("Footer.contact"),
      github: t("Footer.github"),
      sources: t("Footer.sources"),
    },
    hero: {
      eyebrow: t("Home.hero.eyebrow"),
      titleLines: [t("Home.hero.titleLine1"), t("Home.hero.titleLine2")] as const,
      title: `${t("Home.hero.titleLine1")} ${t("Home.hero.titleLine2")}`,
      body: t("Home.hero.body"),
      primary: t("Home.hero.primary"),
      watch: t("Home.hero.watch"),
      note: t("Home.hero.note"),
    },
    how: {
      title: t("Home.how.title"),
      steps: list("Home.how.steps", (key, s) => ({ key, title: t(`${s}.title`), body: t(`${s}.body`) })),
    },
    video: {
      title: t("Home.video.title"),
      // Not shown on the page; they describe the video to search engines
      // and AI agents (structured data, sitemap, Markdown).
      body: t("Home.video.body"),
      label: t("Home.video.label"),
      watch: t("Home.video.watch"),
      playAgain: t("Home.video.playAgain"),
      captions: t("Home.video.captions"),
      play: (seconds: number) => t("Home.video.play", { seconds }),
      fallback: (link: (chunks: ReactNode) => ReactNode) => t.rich("Home.video.fallback", { link }),
      transcript: list("Home.video.transcript", (_, l) => t(l)),
    },
    chicago: {
      title: t("Home.chicago.title"),
      body: t("Home.chicago.body"),
      // The paper the Chicago team asks people to cite for the study.
      report: {
        intro: t("Home.chicago.report.intro"),
        label: t("Home.chicago.report.label"),
        href: t("Home.chicago.report.href"),
      },
      source: t("Home.chicago.source"),
      note: t("Home.chicago.note"),
      stats: list("Home.chicago.stats", (key, s) => ({
        key,
        figure: t(`${s}.figure`),
        title: t(`${s}.title`),
        body: t(`${s}.body`),
        sourceLabel: t(`${s}.sourceLabel`),
        href: t(`${s}.href`),
      })),
    },
    partners: {
      title: t("Home.partners.title"),
      body: t("Home.partners.body"),
      aimsLabel: t("Home.partners.aimsLabel"),
      aims: list("Home.partners.aims", (key, p) => ({ key, label: t(`${p}.label`), body: t(`${p}.body`) })),
      email: t("Home.partners.email"),
      emailHref: `mailto:${CONTACT_EMAIL}?subject=${encodeURIComponent(t("Home.partners.emailSubject"))}`,
    },
    signup: {
      title: t("Home.signup.title"),
      body: t("Home.signup.body"),
      emailLabel: t("Home.signup.emailLabel"),
      emailPlaceholder: t("Home.signup.emailPlaceholder"),
      roleLabel: t("Home.signup.roleLabel"),
      optional: t("Home.signup.optional"),
      roleChoose: t("Home.signup.roleChoose"),
      roles: SIGNUP_ROLES.map((value) => ({ value, label: t(`Home.signup.roles.${value}`) })),
      honeypot: t("Home.signup.honeypot"),
      submit: t("Home.signup.submit"),
      submitting: t("Home.signup.submitting"),
      successTitle: t("Home.signup.successTitle"),
      successBody: (email: string) => t("Home.signup.successBody", { email }),
      privacy: (link: (chunks: ReactNode) => ReactNode) => t.rich("Home.signup.privacy", { link }),
      errors: {
        email: t("Home.signup.errors.email"),
        tooMany: t("Home.signup.errors.tooMany"),
        unavailable: t("Home.signup.errors.unavailable"),
        failed: t("Home.signup.errors.failed"),
      },
    },
    faq: {
      title: t("Home.faq.title"),
      items: list("Home.faq.items", (key, f) => ({ key, question: t(`${f}.question`), answer: t(`${f}.answer`) })),
    },
    confirm: {
      title: t("Confirm.title"),
      description: t("Confirm.description"),
      body: t("Confirm.body"),
      button: t("Confirm.button"),
      confirming: t("Confirm.confirming"),
      successTitle: t("Confirm.successTitle"),
      successBody: t("Confirm.successBody"),
      missing: t("Confirm.missing"),
      invalid: t("Confirm.invalid"),
      expired: t("Confirm.expired"),
      unavailable: t("Confirm.unavailable"),
      back: t("Confirm.back"),
    },
    email: {
      subject: t("Email.subject"),
      heading: t("Email.heading"),
      body: t("Email.body"),
      button: t("Email.button"),
      textLink: (link: string) => t("Email.textLink", { link }),
      expiry: t("Email.expiry"),
    },
    notFound: {
      metaTitle: t("NotFound.metaTitle"),
      eyebrow: t("NotFound.eyebrow"),
      title: t("NotFound.title"),
      back: t("NotFound.back"),
      otherPages: t("NotFound.otherPages"),
      check: {
        title: t("NotFound.check.title"),
        done: t("NotFound.check.done"),
        rows: [
          { label: t("NotFound.check.internet.label"), value: t("NotFound.check.internet.value"), note: t("NotFound.check.internet.note"), ok: true },
          { label: t("NotFound.check.site.label"), value: t("NotFound.check.site.value"), note: t("NotFound.check.site.note"), ok: true },
          { label: t("NotFound.check.page.label"), value: t("NotFound.check.page.value"), note: "", ok: false },
        ],
      },
    },
    whenToUse: {
      recommend: list("WhenToUse.recommend", (_, k) => t(k)),
      avoid: list("WhenToUse.avoid", (_, k) => t(k)),
      howTo: list("WhenToUse.howTo", (_, k) => t(k)),
    },
    pages: INFO_PAGES.map(infoPage),
  };
}

export type Content = Awaited<ReturnType<typeof load>>;

// Loaded once per request (or once per prerender), however many components ask.
export const getContent = cache((locale: string = DEFAULT_LOCALE) => load(locale));

export async function getInfoPage(slug: InfoSlug, locale: string = DEFAULT_LOCALE) {
  const page = (await getContent(locale)).pages.find((p) => p.slug === slug);
  if (!page) throw new Error(`No page named ${slug} in messages/${locale}.json`);
  return page;
}
