package device

import "testing"

func TestKnownLegacyQuirksByEra(t *testing.T) {
	cases := map[string]int{
		"5.1.1": 4, // pre-7 pair + pre-10 mbdb + old-lockdown notes
		"8.4":   2,
		"12.4":  0,
		"16.7":  0,
		"26.1":  1, // modern syslog note only
	}
	for v, want := range cases {
		if got := len(KnownLegacyQuirks(v)); got != want {
			t.Errorf("KnownLegacyQuirks(%q) = %d notes, want %d (%v)", v, got, want, KnownLegacyQuirks(v))
		}
	}
	if KnownLegacyQuirks("garbage") != nil {
		t.Fatal("unparseable versions carry no claims")
	}
}

func TestIsHexUDIDShapes(t *testing.T) {
	good := []string{
		"1234567890abcdef1234567890abcdef12345678", // 40-hex classic
		"1234567890abcdef12345678",                 // 24-hex old
		"1234567890abcdef1234567890abcdef",         // 32-hex A12+
		"ABCDEF0123456789ABCDEF0123456789ABCDEF01", // uppercase tolerated
	}
	bad := []string{"CPID:8960 CPRV:11", "", "zzzz", "0123"}
	for _, s := range good {
		if !IsHexUDID(s) {
			t.Errorf("%q should be hex UDID", s)
		}
	}
	for _, s := range bad {
		if IsHexUDID(s) {
			t.Errorf("%q should NOT be hex UDID", s)
		}
	}
}
