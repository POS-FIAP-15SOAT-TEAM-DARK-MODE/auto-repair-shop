// Package config embeds the gateway routing table.
package config

import _ "embed"

// RoutesYAML is the routing table read by the gateway at startup.
//
//go:embed routes.yaml
var RoutesYAML []byte
