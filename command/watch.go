package command

import (
	"github.com/fsnotify/fsnotify"
	"github.com/urfave/cli"
	"log"
)

func WatchCommand(c *cli.Context) (err error) {

	//defer func() {
	//	if err != nil {
	//		os.Exit(1)
	//	}
	//}()
	//
	//task := getTask(c)
	//if task == nil {
	//	return
	//}
	//
	//err = Exec(task.Directory, task.Run)
	//if err != nil {
	//	return
	//}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = watcher.Close()
	}()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Name)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add("/tmp/foo")
	if err != nil {
		log.Fatal(err)
	}
	<-done

	return
}
