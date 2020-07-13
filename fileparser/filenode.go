package fileparser

import (
	"bytes"
	"path/filepath"
	"strings"
)
import "fmt"

type FileNode struct {
	nodeManager *NodeManager
	file        string
	fileNodeName string
	fileNodeTagName string
	packageName string
	structNodes map[string]*StructNode
	functionNodes map[string]*FunctionNode
}

func (f *FileNode) MergeFunction(functionNames map[string][]string)  {
	for _, functionNode := range f.functionNodes {
		functionNode.Merge(functionNames)
	}
}

func (f *FileNode) MergeStruct(structTypes []string)  {
	for _, structNode := range f.structNodes {
		structNode.Merge(structTypes)
	}
}

func (f *FileNode) DrawFunctionNode(content *bytes.Buffer, receiver map[string]bool){
	if !f.checkFunctionComplex() {
		return
	}

	// function元素
	for _, functionNode := range f.functionNodes {
		functionNode.DrawNode(content, receiver)
	}
}

func (f *FileNode) DrawStructNode(content *bytes.Buffer, record map[string]bool){
	// 检查文件下的所有struct，如果都没有引用其他的struct，就直接跳过，避免图文件过大
	if !f.checkStructComplex() {
		return
	}
	// struct元素
	for _, structNode := range f.structNodes {
		structNode.DrawNode(content, record)
	}
}


func (f *FileNode) DrawFunctionRelation(content *bytes.Buffer, record map[string]bool){
	if !f.checkFunctionComplex() {
		return
	}

	for _, functionNode := range f.functionNodes {
		functionNode.DrawRelation(content, record)
	}
}

func (f *FileNode) DrawStructRelation(content *bytes.Buffer, record map[string]bool){
	if !f.checkStructComplex() {
		return
	}

	for _, structNode := range f.structNodes {
		structNode.DrawRelation(content, record)
	}
}

// 只处理有复杂关联元素的节点
func (f *FileNode) checkFunctionComplex() bool {
	emptyComplexNode := false
	for _, structNode := range f.functionNodes {
		if len(structNode.calledStructs) != 0 {
			emptyComplexNode = true
			break
		}
	}
	return emptyComplexNode
}

// 只处理有复杂关联元素的节点
func (f *FileNode) checkStructComplex() bool {
	emptyComplexNode := false
	for _, structNode := range f.structNodes {
		if len(structNode.complexFields) != 0 {
			emptyComplexNode = true
			break
		}
	}
	return emptyComplexNode
}

func NewFileNode(nodeManager *NodeManager, file string, packageName string) *FileNode {
	return &FileNode{
		nodeManager: nodeManager,
		packageName: packageName,
		file:        file,
		fileNodeName: strings.ReplaceAll(filepath.Base(file), ".", "_"),
		fileNodeTagName: filepath.Base(file),
		structNodes: make(map[string]*StructNode),
		functionNodes: make(map[string]*FunctionNode),
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
