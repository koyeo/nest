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
	"strings"
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
	success := true

	for _, changeTask := range change.TaskList {

		for _, changeDeploy := range changeTask.Deploy {

			if changeDeploy.Type == enums.ChangeTypeDelete || !changeDeploy.Modify {
				continue
			}

			if !success {
				break
			}

			count++
			log.Println(chalk.Green.Color("Deploy start"))
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

			if len(servers) > 0 {
				success = printDeployResult(servers, deployResult)
			}
		}

	}

	if count == 0 {
		fmt.Println(chalk.Green.Color("no change"))
		return
	}

	if success {
		err = core.CommitDeploy(change)
		if err != nil {
			return
		}
		log.Println(chalk.Green.Color("Commit deploy success"))
		notify.DeployDone(count)
	}

	return
}

func printDeployResult(servers []*core.Server, deployResult map[string]error) bool {

	total := len(deployResult)
	success := make([]*core.Server, 0)
	failed := make([]*core.Server, 0)
	for _, v := range servers {
		if deployResult[v.Id] == nil {
			success = append(success, v)
		} else {
			failed = append(failed, v)
		}
	}

	log.Println(chalk.Green.Color("Deploy result"), fmt.Sprintf("total: %d, success: %d, failed: %d", total, len(success), len(failed)))

	if len(success) > 0 {
		if len(success) > 1 {
			log.Println(chalk.Green.Color("Deploy success servers:"))
		} else {
			log.Println(chalk.Green.Color("Deploy success server:"))
		}
		for _, v := range success {
			fmt.Println(v.Id, strings.Repeat(" ", enums.StatusMarginLeft), v.Name, strings.Repeat(" ", enums.StatusMarginLeft), v.Ip)
		}
	}

	if len(failed) > 0 {
		if len(failed) > 1 {
			log.Println(chalk.Red.Color("Deploy failed servers:"))
		} else {
			log.Println(chalk.Red.Color("Deploy failed server:"))
		}
		for _, v := range failed {
			fmt.Println(v.Id, strings.Repeat(" ", enums.StatusMarginLeft), v.Name, strings.Repeat(" ", enums.StatusMarginLeft), v.Ip)
		}
	}

	return total == len(success)
}
