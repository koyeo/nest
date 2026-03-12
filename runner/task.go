package runner

import (
	"context"
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/gozelle/_fs"
	"github.com/koyeo/nest/config"
	"github.com/koyeo/nest/execer"
	"github.com/koyeo/nest/logger"
	"github.com/koyeo/nest/protocol"
	"github.com/koyeo/nest/storage"
	"github.com/koyeo/nest/utils/_tar"
	"github.com/koyeo/nest/utils/unit"
)

// StepEventHandler receives task execution events for visualization.
type StepEventHandler interface {
	OnStepStart(index int, name string)
	OnStepDone(index int, err error)
	OnOutput(content string)
	OnTaskDone(err error)
	// Prompt shows a message to the user and returns their text input.
	// Used for interactive prompts (e.g., conflict resolution during deploy).
	Prompt(message string) string
}

func NewTaskRunner(conf *protocol.Config, task *protocol.Task, key string) *TaskRunner {
	return &TaskRunner{conf: conf, task: task, key: key, uploadedKeys: map[string]string{}}
}

type TaskRunner struct {
	key          string
	conf         *protocol.Config
	task         *protocol.Task
	parents      map[string]bool
	uploadedKeys map[string]string // source basename → object key
	handler      StepEventHandler  // event handler (nil = legacy output)
	stepOffset   int               // global step index offset
	ctx          context.Context
}

// SetContext sets a context for cancellation support.
func (p *TaskRunner) SetContext(ctx context.Context) {
	p.ctx = ctx
}

// SetEventHandler sets the event handler for visualization.
func (p *TaskRunner) SetEventHandler(h StepEventHandler) {
	p.handler = h
}

