package core

import "errors"

var IncludeScriptErr = errors.New("include script error")
var InvalidIncludeScriptRuleErr = errors.New("invalid script include rule")
var InvalidIncludeScriptVarErr = errors.New("invalid script include var")
var ScriptNotExistErr = errors.New("script not exist")
