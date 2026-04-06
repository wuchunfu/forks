package utils

import (
	"os"
	"strings"
)

var Version = "dev"

func init() {
	if data, err := os.ReadFile("VERSION"); err == nil {
		if v := strings.TrimSpace(string(data)); v != "" {
			Version = v
		}
	}
}
