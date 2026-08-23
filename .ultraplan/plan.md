# Implementation Plan: Legacy Reach L1–L4 (old-device driver surface)

## Context
idev's owner wants legacy reach: devices from the checkm8 era down to iOS 5,
as DRIVER verbs only (no exploit code, no downgrade orchestration). Four
items were scoped in docs/RESEARCH.md: L1 recovery/DFU visibility, L2
Developer-Disk-Image mounting for ≤16, L3 legacy lockdown tolerance (iOS
5–9 quirks), L4 offline MBDB backup reader. Everything rides the existing
Handler seam + hermetic-test house style. Ship order chosen by
leverage-per-effort and testability against hardware we may not have.

## Assumptions (owner may veto)
- A1: No raw-USB/cgo dependency ever — breaks CGO_ENABLED=0 six-combo builds.
  Recovery-mode *control* beyond what usbmuxd relays stays out of scope.
- A2: L3 needs real old hardware to be meaningful; we ship graceful
  degradation + a documented hardware runbook, not guessed protocol hacks.
- A3: Backup *restore* (M7) is downstream of L4 and NOT part of L1–L4.
- A4: `enter-recovery` only if lockdownd offers it cleanly on modern iOS;
  exiting recovery depends on what usbmuxd relays (validated in B2 step 1).

## Changes

### Batch A — L4: MBDB backup reader (hermetic, ship first)
- **File**: `device/mbdb.go` (new ~250 lines)
  - `ParseMBDB(data []byte) ([]MBDBRecord, error)` — magic `6d 62 64 62 05 00`,
    records of 11 length-prefixed strings (domain, path, linktarget, datahash,
    encryption-key) + u8 modeHigh, u64 inodeID, u32 mode/uid/gid/mtime/atime/
    ctime, u64 length, u8 propCount, N× (2BL key, 2BL value)
  - `ReadBackup(dir string) (*BackupInfo, error)` — loads Status.plist +
    Info.plist + Manifest.mbdb (pre-iOS10 layout); rejects iOS10+ SQLite
    manifests with a clear pointer message
  - Mode decode mirrors tar semantics (S_IFDIR 0x4000 etc.)
- **File**: `device/mbdb_test.go` — synthetic fixture builder (write mbdb bytes
  in-test); round-trip property: parse → re-serialize → byte-identical;
  truncation fuzz target `FuzzMBDBNoPanic` (30s)
- **File**: `cmd/idev/backupcmd.go`
  - `idev backup inspect <dir> [--domain d] [--app bundle] [--json]` —
    inventory: file count, total size, date range, top-level domains
  - `idev backup inspect --tree <prefix>` — path listing under a domain prefix
- **Tests**: cmd-level table tests against fixture dirs; no device needed

### Batch B — L1: recovery/DFU visibility (protocol work, small)
- Step 1 (validation): capture raw usbmuxd Listen/attached plist for a device
  in Recovery vs DFU vs normal (hardware-in-loop script `scripts/recovery-capture.sh`,
  output committed under testdata/) — establishes exact field shapes
- **File**: `device/recovery.go` (new)
  - Extend attach classification: `Dev.State string` ("normal","recovery","dfu","unknown")
    derived from usbmuxd Properties (ProductType/ConnectionType shapes found in step 1)
  - `RecoveryExit(ctx, udid)` — whatever usbmuxd relays; if nothing relays,
    return typed KindUnsupported with the honest remediation (raw USB needed)
- **File**: `device/device.go` — Dev gains State; Handler unchanged unless verbs need it
- **File**: `cmd/idev/root.go` — `list` prints `[state]`; new `idev recovery exit`
- **Tests**: fixture-plist classification table tests; live-tagged end-to-end

### Batch C — L2: DDI mounter ≤16 (thin CLI over imagemounter)
- **File**: `device/ddi.go` (new)
  - `DDIMount(ctx, udid, imagePath string) error` — gate: ProductVersion major
    ≤16 (else typed KindUnsupported pointing at tunnels), DevModeEnabled check first
  - `DDIStatus(ctx, udid) (mounted bool, images []string, err error)`
  - `DDIUnmount(ctx, udid) error`
  - Optional `--download`: imagemounter.DownloadImageFor (network, flagged loudly)
- **File**: `cmd/idev/ddicmd.go` — `idev ddi mount|status|unmount`
- **Tests**: version-gate predicate table tests (reuse shouldAutoTunnel shape);
  live-tagged mount on real ≤16 hardware

### Batch D — L3: legacy lockdown tolerance (hardware-in-the-loop)
- **File**: `device/legacy.go` (new)
  - Version-aware degradation map: keys known missing pre-iOS7/10
    (CPUArchitecture, WiFiSync settings…) → Info tolerates absence (already partial)
  - Pairing-record path notes per host OS (go-ios owns storage; document only)
- **File**: `docs/LEGACY-MATRIX.md` — support matrix template (iOS 5…16 × verb),
  filled from hardware runs; each discovered quirk becomes a regression test
  with a fixture plist
- **Runbook**: `scripts/legacy-sweep.sh` — runs every read-only verb against a
  target device, captures outputs for the matrix
- **Gate rule**: no quirk fix lands without its fixture test

## Order & batching
A → B → C → D. A is fully hermetic (any wake-up). B/C mix protocol work with
one hardware-validation step each — hardware steps block only their final
live-tagged test, not the merge of gated code. D is explicitly hardware-first.

## Risks
- R1: usbmuxd recovery-message shapes differ per OS version of usbmuxd itself
  (macOS vs iTunes-installed Windows) → fixtures captured per-host, classifier
  keyed on observed fields not assumptions
- R2: DDI download URLs rot (Apple/GitHub moves) → image path flag keeps local
  use alive; downloader is best-effort with honest errors
- R3: old hardware unavailability stalls D/B-final → all code merges behind
  gates; matrix rows stay honestly marked "unverified"
