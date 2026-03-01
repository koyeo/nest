package common

const (
	DefaultConfigFile = "nest.yaml"
	TmpWorkspace      = ".nest"
)

// ConfigFile is the active config file path, settable via --config flag.
var ConfigFile = DefaultConfigFile
