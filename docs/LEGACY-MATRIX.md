# Legacy support matrix (iOS 5 → 16)

Filled from `scripts/legacy-sweep.sh` runs on real hardware. A row without a
capture reference is UNVERIFIED — code may work, we do not claim it.

Legend: ✅ verified working · ⚠️ works with notes (link capture) · ❌ expected
limitation · ⬜ unverified

| verb | iOS 5–6 | 7–9 | 10–12 | 13–15 | 16 | evidence |
|---|---|---|---|---|---|---|
| list/info/battery | ⬜ | ⬜ | ⬜ | ⬜ | ✅ community | |
| pair | ⬜ (plaintext era) | ⬜ | ✅ | ✅ | ✅ | |
| apps/install/uninstall | ⬜ | ⬜ | ✅ | ✅ | ✅ | |
| launch/kill | n/a (DDI) | n/a (DDI) | ⬜ DDI | ⬜ DDI | ⬜ DDI | needs `ddi mount` |
| syslog | ⬜ classic | ⬜ classic | ✅ classic | ✅ classic | ✅ classic | |
| screenshot | ⬜ | ⬜ | ✅ | ✅ | ✅ | instruments service |
| forward/pasteboard/watch | ⬜ | ⬜ | ✅ | ✅ | ✅ | usbmuxd level, version-stable |
| files (AFC) | ⬜ (paths differ) | ⬜ | ✅ | ✅ | ✅ | |
| crash list/pull | ⬜ | ⬜ | ✅ | ✅ | ✅ | |
| ddi mount/status | ⬜ | ⬜ | ✅ target | ✅ target | ✅ target | Batch C live test |
| omega | — | — | ⚠️ 16 only | ⚠️ 16+ | ✅ 16–18/26 | sourced verdict |

## Known quirks (each lands with a fixture test when confirmed)

- pre-iOS7 pairing completes without SSL session upgrade (L3)
- pre-iOS10 backups are MBDB (`idev backup inspect`); iOS10+ are SQLite
- house_arrest sandbox paths differ pre-iOS7
- recovery/DFU devices appear in usbmuxd with chip-dossier serials (L1)
