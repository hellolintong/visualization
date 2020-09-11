package fileparser

import (
	"bytes"
	"fmt"
	"github.com/codeskyblue/go-sh"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type NodeManager struct {
	projectPath string
	packages        map[string][]*FileNode
	structTypes     map[string]map[string]*StructNode
	functionNames   map[string]map[string]bool
	interfaceNames  map[string]map[string]*InterfaceNode
	allFunctions    map[string][]*FunctionNode
	allStructs    	map[string]*StructNode
	knownModuleFunction map[string]bool
}

func (n *NodeManager) GetStructCodeSnippet(baseName string) map[string]string {
	structNode, ok := n.allStructs[baseName]
	if !ok {
		return map[string]string{}
	}
	return structNode.GetCodeSnippet()
}

func (n *NodeManager) getMatchedFunction(baseName string) *FunctionNode {
	var node *FunctionNode
	elems := strings.Split(baseName, "/")
	if len(elems) != 3 {
		return node
	}
	packageName, receiver, name := elems[0], elems[1], elems[2]
	functionNodes, ok := n.allFunctions[name]
	if !ok {
		return node
	}

	for _, functionNode := range functionNodes {
		if functionNode.fileNode.packageName == packageName && functionNode.receiver == receiver {
			node = functionNode
			break
		}
	}
	return node
}

func (n *NodeManager) GetFunctionCodeSnippet(baseName string) map[string]string {
	node := n.getMatchedFunction(baseName)
	if node == nil {
		return map[string]string{}
	}

	return node.GetCodeSnippet()
}

func (n *NodeManager) Relation() map[string][]string {
	relation := make(map[string][]string, 0)
	relation["functions"] = make([]string, 0)
	for _, nodes := range n.allFunctions {
		for _, node := range nodes {
			relation["functions"] = append(relation["functions"], strings.Trim(node.getIdentity(), "\""))
		}
	}
	sort.Strings(relation["functions"])

	relation["structs"] = make([]string, 0)
	for _, node := range n.allStructs {
			relation["structs"] = append(relation["structs"], strings.Trim(node.getIdentity(), "\""))
	}
	sort.Strings(relation["structs"])
	return relation
}

func (n *NodeManager) getStructReceiverLabel(receiver string) string {
	for _, fileNodes := range n.packages {
		for _, fileNode := range fileNodes {
			if structNode, ok := fileNode.structNodes[receiver]; ok == true {
				return structNode.getStructLabel()
			}
		}
	}
	return ""
}

func (n *NodeManager) DrawStruct(baseStruct string, count int) {

	structNode, ok := n.allStructs[baseStruct]
	if !ok {
		return
	}

	content := bytes.NewBuffer([]byte{})
	content.WriteString("digraph gph {")

	// 绘制base struct节点
	record := make(map[string]bool, 0)
	structNode.DrawNode(content, record, count)

	record = make(map[string]bool, 0)
	structNode.DrawRelation(content, record, count)


	content.WriteString("}")

	os.MkdirAll(n.getBaseDir(), 0755)
	baseStruct = strings.Trim(baseStruct, "\"")
	baseStruct = strings.ReplaceAll(baseStruct, "/", "_")
	if err := ioutil.WriteFile(fmt.Sprintf(n.getBaseDir()+"/struct_%s.dot", baseStruct), content.Bytes(), 0644); err == nil {
		if err = sh.Command("/bin/bash", "-c", fmt.Sprintf("dot %s/struct_%s.dot -o %s/struct_%s.png -Tpng", n.getBaseDir(), baseStruct, n.getBaseDir(), baseStruct)).Run(); err == nil {
			log.Println("draw success!")
		} else {
			log.Printf("draw visualization.dot fail, error:%s", err.Error())
		}
	} else {
		log.Printf("write visualization.dot fail, error:%s", err.Error())
	}
}

func (n *NodeManager) getBaseDir() string {
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
		return "."
	}
	projectName := filepath.Base(n.projectPath)
	baseDir := 	path+"/resource/"+projectName
	return baseDir
}

