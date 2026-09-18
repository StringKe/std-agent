package cli

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"
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
	cs := `abcdef1234567890  std-agent_0.2.0_linux_amd64.tar.gz
fedcba0987654321  std-agent_0.2.0_darwin_arm64.tar.gz
1111111111111111  std-agent_0.2.0_windows_amd64.zip
`
	if got := findChecksum(cs, "std-agent_0.2.0_darwin_arm64.tar.gz"); got != "fedcba0987654321" {
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
	// build tar.gz: { stdagent: "fake-binary" }
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

func TestIsBrewCellarPath(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{"/opt/homebrew/Cellar/stdagent/0.0.18/bin/stdagent", true},
		{"/usr/local/Cellar/stdagent/0.0.18/bin/stdagent", true},
		{"/home/linuxbrew/.linuxbrew/Cellar/stdagent/0.0.18/bin/stdagent", true},
		{"/opt/homebrew/bin/stdagent", false},
		{"/Users/x/.local/bin/stdagent", false},
		{"/opt/homebrew/Cellar/tomato-cli/0.1.2/bin/tomato", false},
		{"", false},
	}
	for _, c := range cases {
		if got := isBrewCellarPath(c.path); got != c.want {
			t.Errorf("isBrewCellarPath(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}

func stubUpgradeEnv(t *testing.T, env map[string]string, exe string, exeErr error, linkTarget string, linkErr error) {
	t.Helper()
	oldEnv, oldExe, oldEval := osGetenv, osExecutable, evalSymlinks
	osGetenv = func(k string) string { return env[k] }
	osExecutable = func() (string, error) { return exe, exeErr }
	evalSymlinks = func(_ string) (string, error) { return linkTarget, linkErr }
	t.Cleanup(func() { osGetenv, osExecutable, evalSymlinks = oldEnv, oldExe, oldEval })
}

func TestDetectInstallMethod(t *testing.T) {
	cases := []struct {
		name       string
		env        map[string]string
		exe        string
		exeErr     error
		linkTarget string
		linkErr    error
		want       string
	}{
		{"env override brew", map[string]string{"STDAGENT_INSTALL_METHOD": "brew"}, "/x/stdagent", nil, "/x/stdagent", nil, "brew"},
		{"env override other", map[string]string{"STDAGENT_INSTALL_METHOD": "curl"}, "/opt/homebrew/Cellar/stdagent/0.0.18/bin/stdagent", nil, "", nil, "generic"},
		{"cellar direct", nil, "/opt/homebrew/Cellar/stdagent/0.0.18/bin/stdagent", nil, "/opt/homebrew/Cellar/stdagent/0.0.18/bin/stdagent", nil, "brew"},
		{"symlink into cellar", nil, "/opt/homebrew/bin/stdagent", nil, "/opt/homebrew/Cellar/stdagent/0.0.18/bin/stdagent", nil, "brew"},
		{"curl install", nil, "/Users/x/.local/bin/stdagent", nil, "/Users/x/.local/bin/stdagent", nil, "generic"},
		{"exe error", nil, "", errors.New("no exe"), "", errors.New("no link"), "generic"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			stubUpgradeEnv(t, c.env, c.exe, c.exeErr, c.linkTarget, c.linkErr)
			if got := detectInstallMethod(); got != c.want {
				t.Errorf("detectInstallMethod() = %q, want %q", got, c.want)
			}
		})
	}
}

func TestRunBrewUpgradePinRejected(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))
	err := runBrewUpgrade(cmd, upgradeOptions{Pin: "v0.0.18"})
	if err == nil || !strings.Contains(err.Error(), "pinning") {
		t.Errorf("expected pinning error, got %v", err)
	}
}

func TestRunBrewUpgradeDelegates(t *testing.T) {
	oldFetch, oldExec := fetchLatestTag, execBrewUpgrade
	fetchLatestTag = func() (string, error) { return "v9.9.9", nil }
	called := false
	execBrewUpgrade = func(_ *cobra.Command) error { called = true; return nil }
	t.Cleanup(func() { fetchLatestTag, execBrewUpgrade = oldFetch, oldExec })

	var out bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&out)
	cmd.SetErr(new(bytes.Buffer))
	if err := runBrewUpgrade(cmd, upgradeOptions{}); err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Error("expected brew upgrade delegation")
	}
	if !strings.Contains(out.String(), "delegating") {
		t.Errorf("expected delegating message, got %q", out.String())
	}
}

func TestRunBrewUpgradeUpToDate(t *testing.T) {
	oldFetch, oldExec, oldVer := fetchLatestTag, execBrewUpgrade, versionStr
	fetchLatestTag = func() (string, error) { return "v9.9.9", nil }
	versionStr = "9.9.9"
	called := false
	execBrewUpgrade = func(_ *cobra.Command) error { called = true; return nil }
	t.Cleanup(func() { fetchLatestTag, execBrewUpgrade, versionStr = oldFetch, oldExec, oldVer })

	var out bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&out)
	cmd.SetErr(new(bytes.Buffer))
	if err := runBrewUpgrade(cmd, upgradeOptions{}); err != nil {
		t.Fatal(err)
	}
	if called {
		t.Error("must not delegate when already up-to-date")
	}
	if !strings.Contains(out.String(), "already up-to-date") {
		t.Errorf("expected up-to-date message, got %q", out.String())
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
