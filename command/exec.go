package command

import (
	"bufio"
	"fmt"
	"github.com/ttacon/chalk"
	"log"
	"nest/logger"
	"os/exec"
)



func PipeExec(dir, command string) (err error) {

	log.Println(chalk.Green.Color("PipeExec command:"), command)

	c := exec.Command("bash", "-c", command)
	c.Dir = dir

	stderr, err := c.StderrPipe()
	if err != nil {
		logger.Error("PipeExec command get stderr error: ", err)
	}

	stdout, err := c.StdoutPipe()
	if err != nil {
		logger.Error("PipeExec command get stdout error: ", err)
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
		logger.Error("PipeExec command start error: ", err)
	}

	err = c.Wait()
	if err != nil {
		logger.Error("PipeExec command wait error: ", err)
	}

	return
}
