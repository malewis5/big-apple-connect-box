"use client";

import Link from "next/link";
import { useActionState, useSyncExternalStore } from "react";
import { confirmSignup, type ConfirmState } from "../signup/actions";

export type ConfirmLabels = {
  body: string;
  button: string;
  confirming: string;
  successTitle: string;
  successBody: string;
  missing: string;
  back: string;
};

// The code from the confirmation link, which comes after the # so the server
// never sees it in a page request. Read straight from the address bar; on the
// server (and in the prerendered page) there is none yet.
function subscribe(onChange: () => void) {
  window.addEventListener("hashchange", onChange);
  return () => window.removeEventListener("hashchange", onChange);
}
const readCode = () => decodeURIComponent(window.location.hash.slice(1));
const noCode = () => null;

const INITIAL: ConfirmState = { status: "idle" };

const BUTTON =
  "inline-flex min-h-12 items-center justify-center gap-2 rounded-full bg-zinc-900 px-6 py-3 text-base font-semibold text-white transition hover:bg-zinc-700 disabled:cursor-wait disabled:opacity-80";
const BACK =
  "inline-flex min-h-11 items-center font-medium text-emerald-800 underline underline-offset-4 hover:text-emerald-900";

export default function ConfirmForm({ labels }: { labels: ConfirmLabels }) {
  const code = useSyncExternalStore(subscribe, readCode, noCode);
  const [state, action, pending] = useActionState(confirmSignup, INITIAL);

  if (state.status === "done") {
    return (
      <div role="status" className="mt-8">
        <p className="flex items-center gap-3 text-2xl font-semibold">
          <span className="flex size-9 shrink-0 items-center justify-center rounded-full bg-emerald-50 text-emerald-700">
            <svg viewBox="0 0 20 20" className="size-5" fill="none" stroke="currentColor" strokeWidth="2.5" aria-hidden="true">
              <path d="m5 10.5 3.5 3.5L15 7" strokeLinecap="round" strokeLinejoin="round" />
            </svg>
          </span>
          {labels.successTitle}
        </p>
        <p className="mt-4 text-lg leading-relaxed text-zinc-600">{labels.successBody}</p>
        <Link href="/" className={`mt-6 ${BACK}`}>
          {labels.back}
        </Link>
      </div>
    );
  }

  // Still loading in the browser: show nothing rather than a false "missing".
  if (code === null) return <div className="mt-8 min-h-12" />;

  if (!code) {
    return (
      <div className="mt-8">
        <p className="text-lg leading-relaxed text-zinc-600">{labels.missing}</p>
        <Link href="/#updates" className={`mt-4 ${BACK}`}>
          {labels.back}
        </Link>
      </div>
    );
  }

  return (
    <form action={action} className="mt-8">
      <p className="text-lg leading-relaxed text-zinc-600">{labels.body}</p>
      <input type="hidden" name="token" value={code} />
      <button type="submit" disabled={pending} className={`mt-6 ${BUTTON}`}>
        {pending ? labels.confirming : labels.button}
      </button>
      {state.status === "error" && (
        <div role="alert" className="mt-6">
          <p className="text-lg leading-relaxed text-red-800">{state.message}</p>
          <Link href="/#updates" className={`mt-2 ${BACK}`}>
            {labels.back}
          </Link>
        </div>
      )}
    </form>
  );
}
