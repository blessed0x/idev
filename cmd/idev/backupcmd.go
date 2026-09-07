package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/blessed0x/go-idevice/device"
	"github.com/spf13/cobra"
)

// backupInspect inventories an on-disk pre-iOS10 backup directory. Purely
// offline: no device is touched, so it works for archives shared from
// anywhere (the idevicebackup2-era layout).
func newBackupCmd() *cobra.Command {
	var (
		showJSON bool
		domain   string
		tree     string
	)
	cmd := &cobra.Command{
		Use:   "inspect <backup-dir>",
		Short: "inventory a pre-iOS10 backup directory on this machine",
		Long: `inspect reads a legacy backup's Manifest.mbdb (the layout Apple used
before iOS 10) entirely offline: device identity, encryption state, per-
domain entry counts and sizes.

  --domain HomeDomain        only entries of one domain
  --tree Documents           list paths under a domain-relative prefix
  -a / --app com.x.app       shorthand: domain AppDomain-<bundle>

iOS 10+ backups use a SQLite manifest; inspect says so instead of guessing.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			info, err := device.ReadBackup(args[0])
			if err != nil {
				return err
			}
			if app != "" {
				domain = "AppDomain-" + strings.TrimPrefix(app, "AppDomain-")
			}
			entries := info.Entries
			if domain != "" {
				var kept []device.MBDBRecord
				for _, e := range entries {
					if e.Domain == domain {
						kept = append(kept, e)
					}
				}
				entries = kept
			}
			if tree != "" {
				var kept []device.MBDBRecord
				for _, e := range entries {
					if strings.HasPrefix(e.Path, tree) {
						kept = append(kept, e)
					}
				}
				entries = kept
				sort.Slice(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path })
			}
			type summary struct {
				Dir         string   `json:"dir"`
				DeviceName  string   `json:"device_name"`
				ProductType string   `json:"product_type"`
				IOSVersion  string   `json:"ios_version"`
				State       string   `json:"state"`
				Encrypted   bool     `json:"encrypted"`
				Entries     int      `json:"entries"`
				TotalBytes  uint64   `json:"total_bytes"`
				Domains     []string `json:"domains"`
			}
			var total uint64
			domains := map[string]bool{}
			for _, e := range entries {
				total += e.Size
				domains[e.Domain] = true
			}
			dl := make([]string, 0, len(domains))
			for d := range domains {
				dl = append(dl, d)
			}
			sort.Strings(dl)
			s := summary{
				Dir: info.Dir, DeviceName: info.DeviceName, ProductType: info.ProductType,
				IOSVersion: info.IOSVersion, State: info.State, Encrypted: info.Encrypted,
				Entries: len(entries), TotalBytes: total, Domains: dl,
			}
			return printJSONOr(cmd, showJSON, s, func() error {
				fmt.Fprintf(cmd.OutOrStdout(), "backup: %s\n", info.DeviceName)
				fmt.Fprintf(cmd.OutOrStdout(), "  device: %s (iOS %s)\n", info.ProductType, info.IOSVersion)
				fmt.Fprintf(cmd.OutOrStdout(), "  state: %s, encrypted: %v\n", orDash(info.State), info.Encrypted)
				fmt.Fprintf(cmd.OutOrStdout(), "  entries: %d (%d bytes shown)\n", s.Entries, total)
				if len(dl) > 0 {
					fmt.Fprintf(cmd.OutOrStdout(), "  domains:\n")
					for _, d := range dl {
						fmt.Fprintf(cmd.OutOrStdout(), "    %s\n", d)
					}
				}
				for _, e := range entries {
					kind := "f"
					if e.IsDir() {
						kind = "d"
					}
					fmt.Fprintf(cmd.OutOrStdout(), "    [%s] %s:%s (%d bytes)\n", kind, e.Domain, e.Path, e.Size)
				}
				return nil
			})
		},
	}
	cmd.Flags().BoolVar(&showJSON, "json", false, "machine-readable output")
	cmd.Flags().StringVar(&domain, "domain", "", "only entries of this domain")
	cmd.Flags().StringVar(&tree, "tree", "", "only entries whose path starts with this prefix")
	cmd.Flags().StringVarP(&app, "app", "a", "", "shorthand domain filter: AppDomain-<bundle>")
	return cmd
}

var app string

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
