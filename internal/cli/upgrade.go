package cli

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/inconshreveable/go-update"
	"github.com/spf13/cobra"
)

// upgradeRepo 是发布仓库的 owner/name；默认硬编码到 binary，配合 GH Releases API。
// 用 STDAGENT_REPO_OWNER / STDAGENT_REPO_NAME 环境变量在 fork 场景覆盖。
const (
	defaultUpgradeRepoOwner = "StringKe"
	defaultUpgradeRepoName  = "std-agent"
	upgradeProjectName      = "std-agent" // 归档前缀（goreleaser ProjectName）
	upgradeBinName          = "stdagent"
)

func upgradeRepoOwner() string {
	if v := osGetenv("STDAGENT_REPO_OWNER"); v != "" {
		return v
	}
	return defaultUpgradeRepoOwner
}

func upgradeRepoName() string {
	if v := osGetenv("STDAGENT_REPO_NAME"); v != "" {
		return v
	}
	return defaultUpgradeRepoName
}

func newUpgradeCmd() *cobra.Command {
	var force bool
	var pin string
	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "自我升级到最新版本（或 --version 指定的 tag）",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runUpgrade(cmd, upgradeOptions{Force: force, Pin: pin})
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "已是目标版本时仍强制重装")
	cmd.Flags().StringVar(&pin, "version", "", "指定 tag（如 v0.2.0），默认 latest")
	return cmd
}

type upgradeOptions struct {
	Force bool
	Pin   string
}

func runUpgrade(cmd *cobra.Command, opts upgradeOptions) error {
	tag := opts.Pin
	if tag == "" {
		latest, err := fetchLatestTag()
		if err != nil {
			return fmt.Errorf("fetch latest tag: %w", err)
		}
		tag = latest
	}
	if !strings.HasPrefix(tag, "v") {
		tag = "v" + tag
	}
	cmd.Printf("[upgrade] target: %s (current: %s)\n", tag, versionStr)

	if !opts.Force && versionMatchesTag(versionStr, tag) {
		cmd.Println("[upgrade] already up-to-date; use --force to reinstall")
		return nil
	}

	goos, goarch := runtime.GOOS, runtime.GOARCH
	ver := strings.TrimPrefix(tag, "v")
	archive := fmt.Sprintf("%s_%s_%s_%s", upgradeProjectName, ver, goos, goarch)
	ext := "tar.gz"
	if goos == "windows" {
		ext = "zip"
	}
	archive += "." + ext
	base := fmt.Sprintf("https://github.com/%s/%s/releases/download/%s", upgradeRepoOwner(), upgradeRepoName(), tag)

	cmd.Printf("[upgrade] downloading %s/%s\n", base, archive)
	archData, err := httpGetBytes(base + "/" + archive)
	if err != nil {
		return fmt.Errorf("download archive: %w", err)
	}

	cmd.Println("[upgrade] verifying checksum")
	checksumData, err := httpGetBytes(base + "/checksums.txt")
	if err != nil {
		return fmt.Errorf("download checksums: %w", err)
	}
	expected := findChecksum(string(checksumData), archive)
	if expected == "" {
		return fmt.Errorf("checksum for %s not found in checksums.txt", archive)
	}
	actual := sha256hex(archData)
	if !strings.EqualFold(expected, actual) {
		return fmt.Errorf("checksum mismatch for %s: expected %s actual %s", archive, expected, actual)
	}

	cmd.Println("[upgrade] extracting binary")
	binData, err := extractBinary(archData, ext)
	if err != nil {
		return fmt.Errorf("extract binary: %w", err)
	}

	cmd.Println("[upgrade] replacing current binary (atomic)")
	if err := update.Apply(bytes.NewReader(binData), update.Options{}); err != nil {
		if rerr := update.RollbackError(err); rerr != nil {
			return errors.Join(fmt.Errorf("apply: %w", err), fmt.Errorf("rollback: %w", rerr))
		}
		return fmt.Errorf("apply: %w", err)
	}
	cmd.Printf("[upgrade] done; new version: %s\n", tag)
	cmd.Println("[upgrade] tip: re-run `stdagent version` to confirm")
	return nil
}

// versionMatchesTag 比较 ldflags 注入的 version 与 tag
// version 形如 "0.2.0"（goreleaser 注入不带 v）；tag 形如 "v0.2.0"
func versionMatchesTag(version, tag string) bool {
	if version == "" || version == "dev" {
		return false
	}
	v := strings.TrimPrefix(version, "v")
	t := strings.TrimPrefix(tag, "v")
	return v == t
}

func fetchLatestTag() (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", upgradeRepoOwner(), upgradeRepoName())
	body, err := httpGetBytes(url)
	if err != nil {
		return "", err
	}
	var data struct {
		TagName string `json:"tag_name"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return "", fmt.Errorf("parse github api response: %w", err)
	}
	if data.TagName == "" {
		return "", errors.New("github api returned empty tag_name")
	}
	return data.TagName, nil
}

func httpGetBytes(url string) ([]byte, error) {
	client := &http.Client{Timeout: 60 * time.Second}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "stdagent-upgrade")
	if token := getEnv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http %s: %s", url, resp.Status)
	}
	return io.ReadAll(resp.Body)
}

func sha256hex(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func findChecksum(checksums, archiveName string) string {
	for _, line := range strings.Split(checksums, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[1] == archiveName {
			return fields[0]
		}
	}
	return ""
}

// extractBinary 从归档（tar.gz 或 zip）提取 stdagent binary 内容
func extractBinary(data []byte, ext string) ([]byte, error) {
	binName := upgradeBinName
	if ext == "zip" {
		zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
		if err != nil {
			return nil, err
		}
		want := binName + ".exe"
		for _, f := range zr.File {
			if f.Name == want || f.Name == binName {
				rc, oerr := f.Open()
				if oerr != nil {
					return nil, oerr
				}
				out, rerr := io.ReadAll(rc)
				_ = rc.Close()
				if rerr != nil {
					return nil, rerr
				}
				return out, nil
			}
		}
		return nil, fmt.Errorf("%s not found in zip", want)
	}
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer func() { _ = gz.Close() }()
	tr := tar.NewReader(gz)
	for {
		hdr, terr := tr.Next()
		if errors.Is(terr, io.EOF) {
			break
		}
		if terr != nil {
			return nil, terr
		}
		if hdr.Name == binName {
			return io.ReadAll(tr)
		}
	}
	return nil, fmt.Errorf("%s not found in tar.gz", binName)
}

// getEnv 是 os.Getenv 的薄封装（避免引入 os 依赖到本文件之外的轻测试）
func getEnv(key string) string {
	return osGetenv(key)
}
