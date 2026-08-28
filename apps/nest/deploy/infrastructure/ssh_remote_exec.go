package infrastructure

import "github.com/koyeo/nest/execer"

// SSHRemoteExec implements domain.RemoteExec using SSH.
type SSHRemoteExec struct {
	server *execer.Server
}

// NewSSHRemoteExec creates a new SSHRemoteExec.
func NewSSHRemoteExec(server *execer.Server) *SSHRemoteExec {
	return &SSHRemoteExec{server: server}
}

func (r *SSHRemoteExec) Exec(command string) error {
	return r.server.CombinedExec(command)
}

func (r *SSHRemoteExec) ExecPipe(command string) error {
	return r.server.PipeExec(command)
}
