package main

import (
	"os"
	"runtime/debug"
)

var version = "dev"

func main() {
	if version == "dev" {
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
			version = info.Main.Version
		}
	}
	if err := newRootCmd(version).Execute(); err != nil {
		os.Exit(1)
	}
}
