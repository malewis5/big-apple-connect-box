import { ImageResponse } from "next/og";
import { getContent } from "./content";
import { SITE_NAME } from "./site";

// The alt text has to be known before the image is drawn, so it is written
// here; keep it the same as the hero's title in messages/en.json.
export const alt = `${SITE_NAME}: Is your internet as good as promised?`;
export const size = { width: 1200, height: 630 };
export const contentType = "image/png";

export default async function OpengraphImage() {
  const { hero } = await getContent();
  return new ImageResponse(
    (
      <div
        style={{
          width: "100%",
          height: "100%",
          display: "flex",
          flexDirection: "column",
          justifyContent: "space-between",
          padding: 80,
          background: "linear-gradient(180deg, #ffffff 0%, #f4f4f5 100%)",
          color: "#18181b",
        }}
      >
        <div style={{ display: "flex", alignItems: "center", gap: 20, fontSize: 32, fontWeight: 600 }}>
          <div
            style={{
              width: 26,
              height: 26,
              borderRadius: 13,
              background: "#10b981",
              boxShadow: "0 0 0 10px rgba(16, 185, 129, 0.15)",
            }}
          />
          {SITE_NAME}
        </div>
        <div style={{ display: "flex", flexDirection: "column", gap: 28 }}>
          <div
            style={{
              display: "flex",
              alignSelf: "flex-start",
              padding: "8px 22px",
              borderRadius: 999,
              background: "#ecfdf5",
              color: "#047857",
              fontSize: 28,
            }}
          >
            {hero.eyebrow}
          </div>
          <div style={{ fontSize: 84, fontWeight: 600, letterSpacing: -2, lineHeight: 1.05 }}>
            {hero.title}
          </div>
        </div>
        <div style={{ display: "flex", fontSize: 28, color: "#71717a" }}>bigappleconnect.com</div>
      </div>
    ),
    size,
  );
}
