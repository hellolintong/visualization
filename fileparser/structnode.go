package fileparser

import (
	"bytes"
	"fmt"
	"strings"
)

type StructNode struct {
	fileNode      *FileNode
	name          string
	fields        map[string]string
	complexStructFields map[string]*StructNode
	complexInterfaceFields map[string]*InterfaceNode
}

func NewStructNode(fileNode *FileNode, name string, fields map[string]string) *StructNode {
	return &StructNode{
		fileNode:      fileNode,
		name:          name,
		fields:        fields,
		complexStructFields: make(map[string]*StructNode, 0),
		complexInterfaceFields: make(map[string]*InterfaceNode, 0),
	}
}

func (s *StructNode) Merge(structTypes map[string]map[string]*StructNode, interfaceNames map[string]map[string]*InterfaceNode) {

	// 对于每个field，查看项目中的全部struct，这里只是做简单include判断，如果包括就认为是对应的类型
	for name, t := range s.fields {
		if strings.Contains(t, "map[") {
			index1 := strings.Index(t, "map[")
			index2 := strings.Index(t[index1:], "]")
			t2 := t[index1 + 4: index2]
			t3 := t[index2 + 1: ]
			finalStructType, finalInterfaceType := typeCompare(structTypes, interfaceNames, t2)

			if finalStructType != nil {
				s.complexStructFields[name+":"+t2] = finalStructType
			}

			if finalInterfaceType != nil {
				s.complexInterfaceFields[name+":"+t2] = finalInterfaceType
			}

			finalStructType2, finalInterfaceType2 := typeCompare(structTypes, interfaceNames, t3)

			if finalStructType2 != nil {
				s.complexStructFields[name+":"+t3] = finalStructType2
			}

			if finalInterfaceType2 != nil {
				s.complexInterfaceFields[name+":"+t3] = finalInterfaceType2
			}
		} else {
			finalStructType, finalInterfaceType := typeCompare(structTypes, interfaceNames, t)

			if finalStructType != nil {
				s.complexStructFields[name+":"+t] = finalStructType
			}

			if finalInterfaceType != nil {
				s.complexInterfaceFields[name+":"+t] = finalInterfaceType
			}
		}
	}
}

func (s *StructNode) getIdentity() string {
	return "\"" + s.fileNode.packageName + "/" + s.name + "\""
}

func (s *StructNode) getStructLabel() string {
	buffer := bytes.NewBuffer([]byte{})
	for name, t := range s.fields {
		buffer.WriteString(fmt.Sprintf("%s: %s\\l\\n", name, t))
	}
	label := fmt.Sprintf("struct: %s\\l\\n----\\lpackage: %s\\l\\nfile: %s\\l----\\l%s", s.name, s.fileNode.packageName, s.fileNode.fileNodeTagName, buffer.String())
	return label
}

func (s *StructNode) DrawNode(content *bytes.Buffer, record map[string]bool, count int) {

	content.WriteString(fmt.Sprintf("%s [label=\"%s\", shape=\"box\"];", s.getIdentity(), s.getStructLabel()))
	content.WriteString("\n")

	record[s.getIdentity()] = true
	count--
	if count > 0  {
		for _, node := range s.complexStructFields {
			node.DrawNode(content, record, count)
		}
	}

	if count > 0 {
		for _, node := range s.complexInterfaceFields {
			node.DrawNode(content, record)
		}
	}
}

func (s *StructNode) DrawRelation(content *bytes.Buffer, record map[string]bool, count int) {

	tempRecord := make(map[string]bool, 0)
	for _, node := range s.complexStructFields {
			if tempRecord[node.getIdentity()] == false {
				content.WriteString(fmt.Sprintf("%s -> %s", s.getIdentity(), node.getIdentity()))
				content.WriteString("\n")
				tempRecord[node.getIdentity()] = true
			}
	}
	for _, node := range s.complexInterfaceFields {
		if tempRecord[s.getIdentity() + node.getIdentity()] == false {
			content.WriteString(fmt.Sprintf("%s -> %s", s.getIdentity(), node.getIdentity()))
			content.WriteString("\n")
			tempRecord[s.getIdentity() + node.getIdentity()] = true
		}
	}
	record[s.getIdentity()] = true
	count--
	if count > 0 {
		for _, node := range s.complexStructFields {
			if record[node.getIdentity()] == false {
				node.DrawRelation(content, record, count)
			}
		}
	}
}
