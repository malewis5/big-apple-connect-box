// A small in-memory limit on how often one address or internet address can
// use the signup form. It lives in each server instance's memory, so it is a
// speed bump against floods and form-filling bots, not a guarantee; nothing
// is written down, and entries are dropped once their window passes.
// For a hard limit, add a Vercel Firewall rate-limit rule on POST /.

const hits = new Map<string, number[]>();
let lastSweep = 0;

export function allow(key: string, max: number, windowMs: number, now = Date.now()) {
  // Forget old entries now and then, so the map can't grow without end.
  if (now - lastSweep > windowMs) {
    for (const [k, times] of hits) if (times.every((t) => now - t > windowMs)) hits.delete(k);
    lastSweep = now;
  }
  const recent = (hits.get(key) ?? []).filter((t) => now - t < windowMs);
  if (recent.length >= max) {
    hits.set(key, recent);
    return false;
  }
  recent.push(now);
  hits.set(key, recent);
  return true;
}
