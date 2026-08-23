package device

import (
	"context"
	"fmt"
	"strings"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/gwnodex-bit/idev/internal/log"
)

// L1: legacy boot-stage visibility.
//
// usbmuxd reports devices at every boot stage, not just normal mode. A
// device sitting in Recovery mode (or DFU) still appears in the device list,
// but its serial-style identifier looks different: instead of the 40-char
// hex UDID, Apple's muxd surfaces a chip-dossier string like
// "CPID:8960 CPRV:11 BDID:98 ECID:001F2B..." — that shape is the one stable,
// observable discriminator available without raw USB access.
//
// What each stage can do through this tool:
//
//   normal     everything
//   recovery   visibility only; exiting needs the iBoot protocol over raw
//              USB, which a no-cgo build does not ship (typed refusal)
//   dfu        visibility only; same raw-USB boundary

const (
	StateNormal   = "normal"
	StateRecovery = "recovery"
	StateDFU      = "dfu"
)

// ClassifyState derives the boot-stage label from what usbmuxd told us.
// Pure function so fixtures from real hardware can pin it later
// (.ultraplan plan Batch B step 1).
func ClassifyState(properties ios.DeviceProperties) string {
	sn := properties.SerialNumber
	switch {
	case strings.Contains(sn, "CPID:") || strings.Contains(sn, "PWNED:"):
		return StateRecovery // chip-dossier shape: recovery (or pwned DFU)
	case strings.HasPrefix(sn, "DFU"):
		return StateDFU
	default:
		return StateNormal
	}
}

// RecoveryEnter reboots the device into Recovery mode. The trigger is
// lockdownd's recovery_mode_handler service — starting it IS the request
// (the same contract libimobiledevice's ideviceenterrecovery uses).
func (g *GoIOS) RecoveryEnter(ctx context.Context, udid string) error {
	dev, err := g.entry(ctx, udid)
	if err != nil {
		return err
	}
	err = g.run(ctx, func() error {
		conn, err := ios.ConnectToService(dev, "com.apple.mobile.recovery_mode_handler")
		if err != nil {
			return fmt.Errorf("opening recovery_mode_handler: %w", err)
		}
		_ = conn.Close()
		return nil
	})
	if err != nil {
		return g.wrap(KindConnection, "enter recovery",
			"the device must be in normal mode and trusted", err)
	}
	log.Infof("device %s is rebooting into Recovery mode", udid)
	return nil
}

// RecoveryExit cannot be done honestly without raw-USB iBoot access
// (setenv auto-boot / saveenv / fsboot). Typed refusal with real guidance,
// never a silent lie.
func (g *GoIOS) RecoveryExit(ctx context.Context, udid string) error {
	return &Error{
		Kind: KindUnsupported,
		Op:   "exit recovery",
		Remediation: "exiting Recovery mode needs the iBoot protocol over raw USB,\n" +
			"which idev does not ship (it would require cgo). Use any of:\n" +
			"  irecovery -q && irecovery -c \"setenv auto-boot true\" && irecovery -c saveenv && irecovery -c fsboot\n" +
			"  Apple Configurator (right-click the device > exit recovery)\n" +
			"  pymobiledevice3 recovery exit",
	}
}
