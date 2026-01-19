package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// MCP Protocol Types
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type InitializeResult struct {
	ProtocolVersion string            `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      ServerInfo        `json:"serverInfo"`
}

type ServerCapabilities struct {
	Tools     *ToolsCapability     `json:"tools,omitempty"`
	Resources *ResourcesCapability `json:"resources,omitempty"`
}

type ToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

type ResourcesCapability struct {
	Subscribe   bool `json:"subscribe,omitempty"`
	ListChanged bool `json:"listChanged,omitempty"`
}

type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"inputSchema"`
}

type InputSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties,omitempty"`
	Required   []string            `json:"required,omitempty"`
}

type Property struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
}

type ToolsListResult struct {
	Tools []Tool `json:"tools"`
}

type CallToolParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

type CallToolResult struct {
	Content []ContentBlock `json:"content"`
	IsError bool           `json:"isError,omitempty"`
}

type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// API Key configuration
type APIKeyConfig struct {
	EnvVar      string `json:"env_var"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

// Available API keys configuration
var apiKeyConfigs = map[string]APIKeyConfig{
	// LLM APIs
	"openai": {
		EnvVar:      "OPENAI_API_KEY",
		Description: "OpenAI API key for GPT models",
		Category:    "llm",
	},
	"anthropic": {
		EnvVar:      "ANTHROPIC_API_KEY",
		Description: "Anthropic API key for Claude models",
		Category:    "llm",
	},
	"google_ai": {
		EnvVar:      "GOOGLE_AI_API_KEY",
		Description: "Google AI API key for Gemini models",
		Category:    "llm",
	},
	"cohere": {
		EnvVar:      "COHERE_API_KEY",
		Description: "Cohere API key",
		Category:    "llm",
	},
	// SaaS APIs
	"stripe": {
		EnvVar:      "STRIPE_API_KEY",
		Description: "Stripe API key for payments",
		Category:    "saas",
	},
	"stripe_webhook": {
		EnvVar:      "STRIPE_WEBHOOK_SECRET",
		Description: "Stripe webhook signing secret",
		Category:    "saas",
	},
	"twilio_sid": {
		EnvVar:      "TWILIO_ACCOUNT_SID",
		Description: "Twilio Account SID",
		Category:    "saas",
	},
	"twilio_token": {
		EnvVar:      "TWILIO_AUTH_TOKEN",
		Description: "Twilio Auth Token",
		Category:    "saas",
	},
	"sendgrid": {
		EnvVar:      "SENDGRID_API_KEY",
		Description: "SendGrid API key for emails",
		Category:    "saas",
	},
	"aws_access_key": {
		EnvVar:      "AWS_ACCESS_KEY_ID",
		Description: "AWS Access Key ID",
		Category:    "saas",
	},
	"aws_secret_key": {
		EnvVar:      "AWS_SECRET_ACCESS_KEY",
		Description: "AWS Secret Access Key",
		Category:    "saas",
	},
	// Canva
	"canva_client_id": {
		EnvVar:      "CANVA_CLIENT_ID",
		Description: "Canva OAuth Client ID",
		Category:    "canva",
	},
	"canva_client_secret": {
		EnvVar:      "CANVA_CLIENT_SECRET",
		Description: "Canva OAuth Client Secret",
		Category:    "canva",
	},
	"canva_app_id": {
		EnvVar:      "CANVA_APP_ID",
		Description: "Canva App ID",
		Category:    "canva",
	},
	// Custom/Internal
	"database_url": {
		EnvVar:      "DATABASE_URL",
		Description: "Database connection string",
		Category:    "internal",
	},
	"redis_url": {
		EnvVar:      "REDIS_URL",
		Description: "Redis connection URL",
		Category:    "internal",
	},
	"jwt_secret": {
		EnvVar:      "JWT_SECRET",
		Description: "JWT signing secret",
		Category:    "internal",
	},
	"app_secret": {
		EnvVar:      "APP_SECRET",
		Description: "Application secret key",
		Category:    "internal",
	},
}

type MCPServer struct {
	scanner *bufio.Scanner
}

func NewMCPServer() *MCPServer {
	// Load .env file if it exists (for local development)
	godotenv.Load()

	return &MCPServer{
		scanner: bufio.NewScanner(os.Stdin),
	}
}

func (s *MCPServer) sendResponse(response JSONRPCResponse) {
	data, _ := json.Marshal(response)
	fmt.Println(string(data))
}

func (s *MCPServer) sendError(id interface{}, code int, message string) {
	s.sendResponse(JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &RPCError{
			Code:    code,
			Message: message,
		},
	})
}

func (s *MCPServer) handleInitialize(id interface{}) {
	s.sendResponse(JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result: InitializeResult{
			ProtocolVersion: "2024-11-05",
			Capabilities: ServerCapabilities{
				Tools: &ToolsCapability{
					ListChanged: false,
				},
			},
			ServerInfo: ServerInfo{
				Name:    "api-keys-server",
				Version: "1.0.0",
			},
		},
	})
}

func (s *MCPServer) handleToolsList(id interface{}) {
	// Build enum of available key names
	keyNames := make([]string, 0, len(apiKeyConfigs))
	for name := range apiKeyConfigs {
		keyNames = append(keyNames, name)
	}

	// Build enum of categories
	categories := []string{"llm", "saas", "canva", "internal", "all"}

	tools := []Tool{
		{
			Name:        "get_api_key",
			Description: "Retrieve an API key by its name. Returns the API key value from environment variables.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"key_name": {
						Type:        "string",
						Description: "The name of the API key to retrieve (e.g., 'openai', 'stripe', 'canva_client_id')",
						Enum:        keyNames,
					},
				},
				Required: []string{"key_name"},
			},
		},
		{
			Name:        "list_api_keys",
			Description: "List all available API key names and their descriptions. Does not return actual key values.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"category": {
						Type:        "string",
						Description: "Filter by category: 'llm', 'saas', 'canva', 'internal', or 'all'",
						Enum:        categories,
					},
				},
				Required: []string{},
			},
		},
		{
			Name:        "check_api_key_exists",
			Description: "Check if an API key is configured (has a value set) without revealing the key itself.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"key_name": {
						Type:        "string",
						Description: "The name of the API key to check",
						Enum:        keyNames,
					},
				},
				Required: []string{"key_name"},
			},
		},
	}

	s.sendResponse(JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result: ToolsListResult{
			Tools: tools,
		},
	})
}

func (s *MCPServer) handleToolCall(id interface{}, params CallToolParams) {
	switch params.Name {
	case "get_api_key":
		s.handleGetAPIKey(id, params.Arguments)
	case "list_api_keys":
		s.handleListAPIKeys(id, params.Arguments)
	case "check_api_key_exists":
		s.handleCheckAPIKeyExists(id, params.Arguments)
	default:
		s.sendError(id, -32601, fmt.Sprintf("Unknown tool: %s", params.Name))
	}
}

func (s *MCPServer) handleGetAPIKey(id interface{}, args map[string]interface{}) {
	keyName, ok := args["key_name"].(string)
	if !ok {
		s.sendResponse(JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      id,
			Result: CallToolResult{
				Content: []ContentBlock{{Type: "text", Text: "Error: key_name is required"}},
				IsError: true,
			},
		})
		return
	}

	config, exists := apiKeyConfigs[keyName]
	if !exists {
		s.sendResponse(JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      id,
			Result: CallToolResult{
				Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error: Unknown API key name: %s", keyName)}},
				IsError: true,
			},
		})
		return
	}

	value := os.Getenv(config.EnvVar)
	if value == "" {
		s.sendResponse(JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      id,
			Result: CallToolResult{
				Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("API key '%s' is not configured. Set the %s environment variable.", keyName, config.EnvVar)}},
				IsError: true,
			},
		})
		return
	}

	s.sendResponse(JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result: CallToolResult{
			Content: []ContentBlock{{Type: "text", Text: value}},
		},
	})
}

func (s *MCPServer) handleListAPIKeys(id interface{}, args map[string]interface{}) {
	category := "all"
	if cat, ok := args["category"].(string); ok && cat != "" {
		category = cat
	}

	var result strings.Builder
	result.WriteString("Available API Keys:\n\n")

	categories := map[string]string{
		"llm":      "🤖 LLM APIs",
		"saas":     "☁️ SaaS APIs",
		"canva":    "🎨 Canva APIs",
		"internal": "🔧 Internal/Custom",
	}

	for cat, title := range categories {
		if category != "all" && category != cat {
			continue
		}

		result.WriteString(fmt.Sprintf("%s:\n", title))
		for name, config := range apiKeyConfigs {
			if config.Category == cat {
				configured := "❌"
				if os.Getenv(config.EnvVar) != "" {
					configured = "✅"
				}
				result.WriteString(fmt.Sprintf("  %s %s - %s (env: %s)\n", configured, name, config.Description, config.EnvVar))
			}
		}
		result.WriteString("\n")
	}

	s.sendResponse(JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result: CallToolResult{
			Content: []ContentBlock{{Type: "text", Text: result.String()}},
		},
	})
}

func (s *MCPServer) handleCheckAPIKeyExists(id interface{}, args map[string]interface{}) {
	keyName, ok := args["key_name"].(string)
	if !ok {
		s.sendResponse(JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      id,
			Result: CallToolResult{
				Content: []ContentBlock{{Type: "text", Text: "Error: key_name is required"}},
				IsError: true,
			},
		})
		return
	}

	config, exists := apiKeyConfigs[keyName]
	if !exists {
		s.sendResponse(JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      id,
			Result: CallToolResult{
				Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Error: Unknown API key name: %s", keyName)}},
				IsError: true,
			},
		})
		return
	}

	value := os.Getenv(config.EnvVar)
	if value != "" {
		// Mask the key value for security
		masked := value[:4] + "..." + value[len(value)-4:]
		if len(value) < 12 {
			masked = "****"
		}
		s.sendResponse(JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      id,
			Result: CallToolResult{
				Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("✅ API key '%s' is configured (value: %s)", keyName, masked)}},
			},
		})
	} else {
		s.sendResponse(JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      id,
			Result: CallToolResult{
				Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("❌ API key '%s' is NOT configured. Set %s environment variable.", keyName, config.EnvVar)}},
			},
		})
	}
}

func (s *MCPServer) Run() {
	for s.scanner.Scan() {
		line := s.scanner.Text()
		if line == "" {
			continue
		}

		var request JSONRPCRequest
		if err := json.Unmarshal([]byte(line), &request); err != nil {
			s.sendError(nil, -32700, "Parse error")
			continue
		}

		switch request.Method {
		case "initialize":
			s.handleInitialize(request.ID)
		case "initialized":
			// Notification, no response needed
		case "tools/list":
			s.handleToolsList(request.ID)
		case "tools/call":
			var params CallToolParams
			if err := json.Unmarshal(request.Params, &params); err != nil {
				s.sendError(request.ID, -32602, "Invalid params")
				continue
			}
			s.handleToolCall(request.ID, params)
		default:
			// For unknown methods, just acknowledge if it has an ID
			if request.ID != nil {
				s.sendError(request.ID, -32601, fmt.Sprintf("Method not found: %s", request.Method))
			}
		}
	}
}

func main() {
	server := NewMCPServer()
	server.Run()
}