func (n *NodeManager) DrawFunction(baseFunction string, count int) {
	node := n.getMatchedFunction(baseFunction)
	if node == nil {
		return
	}

	content := bytes.NewBuffer([]byte{})
	content.WriteString("digraph gph {")



	record := make(map[string]bool, 0)
	node.DrawNode(content, record, count)

	record = make(map[string]bool, 0)
	node.DrawRelation(content, record, count)

	content.WriteString("}")

	os.MkdirAll(n.getBaseDir(), 0755)

	baseFunction = strings.ReplaceAll(baseFunction, "/", "_")
	if err := ioutil.WriteFile(fmt.Sprintf(n.getBaseDir()+"/function_%s.dot", baseFunction), content.Bytes(), 0644); err == nil {
		if err = sh.Command("/bin/bash", "-c", fmt.Sprintf("dot %s/function_%s.dot -o %s/function_%s.png -Tpng", n.getBaseDir(), baseFunction,
			n.getBaseDir(), baseFunction)).Run(); err == nil {
			log.Println("draw success!")
		} else {
			log.Printf("draw visualization.dot fail, error:%s", err.Error())
		}
	} else {
		log.Printf("write visualization.dot fail, error:%s", err.Error())
	}
}

func (n *NodeManager) mergeInterfaceImplement() {

	// package
	for _, package_ := range n.packages {
		// file
		for _, filenode := range package_ {
			// interface
			for _, interfaceNode := range filenode.interfaceNodes {
				interfaceNode.mergeImplement(n.functionNames)
			}
		}
	}
}

func (n *NodeManager) mergeFunction() {
	// 归并
	for _, package_ := range n.packages {
		for _, filenode := range package_ {
			filenode.MergeFunction()
		}
	}
}

func (n *NodeManager) mergeStruct() {
	// package
	for _, package_ := range n.packages {
		// file
		for _, filenode := range package_ {
			// structs
			for _, structNode := range filenode.structNodes {
				if _, ok := n.structTypes[filenode.packageName]; ok == false {
					n.structTypes[filenode.packageName] = make(map[string]*StructNode, 0)
				}
				n.structTypes[filenode.packageName][structNode.name] = structNode
			}
		}
	}


	// package
	for _, package_ := range n.packages {
		// file
		for _, filenode := range package_ {
			// functions
			for _, interfaceNode := range filenode.interfaceNodes {
				if _, ok := n.interfaceNames[filenode.packageName]; ok == false {
					n.interfaceNames[filenode.packageName] = make(map[string]*InterfaceNode, 0)
				}
				n.interfaceNames[filenode.packageName][interfaceNode.name] = interfaceNode
			}
		}
	}

	// 归并
	for _, package_ := range n.packages {
		for _, filenode := range package_ {
			filenode.MergeStruct(n.structTypes, n.interfaceNames)
		}
	}

	for _, package_ := range n.packages {
		for _, filenode := range package_ {
			for _, structNode := range filenode.structNodes {
				n.allStructs[structNode.getIdentity()] = structNode
			}
		}
	}

}

func (n *NodeManager) Merge() {
	// 输出结构体依赖
	n.mergeStruct()

	// 输出函数依赖
	n.mergeFunction()

	//  查看方法的实现
	n.mergeInterfaceImplement()
}

