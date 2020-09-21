package fileparser

import (
	"bytes"
	"fmt"
	"strings"
)

type FunctionNode struct {
	fileNode   *FileNode
	name       string
	receiver   string
	parameters []string
	returns    []string
	callee     map[string]*FunctionNode
	caller     map[string]*FunctionNode
	content    string
}

func NewFunctionNode(fileNode *FileNode, name string, receiver string, content string, params []string, returns []string) *FunctionNode {
	if receiver != "" {
		elem := strings.Split(receiver, " ")
		receiver = elem[len(elem)-1]
		receiver = strings.Trim(receiver, "*")
	}
	return &FunctionNode{
		fileNode:   fileNode,
		name:       name,
		receiver:   receiver,
		content:    content,
		callee:     make(map[string]*FunctionNode, 0),
		caller:     make(map[string]*FunctionNode, 0),
		parameters: params,
		returns:    returns,
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
	return "\"" + s.fileNode.packageName + "/" + s.receiver + "/" + s.name + "\""
}

func (s *FunctionNode) GetCalleeCodeSnippet(result map[string]string) map[string]string {
	s.deduceCallee()
	result[s.getIdentity()] = s.content

	for _, callee := range s.callee {
		if _, ok := result[callee.getIdentity()]; ok == false {
			result[callee.getIdentity()] = callee.content
			tmp := callee.GetCalleeCodeSnippet(result)
			for k, v := range tmp {
				result[k] = v
			}
		}
	}

	return result
}

func (s *FunctionNode) GetCallerCodeSnippet(result map[string]string) map[string]string {
	s.deduceCaller()
	result[s.getIdentity()] = s.content
	for _, caller := range s.caller {
		if _, ok := result[caller.getIdentity()]; ok == false {
			result[caller.getIdentity()] = caller.content
			tmp := caller.GetCallerCodeSnippet(result)
			for k, v := range tmp {
				result[k] = v
			}
		}
	}

	return result
}

// 查找函数的调用者
func (s *FunctionNode) deduceCaller() {
	for _, nodes := range s.fileNode.nodeManager.allFunctions {
		for _, node := range nodes {
			if node.getIdentity() == s.getIdentity() {
				continue
			}
			lines := strings.Split(node.content, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				// 跳过函数头，注释
				if !strings.HasPrefix(line, "*") && !strings.HasPrefix(line, "func") && !strings.HasPrefix(line, "//") &&
					!strings.HasPrefix(line, "/*") && (strings.Contains(line, "."+s.name+"(") || strings.Contains(line, " "+s.name+"("))  {
					node2 := node.checkCallerInvolved(line, s)
					if node2 != nil && node2 == s {
						s.caller[node.getIdentity()] = node
					}
				}
			}
		}
	}
}

func (s *FunctionNode) checkCallerInvolved(line string, node *FunctionNode) *FunctionNode{
	if strings.Contains(line, " "+node.name+"(") {
		if node.fileNode.packageName == s.fileNode.packageName && node.receiver == "" {
			return node
		}
	} else if strings.Contains(line, "."+node.name+"(") {
		pos := strings.Index(line, "."+node.name+"(") - 1
		tmp := bytes.NewBuffer([]byte{})
		for pos >= 0 && line[pos] != ' ' {
			tmp.WriteByte(line[pos])
			pos -= 1
		}

		// 是否在模块中
		name := tmp.String()
		if _, ok := s.fileNode.importers[name]; ok {
			if node.fileNode.packageName == name {
				return node
			}
		} else {
			if node.receiver == s.receiver {
				return node
			}
		}

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
	return nil
}

func (s *FunctionNode) checkCalleeInvolved(line string, functionName string, nodes []*FunctionNode) *FunctionNode{
	if strings.Contains(line, " "+functionName+"(") {
		for _, node := range nodes {
			if node.fileNode.packageName == s.fileNode.packageName && node.receiver == "" {
				return node
			}
		}
	} else if strings.Contains(line, "."+functionName+"(") {
		pos := strings.Index(line, "."+functionName+"(") - 1
		tmp := bytes.NewBuffer([]byte{})
		for pos >= 0 && line[pos] != ' ' {
			tmp.WriteByte(line[pos])
			pos -= 1
		}

		// 是否在模块中
		name := tmp.String()
		if _, ok := s.fileNode.importers[name]; ok {
			for _, node := range nodes {
				if node.fileNode.packageName == name {
					return node
				}
			}
		} else {
			for _, node := range nodes {
				if node.receiver == s.receiver {
					return node
				}
			}
		}
		if len(nodes) == 1 {
			return nodes[0]
		}
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
	return nil
}

func (s *FunctionNode) deduceCallee() {
	lines := strings.Split(s.content, "\n")
	// 跳过自己
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// 跳过注释
		if strings.HasPrefix(line, "func") || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "/*") || strings.HasPrefix(line, "*") {
			continue
		}

		// 跳过自己
		if strings.Contains(line, s.name+"(") {
			continue
		}
		nodeManager := s.fileNode.nodeManager
		for functionName, nodes := range nodeManager.allFunctions {
			// 遍历全部现有的函数，查看是否存在调用关系
			node := s.checkCalleeInvolved(line, functionName, nodes)
			if node != nil {
				s.callee[node.getIdentity()] = node
			}
		}
	}
}

func (s *FunctionNode) DrawCallerNode(content *bytes.Buffer, record map[string]bool, count int) {
	s.deduceCaller()
	content.WriteString("\n")
	content.WriteString(fmt.Sprintf("%s [label=%s, shape=\"box\"];", s.getIdentity(), s.getIdentity()))

	record[s.getIdentity()] = true
	count--

	if count > 0 {
		for _, callee := range s.caller {
			callee.DrawCallerNode(content, record, count)
		}
	}
}

func (s *FunctionNode) DrawCallerRelation(content *bytes.Buffer, record map[string]bool, count int) {
	tempRecord := make(map[string]bool, 0)

	for _, caller := range s.caller {
		if tempRecord[caller.getIdentity()] == false {
			content.WriteString(fmt.Sprintf("%s->%s;", caller.getIdentity(), s.getIdentity()))
			content.WriteString("\n")
			tempRecord[caller.getIdentity()] = true
		}
	}

	count--

	record[s.getIdentity()] = true
	if count > 0 {
		for _, caller := range s.caller {
			if record[caller.getIdentity()] == false {
				caller.DrawCallerRelation(content, record, count)
			}
		}
	}
}

func (s *FunctionNode) DrawCalleeNode(content *bytes.Buffer, record map[string]bool, count int) {
	s.deduceCallee()
	content.WriteString("\n")
	content.WriteString(fmt.Sprintf("%s [label=%s, shape=\"box\"];", s.getIdentity(), s.getIdentity()))

	record[s.getIdentity()] = true
	count--

	if count > 0 {
		for _, callee := range s.callee {
			callee.DrawCalleeNode(content, record, count)
		}
	}
}

func (s *FunctionNode) DrawCalleeRelation(content *bytes.Buffer, record map[string]bool, count int) {
	tempRecord := make(map[string]bool, 0)

	for _, callee := range s.callee {
		if tempRecord[callee.getIdentity()] == false {
			content.WriteString(fmt.Sprintf("%s->%s;", s.getIdentity(), callee.getIdentity()))
			content.WriteString("\n")
			tempRecord[callee.getIdentity()] = true
		}
	}

	count--

	record[s.getIdentity()] = true
	if count > 0 {
		for _, callee := range s.callee {
			if record[callee.getIdentity()] == false {
				callee.DrawCalleeRelation(content, record, count)
			}
		}
	}
}
