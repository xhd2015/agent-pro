package doc

import (
	_ "embed"
	"fmt"
)

//go:embed DOC_STYLE_TEST_SPECIFICATION.md
var docStyleTestSpecification string

//go:embed DOC_STYLE_TEST_CODE_SPECIFICATION.md
var docStyleTestCodeSpecification string

func Content(fileName string) (string, error) {
	switch fileName {
	case "DOC_STYLE_TEST_SPECIFICATION.md":
		return docStyleTestSpecification, nil
	case "DOC_STYLE_TEST_CODE_SPECIFICATION.md":
		return docStyleTestCodeSpecification, nil
	default:
		return "", fmt.Errorf("unknown file: %s", fileName)
	}
}
