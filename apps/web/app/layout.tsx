import type { Metadata } from "next";
import { Analytics } from "@vercel/analytics/next";
import { SpeedInsights } from "@vercel/speed-insights/next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";
import { getContent } from "./content";
import { SiteHeader } from "./site-header";
import { DEFAULT_LOCALE, SITE_NAME, SITE_URL } from "./site";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export async function generateMetadata(): Promise<Metadata> {
  const { site } = await getContent();
  return {
    metadataBase: new URL(SITE_URL),
    title: { default: site.title, template: `%s | ${SITE_NAME}` },
    description: site.description,
    applicationName: SITE_NAME,
    keywords: [
      "Big Apple Connect",
      "NYCHA internet",
      "NYCHA Wi-Fi",
      "free internet NYC",
      "home internet speed test",
      "internet reliability",
      "broadband measurement",
      "Raspberry Pi speed test",
      "Netrics",
      "University of Chicago internet study",
    ],
    authors: [{ name: "Matt Lewis", url: "https://github.com/malewis5" }],
    creator: "Matt Lewis",
    category: "technology",
    alternates: { canonical: "/", types: { "text/markdown": "/index.md" } },
    openGraph: {
      type: "website",
      url: "/",
      siteName: SITE_NAME,
      title: site.title,
      description: site.description,
      locale: "en_US",
    },
    twitter: {
      card: "summary_large_image",
      title: site.title,
      description: site.description,
    },
    robots: {
      index: true,
      follow: true,
      googleBot: { index: true, follow: true, "max-image-preview": "large", "max-snippet": -1 },
    },
    // Google Search Console is verified with a DNS TXT record on bigappleconnect.com.
  };
}

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    <html
      lang={DEFAULT_LOCALE}
      className={`${geistSans.variable} ${geistMono.variable} h-full antialiased`}
    >
      <body className="min-h-full flex flex-col">
        <SiteHeader />
        {children}
        <Analytics />
        <SpeedInsights />
      </body>
    </html>
  );
}
