package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type SessionManager struct {
	mu       sync.RWMutex
	sessions map[string]*SessionContext
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*SessionContext),
	}
}

func (s *SessionManager) CreateSession(sessionID string) *SessionContext {
	ctx := NewSessionContext(sessionID)

	s.mu.Lock()
	s.sessions[sessionID] = ctx
	s.mu.Unlock()

	return ctx
}

func (s *SessionManager) GetOrCreate(sessionID string) *SessionContext {
	// fast path（读锁）
	s.mu.RLock()
	ctx, ok := s.sessions[sessionID]
	s.mu.RUnlock()

	if ok {
		return ctx
	}

	// slow path：不在锁内创建
	newCtx := NewSessionContext(sessionID)

	s.mu.Lock()
	defer s.mu.Unlock()

	// double check
	if ctx, ok = s.sessions[sessionID]; ok {
		return ctx
	}

	s.sessions[sessionID] = newCtx
	return newCtx
}

func (s *SessionManager) Delete(sessionID string) {
	s.mu.Lock()
	ctx, ok := s.sessions[sessionID]
	if ok {
		delete(s.sessions, sessionID)
	}
	s.mu.Unlock()

	if !ok {
		return
	}

	// 保存 session 数据（锁外执行）
	if err := s.persistSession(ctx); err != nil {
		slog.Error("保存 session 失败",
			"sessionID", sessionID,
			"error", err,
		)
	}
}

// handleRegisterEndpoint 处理register_endpoint工具请求
func (s *SessionManager) handleRegisterEndpoint(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	session := server.ClientSessionFromContext(ctx)
	slog.Info("处理注册端点请求", "sessionID", session.SessionID())
	sessionContext := s.GetOrCreate(session.SessionID())

	// 从请求中解析参数
	var req RegisterEndpointRequest

	if err := request.BindArguments(&req); err != nil {
		slog.Error("解析请求参数失败", "session", session.SessionID(), "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("解析请求参数失败: %v", err)), nil
	}

	// 注册端点
	if err := sessionContext.RegisterEndpoint(req); err != nil {
		slog.Error("注册端点失败", "session", session.SessionID(), "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("注册端点失败: %v", err)), nil
	}

	// 构建成功响应
	response := RegisterEndpointResponse{
		Success: true,
		Message: fmt.Sprintf("✅ 成功注册端点: %s %s %s",
			req.APIName, req.Method, req.Path),

		Endpoint: EndpointSummary{
			APIName: req.APIName,
			Method:  req.Method,
			Path:    req.Path,

			ParamCount:       len(req.Params),
			RequestBodyCount: len(req.RequestBody),
			ResponseFields:   len(req.ResponseFields),
		},

		Meta: &RegisterMeta{
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		},
	}

	// 记录注册成功
	slog.Info("成功注册端点", "session", session.SessionID(), "APIName", req.APIName, "Method", req.Method, "Path", req.Path, "ParamsCount", len(req.Params))

	// 返回成功响应
	return mcp.NewToolResultJSON(response)
}

// handleListEndpoints 处理list_endpoints工具请求
func (s *SessionManager) handleListEndpoints(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	session := server.ClientSessionFromContext(ctx)
	slog.Info("处理拉取断点列表请求", "sessionID", session.SessionID())

	sessionContext := s.GetOrCreate(session.SessionID())

	// 获取所有已注册的端点
	endpoints := sessionContext.GetRegisteredEndpoints()

	// 构建简化的端点列表项
	endpointList := make([]EndpointListItem, 0, len(endpoints))
	for _, endpoint := range endpoints {
		endpointList = append(endpointList, EndpointListItem{
			APIName: endpoint.APIName,
			Path:    endpoint.Path,
		})
	}

	// 构建响应对象
	response := ListEndpointsResponse{
		Total: len(endpointList),
		List:  endpointList,
	}

	// 记录查询到的端点数量
	slog.Info("查询到端点列表", "session", session.SessionID(), "total", len(endpointList))

	// 返回JSON响应
	return mcp.NewToolResultJSON(response)
}

// handleGenerateDocs 处理generate_docs工具请求
func (s *SessionManager) handleGenerateDocs(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	session := server.ClientSessionFromContext(ctx)
	slog.Info("处理生成断点文档请求", "sessionID", session.SessionID())

	sessionContext := s.GetOrCreate(session.SessionID())

	// 调用GenerateAPIDocs方法
	filePath, err := sessionContext.GenerateAPIDocs()
	if err != nil {
		slog.Error("生成API文档失败", "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("生成API文档失败: %v", err)), nil
	}

	// 获取文件信息
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		slog.Error("获取文件信息失败", "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("获取文件信息失败: %v", err)), nil
	}

	// 构建成功响应
	response := GenerateDocsResponse{
		Success:   true,
		Message:   "✅ 成功生成API文档，格式: markdown",
		FilePath:  filePath,
		FileSize:  fileInfo.Size(),
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
	}

	// 记录文档生成成功
	slog.Info("成功生成API文档", "filePath", filePath, "fileSize", fileInfo.Size())

	// 返回成功响应
	return mcp.NewToolResultJSON(response)

}

func (s *SessionManager) persistSession(ctx *SessionContext) error {
	// 确保目录存在
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 文件路径：data/{sessionID}.json
	filePath := filepath.Join(dir, ctx.ID+".json")

	// JSON pretty（方便排查问题）
	data, err := json.MarshalIndent(ctx, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化失败: %w", err)
	}

	// 写文件（覆盖写）
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("写文件失败: %w", err)
	}

	slog.Info("session 已持久化", "sessionID", ctx.ID, "path", filePath, "endpoints", len(ctx.Endpoints))

	return nil
}
