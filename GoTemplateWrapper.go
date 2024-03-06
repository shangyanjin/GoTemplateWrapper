package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	extension = ".html"
	outDir    = "out"
)

func main() {
	overwrite := flag.Bool("f", false, "Overwrite original files")
	modifyDir := flag.Bool("d", false, "Modify relative directory references")
	addTemplateTag := flag.Bool("t", false, "Add template tags")
	flag.Parse()

	// 获取当前工作目录
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		return
	}

	// 创建输出目录
	if !*overwrite {
		err = os.MkdirAll(outDir, os.ModePerm)
		if err != nil {
			fmt.Println("Error creating output directory:", err)
			return
		}
	}

	// 遍历当前目录下的所有文件
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println("Error walking directory:", err)
			return nil
		}

		// 忽略输出目录
		if !*overwrite && strings.HasPrefix(path, filepath.Join(dir, outDir)) {
			return nil
		}

		// 只处理扩展名为.html的文件
		if !info.IsDir() && strings.HasSuffix(info.Name(), extension) {
			modifyFile(path, *overwrite, *modifyDir, *addTemplateTag)
		}

		return nil
	})

	if err != nil {
		fmt.Println("Error walking directory:", err)
	}
}

func modifyFile(filename string, overwrite, modifyDir, addTemplateTag bool) {
	// 读取文件内容
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	// 正则表达式匹配相对路径引用
	pattern := `(?i)(href|src)="([^"]+?)"`
	regex := regexp.MustCompile(pattern)

	// 修改相对路径引用
	modifiedContent := regex.ReplaceAllFunc(content, func(match []byte) []byte {
		parts := regex.FindSubmatch(match)
		if len(parts) > 2 && !bytes.HasPrefix(parts[2], []byte("http")) {
			// 如果相对路径是“#”，则不进行处理
			if bytes.HasPrefix(parts[2], []byte("#")) {
				return match
			}
			if modifyDir {
				return []byte(fmt.Sprintf("%s=\"static/%s\"", parts[1], parts[2]))
			}
		}
		return match
	})

	// 处理模板标签
	contentStr := trimSpaceNewLine(string(modifiedContent))
	// 移除现有的模板标签
	contentStr = removeTemplateTag(contentStr)

	if addTemplateTag {
		dir, file := filepath.Split(filename)
		dir = filepath.Base(filepath.ToSlash(dir))
		file = strings.TrimSuffix(file, filepath.Ext(file))
		defineStatement := fmt.Sprintf("{{define \"%s/%s\"}}\n", dir, file)
		//defineStatement := fmt.Sprintf("{{define \"%s/%s\"}}\n", strings.TrimPrefix(dir, filepath.Join(dir, "../")), strings.TrimSuffix(file, filepath.Ext(file)))
		endStatement := "\n{{end}}"
		contentStr = defineStatement + contentStr + endStatement
	}

	// 获取输出文件路径
	outPath := filename
	if !overwrite {
		outPath = filepath.Join(outDir, filepath.Base(filename))
	}

	// 写入输出文件
	err = ioutil.WriteFile(outPath, []byte(contentStr), 0644)
	if err != nil {
		fmt.Println("Error writing file:", err)
		return
	}

	fmt.Printf("File %s modified successfully\n", outPath)
}

func trimSpaceNewLine(s string) string {
	return strings.TrimSpace(s)
}

func removeTemplateTag(s string) string {
	re := regexp.MustCompile(`(?i)^{{define\s+["']?[^"']+["']?}}\n?|^{{end}}\n?`)
	return re.ReplaceAllString(s, "")
}
