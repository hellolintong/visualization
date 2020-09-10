package visualization

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"visualization/fileparser"
)

const (
	maxDeep = 100
)

// 递归搜索，找出全部go文件, 并且忽略掉proto生成的文件
func visit(path string, deep int) ([]string, error) {
	if deep > maxDeep {
		return nil, fmt.Errorf("can't visit too deep dirs")
	}
	files := make([]string, 0)
	rd, err := ioutil.ReadDir(path)
	if err != nil {
		log.Printf("can't read dir:%s, error:%s", path, err.Error())
		return nil, err
	}
	for _, fi := range rd {
		subPath := filepath.Join(path, fi.Name())
		if fi.IsDir() {
			if fi.Name() != "vendor" {
				if subFiles, err := visit(subPath, deep+1); err == nil {
					files = append(files, subFiles...)
				}
			}
		} else {
			if strings.HasSuffix(subPath, ".go") && !strings.HasSuffix(subPath, ".pb.go") && !strings.HasSuffix(subPath, "_test.go") {
				files = append(files, subPath)
			}
		}
	}
	return files, nil
}


func NewParser(path string) (fileparser.Parser, error) {
	files, err := visit(path, 0)
	if err != nil {
		return nil, err
	}

	nodeManager := fileparser.NewParser(path)
	for _, file := range files {
		err := nodeManager.Inspect(file)
		if err != nil {
			log.Printf("can't inspect file:%s, error:%s", file, err.Error())
			os.Exit(-1)
		}
	}

	nodeManager.Merge()
	return nodeManager, nil
}

