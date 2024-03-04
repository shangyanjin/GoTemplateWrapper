package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func trimSpaceNewLine(s string) string {
	return strings.TrimSpace(s)
}

// normalizePath 处理路径，确保路径中不包含重复的斜杠，并处理可能的Windows路径问题
func normalizePath(path string) string {
	// 统一路径分隔符为正斜杠
	normalizedPath := strings.ReplaceAll(path, "\\", "/")
	// 使用正则表达式替换所有重复的斜杠为单个斜杠
	re := regexp.MustCompile(`/{2,}`)
	normalizedPath = re.ReplaceAllString(normalizedPath, "/")
	return normalizedPath
}

func processFile(path, rootDir, outDir string) error {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	contentStr := trimSpaceNewLine(string(content))

	// 移除现有的 {{define}} 和 {{end}} 标签
	defineRegex := regexp.MustCompile(`(?m)^{{define .+}}\n?`)
	endRegex := regexp.MustCompile(`(?m)\n?{{end}}$`)
	contentStr = defineRegex.ReplaceAllString(contentStr, "")
	contentStr = endRegex.ReplaceAllString(contentStr, "")

	relPath, _ := filepath.Rel(rootDir, path)
	normalizedRelPath := normalizePath(relPath)
	dir, file := filepath.Split(normalizedRelPath)
	fileNameWithoutExt := strings.TrimSuffix(file, filepath.Ext(file))
	defineStatement := fmt.Sprintf("{{define \"%s/%s\"}}\n", dir, fileNameWithoutExt)
	endStatement := "\n{{end}}"
	//对defineStatement 处理，// 替换为/
	defineStatement = strings.Replace(defineStatement, "//", "/", -1)
	contentStr = defineStatement + contentStr + endStatement

	// 生成输出路径并确保目录结构存在
	outputPath := filepath.Join(outDir, relPath)
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return err
	}

	return ioutil.WriteFile(outputPath, []byte(contentStr), 0644)
}

func main() {
	rootDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		return
	}

	if len(os.Args) > 1 {
		rootDir = os.Args[1]
	}

	outDir := filepath.Join(rootDir, "out")

	fmt.Println("Processing directory:", rootDir)
	fmt.Println("Outputting to:", outDir)

	err = filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".html" && !strings.Contains(path, filepath.Join(string(filepath.Separator), "out", string(filepath.Separator))) {
			fmt.Println("Processing:", path)
			return processFile(path, rootDir, outDir)
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}
}
