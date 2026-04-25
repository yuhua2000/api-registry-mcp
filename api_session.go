package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type SessionContext struct {
	ID        string
	Endpoints map[string]RegisterEndpointRequest // key: api_name:method:path
	CreatedAt time.Time
	Path      string
}

func NewSessionContext(sessionID string) *SessionContext {
	return &SessionContext{
		ID:        sessionID,
		Endpoints: make(map[string]RegisterEndpointRequest),
		CreatedAt: time.Now(),
		Path:      filepath.Join(dir, sessionID+".md"),
	}
}

// RegisterEndpoint 注册一个新的API端点
func (s *SessionContext) RegisterEndpoint(endpoint RegisterEndpointRequest) error {
	if endpoint.APIName == "" {
		return fmt.Errorf("API名称不能为空")
	}

	if endpoint.Path == "" {
		return fmt.Errorf("路径不能为空")
	}
	if endpoint.Method == "" {
		return fmt.Errorf("请求方法不能为空")
	}

	key := fmt.Sprintf("%s:%s:%s", endpoint.APIName, endpoint.Method, endpoint.Path)
	s.Endpoints[key] = endpoint
	slog.Info("注册端点", "key", key)

	// 将端点信息追加到Markdown文件
	if err := s.appendToMarkdown(endpoint); err != nil {
		slog.Error("写入Markdown失败", "session", s.ID, "err", err)
	} else {
		slog.Info("写入Markdown成功", "session", s.ID, "api", endpoint.APIName, "endpoint", endpoint.Path)
	}

	return nil
}

// GetRegisteredEndpoints 获取所有已注册的端点
func (s *SessionContext) GetRegisteredEndpoints() []RegisterEndpointRequest {
	endpoints := make([]RegisterEndpointRequest, 0, len(s.Endpoints))
	for _, endpoint := range s.Endpoints {
		endpoints = append(endpoints, endpoint)
	}
	return endpoints
}

// GetEndpointsByAPI 获取特定API的所有端点
func (s *SessionContext) GetEndpointsByAPI(apiName, apiVersion string) []RegisterEndpointRequest {
	var result []RegisterEndpointRequest
	for _, endpoint := range s.Endpoints {
		if endpoint.APIName == apiName {
			result = append(result, endpoint)
		}
	}
	return result
}

