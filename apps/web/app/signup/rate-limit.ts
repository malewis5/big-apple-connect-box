// A small in-memory limit on how often one address or internet address can
// use the signup form. It lives in each server instance's memory, so it is a
// speed bump against floods and form-filling bots, not a guarantee; nothing
// is written down, and entries are dropped once their window passes.
// For a hard limit, add a Vercel Firewall rate-limit rule on POST /.

// Each entry keeps its own window, because the form uses different ones: ten
// minutes for an internet address, an hour for an email address.
const hits = new Map<string, { windowMs: number; times: number[] }>();
let lastSweep = 0;
const SWEEP_EVERY = 10 * 60_000;

export function allow(key: string, max: number, windowMs: number, now = Date.now()) {
  // Forget old entries now and then, so the map can't grow without end. Each
  // entry is judged by its own window, so a sweep never drops an email
  // address's hourly count early.
  if (now - lastSweep > SWEEP_EVERY) {
    for (const [k, entry] of hits) if (entry.times.every((t) => now - t >= entry.windowMs)) hits.delete(k);
    lastSweep = now;
  }
  const recent = (hits.get(key)?.times ?? []).filter((t) => now - t < windowMs);
  if (recent.length >= max) {
    hits.set(key, { windowMs, times: recent });
    return false;
  }
  recent.push(now);
  hits.set(key, { windowMs, times: recent });
  return true;
}
