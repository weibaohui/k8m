package kubectl

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/weibaohui/k8m/internal/utils"
)

type Docs struct {
	Trees []TreeNode
}

func NewDocs() *Docs {
	d := &Docs{}
	if len(trees) == 0 {
		initDoc()
	}
	d.Trees = trees
	return d
}

var trees []TreeNode

// TreeNode 表示树形结构的节点
type TreeNode struct {
	ID              string      `json:"id"`
	Label           string      `json:"label"`
	Description     string      `json:"description,omitempty"`
	Type            string      `json:"type,omitempty"`
	Ref             string      `json:"ref,omitempty"`
	VendorExtension interface{} `json:"vendor_extension,omitempty"`
	Children        []TreeNode  `json:"children,omitempty"`
}

// SchemaDefinition 表示根定义
type SchemaDefinition struct {
	Name  string      `json:"name"`
	Value SchemaValue `json:"value"`
}

// SchemaValue 表示定义的值
type SchemaValue struct {
	Description string           `json:"description"`
	Properties  SchemaProperties `json:"properties"`
	Type        SchemaType       `json:"type"`
	// VendorExtension []VendorExtension `json:"vendor_extension,omitempty"`
}

// SchemaProperties 表示属性
type SchemaProperties struct {
	AdditionalProperties []Property `json:"additional_properties"`
}

// Property 表示单个属性
type Property struct {
	Name  string        `json:"name"`
	Value PropertyValue `json:"value"`
}

// PropertyValue 表示属性的值
type PropertyValue struct {
	Description string      `json:"description,omitempty"`
	Type        *SchemaType `json:"type,omitempty"`
	Ref         string      `json:"_ref,omitempty"`
}

// SchemaType 表示类型
type SchemaType struct {
	Value []string `json:"value"`
}

// definitionsMap 存储所有定义，以便处理引用
var definitionsMap map[string]SchemaDefinition

// parseOpenAPISchema 解析 OpenAPI Schema JSON 字符串并返回根 TreeNode
func parseOpenAPISchema(schemaJSON string) (TreeNode, error) {
	var def SchemaDefinition
	err := json.Unmarshal([]byte(schemaJSON), &def)
	if err != nil {
		return TreeNode{}, err
	}

	rootDef := def
	return buildTree(rootDef), nil
}

// buildTree 根据 SchemaDefinition 构建 TreeNode
func buildTree(def SchemaDefinition) TreeNode {
	labelParts := strings.Split(def.Name, ".")
	label := labelParts[len(labelParts)-1]

	nodeType := ""
	if len(def.Value.Type.Value) > 0 {
		nodeType = def.Value.Type.Value[0]
	}

	var children []TreeNode
	for _, prop := range def.Value.Properties.AdditionalProperties {
		children = append(children, buildPropertyNode(prop))
	}

	return TreeNode{
		ID:          def.Name,
		Label:       label,
		Description: def.Value.Description,
		Type:        nodeType,
		Children:    children,
	}
}

// buildPropertyNode 根据 Property 构建 TreeNode
func buildPropertyNode(prop Property) TreeNode {
	label := prop.Name
	nodeID := prop.Name
	description := prop.Value.Description
	nodeType := ""
	ref := ""

	if prop.Value.Type != nil && len(prop.Value.Type.Value) > 0 {
		nodeType = prop.Value.Type.Value[0]
	}
	if prop.Value.Ref != "" {
		ref = prop.Value.Ref
	}

	var children []TreeNode

	// 如果有引用，查找定义并递归构建子节点
	if ref != "" {
		// 假设 ref 的格式为 "#/definitions/io.k8s.apimachinery.pkg.apis.meta.v1.ObjectMeta"
		refParts := strings.Split(ref, "/")
		refName := refParts[len(refParts)-1]
		// 构建完整的引用路径
		fullRef := strings.Join(refParts[1:], ".")

		if def, exists := definitionsMap[fullRef]; exists {
			childNode := buildTree(def)
			children = append(children, childNode)
		} else {
			// 如果引用的定义不存在，可以记录为一个叶子节点或处理为需要进一步扩展
			children = append(children, TreeNode{
				ID:          ref,
				Label:       refName,
				Description: "Referenced definition not found",
			})
		}
	}

	return TreeNode{
		ID:          nodeID,
		Label:       label,
		Description: description,
		Type:        nodeType,
		Ref:         ref,
		Children:    children,
	}
}

// printTree 递归打印 TreeNode
func printTree(node TreeNode, level int) {
	indent := strings.Repeat("  ", level)
	fmt.Printf("%s%s (ID: %s)\n", indent, node.Label, node.ID)
	if node.Description != "" {
		fmt.Printf("%s  Description: %s\n", indent, node.Description)
	}
	if node.Type != "" {
		fmt.Printf("%s  Type: %s\n", indent, node.Type)
	}
	if node.Ref != "" {
		fmt.Printf("%s  Ref: %s\n", indent, node.Ref)
	}

	for _, child := range node.Children {
		printTree(child, level+1)
	}
}

func initDoc() {

	// 获取 OpenAPI Schema
	openAPISchema, err := kubectl.client.DiscoveryClient.OpenAPISchema()
	if err != nil {
		fmt.Printf("Error fetching OpenAPI schema: %v\n", err)
		os.Exit(1)
	}

	// 将 OpenAPI Schema 转换为 JSON 字符串
	schemaBytes, err := json.Marshal(openAPISchema)
	if err != nil {
		fmt.Printf("Error marshaling OpenAPI schema to JSON: %v\n", err)
		os.Exit(1)
	}
	// os.WriteFile("def.json", schemaBytes, 0644)
	// 打印部分 Schema 以供调试
	// fmt.Println(string(schemaBytes))

	// 解析 OpenAPI Schema 并构建 TreeNode 结构
	// 由于客户端返回的 OpenAPI Schema 格式与之前的硬编码 JSON 不同，
	// 我们需要提取 "definitions" 部分并转换为所需的格式。

	var openapiSchema map[string]interface{}
	err = json.Unmarshal(schemaBytes, &openapiSchema)
	if err != nil {
		fmt.Printf("Error unmarshaling OpenAPI schema: %v\n", err)
		os.Exit(1)
	}
	// 提取 swagger 下的第一个 definitions
	definitionsSw, ok := openapiSchema["definitions"].(map[string]interface{})
	if !ok {
		fmt.Printf("No definitions found in OpenAPI schema top level\n")
		os.Exit(1)
	}

	definitionList, ok := definitionsSw["additional_properties"].([]interface{})

	if !ok {
		fmt.Printf("No definitions found in OpenAPI schema\n")
		os.Exit(1)
	}

	for _, item := range definitionList {
		definition, ok := item.(map[string]interface{})
		jsonUtils := utils.JSONUtils{}
		jstr := jsonUtils.ToJSON(definition)
		if !ok {
			fmt.Printf("convert definition error\n")
			os.Exit(1)
		}

		// 解析 Schema 并构建树形结构
		treeRoot, err := parseOpenAPISchema(jstr)
		if err != nil {
			fmt.Printf("Error parsing OpenAPI schema: %v\n", err)
			os.Exit(1)
		}
		trees = append(trees, treeRoot)
		// // 打印树形结构
		// printTree(treeRoot, 0)

	}
}

func (d *Docs) ListNames() {
	for _, tree := range d.Trees {
		log.Println(tree.Label)
	}
}
