package fileparser

import (
	"bytes"
	"fmt"
	"strings"
)

type FunctionNode struct {
	fileNode      *FileNode
	name          string
	receiver      string
	calledStructs map[string]map[string]bool
	body          string
}

func NewFunctionNode(fileNode *FileNode, name string, receiver string, body string) *FunctionNode {
	if receiver != "" {
		elem := strings.Split(receiver, " ")
		receiver = elem[len(elem) - 1]
		receiver = strings.Trim(receiver, "*")
	}
	return &FunctionNode{
		fileNode:      fileNode,
		name:          name,
		receiver:      receiver,
		body:          body,
		calledStructs: make(map[string]map[string]bool, 0),
	}
}

// 关联函数调用
func (s *FunctionNode) Merge(functionNames map[string][]string)  {
	for receiver, names := range functionNames {
		for _, name := range names {
			// 如果是自身模块的调动，就忽略掉
			if receiver != s.receiver && s.receiver != "" && receiver != "" {
				if strings.Contains(s.body, "." + name+"(") {
					if _, ok := s.calledStructs[receiver]; ok == false {
						s.calledStructs[receiver] = make(map[string]bool)
					}
					// receiver和function
					s.calledStructs[receiver][name] = true
				}
			}
		}
	}
}


func (s *FunctionNode) DrawNode(content *bytes.Buffer, receiver map[string]bool){
	// 如果没有函数引用，就不绘制
	if len(s.calledStructs) == 0 {
		return
	}

	if s.receiver == "" {
		return
	}

	// 如果存在接收者，则设置接收者，注意只设置一次
	if _, ok := receiver[s.receiver]; ok == false {
		// 设置自己的node
		content.WriteString("\n")
		label := fmt.Sprintf("package:%s \\l file:%s \\l struct:%s \\l ", s.fileNode.packageName, s.fileNode.fileNodeTagName, s.receiver)
		content.WriteString(fmt.Sprintf("%s [label=\"%s\", shape=\"box\"];", s.receiver, label))
		receiver[s.receiver] = true
	}

	// 设置函数节点
	if _, ok := receiver[s.receiver + "_" + s.name]; ok == false {
		// function
		content.WriteString(fmt.Sprintf("%s [label=\"%s\", shape=\"box\"];", s.receiver + "_" + s.name, "function:" + s.name))
		content.WriteString("\n")
		receiver[s.receiver + "_" + s.name] = true
	}

	// 设置关联对象的node
	for calledReceiver, _ := range s.calledStructs {
		// 记录struct 节点
		if _, ok := receiver[calledReceiver]; ok == false {
			label := s.fileNode.nodeManager.getFunctionReceiverLabel(calledReceiver)
			if label == "" {
				label = calledReceiver
			}
			// struct
			content.WriteString(fmt.Sprintf("%s [label=\"%s\", shape=\"box\"];", calledReceiver + "v", label))
			content.WriteString("\n")
			receiver[calledReceiver] = true
		}

		//// 记录函数节点
		//for function := range functions {
		//	if receiver[function + "_" + calledReceiver] == false {
		//		content.WriteString(fmt.Sprintf("%s [label=\"%s\", shape=\"box\"];", calledReceiver + "_" + function, function))
		//	}
		//	receiver[function + "_" + calledReceiver] = true
		//}
	}
}

func (s *FunctionNode) DrawRelation(content *bytes.Buffer, record map[string]bool){
	if s.receiver == "" {
		return
	}
	if len(s.calledStructs) == 0 {
		return
	}
	if _, ok := record[s.receiver +"_"+s.name]; ok == false {
		content.WriteString(fmt.Sprintf("%s->%s [style=\"dashed\"];", s.receiver + "v", s.receiver + "_" + s.name))
		content.WriteString("\n")
		record[s.receiver + "_" + s.name] = true
	}
	for calledReceiver, _ := range s.calledStructs {
		if record[s.receiver + "_" + s.name +"_" + calledReceiver] == false {
			content.WriteString(fmt.Sprintf("%s->%s;", s.receiver + "_" + s.name, calledReceiver + "v"))
			content.WriteString("\n")
			record[s.receiver + "_" + s.name + "_" + calledReceiver] = true
		}
	}
}