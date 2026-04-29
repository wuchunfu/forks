package main

import (
	"embed"

	"github.com/cicbyte/forks/assets"
	"github.com/cicbyte/forks/cmd"
)

//go:embed web/dist
var webDist embed.FS

func main() {
	assets.WebDist = webDist
	cmd.Execute()
}
