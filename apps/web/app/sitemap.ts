import type { MetadataRoute } from "next";
import { getContent } from "./content";
import { DEMO_VIDEO, HOME_UPDATED, SITE_NAME, SITE_URL } from "./site";

// Each page with the day its words last changed (kept by hand in site.ts).
// Search engines use that date to decide what to read again. They ignore the
// older "priority" and "change frequency" fields, so those are left out.
// The confirm page (/confirm) is left out: it only works from an email link.
export default async function sitemap(): Promise<MetadataRoute.Sitemap> {
  const { video, pages } = await getContent();
  return [
    {
      url: SITE_URL,
      lastModified: HOME_UPDATED,
      // The demo video on the home page. Next writes these into the XML as
      // they are, so they must not contain & or <.
      videos: [
        {
          title: `${SITE_NAME}: ${video.title}`,
          description: video.body,
          thumbnail_loc: `${SITE_URL}${DEMO_VIDEO.files.thumbnail}`,
          content_loc: `${SITE_URL}${DEMO_VIDEO.files.mp4}`,
          duration: DEMO_VIDEO.seconds,
          publication_date: DEMO_VIDEO.uploaded,
          family_friendly: "yes",
        },
      ],
    },
    ...pages.map((page) => ({ url: `${SITE_URL}/${page.slug}`, lastModified: page.updated })),
  ];
}
