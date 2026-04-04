package main

import (
	"embed"

	"forks.com/m/assets"
	"forks.com/m/cmd"
)

//go:embed web/dist
var webDist embed.FS

func main() {
	assets.WebDist = webDist
	cmd.Execute()
}
