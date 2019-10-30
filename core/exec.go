package core

import (
	"bufio"
	"fmt"
	"github.com/ttacon/chalk"
	"golang.org/x/crypto/ssh"
	"log"
	"nest/logger"
	"os/exec"
	"time"
)

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

func PipeExec(dir, command string) (err error) {

	log.Println(chalk.Green.Color("Exec command:"), command)

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

	err = c.Run()
	if err != nil {
		logger.Error("Exec command run error: ", err)
	}

	return
}

func SSHPipeExec(sshClient *ssh.Client, command string) (err error) {

	session, err := sshClient.NewSession()
	if err != nil {
		logger.Error("New remote ssh session error: ", err)
		return
	}

	defer func() {
		_ = session.Close()
	}()

	log.Println(chalk.Green.Color("Run remote ssh command:"), command)

	stderr, err := session.StderrPipe()
	if err != nil {
		logger.Error("Run remote ssh command get stderr error: ", err)
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		logger.Error("Run remote ssh command get stdout error: ", err)
	}

	out := make(chan string, 1048576)
	defer func() {
		for len(out) > 0 {
			time.Sleep(500 * time.Millisecond)
			close(out)
		}
	}()

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

	err = session.Run(command)
	if err != nil {
		logger.Error("Run remote ssh command run error: ", err)
	}

	return
}
