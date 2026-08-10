package specdata

import "embed"

const (
	SpecVersion     = "1.0.0-rc.1"
	ProtocolVersion = "1.0"
)

// Schemas contains the normative control-plane schemas needed by dota when it
// is installed outside this repository.
//
//go:embed schema/*.json
var Schemas embed.FS
