package core

import "time"

func BuildTimestamp() string {
	return time.Now().Format("2006_01_02_15_04_05")
}
