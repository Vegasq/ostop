# ostop - OpenSearch Terminal UI

A terminal-based monitoring dashboard for OpenSearch clusters, focused on cluster operations and capacity planning.

![ostop demo](demo.gif)

**[📚 Documentation Website](https://ostop.mkla.dev)**

## Features

- 🏥 **Cluster Health** - Real-time cluster status monitoring
- 📊 **Node Statistics** - Detailed per-node metrics with JVM heap, CPU, RAM, and disk usage
- 📑 **Index Overview** - Monitor indices with health status, documents, and storage
- 🔍 **Index Schema Viewer** - Drill down into individual indices to explore field mappings, types, and analyzers
- 🔀 **Shard Distribution** - Visualize shard allocation across nodes with balance analysis
- 🎯 **Resource Dashboard** - Cluster-wide aggregate metrics and capacity planning insights
- 📈 **Live Metrics** - Real-time graphs showing indexing and search rates with auto-refresh
- 📊 **Visual Metrics** - Color-coded bar charts, health indicators, and Braille-rendered graphs
- 🎨 **Split-Panel UI** - Navigate between cluster overview, nodes, indices, shards, and resources
- 🔐 **AWS Support** - Native AWS OpenSearch support with SigV4 signing
- ⌨️  **Keyboard Driven** - Efficient terminal-based workflow with Vim-like navigation

## Installation

```bash
# Build from source
go build -o ostop main.go

# Or run directly
go run main.go --endpoint <your-endpoint>
```

## Usage

### Local OpenSearch

```bash
# Start OpenSearch in Docker
docker run -d \
  --name opensearch-dev \
  -p 9200:9200 \
  -p 9600:9600 \
  -e "discovery.type=single-node" \
  -e "DISABLE_SECURITY_PLUGIN=true" \
  opensearchproject/opensearch:latest

# Connect with ostop
./ostop --endpoint http://localhost:9200
```

### AWS OpenSearch

```bash
# Using default AWS credentials
./ostop \
  --endpoint https://search-mydomain-xxx.us-east-1.es.amazonaws.com \
  --region us-east-1

# Using specific AWS profile
./ostop \
  --endpoint https://search-mydomain-xxx.us-east-1.es.amazonaws.com \
  --region us-east-1 \
  --profile my-profile
```

### Command Line Options

```
--endpoint <url>      OpenSearch endpoint URL (required)
--region <region>     AWS region (required for AWS OpenSearch)
--profile <name>      AWS profile name (optional)
--insecure            Skip TLS verification (development only)
--version             Show version information
```

## Keyboard Shortcuts

### Navigation
- `↑/k` - Move up (menu navigation, index selection, or scroll up in right panel)
- `↓/j` - Move down (menu navigation, index selection, or scroll down in right panel)
- `Tab` - Switch between left and right panels
- `Enter` - Select view (when in left panel) or drill into index schema (in indices view)
- `Esc/Backspace` - Return from index schema view to indices list

### Scrolling (Right Panel)
- `PgUp/b` - Scroll up one page
- `PgDn/f/Space` - Scroll down one page
- `Home/g` - Jump to top
- `End/G` - Jump to bottom

### Actions
- `r` - Refresh data from cluster (manual refresh for all views except Live Metrics)
- `q` - Quit application
- `Ctrl+C` - Force quit

## Requirements

- Go 1.21 or later
- Access to OpenSearch cluster (local or AWS)
- For AWS: Valid AWS credentials configured

## AWS Authentication

ostop uses the standard AWS credential chain:
1. Environment variables (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`)
2. Shared credentials file (`~/.aws/credentials`)
3. IAM roles (EC2 instance profiles, ECS task roles)

This matches the behavior of `awscurl` and other AWS CLI tools.

## Development

```bash
# Install dependencies
go mod download

# Run tests
go test ./...

# Build
go build -o ostop main.go

# Build with optimizations
go build -o ostop -ldflags="-s -w" main.go
```

## License

MIT
