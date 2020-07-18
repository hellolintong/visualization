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
	"strings"
)

type NodeManager struct {
	packages        map[string][]*FileNode
	pointedPackages map[string]bool
	pointedStructs  map[string]bool
	structTypes     map[string]map[string]bool
	functionNames   map[string]map[string]bool
	interfaceNames  map[string]map[string]bool
	detail          bool
	allField        bool
}

func (n *NodeManager) getFunctionReceiverLabel(receiver string) string {
	temp := n.detail
	n.detail = false
	result := n.getStructReceiverLabel(receiver)
	n.detail = temp
	return result
}

func (n *NodeManager) getStructReceiverLabel(receiver string) string {
	for _, fileNodes := range n.packages {
		for _, fileNode := range fileNodes {
			if structNode, ok := fileNode.structNodes[receiver]; ok == true {
				return structNode.getStructLabel(n.detail)
			}
		}
	}
	return ""
}

func (n *NodeManager) drawStruct() {
	content := bytes.NewBuffer([]byte{})
	content.WriteString("digraph gph {")

	// 绘制struct节点
	record := map[string]bool{}
	for _, package_ := range n.packages {
		for _, filenode := range package_ {
			filenode.DrawStructNode(content, record)
		}
	}

	// 绘制interface节点
	record = map[string]bool{}
	for _, package_ := range n.packages {
		for _, filenode := range package_ {
			filenode.DrawInterfaceNode(content, record)
		}
	}

	// 绘制struct节点关系
	for _, package_ := range n.packages {
		for _, filenode := range package_ {
			filenode.DrawStructRelation(content, record)
		}
	}

	// 绘制interface节点
	for _, package_ := range n.packages {
		for _, filenode := range package_ {
			filenode.DrawInterfaceRelation(content, record)
		}
	}

	content.WriteString("}")

	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}

	os.MkdirAll(path+"/data/", 0755)

	if err := ioutil.WriteFile(path+"/data/struct_visualization.dot", content.Bytes(), 0644); err == nil {
		if err = sh.Command("/bin/bash", "-c", fmt.Sprintf("dot %s/data/struct_visualization.dot -o %s/data/struct_visualization.png -Tpng", path, path)).Run(); err == nil {
			log.Println("draw success!")
		} else {
			log.Printf("draw visualization.dot fail, error:%s", err.Error())
		}
	} else {
		log.Printf("write visualization.dot fail, error:%s", err.Error())
	}
}

func (n *NodeManager) drawFunction() {
	content := bytes.NewBuffer([]byte{})
	content.WriteString("digraph gph {")

	receiver := make(map[string]bool, 0)
	for _, package_ := range n.packages {
		for _, filenode := range package_ {
			filenode.DrawFunctionNode(content, receiver)
		}
	}

	record := map[string]bool{}
	for _, package_ := range n.packages {
		for _, filenode := range package_ {
			filenode.DrawFunctionRelation(content, record)
		}
	}

	content.WriteString("}")

	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}

	os.MkdirAll(path+"/data/", 0755)

	if err := ioutil.WriteFile(path+"/data/function_visualization.dot", content.Bytes(), 0644); err == nil {
		if err = sh.Command("/bin/bash", "-c", fmt.Sprintf("dot %s/data/function_visualization.dot -o %s/data/function_visualization.png -Tpng", path, path)).Run(); err == nil {
			log.Println("draw success!")
		} else {
			log.Printf("draw visualization.dot fail, error:%s", err.Error())
		}
	} else {
		log.Printf("write visualization.dot fail, error:%s", err.Error())
	}
}

func (n *NodeManager) Draw() {
	n.drawStruct()
	n.drawFunction()
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
			filenode.MergeFunction(n.structTypes, n.interfaceNames)
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
					n.structTypes[filenode.packageName] = make(map[string]bool, 0)
				}
				n.structTypes[filenode.packageName][structNode.name] = true
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
					n.interfaceNames[filenode.packageName] = make(map[string]bool, 0)
				}
				n.interfaceNames[filenode.packageName][interfaceNode.name] = true
			}
		}
	}

	// 归并
	for _, package_ := range n.packages {
		for _, filenode := range package_ {
			filenode.MergeStruct(n.structTypes, n.interfaceNames)
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

func (n *NodeManager) Inspect(file string) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
	if err != nil {
		log.Printf("can't parse file:%s, error:%s", file, err.Error())
		return err
	}

	if len(n.pointedPackages) != 0 {
		if _, ok := n.pointedPackages[f.Name.Name]; ok == false {
			return nil
		}
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
				fields[v.Names[0].Name] = typeInSource
			} else {
				// 匿名成员变量
				fields[typeInSource] = typeInSource
			}
		}
		structNode := NewStructNode(fileParser, t.Name.Name, fields)

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
		interfaceNode := NewInterfaceNode(fileParser, t.Name.Name, methods)
		fileParser.interfaceNodes[t.Name.Name] = interfaceNode

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

		body := string(content)[int(x.Body.Lbrace)-1 : int(x.Body.Rbrace)]
		functionNode := NewFunctionNode(fileParser, x.Name.Name, receiver, body, parameters, returns)
		fileParser.functionNodes[functionNode.name] = functionNode
		return true
	}

	ast.Inspect(f, interfaceParser)

	ast.Inspect(f, structParser)

	ast.Inspect(f, functionParser)

	if _, ok := n.packages[f.Name.Name]; !ok {
		n.packages[f.Name.Name] = make([]*FileNode, 0)
	}

	n.packages[f.Name.Name] = append(n.packages[f.Name.Name], fileParser)

	return nil
}
