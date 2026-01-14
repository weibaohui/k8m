package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/weibaohui/k8m/pkg/plugins/modules/inspection/models"
)

func main() {
	outDir := flag.String("out", "pkg/plugins/modules/inspection/doc", "输出目录（相对项目根目录或绝对路径）")
	flag.Parse()

	absOutDir, err := filepath.Abs(*outDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "解析输出目录失败: %v\n", err)
		os.Exit(1)
	}

	if err := os.MkdirAll(absOutDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "创建输出目录失败: %v\n", err)
		os.Exit(1)
	}

	scripts := make([]models.InspectionLuaScript, 0, len(models.BuiltinLuaScripts))
	for _, s := range models.BuiltinLuaScripts {
		scripts = append(scripts, s)
	}

	sort.Slice(scripts, func(i, j int) bool {
		return scripts[i].ScriptCode < scripts[j].ScriptCode
	})

	indexItems := make([]indexItem, 0, len(scripts))
	for _, s := range scripts {
		if strings.TrimSpace(s.ScriptCode) == "" {
			fmt.Fprintf(os.Stderr, "跳过 ScriptCode 为空的脚本: %s\n", s.Name)
			continue
		}
		fileName := sanitizeFileName(s.ScriptCode) + ".md"
		outPath := filepath.Join(absOutDir, fileName)
		if err := writeScriptDoc(outPath, s); err != nil {
			fmt.Fprintf(os.Stderr, "写入脚本文档失败: script=%s, err=%v\n", s.ScriptCode, err)
			os.Exit(1)
		}
		indexItems = append(indexItems, indexItem{
			ScriptCode: s.ScriptCode,
			Name:       s.Name,
			FileName:   fileName,
		})
	}

	readmePath := filepath.Join(absOutDir, "README.md")
	if err := writeReadme(readmePath, indexItems); err != nil {
		fmt.Fprintf(os.Stderr, "写入 README.md 失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("已生成 %d 个脚本文档到: %s\n", len(indexItems), absOutDir)
}

type indexItem struct {
	ScriptCode string
	Name       string
	FileName   string
}

func writeReadme(path string, items []indexItem) error {
	var b strings.Builder
	b.WriteString("# 内置巡检 Lua 脚本索引\n\n")
	b.WriteString("说明：本文档索引由程序自动生成，脚本内容以 Go 内置脚本为准。\n\n")
	b.WriteString("## 脚本列表\n\n")
	for _, it := range items {
		b.WriteString(fmt.Sprintf("- [%s | %s](%s)\n", it.ScriptCode, it.Name, it.FileName))
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func writeScriptDoc(path string, s models.InspectionLuaScript) error {
	fence := chooseFence(s.Script)

	var b strings.Builder
	b.WriteString("# " + s.Name + "\n\n")
	b.WriteString("## 介绍\n\n")
	b.WriteString(s.Description + "\n\n")
	b.WriteString("## 信息\n\n")
	b.WriteString(fmt.Sprintf("- ScriptCode: %s\n", s.ScriptCode))
	b.WriteString(fmt.Sprintf("- Kind: %s\n", s.Kind))
	b.WriteString(fmt.Sprintf("- Group: %s\n", s.Group))
	b.WriteString(fmt.Sprintf("- Version: %s\n", s.Version))
	b.WriteString(fmt.Sprintf("- TimeoutSeconds: %d\n\n", s.TimeoutSeconds))
	b.WriteString("## 代码\n\n")
	b.WriteString(fence + "lua\n")
	b.WriteString(s.Script)
	if !strings.HasSuffix(s.Script, "\n") {
		b.WriteString("\n")
	}
	b.WriteString(fence + "\n")

	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func sanitizeFileName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, string(os.PathSeparator), "-")
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, "\\", "-")
	name = strings.ReplaceAll(name, " ", "_")
	re := regexp.MustCompile(`[^0-9A-Za-z._-]+`)
	name = re.ReplaceAllString(name, "_")
	if name == "" {
		return "script"
	}
	return name
}

func chooseFence(code string) string {
	maxRun := 0
	run := 0
	for _, r := range code {
		if r == '`' {
			run++
			if run > maxRun {
				maxRun = run
			}
			continue
		}
		run = 0
	}
	if maxRun < 3 {
		return "```"
	}
	return strings.Repeat("`", maxRun+1)
}

