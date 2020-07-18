package fileparser

import (
	"bytes"
	"fmt"
)

type StructNode struct {
	fileNode      *FileNode
	name          string
	fields        map[string]string
	complexFields map[string]string
}

func NewStructNode(fileNode *FileNode, name string, fields map[string]string) *StructNode {
	return &StructNode{
		fileNode:      fileNode,
		name:          name,
		fields:        fields,
		complexFields: make(map[string]string, 0),
	}
}

func (s *StructNode) Merge(structTypes map[string]map[string]bool, interfaceNames map[string]map[string]bool) {
	// 对于每个field，查看项目中的全部struct，这里只是做简单include判断，如果包括就认为是对应的类型
	for name, t := range s.fields {
		keyFinalReceiver, finalReceiver := typeCompare(structTypes, interfaceNames, t)

		if finalReceiver != "" {
			s.complexFields[finalReceiver] = name + ":" + t
		}
		if keyFinalReceiver != "" {
			s.complexFields[keyFinalReceiver] = name + ":" + t
		}
	}
}

func (s *StructNode) getStructLabel(detail bool) string {
	if detail {
		buffer := bytes.NewBuffer([]byte{})
		for name, t := range s.fields {
			buffer.WriteString(fmt.Sprintf("%s: %s\\l\\n", name, t))
		}
		label := fmt.Sprintf("struct: %s\\l\\n----\\lpackage: %s\\l\\nfile: %s\\l----\\l%s", s.name, s.fileNode.packageName, s.fileNode.fileNodeTagName, buffer.String())
		return label
	} else {
		label := fmt.Sprintf("struct: %s\\l\\n----\\lpackage: %s\\l\\nfile: %s", s.name, s.fileNode.packageName, s.fileNode.fileNodeTagName)
		return label
	}
}

func (s *StructNode) DrawNode(content *bytes.Buffer, record map[string]bool) {
	if !s.fileNode.nodeManager.allField && len(s.complexFields) == 0 {
		return
	}
	if _, ok := record[s.name]; ok == false {
		content.WriteString(fmt.Sprintf("%s [label=\"%s\", shape=\"box\"];", s.name+"v", s.getStructLabel(s.fileNode.nodeManager.detail)))
		content.WriteString("\n")
		record[s.name] = true
	}
	for dest, _ := range s.complexFields {
		if _, ok := record[dest]; ok == false {
			label := s.fileNode.nodeManager.getStructReceiverLabel(dest)
			if label == "" {
				label = dest
			}
			content.WriteString(fmt.Sprintf("%s [label=\"%s\", shape=\"box\"];", dest+"v", label))
			content.WriteString("\n")
			record[dest] = true
		}
	}
}

func (s *StructNode) DrawRelation(content *bytes.Buffer, record map[string]bool) {
	for dest, label := range s.complexFields {
		if _, ok := record[s.name+"_"+dest]; ok == false {
			content.WriteString(fmt.Sprintf("%s->%s [label=\"%s\"];", s.name+"v", dest+"v", label))
			content.WriteString("\n")
		}
		record[s.name+"_"+dest] = true
	}
}
