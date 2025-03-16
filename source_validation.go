package grader

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

func (g *Grader) GetHeaderC_Cpp() ([]string, error) {
	fileName := g.sourceFile
	fileType := g.typeFile
	if fileType != "c" && fileType != "cpp" {
		return nil, fmt.Errorf("❌ Error file type %s: %v", fileType)
	}
	code, err := os.ReadFile(filepath.Join(g.pathOfSource, fileName+"."+fileType))
	if err != nil {
		return nil, fmt.Errorf("❌ Error reading source code %s: %v", fileName, err)
	}
	includeRegex := regexp.MustCompile(`#include\s*[<"]([^>"]+)[>"]`)

	matches := includeRegex.FindAllStringSubmatch(string(code), -1)

	// แสดงรายการ Header ที่พบ
	// fmt.Println("🔍 Header C, C++ in " + fileName + " Files :")
	// fmt.Println(len(matches))
	var headerList []string
	for _, match := range matches {
		headerList = append(headerList, match[1])
		// fmt.Println(" -", match[1])
	}
	// fmt.Print(headerList)
	return headerList, nil
}

func (g *Grader) ValidationSourceCodeByRegex(regexCmd string) ([]string, error) {
	fileName := g.sourceFile
	fileType := g.typeFile

	code, err := os.ReadFile(filepath.Join(g.pathOfSource, fileName+"."+fileType))
	if err != nil {
		return nil, fmt.Errorf("❌ Error reading source code %s: %v", fileName, err)
	}
	// includeRegex := regexp.MustCompile(`#include\s*[<"]([^>"]+)[>"]`)
	includeRegex := regexp.MustCompile(regexCmd)

	matches := includeRegex.FindAllStringSubmatch(string(code), -1)
	var regex_list []string
	// แสดงรายการ Header ที่พบ
	// fmt.Println("🔍 Files RegEx in " + fileName + " :")
	for _, match := range matches {
		// fmt.Println(" -", match[1])
		regex_list = append(regex_list, match[1])
	}
	return regex_list, nil
}
