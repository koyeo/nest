package _string

import (
	"encoding/json"
	"fmt"
)

func StringsToJson(s []string) string {
	d, _ := json.Marshal(s)
	return fmt.Sprintf(string(d))
}
