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

type DeployStart struct {
	Task   *core.Task
	Server *core.Server
	Start  *core.DaemonStart
}

func NewDeployStart(task *core.Task, server *core.Server, start *core.DaemonStart) *DeployStart {
	return &DeployStart{Task: task, Server: server, Start: start}
}

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
	ctx.SetCli(c)

	count := 0
	deploySuccess := true

	deployStarts := make([]*DeployStart, 0)
	for _, changeTask := range change.TaskList {

		for _, changeDeploy := range changeTask.Deploy {

			if changeDeploy.Type == enums.ChangeTypeDelete || !changeDeploy.Modify {
				continue
			}

			if !deploySuccess {
				break
			}

			count++

			if changeTask.Task.Name != "" {
				log.Println(chalk.Green.Color(fmt.Sprintf("Task: %s(%s)", changeTask.Task.Id, changeTask.Task.Name)))
			} else {
				log.Println(chalk.Green.Color(fmt.Sprintf("Task: %s", changeTask.Task.Id)))
			}
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
				if deploy.Daemon != nil && deploy.Daemon.Start != nil {
					deployStarts = append(deployStarts, NewDeployStart(task, server, deploy.Daemon.Start))
				}
			}

			if len(servers) > 0 {
				deploySuccess = printDeployResult(servers, deployResult)
			}
		}

	}

	if count == 0 {
		fmt.Println(chalk.Green.Color("no change"))
		return
	}

	if !deploySuccess {
		return
	}

	if len(deployStarts) == 0 {
		err = core.CommitDeploy(change)
		if err != nil {
			return
		}
		log.Println(chalk.Green.Color("Commit deploy Success"))
		notify.DeployDone(count)
		return
	}

	startSuccess := true
	startResult := make(map[string]error)
	servers := make([]*core.Server, 0)
	for _, deployStart := range deployStarts {
		err = core.ExecStart(ctx, deployStart.Task, deployStart.Server, deployStart.Start)
		startResult[deployStart.Server.Id] = err
		servers = append(servers, deployStart.Server)
	}

	startSuccess = printStartResult(servers, startResult)

	if startSuccess {
		err = core.CommitDeploy(change)
		if err != nil {
			return
		}
		log.Println(chalk.Green.Color("Commit deploy Success"))
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

	if len(failed) > 0 {
		log.Println(chalk.Red.Color("Deploy result"), fmt.Sprintf("total: %d, success: %d, failed: %d", total, len(success), len(failed)))
	} else {
		log.Println(chalk.Green.Color("Deploy result"), fmt.Sprintf("total: %d, success: %d, failed: %d", total, len(success), len(failed)))
	}

	if len(success) > 0 {
		if len(success) > 1 {
			log.Println(chalk.Green.Color("Deploy success servers:"))
		} else {
			log.Println(chalk.Green.Color("Deploy success server:"))
		}
		for _, v := range success {
			fmt.Println(v.Id, strings.Repeat(" ", enums.StatusMarginLeft), v.Name, strings.Repeat(" ", enums.StatusMarginLeft), v.SSH.Ip)
		}
	}

	if len(failed) > 0 {
		if len(failed) > 1 {
			log.Println(chalk.Red.Color("Deploy failed servers:"))
		} else {
			log.Println(chalk.Red.Color("Deploy failed server:"))
		}
		for _, v := range failed {
			fmt.Println(v.Id, strings.Repeat(" ", enums.StatusMarginLeft), v.Name, strings.Repeat(" ", enums.StatusMarginLeft), v.SSH.Ip)
		}
	}

	return total == len(success)
}

func printStartResult(servers []*core.Server, deployResult map[string]error) bool {

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

	if len(failed) > 0 {
		log.Println(chalk.Red.Color("Start result"), fmt.Sprintf("total: %d, success: %d, failed: %d", total, len(success), len(failed)))
	} else {
		log.Println(chalk.Green.Color("Start result"), fmt.Sprintf("total: %d, success: %d, failed: %d", total, len(success), len(failed)))
	}

	if len(success) > 0 {
		if len(success) > 1 {
			log.Println(chalk.Green.Color("Start success servers:"))
		} else {
			log.Println(chalk.Green.Color("Start success server:"))
		}
		for _, v := range success {
			fmt.Println(v.Id, strings.Repeat(" ", enums.StatusMarginLeft), v.Name, strings.Repeat(" ", enums.StatusMarginLeft), v.SSH.Ip)
		}
	}

	if len(failed) > 0 {
		if len(failed) > 1 {
			log.Println(chalk.Red.Color("Start failed servers:"))
		} else {
			log.Println(chalk.Red.Color("Start failed server:"))
		}
		for _, v := range failed {
			fmt.Println(v.Id, strings.Repeat(" ", enums.StatusMarginLeft), v.Name, strings.Repeat(" ", enums.StatusMarginLeft), v.SSH.Ip)
		}
	}

	return total == len(success)
}
