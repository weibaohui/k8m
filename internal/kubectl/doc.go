package kubectl

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
)

// TreeNode 定义树形节点结构，包含更多字段
type TreeNode struct {
	Name        string      `json:"name"`
	Type        string      `json:"type,omitempty"`
	Children    []*TreeNode `json:"children,omitempty"`
	Description string      `json:"description,omitempty"`
	Required    bool        `json:"required,omitempty"`
	Default     interface{} `json:"default,omitempty"`
}

// CacheData 定义缓存的数据结构
type CacheData struct {
	Schema map[string]interface{} `json:"schema"`
}

var (
	cache     *CacheData
	cacheOnce sync.Once
	cacheErr  error
	cacheMux  sync.RWMutex
)

// Doc 函数用于生成并输出树形结构的 JSON 数据
func Doc() {
	// 设置日志格式
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// 获取 OpenAPI schemas，支持内存缓存
	schema, err := GetOpenAPISchemasWithCache()
	if err != nil {
		log.Fatalf("Failed to get OpenAPI schemas: %v", err)
	}

	// 构建树形结构，处理引用
	tree, err := BuildTreeWithRefs(schema)
	if err != nil {
		log.Fatalf("Failed to build tree: %v", err)
	}

	// 输出树形结构为 JSON
	treeJSON, err := json.MarshalIndent(tree, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal tree: %v", err)
	}

	fmt.Println(string(treeJSON))
}

// GetOpenAPISchemasWithCache 获取 OpenAPI schemas，支持内存缓存
func GetOpenAPISchemasWithCache() (map[string]interface{}, error) {
	cacheOnce.Do(func() {
		log.Println("Fetching OpenAPI schema from Kubernetes API")
		schema, err := GetOpenAPISchemas()
		if err != nil {
			cacheErr = err
			return
		}

		cacheMux.Lock()
		cache = &CacheData{
			Schema: schema,
		}
		cacheMux.Unlock()

		log.Println("OpenAPI schema cached in memory")
	})

	cacheMux.RLock()
	defer cacheMux.RUnlock()

	if cache == nil {
		return nil, cacheErr
	}

	log.Println("Loading OpenAPI schema from in-memory cache")
	return cache.Schema, nil
}

// GetOpenAPISchemas 获取 Kubernetes OpenAPI 规范中的所有 schema
func GetOpenAPISchemas() (map[string]interface{}, error) {

	discoveryClient := kubectl.client.Discovery()

	// 获取 OpenAPI 规范
	openAPISchema, err := discoveryClient.OpenAPISchema()
	if err != nil {
		return nil, fmt.Errorf("failed to get OpenAPI schema: %v", err)
	}

	// 将 OpenAPISchema 转换为原始 JSON
	raw, err := json.Marshal(openAPISchema)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal OpenAPI schema: %v", err)
	}

	// 解析 JSON
	var schema map[string]interface{}
	if err := json.Unmarshal(raw, &schema); err != nil {
		return nil, fmt.Errorf("failed to unmarshal OpenAPI schema: %v", err)
	}

	return schema, nil
}

// BuildTreeWithRefs 构建树形结构，处理引用
func BuildTreeWithRefs(schema map[string]interface{}) ([]*TreeNode, error) {
	// 适配不同 Kubernetes 版本的 OpenAPI 规范结构
	definitions, ok := schema["definitions"].(map[string]interface{})
	if !ok {
		components, ok := schema["components"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("no definitions or components found in OpenAPI schema")
		}
		definitions, ok = components["schemas"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("no schemas found in components")
		}
	}

	// 使用同步映射存储已解析的引用，避免重复解析
	var mutex sync.Mutex
	resolvedRefs := make(map[string]*TreeNode)

	var trees []*TreeNode
	var wg sync.WaitGroup
	var parseErr error
	sem := make(chan struct{}, 10) // 并发限制，防止过多并发

	for name, def := range definitions {
		defMap, ok := def.(map[string]interface{})
		if !ok {
			continue
		}

		wg.Add(1)
		sem <- struct{}{}
		go func(name string, defMap map[string]interface{}) {
			defer wg.Done()
			defer func() { <-sem }()

			node, err := parseDefinition(name, defMap, definitions, resolvedRefs, &mutex)
			if err != nil {
				log.Printf("Failed to parse definition %s: %v", name, err)
				parseErr = err
				return
			}

			mutex.Lock()
			trees = append(trees, node)
			mutex.Unlock()
		}(name, defMap)
	}

	wg.Wait()
	if parseErr != nil {
		return nil, parseErr
	}

	return trees, nil
}

