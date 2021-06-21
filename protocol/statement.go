package protocol

import (
	"fmt"
	"regexp"
)

var (
	varReg = regexp.MustCompile(`[a-zA-Z_][a-zA-Z0-9_]*`)
)

type Statement string

func (p Statement) Render(vars map[string]string) (res string, err error) {
	
	res = string(p)
	
	if vars == nil {
		return
	}
	for k, v := range vars {
		if !varReg.MatchString(k) {
			err = fmt.Errorf("illegal var name '%s'", k)
			return
		}
		var reg *regexp.Regexp
		reg, err = regexp.Compile(`\{\{\s*` + k + `\s*\}\}`)
		if err != nil {
			err = fmt.Errorf("compile var '%s' regexp error: %s", k, err)
			return
		}
		res = reg.ReplaceAllString(res, v)
	}
	
	return
}
