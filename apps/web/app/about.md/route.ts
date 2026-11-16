import { MARKDOWN_HEADERS, infoMarkdown } from "../markdown";

export async function GET() {
  return new Response(await infoMarkdown("about"), { headers: MARKDOWN_HEADERS("/about") });
}
