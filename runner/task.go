package runner

import (
	"context"
	"crypto/sha1"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/gozelle/_color"
	"github.com/gozelle/_exec"
	"github.com/gozelle/_fs"
	"github.com/koyeo/nest/config"
	"github.com/koyeo/nest/logger"
	"github.com/koyeo/nest/protocol"
	"github.com/koyeo/nest/storage"
	"github.com/koyeo/nest/utils/_tar"
	"github.com/koyeo/nest/utils/unit"
)

func NewTaskRunner(conf *protocol.Config, task *protocol.Task, key string) *TaskRunner {
	return &TaskRunner{conf: conf, task: task, key: key, uploadedKeys: map[string]string{}}
}

type TaskRunner struct {
	key          string
	conf         *protocol.Config
	task         *protocol.Task
	parents      map[string]bool
	uploadedKeys map[string]string // source basename → object key
}

func (p TaskRunner) prepareEnviron() []string {
	envs := map[string]string{}
	// Start with the current process environment so PATH etc. are preserved
	for _, entry := range os.Environ() {
		for i := 0; i < len(entry); i++ {
			if entry[i] == '=' {
				envs[entry[:i]] = entry[i+1:]
				break
			}
		}
	}
	// Overlay global config envs
	for k, v := range p.conf.Envs {
		envs[k] = v
	}
	// Overlay task-specific envs (highest priority)
	for k, v := range p.task.Envs {
		envs[k] = v
	}
	environ := make([]string, 0, len(envs))
	for k, v := range envs {
		environ = append(environ, fmt.Sprintf("%s=%s", k, v))
	}
	return environ
}

func (p TaskRunner) Exec() (err error) {
	for _, step := range p.task.Steps {
		if step.Use != "" {
			if err = p.use(step.Use); err != nil {
				return
			}
		} else if step.Upload != nil {
			if err = p.upload(step.Upload); err != nil {
				return
			}
		} else if step.Deploy != nil {
			if err = p.deploy(step.Deploy); err != nil {
				return
			}
		} else {
			if err = p.execute(step); err != nil {
				return
			}
		}
	}
	return
}

func (p TaskRunner) use(key string) (err error) {

	defer func() {
		if err == nil {
			p.printExecUseEnd()
		}
	}()

	// check  circle dependency
	if p.parents != nil {
		if _, ok := p.parents[key]; ok {
			err = fmt.Errorf("task: %s depend task: %s circlely", p.key, key)
			return
		}
	}

	task, ok := p.conf.Tasks[key]
	if !ok {
		err = fmt.Errorf("use task: '%s' not found", key)
		return
	}
	taskRunner := NewTaskRunner(p.conf, task, key)

	// store parent task key to avoid circle dependency
	taskRunner.parents = map[string]bool{
		p.key: true,
	}
	p.printExecUseStart(key, task.Comment)
	err = taskRunner.Exec()
	if err != nil {
		return
	}
	return
}

