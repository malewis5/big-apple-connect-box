import { MARKDOWN_HEADERS, homeMarkdown } from "../markdown";

export async function GET() {
  return new Response(await homeMarkdown(), { headers: MARKDOWN_HEADERS("/") });
}
