package execer

import (
	"bufio"
	"github.com/koyeo/nest/enums"
	"github.com/koyeo/nest/logger"
	"github.com/ttacon/chalk"
	"log"
	"os"
	"os/exec"
	"sync"
)

func RunCommand(shell, dir, command string) (err error) {
	
	log.Println(chalk.Green.Color("[Exec command]"), command)
	
	if shell == "" {
		shell = enums.DefaultShell
	} else {
		log.Println(chalk.Green.Color("[Use shell]"), shell)
	}
	
	c := exec.Command(shell, "-c", command)
	c.Dir = dir
	c.Env = os.Environ()
	c.Stdin = os.Stdin
	
	stdout, err := c.StdoutPipe()
	if err != nil {
		logger.Error("[Exec error]", err)
	}
	
	c.Stderr = c.Stdout
	
	out := make(chan []byte)
	defer func() {
		close(out)
	}()
	
	var wg sync.WaitGroup
	
	wg.Add(1)
	go func() {
		scanner := bufio.NewScanner(stdout)
		scanner.Split(bufio.ScanBytes)
		for scanner.Scan() {
			_, _ = os.Stdout.Write(scanner.Bytes())
		}
		wg.Done()
	}()
	
	err = c.Run()
	wg.Wait()
	
	if err != nil {
		return
	}
	
	return
}

func RunScript(shell, dir, file string) (err error) {
	
	if shell == "" {
		shell = enums.DefaultShell
	} else {
		log.Println(chalk.Green.Color("[Use shell]"), shell)
	}
	
	c := exec.Command(shell, file)
	c.Dir = dir
	c.Env = os.Environ()
	c.Stdin = os.Stdin
	
	stdout, err := c.StdoutPipe()
	if err != nil {
		logger.Error("[Exec error]", err)
	}
	
	c.Stderr = c.Stdout
	
	out := make(chan []byte)
	defer func() {
		close(out)
	}()
	
	var wg sync.WaitGroup
	
	wg.Add(1)
	go func() {
		scanner := bufio.NewScanner(stdout)
		scanner.Split(bufio.ScanBytes)
		for scanner.Scan() {
			_, _ = os.Stdout.Write(scanner.Bytes())
		}
		wg.Done()
	}()
	
	err = c.Run()
	wg.Wait()
	
	if err != nil {
		return
	}
	
	return
}
