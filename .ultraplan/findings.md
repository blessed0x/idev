# Findings (exploration evidence, read before planning anything here)

## go-ios v1.3.2 facts (verified from module source on disk)

- `ios.ListDevices()` / `ios.Listen()` have NO recovery/DFU awareness — zero
  matches for recovery/dfu/autoboot in listdevices.go, listen.go,
  usbmuxconnection.go. Whatever usbmuxd reports for non-normal devices is
  either dropped or mislabeled by DeviceEntry parsing today.
  => L1 cannot reuse upstream; we extend classification ourselves from raw
  attached/list message fields. `device/watch` already exercises ios.Listen,
  so the plumbing to capture real messages exists in-repo.
- imagemounter is complete and public:
  - NewDeveloperDiskImageMounter(device, *semver.Version); MountImage(path),
    UnmountImage, ListImages
  - package funcs: MountImage(device,path) / UnmountImage(device) /
    IsDevModeEnabled(device)
  - imagedownloader: DownloadImageFor(device, baseDir) (network),
    MatchAvailable(version), Download17Plus (irrelevant to us ≤16 gate)
  - mount flow = sendUploadRequest → waitForUploadComplete → status check;
    deviceRefusedError exists (dev-mode-off shape)
- notificationproxy: Observe() registers but the receive loop
  (`newNotification`) is UNEXPORTED — external consumers cannot stream.
  Confirmed reason notify was deferred; a ~80-line plist-framing port over
  lockdown service com.apple.mobile.notification_proxy unlocks it later.

## MBDB format (training knowledge; fixture tests will pin it)
- Manifest.mbdb magic: `6D 62 64 62 05 00` ("mbdb" + 0x05 0x00)
- Record: len-prefixed(2B BE) strings ×5 [domain, path, linktarget, datahash,
  encryption-key] then u8 modeHigh, u64 inodeID(?), u32 mode, u32 uid, u32 gid,
  u32 mtime, u32 atime, u32 ctime, u64 length, u8 propCount, then propCount ×
  (2BL key + 2BL value). Records repeat until EOF.
- Pre-iOS10 backup dir layout: Info.plist, Status.plist, Manifest.plist,
  Manifest.mbdb + content files named by SHA1(domain-path) hex
- iOS10+ replaced mbdb with Manifest.db (SQLite) — reader must detect and
  redirect ("use M7 SQLite path")
- Round-trip re-serialization gives us a free property test without golden
  fixtures from real devices

## Recovery semantics (uncertain — flagged, not assumed)
- Entering recovery programmatically: believed lockdownd-level request on
  modern iOS; NOT verified against source this pass. Gate behind Batch B step 1
  hardware validation before any code claims it.
- Exiting recovery: irecovery uses raw USB iBoot protocol (setenv auto-boot/
  saveenv/fsboot). Whether usbmuxd relays any control channel to recovery-mode
  devices is UNKNOWN — that is precisely what scripts/recovery-capture.sh
  answers. If it relays nothing: typed KindUnsupported + honest text (A1 holds).