// appendToMarkdown 将端点信息追加到Markdown文件
func (s *SessionContext) appendToMarkdown(endpoint RegisterEndpointRequest) error {
	// ✅ 1. 确保目录存在
	dir := filepath.Dir(s.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// ✅ 2. 打开文件（追加模式）
	file, err := os.OpenFile(s.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	// 使用提取的方法生成单个端点的Markdown内容
	doc := generateEndpointMarkdown(endpoint)

	// 写入文件
	if _, err := file.WriteString(doc); err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	return nil
}

// GenerateAPIDocs 生成完整的API文档
func (s *SessionContext) GenerateAPIDocs() (string, error) {
	// 创建新的Markdown文件
	dir, err := os.MkdirTemp("", "api_docs")
	if err != nil {
		return "", err
	}

	file, err := os.Create(filepath.Join(dir, "api_docs.md"))
	if err != nil {
		return "", fmt.Errorf("创建文件失败: %v", err)
	}
	defer file.Close()

	// 写入文件头
	header := `# API 文档

本文档由OpenAPI Registry MCP服务自动生成。

## 目录

`

	endpoints := s.GetRegisteredEndpoints()

	// 按API名称分组
	apiGroups := make(map[string][]RegisterEndpointRequest)
	for _, endpoint := range endpoints {
		apiGroups[endpoint.APIName] = append(apiGroups[endpoint.APIName], endpoint)
	}

	// 生成目录
	for apiName, apiEndpoints := range apiGroups {
		header += fmt.Sprintf("- [%s](#%s)\n", apiName, apiName)
		for _, endpoint := range apiEndpoints {
			header += fmt.Sprintf("  - [%s %s](#%s-%s)\n",
				endpoint.Method, endpoint.Path,
				endpoint.Method, endpoint.Path)
		}
	}

	header += "\n---\n\n"

	if _, err := file.WriteString(header); err != nil {
		return "", fmt.Errorf("写入文件头失败: %v", err)
	}

	// 为每个API生成文档
	for apiName, apiEndpoints := range apiGroups {
		apiDoc := fmt.Sprintf("## %s\n\n", apiName)

		for _, endpoint := range apiEndpoints {
			apiDoc += generateEndpointMarkdown(endpoint)
		}

		if _, err := file.WriteString(apiDoc); err != nil {
			return "", fmt.Errorf("写入API文档失败: %v", err)
		}
	}

	return filepath.Join(dir, "api_docs.md"), nil
}

// generateEndpointMarkdown 生成单个API端点的Markdown文档
func generateEndpointMarkdown(endpoint RegisterEndpointRequest) string {
	doc := fmt.Sprintf(`# %s API

## 端点: %s %s 

**描述:** %s

**注册时间:** %s
`,
		endpoint.APIName,
		endpoint.Method,
		endpoint.Path,
		endpoint.Description,
		time.Now().Format("2006-01-02 15:04:05"),
	)

	// ========================
	// 请求参数（query/path/header）
	// ========================
	if len(endpoint.Params) > 0 {
		doc += `
### 请求参数

| 参数名 | 位置 | 类型 | 必填 | 描述 | 默认值 |
|--------|------|------|------|------|--------|
`

		for _, param := range endpoint.Params {
			required := "否"
			if param.Required {
				required = "是"
			}

			defaultValue := "-"
			if param.Default != nil {
				data, err := json.Marshal(param.Default)
				if err != nil {
					defaultValue = fmt.Sprintf("%v", param.Default)
				} else {
					defaultValue = string(data)
				}
			}

			in := param.In
			if in == "" {
				in = "-"
			}

			doc += fmt.Sprintf("| %s | %s | %s | %s | %s | %s |\n",
				param.Name,
				in,
				param.Type,
				required,
				param.Description,
				defaultValue,
			)
		}
	}

	// ========================
	// 请求体（body）
	// ========================
	if len(endpoint.RequestBody) > 0 {
		doc += "\n### 请求体\n"

		if endpoint.RequestBodyFormat != "" {
			doc += fmt.Sprintf("\n**Content-Type:** %s\n\n", endpoint.RequestBodyFormat)
		}

		doc += "| 字段名 | 类型 | 必填 | 描述 | 默认值 |\n"
		doc += "|--------|------|------|------|--------|\n"

		for _, param := range endpoint.RequestBody {
			required := "否"
			if param.Required {
				required = "是"
			}

			defaultValue := "-"
			if param.Default != nil {
				data, err := json.Marshal(param.Default)
				if err != nil {
					defaultValue = fmt.Sprintf("%v", param.Default)
				} else {
					defaultValue = string(data)
				}
			}

			doc += fmt.Sprintf("| %s | %s | %s | %s | %s |\n",
				param.Name,
				param.Type,
				required,
				param.Description,
				defaultValue,
			)
		}
	}

	if endpoint.ParamExample != "" {
		doc += "\n### 请求示例\n\n"

		lang, content := formatExample(endpoint.ParamExample)

		if lang != "" {
			doc += fmt.Sprintf("```%s\n%s\n```\n", lang, content)
		} else {
			doc += fmt.Sprintf("```\n%s\n```\n", content)
		}
	}

	// ========================
	// 响应
	// ========================
	if endpoint.ResponseFormat != "" {
		doc += fmt.Sprintf("\n### 响应格式\n\n%s\n", endpoint.ResponseFormat)
	}

	if len(endpoint.ResponseFields) > 0 {
		doc += "\n### 响应字段\n\n"
		doc += "| 字段名 | 类型 | 必返 | 描述 | 示例 |\n"
		doc += "|--------|------|------|------|------|\n"

		for _, field := range endpoint.ResponseFields {
			required := "否"
			if field.Required {
				required = "是"
			}

			example := "-"
			if field.Example != nil {
				data, err := json.Marshal(field.Example)
				if err != nil {
					example = fmt.Sprintf("%v", field.Example)
				} else {
					example = string(data)
				}
			}

			doc += fmt.Sprintf("| %s | %s | %s | %s | %s |\n",
				field.Name,
				field.Type,
				required,
				field.Description,
				example,
			)
		}
	}

	// ========================
	// 响应示例
	// ========================
	if endpoint.ResponseExample != "" {
		doc += "\n### 响应示例\n\n"

		lang, content := formatExample(endpoint.ResponseExample)

		if lang != "" {
			doc += fmt.Sprintf("```%s\n%s\n```\n", lang, content)
		} else {
			doc += fmt.Sprintf("```\n%s\n```\n", content)
		}
	}

	doc += "\n---\n\n"

	return doc
}

func formatExample(example string) (lang string, content string) {
	example = strings.TrimSpace(example)

	var obj interface{}
	if err := json.Unmarshal([]byte(example), &obj); err == nil {
		// JSON → 格式化
		data, err := json.MarshalIndent(obj, "", "  ")
		if err == nil {
			return "json", string(data)
		}
	}

	// fallback：非 JSON
	return "", example
}
