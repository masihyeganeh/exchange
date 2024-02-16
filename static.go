// Package exchange is used for embedding static files
package exchange

import "embed"

// StaticFiles are static files embedded in the executable
//
//go:embed static
var StaticFiles embed.FS
