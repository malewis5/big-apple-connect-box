import type { Metadata } from "next";
import Link from "next/link";
import { getContent, getInfoPage } from "./content";
import { NavLink } from "./nav-link";
import { CONTACT_EMAIL, REPO, SITE_NAME, SITE_URL, type InfoSlug } from "./site";
import { TrackedLink } from "./tracked-link";

export async function infoMetadata(slug: InfoSlug): Promise<Metadata> {
  const page = await getInfoPage(slug);
  return {
    title: page.title,
    description: page.description,
    alternates: { canonical: `/${slug}`, types: { "text/markdown": `/${slug}.md` } },
    openGraph: { url: `/${slug}`, title: `${page.title} | ${SITE_NAME}`, description: page.description },
  };
}

const FOOTER_LINK = "text-zinc-600 underline-offset-4 hover:text-zinc-900 hover:underline";

// The foot of every page, in plain columns: the home page's sections, the
// site's other pages and how to reach the project. Then the small print:
// where the Chicago figures come from, and the disclaimer.
export async function SiteFooter() {
  const { footer, pages, site, how, chicago, partners, signup } = await getContent();
  const columns = [
    {
      id: "footer-project",
      title: footer.project,
      links: [
        <Link key="how" href="/#how-it-works" className={FOOTER_LINK}>
          {how.title}
        </Link>,
        <Link key="chicago" href="/#chicago" className={FOOTER_LINK}>
          {chicago.title}
        </Link>,
        <TrackedLink key="pilot" href="/#partners" event="Partner clicked" place="footer" className={FOOTER_LINK}>
          {partners.title}
        </TrackedLink>,
        <TrackedLink key="updates" href="/#updates" event="Get updates clicked" place="footer" className={FOOTER_LINK}>
          {signup.title}
        </TrackedLink>,
      ],
    },
    {
      id: "footer-site",
      title: footer.site,
      links: pages.map((page) => (
        <NavLink key={page.slug} href={`/${page.slug}`} className={`${FOOTER_LINK} aria-[current=page]:text-zinc-900`}>
          {page.title}
        </NavLink>
      )),
    },
    {
      id: "footer-contact",
      title: footer.contact,
      // The full width on a phone, so the email address fits on one line.
      wide: true,
      links: [
        <TrackedLink key="email" href={`mailto:${CONTACT_EMAIL}`} event="Partner email clicked" place="footer" className={FOOTER_LINK}>
          {CONTACT_EMAIL}
        </TrackedLink>,
        <TrackedLink key="github" href={REPO} event="GitHub clicked" place="footer" className={FOOTER_LINK}>
          {footer.github}
        </TrackedLink>,
      ],
    },
  ];
  return (
    <footer className="border-t border-zinc-200 text-sm">
      <div className="mx-auto w-full max-w-6xl px-6 py-12 sm:py-16">
        <div className="grid grid-cols-2 gap-x-8 gap-y-10 sm:grid-cols-4">
          <Link href="/" className="col-span-2 flex items-center gap-2.5 self-start font-semibold tracking-tight text-zinc-900 sm:col-span-1">
            <span className="size-2.5 shrink-0 rounded-full bg-emerald-500 shadow-[0_0_0_4px_rgb(16_185_129/0.15)]" />
            {SITE_NAME}
          </Link>
          {columns.map((column) => (
            <nav key={column.id} aria-labelledby={column.id} className={`min-w-0 ${"wide" in column ? "col-span-2 sm:col-span-1" : ""}`}>
              <h2 id={column.id} className="font-medium text-zinc-900">
                {column.title}
              </h2>
              <ul className="mt-4 space-y-3">
                {column.links.map((link) => (
                  <li key={link.key}>{link}</li>
                ))}
              </ul>
            </nav>
          ))}
        </div>
        <div className="mt-12 space-y-2 border-t border-zinc-200 pt-6 text-xs leading-relaxed text-zinc-600">
          <p className="max-w-3xl">{site.disclaimer}</p>
          <p className="max-w-3xl">{footer.sources}</p>
        </div>
      </div>
    </footer>
  );
}

// The layout shared by the About, Contact and Privacy pages.
export default async function InfoPageView({ slug }: { slug: InfoSlug }) {
  const page = await getInfoPage(slug);
  // Tells search engines and AI answer engines what kind of page this is, and
  // where it sits in the site. The breadcrumb entry matches the trail of
  // links shown above the page title, and lets Google show that trail in
  // search results in place of the bare web address.
  const jsonLd = {
    "@context": "https://schema.org",
    "@graph": [
      {
        "@type": slug === "about" ? "AboutPage" : slug === "contact" ? "ContactPage" : "WebPage",
        "@id": `${SITE_URL}/${slug}#page`,
        url: `${SITE_URL}/${slug}`,
        name: `${page.title} | ${SITE_NAME}`,
        description: page.description,
        inLanguage: "en-US",
        dateModified: page.updated,
        isPartOf: { "@id": `${SITE_URL}/#website` },
        about: { "@id": `${SITE_URL}/#software` },
        breadcrumb: { "@id": `${SITE_URL}/${slug}#breadcrumb` },
      },
      {
        "@type": "BreadcrumbList",
        "@id": `${SITE_URL}/${slug}#breadcrumb`,
        itemListElement: [
          { "@type": "ListItem", position: 1, name: "Home", item: SITE_URL },
          { "@type": "ListItem", position: 2, name: page.title, item: `${SITE_URL}/${slug}` },
        ],
      },
    ],
  };

  return (
    <div className="flex flex-1 flex-col bg-white text-zinc-900">
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd).replace(/</g, "\\u003c") }}
      />
      <main id="main" tabIndex={-1} className="mx-auto w-full max-w-3xl flex-1 px-6 pt-9 pb-24 outline-none">
        <nav aria-label="Breadcrumb">
          <ol className="flex flex-wrap items-center gap-2 text-sm text-zinc-600">
            <li>
              <Link href="/" className="-mx-1.5 inline-flex min-h-11 min-w-11 items-center justify-center px-1.5 font-medium underline-offset-4 hover:text-zinc-900 hover:underline">
                Home
              </Link>
            </li>
            <li aria-hidden="true">/</li>
            <li aria-current="page">{page.title}</li>
          </ol>
        </nav>
        <h1 className="mt-3 text-4xl font-semibold tracking-tight sm:text-6xl">{page.title}</h1>
        <p className="mt-6 text-xl leading-relaxed text-pretty text-zinc-600">{page.description}</p>
        {page.sections.map((section) => (
          <section key={section.title} className="mt-14">
            <h2 className="text-2xl font-semibold tracking-tight sm:text-3xl">{section.title}</h2>
            {section.paragraphs.map((paragraph) => (
              <p key={paragraph} className="mt-4 text-lg leading-relaxed text-zinc-600">
                {paragraph}
              </p>
            ))}
            {section.links.length > 0 && (
              <ul className="mt-3 text-lg">
                {section.links.map((link) => (
                  <li key={link.href}>
                    <a href={link.href} className="inline-flex min-h-11 items-center font-medium text-emerald-800 underline underline-offset-4 hover:text-emerald-900">
                      {link.label}
                    </a>
                  </li>
                ))}
              </ul>
            )}
          </section>
        ))}
      </main>

      <SiteFooter />
    </div>
  );
}