// upload compresses the local source and uploads it to cloud storage.
func (p *TaskRunner) upload(u *protocol.Upload) error {
	cfg := config.Load()
	cred, err := cfg.DecryptStorage(u.Storage)
	if err != nil {
		return fmt.Errorf("storage '%s': %s", u.Storage, err)
	}

	store, err := storage.NewFromCredential(cred)
	if err != nil {
		return fmt.Errorf("create storage client error: %s", err)
	}

	// Verify source exists
	ok, err := _fs.Exists(u.Source)
	if err != nil {
		return fmt.Errorf("check source error: %s", err)
	}
	if !ok {
		return fmt.Errorf("upload source not found: %s", u.Source)
	}

	// Compress source
	sourceName := path.Base(u.Source)
	bundleName := fmt.Sprintf("%s.tar.gz", sourceName)
	tmpDir := NestTmpDir()
	bundlePath := fmt.Sprintf("%s/%s", tmpDir, bundleName)
	defer cleanNestTempDir()

	sourceFile, err := os.Open(u.Source)
	if err != nil {
		return fmt.Errorf("open source error: %s", err)
	}
	defer func() { _ = sourceFile.Close() }()

	if err = _tar.Compress([]*os.File{sourceFile}, bundlePath); err != nil {
		return fmt.Errorf("compress error: %s", err)
	}

	// Compute SHA1 for object key
	bundleData, err := os.ReadFile(bundlePath)
	if err != nil {
		return fmt.Errorf("read bundle error: %s", err)
	}
	sha := fmt.Sprintf("%x", sha1.Sum(bundleData))

	// Build object key: nest/{sha1(cwd)}/{basename}.tar.gz
	cwd, _ := os.Getwd()
	cwdHash := fmt.Sprintf("%x", sha1.Sum([]byte(cwd)))[:8]
	objectKey := fmt.Sprintf("nest/%s/%s", cwdHash, bundleName)

	// Upload
	bundleFile, err := os.Open(bundlePath)
	if err != nil {
		return fmt.Errorf("open bundle error: %s", err)
	}
	defer func() { _ = bundleFile.Close() }()

	info, _ := bundleFile.Stat()
	logger.Step(p.key, p.task.Comment, "☁️",
		_color.New(_color.FgCyan).Sprintf("[%s]", u.Storage),
		_color.New(_color.FgMagenta).Sprintf("%s → %s (%s)", u.Source, objectKey, unit.ByteSize(info.Size())),
	)

	ctx := context.Background()
	if err = store.Upload(ctx, objectKey, bundleFile, info.Size()); err != nil {
		return fmt.Errorf("upload to storage error: %s", err)
	}

	// Track for deploy via
	p.uploadedKeys[sourceName] = objectKey
	_ = sha // sha available for logging

	logger.Step(p.key, p.task.Comment, "✅",
		_color.New(_color.FgHiGreen).Sprint("uploaded"),
	)
	return nil
}

func (p *TaskRunner) deploy(deploy *protocol.Deploy) (err error) {
	servers := map[string]*protocol.Server{}
	runners := make([]*ServerRunner, 0)
	defer func() {
		for _, v := range runners {
			v.Close()
		}
	}()
	for _, v := range deploy.Servers {
		if v.Use != "" {
			server, ok := p.conf.Servers[v.Use]
			if !ok {
				err = fmt.Errorf("deploy use server: '%s' not exists", v.Use)
				return
			}
			servers[server.Host] = server
		} else {
			servers[v.Host] = v
		}
	}

	// Check if deploying via cloud storage
	if deploy.Via != "" {
		return p.deployViaStorage(deploy, servers)
	}

	// Original SFTP upload path
	for key, server := range servers {
		if server.Host == "" {
			err = fmt.Errorf("deploy server host is empty")
			return
		}
		serverRunner := NewServerRunner(p.conf, p, server, key)
		runners = append(runners, serverRunner)
		for _, mapper := range deploy.Mappers {
			err = serverRunner.Upload(mapper.Source, mapper.Target)
			if err != nil {
				return
			}
		}
	}
	for key, server := range servers {
		serverRunner := NewServerRunner(p.conf, p, server, key)
		runners = append(runners, serverRunner)
		for _, execute := range deploy.Executes {
			if execute.Run != "" {
				p.printServerExec(server, execute.Run)
				err = serverRunner.PipeExec(execute.Run)
				if err != nil {
					err = fmt.Errorf("server execute error: %s", err)
					return
				}
			}
		}
	}

	return
}

