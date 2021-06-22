package execer

import (
	"bufio"
	"github.com/koyeo/nest/constant"
	"github.com/koyeo/nest/logger"
	"github.com/ttacon/chalk"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
)

func Exec(dir, command string) (out string, err error) {
	
	c := exec.Command(constant.BASH, "-c", command)
	c.Dir = dir
	
	res, err := c.CombinedOutput()
	if err != nil {
		return
	}
	
	out = strings.TrimSpace(string(res))
	
	return
}

func RunCommand(shell, dir, command string) (err error) {
	
	if shell == "" {
		shell = constant.BASH
	} else {
		log.Println(chalk.Green.Color("[Use shell]"), shell)
	}
	
	c := exec.Command(shell, "-c", command)
	c.Dir = dir
	c.Env = os.Environ()
	c.Stdin = os.Stdin
	
	stderr, err := c.StderrPipe()
	if err != nil {
		logger.Error("[Run remote ssh command get stderr error]", err)
	}
	
	stdout, err := c.StdoutPipe()
	if err != nil {
		logger.Error("[Exec get stdout pipe error]", err)
		return
	}
	
	c.Stderr = c.Stdout
	
	out := make(chan []byte)
	defer func() {
		close(out)
	}()
	
	var wg sync.WaitGroup
	
	wg.Add(2)
	go func() {
		scanner := bufio.NewScanner(stdout)
		scanner.Split(bufio.ScanBytes)
		for scanner.Scan() {
			_, _ = os.Stdout.Write(scanner.Bytes())
		}
		wg.Done()
	}()
	
	go func() {
		scanner := bufio.NewScanner(stderr)
		scanner.Split(bufio.ScanBytes)
		for scanner.Scan() {
			_, _ = os.Stderr.Write(scanner.Bytes())
		}
		wg.Done()
	}()
	
	err = c.Run()
	wg.Wait()
	
	if err != nil {
		logger.Error("[Exec error]", err)
		return
	}
	
	return
}

func RunScript(shell, dir, file string) (err error) {
	
	if shell == "" {
		shell = constant.BASH
	} else {
		log.Println(chalk.Green.Color("[Use shell]"), shell)
	}
	
	c := exec.Command(shell, file)
	c.Dir = dir
	c.Env = os.Environ()
	c.Stdin = os.Stdin
	
	stderr, err := c.StderrPipe()
	if err != nil {
		logger.Error("[Run remote ssh command get stderr error]", err)
	}
	
	stdout, err := c.StdoutPipe()
	if err != nil {
		logger.Error("[Exec get stdout pipe error]", err)
		return
	}
	
	c.Stderr = c.Stdout
	
	out := make(chan []byte)
	defer func() {
		close(out)
	}()
	
	var wg sync.WaitGroup
	
	wg.Add(2)
	go func() {
		scanner := bufio.NewScanner(stdout)
		scanner.Split(bufio.ScanBytes)
		for scanner.Scan() {
			_, _ = os.Stdout.Write(scanner.Bytes())
		}
		wg.Done()
	}()
	
	go func() {
		scanner := bufio.NewScanner(stderr)
		scanner.Split(bufio.ScanBytes)
		for scanner.Scan() {
			_, _ = os.Stderr.Write(scanner.Bytes())
		}
		wg.Done()
	}()
	
	err = c.Run()
	wg.Wait()
	
	if err != nil {
		logger.Error("[Exec error]", err)
		return
	}
	
	return
}

func HomePath() (path string, err error) {
	
	path, err = Exec("", "echo ~")
	if err != nil {
		return
	}
	
	return
}
