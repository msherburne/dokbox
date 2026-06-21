package docker

import (
	"archive/tar"
	"bytes"
	"testing"
)

func TestListArchiveDirectoryEntriesReturnsImmediateChildren(t *testing.T) {
	archive := newTestArchive(t, []tar.Header{
		{Name: "./", Mode: 0o755, Typeflag: tar.TypeDir},
		{Name: "./etc/", Mode: 0o755, Typeflag: tar.TypeDir},
		{Name: "./etc/hosts", Mode: 0o644, Size: int64(len("hosts"))},
		{Name: "./var/log/", Mode: 0o755, Typeflag: tar.TypeDir},
		{Name: "./README.md", Mode: 0o644, Size: int64(len("readme"))},
	})

	entries, err := listArchiveDirectoryEntries(archive, "/")
	if err != nil {
		t.Fatalf("expected archive listing to succeed, got %v", err)
	}

	if len(entries) != 3 {
		t.Fatalf("expected three top-level entries, got %#v", entries)
	}

	if entries[0].Name != "etc" || !entries[0].IsDir || entries[0].Path != "/etc" {
		t.Fatalf("unexpected first directory entry: %#v", entries[0])
	}
	if entries[1].Name != "var" || !entries[1].IsDir || entries[1].Path != "/var" {
		t.Fatalf("unexpected second directory entry: %#v", entries[1])
	}
	if entries[2].Name != "README.md" || entries[2].IsDir || entries[2].Path != "/README.md" {
		t.Fatalf("unexpected file entry: %#v", entries[2])
	}
}

func TestListArchiveDirectoryEntriesPreservesNestedDirectoryPaths(t *testing.T) {
	archive := newTestArchive(t, []tar.Header{
		{Name: "./", Mode: 0o755, Typeflag: tar.TypeDir},
		{Name: "./config.yaml", Mode: 0o644, Size: int64(len("cfg"))},
		{Name: "./subdir/", Mode: 0o755, Typeflag: tar.TypeDir},
	})

	entries, err := listArchiveDirectoryEntries(archive, "/etc")
	if err != nil {
		t.Fatalf("expected archive listing to succeed, got %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("expected two nested entries, got %#v", entries)
	}

	if entries[0].Path != "/etc/subdir" || !entries[0].IsDir {
		t.Fatalf("expected nested directory path, got %#v", entries[0])
	}
	if entries[1].Path != "/etc/config.yaml" || entries[1].IsDir {
		t.Fatalf("expected nested file path, got %#v", entries[1])
	}
}

func newTestArchive(t *testing.T, headers []tar.Header) *bytes.Buffer {
	t.Helper()

	var buffer bytes.Buffer
	writer := tar.NewWriter(&buffer)
	for _, header := range headers {
		header := header
		if err := writer.WriteHeader(&header); err != nil {
			t.Fatalf("expected header write to succeed, got %v", err)
		}
		if header.Typeflag != tar.TypeDir && header.Size > 0 {
			payload := bytes.Repeat([]byte("x"), int(header.Size))
			if _, err := writer.Write(payload); err != nil {
				t.Fatalf("expected payload write to succeed, got %v", err)
			}
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("expected archive close to succeed, got %v", err)
	}

	return &buffer
}
