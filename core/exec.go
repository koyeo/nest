package core

import "os/exec"

func Exec(dir, command string) (out string, err error) {

	c := exec.Command("bash", "-c", command)
	c.Dir = dir

	res, err := c.CombinedOutput()
	if err != nil {
		return
	}

	out = string(res)
	return
}
