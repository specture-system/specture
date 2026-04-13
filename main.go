package main

import (
	_ "embed"
	"strings"

	"github.com/specture-system/specture/cmd"
)

//go:embed VERSION
var versionFile string

var (
	version = strings.TrimSpace(versionFile)
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	cmd.SetVersion(version, commit)
	cmd.Execute()
}
