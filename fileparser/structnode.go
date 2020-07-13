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
		finalReceiver := ""
		for _, structType := range structTypes {
			// 精准匹配
			if t == structType || t == "*"+structType {
				finalReceiver = structType
				break
			}
			// 取最长匹配（避免像map这种干扰)
			if strings.Contains(t, structType) {
				if len(structType) > len(finalReceiver) {
					finalReceiver = structType
				}
			}
		}
		if finalReceiver != "" {
			s.complexFields[finalReceiver] = true
		}
	}
}

func (s *StructNode) getStructLabel(detail bool) string {
	if detail {
		buffer := bytes.NewBuffer([]byte{})
		for name, t := range s.fields {
			buffer.WriteString(fmt.Sprintf("%s:%s\\l", name, t))
		}
		label := fmt.Sprintf("package:%s \\l file:%s \\l struct:%s \\l %s", s.fileNode.packageName, s.fileNode.fileNodeTagName, s.name, buffer.String())
		return label
	} else {
		label := fmt.Sprintf("package:%s \\l file:%s \\l struct:%s", s.fileNode.packageName, s.fileNode.fileNodeTagName, s.name)
		return label
	}
}

func (s *StructNode) DrawNode(content *bytes.Buffer, record map[string]bool){
	if len(s.complexFields) == 0 {
		return
	}
	if _, ok := record[s.name]; ok == false {
		content.WriteString(fmt.Sprintf("%s [label=\"%s\", shape=\"box\"];", s.name, s.getStructLabel(true)))
		content.WriteString("\n")
		record[s.name] = true
	}
	for dest, _ := range s.complexFields {
		if _, ok := record[dest]; ok == false {
			label := s.fileNode.nodeManager.getReceiverLabel(dest, true)
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