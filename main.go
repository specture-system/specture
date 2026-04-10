package main

import "github.com/specture-system/specture/cmd"

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	cmd.SetVersion(version, commit)
	cmd.Execute()
}
