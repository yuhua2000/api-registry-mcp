# OpenAPI Registry MCP Service

[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go)](https://golang.org/)
[![MCP Version](https://img.shields.io/badge/MCP-v0.49.0-blue)](https://github.com/mark3labs/mcp-go)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

一个基于 MCP (Model Context Protocol) 的 OpenAPI 端点注册和文档生成服务。通过自动代码审计来识别、注册 API 端点，并生成标准化文档。

## ✨ 功能特性

- 🔍 **自动代码审计** - 扫描源代码识别 API 端点
- 📝 **结构化注册** - 按 OpenAPI 标准注册 API 端点
- 📚 **文档生成** - 自动生成 Markdown 格式的 API 文档
- 🔧 **会话管理** - 支持多会话并发处理
- 🚀 **实时响应** - 基于 SSE (Server-Sent Events) 提供实时通信

## 📋 核心工具

### 1. 端点注册工具 (`register_endpoint`)
用于在代码审计过程中，将发现的 API 端点按统一规范进行注册。

**参数要求:**
- `api_name`: API 业务名称（如 "用户管理API"）
- `description`: API 功能描述
- `path`: API 路径（RESTful 风格）
- `method`: HTTP 方法（GET/POST/PUT/DELETE/PATCH）
- `params`: 请求参数定义
- `request_body`: 请求体字段定义
- `response_fields`: 响应字段定义

### 2. 端点查询工具 (`list_endpoints`)
查看已注册的所有 API 端点列表。

### 3. 文档生成工具 (`generate_docs`)
生成完整的 API 文档报告（Markdown 格式）。

## 🚀 快速开始

### 环境要求
- Go 1.26+
- MCP 客户端（如 Claude Desktop, Cursor, Windsurf 等）

### 安装与运行

```bash
# 克隆项目
git clone https://github.com/yuhua2000/api-registry-mcp.git
cd api-registry-mcp

# 安装依赖
go mod download

# 运行服务
go run main.go
```

服务将在默认端口 `8080` 启动。您可以使用以下命令测试服务：

```bash
# 测试服务是否正常运行
curl http://localhost:8080/
```

### MCP 客户端配置

在您的 MCP 客户端配置文件中添加：

**Claude Desktop 配置 (claude_desktop_config.json):**
```json
{
  "mcpServers": {
    "api-registry-mcp": {
      "url": "http://localhost:8080"
    }
  }
}
```

## 🏗️ 项目结构

```
api-registry-mcp/
├── main.go              # 主程序入口，MCP 服务器配置
├── type.go             # 数据类型定义（请求/响应结构）
├── manager.go          # 会话管理器，处理工具请求
├── api_session.go      # 会话上下文，实现业务逻辑
├── go.mod              # Go 模块定义
├── go.sum              # 依赖校验
├── skills.md           # MCP 技能定义文档
└── data/               # 数据存储目录（自动生成）
```

## 🤖 AI 技能使用

项目包含完整的 MCP 技能定义 (`skills.md`)，供 AI 代理自动化 API 文档生成：

**核心流程：**
1. **代码审计** - AI 扫描项目代码，识别所有 API 端点
2. **自动注册** - 按 OpenAPI 标准注册发现的端点  
3. **文档生成** - 一键生成完整的 Markdown API 文档

**支持框架：** Python (Flask/FastAPI), JavaScript (Express/NestJS), Go (Gin/Echo), Java (Spring Boot), C# (ASP.NET Core)

**使用方式：** 配置 MCP 服务器后，AI 代理即可自动调用此技能完成 API 文档生成。

## 🔧 API 示例

### 注册 API 端点

```json
{
  "api_name": "用户管理API",
  "description": "获取单个用户信息",
  "path": "/users/{id}",
  "method": "GET",
  "params": [
    {
      "name": "id",
      "in": "path",
      "type": "integer",
      "required": true,
      "description": "用户ID"
    }
  ],
  "response_format": "application/json",
  "response_fields": [
    {
      "name": "id",
      "type": "integer",
      "required": true,
      "description": "用户ID",
      "example": 123
    },
    {
      "name": "name",
      "type": "string",
      "required": true,
      "description": "用户名",
      "example": "张三"
    }
  ],
  "response_example": "{\"id\":123,\"name\":\"张三\"}"
}
```

### 查询端点列表

```bash
# 返回所有已注册的端点
{
  "total": 5,
  "list": [
    {
      "api_name": "用户管理API",
      "path": "/users"
    },
    {
      "api_name": "用户管理API",
      "path": "/users/{id}"
    }
  ]
}
```

### 生成的文档示例

生成的 Markdown 文档包含：
- API 概述和描述
- 请求参数表格
- 请求体和响应字段定义
- 完整的请求/响应示例

## 🛠️ 开发指南

### 添加新的工具

1. 在 `main.go` 中定义工具描述
2. 在 `type.go` 中添加对应的请求/响应结构体
3. 在 `manager.go` 中实现处理函数
4. 在 `api_session.go` 中实现具体业务逻辑

### 测试

```bash
# 运行测试（需要添加测试文件）
go test ./...
```

### 构建

```bash
# 构建可执行文件
go build -o api-registry-mcp main.go

# 运行构建后的程序
./api-registry-mcp
```

## 📊 数据存储

注册的 API 端点数据以 JSON 格式存储在 `data/` 目录下：
- 每个会话一个文件：`data/{session_id}.json`
- 自动生成的文档存储在临时目录中

## 🤝 贡献

欢迎贡献！请遵循以下步骤：

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 📝 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🙏 致谢

- [MCP-Go](https://github.com/mark3labs/mcp-go) - MCP 协议的 Go 实现
- 所有贡献者和用户

## 📞 支持

如果您遇到问题或有建议：
- 在 [GitHub Issues](https://github.com/yuhua2000/api-registry-mcp/issues) 中报告问题
- 提供详细的错误描述和复现步骤

---

**Made with ❤️ for the MCP community**

<p align="center">
  ⚡ 用 Go 和 MCP 构建的下一代 API 文档工具 ⚡
</p>
