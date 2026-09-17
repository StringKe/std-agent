package cli

import "os"

// osGetenv wraps os.Getenv so tests can stub it
var osGetenv = os.Getenv