// outputWriter returns an io.Writer. When handler is set, output goes nowhere
// (it’s captured at the OS level by the webui). Otherwise, os.Stdout.
func (p *TaskRunner) outputWriter() io.Writer {
	return os.Stdout
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

// StepNames returns the display names of all steps in the task, recursively expanding use references.
func (p TaskRunner) StepNames() []string {
	return p.collectStepNames(p.task, 0)
}

// StepDetail holds structured info about a step for the webui.
type StepDetail struct {
	Name    string `json:"name"`
	Depth   int    `json:"depth"`
	IsGroup bool   `json:"is_group"` // true for "use" headers
}

// StepDetails returns structured step metadata for the webui tree view.
func (p TaskRunner) StepDetails() []StepDetail {
	return p.collectStepDetails(p.task, 0)
}

func (p TaskRunner) collectStepDetails(task *protocol.Task, depth int) []StepDetail {
	var details []StepDetail
	for _, step := range task.Steps {
		if step.Use != "" {
			comment := step.Use
			if t, ok := p.conf.Tasks[step.Use]; ok && t.Comment != "" {
				comment = t.Comment
			}
			details = append(details, StepDetail{Name: comment, Depth: depth, IsGroup: true})
			if subTask, ok := p.conf.Tasks[step.Use]; ok {
				details = append(details, p.collectStepDetails(subTask, depth+1)...)
			}
		} else if step.Upload != nil {
			details = append(details, StepDetail{Name: fmt.Sprintf("upload: %s", step.Upload.Source), Depth: depth})
		} else if step.Deploy != nil {
			details = append(details, StepDetail{Name: "deploy", Depth: depth})
		} else if step.Run != "" {
			details = append(details, StepDetail{Name: step.Run, Depth: depth})
		}
	}
	return details
}

func (p TaskRunner) collectStepNames(task *protocol.Task, depth int) []string {
	var names []string
	prefix := ""
	if depth > 0 {
		prefix = strings.Repeat("  ", depth-1) + "└ "
	}
	for _, step := range task.Steps {
		if step.Use != "" {
			// Add the "use" header
			names = append(names, prefix+fmt.Sprintf("▶ %s", step.Use))
			// Recursively expand the referenced task's steps
			if subTask, ok := p.conf.Tasks[step.Use]; ok {
				subNames := p.collectStepNames(subTask, depth+1)
				names = append(names, subNames...)
			}
		} else if step.Upload != nil {
			names = append(names, prefix+fmt.Sprintf("upload: %s", step.Upload.Source))
		} else if step.Deploy != nil {
			names = append(names, prefix+"deploy")
		} else if step.Run != "" {
			names = append(names, prefix+step.Run)
		}
	}
	return names
}

func (p TaskRunner) Exec() (err error) {
	globalIdx := p.stepOffset
	for _, step := range p.task.Steps {
		// Check for cancellation before each step
		if p.ctx != nil {
			select {
			case <-p.ctx.Done():
				return fmt.Errorf("cancelled")
			default:
			}
		}
		if step.Use != "" {
			p.sendStepStart(globalIdx)
			globalIdx++
			// The child task's steps occupy the next N indices
			childStepCount := p.countSubSteps(step.Use)
			if err = p.use(step.Use, globalIdx); err != nil {
				p.sendStepDone(err)
				return
			}
			p.sendStepDone(nil)
			globalIdx += childStepCount
		} else if step.Upload != nil {
			p.sendStepStart(globalIdx)
			if err = p.upload(step.Upload); err != nil {
				p.sendStepDone(err)
				return
			}
			p.sendStepDone(nil)
			globalIdx++
		} else if step.Deploy != nil {
			p.sendStepStart(globalIdx)
			if err = p.deploy(step.Deploy); err != nil {
				p.sendStepDone(err)
				return
			}
			p.sendStepDone(nil)
			globalIdx++
		} else {
			p.sendStepStart(globalIdx)
			if err = p.execute(step); err != nil {
				p.sendStepDone(err)
				return
			}
			p.sendStepDone(nil)
			globalIdx++
		}
	}
	return
}

// countSubSteps counts the total flattened steps for a use reference (recursive).
func (p TaskRunner) countSubSteps(key string) int {
	task, ok := p.conf.Tasks[key]
	if !ok {
		return 0
	}
	count := 0
	for _, step := range task.Steps {
		if step.Use != "" {
			count++ // The "use" header itself
			count += p.countSubSteps(step.Use)
		} else {
			count++
		}
	}
	return count
}

func (p TaskRunner) sendStepStart(index int) {
	if p.handler != nil {
		name := ""
		// We don't need to pass name here; the handler already knows step names
		p.handler.OnStepStart(index, name)
	}
}

func (p TaskRunner) sendStepDone(err error) {
	if p.handler != nil {
		p.handler.OnStepDone(p.stepOffset, err)
	}
}

// tuiLog routes a log message through handler output or falls back to logger.Step.
func (p TaskRunner) tuiLog(emoji string, args ...string) {
	if p.handler != nil {
		msg := emoji
		for _, a := range args {
			if a != "" {
				msg += " " + a
			}
		}
		p.handler.OnOutput(msg + "\n")
	} else {
		logger.Step(p.key, p.task.Comment, emoji, args...)
	}
}

func (p TaskRunner) use(key string, childOffset int) (err error) {

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

	// Propagate handler, context, and step offset to child runner
	taskRunner.handler = p.handler
	taskRunner.ctx = p.ctx
	taskRunner.stepOffset = childOffset

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
	// Resolve storage alias to global config name
	globalName, err := p.conf.ResolveStorage(u.Storage)
	if err != nil {
		return err
	}

	cfg := config.Load()
	cred, err := cfg.DecryptStorage(globalName)
	if err != nil {
		return fmt.Errorf("storage '%s' (config '%s'): %s", u.Storage, globalName, err)
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
	tmpDir := NestTmpDir()
	bundlePath := fmt.Sprintf("%s/%s.tar.gz", tmpDir, sourceName)
	defer cleanNestTempDir()

	sourceFile, err := os.Open(u.Source)
	if err != nil {
		return fmt.Errorf("open source error: %s", err)
	}
	defer func() { _ = sourceFile.Close() }()

	if err = _tar.Compress([]*os.File{sourceFile}, bundlePath); err != nil {
		return fmt.Errorf("compress error: %s", err)
	}

	// Compute SHA1 of bundle content for object key
	bundleData, err := os.ReadFile(bundlePath)
	if err != nil {
		return fmt.Errorf("read bundle error: %s", err)
	}
	sha := fmt.Sprintf("%x", sha1.Sum(bundleData))
	objectKey := fmt.Sprintf("nest/%s.tar.gz", sha)

	// Get local bundle size
	bundleFile, err := os.Open(bundlePath)
	if err != nil {
		return fmt.Errorf("open bundle error: %s", err)
	}
	defer func() { _ = bundleFile.Close() }()

	info, _ := bundleFile.Stat()
	localSize := info.Size()

	p.tuiLog("☁️",
		fmt.Sprintf("[%s]", u.Storage),
		fmt.Sprintf("%s → %s (%s)", u.Source, objectKey, unit.ByteSize(localSize)),
	)

	ctx := context.Background()

	// Check if the same object already exists (dedup by content hash + size)
	remoteSize, err := store.Head(ctx, objectKey)
	if err != nil {
		return fmt.Errorf("check remote object error: %s", err)
	}

	if remoteSize == localSize {
		p.tuiLog("⏭️", "skipped (already exists)")
	} else {
		if err = store.Upload(ctx, objectKey, bundleFile, localSize); err != nil {
			return fmt.Errorf("upload to storage error: %s", err)
		}
		p.tuiLog("✅", "uploaded")
	}

	// Track for deploy via
	p.uploadedKeys[sourceName] = objectKey
	return nil
}

func (p *TaskRunner) deploy(deploy *protocol.Deploy) (err error) {
	servers := map[string]*protocol.Server{}
	runners := map[string]*ServerRunner{} // key → runner, reused for files + executes
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

	// Create one ServerRunner per server (reused for files + executes)
	for key, server := range servers {
		if server.Host == "" {
			err = fmt.Errorf("deploy server host is empty")
			return
		}
		runners[key] = NewServerRunner(p.conf, p, server, key)
	}

	// Upload files
	for key, server := range servers {
		serverRunner := runners[key]
		_ = server // used for logging in deployFileViaStorage

		for _, file := range deploy.Files {
			if file.Storage != "" {
				// Transfer via cloud storage (upload → presigned URL → remote download)
				if err = p.deployFileViaStorage(serverRunner, server, file.Storage, file.Source, file.Target, deploy.ConflictStrategy); err != nil {
					return
				}
			} else {
				// Default: direct SFTP upload (tar + extract)
				err = serverRunner.Upload(file.Source, file.Target, deploy.ConflictStrategy)
				if err != nil {
					return
				}
			}
		}
	}

	// Clean up cloud storage objects after all servers have downloaded
	p.cleanUploadedObjects(deploy.Files)

	// Execute post-deploy commands (reuse same runners)
	for key, server := range servers {
		serverRunner := runners[key]
		for _, execute := range deploy.Executes {
			if execute.Run != "" {
				cmd := execute.Run
				if deploy.ShellInit != "" {
					cmd = deploy.ShellInit + " && " + cmd
				}
				if deploy.Cwd != "" {
					cmd = "cd " + deploy.Cwd + " && " + cmd
				}
				p.printServerExec(server, execute.Run)
				if err = serverRunner.PipeExec(cmd); err != nil {
					err = fmt.Errorf("server execute error: %s", err)
					return
				}
			}
		}
	}

	return
}

// cleanUploadedObjects deletes all cloud storage objects that were uploaded during this deploy.
// Called after all servers have finished downloading to avoid breaking multi-server deploys.
func (p *TaskRunner) cleanUploadedObjects(files []*protocol.FileMapping) {
	if len(p.uploadedKeys) == 0 {
		return
	}

	// Group object keys by storage alias
	type storageGroup struct {
		alias string
		keys  []string
	}
	groups := map[string]*storageGroup{}

	for _, file := range files {
		if file.Storage == "" {
			continue
		}
		objectKey, ok := p.uploadedKeys[file.Source]
		if !ok {
			continue
		}
		if g, exists := groups[file.Storage]; exists {
			// Dedup: only add if not already in list
			found := false
			for _, k := range g.keys {
				if k == objectKey {
					found = true
					break
				}
			}
			if !found {
				g.keys = append(g.keys, objectKey)
			}
		} else {
			groups[file.Storage] = &storageGroup{
				alias: file.Storage,
				keys:  []string{objectKey},
			}
		}
	}

	ctx := context.Background()
	for _, g := range groups {
		globalName, err := p.conf.ResolveStorage(g.alias)
		if err != nil {
			continue
		}
		cfg := config.Load()
		cred, err := cfg.DecryptStorage(globalName)
		if err != nil {
			continue
		}
		store, err := storage.NewFromCredential(cred)
		if err != nil {
			continue
		}
		if err = store.DeleteObjects(ctx, g.keys); err != nil {
			p.tuiLog("⚠️", fmt.Sprintf("failed to clean cloud objects: %s", err))
		} else {
			p.tuiLog("🧹", fmt.Sprintf("cleaned %d cloud object(s) from %s", len(g.keys), g.alias))
		}
	}
}

// deployFileViaStorage uploads a local file/dir to cloud storage, then downloads on remote.
// Uses the same DeployService.Deploy() flow as the SFTP path for extraction + conflict resolution.
func (p *TaskRunner) deployFileViaStorage(
	serverRunner *ServerRunner,
	server *protocol.Server,
	storageAlias, localPath, target, conflictStrategy string,
) error {
	// Resolve storage alias
	globalName, err := p.conf.ResolveStorage(storageAlias)
	if err != nil {
		return err
	}

	cfg := config.Load()
	cred, err := cfg.DecryptStorage(globalName)
	if err != nil {
		return fmt.Errorf("storage '%s' (config '%s'): %s", storageAlias, globalName, err)
	}

	store, err := storage.NewFromCredential(cred)
	if err != nil {
		return fmt.Errorf("create storage client error: %s", err)
	}

	// Upload local source to storage (compress + dedup)
	objectKey, bundleHash, err := p.uploadToStorage(store, storageAlias, localPath)
	if err != nil {
		return err
	}

	// Generate pre-signed URL
	ctx := context.Background()
	url, err := store.PresignedURL(ctx, objectKey, 1*time.Hour)
	if err != nil {
		return fmt.Errorf("generate download URL error: %s", err)
	}

	sourceName := path.Base(localPath)
	p.tuiLog("⬇️",
		fmt.Sprintf("[%s]", server.Name()),
		fmt.Sprintf("downloading %s via %s", sourceName, storageAlias),
	)

	// Download tar.gz on remote server
	bundleName := fmt.Sprintf("%s.tar.gz", sourceName)
	bundleRemotePath := fmt.Sprintf("/tmp/nest-%s", bundleName)

	downloadCmd := fmt.Sprintf("curl -fsSL '%s' -o %s", url, bundleRemotePath)
	if err = serverRunner.PipeExec(downloadCmd); err != nil {
		return fmt.Errorf("remote download error: %s", err)
	}

	// Compute targetDir — same logic as SFTP path:
	// trailing "/" means target is a directory; otherwise parent is the dir.
	targetDir := target
	if !strings.HasSuffix(target, "/") {
		targetDir = filepath.Join(target, "../")
	}

	// Deploy using shared service (extract → conflict resolution → move)
	if err = serverRunner.deployBundle(bundleRemotePath, targetDir, bundleName, bundleHash, conflictStrategy); err != nil {
		return err
	}

	// Clean up remote temp bundle
	_ = serverRunner.CombinedExec(fmt.Sprintf("rm -f %s", bundleRemotePath))

	p.tuiLog("✅", fmt.Sprintf("deployed %s → %s", sourceName, targetDir))
	return nil
}

// uploadToStorage compresses and uploads a local path to cloud storage, with dedup.
func (p *TaskRunner) uploadToStorage(store storage.ObjectStorage, alias, localPath string) (string, string, error) {
	sourceName := path.Base(localPath)

	// Check if already uploaded in this run (dedup across servers)
	if key, ok := p.uploadedKeys[localPath]; ok {
		p.tuiLog("⏭️", fmt.Sprintf("[%s] %s already uploaded", alias, sourceName))
		return key, "", nil
	}

	ok, err := _fs.Exists(localPath)
	if err != nil {
		return "", "", fmt.Errorf("check source error: %s", err)
	}
	if !ok {
		return "", "", fmt.Errorf("source not found: %s", localPath)
	}

	// Compress
	tmpDir := NestTmpDir()
	bundlePath := fmt.Sprintf("%s/%s.tar.gz", tmpDir, sourceName)
	defer cleanNestTempDir()

	sourceFile, err := os.Open(localPath)
	if err != nil {
		return "", "", fmt.Errorf("open source error: %s", err)
	}
	defer func() { _ = sourceFile.Close() }()

	if err = _tar.Compress([]*os.File{sourceFile}, bundlePath); err != nil {
		return "", "", fmt.Errorf("compress error: %s", err)
	}

	// Compute hash for dedup
	bundleData, err := os.ReadFile(bundlePath)
	if err != nil {
		return "", "", fmt.Errorf("read bundle error: %s", err)
	}
	sha := fmt.Sprintf("%x", sha1.Sum(bundleData))
	objectKey := fmt.Sprintf("nest/%s.tar.gz", sha)
	bundleHashSum := sha256.Sum256(bundleData)
	bundleHash := fmt.Sprintf("%x", bundleHashSum[:])

	bundleFile, err := os.Open(bundlePath)
	if err != nil {
		return "", "", fmt.Errorf("open bundle error: %s", err)
	}
	defer func() { _ = bundleFile.Close() }()

	info, _ := bundleFile.Stat()
	localSize := info.Size()

	p.tuiLog("☁️",
		fmt.Sprintf("[%s]", alias),
		fmt.Sprintf("%s → %s (%s)", localPath, objectKey, unit.ByteSize(localSize)),
	)

	ctx := context.Background()
	remoteSize, err := store.Head(ctx, objectKey)
	if err != nil {
		return "", "", fmt.Errorf("check remote object error: %s", err)
	}

	if remoteSize == localSize {
		p.tuiLog("⏭️", "skipped (already exists)")
	} else {
		if err = store.Upload(ctx, objectKey, bundleFile, localSize); err != nil {
			return "", "", fmt.Errorf("upload to storage error: %s", err)
		}
		p.tuiLog("✅", "uploaded")
	}

	// Track for dedup across multiple servers
	p.uploadedKeys[localPath] = objectKey
	return objectKey, bundleHash, nil
}

func (p TaskRunner) execute(step *protocol.Step) (err error) {
	if step.Run == "" {
		return
	}
	p.printExec(step.Run)
	runner := execer.NewRunner()
	if p.ctx != nil {
		runner.SetContext(p.ctx)
	}
	runner.AddCommand(step.Run)
	runner.SetEnviron(p.prepareEnviron())
	if p.task.Workspace != "" {
		runner.SetDir(p.task.Workspace)
	}
	err = runner.PipeOutput()
	if err != nil {
		// If cancelled, return clean error
		if p.ctx != nil && p.ctx.Err() != nil {
			return fmt.Errorf("cancelled")
		}
		err = fmt.Errorf("runner pipe exec error: %s", err)
		return
	}
	return
}

func (p TaskRunner) PrintStart() {
	p.tuiLog("🕘", "start")
}

func (p TaskRunner) PrintSuccess() {
	p.tuiLog("🎉", "success")
}

func (p TaskRunner) PrintFailed() {
	p.tuiLog("❌️", "failed")
}

func (p TaskRunner) printExecUseStart(key, comment string) {
	if comment != "" {
		key = comment
	}
	p.tuiLog("👉", key)
}

func (p TaskRunner) printExecUseEnd() {
	p.tuiLog("👈")
}

func (p TaskRunner) printServerExec(server *protocol.Server, command string) {
	p.tuiLog(
		"🏃",
		fmt.Sprintf("[%s]", server.Name()),
		command,
	)
}

func (p TaskRunner) printExec(command string) {
	p.tuiLog("🏃", command)
}
