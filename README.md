# MCP API Keys Server

A secure MCP (Model Context Protocol) server written in Go that manages and retrieves API keys from environment variables. Runs in Docker for easy deployment and isolation.

## Features

- **Secure API Key Management**: Store keys in environment variables, never in code
- **Multiple API Categories**: LLM, SaaS, Canva, and custom/internal APIs
- **Docker Ready**: Containerized for consistent deployment
- **MCP Protocol Compliant**: Works with Claude Code and other MCP clients

## Available Tools

| Tool | Description |
|------|-------------|
| `get_api_key` | Retrieve an API key by name |
| `list_api_keys` | List all available API keys (without revealing values) |
| `check_api_key_exists` | Check if an API key is configured |

## Supported API Keys

### LLM APIs
- `openai` - OpenAI API key
- `anthropic` - Anthropic API key
- `google_ai` - Google AI API key
- `cohere` - Cohere API key

### SaaS APIs
- `stripe` - Stripe API key
- `stripe_webhook` - Stripe webhook secret
- `twilio_sid` - Twilio Account SID
- `twilio_token` - Twilio Auth Token
- `sendgrid` - SendGrid API key
- `aws_access_key` - AWS Access Key ID
- `aws_secret_key` - AWS Secret Access Key

### Canva APIs
- `canva_client_id` - Canva OAuth Client ID
- `canva_client_secret` - Canva OAuth Client Secret
- `canva_app_id` - Canva App ID

### Internal/Custom
- `database_url` - Database connection string
- `redis_url` - Redis connection URL
- `jwt_secret` - JWT signing secret
- `app_secret` - Application secret key

## Quick Start

### 1. Clone the Repository

```bash
git clone https://github.com/yourusername/mcp-api-keys-server.git
cd mcp-api-keys-server
```

### 2. Configure Environment Variables

```bash
# Copy the example file
cp .env.example .env

# Edit .env with your actual API keys
nano .env  # or your preferred editor
```

### 3. Build and Run with Docker

```bash
# Build the Docker image
docker build -t mcp-api-keys-server .

# Or use docker-compose
docker-compose build
```

### 4. Configure Claude Code

Add to your Claude Code MCP settings (`~/.claude/claude_desktop_config.json` or similar):

```json
{
  "mcpServers": {
    "api-keys": {
      "command": "docker",
      "args": [
        "run",
        "-i",
        "--rm",
        "--env-file", "/path/to/your/.env",
        "mcp-api-keys-server"
      ]
    }
  }
}
```

Or if using docker-compose, create a wrapper script:

```bash
#!/bin/bash
# run-mcp-server.sh
cd /path/to/mcp-api-keys-server
docker-compose run --rm mcp-api-keys
```

Then configure:

```json
{
  "mcpServers": {
    "api-keys": {
      "command": "/path/to/run-mcp-server.sh"
    }
  }
}
```

## Local Development

### Prerequisites
- Go 1.21+
- Docker (optional)

### Run Locally (without Docker)

```bash
# Install dependencies
go mod download

# Create .env file
cp .env.example .env
# Edit .env with your keys

# Build and run
go build -o mcp-server .
./mcp-server
```

### Test the Server

You can test the MCP server by sending JSON-RPC messages:

```bash
# Initialize
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | ./mcp-server

# List tools
echo '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}' | ./mcp-server

# List API keys
echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"list_api_keys","arguments":{"category":"all"}}}' | ./mcp-server
```

## Security Best Practices

1. **Never commit `.env` files** - The `.gitignore` is configured to prevent this
2. **Use `.env.example`** - Document required variables without exposing secrets
3. **Rotate keys regularly** - Update your `.env` file periodically
4. **Use Docker secrets in production** - For Swarm/Kubernetes deployments
5. **Limit access** - Run the container as non-root user (already configured)

## Adding New API Keys

Edit `main.go` and add to the `apiKeyConfigs` map:

```go
"new_service": {
    EnvVar:      "NEW_SERVICE_API_KEY",
    Description: "New Service API key",
    Category:    "saas",  // or "llm", "canva", "internal"
},
```

Then update `.env.example`:

```bash
NEW_SERVICE_API_KEY=your-new-service-key-here
```

## Project Structure

```
mcp-api-keys-server/
├── main.go              # MCP server implementation
├── go.mod               # Go module definition
├── Dockerfile           # Multi-stage Docker build
├── docker-compose.yml   # Docker Compose configuration
├── .env.example         # Example environment variables
├── .gitignore           # Git ignore rules
└── README.md            # This file
```

## License

MIT License - feel free to use and modify as needed.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

Remember: Never include actual API keys in commits!
