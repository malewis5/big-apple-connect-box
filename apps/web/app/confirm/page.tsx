import type { Metadata } from "next";
import { getContent } from "../content";
import { SiteFooter } from "../info-page";
import ConfirmForm from "./confirm-form";

// Where the link in the confirmation email lands: /confirm#<code>. The page
// itself is static; the form reads the code in the browser (confirm-form.tsx).
// It isn't in the sitemap, and search engines are asked not to list it.
export async function generateMetadata(): Promise<Metadata> {
  const { confirm } = await getContent();
  return {
    title: confirm.title,
    description: confirm.description,
    alternates: { canonical: null },
    robots: { index: false, follow: true },
  };
}

export default async function ConfirmPage() {
  const { confirm } = await getContent();
  return (
    <div className="flex flex-1 flex-col bg-white text-zinc-900">
      <main id="main" tabIndex={-1} className="mx-auto w-full max-w-2xl flex-1 px-6 pt-12 pb-24 outline-none sm:pt-20">
        <h1 className="text-4xl font-semibold tracking-tight text-balance sm:text-6xl">{confirm.title}</h1>
        <ConfirmForm
          labels={{
            body: confirm.body,
            button: confirm.button,
            confirming: confirm.confirming,
            successTitle: confirm.successTitle,
            successBody: confirm.successBody,
            missing: confirm.missing,
            back: confirm.back,
          }}
        />
      </main>
      <SiteFooter />
    </div>
  );
}
