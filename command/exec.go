package command

import (
	"bufio"
	"fmt"
	"nest/logger"
	"os/exec"
)

func Exec(dir, command string) (err error) {

	c := exec.Command("bash", "-c", command)
	c.Dir = dir

	stderr, err := c.StderrPipe()
	if err != nil {
		logger.Error("Exec command get stderr error: ", err)
	}

	stdout, err := c.StdoutPipe()
	if err != nil {
		logger.Error("Exec command get stdout error: ", err)
	}

	out := make(chan string)
	defer close(out)

	go func() {
		scanner := bufio.NewScanner(stdout)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			out <- scanner.Text()
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			out <- scanner.Text()
		}
	}()

	go func() {
		for {
			m := <-out
			if m != "" {
				fmt.Println(m)
			}
		}
	}()

	err = c.Start()
	if err != nil {
		logger.Error("Exec command start error: ", err)
	}

	err = c.Wait()
	if err != nil {
		logger.Error("Exec command wait error: ", err)
	}

	return
}
