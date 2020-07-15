package fileparser

import (
	"bytes"
	"fmt"
)

type InterfaceNode struct {
	fileNode      *FileNode
	name          string
	implementStruct map[string]bool
	methods map[string]string
}

func NewInterfaceNode(fileNode *FileNode, name string) *InterfaceNode {
	return &InterfaceNode{
		fileNode: fileNode,
		name: name,
		implementStruct: make(map[string]bool, 0),
		methods: make(map[string]string, 0),
	}
}

func (i *InterfaceNode) mergeImplement(functionNames map[string][]string) {
	label:
	for receiver, functions := range functionNames {
		// 获取该结构体的方法
		tmp := make(map[string]bool, 0)
		for _, function := range functions {
			tmp[function] = true
		}

		// 查看接口中的全部方法是否都实现了
		for method, _ := range i.methods {
			if _, ok := tmp[method]; !ok {
				continue label
			}
		}
		i.implementStruct[receiver] = true
	}
}

func (i *InterfaceNode) getLabelDescribe() string{
		buffer := bytes.NewBuffer([]byte{})
		for _, describe := range i.methods {
			buffer.WriteString(fmt.Sprintf("%s\\l", describe))
		}
		label := fmt.Sprintf("package:%s \\l file:%s \\l interface:%s \\l ---------- \\l %s", i.fileNode.packageName, i.fileNode.fileNodeTagName, i.name, buffer.String())
		return label
}


func (i *InterfaceNode) DrawNode(content *bytes.Buffer, record map[string]bool){
	if _, ok := record[i.name]; ok == false {
		content.WriteString(fmt.Sprintf("%s [label=\"%s\", shape=\"box\"];", i.name + "v", i.getLabelDescribe()))
		content.WriteString("\n")
		record[i.name] = true
	}
}

func (i *InterfaceNode) DrawRelation(content *bytes.Buffer, record map[string]bool){
	for structName, _ := range i.implementStruct {
		if _, ok := record[structName + "_" + i.name]; ok == false {
			content.WriteString(fmt.Sprintf("%s->%s [style=\"dashed\"];", structName + "v", i.name + "v"))
			content.WriteString("\n")
			record[structName + "_" + i.name] = true
		}
	}
}