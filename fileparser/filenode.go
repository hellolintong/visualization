package fileparser

import (
	"bytes"
	"path/filepath"
	"strings"
)
import "fmt"

type FileNode struct {
	nodeManager     *NodeManager
	file            string
	fileNodeName    string
	fileNodeTagName string
	packageName     string
	structNodes     map[string]*StructNode
	interfaceNodes  map[string]*InterfaceNode
	functionNodes   map[string][]*FunctionNode
	importers       map[string]string
}

func (f *FileNode) MergeFunction() {
	// 先收集所有用到函数
	for _, functionNodes := range f.functionNodes {
		for _, functionNode := range functionNodes {
			functionNode.Merge()
		}
	}
}

func (f *FileNode) MergeStruct(structTypes map[string]map[string]*StructNode, interfaceNames map[string]map[string]*InterfaceNode) {
	for _, structNode := range f.structNodes {
		structNode.Merge(structTypes, interfaceNames)
	}
}

func NewFileNode(nodeManager *NodeManager, file string, packageName string) *FileNode {
	return &FileNode{
		nodeManager:     nodeManager,
		packageName:     packageName,
		file:            file,
		fileNodeName:    strings.ReplaceAll(filepath.Base(file), ".", "_"),
		fileNodeTagName: filepath.Base(file),
		structNodes:     make(map[string]*StructNode, 0),
		interfaceNodes:  make(map[string]*InterfaceNode, 0),
		functionNodes:   make(map[string][]*FunctionNode, 0),
		importers:       make(map[string]string, 0),
	}
}

func (f *FileNode) String() string {
	buffer := bytes.NewBuffer([]byte{})
	buffer.WriteString(fmt.Sprintln("filename:" + f.file))
	for _, fileStruct := range f.structNodes {
		for key, value := range fileStruct.fields {
			buffer.WriteString(fmt.Sprintln("struct:" + key + ", type:" + value))
		}
	}
	return buffer.String()
}
