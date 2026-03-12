package execer

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func NewRunner() *Runner {
	return &Runner{
		shell:  "bash",
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
}

type Runner struct {
	commands []string
	shell    string
	dir      string
	environ  []string
	stdout   io.Writer
	stderr   io.Writer
	ctx      context.Context
}

// SetContext sets a context for cancellation support.
func (p *Runner) SetContext(ctx context.Context) *Runner {
	p.ctx = ctx
	return p
}

// AddCommand sets execution commands.
func (p *Runner) AddCommand(commands ...string) *Runner {
	p.commands = commands
	return p
}

// SetEnviron sets environment variables.
func (p *Runner) SetEnviron(environ []string) *Runner {
	p.environ = environ
	return p
}

// SetDir sets working directory.
func (p *Runner) SetDir(path string) *Runner {
	p.dir = path
	return p
}

// SetShell sets the shell binary.
func (p *Runner) SetShell(path string) *Runner {
	p.shell = path
	return p
}

// SetOutput sets stdout and stderr writers. Defaults to os.Stdout/os.Stderr.
func (p *Runner) SetOutput(stdout, stderr io.Writer) *Runner {
	p.stdout = stdout
	p.stderr = stderr
	return p
}

// CombinedOutput returns final result as string.
func (p *Runner) CombinedOutput() (result string, err error) {
	var res []byte
	for _, v := range p.commands {
		c := exec.Command(p.shell, "-c", v)
		c.Dir = p.dir
		res, err = c.CombinedOutput()
		if err != nil {
			return
		}
		result += strings.TrimSpace(string(res))
	}
	return
}

// PipeOutput streams output in real time to the configured writers.
func (p *Runner) PipeOutput() error {
	for _, v := range p.commands {
		if err := p.pipeExec(v); err != nil {
			return err
		}
	}
	return nil
}

func (p *Runner) wrapCmd(cmd *exec.Cmd) {
	cmd.Dir = p.dir
	if len(p.environ) > 0 {
		cmd.Env = p.environ
	} else {
		cmd.Env = os.Environ()
	}
}

func (p *Runner) pipeExec(command string) (err error) {
	var c *exec.Cmd
	if p.ctx != nil {
		c = exec.CommandContext(p.ctx, p.shell, "-c", command)
	} else {
		c = exec.Command(p.shell, "-c", command)
	}
	p.wrapCmd(c)

	// Only connect stdin when NOT in TUI mode (TUI owns the terminal)
	if p.stdout == os.Stdout {
		c.Stdin = os.Stdin
	}

	// Let Go's exec package handle the pipe copying internally (32KB buffer)
	c.Stdout = p.stdout

	// Capture stderr tail for error reporting
	stderrTail := newTailBuffer(4096)
	c.Stderr = io.MultiWriter(p.stderr, stderrTail)

	err = c.Run()
	if err != nil {
		tail := strings.TrimSpace(stderrTail.String())
		if tail != "" {
			err = fmt.Errorf("%s\n%s", err, tail)
		}
	}
	return
}
