#!/bin/sh
# Installs the Big Apple Connect Box on a Raspberry Pi or other Linux system
# that uses systemd. Run it as root, from a folder that holds this script,
# the big-apple-connect-box program and the two systemd files:
#
#	sudo sh install.sh
#
# Running it again installs a newer program over the old one. Saved scans
# are kept.
set -eu

name=big-apple-connect-box
cd "$(dirname "$0")"

# The account the program runs as. It can't log in and owns only its scans.
if ! id "$name" >/dev/null 2>&1; then
	useradd --system --no-create-home --shell /usr/sbin/nologin "$name"
fi

install -m 755 "$name" /usr/local/bin/
install -m 644 "$name.service" "$name.timer" /etc/systemd/system/
systemctl daemon-reload

# From the next boot, hold time-sync.target until the clock has synced.
systemctl enable systemd-time-wait-sync.service

systemctl enable --now "$name.timer"

echo "Installed. Scans are saved to /var/lib/$name/scans.db."
systemctl list-timers "$name.timer" --no-pager