// deployViaStorage downloads artifacts from cloud storage on the remote server.
func (p *TaskRunner) deployViaStorage(deploy *protocol.Deploy, servers map[string]*protocol.Server) error {
	cfg := config.Load()
	cred, err := cfg.DecryptStorage(deploy.Via)
	if err != nil {
		return fmt.Errorf("storage '%s': %s", deploy.Via, err)
	}

	store, err := storage.NewFromCredential(cred)
	if err != nil {
		return fmt.Errorf("create storage client error: %s", err)
	}

	runners := make([]*ServerRunner, 0)
	defer func() {
		for _, v := range runners {
			v.Close()
		}
	}()

	ctx := context.Background()

	for key, server := range servers {
		if server.Host == "" {
			return fmt.Errorf("deploy server host is empty")
		}
		serverRunner := NewServerRunner(p.conf, p, server, key)
		runners = append(runners, serverRunner)

		for _, mapper := range deploy.Mappers {
			sourceName := path.Base(mapper.Source)
			objectKey, ok := p.uploadedKeys[sourceName]
			if !ok {
				// Fallback: construct key from convention
				cwd, _ := os.Getwd()
				cwdHash := fmt.Sprintf("%x", sha1.Sum([]byte(cwd)))[:8]
				objectKey = fmt.Sprintf("nest/%s/%s.tar.gz", cwdHash, sourceName)
			}

			// Generate pre-signed URL (1 hour)
			url, err := store.PresignedURL(ctx, objectKey, 1*time.Hour)
			if err != nil {
				return fmt.Errorf("generate download URL error: %s", err)
			}

			logger.Step(p.key, p.task.Comment, "⬇️",
				_color.New(_color.FgCyan).Sprintf("[%s]", server.Name()),
				_color.New(_color.FgMagenta).Sprintf("downloading %s via %s", sourceName, deploy.Via),
			)

			// Prepare target directory and download on remote server
			targetDir := mapper.Target
			bundleName := fmt.Sprintf("%s.tar.gz", sourceName)
			bundleRemotePath := fmt.Sprintf("/tmp/nest-%s", bundleName)

			// Download via curl on remote server
			downloadCmd := fmt.Sprintf("curl -fsSL '%s' -o %s", url, bundleRemotePath)
			if err = serverRunner.PipeExec(downloadCmd); err != nil {
				return fmt.Errorf("remote download error: %s", err)
			}

			// Extract and deploy using the same flow as SFTP
			extractCmd := fmt.Sprintf("mkdir -p %s && tar -xzf %s -C %s && rm -f %s",
				targetDir, bundleRemotePath, targetDir, bundleRemotePath)
			if err = serverRunner.PipeExec(extractCmd); err != nil {
				return fmt.Errorf("remote extract error: %s", err)
			}

			logger.Step(p.key, p.task.Comment, "✅",
				_color.New(_color.FgHiGreen).Sprintf("deployed %s → %s", sourceName, targetDir),
			)
		}
	}

	// Execute post-deploy commands
	for key, server := range servers {
		serverRunner := NewServerRunner(p.conf, p, server, key)
		runners = append(runners, serverRunner)
		for _, execute := range deploy.Executes {
			if execute.Run != "" {
				p.printServerExec(server, execute.Run)
				if err := serverRunner.PipeExec(execute.Run); err != nil {
					return fmt.Errorf("server execute error: %s", err)
				}
			}
		}
	}

	return nil
}

func (p TaskRunner) execute(step *protocol.Step) (err error) {
	if step.Run == "" {
		return
	}
	p.printExec(step.Run)
	runner := _exec.NewRunner()
	runner.AddCommand(step.Run)
	runner.SetEnviron(p.prepareEnviron())
	if p.task.Workspace != "" {
		runner.SetDir(p.task.Workspace)
	}
	err = runner.PipeOutput()
	if err != nil {
		err = fmt.Errorf("runner pipe exec error: %s", err)
		return
	}
	return
}

func (p TaskRunner) PrintStart() {
	logger.Step(p.key, p.task.Comment, "🕘", "start")
}

func (p TaskRunner) PrintSuccess() {
	logger.Step(p.key, p.task.Comment, "🎉", _color.New(_color.FgHiGreen).Sprint("success"))
}

func (p TaskRunner) PrintFailed() {
	logger.Step(p.key, p.task.Comment, "❌️", _color.New(_color.FgHiRed).Sprint("failed"))
}

func (p TaskRunner) printExecUseStart(key, comment string) {
	if comment != "" {
		key = comment
	}
	logger.Step(p.key, p.task.Comment, "👉", key)
}

func (p TaskRunner) printExecUseEnd() {
	logger.Step(p.key, p.task.Comment, "👈")
}

func (p TaskRunner) printServerExec(server *protocol.Server, command string) {
	logger.Step(
		p.key,
		p.task.Comment,
		"🏃",
		_color.New(_color.FgCyan).Sprintf("[%s]", server.Name()),
		_color.New(_color.FgWhite).Sprintf("%s", command),
	)
}

func (p TaskRunner) printExec(command string) {
	logger.Step(
		p.key,
		p.task.Comment,
		"🏃",
		_color.New(_color.FgWhite).Sprintf("%s", command),
	)
}
