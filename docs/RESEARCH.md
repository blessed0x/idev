# idev ecosystem leverage research

Every tool in the usbmuxd/libimobiledevice orbit that could give idev
leverage — alive or dead, famous or niche — with the merge verdict.
Owner directive: license concerns are waived for this project; we port
functionality into Go rather than wrapping binaries.

## The landscape (verified 2026-08-23)

| project | ★ | state | what it gives us | verdict |
|---|---|---|---|---|
| libimobiledevice suite (libimobiledevice, usbmuxd, ifuse, ideviceinstaller) | 8.1k/1.7k/1k/1.4k | slow-alive | the protocol reference implementations; semantics oracle for every service we surface | **reference only** — we already ride go-ios |
| idevicebackup2 | part of limd | slow-alive | THE backup/restore CLI: full-device backup, restore, system backups | **M7 flagship** — biggest missing driver feature (see plan below) |
| pymobiledevice3 | 2.6k | very alive | best-documented semantics for developer services; our UX benchmark | reference for behavior parity per milestone |
| go-ios | 2.2k | very alive | our engine — all transports below already vendored | merged by construction |
| idevicerestore | 1.9k | alive | firmware restore/upgrade orchestration | **out of scope** (restore = bricking territory, not a driver verb) |
| libirecovery / irecovery | 650 | alive | talks to iBoot/iBSS in DFU & Recovery mode over USB | **legacy M-L1**: recovery-mode device detection via usbmuxd + reboot-out-of-recovery; full DFU client needs raw USB (cgo) — deferred |
| gidevice (Go) | 294 | dormant 2023-12 | prior Go art: confirms demand, some API shapes worth reading | mine for ideas, do not depend |
| usbfluxd (core-dev) | small | semi-alive | relays usbmuxd across TCP so network machines see your phone | **M-N1**: `idev bridge <host>` would give WiFi-everywhere without pairing dance |
| mobiledevice (homebrew C tool) | ~dead | dead | proved the "tiny CLI for install/syslog/backup" niche; its users are our users | inspiration; feature checklist absorbed |
| tcprelay.py / itunnel / ios-deploy | dead classics | dead | each died doing one thing we already ship better (forward/install/watch) | validated demand; nothing to port |
| iMazing (closed) | commercial | alive | the product spec for backup/file UX: app sandbox browsing, export wizards | UX benchmark for `files` + future `backup` |
| tsschecker / futurerestore world | alive | niche | SHSH blob saving + downgrade orchestration | **out of scope**: jailbreak-adjacent by owner directive |

## Legacy reach: checkm8 era and iOS 5–12

What works TODAY through plain usbmuxd/lockdownd (no exploit needed):

- The usbmuxd wire protocol is unchanged since iOS 2 — old devices enumerate,
  pair and serve lockdownd identically. Most verbs (list/info/battery/apps/
  install/forward/pasteboard/watch) should simply work on iOS 5–12 hardware.
- syslog classic relay exists ≤16 — richer than on modern devices (raw format).
- Crash logs live on the device since iOS 4 era (`crash` verb covers them).
- AFC media partition + house_arrest sandboxes exist since iOS 4 (`files`).

Legacy-specific work queued:

| item | era | notes |
|---|---|---|
| L1 recovery/DFU visibility | all | usbmuxd reports recovery-mode devices; detect + label them in `list`, offer `recovery exit` (auto-boot / reboot) — irecovery's safe subset |
| L2 DDI mounter for ≤16 | 8–16 | imagemounter + Developer Disk Image download; unblocks launch/instruments on old devices without Xcode |
| L3 legacy lockdown quirks | 5–9 | pre-iOS10 wants plaintext lockdown sessions before pairing completes; test matrix needed against real old hardware |
| L4 MBDB backup reader | ≤9 | offline parser for old-format backups on disk (pure Go, no device needed) — pairs with M7 |
| out of scope | any | checkm8/pwned-DFU exploitation itself (owner directive: driver, not jailbreak) |

## M7 flagship plan: backup2 (the idevicebackup2/iMazing category)

The single highest-value missing driver feature. Port of the MobileBackup2
device-link flow (same framing Omega already drives for restores, so the
house knows the protocol):

1. `backup` — full or incremental (plist manifest + mbdb-style file store),
   password handling for encrypted backups
2. `backup list` — inventory a backup on disk (apps, sizes, dates)
3. `restore` — settings/app data restore with explicit destructive gates
4. L4 reader for pre-iOS10 MBDB archives

Effort estimate: the link-layer plumbing exists (device-link frames from
omega.go); the remaining work is manifest math and will-change detection —
roughly the size of the whole current device package, split across several
batches.

## Already shipped from this research (this batch)

- `location set lat,lon | clear | gpx` — simlocation
- `crash list/pull/clear [--pattern]` — crashreport store
- `files ls/pull/push/rm/mkdir [-a bundle]` — AFC media partition +
  house_arrest sandboxes (the iMazing-lite core)
