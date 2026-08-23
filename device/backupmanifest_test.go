package device

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// mbdbFixture builds a manifest with one of each interesting entry shape.
func mbdbFixture() []MBDBRecord {
	return []MBDBRecord{
		{Domain: "HomeDomain", Path: "/Library/Preferences", Inode: 2,
			Mode: 0o40755, UID: 501, GID: 501, Mtime: 1700000000, Atime: 1700000001,
			Ctime: 1700000002, Size: 170},
		{Domain: "HomeDomain", Path: "/Library/prefs.plist", DataHash: "\xab\xcd\xef",
			Inode: 3, Mode: 0o100644, UID: 501, GID: 501,
			Mtime: 1700000010, Atime: 1700000011, Ctime: 1700000012, Size: 4096,
			Props: [][2]string{{"com.apple.backup.RecordType", "file"}}},
		{Domain: "AppDomain-com.example.app", Path: "Documents/notes.txt",
			LinkTarget: "", Inode: 9, Mode: 0o100600, UID: 33,
			GID: 33, Mtime: 1699999999, Atime: 1699999999, Ctime: 1699999999,
			Size: 42},
	}
}

func TestMBDBRoundTripByteIdentical(t *testing.T) {
	fixture := SerializeMBDB(mbdbFixture())
	parsed, err := ParseMBDB(fixture)
	if err != nil {
		t.Fatalf("parse own fixture: %v", err)
	}
	re := SerializeMBDB(parsed)
	if !bytes.Equal(fixture, re) {
		t.Fatal("round-trip must be byte-identical (parser and serializer disagree)")
	}
}

func TestParseMBDBRejectsGarbage(t *testing.T) {
	for _, data := range [][]byte{
		nil,
		[]byte("not mbdb at all"),
		append(append([]byte{}, mbdbMagic...), 0x00), // magic then truncated length
	} {
		if _, err := ParseMBDB(data); err == nil {
			t.Fatalf("garbage %q must be rejected", data)
		}
	}
}

func TestMBDBRecordAccessors(t *testing.T) {
	recs := mbdbFixture()
	if !recs[0].IsDir() {
		t.Fatalf("040755 must classify as directory, got %o", recs[0].Mode)
	}
	if recs[1].IsDir() || recs[1].IsDir() && false {
		_ = recs[1]
	}
	if recs[1].IsDir() {
		t.Fatal("100644 must not be a directory")
	}
	if got := recs[1].DataHashHex(); got != "abcdef" {
		t.Fatalf("hash hex mismatch: %q", got)
	}
}

// FuzzMBDBNoPanic smashes the parser: truncations and hostile lengths must
// produce clean errors naming the offset — never panics, never partial use.
func FuzzMBDBNoPanic(f *testing.F) {
	f.Add(SerializeMBDB(mbdbFixture()))
	f.Add(mbdbMagic)
	good := SerializeMBDB(mbdbFixture())
	for _, n := range []int{8, 20, 60, len(good) / 2, len(good) - 1} {
		if n < len(good) {
			f.Add(good[:n])
		}
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = ParseMBDB(data) // clean error or success; panic is the failure
	})
}

func TestReadBackupInventory(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "Manifest.mbdb"), SerializeMBDB(mbdbFixture()), 0o644)
	os.WriteFile(filepath.Join(dir, "Status.plist"),
		[]byte(`<?xml version="1.0"?><plist version="1.0"><dict>
<key>BackupState</key><string>finished</string>
<key>IsEncrypted</key><true/></dict></plist>`), 0o644)
	os.WriteFile(filepath.Join(dir, "Info.plist"),
		[]byte(`<?xml version="1.0"?><plist version="1.0"><dict>
<key>Device Name</key><string>Old Phone</string>
<key>Product Type</key><string>iPhone7,2</string>
<key>Product Version</key><string>8.4</string></dict></plist>`), 0o644)

	info, err := ReadBackup(dir)
	if err != nil {
		t.Fatal(err)
	}
	if info.DeviceName != "Old Phone" || info.ProductType != "iPhone7,2" || info.IOSVersion != "8.4" {
		t.Fatalf("Info.plist fields lost: %+v", info)
	}
	if !info.Encrypted || info.State != "finished" {
		t.Fatalf("Status.plist fields lost: %+v", info)
	}
	if len(info.Entries) != 3 || info.Domains()["AppDomain-com.example.app"] != 1 {
		t.Fatalf("entries miscounted: %+v", info.Domains())
	}
}

func TestReadBackupDetectsSQLiteManifest(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "Manifest.db"), []byte("sqlite"), 0o644)
	if _, err := ReadBackup(dir); err == nil || !bytes.Contains([]byte(err.Error()), []byte("iOS10+")) {
		t.Fatalf("modern backup must redirect with a clear message, got %v", err)
	}
}

func TestReadBackupMissingManifest(t *testing.T) {
	if _, err := ReadBackup(t.TempDir()); err == nil {
		t.Fatal("empty dir must error")
	}
}
