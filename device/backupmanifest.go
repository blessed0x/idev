package device

// Pre-iOS10 backup manifests: Manifest.mbdb.
//
// The format is a flat record stream behind a six-byte magic
// ("mbdb" 0x05 0x00). Every record describes one filesystem entry of the
// backup: five length-prefixed metadata strings, fixed-width stat fields,
// then inline properties. File CONTENTS live outside the manifest, named by
// SHA1(domain-path); this parser reads the manifest itself, which is what
// `idev backup inspect` inventories.
//
// iOS10 and later replaced this format with a SQLite Manifest.db; ReadBackup
// detects that shape and points at the modern path instead of misparsing.

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	howett "howett.net/plist"
)

var mbdbMagic = []byte{'m', 'b', 'd', 'b', 0x05, 0x00}

// MBDBRecord is one entry of a pre-iOS10 backup manifest.
type MBDBRecord struct {
	Domain        string
	Path          string
	LinkTarget    string
	DataHash      string // raw bytes; use DataHashHex()
	EncryptionKey string

	// Reserved is the undocumented byte between the strings block and the
	// stat fields. Real backups carry 0 here; kept opaque on purpose.
	Reserved uint8
	Inode    uint64
	Mode     uint32
	UID      uint32
	GID      uint32
	Mtime    uint32
	Atime    uint32
	Ctime    uint32
	Size     uint64

	Props [][2]string
}

// DataHashHex renders the content hash the way backup tools print it.
func (r MBDBRecord) DataHashHex() string {
	return hex.EncodeToString([]byte(r.DataHash))
}

// IsDir reports whether the entry is a directory (unix S_IFDIR bit).
func (r MBDBRecord) IsDir() bool { return r.Mode&0o170000 == 0o040000 }

// ModTime converts the manifest's mtime to wall-clock time.
func (r MBDBRecord) ModTime() time.Time {
	return time.Unix(int64(r.Mtime), 0)
}

// ParseMBDB parses a Manifest.mbdb image from memory. Truncated input yields
// an error naming the record offset — never a partial slice silently used.
func ParseMBDB(data []byte) ([]MBDBRecord, error) {
	if !bytes.HasPrefix(data, mbdbMagic) {
		return nil, fmt.Errorf("not an mbdb manifest (bad magic)")
	}
	var out []MBDBRecord
	pos := len(mbdbMagic)
	for pos < len(data) {
		rec, next, err := parseMBDBRecord(data, pos)
		if err != nil {
			return nil, fmt.Errorf("record at offset %d: %w", pos, err)
		}
		out = append(out, rec)
		pos = next
	}
	return out, nil
}

// lpString reads one 2-byte big-endian length-prefixed string.
func lpString(data []byte, pos int) (string, int, error) {
	if pos+2 > len(data) {
		return "", 0, fmt.Errorf("truncated string length")
	}
	n := int(binary.BigEndian.Uint16(data[pos : pos+2]))
	pos += 2
	if pos+n > len(data) {
		return "", 0, fmt.Errorf("truncated string (%d bytes claimed)", n)
	}
	return string(data[pos : pos+n]), pos + n, nil
}

func parseMBDBRecord(data []byte, pos int) (MBDBRecord, int, error) {
	var rec MBDBRecord
	var err error
	if rec.Domain, pos, err = lpString(data, pos); err != nil {
		return rec, pos, err
	}
	if rec.Path, pos, err = lpString(data, pos); err != nil {
		return rec, pos, err
	}
	if rec.LinkTarget, pos, err = lpString(data, pos); err != nil {
		return rec, pos, err
	}
	if rec.DataHash, pos, err = lpString(data, pos); err != nil {
		return rec, pos, err
	}
	if rec.EncryptionKey, pos, err = lpString(data, pos); err != nil {
		return rec, pos, err
	}
	fixed := 1 + 8 + 4 + 4 + 4 + 4 + 4 + 4 + 8 + 1 // modeHigh,inode,mode,uid,gid,mtime,atime,ctime,size,propCount
	if pos+fixed > len(data) {
		return rec, pos, fmt.Errorf("truncated stat fields")
	}
	rec.Reserved = data[pos]
	pos++
	rec.Inode = binary.BigEndian.Uint64(data[pos:])
	pos += 8
	rec.Mode = binary.BigEndian.Uint32(data[pos:])
	pos += 4
	rec.UID = binary.BigEndian.Uint32(data[pos:])
	pos += 4
	rec.GID = binary.BigEndian.Uint32(data[pos:])
	pos += 4
	rec.Mtime = binary.BigEndian.Uint32(data[pos:])
	pos += 4
	rec.Atime = binary.BigEndian.Uint32(data[pos:])
	pos += 4
	rec.Ctime = binary.BigEndian.Uint32(data[pos:])
	pos += 4
	rec.Size = binary.BigEndian.Uint64(data[pos:])
	pos += 8
	propCount := int(data[pos])
	pos++
	for i := 0; i < propCount; i++ {
		var k, v string
		if k, pos, err = lpString(data, pos); err != nil {
			return rec, pos, err
		}
		if v, pos, err = lpString(data, pos); err != nil {
			return rec, pos, err
		}
		rec.Props = append(rec.Props, [2]string{k, v})
	}
	return rec, pos, nil
}

