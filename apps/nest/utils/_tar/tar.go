package _tar

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
)

func Compress(files []*os.File, dest string) error {
	d, _ := os.Create(dest)
	defer d.Close()
	gw := gzip.NewWriter(d)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()
	for _, file := range files {
		err := compress(file, "", tw)
		if err != nil {
			return err
		}
	}
	return nil
}

func compress(file *os.File, prefix string, tw *tar.Writer) error {
	info, err := file.Stat()
	if err != nil {
		return err
	}
	if info.IsDir() {
		if prefix == "" {
			prefix = info.Name()
		} else {
			prefix = prefix + "/" + info.Name()
		}
		var fileInfos []os.FileInfo
		fileInfos, err = file.Readdir(-1)
		if err != nil {
			return err
		}
		for _, fi := range fileInfos {
			childPath := file.Name() + "/" + fi.Name()
			// Check for symlinks via Lstat
			linfo, lErr := os.Lstat(childPath)
			if lErr != nil {
				continue // skip unreadable entries
			}
			if linfo.Mode()&os.ModeSymlink != 0 {
				// Try to resolve — skip if broken
				if _, resolveErr := os.Stat(childPath); resolveErr != nil {
					fmt.Fprintf(os.Stderr, "⚠ skipping broken symlink: %s\n", childPath)
					continue
				}
			}
			var f *os.File
			f, err = os.Open(childPath)
			if err != nil {
				return err
			}
			err = compress(f, prefix, tw)
			if err != nil {
				return err
			}
		}
	} else {
		var header *tar.Header
		header, err = tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		if prefix != "" {
			header.Name = prefix + "/" + header.Name
		}
		err = tw.WriteHeader(header)
		if err != nil {
			return err
		}
		_, err = io.Copy(tw, file)
		file.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

