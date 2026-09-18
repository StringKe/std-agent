package cli

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

// osGetenv wraps os.Getenv so tests can stub it
var osGetenv = os.Getenv

// osExecutable wraps os.Executable so tests can stub it
var osExecutable = os.Executable

// evalSymlinks wraps filepath.EvalSymlinks so tests can stub it
var evalSymlinks = filepath.EvalSymlinks

// execBrewUpgrade runs `brew upgrade` against the tap; stubbed in tests
var execBrewUpgrade = func(cmd *cobra.Command) error {
	c := exec.Command("brew", "upgrade", brewTapRef)
	c.Stdout = cmd.OutOrStdout()
	c.Stderr = cmd.ErrOrStderr()
	return c.Run()
}
