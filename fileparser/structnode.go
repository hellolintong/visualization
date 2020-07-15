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
		keyFinalReceiver := ""
		/*
		uint8 8位无符号整型(0 to 255)
		uint16 16位无符号整型(0 to 65535)
		uint32 32位无符号整型(0 to 4294967295)
		uint64 64位无符号整型(0 to 18446744073709551615)
		int8 8位有符号整型(-128 to 127)
		int16 16位有符号整型(-32768 to 32767)
		int32 32位有符号整型(-2147483648 to 2147483647)
		int64 64位有符号整型(-9223372036854775808 to 9223372036854775807)
		float32 32位浮点类型
		float64 64位浮点类型
		complex32 由float32实部+虚部
		complex64 由float64实部+虚部
		byte uint8的别名
		rune int32的别名
		平台相关的类型
		uint，int 32或者是64位
		uintptr 一个足够表示指针的无符号整数
		 */
		for _, structType := range structTypes {
			// 基础类型
			if t == "string" || t == "int" || t == "uint" || t == "uint8" || t == "uint16" || t == "uint32" || t == "uint64" ||
				t == "int8" || t == "int16" || t == "int32" || t == "int64" || t == "float32" ||
				t == "float64" || t == "complex32" || t == "complex64" || t == "byte" || t == "rune" || t == "uintptr" {
				break
			}
			// 精准匹配
			if t == structType || t == "*"+structType {
				finalReceiver = structType
				break
			}
			// 取最长匹配（避免像map这种干扰)
			if strings.Contains(t, "]"+structType) || strings.Contains(t, "]*" + structType){
				if len(structType) > len(finalReceiver) {
					finalReceiver = structType
				}
			}

			if strings.HasPrefix(t, "map") && strings.Contains(t, "["+structType+ "]") || strings.Contains(t, "[*" + structType+"]"){
				if len(structType) > len(keyFinalReceiver) {
					keyFinalReceiver = structType
				}
			}
		}
		if finalReceiver != "" {
			s.complexFields[finalReceiver] = true
		}
		if keyFinalReceiver != "" {
			s.complexFields[keyFinalReceiver] = true
		}
	}
}

func (s *StructNode) getStructLabel(detail bool) string {
	if detail {
		buffer := bytes.NewBuffer([]byte{})
		for name, t := range s.fields {
			buffer.WriteString(fmt.Sprintf("%s:%s\\l", name, t))
		}
		label := fmt.Sprintf("package:%s \\l file:%s \\l struct:%s \\l ---------- \\l %s", s.fileNode.packageName, s.fileNode.fileNodeTagName, s.name, buffer.String())
		return label
	} else {
		label := fmt.Sprintf("package:%s \\l file:%s \\l struct:%s", s.fileNode.packageName, s.fileNode.fileNodeTagName, s.name)
		return label
	}
}

func (s *StructNode) DrawNode(content *bytes.Buffer, record map[string]bool){
	if !s.fileNode.nodeManager.allField && len(s.complexFields) == 0 {
		return
	}
	if _, ok := record[s.name]; ok == false {
		content.WriteString(fmt.Sprintf("%s [label=\"%s\", shape=\"box\"];", s.name + "v", s.getStructLabel(s.fileNode.nodeManager.detail)))
		content.WriteString("\n")
		record[s.name] = true
	}
	for dest, _ := range s.complexFields {
		if _, ok := record[dest]; ok == false {
			label := s.fileNode.nodeManager.getStructReceiverLabel(dest)
			if label == "" {
				label = dest
			}
			content.WriteString(fmt.Sprintf("%s [label=\"%s\", shape=\"box\"];", dest + "v", label))
			content.WriteString("\n")
			record[dest] = true
		}
	}
}

func (s *StructNode) DrawRelation(content *bytes.Buffer, record map[string]bool){
	for dest, _ := range s.complexFields {
		if _, ok := record[s.name + "_" + dest]; ok == false {
			content.WriteString(fmt.Sprintf("%s->%s;", s.name + "v", dest + "v"))
			content.WriteString("\n")
		}
		record[s.name + "_" + dest] = true
	}
}