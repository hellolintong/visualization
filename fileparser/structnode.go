package fileparser

import (
	"bytes"
	"fmt"
	"strings"
)

type StructNode struct {
	fileNode *FileNode
	name   string
	fields map[string]string
	complexFields map[string]bool
}

func NewStructNode(fileNode *FileNode, name string) *StructNode {
	return &StructNode{
		fileNode: fileNode,
		name: name,
		fields: make(map[string]string, 0),
		complexFields: make(map[string]bool, 0),
	}
}

func (s *StructNode) Merge(structTypes []string)  {
	// 对于每个field，查看项目中的全部struct，这里只是做简单include判断，如果包括就认为是对应的类型
	for _, t := range s.fields {
		for _, structType := range structTypes {
			if strings.Contains(t, structType) {
				s.complexFields[structType] = true
				break
			}
		}
	}
}

func (s *StructNode) DrawNode(content *bytes.Buffer, record map[string]bool){
	if len(s.complexFields) == 0 {
		return
	}
	if _, ok := record[s.name]; ok == false {
		label := fmt.Sprintf("package:%s \\l file:%s \\l struct:%s \\l", s.fileNode.packageName, s.fileNode.fileNodeTagName, s.name)
		content.WriteString(fmt.Sprintf("%s [label=\"%s\", shape=\"box\"];", s.name, label))
		content.WriteString("\n")
		record[s.name] = true
	}
	for dest, _ := range s.complexFields {
		if _, ok := record[dest]; ok == false {
			label := s.fileNode.nodeManager.getReceiverLabel(dest)
			if label == "" {
				label = dest
			}
			content.WriteString(fmt.Sprintf("%s [label=\"%s\", shape=\"box\"];", dest, label))
			content.WriteString("\n")
			record[dest] = true
		}
	}
}

func (s *StructNode) DrawRelation(content *bytes.Buffer, record map[string]bool){
	for dest, _ := range s.complexFields {
		if _, ok := record[s.name + "_" + dest]; ok == false {
			content.WriteString(fmt.Sprintf("%s->%s;", s.name, dest))
			content.WriteString("\n")
		}
		record[s.name + "_" + dest] = true
	}
}