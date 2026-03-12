package execer

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type Server struct {
	Key          string
	Host         string
	Port         int
	User         string
	Password     string
	IdentityFile string
	sshClient    *ssh.Client
	sftpClient   *sftp.Client
	stdout       io.Writer
	stderr       io.Writer
	ctx          context.Context
}

// SetContext sets a context for cancellation support.
func (p *Server) SetContext(ctx context.Context) {
	p.ctx = ctx
}

func (p *Server) SSHClient() *ssh.Client {
	return p.sshClient
}

func (p *Server) SFTPClient() *sftp.Client {
	return p.sftpClient
}

// SetOutput sets stdout and stderr writers for piped command execution.
func (p *Server) SetOutput(stdout, stderr io.Writer) {
	p.stdout = stdout
	p.stderr = stderr
}

func (p *Server) getStdout() io.Writer {
	if p.stdout != nil {
		return p.stdout
	}
	return os.Stdout
}

func (p *Server) getStderr() io.Writer {
	if p.stderr != nil {
		return p.stderr
	}
	return os.Stderr
}

func (p *Server) Close() {
	if p.sftpClient != nil {
		_ = p.sftpClient.Close()
	}
	if p.sshClient != nil {
		_ = p.sshClient.Close()
	}
}

func (p Server) getPublicKey(path string) (auth ssh.AuthMethod, err error) {
	home, err := HomeDir()
	if err != nil {
		return
	}
	if strings.HasPrefix(path, "~") {
		path = fmt.Sprintf("%s%s", home, strings.TrimPrefix(path, "~"))
	}
	key, err := os.ReadFile(path)
	if err != nil {
		return
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return
	}
	auth = ssh.PublicKeys(signer)
	return
}

func (p *Server) InitSSH() (err error) {
	if p.sshClient == nil {
		p.sshClient, err = NewSSHClient(p)
		if err != nil {
			return
		}
	}
	return
}

func (p *Server) InitSFTP() (err error) {
	err = p.InitSSH()
	if err != nil {
		err = fmt.Errorf("connect server error: %s", err)
		return
	}
	if p.sftpClient == nil {
		p.sftpClient, err = sftp.NewClient(p.sshClient)
		if err != nil {
			err = fmt.Errorf("init sftp client error: %s", err)
			return
		}
	}
	return
}

func (p *Server) Ping() (err error) {
	err = p.InitSSH()
	if err != nil {
		return
	}
	session, err := p.SSHClient().NewSession()
	if err != nil {
		return
	}
	defer func() {
		_ = session.Close()
	}()
	result, err := session.CombinedOutput(`echo "pong"`)
	if err != nil {
		err = fmt.Errorf("ping error: %s", err)
		return
	}
	ok := strings.TrimSpace(string(result))
	if ok != "pong" {
		err = fmt.Errorf("ping error: server not response 'pong'")
		return
	}
	return
}

func (p *Server) CombinedExec(command string) error {
	session, err := p.SSHClient().NewSession()
	if err != nil {
		return err
	}
	defer func() {
		_ = session.Close()
	}()
	out, err := session.CombinedOutput(wrapLoginShell(command))
	if err != nil {
		err = fmt.Errorf("%s", strings.TrimSpace(string(out)))
		return err
	}
	return nil
}

func (p *Server) PipeExec(command string) (err error) {
	err = p.InitSSH()
	if err != nil {
		return
	}
	session, err := p.SSHClient().NewSession()
	if err != nil {
		return
	}
	defer func() {
		_ = session.Close()
	}()

	// Watch for context cancellation — close session to interrupt remote command
	if p.ctx != nil {
		go func() {
			<-p.ctx.Done()
			_ = session.Close()
		}()
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		err = fmt.Errorf("fetch stderr pipe error: %s", err)
		return
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		err = fmt.Errorf("fetch stdout pipe error: %s", err)
		return
	}

	stdoutW := p.getStdout()
	stderrW := p.getStderr()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		scanner := bufio.NewScanner(stdout)
		scanner.Split(bufio.ScanBytes)
		for scanner.Scan() {
			_, _ = stdoutW.Write(scanner.Bytes())
		}
		wg.Done()
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		scanner.Split(bufio.ScanBytes)
		for scanner.Scan() {
			_, _ = stderrW.Write(scanner.Bytes())
		}
		wg.Done()
	}()

	err = session.Run(wrapLoginShell(command))

	// Always wait for pipe goroutines to finish flushing output,
	// so that all stderr/stdout is visible before the error is reported.
	wg.Wait()

	// If context was cancelled, report a clean cancellation error
	if p.ctx != nil && p.ctx.Err() != nil {
		err = fmt.Errorf("cancelled")
		return
	}

	if err != nil {
		err = fmt.Errorf("session run command error: %s", err)
		return
	}

	return
}

// wrapLoginShell wraps a command to run in a login shell,
// so that /etc/profile, ~/.bash_profile, ~/.bashrc etc. are loaded.
// This ensures PATH includes tools like nvm, pyenv, etc.
func wrapLoginShell(command string) string {
	return fmt.Sprintf(`bash -l -c %q`, command)
}
