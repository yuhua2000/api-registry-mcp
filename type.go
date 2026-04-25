package main

type RegisterEndpointRequest struct {
	APIName     string `json:"api_name" jsonschema:"API名称，例如：用户管理API、订单API"`
	Description string `json:"description,omitempty" jsonschema:"API端点描述"`

	Path   string `json:"path" jsonschema:"API端点路径，例如：/users、/orders/{id}"`
	Method string `json:"method" jsonschema:"HTTP请求方法：GET、POST、PUT、DELETE、PATCH"`

	// 请求参数（query/path/header）
	Params []ParamDefinition `json:"params,omitempty" jsonschema:"请求参数定义列表（query/path/header）"`

	// 请求体（单独语义，不建议混在 Params）
	RequestBody       []ParamDefinition `json:"request_body,omitempty" jsonschema:"请求体字段定义（body参数）"`
	RequestBodyFormat string            `json:"request_body_format,omitempty" jsonschema:"请求体格式（HTTP Content-Type），例如：application/json"`
	ParamExample      string            `json:"param_example,omitempty" jsonschema:"请求参数示例"`

	// 响应
	ResponseFormat  string          `json:"response_format,omitempty" jsonschema:"响应格式：json、xml、text"`
	ResponseFields  []ResponseField `json:"response_fields,omitempty" jsonschema:"响应字段定义"`
	ResponseExample string          `json:"response_example,omitempty" jsonschema:"响应示例"`
}

// ParamDefinition 定义API请求参数
type ParamDefinition struct {
	Name string `json:"name" jsonschema:"参数名称"`
	// 参数位置（对齐 OpenAPI: in）
	In string `json:"in,omitempty" jsonschema:"参数位置：query、path、body、header"`
	// 数据类型（对齐 OpenAPI: schema.type）
	Type        string      `json:"type" jsonschema:"参数数据类型：string、number、integer、boolean、array、object"`
	Required    bool        `json:"required,omitempty" jsonschema:"是否必填"`
	Description string      `json:"description,omitempty" jsonschema:"参数描述"`
	Default     interface{} `json:"default,omitempty" jsonschema:"默认值"`
}

// ResponseField 定义响应字段的结构
type ResponseField struct {
	Name        string `json:"name" jsonschema:"字段名称"`
	Type        string `json:"type" jsonschema:"字段数据类型：string、number、integer、boolean、array、object"`
	Description string `json:"description,omitempty" jsonschema:"字段描述"`
	// 可选：示例值（非常有用）
	Example interface{} `json:"example,omitempty" jsonschema:"字段示例值"`
	// 可选：是否必返回（很多接口其实不是全返回）
	Required bool `json:"required,omitempty" jsonschema:"是否必定返回"`
}

// EndpointListItem 端点列表项，简化版本的端点信息
type EndpointListItem struct {
	APIName string `json:"api_name" jsonschema:"API名称"`
	BaseURL string `json:"base_url" jsonschema:"API基础URL"`
	Path    string `json:"path" jsonschema:"API端点路径"`
}

// ListEndpointsResponse 定义list_endpoints工具的响应结构体
type ListEndpointsResponse struct {
	Total int                `json:"total" jsonschema:"总端点数量"`
	List  []EndpointListItem `json:"list" jsonschema:"端点列表"`
}

// RegisterEndpointResponse 定义register_endpoint工具的响应结构体
type RegisterEndpointResponse struct {
	Success bool   `json:"success" jsonschema:"操作是否成功"`
	Message string `json:"message" jsonschema:"操作结果消息"`

	// 核心返回（轻量信息，方便展示）
	Endpoint EndpointSummary `json:"endpoint" jsonschema:"端点摘要信息"`

	// 可选：扩展信息
	Meta *RegisterMeta `json:"meta,omitempty" jsonschema:"附加信息"`
}

type EndpointSummary struct {
	APIName string `json:"api_name"`
	Method  string `json:"method"`
	Path    string `json:"path"`

	ParamCount       int `json:"param_count"`
	RequestBodyCount int `json:"request_body_count"`
	ResponseFields   int `json:"response_fields_count"`
}

type RegisterMeta struct {
	Timestamp string `json:"timestamp"`
}

// GenerateDocsRequest 定义generate_docs工具的请求结构体
type GenerateDocsRequest struct {
}

// GenerateDocsResponse 定义generate_docs工具的响应结构体
type GenerateDocsResponse struct {
	Success   bool   `json:"success" jsonschema:"操作是否成功"`
	Message   string `json:"message" jsonschema:"操作结果消息"`
	FilePath  string `json:"file_path" jsonschema:"生成的文档文件路径"`
	FileSize  int64  `json:"file_size" jsonschema:"文件大小（字节）"`
	Timestamp string `json:"timestamp" jsonschema:"生成时间戳"`
}
