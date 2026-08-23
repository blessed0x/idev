<div align="center">

# idev

**Your iPhone/iPad, from the terminal. In Go.**

Pair it, install to it, launch apps, stream logs, take screenshots,
forward ports, read the clipboard — one static binary, no pip, no Python.

*The pymobiledevice3 everyday surface, rebuilt pure-Go.*

</div>

---

## Why

[pymobiledevice3](https://github.com/doronz88/pymobiledevice3) proved how much
an iPhone can do over plain usbmuxd/lockdownd — but using it means managing a
Python environment for every machine (and every CI runner). **idev** ships the
everyday 80% of that surface as one ~15 MB static binary built on
[danielpaulus/go-ios](https://github.com/danielpaulus/go-ios), with the parts
people actually script daily polished first: typed errors with real remediation
text, deterministic exit codes, `--json` everywhere, and an iOS 17+ developer
tunnel that manages itself.

## Install

```bash
go install github.com/xscope0/idev/cmd/idev@latest        # any OS with Go
```

Prebuilt binaries land with releases (macOS/Linux/Windows × arm64/amd64).

## Commands

| command | what it does |
|---|---|
| `idev list` | who's plugged in (USB + WiFi deduped by UDID) |
| `idev doctor` | probe usbmuxd + per-OS fix when it's down |
| `idev pair` | trust dialog pairing (`--supervised p12:pw` for supervised) |
| `idev info` | name, iOS version, model, arch (`--json`) |
| `idev battery` | charge level + state |
| `idev apps [--system]` | installed apps |
| `idev install App.ipa` | Xcode zip-conduit install |
| `idev uninstall <bundle>` | remove an app |
| `idev launch <bundle>` | run it, print pid (`--env`, args) |
| `idev kill <pid>` | stop a process |
| `idev syslog [--process p] [--contains s]` | parsed live logs; iOS 17+ streams os_trace over a tunnel idev starts for you |
| `idev watch` | attach/detach events |
| `idev screenshot [f.png]` | screen capture (timestamped default) |
| `idev devmode` | iOS 16+ Developer Mode switch status |
| `idev forward <host> <dev>` | iproxy-style port relay (every iOS version) |
| `idev pasteboard get\|set` | device clipboard |
| `idev restart` / `idev shutdown` | power control (`--yes` for scripts) |
| `idev omega` | revoke/cert blacklist remover (jailbreak.party Omega port) |

## How it compares to pymobiledevice3

Honest table, because you should pick tools on facts:

**idev is better at**

- **Zero environment**: one static binary vs pip + virtualenv + Python version pinning. CI just works.
- **Self-managed tunnels**: when an operation hits the iOS 17+ CoreDevice gate, idev spawns the userspace tunnel itself, waits for readiness, retries, and tears it down on exit. pymobiledevice3 asks you to run its remote-tunnel service in another terminal.
- **Scriptability contract**: every failure is a typed error with an exit code (64–70 by class: usage/not-found/unsupported/permission/timeout/connection) and remediation text written for humans.
- **Transparent syslog on stock iOS 17+**: classic relay removed by Apple? idev falls back to os_trace over the same auto-tunnel without you changing commands.
- **UDID-level device identity**: one phone on USB + WiFi shows once, preferred transport chosen deterministically; ambiguous targets are refused, never guessed.
- **Omega**: the jailbreak.party revoke/cert blacklist remover, with a sourced per-iOS-version verdict (16–18 proven, 26.x live-verified, 27 hard-blocked). Nothing like it exists elsewhere.
- **Cold start & size**: sub-100ms invocations vs Python interpreter boot per call.

**pymobiledevice3 is ahead today**

backup2 (full backup/restore), afc file browsing, profile management, crash-log
collection, packet capture, GPS simulation, notification proxy, SpringBoard
control, webinspector REPL, activation/ecc flows, and sheer community mileage.

**The roadmap closes most of that gap fast** — go-ios already ships the
transports (afc, misagent, crashreport, simlocation, notificationproxy, pcap),
so parity there is mostly CLI surfacing + tests, not protocol work:

| milestone | adds |
|---|---|
| M1 | `afc pull/push/ls/rm` (app + media filesystems) |
| M2 | `crash copy/live/clear` |
| M3 | `profile list/install/remove` |
| M4 | `location <lat> <lon>` simulation |
| M5 | `notify post/observe` |
| M6 | `pcap` interface capture |

Want one of these next? Open an issue — the seam pattern (Handler interface +
hermetic fakes) makes each roughly a weekend.

## iOS version notes

- Pairing/info/battery/apps/forward/pasteboard/watch: **every iOS version**.
- launch/kill/install/screenshot: need a developer tunnel on **iOS 17+**
  (Developer Disk Image era before that) — idev handles both generations and
  starts the tunnel automatically when needed. `IDEV_NO_AUTO_TUNNEL=1` prints
  the manual command instead of spawning one.
- `omega`: supported 16–18 and 26 (live-verified on 26.1); 19–25 were never
  released; 27 hard-blocked (Apple changed the backup system); future versions
  get an untested warning.

## Development

```bash
make build   # → bin/idev
make test    # -race suite
make qa      # lint (gofmt/vet/tidy/staticcheck/govulncheck) + race + 6-combo cross-compile
```

House rules: every behavior change ships a test; network paths are
hermetic-httptest; hardware-dependent tests live behind `-tags live`.

## Credits

- [danielpaulus/go-ios](https://github.com/danielpaulus/go-ios) — the pure-Go Apple protocol stack doing the heavy lifting
- [doronz88/pymobiledevice3](https://github.com/doronz88/pymobiledevice3) — the original proof of how far lockdown services go; UX benchmark throughout
- [jailbreakdotparty/Omega](https://github.com/jailbreakdotparty/Omega) — the blacklist remover, ported faithfully with permission of its design
- [xscope0/xkvm-ios-injector](https://github.com/xscope0/xkvm-ios-injector) — idev's device subsystem was battle-tested there first

## License

MIT
