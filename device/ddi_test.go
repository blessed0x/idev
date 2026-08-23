package device

import "testing"

func TestDDISupportedVersionGate(t *testing.T) {
	cases := map[string]bool{
		"16.7.1": true, "12.4": true, "8.4": true,
		"17.0": false, "26.1": false, "garbage": false,
	}
	for v, want := range cases {
		if got := ddiSupported(v); got != want {
			t.Errorf("ddiSupported(%q)=%v want %v", v, got, want)
		}
	}
}
