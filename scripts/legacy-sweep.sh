#!/usr/bin/env bash
# legacy-sweep: run every read-only idev verb against the connected device
# and capture outputs for docs/LEGACY-MATRIX.md. Read-only by design —
# nothing here pairs, installs, deletes or reboots.
#
# usage: UDID=<udid> bash scripts/legacy-sweep.sh [outdir]
set -uo pipefail

UDID="${UDID:?export UDID=<the device's udid>}"
OUT="${1:-legacy-capture-$(date +%Y%m%d-%H%M%S)}"
mkdir -p "$OUT"

run() { # name, args...
  local name="$1"; shift
  echo "== idev $* ==" | tee "$OUT/$name.txt"
  idev --udid "$UDID" "$@" >"$OUT/$name.body" 2>"$OUT/$name.err"
  echo "exit=$?" | tee -a "$OUT/$name.txt"
  head -c 2000 "$OUT/$name.body"
}

run info info
run battery battery
run apps apps
run devmode devmode
run files-ls files ls /
run crash-list crash list
run syslog-sample timeout 3 idev --udid "$UDID" syslog >"$OUT/syslog.body" 2>&1; echo "syslog exit=$?"
run ddi-status ddi status
run watch-sample timeout 2 idev --udid "$UDID" watch >/dev/null 2>&1; echo "watch exit=$?"

echo
echo "captured to $OUT/ — paste relevant rows into docs/LEGACY-MATRIX.md"
