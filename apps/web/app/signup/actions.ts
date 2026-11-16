"use server";

import { track } from "@vercel/analytics/server";
import { headers } from "next/headers";
import { Resend } from "resend";
import { getContent } from "../content";
import { CONTACT_EMAIL, SIGNUP_ROLES, SITE_NAME, SITE_URL, type SignupRole } from "../site";
import { confirmationEmail } from "./email";
import { allow } from "./rate-limit";
import { signToken, verifyToken } from "./token";

// Email updates with double opt-in, through Resend:
//
// 1. requestSignup checks the form and emails a link to /confirm#<code>.
//    Nothing is saved yet; the code carries the address (see token.ts).
// 2. The person opens the link and presses "Confirm my email", which calls
//    confirmSignup. Only then is the address added to Resend, with its role.
//    The press is needed because some mail filters open every link in an
//    email, which would otherwise confirm addresses nobody confirmed.
//
// Environment variables (README.md, "Email updates"):
//   RESEND_API_KEY     sends the email and saves the contact
//   SIGNUP_SECRET      signs the links; any long random string
//   RESEND_SEGMENT_ID  optional: the segment new contacts join
//   SIGNUP_FROM        optional: the From line; defaults to the project address
// Under `next dev` without RESEND_API_KEY, the link is printed in the terminal
// instead of sent, and confirming only prints what would be saved.

const isDev = process.env.NODE_ENV === "development";

function config() {
  const secret = process.env.SIGNUP_SECRET || (isDev ? "development-only-signup-secret" : "");
  const apiKey = process.env.RESEND_API_KEY ?? "";
  return { secret, apiKey, segmentId: process.env.RESEND_SEGMENT_ID ?? "" };
}

// An email address that a mail server could accept: one @, a dot in the
// domain, no spaces, and no longer than the 254 characters the standard allows.
const EMAIL = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

async function clientAddress() {
  const h = await headers();
  return { ip: h.get("x-forwarded-for")?.split(",")[0]?.trim() || "unknown", origin: h.get("origin") };
}

// Analytics must never break a signup.
async function record(event: string, properties: Record<string, string>) {
  try {
    await track(event, properties);
  } catch {}
}

export type SignupState =
  | { status: "idle" }
  | { status: "sent"; email: string; message: string }
  | { status: "error"; field?: "email"; message: string; email: string; role: string };

export async function requestSignup(_: SignupState, form: FormData): Promise<SignupState> {
  const { signup, email: words, site } = await getContent();
  const email = String(form.get("email") ?? "").trim().toLowerCase();
  const rawRole = String(form.get("role") ?? "");
  const role: SignupRole | "" = SIGNUP_ROLES.includes(rawRole as SignupRole) ? (rawRole as SignupRole) : "";
  const fail = (message: string, field?: "email"): SignupState => ({ status: "error", field, message, email, role });

  // The hidden field people can't see. Only bots fill it in. Tell them it
  // worked, so they don't try again, and send nothing.
  if (String(form.get("company") ?? "") !== "") return { status: "sent", email, message: signup.successBody(email) };

  if (email.length > 254 || !EMAIL.test(email)) return fail(signup.errors.email, "email");

  const { ip, origin } = await clientAddress();
  if (!allow(`ip:${ip}`, 5, 10 * 60_000) || !allow(`email:${email}`, 3, 60 * 60_000)) return fail(signup.errors.tooMany);

  const { secret, apiKey } = config();
  if (!secret) {
    console.error("[signup] SIGNUP_SECRET is not set.");
    return fail(signup.errors.unavailable);
  }
  // Server Actions only run when the request's Origin matches the site, so
  // the origin is this site's own address: the live site, a preview, or localhost.
  const link = `${origin ?? SITE_URL}/confirm#${signToken({ email, role }, secret)}`;

  if (!apiKey) {
    if (!isDev) {
      console.error("[signup] RESEND_API_KEY is not set.");
      return fail(signup.errors.unavailable);
    }
    console.info(`[signup] RESEND_API_KEY is not set, so nothing was sent. Confirmation link for ${email}:\n${link}`);
    return { status: "sent", email, message: signup.successBody(email) };
  }

  const message = confirmationEmail(words, site.disclaimer, link);
  const { error } = await new Resend(apiKey).emails.send({
    from: process.env.SIGNUP_FROM || `${SITE_NAME} <${CONTACT_EMAIL}>`,
    to: email,
    replyTo: CONTACT_EMAIL,
    subject: message.subject,
    html: message.html,
    text: message.text,
  });
  if (error) {
    console.error("[signup] Resend couldn't send the confirmation email:", error);
    return fail(signup.errors.failed);
  }
  await record("Signup requested", { role: role || "none" });
  return { status: "sent", email, message: signup.successBody(email) };
}

export type ConfirmState = { status: "idle" } | { status: "done" } | { status: "error"; message: string };

export async function confirmSignup(_: ConfirmState, form: FormData): Promise<ConfirmState> {
  const { confirm } = await getContent();
  const token = String(form.get("token") ?? "");
  if (!token) return { status: "error", message: confirm.missing };

  const { ip } = await clientAddress();
  if (!allow(`confirm:${ip}`, 10, 10 * 60_000)) return { status: "error", message: confirm.unavailable };

  const { secret, apiKey, segmentId } = config();
  if (!secret) {
    console.error("[signup] SIGNUP_SECRET is not set.");
    return { status: "error", message: confirm.unavailable };
  }
  const claim = verifyToken(token, secret);
  if (!claim.ok) return { status: "error", message: claim.reason === "expired" ? confirm.expired : confirm.invalid };

  if (!apiKey) {
    if (!isDev) {
      console.error("[signup] RESEND_API_KEY is not set.");
      return { status: "error", message: confirm.unavailable };
    }
    console.info(`[signup] RESEND_API_KEY is not set, so nothing was saved. Would add ${claim.email} (role: ${claim.role || "none"}).`);
    return { status: "done" };
  }

  const resend = new Resend(apiKey);
  // The role is a custom contact property named `role` (create it once in
  // Resend under Audience, Properties). Leave it out when they skipped it.
  const properties = claim.role ? { role: claim.role } : undefined;
  const created = await resend.contacts.create({
    email: claim.email,
    unsubscribed: false,
    properties,
    segments: segmentId ? [{ id: segmentId }] : undefined,
  });
  if (created.error) {
    // Usually: the address is already a contact, perhaps one who unsubscribed
    // before. They have just confirmed again, so subscribe them again.
    const updated = await resend.contacts.update({ email: claim.email, unsubscribed: false, properties });
    if (updated.error) {
      console.error("[signup] Resend couldn't save the contact:", created.error, updated.error);
      return { status: "error", message: confirm.unavailable };
    }
    if (segmentId) {
      const added = await resend.contacts.segments.add({ email: claim.email, segmentId });
      if (added.error) console.error("[signup] Resend couldn't add the contact to the segment:", added.error);
    }
  }
  await record("Signup confirmed", { role: claim.role || "none" });
  return { status: "done" };
}
