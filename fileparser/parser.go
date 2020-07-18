package fileparser

import (
	"strings"
)

type Parser interface {
	Merge()
	Draw()
	Inspect(file string) error
}

func NewParser(detail bool, allField bool, packages []string, structs []string) Parser {
	pointedPackages := map[string]bool{}
	for _, packageName := range packages {
		pointedPackages[packageName] = true
	}

	pointedStructs := map[string]bool{}
	for _, structName := range structs {
		pointedStructs[structName] = true
	}

	return &NodeManager{
		packages:        make(map[string][]*FileNode, 0),
		structTypes:     make(map[string]map[string]bool, 0),
		functionNames:   make(map[string]map[string]bool, 0),
		interfaceNames:  make(map[string]map[string]bool, 0),
		pointedPackages: pointedPackages,
		pointedStructs:  pointedStructs,
		detail:          detail,
		allField:        allField,
	}
}

func getKeyStructType(fieldType string, structType string) string {
	// 除去chan的影响
	if strings.HasPrefix(fieldType, "chan") {
		fieldType = fieldType[len("chan"):]
	}

	fieldType = strings.ReplaceAll(fieldType, " ", "")
	if strings.HasPrefix(fieldType, "map") && strings.Contains(fieldType, "["+structType+"]") || strings.Contains(fieldType, "[*"+structType+"]") {
		return structType
	}
	return ""
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

func typeCompare(structTypes map[string]map[string]bool, interfaceNames map[string]map[string]bool, fieldType string) (string, string) {

	// 先去掉包前缀
	packages := make([]string, 0)
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
						packageName := subFieldType[i+1 : index]
						packages = append(packages, packageName)
						subFieldType = subFieldType[index+1:]
						break
					}
				}
				if !found {
					packageName := subFieldType[:index]
					packages = append(packages, packageName)
					subFieldType = subFieldType[index+1:]
				}
			} else {
				break
			}
		}
	}

	finalType := ""
	keyFinalType := ""

	// 如果有包名，就先以包为标准
	if len(packages) != 0  {
		// 针对每个包
		for _, packageName := range packages {
			if _, ok := structTypes[packageName]; ok == true {
				for structType, _ := range structTypes[packageName] {
					tmp := getStructType(fieldType, packageName + "." + structType)
					if len(tmp) > len(finalType) {
						finalType = tmp
					}
					tmp = getKeyStructType(fieldType, packageName + "." + structType)
					if len(tmp) > len(keyFinalType) {
							keyFinalType = tmp
					}
				}
			}

			if _, ok := interfaceNames[packageName]; ok == true {
				for structType, _ := range interfaceNames[packageName] {
					tmp := getStructType(fieldType, packageName + "." + structType)
					if len(tmp) > len(finalType) {
						finalType = tmp
					}
					tmp = getKeyStructType(fieldType, packageName + "." + structType)
					if len(tmp) > len(keyFinalType) {
						keyFinalType = tmp
					}
				}
			}
		}
	}  else {
		for _, types := range structTypes {
			for structType, _ := range types {
				tmp := getStructType(fieldType, structType)
				if len(tmp) > len(finalType) {
					finalType = tmp
				}
				tmp = getKeyStructType(fieldType, structType)
				if len(tmp) > len(keyFinalType) {
					keyFinalType = tmp
				}
			}
		}

		for _, types := range interfaceNames {
			for structType, _ := range types {
				tmp := getStructType(fieldType, structType)
				if len(tmp) > len(finalType) {
					finalType = tmp
				}
				tmp = getKeyStructType(fieldType, structType)
				if len(tmp) > len(keyFinalType) {
					keyFinalType = tmp
				}
			}
		}
	}
	if strings.Contains(keyFinalType, ".") {
		elems := strings.Split(keyFinalType, ".")
		keyFinalType = elems[len(elems) - 1]
	}
	if strings.Contains(finalType, ".") {
		elems := strings.Split(finalType, ".")
		finalType = elems[len(elems) - 1]
	}
	return keyFinalType, finalType
}
