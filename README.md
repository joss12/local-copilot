# Local Copilot

An offline code completion engine that learns from YOUR codebase. Built with Go and TypeScript.

![Demo](https://img.shields.io/badge/status-working-brightgreen)
![Go](https://img.shields.io/badge/go-1.21+-blue)
![VSCode](https://img.shields.io/badge/vscode-1.85+-blue)

## 🎯 Features

- ⚡ **Lightning Fast** - Sub-3ms response times
- 🔒 **Completely Offline** - No internet required, no data leaves your machine
- 🎯 **Context-Aware** - Learns from your actual codebase
- 🔧 **Smart Suggestions** - Function signatures, variables, types
- 🚀 **Production Ready** - Clean architecture, tested and working



## 🚀 Quick Start

### Prerequisites

- Go 1.21+
- Node.js 18+
- pnpm
- VSCode 1.85+

### Backend Setup
```bash
# Navigate to backend directory
cd backend

# Build the binary
go build -o copilot ./cmd/main.go

# Run the backend
./copilot --port 8089
```

### Extension Setup
```bash
# Navigate to extension directory
cd extension

# Install dependencies
pnpm install

# Compile TypeScript
pnpm run compile

# Open in VSCode and press F5 to launch Extension Development Host
code .
```

## 📖 Usage

1. **Start the backend:**
```bash
   cd backend
   ./copilot --port 8089
```

2. **Open VSCode** and press `F5` to launch the Extension Development Host

3. **Open a workspace** with Go/JS/TS code

4. **Index your workspace:**
   - Press `Ctrl+Shift+P`
   - Run: `Local Copilot: Index Workspace`

5. **Start coding!** - Suggestions appear automatically as you type

## 🎮 Commands

Access via Command Palette (`Ctrl+Shift+P`):

- `Local Copilot: Start Backend` - Start the Go backend server
- `Local Copilot: Stop Backend` - Stop the backend server
- `Local Copilot: Index Workspace` - Index current workspace

## ⚙️ Configuration

Available settings (File → Preferences → Settings → Local Copilot):

- `localCopilot.backendPort` - Backend server port (default: 8080)
- `localCopilot.backendPath` - Path to copilot binary (optional)

## 🧪 Supported Languages

- ✅ Go (full support)
- 🚧 JavaScript (architecture ready)
- 🚧 TypeScript (architecture ready)

## 🔬 Technical Details

### Backend Components

- **File Scanner** - Finds source files, skips node_modules, .git, etc.
- **AST Parser** - Extracts functions, variables, types using `go/parser`
- **SQLite Database** - Stores indexed symbols with metadata
- **Pattern Matcher** - Searches and ranks suggestions by confidence
- **REST API** - 3 endpoints: /health, /index, /suggest

### Performance

- **Indexing:** ~300ms for 7 files, 60 symbols
- **Suggestions:** 2-5ms response time
- **Database:** Lightweight SQLite, ~1MB for typical projects

### Algorithm

1. User types partial symbol (e.g., "NewData")
2. Extension sends context to backend
3. Backend searches database for matches
4. Results ranked by:
   - Prefix match (+0.3 confidence)
   - Same file (+0.2 confidence)
   - Base confidence (0.5)
5. Suggestions returned sorted by confidence

### LLM Integration (Experimental)

Optional integration with local LLMs via Ollama:
```bash
# Install Ollama
curl -fsSL https://ollama.com/install.sh | sh

# Pull a model
ollama pull deepseek-coder:1.3b-base

# Run backend with LLM support
./copilot --port 8089 --llm
```

**Note:** LLM mode is slower (~10s vs 3ms) but provides AI-powered completions.

## 📊 Benchmarks

| Operation | Time | Details |
|-----------|------|---------|
| Indexing | 250ms | 7 files, 60 symbols |
| Suggestion | 2.5ms | Pattern matching |
| LLM Suggestion | 7-10s | DeepSeek Coder 1.3B |

## 🛠️ Development

### Running Tests
```bash
# Backend tests
cd backend
go test ./...

# Extension tests
cd extension
pnpm test
```

### Building for Production
```bash
# Backend
cd backend
go build -o copilot ./cmd/main.go

# Extension
cd extension
pnpm run vscode:prepublish
pnpm exec vsce package
```

## 🚧 Roadmap

- [x] Go AST parser
- [x] Pattern matching engine
- [x] VSCode extension
- [x] Clean signature formatting
- [x] LLM integration (experimental)
- [ ] JavaScript/TypeScript parser
- [ ] Streaming suggestions (SSE)
- [ ] Multi-line completions
- [ ] Context-aware ranking improvements

## 🤝 Contributing

Contributions welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## 📝 License

MIT License - feel free to use this project for learning and portfolio purposes.

## 🙏 Acknowledgments

- Built with Go's excellent `go/parser` and `go/ast` packages
- VSCode extension API
- Ollama for local LLM support

## 📧 Contact

Built by BornToShine - [GitHub Profile](https://github.com/joss12)

---

**⭐ Star this repo if you found it useful!**
