package device

// L3: legacy lockdown tolerance. iOS's lockdown surface shifted over the
// years; these helpers let idev answer honestly about what a given version
// can do instead of failing mysteriously mid-verb. Every claim here is a
// documented ecosystem fact (libimobiledevice tool matrix), and each row
// gets a fixture test when hardware confirms it (LEGACY-MATRIX.md).

import "strings"

// KnownLegacyQuirks returns advisory notes for a ProductVersion. Empty when
// nothing special applies. Ordered oldest-behavior first.
func KnownLegacyQuirks(productVersion string) []string {
	v, err := ParseVersion(productVersion)
	if err != nil {
		return nil
	}
	maj := v[0]
	var out []string
	if maj <= 9 {
		out = append(out,
			"backups use the pre-iOS10 MBDB manifest (idev backup inspect reads them)",
			"pairing on this era completes without an SSL session upgrade",
		)
	}
	if maj <= 6 {
		out = append(out,
			"lockdown predates several modern keys; info shows fewer fields",
			"app sandboxes via house_arrest exist since iOS 4 but paths differ",
		)
	}
	if maj >= 17 {
		out = append(out,
			"classic syslog relay removed by Apple; os_trace over tunnel is the path",
		)
	}
	return out
}

// IsHexUDID reports whether a serial-shaped string is a normal-mode UDID
// (24/25/32/40 hex chars depending on chip generation). Recovery/DFU devices
// surface chip-dossier strings instead.
func IsHexUDID(s string) bool {
	if len(s) != 24 && len(s) != 25 && len(s) != 32 && len(s) != 40 {
		return false
	}
	for _, c := range strings.ToLower(s) {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}