// parseDefinition 解析单个定义，处理引用
func parseDefinition(name string, defMap map[string]interface{}, definitions map[string]interface{}, resolvedRefs map[string]*TreeNode, mutex *sync.Mutex) (*TreeNode, error) {
	node := &TreeNode{
		Name: name,
	}

	// 处理类型
	if typ, ok := defMap["type"].(string); ok {
		node.Type = typ
	}

	// 处理描述
	if desc, ok := defMap["description"].(string); ok {
		node.Description = desc
	}

	// 处理 required 字段
	requiredFields := make(map[string]bool)
	if required, ok := defMap["required"].([]interface{}); ok {
		for _, field := range required {
			if fieldStr, ok := field.(string); ok {
				requiredFields[fieldStr] = true
			}
		}
	}

	// 递归处理 properties
	if properties, ok := defMap["properties"].(map[string]interface{}); ok {
		children, err := parseProperties(properties, requiredFields, definitions, resolvedRefs, mutex)
		if err != nil {
			return nil, err
		}
		node.Children = children
	}

	// 处理数组类型的项
	if typ, ok := defMap["type"].(string); ok && typ == "array" {
		if items, ok := defMap["items"].(map[string]interface{}); ok {
			if itemProps, ok := items["properties"].(map[string]interface{}); ok {
				children, err := parseProperties(itemProps, requiredFields, definitions, resolvedRefs, mutex)
				if err != nil {
					return nil, err
				}
				node.Children = children
			}
		}
	}

	return node, nil
}

// parseProperties 递归解析 properties，构建子树，并处理引用
func parseProperties(properties map[string]interface{}, requiredFields map[string]bool, definitions map[string]interface{}, resolvedRefs map[string]*TreeNode, mutex *sync.Mutex) ([]*TreeNode, error) {
	var nodes []*TreeNode
	var wg sync.WaitGroup
	var parseErr error
	sem := make(chan struct{}, 20) // 并发限制

	for propName, prop := range properties {
		propMap, ok := prop.(map[string]interface{})
		if !ok {
			continue
		}

		wg.Add(1)
		sem <- struct{}{}
		go func(propName string, propMap map[string]interface{}) {
			defer wg.Done()
			defer func() { <-sem }()

			node := &TreeNode{
				Name:     propName,
				Required: requiredFields[propName],
			}

			// 处理字段类型
			if typ, ok := propMap["type"].(string); ok {
				node.Type = typ
			}

			// 处理引用类型
			if ref, ok := propMap["$ref"].(string); ok {
				refName := extractRefName(ref)
				if refName != "" {
					// 检查是否已经解析过
					mutex.Lock()
					if existing, found := resolvedRefs[refName]; found {
						node.Type = existing.Name
						node.Description = existing.Description
						node.Children = existing.Children
						mutex.Unlock()
					} else {
						mutex.Unlock()
						// 查找定义并解析
						def, exists := definitions[refName]
						if !exists {
							log.Printf("Reference %s not found in definitions", refName)
							return
						}
						defMap, ok := def.(map[string]interface{})
						if !ok {
							log.Printf("Definition for %s is not a map", refName)
							return
						}
						childNode, err := parseDefinition(refName, defMap, definitions, resolvedRefs, mutex)
						if err != nil {
							log.Printf("Failed to parse referenced definition %s: %v", refName, err)
							return
						}

						// 添加到引用映射中
						mutex.Lock()
						resolvedRefs[refName] = childNode
						mutex.Unlock()

						node.Type = childNode.Type
						node.Description = childNode.Description
						node.Children = childNode.Children
					}
				}
			}

			// 处理描述
			if desc, ok := propMap["description"].(string); ok {
				node.Description = desc
			}

			// 处理默认值
			if defVal, ok := propMap["default"]; ok {
				node.Default = defVal
			}

			// 递归处理子属性
			if subProps, ok := propMap["properties"].(map[string]interface{}); ok {
				children, err := parseProperties(subProps, requiredFields, definitions, resolvedRefs, mutex)
				if err != nil {
					parseErr = err
					return
				}
				node.Children = children
			}

			// 处理数组类型的项
			if typ, ok := propMap["type"].(string); ok && typ == "array" {
				if items, ok := propMap["items"].(map[string]interface{}); ok {
					if itemProps, ok := items["properties"].(map[string]interface{}); ok {
						children, err := parseProperties(itemProps, requiredFields, definitions, resolvedRefs, mutex)
						if err != nil {
							parseErr = err
							return
						}
						node.Children = children
					}
				}
			}

			// 添加节点到结果
			mutex.Lock()
			nodes = append(nodes, node)
			mutex.Unlock()
		}(propName, propMap)
	}

	wg.Wait()
	if parseErr != nil {
		return nil, parseErr
	}

	return nodes, nil
}

// extractRefName 提取引用名称，例如 "#/definitions/io.k8s.api.core.v1.Pod" 提取 "io.k8s.api.core.v1.Pod"
func extractRefName(ref string) string {
	parts := strings.Split(ref, "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}
