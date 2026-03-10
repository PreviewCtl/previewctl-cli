package yaml

import (
	"embed"
	"io/fs"
)

//go:embed default/*
var embededFs embed.FS

func GetDefaultYamlV1() ([]byte, error) {
	return fs.ReadFile(embededFs, "default/defaultv1.yml")
}
