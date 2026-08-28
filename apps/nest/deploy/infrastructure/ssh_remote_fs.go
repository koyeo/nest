package infrastructure

import (
	"fmt"
	"io/fs"
	"strings"
	"time"

	"github.com/koyeo/nest/deploy/domain"
	"github.com/koyeo/nest/execer"
)

// sftpFileInfo wraps os.FileInfo to implement domain.FileInfo.
type sftpFileInfo struct {
	name    string
	size    int64
	mode    fs.FileMode
	modTime time.Time
	isDir   bool
}

func (f *sftpFileInfo) Name() string       { return f.name }
func (f *sftpFileInfo) Size() int64        { return f.size }
func (f *sftpFileInfo) Mode() fs.FileMode  { return f.mode }
func (f *sftpFileInfo) ModTime() time.Time { return f.modTime }
func (f *sftpFileInfo) IsDir() bool        { return f.isDir }

// SSHRemoteFS implements domain.RemoteFS using SFTP and SSH.
type SSHRemoteFS struct {
	server *execer.Server
}

// NewSSHRemoteFS creates a new SSHRemoteFS with the given exec server.
func NewSSHRemoteFS(server *execer.Server) *SSHRemoteFS {
	return &SSHRemoteFS{server: server}
}

func (r *SSHRemoteFS) Stat(path string) (domain.FileInfo, error) {
	info, err := r.server.SFTPClient().Stat(path)
	if err != nil {
		return nil, err
	}
	return &sftpFileInfo{
		name:    info.Name(),
		size:    info.Size(),
		mode:    info.Mode(),
		modTime: info.ModTime(),
		isDir:   info.IsDir(),
	}, nil
}

func (r *SSHRemoteFS) MkdirAll(path string) error {
	return r.server.SFTPClient().MkdirAll(path)
}

func (r *SSHRemoteFS) ReadDir(path string) ([]string, error) {
	entries, err := r.server.SFTPClient().ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("read remote dir error: %s", err)
	}
	var names []string
	for _, e := range entries {
		name := e.Name()
		// Skip .nest metadata directory
		if strings.HasPrefix(name, ".nest") {
			continue
		}
		names = append(names, name)
	}
	return names, nil
}

func (r *SSHRemoteFS) ReadFile(path string) ([]byte, error) {
	file, err := r.server.SFTPClient().Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	data := make([]byte, info.Size())
	_, err = file.Read(data)
	if err != nil && err.Error() != "EOF" {
		return nil, err
	}
	return data, nil
}

func (r *SSHRemoteFS) WriteFile(path string, data []byte) error {
	file, err := r.server.SFTPClient().Create(path)
	if err != nil {
		return fmt.Errorf("create remote file error: %s", err)
	}
	defer func() { _ = file.Close() }()

	_, err = file.Write(data)
	return err
}

func (r *SSHRemoteFS) Remove(path string) error {
	// SFTP Remove only handles files; use SSH rm -rf for directories
	info, err := r.server.SFTPClient().Stat(path)
	if err != nil {
		return nil // Already gone
	}
	if info.IsDir() {
		// Use SSH command for recursive directory removal
		session, err := r.server.SSHClient().NewSession()
		if err != nil {
			return err
		}
		defer func() { _ = session.Close() }()
		return session.Run(fmt.Sprintf("rm -rf %s", path))
	}
	return r.server.SFTPClient().Remove(path)
}

func (r *SSHRemoteFS) Rename(src, dst string) error {
	return r.server.SFTPClient().Rename(src, dst)
}

func (r *SSHRemoteFS) FileHash(path string) (string, error) {
	// Try sha256sum (Linux), fallback to shasum (macOS)
	cmd := fmt.Sprintf("sha256sum %s 2>/dev/null || shasum -a 256 %s", path, path)
	session, err := r.server.SSHClient().NewSession()
	if err != nil {
		return "", fmt.Errorf("create session error: %s", err)
	}
	defer func() { _ = session.Close() }()

	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return "", fmt.Errorf("compute hash error: %s", err)
	}

	parts := strings.Fields(strings.TrimSpace(string(output)))
	if len(parts) == 0 {
		return "", fmt.Errorf("unexpected hash output: %s", string(output))
	}
	return parts[0], nil
}

func (r *SSHRemoteFS) FileModTime(path string) (time.Time, error) {
	info, err := r.server.SFTPClient().Stat(path)
	if err != nil {
		return time.Time{}, fmt.Errorf("stat remote file error: %s", err)
	}
	return info.ModTime().UTC(), nil
}
