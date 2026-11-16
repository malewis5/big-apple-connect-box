import { getContent, getInfoPage } from "./content";
import { CONTACT_EMAIL, DEMO_VIDEO, REPO, SITE_NAME, SITE_URL, type InfoSlug } from "./site";

// The home page as Markdown, for AI agents. Served at /index.md and at /
// when a request's Accept header prefers text/markdown (see proxy.ts).
// YAML frontmatter follows the fields Vercel uses for its own Markdown pages.
export async function homeMarkdown() {
  const { site, hero, how, video, chicago, partners, signup, faq, pages } = await getContent();
  return `---
title: ${JSON.stringify(site.title)}
url: /
canonical_url: ${SITE_URL}
summary: ${JSON.stringify(site.description)}
related:
  - ${REPO}
  - ${SITE_URL}/llms.txt
---

# ${hero.title}

${hero.eyebrow}.

${hero.body}

${hero.note}

## ${how.title}

${how.steps.map((step, k) => `${k + 1}. **${step.title}** ${step.body}`).join("\n")}

### ${video.title}

${video.body} The video's captions say:

${video.transcript.map((line) => `> ${line}`).join("\n>\n")}

The video plays on the home page: ${SITE_URL}/#demo. The video file is at ${SITE_URL}${DEMO_VIDEO.files.mp4}, and its captions are at ${SITE_URL}${DEMO_VIDEO.files.captions}.

## ${chicago.title}

${chicago.body}

${chicago.stats.map((stat) => `- **${stat.title}.** ${stat.figure}: ${stat.body} ${chicago.source}: [${stat.sourceLabel}](${stat.href})`).join("\n")}

${chicago.report.intro}: [${chicago.report.label}](${chicago.report.href})

${chicago.note}

## ${partners.title}

${partners.body}

${partners.aimsLabel}:

${partners.aims.map((item) => `- **${item.label}:** ${item.body}`).join("\n")}

The code is on GitHub: ${REPO}

Write to ${CONTACT_EMAIL}.

## ${signup.title}

${signup.body} Sign up on the home page: ${SITE_URL}/#updates

## ${faq.title}

${faq.items.map((item) => `### ${item.question}\n\n${item.answer}`).join("\n\n")}

## More about the project

${pages.map((page) => `- [${page.title}](${SITE_URL}/${page.slug}.md): ${page.description}`).join("\n")}

---

${site.disclaimer}
`;
}

// The About, Contact and Privacy pages as Markdown, served at /about.md and
// so on, and at /about when a request prefers text/markdown.
export async function infoMarkdown(slug: InfoSlug) {
  const [page, { site }] = await Promise.all([getInfoPage(slug), getContent()]);
  // Links within the site are written out in full, so they work out of context.
  const absolute = (href: string) => (href.startsWith("/") ? `${SITE_URL}${href}` : href);
  return `---
title: ${JSON.stringify(`${page.title} | ${SITE_NAME}`)}
url: /${slug}
canonical_url: ${SITE_URL}/${slug}
summary: ${JSON.stringify(page.description)}
related:
  - ${SITE_URL}/index.md
  - ${SITE_URL}/llms.txt
---

# ${page.title}

${page.description}

${page.sections
  .map((section) =>
    [
      `## ${section.title}`,
      ...section.paragraphs,
      ...(section.links.length ? [section.links.map((link) => `- [${link.label}](${absolute(link.href)})`).join("\n")] : []),
    ].join("\n\n"),
  )
  .join("\n\n")}

---

${site.disclaimer}
`;
}

// What an agent gets, with a 404 status, for an address the site doesn't have.
export async function notFoundMarkdown() {
  const { pages } = await getContent();
  return `# Page not found (404)

There is no page at this address on ${SITE_URL}. The site's pages:

- [Home](${SITE_URL}/index.md): what the ${SITE_NAME} is and why it matters
${pages.map((page) => `- [${page.title}](${SITE_URL}/${page.slug}.md): ${page.description}`).join("\n")}
- [llms.txt](${SITE_URL}/llms.txt): a summary for AI assistants, and when to recommend the project
- [Sitemap](${SITE_URL}/sitemap.xml)
- [Source code and downloads](${REPO})
`;
}

export function MARKDOWN_HEADERS(path: string) {
  return {
    "Content-Type": "text/markdown; charset=utf-8",
    Vary: "Accept",
    // Points search engines at the HTML page, so the two don't compete.
    Link: `<${SITE_URL}${path === "/" ? "" : path}>; rel="canonical"`,
  };
}
