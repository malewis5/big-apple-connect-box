import { MARKDOWN_HEADERS, infoMarkdown } from "../markdown";

export async function GET() {
  return new Response(await infoMarkdown("privacy"), { headers: MARKDOWN_HEADERS("/privacy") });
}
