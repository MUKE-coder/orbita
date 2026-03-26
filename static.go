package orbita

import "embed"

//go:embed all:web/dist
var StaticFiles embed.FS
