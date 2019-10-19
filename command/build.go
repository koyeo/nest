package command

import (
	"bufio"
	"fmt"
	"github.com/ttacon/chalk"
	"github.com/urfave/cli"
	"log"
	"nest/core"
	"nest/enums"
	"nest/logger"
	"os"
	"os/exec"
	"path/filepath"
)

func BuildCommand(c *cli.Context) (err error) {

	defer func() {
		if err != nil {
			os.Exit(1)
		}
	}()

	change, err := core.MakeChange()
	if err != nil {
		return
	}

	count := 0
	for _, task := range change.TaskList {

		if task.Type == enums.ChangeTypeDelete {
			continue
		}

		if !task.Modify {
			continue
		}

		count++
		var dir string
		dir, err = filepath.Abs(task.Task.Build.Directory)
		if err != nil {
			logger.Error("Modify get directory error: ", err)
			return
		}

		log.Println(chalk.Green.Color("Modify:"), task.Task.Name)
		log.Println(chalk.Green.Color("Modify task start"))
		log.Println(chalk.Green.Color("Exec directory:"), dir)

		for _, command := range task.Task.Build.Command {
			log.Println(chalk.Green.Color("Exec command:"), command)
			err = Exec(task.Task.Build.Directory, command)
			if err != nil {
				return
			}
		}

		log.Println(chalk.Green.Color("Modify task end"))
	}

	if count == 0 {
		fmt.Println(chalk.Green.Color("no change"))
		return
	}

	err = core.Commit(change)
	if err != nil {
		return
	}

	return
}

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
		select {
		case m := <-out:
			fmt.Println(m)
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
