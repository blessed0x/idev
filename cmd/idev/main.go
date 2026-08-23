// idev — control a connected iPhone/iPad from the terminal.
//
// The pure-Go take on pymobiledevice3's everyday surface: pairing,
// installs, process control, syslog (os_trace over an auto-managed
// developer tunnel on iOS 17+), screenshots, port forwarding and the
// clipboard — one static binary, typed errors with real remediation text,
// no pip and no virtualenv. Built on danielpaulus/go-ios.
package main

import (
	"os"

	"github.com/xscope0/idev/device"
	"github.com/xscope0/idev/internal/log"
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		log.Errorf("%v", err)
		os.Exit(device.ExitCode(err))
	}
}
