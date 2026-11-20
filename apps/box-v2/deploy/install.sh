#!/bin/sh
# Installs the Big Apple Connect Box on a Raspberry Pi or other Linux system
# that uses systemd. Run it as root, from a folder that holds this script,
# the big-apple-connect-box program and the four systemd files:
#
#	sudo sh install.sh
#
# Running it again installs a newer program over the old one. Saved scans
# and pings are kept.
set -eu

name=big-apple-connect-box
cd "$(dirname "$0")"

# The account the program runs as. It can't log in and owns only what the
# program saves.
if ! id "$name" >/dev/null 2>&1; then
	useradd --system --no-create-home --shell /usr/sbin/nologin "$name"
fi

install -m 755 "$name" /usr/local/bin/
install -m 644 "$name.service" "$name.timer" "$name-ping.service" "$name-ping.timer" /etc/systemd/system/
systemctl daemon-reload

# From the next boot, hold time-sync.target until the clock has synced. The
# speed test's timer waits for that. The pings' timer doesn't.
systemctl enable systemd-time-wait-sync.service

# Restarted, not only started, so that a timer already running takes up a
# changed file. A restart doesn't run a speed test, and the pings carry on a
# minute after their last round.
systemctl enable "$name.timer" "$name-ping.timer"
systemctl restart "$name.timer" "$name-ping.timer"

echo "Installed. Scans and pings are saved to /var/lib/$name/scans.db."
systemctl list-timers "$name.timer" "$name-ping.timer" --no-pager
