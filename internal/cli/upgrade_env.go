package cli

import "os"

// osGetenv 包一层方便测试桩替换
var osGetenv = os.Getenv
