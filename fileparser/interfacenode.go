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

func (i *InterfaceNode) getLabelDescribe(detail bool) string {
	if detail {
		buffer := bytes.NewBuffer([]byte{})
		for _, describe := range i.methods {
			buffer.WriteString(fmt.Sprintf("%s\\l\\n", describe))
		}
		label := fmt.Sprintf("interface: %s\\l\\n----\\lpackage: %s\\l\\nfile: %s\\l-----\\l%s", i.name, i.fileNode.packageName, i.fileNode.fileNodeTagName, buffer.String())
		return label
	} else {
		label := fmt.Sprintf("interface: %s\\l\\n----\\lpackage: %s\\l\\nfile: %s", i.name, i.fileNode.packageName, i.fileNode.fileNodeTagName)
		return label
	}
}

func (i *InterfaceNode) DrawNode(content *bytes.Buffer, record map[string]bool) {
	if _, ok := record[i.name]; ok == false {
		content.WriteString(fmt.Sprintf("%s [label=\"%s\", shape=\"box\"];", i.name+"v", i.getLabelDescribe(i.fileNode.nodeManager.detail)))
		content.WriteString("\n")
		record[i.name] = true
	}
}

func (i *InterfaceNode) DrawRelation(content *bytes.Buffer, record map[string]bool) {
	for structName, _ := range i.implementStruct {
		if _, ok := record[structName+"_"+i.name]; ok == false {
			content.WriteString(fmt.Sprintf("%s->%s [style=\"dashed\"];", structName+"v", i.name+"v"))
			content.WriteString("\n")
			record[structName+"_"+i.name] = true
		}
	}
}
