package notify

import (
	"fmt"
	"github.com/gen2brain/beeep"
)

func Alert(title, message string, icon ...string) {
	err := beeep.Beep(beeep.DefaultFreq, beeep.DefaultDuration)
	if err != nil {
		return
	}
	if len(icon) == 0 {
		icon = make([]string, 1)
	}
	err = beeep.Alert(title, message, icon[0])
	if err != nil {
		panic(err)
	}
}

func BuildDone(count int) {

	var message string

	if count > 1 {
		message = fmt.Sprintf("exec %d tasks", count)
	} else {
		message = fmt.Sprintf("exec %d task", count)
	}

	Alert("Build done", message, "assets/success.png")
}
