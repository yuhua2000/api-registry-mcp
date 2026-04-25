package main

import (
	"context"
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const RegisterEndpointToolDescription = `
注册 OpenAPI 风格的 API 端点（基于代码静态分析）

## 作用
将代码中识别到的 API 端点进行结构化注册，形成统一 API 资产，用于接口治理、文档生成与安全审计。

## 输入来源（自动解析代码）
系统会从源码中自动提取以下信息：

### API 基础信息
- api_name：业务模块名称（如 用户管理API）
- description：接口功能说明（基于代码语义推导）

### 路由信息
- path：RESTful 路径（如 /users/{id}）
- method：HTTP 方法（GET / POST / PUT / DELETE / PATCH）

### 请求参数（params）
适用于 query / path / header 参数：
- name：参数名（必须与代码一致）
- in：参数位置（必须明确）
- type：数据类型（string / integer / boolean / array / object）
- required：是否必填
- description：参数说明
- default：默认值（可选）

### 请求体（request_body）
仅适用于 body 类型请求：
- request_body：结构化请求体字段
- request_body_format：请求体格式（必须使用标准 MIME）

支持格式：
- application/json
- application/xml
- application/x-www-form-urlencoded
- multipart/form-data
- application/octet-stream
- text/plain

### 请求示例
- param_example：请求示例（必须与真实参数结构一致，支持多行 JSON 格式）

## 输出结果
注册后的 API 端点结构，包含：
- 完整请求参数
- 请求体结构
- 响应结构定义

### 响应信息
- response_format：响应格式（如 application/json）
- response_fields：响应字段定义（重点结构）
	字段说明
	  - name：字段名称（必须与实际返回一致）
	  - type：字段类型（string / integer / boolean / array / object）
	  - description：字段说明（基于真实语义）
- response_example：真实响应示例（必须与真实参数结构一致，支持多行 JSON 格式）

## 使用场景
- 自动 API 审计
- OpenAPI 文档生成
- 微服务接口资产沉淀
- 接口规范治理与校验

## 强约束
- 所有数据必须来源代码，不允许猜测或补全
- params 与 request_body 必须严格分离
- path 必须与代码路由完全一致
- method 仅允许标准 HTTP 方法
- request_body_format 必须使用标准 MIME 类型
- 必须逐个端点注册，不允许跳过或合并
- 支持重复注册：相同 path + method 会覆盖旧数据
- 支持暂停 / 继续执行
- 支持中途上下文压缩（避免长上下文影响执行）
`

var (
	httpAddr = ":8080" // 默认HTTP端口
	dir      = "data"
)

func main() {
	s := NewSessionManager()

	hooks := &server.Hooks{}
	hooks.AddOnRegisterSession(func(ctx context.Context, session server.ClientSession) {
		sessionID := session.SessionID()

		s.CreateSession(sessionID)

		slog.Info("session create", "sessionID", sessionID)
	})

	hooks.AddOnUnregisterSession(func(ctx context.Context, session server.ClientSession) {
		sessionID := session.SessionID()

		s.Delete(sessionID)

		slog.Info("session deleted", "sessionID", sessionID)
	})

	// 创建MCP服务器
	mcpServer := server.NewMCPServer(
		"OpenAPI Registry Service",
		"1.0.0",
		server.WithHooks(hooks),
	)

	// 注册工具：开放API格式注册接口
	registerAPITool := mcp.NewTool(
		"register_endpoint",
		mcp.WithDescription(RegisterEndpointToolDescription),
		mcp.WithInputSchema[RegisterEndpointRequest](),
		mcp.WithOutputSchema[RegisterEndpointResponse](),
	)

	mcpServer.AddTool(registerAPITool, s.handleRegisterEndpoint)

	// 注册工具：已注册开放API端点查询工具
	listEndpointsTool := mcp.NewTool(
		"list_endpoints",
		mcp.WithDescription("查询所有已注册的开放API端点"),
		mcp.WithInputSchema[map[string]interface{}](), // 无参数
		mcp.WithOutputSchema[ListEndpointsResponse](),
	)

	mcpServer.AddTool(listEndpointsTool, s.handleListEndpoints)

	// 注册工具：生成API文档
	generateDocsTool := mcp.NewTool(
		"generate_docs",
		mcp.WithDescription(`生成已注册API的完整文档报告
	
	使用此工具可以生成所有已注册API端点的完整文档报告，包括：
	- 所有API端点的详细描述
	- 请求参数表格
	- 响应格式和字段说明
	- 参数和响应示例`),
		mcp.WithInputSchema[GenerateDocsRequest](),
		mcp.WithOutputSchema[GenerateDocsResponse](),
	)

	mcpServer.AddTool(generateDocsTool, s.handleGenerateDocs)

	slog.Info("📋 可用工具:")
	slog.Info("   1. register_endpoint - 注册开放API格式端点")
	slog.Info("   2. list_endpoints - 查询已注册API端点列表")
	slog.Info("   3. generate_docs - 生成API文档报告")
	// 启动服务器
	slog.Info("🚀 启动OpenAPI端点注册MCP服务...")

	slog.Info("🌐 SSE服务器监听地址", "addr", httpAddr)

	sseServer := server.NewSSEServer(mcpServer)
	_ = sseServer.Start(httpAddr)
}
