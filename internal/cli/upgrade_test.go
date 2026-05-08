package cli

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"strings"
	"testing"
)

func TestVersionMatchesTag(t *testing.T) {
	cases := []struct {
		v, t string
		want bool
	}{
		{"0.2.0", "v0.2.0", true},
		{"v0.2.0", "v0.2.0", true},
		{"0.2.0", "0.2.0", true},
		{"0.2.0", "v0.2.1", false},
		{"dev", "v0.2.0", false},
		{"", "v0.2.0", false},
	}
	for _, c := range cases {
		if got := versionMatchesTag(c.v, c.t); got != c.want {
			t.Errorf("versionMatchesTag(%q, %q) = %v, want %v", c.v, c.t, got, c.want)
		}
	}
}

func TestFindChecksum(t *testing.T) {
	cs := `abcdef1234567890  std-ai_0.2.0_linux_amd64.tar.gz
fedcba0987654321  std-ai_0.2.0_darwin_arm64.tar.gz
1111111111111111  std-ai_0.2.0_windows_amd64.zip
`
	if got := findChecksum(cs, "std-ai_0.2.0_darwin_arm64.tar.gz"); got != "fedcba0987654321" {
		t.Errorf("got %q", got)
	}
	if got := findChecksum(cs, "missing.tar.gz"); got != "" {
		t.Errorf("expect empty for missing, got %q", got)
	}
}

func TestSHA256Hex(t *testing.T) {
	got := sha256hex([]byte("hello"))
	want := "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestExtractBinaryTarGz(t *testing.T) {
	// 构造 tar.gz: { stdagent: "fake-binary" }
	var gzBuf bytes.Buffer
	gw := gzip.NewWriter(&gzBuf)
	tw := tar.NewWriter(gw)
	body := []byte("fake-binary-content")
	_ = tw.WriteHeader(&tar.Header{Name: "stdagent", Mode: 0o755, Size: int64(len(body))})
	_, _ = tw.Write(body)
	_ = tw.Close()
	_ = gw.Close()

	got, err := extractBinary(gzBuf.Bytes(), "tar.gz")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, body) {
		t.Errorf("got %q want %q", got, body)
	}
}

func TestExtractBinaryTarGzNotFound(t *testing.T) {
	var gzBuf bytes.Buffer
	gw := gzip.NewWriter(&gzBuf)
	tw := tar.NewWriter(gw)
	body := []byte("other")
	_ = tw.WriteHeader(&tar.Header{Name: "other-binary", Mode: 0o755, Size: int64(len(body))})
	_, _ = tw.Write(body)
	_ = tw.Close()
	_ = gw.Close()

	_, err := extractBinary(gzBuf.Bytes(), "tar.gz")
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected not-found error, got %v", err)
	}
}

func TestExtractBinaryZipExe(t *testing.T) {
	var zipBuf bytes.Buffer
	zw := zip.NewWriter(&zipBuf)
	w, _ := zw.Create("stdagent.exe")
	_, _ = w.Write([]byte("windows-binary"))
	_ = zw.Close()

	got, err := extractBinary(zipBuf.Bytes(), "zip")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "windows-binary" {
		t.Errorf("got %q", got)
	}
}
