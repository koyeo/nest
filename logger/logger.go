package logger

import (
	"fmt"
	"github.com/gozelle/_color"
	"strings"
	"time"
)

func timestamp() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func Step(taskKey, taskComment, emoji string, args ...string) {
	for i := 0; i < len(args); i++ {
		if args[i] == "" {
			args = append(args[0:i], args[i+1:]...)
		}
	}
	if taskComment != "" {
		taskKey = taskComment
	}
	fmt.Printf("%s %s %s %s\n",
		_color.WhiteString(timestamp()),
		_color.New(_color.FgBlack, _color.BgWhite).Sprintf("[%s]", taskKey),
		emoji,
		strings.Join([]string{
			strings.Join(args, " "),
		}, " "),
	)
}

func Print(msg string) {
	fmt.Printf(fmt.Sprintf("%s %s", _color.RedString(timestamp()), msg))
}

func Error(err error) {
	fmt.Printf("%s %s\n",
		_color.WhiteString(timestamp()),
		_color.New(_color.FgRed).Sprintf("%s", err),
	)
}
