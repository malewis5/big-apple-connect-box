import type { Content } from "../content";

// The confirmation email, as HTML and plain text. Kept plain on purpose: one
// heading, one sentence, one button, and the link written out for mail apps
// that hide buttons. Styles are inline because most mail apps drop <style>.
// The written-out link is one long unbroken string, so its paragraph may break
// anywhere; otherwise it runs past the edge of the card.

const escape = (text: string) =>
  text.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/"/g, "&quot;");

export function confirmationEmail(words: Content["email"], disclaimer: string, link: string) {
  const html = `<!doctype html>
<html lang="en">
<body style="margin:0;padding:0;background:#f4f4f5;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Helvetica,Arial,sans-serif;color:#18181b;">
  <div style="max-width:520px;margin:0 auto;padding:40px 24px;">
    <div style="background:#ffffff;border-radius:24px;padding:32px;">
      <p style="margin:0 0 24px;font-size:14px;font-weight:600;">Big Apple Connect Box</p>
      <h1 style="margin:0 0 16px;font-size:28px;line-height:1.2;">${escape(words.heading)}</h1>
      <p style="margin:0 0 28px;font-size:17px;line-height:1.6;color:#52525b;">${escape(words.body)}</p>
      <a href="${escape(link)}" style="display:inline-block;background:#18181b;color:#ffffff;text-decoration:none;font-weight:600;font-size:16px;padding:14px 28px;border-radius:999px;">${escape(words.button)}</a>
      <p style="margin:28px 0 0;font-size:14px;line-height:1.6;color:#52525b;overflow-wrap:anywhere;word-break:break-all;">${escape(words.textLink(link))}</p>
      <p style="margin:16px 0 0;font-size:14px;line-height:1.6;color:#52525b;">${escape(words.expiry)}</p>
    </div>
    <p style="margin:24px 8px 0;font-size:12px;line-height:1.6;color:#52525b;">${escape(disclaimer)}</p>
  </div>
</body>
</html>`;
  const text = [words.heading, "", words.body, "", words.textLink(link), "", words.expiry, "", "--", disclaimer].join("\n");
  return { subject: words.subject, html, text };
}
