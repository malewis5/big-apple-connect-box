"use client";

import { track } from "@vercel/analytics";
import { useEffect, useRef, useState, type ReactNode } from "react";
import { DEMO_VIDEO as VIDEO } from "./site";
import { CARD } from "./ui";

// AV1 is the smallest, for browsers that can play it; H.264 plays everywhere.
const DEMO_VIDEO = VIDEO.files;

// The words on and around the player, from messages/en.json (Home.video).
// The server resolves them and passes them in, so no catalog reaches the browser.
export type DemoVideoLabels = {
  label: string;
  play: string;
  watch: string;
  playAgain: string;
  captions: string;
  fallback: ReactNode;
};

// Many people reading this page are on the connection they want to check, so
// the video never starts by itself and downloads nothing until someone presses
// play. Until then the poster and one play button stand in for it.

function setTrackMode(track: TextTrack, mode: TextTrackMode) {
  track.mode = mode;
}

// The player alone. The "How it works" section on the home page places it.
export default function DemoVideo({ labels }: { labels: DemoVideoLabels }) {
  const video = useRef<HTMLVideoElement>(null);
  const [caption, setCaption] = useState(""); // kept while it fades out
  const [captionShown, setCaptionShown] = useState(false);
  // "idle" shows the poster and play button; "starting" keeps the poster up
  // until the first frame arrives; "ended" rests on the last frame with a
  // button to play again.
  const [stage, setStage] = useState<"idle" | "starting" | "playing" | "ended">("idle");

  useEffect(() => {
    const el = video.current;
    if (!el) return;
    const onEnded = () => setStage("ended");
    const onPlaying = () => setStage("playing");
    el.addEventListener("ended", onEnded);
    el.addEventListener("playing", onPlaying);
    return () => {
      el.removeEventListener("ended", onEnded);
      el.removeEventListener("playing", onPlaying);
    };
  }, []);

  function play() {
    const el = video.current;
    if (!el) return;
    if (stage === "ended") el.currentTime = 0;
    track("Video played", { again: stage === "ended" });
    setStage((s) => (s === "idle" ? "starting" : s));
    el.play().catch(() => setStage((s) => (s === "starting" ? "idle" : s)));
  }

  // Captions in the page's own style. Browsers draw captions their own way
  // (Chrome ignores a caption background), so CSS hides the native cues inline
  // and the page draws each line as the same white pill as the shared copies.
  // Keep the track showing so the player's caption menu reflects its state.
  // In full screen, where nothing of the page shows, the browser draws them.
  // Turning captions off in the player's menu turns the pill off too.
  useEffect(() => {
    const el = video.current;
    const track = el?.textTracks[0];
    if (!el || !track) return;
    let on = true;
    const fullScreen = () =>
      document.fullscreenElement === el || (el as HTMLVideoElement & { webkitDisplayingFullscreen?: boolean }).webkitDisplayingFullscreen;
    const show = () => {
      const cue = track.activeCues?.[0] as VTTCue | undefined;
      const visible = on && !fullScreen() && !!cue;
      if (visible) setCaption(cue.text);
      setCaptionShown(visible);
    };
    const sync = () => {
      if (track.mode === "disabled") on = false;
      else on = true;
      el.classList.toggle("demo-inline-captions", !fullScreen());
      const want = on ? "showing" : "disabled";
      if (track.mode !== want) setTrackMode(track, want); // fires change again, which then matches
      show();
    };
    sync();
    track.addEventListener("cuechange", show);
    el.textTracks.addEventListener("change", sync);
    document.addEventListener("fullscreenchange", sync);
    el.addEventListener("webkitbeginfullscreen", sync);
    el.addEventListener("webkitendfullscreen", sync);
    return () => {
      track.removeEventListener("cuechange", show);
      el.textTracks.removeEventListener("change", sync);
      document.removeEventListener("fullscreenchange", sync);
      el.removeEventListener("webkitbeginfullscreen", sync);
      el.removeEventListener("webkitendfullscreen", sync);
    };
  }, []);

  const covered = stage === "idle" || stage === "starting";

  return (
    <div
      className={`relative w-full min-w-0 overflow-hidden [container-type:inline-size] ${CARD}`}
    >
      <video
        ref={video}
        className="block aspect-video w-full bg-zinc-100 object-cover"
        width={1920}
        height={1080}
        preload="none"
        muted
        playsInline
        controls={stage === "playing"}
        aria-label={labels.label}
      >
        <source src={DEMO_VIDEO.webm} type='video/webm; codecs="av01.0.08M.10"' />
        <source src={DEMO_VIDEO.mp4} type="video/mp4" />
        <track src={DEMO_VIDEO.captions} kind="captions" srcLang="en" label={labels.captions} default />
        <p className="p-6 text-zinc-600">{labels.fallback}</p>
      </video>
      {/* Sized to the video's width, like the captions in the shared copies:
          about 2.4% of the width, and small and tucked low on a phone. The words are also in
          the track and on the Markdown page. */}
      <div
        aria-hidden="true"
        className="pointer-events-none absolute inset-x-0 bottom-[16%] flex justify-center px-[4%] max-sm:bottom-[7%] max-sm:px-3"
      >
        <p
          className={`max-w-full rounded-full bg-white/95 px-[1.1em] py-[0.45em] text-center text-[max(0.8rem,2.4cqw)] leading-snug font-medium text-pretty text-zinc-900 shadow-sm transition-opacity duration-300 motion-reduce:transition-none max-sm:rounded-2xl max-sm:px-2.5 max-sm:py-1 max-sm:text-[0.6875rem] ${captionShown ? "opacity-100" : "opacity-0"}`}
        >
          {caption || "\u00a0"}
        </p>
      </div>
      {/* Before it plays: the poster, and one large
          light button over it. The whole picture is the target. */}
      {covered && (
        <button
          type="button"
          onClick={play}
          aria-label={labels.play}
          className="group absolute inset-0 flex cursor-pointer items-center justify-center focus-visible:outline-offset-[-4px]"
        >
          {/* eslint-disable-next-line @next/next/no-img-element -- the poster is a pre-sized WebP; the optimizer adds nothing */}
          <img
            src={DEMO_VIDEO.poster}
            alt=""
            className={`absolute inset-0 size-full object-cover transition duration-500 ease-out group-hover:scale-[1.02] motion-reduce:transition-none ${stage === "starting" ? "opacity-90" : ""}`}
          />
          <span aria-hidden="true" className="absolute inset-0 bg-zinc-950/10 transition-colors duration-300 group-hover:bg-zinc-950/0" />
          {/* The circle sits at the exact center; the label hangs below it. */}
          <span className="relative">
            <span className="relative flex size-16 items-center justify-center sm:size-24">
              {stage === "idle" && (
                <span
                  aria-hidden="true"
                  className="absolute inset-0 animate-play-halo rounded-full bg-white motion-reduce:hidden"
                />
              )}
              <span className="relative flex size-full items-center justify-center rounded-full bg-white/95 text-zinc-900 shadow-[0_8px_30px_-6px_rgb(0_0_0/0.35)] ring-1 ring-zinc-900/10 backdrop-blur-sm transition duration-300 ease-out group-hover:scale-110 group-hover:bg-white group-active:scale-95 motion-reduce:transition-none">
                {stage === "starting" ? (
                  <span
                    aria-hidden="true"
                    className="size-7 animate-spin rounded-full border-[3px] border-zinc-300 border-t-zinc-900 motion-reduce:animate-none"
                  />
                ) : (
                  // The triangle's centroid (not its box) sits at the
                  // viewBox center, which is where it looks centered.
                  <svg viewBox="0 0 24 24" aria-hidden="true" className="size-7 sm:size-9" fill="currentColor">
                    <path d="M8 5.6v12.8a1 1 0 0 0 1.5.86l10.2-6.4a1 1 0 0 0 0-1.72L9.5 4.74A1 1 0 0 0 8 5.6Z" />
                  </svg>
                )}
              </span>
            </span>
            <span
              aria-hidden="true"
              className="absolute top-full left-1/2 mt-2 -translate-x-1/2 sm:mt-3 rounded-full bg-white/95 px-3 py-1 whitespace-nowrap text-sm font-medium text-zinc-900 tabular-nums shadow-sm ring-1 ring-zinc-900/10"
            >
              {labels.watch} &middot; 0:{String(VIDEO.seconds).padStart(2, "0")}
            </span>
          </span>
        </button>
      )}
      {/* At the end, one light button over the last frame, like the
          caption pills. The player's dark controls hide meanwhile. */}
      {stage === "ended" && (
        <div className="absolute inset-0 flex items-center justify-center">
          <button
            type="button"
            onClick={play}
            className="inline-flex min-h-11 items-center gap-2 rounded-full bg-white/95 px-5 py-2.5 text-sm font-medium text-zinc-900 shadow-sm ring-1 ring-zinc-900/10 transition hover:bg-white"
          >
            <svg
              viewBox="0 0 20 20"
              aria-hidden="true"
              className="size-4"
              fill="none"
              stroke="currentColor"
              strokeWidth="1.75"
              strokeLinecap="round"
              strokeLinejoin="round"
            >
              <path d="M3.5 10a6.5 6.5 0 1 0 2-4.7" />
              <path d="M3.5 3.5v3.5H7" />
            </svg>
            {labels.playAgain}
          </button>
        </div>
      )}
    </div>
  );
}
