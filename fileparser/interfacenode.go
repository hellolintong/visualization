package fileparser

import (
	"bytes"
	"fmt"
)

type InterfaceNode struct {
	fileNode        *FileNode
	name            string
	implementStruct map[string]bool
	methods         map[string]string
}

func NewInterfaceNode(fileNode *FileNode, name string, methods map[string]string) *InterfaceNode {
	return &InterfaceNode{
		fileNode:        fileNode,
		name:            name,
		implementStruct: make(map[string]bool, 0),
		methods:         methods,
	}
}

func (i *InterfaceNode) getIdentity() string {
	return "\"" + i.fileNode.packageName + "/" + i.name + "\""
}

func (i *InterfaceNode) mergeImplement(functionNames map[string]map[string]bool) {
label:
	for receiver, functions := range functionNames {
		// 查看接口中的全部方法是否都实现了
		for method, _ := range i.methods {
			if _, ok := functions[method]; !ok {
				continue label
			}
		}
		i.implementStruct[receiver] = true
	}
}

func (i *InterfaceNode) getLabelDescribe() string {
	buffer := bytes.NewBuffer([]byte{})
	for _, describe := range i.methods {
		buffer.WriteString(fmt.Sprintf("%s\\l\\n", describe))
	}
	label := fmt.Sprintf("interface: %s\\l\\n----\\lpackage: %s\\l\\nfile: %s\\l-----\\l%s", i.name, i.fileNode.packageName, i.fileNode.fileNodeTagName, buffer.String())
	return label
}

func (i *InterfaceNode) DrawNode(content *bytes.Buffer, record map[string]bool) {
	if record[i.getIdentity()] == true {
		return
	}
	content.WriteString(fmt.Sprintf("%s [label=\"%s\", shape=\"box\"];", i.getIdentity(), i.getLabelDescribe()))
	content.WriteString("\n")
	record[i.getIdentity()] = true
}

