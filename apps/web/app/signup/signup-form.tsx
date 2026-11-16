"use client";

import { useActionState, useId, type ReactNode } from "react";
import { requestSignup, type SignupState } from "./actions";

// The words the form shows, resolved on the server from messages/en.json
// (Home.signup). Messages after sending come back from the action.
export type SignupLabels = {
  emailLabel: string;
  emailPlaceholder: string;
  roleLabel: string;
  optional: string;
  roleChoose: string;
  roles: { value: string; label: string }[];
  honeypot: string;
  submit: string;
  submitting: string;
  successTitle: string;
  privacy: ReactNode;
};

const INITIAL: SignupState = { status: "idle" };

// The signup for email updates. It sits on the dark card at the foot of the
// home page, so its colors are the light-on-dark versions of the site's own.
// Without JavaScript it still works: the form posts to the Server Action and
// the page comes back with the result.
export default function SignupForm({ labels }: { labels: SignupLabels }) {
  const [state, action, pending] = useActionState(requestSignup, INITIAL);
  const id = useId();

  if (state.status === "sent") {
    return (
      <div role="status" className="rounded-2xl bg-white/10 p-6 ring-1 ring-white/15 sm:p-8">
        <p className="flex items-center gap-3 text-xl font-semibold text-white">
          <span className="flex size-8 shrink-0 items-center justify-center rounded-full bg-emerald-400 text-zinc-900">
            <svg viewBox="0 0 20 20" className="size-4" fill="none" stroke="currentColor" strokeWidth="2.5" aria-hidden="true">
              <path d="m5 10.5 3.5 3.5L15 7" strokeLinecap="round" strokeLinejoin="round" />
            </svg>
          </span>
          {labels.successTitle}
        </p>
        <p className="mt-3 text-lg leading-relaxed text-zinc-300">{state.message}</p>
      </div>
    );
  }

  const error = state.status === "error" ? state : null;
  const emailError = error?.field === "email";

  return (
    <form action={action} className="relative min-w-0">
      <div>
        <label htmlFor={`${id}-email`} className="block text-sm font-medium text-white">
          {labels.emailLabel}
        </label>
        <input
          id={`${id}-email`}
          name="email"
          type="email"
          required
          autoComplete="email"
          inputMode="email"
          spellCheck={false}
          maxLength={254}
          defaultValue={error?.email}
          placeholder={labels.emailPlaceholder}
          aria-invalid={emailError || undefined}
          aria-describedby={error ? `${id}-error` : undefined}
          className="mt-2 block min-h-12 w-full rounded-xl bg-white px-4 py-3 text-base text-zinc-900 placeholder:text-zinc-500 focus:outline-2 focus:outline-offset-2 focus:outline-emerald-400 aria-[invalid]:outline-2 aria-[invalid]:outline-red-300"
        />
      </div>

      <div className="mt-5">
        <label htmlFor={`${id}-role`} className="block text-sm font-medium text-white">
          {labels.roleLabel} <span className="font-normal text-zinc-300">({labels.optional})</span>
        </label>
        <select
          id={`${id}-role`}
          name="role"
          defaultValue={error?.role ?? ""}
          className="mt-2 block min-h-12 w-full rounded-xl bg-white px-4 py-3 text-base text-zinc-900 focus:outline-2 focus:outline-offset-2 focus:outline-emerald-400"
        >
          <option value="">{labels.roleChoose}</option>
          {labels.roles.map((role) => (
            <option key={role.value} value={role.value}>
              {role.label}
            </option>
          ))}
        </select>
      </div>

      {/* For bots only: hidden from people and screen readers, skipped by Tab.
          A filled-in value means a bot, and nothing is sent. */}
      <div aria-hidden="true" className="absolute -left-[9999px] size-px overflow-hidden">
        <label htmlFor={`${id}-company`}>{labels.honeypot}</label>
        <input id={`${id}-company`} name="company" type="text" tabIndex={-1} autoComplete="off" defaultValue="" />
      </div>

      {error && (
        <p id={`${id}-error`} role="alert" className="mt-5 text-base leading-relaxed text-red-200">
          {error.message}
        </p>
      )}

      <button
        type="submit"
        disabled={pending}
        className="mt-6 flex min-h-12 w-full items-center justify-center rounded-full bg-white px-6 py-3 text-base font-semibold text-zinc-900 transition hover:bg-zinc-200 disabled:cursor-wait disabled:opacity-80 sm:w-auto"
      >
        {pending ? labels.submitting : labels.submit}
      </button>
      <p className="mt-4 text-sm leading-relaxed text-zinc-300">{labels.privacy}</p>
    </form>
  );
}
