package frontend

import "embed"

//go:embed dist
var DistFS embed.FS

//go:embed template.html
var TemplateHTML string
