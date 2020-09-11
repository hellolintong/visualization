package fileparser

import (
	"strings"
)

type Parser interface {
	Merge()
	DrawFunction(baseName string, count int)
	DrawStruct(baseName string, count int)
	GetStructCodeSnippet(baseName string) map[string]string
	GetFunctionCodeSnippet(baseName string) map[string]string
	Inspect(file string) error
	Relation() map[string][]string
}

func NewParser(projectPath string) Parser {

	return &NodeManager{
		projectPath: projectPath,
		packages:        make(map[string][]*FileNode, 0),
		structTypes:     make(map[string]map[string]*StructNode, 0),
		functionNames:   make(map[string]map[string]bool, 0),
		interfaceNames:  make(map[string]map[string]*InterfaceNode, 0),
		allFunctions: make(map[string][]*FunctionNode, 0),
		allStructs: make(map[string]*StructNode, 0),
		knownModuleFunction: make(map[string]bool, 0),
	}
}

func getStructType(fieldType string, structType string) string {
	// 除去chan的影响
	if strings.HasPrefix(fieldType, "chan") {
		fieldType = fieldType[len("chan"):]
	}
	fieldType = strings.ReplaceAll(fieldType, " ", "")

	// 基础类型
	if fieldType == "string" || fieldType == "int" || fieldType == "uint" || fieldType == "uint8" || fieldType == "uint16" || fieldType == "uint32" || fieldType == "uint64" ||
		fieldType == "int8" || fieldType == "int16" || fieldType == "int32" || fieldType == "int64" || fieldType == "float32" ||
		fieldType == "float64" || fieldType == "complex32" || fieldType == "complex64" || fieldType == "byte" || fieldType == "rune" || fieldType == "uintptr" {
		return ""
	}
	// 精准匹配
	if fieldType == structType || fieldType == "*"+structType {
		return structType
	}
	// 取最长匹配（避免像map这种干扰)
	if strings.Contains(fieldType, "]"+structType) || strings.Contains(fieldType, "]*"+structType) {
		return structType
	}
	return ""
}

func typeCompare(structTypes map[string]map[string]*StructNode, interfaceNames map[string]map[string]*InterfaceNode, fieldType string) (*StructNode, *InterfaceNode) {

	// 先去掉包前缀
	var packageName string
	if strings.Contains(fieldType, ".") {
		var subFieldType string
		subFieldType = fieldType
		for {
			index := strings.Index(subFieldType, ".")
			if index != -1 {
				found := false
				for i := index; i >= 0; i-- {
					 if subFieldType[i] == ' ' || subFieldType[i] == '*' || subFieldType[i] == ']' {
					 	found = true
						packageName = subFieldType[i+1 : index]
						subFieldType = subFieldType[index+1:]
						break
					}
				}
				if !found {
					packageName = subFieldType[:index]
					subFieldType = subFieldType[index+1:]
				}
			} else {
				break
			}
		}
	}

	var finalStructTypeStr string

	var finalStructType *StructNode

	var finalInterfaceType *InterfaceNode

	// 如果有包名，就先以包为标准
	if len(packageName) != 0  {
		if _, ok := structTypes[packageName]; ok == true {
			// 找到最合适的匹配点
			for structType, value := range structTypes[packageName] {
				tmp := getStructType(fieldType, packageName + "." + structType)
				if len(tmp) > len(finalStructTypeStr) {
					finalStructTypeStr = tmp
					finalStructType = value
				}
			}
		}

		if _, ok := interfaceNames[packageName]; ok == true {
			for structType, value := range interfaceNames[packageName] {
				tmp := getStructType(fieldType, packageName + "." + structType)
				if len(tmp) > len(finalStructTypeStr) {
					finalStructTypeStr = tmp
					finalInterfaceType = value
				}
			}
		}
	} else {
		for _, types := range structTypes {
			for structType, value := range types {
				tmp := getStructType(fieldType, structType)
				if len(tmp) > len(finalStructTypeStr) {
					finalStructTypeStr = tmp
					finalStructType = value
				}
			}
		}

		for _, types := range interfaceNames {
			for structType, value := range types {
				tmp := getStructType(fieldType, structType)
				if len(tmp) > len(finalStructTypeStr) {
					finalStructTypeStr = tmp
					finalInterfaceType = value
				}
			}
		}
	}

	return finalStructType, finalInterfaceType
}
