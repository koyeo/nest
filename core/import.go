package core

import (
	"fmt"
	"nest/enums"
	"nest/storage"
	"os"
	"regexp"
	"strings"
)

func GoPackages(src []byte) (packages []string) {

	r1 := regexp.MustCompile(`import\s+\(([\s\S]*?)\)`)
	r2 := regexp.MustCompile(`import\s+([A-Za-z0-9-._/"]*\s*"[A-Za-z0-9-._/"]+")`)
	r3 := regexp.MustCompile(`([A-Za-z0-9-._/"]*)\s*"([A-Za-z0-9-._/"]+)"`)

	blocks := make([]string, 0)

	res := r1.FindAllSubmatch(src, -1)
	for _, v1 := range res {
		if len(v1) < 0 {
			continue
		}

		for k2, v2 := range v1 {
			if k2 > 0 {
				block := strings.TrimSpace(string(v2))
				blocks = append(blocks, block)
			}

		}
	}

	res = r2.FindAllSubmatch(src, -1)
	for _, v1 := range res {
		if len(v1) < 0 {
			continue
		}

		for k2, v2 := range v1 {
			if k2 > 0 {
				block := strings.TrimSpace(string(v2))
				blocks = append(blocks, block)
			}
		}
	}

	packages = make([]string, 0)

	for _, block := range blocks {
		items := strings.Split(block, "\n")
		for _, v := range items {
			v = strings.TrimSpace(v)
			strs := r3.FindAllStringSubmatch(v, -1)
			for _, v1 := range strs {
				if len(v1) == 2 {
					packages = append(packages, strings.TrimSpace(v1[1]))
				}
				if len(v1) == 3 {
					packages = append(packages, strings.TrimSpace(v1[2]))
				}
			}
		}
	}

	return
}

func IsProjectPackage(packagePath string) (root string, ok bool, err error) {

	wd, err := os.Getwd()
	if err != nil {
		return
	}

	paths := strings.Split(wd, "/")
	l := len(paths)
	if l == 0 {
		err = fmt.Errorf("get project root error")
		return
	}

	root = paths[l-1]

	if strings.HasPrefix(packagePath, root) {
		ok = true
	}

	return
}

func GoPackageFiles(watchFiles *WatchFiles, filePath string) (err error) {

	if !storage.Exist(filePath) {
		return
	}

	data, err := storage.Read(filePath)
	if err != nil {
		return
	}

	packagePaths := GoPackages(data)
	var ok bool
	var root string
	for _, v := range packagePaths {
		root, ok, err = IsProjectPackage(v)
		if err != nil {
			return
		}
		if !ok {
			continue
		}
		path := strings.TrimPrefix(v, root+"/")
		if storage.Exist(path) {
			var files []string
			files, err = storage.Files(path, enums.GoSuffix)
			if err != nil {
				return
			}
			for _, vv := range files {
				if !watchFiles.Has(vv) {
					watchFiles.Add(vv)
					err = GoPackageFiles(watchFiles, vv)
					if err != nil {
						return
					}
				}
			}
		}
	}

	return
}
