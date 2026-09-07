package device

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/imagemounter"
	"github.com/blessed0x/idev/internal/log"
)

// L2: Developer Disk Image mounting for iOS <=16 devices. On 17+ the same
// services live behind the CoreDevice tunnel instead (see tunnel.go), so the
// verbs gate hard on version and point at the right mechanism either way.
//
// Mounting needs a DDI image matching the device's exact build;
// DownloadImageFor fetches it from go-ios's artifact mirror when --download
// is passed, otherwise --image names a local .dmg/.ddi.

func ddiSupported(version string) bool {
	v, err := ParseVersion(version)
	if err != nil {
		return false // unknown -> not supported; tunnels are the modern path
	}
	return v[0] <= 16
}

func (g *GoIOS) requireDDICapable(ctx context.Context, udid string) (ios.DeviceEntry, error) {
	dev, err := g.entry(ctx, udid)
	if err != nil {
		return dev, err
	}
	info, err := g.Info(ctx, udid)
	if err != nil {
		return dev, err
	}
	if !ddiSupported(info.ProductVersion) {
		return dev, &Error{
			Kind: KindUnsupported,
			Op:   "mount developer disk image",
			Remediation: fmt.Sprintf("iOS %s serves these services over a developer\n"+
				"tunnel instead (idev starts one automatically); DDI mounting covers\n"+
				"iOS 16 and older.", info.ProductVersion),
		}
	}
	if enabled, derr := imagemounter.IsDevModeEnabled(dev); derr == nil && !enabled {
		return dev, &Error{
			Kind: KindPermission,
			Op:   "mount developer disk image",
			Remediation: "Developer Mode is OFF: Settings > Privacy & Security >\n" +
				"Developer Mode > turn on (see `idev devmode`).",
		}
	}
	return dev, nil
}

// DDIMount mounts a local Developer Disk Image. With download=true and no
// imagePath, go-ios fetches the matching image first (network).
func (g *GoIOS) DDIMount(ctx context.Context, udid, imagePath string, download bool) error {
	dev, err := g.requireDDICapable(ctx, udid)
	if err != nil {
		return err
	}
	err = g.run(ctx, func() error {
		if imagePath == "" {
			if !download {
				return fmt.Errorf("give --image <path.dmg> or --download to fetch the matching image")
			}
			log.Infof("downloading the matching Developer Disk Image (network)..")
			p, err := imagemounter.DownloadImageFor(dev, os.TempDir())
			if err != nil {
				return fmt.Errorf("downloading: %w", err)
			}
			imagePath = p
			log.Infof("image ready: %s", p)
		}
		return imagemounter.MountImage(dev, imagePath)
	})
	if err != nil {
		return g.wrap(KindConnection, "mount developer disk image",
			"keep the device unlocked during the mount", err)
	}
	log.Infof("developer disk image mounted")
	return nil
}

// DDIStatus reports whether an image is currently mounted.
func (g *GoIOS) DDIStatus(ctx context.Context, udid string) ([]string, error) {
	dev, err := g.entry(ctx, udid)
	if err != nil {
		return nil, err
	}
	var mounted [][]byte
	err = g.run(ctx, func() error {
		im, ierr := imagemounter.NewImageMounter(dev)
		if ierr != nil {
			return ierr
		}
		defer im.Close()
		mounted, ierr = im.ListImages()
		return ierr
	})
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "invalidservice") || tunnelGate(err) {
			return nil, g.wrap(KindNotFound, "ddi status",
				"this service answers only on iOS <=16 with Developer Mode ON", err)
		}
		return nil, g.wrap(KindConnection, "ddi status", "", err)
	}
	signatures := make([]string, 0, len(mounted))
	for _, sig := range mounted {
		signatures = append(signatures, fmt.Sprintf("%x", sig)[:16]+"…")
	}
	return signatures, nil
}

// DDIUnmount detaches the current image.
func (g *GoIOS) DDIUnmount(ctx context.Context, udid string) error {
	dev, err := g.entry(ctx, udid)
	if err != nil {
		return err
	}
	err = g.run(ctx, func() error { return imagemounter.UnmountImage(dev) })
	if err != nil {
		return g.wrap(KindConnection, "unmount developer disk image", "", err)
	}
	log.Infof("developer disk image unmounted")
	return nil
}
