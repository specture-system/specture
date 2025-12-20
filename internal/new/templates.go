package new

import (
	"embed"
)

//go:embed templates/*.md
var templateFiles embed.FS
