import { createHmac, timingSafeEqual } from "node:crypto";
import { SIGNUP_ROLES, type SignupRole } from "../site";

// The code in a confirmation link. It carries the email address and the
// optional role, signed with SIGNUP_SECRET, so nothing has to be stored
// before someone confirms: an address that is never confirmed is never kept.
// The link puts it after a # (/confirm#<code>), so it isn't sent to the
// server when the page loads, and doesn't end up in logs or Referer headers.

// How long a link works, in milliseconds. The email and the confirm page say 7 days.
export const TOKEN_MAX_AGE = 7 * 24 * 60 * 60 * 1000;

export type SignupClaim = { email: string; role: SignupRole | "" };

function signature(body: string, secret: string) {
  return createHmac("sha256", secret).update(body).digest("base64url");
}

export function signToken(claim: SignupClaim, secret: string, now = Date.now()) {
  const body = Buffer.from(JSON.stringify({ e: claim.email, r: claim.role, t: now })).toString("base64url");
  return `${body}.${signature(body, secret)}`;
}

export function verifyToken(
  token: string,
  secret: string,
  now = Date.now(),
): ({ ok: true } & SignupClaim) | { ok: false; reason: "invalid" | "expired" } {
  const [body, sig, ...rest] = token.trim().split(".");
  if (!body || !sig || rest.length) return { ok: false, reason: "invalid" };
  const expected = Buffer.from(signature(body, secret));
  const given = Buffer.from(sig);
  if (expected.length !== given.length || !timingSafeEqual(expected, given)) return { ok: false, reason: "invalid" };
  let data: unknown;
  try {
    data = JSON.parse(Buffer.from(body, "base64url").toString("utf8"));
  } catch {
    return { ok: false, reason: "invalid" };
  }
  const { e, r, t } = (data ?? {}) as { e?: unknown; r?: unknown; t?: unknown };
  if (typeof e !== "string" || typeof t !== "number") return { ok: false, reason: "invalid" };
  const role = SIGNUP_ROLES.includes(r as SignupRole) ? (r as SignupRole) : "";
  if (now - t > TOKEN_MAX_AGE || t > now + 60_000) return { ok: false, reason: "expired" };
  return { ok: true, email: e, role };
}
