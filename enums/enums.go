package enums

const (
	ConfigFile              = "nest.yml"
	WorkspaceDir            = ".nest"
	DataDir                 = "data"
	DataFile                = "nest.db"
	BinDir                  = "bin"
	RemoteOptScriptDir      = "/opt/script"
	RemoteScriptDir         = "script"
	GoSuffix                = ".go"
	FirstLevel              = 1
	SecondLevel             = 2
	ThirdLevel              = 4
	StatusMarginLeft        = 6
	ChangeTypeNone          = 0
	ChangeTypeNew           = 1
	ChangeTypeUpdate        = 2
	ChangeTypeDelete        = 3
	ChangeTypeForce         = 4
	ScriptTypeBefore        = "before"
	ScriptTypeAfter         = "after"
	ScriptExtendIdent       = "@"
	Q                       = "q"
	Quiet                   = "quiet"
	D                       = "d"
	Daemon                  = "daemon"
	BuildBinConst           = "__BIN__"
	BinSourceBuild          = "build"
	BinSourceUrl            = "url"
	ZipSuffix               = ".zip"
	DeploySourceBuild       = "build"
	ScriptPositionSplitFlag = "@"
	ScriptPositionBefore    = "before"
	ScriptPositionAfter     = "after"
	ScriptVarNameFlag       = "$"
	ScriptVarSplitFlag      = ","
	ScriptVarAssignFlag     = "="
	ScriptVarWrapFlagLeft   = "("
	ScriptVarWrapFlagRight  = ")"
	RelativePathPrefix      = "./"
)
