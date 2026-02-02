# Contributing to DEFLOT

First off, thank you for considering contributing to DEFLOT! It's people like you that make DEFLOT such a great tool for the security community.

## Code of Conduct

This project and everyone participating in it is governed by basic principles of respect and professionalism. By participating, you are expected to uphold this standard.

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check existing issues to avoid duplicates. When you create a bug report, include as many details as possible:

- **Use a clear and descriptive title**
- **Describe the exact steps to reproduce the problem**
- **Provide specific examples** (command line, target domain, configuration)
- **Describe the behavior you observed** and explain what you expected
- **Include logs and screenshots** if applicable
- **Specify your environment** (OS, Go version, DEFLOT version)

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion:

- **Use a clear and descriptive title**
- **Provide a step-by-step description** of the suggested enhancement
- **Explain why this enhancement would be useful** to most DEFLOT users
- **List any alternative solutions** you've considered

### Pull Requests

- Fill in the required template
- Follow the Go coding style conventions
- Include thoughtful comments for complex logic
- Update documentation as needed
- Ensure all tests pass
- Create focused PRs (one feature/fix per PR)

## Development Setup

1. **Fork and clone the repository**:
   ```bash
   git clone https://github.com/bratyabasu07/deflot.git
   cd deflot
   ```

2. **Install dependencies**:
   ```bash
   go mod download
   ```

3. **Build the project**:
   ```bash
   go build -o deflot
   ```

4. **Run tests**:
   ```bash
   go test ./...
   ```

## Coding Guidelines

- **Code Style**: Follow standard Go conventions (`gofmt`, `golint`)
- **Documentation**: Add comments for all exported functions and types
- **Error Handling**: Always handle errors appropriately
- **Testing**: Add tests for new features
- **Commits**: Write clear, concise commit messages

## Project Structure

```
deflot/
├── cmd/              # CLI commands (root, config, server)
├── internal/
│   ├── config/       # Configuration management
│   ├── context/      # Application context
│   ├── dedup/        # Deduplication engine
│   ├── filters/      # URL classification filters
│   ├── integrations/ # External tool integrations
│   ├── output/       # Output writers
│   ├── pipeline/     # Core streaming pipeline
│   ├── server/       # Web interface server
│   ├── sources/      # Passive data sources
│   ├── status/       # HTTP status checker
│   ├── summary/      # Statistics and reporting
│   ├── targetlist/   # Batch target processing
│   └── ui/           # Terminal UI components
└── main.go
```

## Adding New Features

### Adding a New Data Source

1. Create a new file in `internal/sources/`
2. Implement the `Source` interface
3. Register it in `cmd/root.go`

### Adding a New Filter

1. Add pattern to `internal/filters/engine.go`
2. Update filter configuration in `internal/context/context.go`
3. Add CLI flag in `cmd/root.go`

## Questions?

Feel free to open an issue with the `question` label if you have any questions about contributing.

Thank you for contributing to DEFLOT! 🚀
