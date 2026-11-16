import type { MetadataRoute } from "next";
import { getContent } from "./content";
import { SITE_NAME } from "./site";

export default async function manifest(): Promise<MetadataRoute.Manifest> {
  const { site } = await getContent();
  return {
    name: SITE_NAME,
    short_name: "Connect Box",
    description: site.description,
    start_url: "/",
    display: "standalone",
    background_color: "#ffffff",
    theme_color: "#ffffff",
    icons: [
      { src: "/icon.svg", sizes: "any", type: "image/svg+xml" },
      { src: "/apple-icon.png", sizes: "180x180", type: "image/png" },
    ],
  };
}
