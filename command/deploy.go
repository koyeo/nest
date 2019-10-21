package command

import (
	"fmt"
	"github.com/ttacon/chalk"
	"github.com/urfave/cli"
	"log"
	"nest/core"
	"nest/enums"
	"nest/logger"
	"nest/notify"
	"os"
)

func DeployCommand(c *cli.Context) (err error) {

	defer func() {
		if err != nil {
			os.Exit(1)
		}
	}()

	change, err := core.MakeChange()
	if err != nil {
		return
	}

	env := mustGetEnv(c, 0)
	if env == nil {
		return
	}

	ctx, err := core.Prepare()
	if err != nil {
		return
	}

	count := 0

	for _, changeTask := range change.TaskList {

		for _, changeDeploy := range changeTask.Deploy {

			if changeDeploy.Type == enums.ChangeTypeDelete || !changeDeploy.Modify {
				continue
			}

			count++

			task := ctx.GetTask(changeDeploy.TaskId)
			if task == nil {
				err = fmt.Errorf("task \"%s\" is nil", changeDeploy.TaskId)
				logger.Error("Deploy error: ", err)
				return
			}

			env := ctx.GetEnv(changeDeploy.EnvId)
			if env == nil {
				err = fmt.Errorf("env \"%s\" is nil", changeDeploy.EnvId)
				logger.Error("Deploy error: ", err)
				return
			}

			deploy := task.GetDeploy(changeDeploy.EnvId)
			if deploy == nil {
				err = fmt.Errorf("deploy \"%s\" is nil", changeDeploy.EnvId)
				logger.Error("Deploy error: ", err)
				return
			}

			var servers []*core.Server
			for _, serverId := range deploy.Server {
				server := env.GetServer(serverId)
				if server == nil {
					err = fmt.Errorf("server \"%s\" is nil", serverId)
					logger.Error("Deploy error: ", err)
					return
				}
				if server.SSH == nil {
					err = fmt.Errorf("server \"%s\" ssh is nil", serverId)
					logger.Error("Deploy error: ", err)
					return
				}
				servers = append(servers, server)
			}

			deployResult := make(map[string]error)
			for _, server := range servers {
				err = core.ExecDeploy(ctx, task, server, deploy, changeDeploy)
				deployResult[server.Id] = err
			}

			logger.DebugPrint(deployResult)
		}

	}

	if count == 0 {
		fmt.Println(chalk.Green.Color("no change"))
		return
	}

	//err = core.Commit(change)
	//if err != nil {
	//	return
	//}
	log.Println(chalk.Green.Color("Commit success"))

	notify.DeployDone(count)

	return
}
