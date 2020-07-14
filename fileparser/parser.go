package fileparser

type Parser interface {
	Merge()
	Draw()
	Inspect(file string) error
}

func NewParser(detail bool) Parser {
	return &NodeManager{
		packages:      make(map[string][]*FileNode, 0),
		structTypes:   make([]string, 0),
		functionNames: make(map[string][]string, 0),
		detail: detail,
	}
}