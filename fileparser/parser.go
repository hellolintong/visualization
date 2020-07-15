package fileparser

type Parser interface {
	Merge()
	Draw()
	Inspect(file string) error
}

func NewParser(detail bool, allField bool) Parser {
	return &NodeManager{
		packages:      make(map[string][]*FileNode, 0),
		structTypes:   make([]string, 0),
		functionNames: make(map[string][]string, 0),
		interfaceNames: make(map[string]bool, 0),
		detail: detail,
		allField: allField,
	}
}