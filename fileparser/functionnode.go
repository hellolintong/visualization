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
	parameters    []string
	returns       []string
	callee 		  map[string]*FunctionNode
	body          string
}

func NewFunctionNode(fileNode *FileNode, name string, receiver string, body string, params []string, returns []string) *FunctionNode {
	if receiver != "" {
		elem := strings.Split(receiver, " ")
		receiver = elem[len(elem)-1]
		receiver = strings.Trim(receiver, "*")
	}
	return &FunctionNode{
		fileNode:      fileNode,
		name:          name,
		receiver:      receiver,
		body:          body,
		callee: make(map[string]*FunctionNode, 0),
		parameters:    params,
		returns:       returns,
	}
}

func (s *FunctionNode) Merge() {
	nodeManager := s.fileNode.nodeManager
	if _, ok := nodeManager.allFunctions[s.name]; !ok {
		nodeManager.allFunctions[s.name] = make([]*FunctionNode, 0)
	}
	nodeManager.allFunctions[s.name] = append(nodeManager.allFunctions[s.name], s)
}

func (s *FunctionNode) getIdentity() string {
	return "\""+ s.fileNode.packageName + "/" + s.receiver + "/" + s.name + "\""
}

func (s *FunctionNode) Deduce() {
	nodeManager := s.fileNode.nodeManager
	lines := strings.Split(s.body, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// 跳过注释
		if strings.HasPrefix(line, "//") || strings.HasPrefix(line, "/*") || strings.HasPrefix(line, "*") {
			continue
		}

		// 遍历全部现有的函数，查看是否存在调用关系
		for functionName, nodes := range nodeManager.allFunctions {
			if strings.Contains(line, " " + functionName+"(") {
				for _, node := range nodes {
					if node.fileNode.packageName == s.fileNode.packageName && node.receiver == "" {
						s.callee[node.getIdentity()] = node
					}
				}
			} else if strings.Contains(line, "." + functionName+"(") {
				pos := strings.Index(line, "." + functionName+"(") - 1
				tmp := bytes.NewBuffer([]byte{})
				for pos >= 0 && line[pos] != ' ' {
					tmp.WriteByte(line[pos])
					pos -= 1
				}

				found := false
				// 是否在模块中
				name := tmp.String()
				if _, ok  := s.fileNode.importers[name]; ok {
					for _, node := range nodes {
						if node.fileNode.packageName == name {
							s.callee[node.getIdentity()] = node
							found = true
							break
						}
					}
				} else {
					for _, node := range nodes {
						if node.receiver == s.receiver {
							s.callee[node.getIdentity()] = node
							found = true
							break
						}
					}
				}
				if found == false {
					//if len(nodes) == 1 && s.fileNode.nodeManager.knownModuleFunction[nodes[0].name] == true {
					//	s.callee[nodes[0].getIdentity()] = nodes[0]
					//	break
					//}
					//// 优先找同一个模块里面的
					//fmt.Println(s.fileNode.file)
					//fmt.Println(s.getIdentity())
					//fmt.Println(line)
					//for i, node := range nodes {
					//	fmt.Printf("node %d: %s\n", i , node.getIdentity())
					//}
					//
					//index := -1
					//fmt.Println("输入对应的node序号:")
					//_, _ = fmt.Scanf("%d", &index)
					//if index >= 0 {
					//	s.callee[nodes[index].getIdentity()] = nodes[index]
					//}
					//if len(nodes) == 1 && unicode.IsLower([]rune(nodes[0].name)[0]) && index == 0{
					//	s.fileNode.nodeManager.knownModuleFunction[nodes[0].name] = true
					//}
				}
			}
		}
	}
}


func (s *FunctionNode) DrawNode(content *bytes.Buffer, record map[string]bool, count int) {
	s.Deduce()
	content.WriteString("\n")
	content.WriteString(fmt.Sprintf("%s [label=%s, shape=\"box\"];",  s.getIdentity(), s.getIdentity()))

	record[s.getIdentity()] = true
	count--

	if count > 0 {
		for _, callee := range s.callee {
			callee.DrawNode(content, record, count)
		}
	}
}

func (s *FunctionNode) DrawRelation(content *bytes.Buffer, record map[string]bool, count int) {
	tempRecord := make(map[string]bool, 0)

	for _, callee := range s.callee {
			if tempRecord[callee.getIdentity()] == false {
				content.WriteString(fmt.Sprintf("%s->%s;", s.getIdentity(), callee.getIdentity()))
				content.WriteString("\n")
				tempRecord[callee.getIdentity()] = true
			}
	}

	count--

	record[s.getIdentity()] = true;
	if count > 0 {
		for _, callee := range s.callee {
			if record[callee.getIdentity()] == false {
				callee.DrawRelation(content, record, count)
			}
		}
	}
}

