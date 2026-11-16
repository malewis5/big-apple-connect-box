import { MARKDOWN_HEADERS, infoMarkdown } from "../markdown";

export async function GET() {
  return new Response(await infoMarkdown("contact"), { headers: MARKDOWN_HEADERS("/contact") });
}
