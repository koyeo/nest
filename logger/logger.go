package logger

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ttacon/chalk"
	"log"
	"os"
	"runtime/debug"
	"strings"
	"time"
)

func Done() {
	log.Println(chalk.Green.Color("command execute done"))
}

func MakeDone() {
	log.Println(chalk.Green.Color("Make done"))
}

func Successf(format string, a ...interface{}) {
	log.Println(chalk.Green.Color(fmt.Sprintf(format, a...)))
}

func ReadFile(path string) {
	wd, _ := os.Getwd()
	path = strings.TrimPrefix(path, wd+"/")
	log.Println(chalk.Green.Color("Read file:"), chalk.Green.Color(chalk.Bold.TextStyle(path)))
}

func MakeDirSuccess(path string) {
	wd, _ := os.Getwd()
	path = strings.TrimPrefix(path, wd+"/")
	log.Println(chalk.Green.Color("Make dir:"), chalk.Green.Color(chalk.Bold.TextStyle(path)))
}

func MakeFileSuccess(path string) {
	wd, _ := os.Getwd()
	path = strings.TrimPrefix(path, wd+"/")
	log.Println(chalk.Green.Color("Make file:"), chalk.Green.Color(chalk.Bold.TextStyle(path)))
}

func CleanFileSuccess(path string) {
	wd, _ := os.Getwd()
	path = strings.TrimPrefix(path, wd+"/")
	log.Println(chalk.Cyan.Color("Clean file:"), chalk.Cyan.Color(chalk.Bold.TextStyle(path)))
}

func TemplateError(msg string, err error) {
	if err == nil {
		err = errors.New("")
	}
	errMsg := err.Error()
	log.Println(chalk.Red.Color(msg), chalk.Red.Color(chalk.Bold.TextStyle(errMsg)))
}

func Success(msg string) {
	log.Println(chalk.Green.Color(msg), chalk.Green.Color(msg))
}

func Print(msg string) {
	fmt.Printf(fmt.Sprintf("%s %s", time.Now().Format(time.RFC3339), msg))
}

func Error(msg string, err error) {
	if err == nil {
		err = errors.New("")
	}
	log.Println(chalk.Red.Color(msg), chalk.Red.Color(chalk.Bold.TextStyle(err.Error())))
}

func Fatal(msg string, err error) {
	if err == nil {
		err = errors.New("")
	}
	log.Println(msg, err.Error())
	debug.PrintStack()
	os.Exit(1)
}

func DebugPrint(elem interface{}) {
	c, err := json.MarshalIndent(elem, "", "\t")
	if err != nil {
		fmt.Println("Call debug print error", err)
	}
	fmt.Println(string(c))
}
