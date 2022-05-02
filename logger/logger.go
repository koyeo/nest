package logger

import (
	"fmt"
	"github.com/gozelle/_color"
	"strings"
	"time"
)

func PrintStep(step string, args ...string) {
	for i := 0; i < len(args); i++ {
		if args[i] == "" {
			args = append(args[0:i], args[i+1:]...)
		}
	}
	fmt.Printf("%s %s\n",
		_color.WhiteString(time.Now().Format("2006-01-02 15:04:05")),
		strings.Join([]string{
			_color.New(_color.FgHiGreen, _color.Bold).Sprint("[Nest]"),
			_color.New(_color.FgHiGreen).Sprintf("[%s]", step),
			strings.Join(args, " "),
		}, " "),
	)
}

func Print(msg string) {
	fmt.Printf(fmt.Sprintf("%s %s", _color.RedString(time.Now().Format("2006-01-02 15:04:05")), msg))
}
