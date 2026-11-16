import { getContent } from "./content";
import DemoVideo from "./demo-video";
import { SiteFooter } from "./info-page";
import SignupForm from "./signup/signup-form";
import { CONTACT_EMAIL, DEMO_VIDEO, HOME_UPDATED, REPO, SITE_NAME, SITE_URL } from "./site";
import { TrackedLink } from "./tracked-link";
import { CONTAINER, SECTION, SECTION_LEAD, SECTION_TITLE } from "./ui";

// The maker's own public profiles. They tell search engines and AI answer
// engines which Matt Lewis this is. Each one is a profile he links to from
// his GitHub page or his own website. Only list a profile that really exists.
const AUTHOR_PROFILES = [
  "https://github.com/malewis5",
  "https://www.linkedin.com/in/malewis5",
  "https://twitter.com/mttlws",
  "https://www.mattech.fyi",
];

const PRIMARY_BUTTON =
  "inline-flex min-h-12 items-center justify-center rounded-full bg-zinc-900 px-7 py-3 text-base font-semibold text-white transition hover:bg-zinc-700";

// The page reads top to bottom as an argument: what it is,
// how it works, the evidence, where it stands, then the two asks (partners,
// updates) and the remaining questions. Every call to action near the top
// jumps to one of the two asks.
export default async function Home() {
  const c = await getContent();
  const { hero, how, video, chicago, partners, signup, faq } = c;

  // Structured data for search engines and AI answer engines. Every entry
  // describes something the page shows, in the same words, as Google requires.
  const jsonLd = {
    "@context": "https://schema.org",
    "@graph": [
      {
        "@type": "WebSite",
        "@id": `${SITE_URL}/#website`,
        url: SITE_URL,
        name: SITE_NAME,
        description: c.site.description,
        inLanguage: "en-US",
        publisher: { "@id": `${SITE_URL}/#author` },
      },
      {
        "@type": "WebPage",
        "@id": `${SITE_URL}/#webpage`,
        url: SITE_URL,
        name: c.site.title,
        headline: hero.title,
        description: hero.body,
        inLanguage: "en-US",
        dateModified: HOME_UPDATED,
        isPartOf: { "@id": `${SITE_URL}/#website` },
        about: { "@id": `${SITE_URL}/#software` },
        author: { "@id": `${SITE_URL}/#author` },
        hasPart: [
          { "@id": `${SITE_URL}/#how-it-works` },
          { "@id": `${SITE_URL}/#video` },
          { "@id": `${SITE_URL}/#why-it-matters` },
          { "@id": `${SITE_URL}/#chicago` },
          { "@id": `${SITE_URL}/#faq` },
        ],
        // The published research behind the Chicago figures on the page.
        citation: chicago.stats.map((stat) => ({ "@type": "CreativeWork", name: stat.sourceLabel, url: stat.href })),
      },
      {
        "@type": "Person",
        "@id": `${SITE_URL}/#author`,
        name: "Matt Lewis",
        description: `Matt Lewis makes and maintains the ${SITE_NAME} as an independent, open-source project that measures home internet for NYCHA residents.`,
        url: "https://github.com/malewis5",
        sameAs: AUTHOR_PROFILES,
        contactPoint: { "@type": "ContactPoint", email: CONTACT_EMAIL, contactType: "partnerships and press" },
      },
      {
        "@type": "SoftwareApplication",
        "@id": `${SITE_URL}/#software`,
        name: SITE_NAME,
        description:
          "An open-source program that measures home internet speed, latency and reliability on a regular schedule and saves the results on your own computer.",
        applicationCategory: "UtilitiesApplication",
        operatingSystem: "macOS, Windows, Linux, Raspberry Pi OS",
        isAccessibleForFree: true,
        offers: { "@type": "Offer", price: "0", priceCurrency: "USD" },
        license: `${REPO}/blob/main/LICENSE`,
        codeRepository: REPO,
        downloadUrl: `${REPO}/releases/latest`,
        author: { "@id": `${SITE_URL}/#author` },
        url: SITE_URL,
        sameAs: [REPO],
        audience: {
          "@type": "Audience",
          audienceType: "NYCHA residents and New York City public housing households",
        },
      },
      // The three steps under "How it works".
      {
        "@type": "HowTo",
        "@id": `${SITE_URL}/#how-it-works`,
        name: `${how.title}: ${SITE_NAME}`,
        description: hero.body,
        step: how.steps.map((step, k) => ({ "@type": "HowToStep", position: k + 1, name: step.title, text: step.body })),
      },
      // The demo video, with its captions as the transcript, so search engines
      // and AI answer engines can show and quote it.
      {
        "@type": "VideoObject",
        "@id": `${SITE_URL}/#video`,
        name: `${SITE_NAME}: ${video.title}`,
        description: video.body,
        thumbnailUrl: [`${SITE_URL}${DEMO_VIDEO.files.thumbnail}`, `${SITE_URL}${DEMO_VIDEO.files.poster}`],
        uploadDate: DEMO_VIDEO.uploaded,
        duration: DEMO_VIDEO.duration,
        contentUrl: `${SITE_URL}${DEMO_VIDEO.files.mp4}`,
        encodingFormat: "video/mp4",
        width: 1280,
        height: 720,
        inLanguage: "en-US",
        isFamilyFriendly: true,
        isAccessibleForFree: true,
        transcript: video.transcript.join(" "),
        caption: {
          "@type": "MediaObject",
          contentUrl: `${SITE_URL}${DEMO_VIDEO.files.captions}`,
          encodingFormat: "text/vtt",
          inLanguage: "en-US",
        },
        url: `${SITE_URL}/#demo`,
        isPartOf: { "@id": `${SITE_URL}/#webpage` },
        about: { "@id": `${SITE_URL}/#software` },
        creator: { "@id": `${SITE_URL}/#author` },
        publisher: { "@id": `${SITE_URL}/#author` },
      },
      // The three findings in the Chicago section.
      {
        "@type": "ItemList",
        "@id": `${SITE_URL}/#chicago`,
        name: chicago.title,
        description: chicago.body,
        itemListElement: chicago.stats.map((stat, k) => ({
          "@type": "ListItem",
          position: k + 1,
          name: `${stat.figure}: ${stat.title}`,
          description: stat.body,
          url: stat.href,
        })),
      },
      {
        "@type": "FAQPage",
        "@id": `${SITE_URL}/#faq`,
        mainEntity: faq.items.map((item) => ({
          "@type": "Question",
          name: item.question,
          acceptedAnswer: { "@type": "Answer", text: item.answer },
        })),
      },
    ],
  };

  return (
    <div className="flex-1 bg-white text-zinc-900">
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd).replace(/</g, "\\u003c") }}
      />
      <main id="main" tabIndex={-1} className="outline-none">
        {/* 1. The cover, centered: one column, so the eye has one line to
            follow. It's the only centered part of the page; the sections below
            read from the left. Words only: the headline, the paragraph, one
            button, a link to the video and the fine print. Partners have their
            own door in the header. */}
        <section className="pt-10 pb-12 sm:pt-12 sm:pb-16">
          <div className={`${CONTAINER} flex flex-col items-center text-center`}>
            <p className="text-base font-medium text-zinc-600">{hero.eyebrow}</p>
            {/* Two lines on every screen. The longer line is about 10em wide, so
                the size is a tenth of the column (the screen less its 3rem of
                padding), between 2rem and 5.5rem. */}
            <h1 className="mt-3 text-[clamp(2rem,calc((100vw_-_3rem)/10),5.5rem)] leading-none font-semibold tracking-[-0.035em] break-words">
              <span className="block">{hero.titleLines[0]}</span>{" "}
              <span className="block">{hero.titleLines[1]}</span>
            </h1>
            <p className="mt-6 max-w-2xl text-lg leading-relaxed text-balance text-zinc-600 sm:mt-8 sm:text-xl">{hero.body}</p>
            <div className="mt-8 flex flex-col items-center gap-5 sm:flex-row sm:gap-8">
              <TrackedLink href="#updates" event="Get updates clicked" place="hero" className={PRIMARY_BUTTON}>
                {hero.primary}
              </TrackedLink>
              <TrackedLink href="#how-it-works" event="Watch clicked" place="hero" className="text-base font-medium text-zinc-900 underline decoration-zinc-900/25 underline-offset-4 hover:text-zinc-900 hover:decoration-zinc-900">
                {hero.watch}{" "}
                <time dateTime={`PT${DEMO_VIDEO.seconds}S`} className="text-zinc-600 tabular-nums">
                  {`(${Math.floor(DEMO_VIDEO.seconds / 60)}:${String(DEMO_VIDEO.seconds % 60).padStart(2, "0")})`}
                </time>
              </TrackedLink>
            </div>
            {/* The fine print: the disclaimer. It stays above the fold because
                the project's name borrows the city program's. */}
            <p className="mt-12 text-sm leading-relaxed text-zinc-600">{hero.note}</p>
          </div>
        </section>

        {/* 2. How it works: the steps on the left edge, under the cover's words,
            and the video beside them, filling the space the cover leaves on the
            right. On a phone the steps come first, then the video. */}
        <section id="how-it-works" aria-labelledby="how-title" className="bg-white pt-8 pb-16 sm:pt-12 sm:pb-32">
          <div className={`${CONTAINER} lg:grid lg:grid-cols-[2fr_3fr] lg:items-center lg:gap-16`}>
            <div>
              <h2 id="how-title" className={SECTION_TITLE}>
                {how.title}
              </h2>
              <ol className="mt-8 space-y-6 sm:mt-10 sm:space-y-8">
                {how.steps.map((step, k) => (
                  <li key={step.key}>
                    {/* The number sits on the title's line, like a list marker.
                        The <ol> already gives screen readers the order. */}
                    <h3 className="flex items-center gap-3 text-xl font-semibold">
                      <span aria-hidden="true" className="flex size-7 shrink-0 items-center justify-center rounded-full bg-emerald-50 text-base font-semibold text-emerald-800">
                        {k + 1}
                      </span>
                      {step.title}
                    </h3>
                    <p className="mt-1.5 pl-10 text-lg leading-relaxed text-pretty text-zinc-600">{step.body}</p>
                  </li>
                ))}
              </ol>
            </div>
            <div id="demo" className="mt-10 min-w-0 lg:mt-0">
              <DemoVideo
                labels={{
                  label: video.label,
                  play: video.play(DEMO_VIDEO.seconds),
                  watch: video.watch,
                  playAgain: video.playAgain,
                  captions: video.captions,
                  fallback: video.fallback((chunks) => (
                    <a href={DEMO_VIDEO.files.mp4} className="underline">
                      {chunks}
                    </a>
                  )),
                }}
              />
            </div>
          </div>
        </section>

        {/* 3. The evidence: what Chicago's measuring found, with sources. */}
        <section id="chicago" aria-labelledby="chicago-title" className={SECTION}>
          <div className={CONTAINER}>
            <h2 id="chicago-title" className={SECTION_TITLE}>
              {chicago.title}
            </h2>
            <p className={SECTION_LEAD}>{chicago.body}</p>
            {/* Each finding reads as an argument: the claim, the number, the
                sentence that explains the number, then its source. On a wide
                screen the three share their rows, so the numbers and the
                sources line up however the claims wrap. */}
            <div className="mt-12 grid gap-10 md:grid-cols-3 md:gap-x-12 md:gap-y-0">
              {chicago.stats.map((stat) => (
                <figure key={stat.key} className="min-w-0 break-words md:row-span-4 md:grid md:grid-rows-subgrid">
                  <h3 className="text-xl font-semibold text-pretty">{stat.title}</h3>
                  <p className="mt-4 text-5xl font-semibold tracking-tight tabular-nums">{stat.figure}</p>
                  <p className="mt-2 text-base leading-relaxed text-zinc-600">{stat.body}</p>
                  <figcaption className="mt-3 text-sm leading-relaxed text-zinc-600">
                    {chicago.source}:{" "}
                    <a href={stat.href} className="underline decoration-zinc-900/25 underline-offset-4 hover:text-zinc-900 hover:decoration-zinc-900">
                      {stat.sourceLabel}
                    </a>
                  </figcaption>
                </figure>
              ))}
            </div>
            {/* The small print, in one place: the study's report, then the disclaimer. */}
            <div className="mt-12 max-w-3xl space-y-1 text-sm leading-relaxed text-zinc-600">
              <p>
                {chicago.report.intro}:{" "}
                <a href={chicago.report.href} className="underline decoration-zinc-900/25 underline-offset-4 hover:text-zinc-900 hover:decoration-zinc-900">
                  {chicago.report.label}
                </a>
              </p>
              <p>{chicago.note}</p>
            </div>
          </div>
        </section>

        {/* 4. For partners, the city above all. On a wide screen: the ask and
            the email button on the left, the project's aims on the right, so
            the section fills the page's width. On a phone it stacks: the ask,
            the aims, then the email. */}
        <section id="partners" aria-labelledby="partners-title" className={SECTION}>
          <div className={`${CONTAINER} lg:grid lg:grid-cols-2 lg:grid-rows-[auto_1fr] lg:gap-x-16`}>
            <div className="lg:col-start-1 lg:row-start-1">
              <h2 id="partners-title" className={SECTION_TITLE}>
                {partners.title}
              </h2>
              <p className={SECTION_LEAD}>{partners.body}</p>
            </div>
            <div className="mt-12 lg:col-start-2 lg:row-span-2 lg:row-start-1 lg:mt-0">
              <h3 className="text-sm font-medium text-zinc-600">{partners.aimsLabel}</h3>
              <dl className="mt-3 space-y-4">
                {partners.aims.map((item) => (
                  <div key={item.key}>
                    <dt className="text-lg font-semibold">{item.label}</dt>
                    <dd className="mt-1 text-base leading-relaxed text-zinc-600">{item.body}</dd>
                  </div>
                ))}
              </dl>
            </div>
            <div className="mt-12 lg:col-start-1 lg:row-start-2 lg:mt-10">
              <TrackedLink href={partners.emailHref} event="Partner email clicked" place="partners" className={`${PRIMARY_BUTTON} text-center [overflow-wrap:anywhere]`}>
                {partners.email}
              </TrackedLink>
            </div>
          </div>
        </section>

        {/* 5. Get updates. The title and its line sit on the page at the left
            edge, like every other section's; the form is on the one dark card
            on the page, beside them on a wide screen and under them on a phone. */}
        <section id="updates" aria-labelledby="updates-title" className={SECTION}>
          <div className={`${CONTAINER} lg:grid lg:grid-cols-2 lg:items-start lg:gap-x-16`}>
            <div>
              <h2 id="updates-title" className={SECTION_TITLE}>
                {signup.title}
              </h2>
              <p className={SECTION_LEAD}>{signup.body}</p>
            </div>
            <div className="mt-10 rounded-3xl bg-zinc-900 p-6 text-white sm:p-10 lg:mt-0">
              <SignupForm
                labels={{
                  emailLabel: signup.emailLabel,
                  emailPlaceholder: signup.emailPlaceholder,
                  roleLabel: signup.roleLabel,
                  optional: signup.optional,
                  roleChoose: signup.roleChoose,
                  roles: signup.roles,
                  honeypot: signup.honeypot,
                  submit: signup.submit,
                  submitting: signup.submitting,
                  successTitle: signup.successTitle,
                  privacy: signup.privacy((chunks) => (
                    <a href="/privacy" className="font-medium text-white underline underline-offset-4 hover:text-zinc-200">
                      {chunks}
                    </a>
                  )),
                }}
              />
            </div>
          </div>
        </section>

        {/* 6. The questions that are left. */}
        <section id="faq" aria-labelledby="faq-title" className={SECTION}>
          <div className={`${CONTAINER} lg:grid lg:grid-cols-[1fr_2fr] lg:items-start lg:gap-16`}>
            <h2 id="faq-title" className={SECTION_TITLE}>
              {faq.title}
            </h2>
            <div className="mt-8 divide-y divide-zinc-900/10 border-y border-zinc-900/10 lg:mt-0">
              {faq.items.map((item) => (
                <details key={item.key} className="group">
                  <summary className="flex cursor-pointer list-none items-center justify-between gap-6 py-5 text-lg font-semibold sm:py-6 [&::-webkit-details-marker]:hidden">
                    <h3>{item.question}</h3>
                    <svg
                      viewBox="0 0 20 20"
                      aria-hidden="true"
                      className="size-5 shrink-0 text-zinc-500 transition group-open:rotate-180 motion-reduce:transition-none"
                      fill="none"
                      stroke="currentColor"
                      strokeWidth="1.5"
                    >
                      <path d="M5 8l5 5 5-5" strokeLinecap="round" strokeLinejoin="round" />
                    </svg>
                  </summary>
                  <p className="-mt-2 max-w-2xl pb-5 text-lg leading-relaxed text-zinc-600 sm:-mt-3 sm:pb-6">{item.answer}</p>
                </details>
              ))}
            </div>
          </div>
        </section>
      </main>

      <SiteFooter />
    </div>
  );
}
