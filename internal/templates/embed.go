// Package templates содержит встроенные шаблоны HTML.
package templates

import "embed"

//go:embed *.html
var FS embed.FS