// 解析源文件，解析出对应的结构体，接口，函数
func (n *NodeManager) Inspect(file string) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
	if err != nil {
		log.Printf("can't parse file:%s, error:%s", file, err.Error())
		return err
	}

	content, err := ioutil.ReadFile(file)
	if err != nil {
		log.Printf("can't read file:%s content, error:%s", file, err.Error())
		return err
	}

	fileParser := NewFileNode(n, file, f.Name.Name)

	structParser := func(n ast.Node) bool {
		t, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}

		if t.Type == nil {
			return true
		}

		x, ok := t.Type.(*ast.StructType)
		if !ok {
			return true
		}

		fields := make(map[string]string, 0)
		for _, v := range x.Fields.List {
			typeExpr := v.Type
			start := typeExpr.Pos() - 1
			end := typeExpr.End() - 1
			typeInSource := string(content)[start:end]
			// 去掉无用的空格
			typeInSource = strings.Trim(typeInSource, " ")
			if len(v.Names) > 0 {
				// name -> type
				fields[v.Names[0].Name] = typeInSource
			} else {
				// 匿名成员变量
				fields[typeInSource] = typeInSource
			}
		}

		i := x.Pos()
		for content[i] != '\n'  && i >= 0 {
			i--
		}
		structNode := NewStructNode(fileParser, t.Name.Name, string(content)[i + 1: x.End()], fields)

		fileParser.structNodes[structNode.name] = structNode
		return true
	}

	interfaceParser := func(n ast.Node) bool {
		t, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}

		if t.Type == nil {
			return true
		}

		x, ok := t.Type.(*ast.InterfaceType)
		if !ok {
			return true
		}

		methods := make(map[string]string, 0)
		if !x.Incomplete {
			functions := strings.Split(string(content[x.Methods.Opening:x.Methods.Closing]), "\n")
			for _, function := range functions {
				function = strings.Trim(function, "\t")
				if strings.Contains(function, "(") && strings.Contains(function, ")") {
					name := strings.Split(function, "(")[0]
					methods[name] = function
				}
			}
		}
		interfaceNode := NewInterfaceNode(fileParser, t.Name.Name, string(content[t.Pos() - 1: t.End()]), methods)
		fileParser.interfaceNodes[t.Name.Name] = interfaceNode

		for name, _ := range methods {
			functionNode := NewFunctionNode(fileParser, name, t.Name.Name,  string(content[t.Pos() - 1: t.End()]), []string{},  []string{})
			if _, ok := fileParser.functionNodes[functionNode.name]; ok == false {
				fileParser.functionNodes[functionNode.name] = make([]*FunctionNode, 0)
			}
			fileParser.functionNodes[functionNode.name] = append(fileParser.functionNodes[functionNode.name], functionNode)
		}

		return true
	}

	functionParser := func(n ast.Node) bool {
		x, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}
		receiver := ""

		if x.Recv != nil {
			receiver = string(content)[int(x.Recv.Opening) : int(x.Recv.Closing)-1]
			elems := strings.Split(receiver, " ")
			if len(elems) == 2 {
				receiver = elems[1]
				receiver = strings.TrimLeft(receiver, "*")
			}
		}
		parameters := make([]string, 0)
		returns := make([]string, 0)

		// 设置对应的参数和返回值
		if x.Type.Params != nil && x.Type.Params.List != nil {
			for _, param := range x.Type.Params.List {
				paramType := string(content)[int(param.Type.Pos())-1 : int(param.Type.End())-1]
				paramType = strings.Trim(paramType, " ")
				paramType = strings.TrimLeft(paramType, "*")
				parameters = append(parameters, paramType)
			}
		}

		if x.Type.Results != nil && x.Type.Results.List != nil {
			for _, result := range x.Type.Results.List {
				resultType := string(content)[int(result.Type.Pos())-1 : int(result.Type.End())]
				resultType = strings.Trim(resultType, " ")
				resultType = strings.TrimLeft(resultType, "*")
				returns = append(returns, resultType)
			}
		}

		//content := string(content)[int(x.Body.Lbrace)-1 : int(x.Body.Rbrace)]
		functionNode := NewFunctionNode(fileParser, x.Name.Name,  receiver, string(content[x.Pos() - 1: x.End()]), parameters, returns)
		if _, ok := fileParser.functionNodes[functionNode.name]; ok == false {
			fileParser.functionNodes[functionNode.name] = make([]*FunctionNode, 0)
		}
		fileParser.functionNodes[functionNode.name] = append(fileParser.functionNodes[functionNode.name], functionNode)
		return true
	}

	importParser := func(n ast.Node) bool {
		x, ok := n.(*ast.ImportSpec)
		if !ok {
			return true
		}
		if x.Name != nil {
			fileParser.importers[x.Name.Name] = x.Path.Value
		} else {
			elems := strings.Split("/", x.Path.Value)
			var name string
			if len(elems) == 1 {
				name = x.Path.Value
			} else {
				name = elems[len(elems) - 1]
			}
			fileParser.importers[name] = x.Path.Value
		}
		return true
	}

	ast.Inspect(f, interfaceParser)

	ast.Inspect(f, structParser)

	ast.Inspect(f, functionParser)

	ast.Inspect(f, importParser)

	if _, ok := n.packages[f.Name.Name]; !ok {
		n.packages[f.Name.Name] = make([]*FileNode, 0)
	}

	n.packages[f.Name.Name] = append(n.packages[f.Name.Name], fileParser)

	return nil
}
