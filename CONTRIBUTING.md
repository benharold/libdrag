# Contributing to libdrag

Thank you for your interest in contributing to libdrag! This document provides guidelines for contributing to the project.

## Code of Conduct

Please be respectful and professional in all interactions. We welcome contributions from developers of all skill levels.

## Getting Started

1. Fork the repository
2. Clone your fork locally
3. Create a feature branch: `git checkout -b feature/your-feature-name`
4. Make your changes
5. Run tests: `go test ./...`
6. Commit your changes with clear commit messages
7. Push to your fork and submit a pull request

## Development Setup

```bash
# Clone the repository
git clone https://github.com/benharold/libdrag.git
cd libdrag

# Run the demo to verify everything works
go run cmd/libdrag/main.go

# Run tests
go test ./...
```

## Pull Request Process

1. Ensure your code follows Go conventions and best practices
2. Add tests for new functionality
3. Update documentation as needed
4. Ensure all tests pass
5. Update CHANGELOG.md with your changes
6. Submit a pull request with a clear description

## Areas for Contribution

- Additional racing formats (IHRA variations, bracket racing, etc.)
- Performance optimizations
- Cross-platform testing
- Documentation improvements
- Example applications
- Bug fixes and stability improvements
- Mobile platform integration helpers

## Coding Standards

- Follow standard Go formatting (`go fmt`)
- Write clear, documented code
- Include unit tests for new functionality
- Keep public APIs simple and well-documented
- Use meaningful variable and function names

## Reporting Issues

When reporting issues, please include:
- Go version
- Operating system
- Steps to reproduce
- Expected vs actual behavior
- Relevant code snippets or logs

## Questions?

Feel free to open an issue for questions or join discussions about potential features.
