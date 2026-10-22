package utils

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestCreateTarball(t *testing.T) {
	source := t.TempDir()
	if err := os.WriteFile(filepath.Join(source, "a.txt"), []byte("hello"), 0640); err != nil {
		t.Fatalf("failed to prepare source: %v", err)
	}
	if err := os.Mkdir(filepath.Join(source, "sub"), 0755); err != nil {
		t.Fatalf("failed to prepare source: %v", err)
	}
	if err := os.WriteFile(filepath.Join(source, "sub", "b.txt"), []byte("world"), 0600); err != nil {
		t.Fatalf("failed to prepare source: %v", err)
	}
	if err := os.Symlink("a.txt", filepath.Join(source, "link")); err != nil {
		t.Fatalf("failed to prepare source: %v", err)
	}

	target := filepath.Join(t.TempDir(), "archive.tar.gz")
	if err := createTarball(source, target); err != nil {
		t.Fatalf("createTarball returned an error: %v", err)
	}

	headers, contents := readTarball(t, target)

	if got := contents["a.txt"]; got != "hello" {
		t.Errorf("a.txt = %q, want %q", got, "hello")
	}
	if got := contents[filepath.ToSlash("sub/b.txt")]; got != "world" {
		t.Errorf("sub/b.txt = %q, want %q", got, "world")
	}

	sub, ok := headers["sub"]
	if !ok {
		t.Fatalf("directory entry 'sub' missing from archive, got %v", keys(headers))
	}
	if sub.Typeflag != tar.TypeDir {
		t.Errorf("'sub' typeflag = %d, want %d", sub.Typeflag, tar.TypeDir)
	}

	link, ok := headers["link"]
	if !ok {
		t.Fatalf("symlink entry 'link' missing from archive, got %v", keys(headers))
	}
	if link.Typeflag != tar.TypeSymlink {
		t.Errorf("'link' typeflag = %d, want %d", link.Typeflag, tar.TypeSymlink)
	}
	if link.Linkname != "a.txt" {
		t.Errorf("'link' linkname = %q, want %q", link.Linkname, "a.txt")
	}

	info, err := os.Stat(target)
	if err != nil {
		t.Fatalf("failed to stat archive: %v", err)
	}
	if info.Size() == 0 {
		t.Error("archive is empty")
	}
}

func TestCreateTarballMissingSource(t *testing.T) {
	target := filepath.Join(t.TempDir(), "archive.tar.gz")
	if err := createTarball(filepath.Join(t.TempDir(), "missing"), target); err == nil {
		t.Fatal("expected an error for a missing source directory")
	}
}

func readTarball(t *testing.T, path string) (map[string]*tar.Header, map[string]string) {
	t.Helper()

	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("failed to open archive: %v", err)
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		t.Fatalf("failed to open gzip stream: %v", err)
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)
	headers := make(map[string]*tar.Header)
	contents := make(map[string]string)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("failed to read archive entry: %v", err)
		}
		headers[header.Name] = header

		if header.Typeflag == tar.TypeReg {
			data, err := io.ReadAll(tarReader)
			if err != nil {
				t.Fatalf("failed to read %s: %v", header.Name, err)
			}
			contents[header.Name] = string(data)
		}
	}
	return headers, contents
}

func keys(headers map[string]*tar.Header) []string {
	result := make([]string, 0, len(headers))
	for name := range headers {
		result = append(result, name)
	}
	return result
}