// SerializeMBDB re-encodes records to the exact wire format. Round-tripping
// ParseMBDB -> SerializeMBDB is byte-identical, which is how the parser is
// tested without golden files from real devices.
func SerializeMBDB(records []MBDBRecord) []byte {
	var buf bytes.Buffer
	buf.Write(mbdbMagic)
	for _, rec := range records {
		writeLP(&buf, rec.Domain)
		writeLP(&buf, rec.Path)
		writeLP(&buf, rec.LinkTarget)
		writeLP(&buf, rec.DataHash)
		writeLP(&buf, rec.EncryptionKey)
		buf.WriteByte(rec.Reserved)
		var u64 [8]byte
		binary.BigEndian.PutUint64(u64[:], rec.Inode)
		buf.Write(u64[:])
		var u32 [4]byte
		for _, v := range []uint32{rec.Mode, rec.UID, rec.GID, rec.Mtime, rec.Atime, rec.Ctime} {
			binary.BigEndian.PutUint32(u32[:], v)
			buf.Write(u32[:])
		}
		binary.BigEndian.PutUint64(u64[:], rec.Size)
		buf.Write(u64[:])
		buf.WriteByte(uint8(len(rec.Props)))
		for _, kv := range rec.Props {
			writeLP(&buf, kv[0])
			writeLP(&buf, kv[1])
		}
	}
	return buf.Bytes()
}

func writeLP(buf *bytes.Buffer, s string) {
	var l [2]byte
	binary.BigEndian.PutUint16(l[:], uint16(len(s)))
	buf.Write(l[:])
	buf.WriteString(s)
}

// BackupInfo is the inventory `idev backup inspect` reports.
type BackupInfo struct {
	Dir         string
	DeviceName  string
	ProductType string
	IOSVersion  string
	LastBackup  time.Time
	Encrypted   bool
	State       string // Status.plist BackupState (e.g. "finished")
	Entries     []MBDBRecord
}

// Domains tallies entries per domain.
func (b *BackupInfo) Domains() map[string]int {
	out := map[string]int{}
	for _, e := range b.Entries {
		out[e.Domain]++
	}
	return out
}

// ReadBackup inventories an on-disk pre-iOS10 backup directory.
func ReadBackup(dir string) (*BackupInfo, error) {
	if _, err := os.Stat(filepath.Join(dir, "Manifest.db")); err == nil {
		return nil, fmt.Errorf("%s is an iOS10+ backup (SQLite Manifest.db) — the mbdb reader covers pre-iOS10 only", dir)
	}
	data, err := os.ReadFile(filepath.Join(dir, "Manifest.mbdb"))
	if err != nil {
		return nil, fmt.Errorf("no Manifest.mbdb in %s: %w", dir, err)
	}
	entries, err := ParseMBDB(data)
	if err != nil {
		return nil, err
	}
	info := &BackupInfo{Dir: dir, Entries: entries}
	if st, err := readPlistFile(filepath.Join(dir, "Status.plist")); err == nil {
		if v, ok := st["BackupState"].(string); ok {
			info.State = v
		}
		if v, ok := st["IsEncrypted"].(bool); ok && v {
			info.Encrypted = true
		}
	}
	if inf, err := readPlistFile(filepath.Join(dir, "Info.plist")); err == nil {
		if v, ok := inf["Device Name"].(string); ok {
			info.DeviceName = v
		}
		if v, ok := inf["Product Type"].(string); ok {
			info.ProductType = v
		}
		if v, ok := inf["Product Version"].(string); ok {
			info.IOSVersion = v
		}
	}
	return info, nil
}

// readPlistFile is a tiny helper so both plists degrade gracefully: a missing
// or malformed sidecar never fails the whole inventory.
func readPlistFile(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var d map[string]interface{}
	if _, err := howett.Unmarshal(data, &d); err != nil {
		return nil, err
	}
	return d, nil
}
