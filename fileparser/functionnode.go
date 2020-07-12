package fileparser

import (
	"bytes"
	"fmt"
	"strings"
)

type FunctionNode struct {
	fileNode *FileNode
	name   string
	receiver string
	callFunctions map[string]bool
	body string
}

func NewFunctionNode(fileNode *FileNode, name string, receiver string, body string) *FunctionNode {
	if receiver != "" {
		elem := strings.Split(receiver, " ")
		receiver = elem[len(elem) - 1]
		receiver = strings.Trim(receiver, "*")
	}
	return &FunctionNode{
		fileNode: fileNode,
		name: name,
		receiver: receiver,
		body: body,
		callFunctions: make(map[string]bool, 0),
	}
}

// 关联函数调用
func (s *FunctionNode) Merge(functionNames map[string][]string)  {
	for receiver, names := range functionNames {
		for _, name := range names {
			// 如果是自身模块的调动，就忽略掉
			if receiver != s.receiver && s.receiver != "" && receiver != "" {
				if strings.Contains(s.body, name+"(") {
					s.callFunctions[receiver] = true
				}
			}
		}
	}
}

func (s *FunctionNode) DrawNode(content *bytes.Buffer, receiver map[string]bool){
	// 如果没有函数引用，就不绘制
	if len(s.callFunctions) == 0 {
		return
	}

	// 如果存在接收者，则设置接收者，注意只设置一次
	_ , ok := receiver[s.receiver]
	if s.receiver != "" && ok == false {
		// 设置自己的node
		content.WriteString("\n")
		label := fmt.Sprintf("package:%s \\l file:%s \\l struct:%s \\l", s.fileNode.packageName, s.fileNode.fileNodeTagName, s.receiver)
		content.WriteString(fmt.Sprintf("%s [label=\"%s\", shape=\"box\"];", s.receiver, label))
	}
	// 设置关联对象的node
	for dest := range s.callFunctions {
		if ok := receiver[dest]; ok == false {
			label := s.fileNode.nodeManager.getReceiverLabel(dest)
			if label == "" {
				label = dest
			}
			content.WriteString(fmt.Sprintf("%s [label=\"%s\", shape=\"box\"];", dest, label))
		}
		receiver[dest] = true
	}
	receiver[s.receiver] = true
}

func (s *FunctionNode) DrawRelation(content *bytes.Buffer, record map[string]bool){
	if s.receiver != "" {
		for dest := range s.callFunctions {
			if record[s.receiver + "_" + dest] == false {
				record[s.receiver+"_"+dest] = true
				content.WriteString(fmt.Sprintf("%s->%s;", s.receiver, dest))
				content.WriteString("\n")
			}
		}
	}
}