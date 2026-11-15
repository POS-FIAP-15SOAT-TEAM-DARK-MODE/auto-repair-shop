package api

import _ "embed"

// SwaggerBytes is the gateway's own OpenAPI document (/swagger.yaml).
//
//go:embed swagger.yaml
var SwaggerBytes []byte

// IndexHTML is the documentation index page (/swagger/index.html).
//
//go:embed index.html
var IndexHTML []byte
