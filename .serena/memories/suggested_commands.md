# Suggested Commands for Migration-Go

## Building
```bash
# Build all packages
go build -v ./...

# Build just the main package
go build
```

## Testing
```bash
# Run all tests
go test -v

# Run specific test
go test -v -run Example_basic

# Run all example tests
go test -v -run Example
```

## Code Quality
```bash
# Format code
go fmt ./...

# Vet code
go vet ./...

# Tidy dependencies
go mod tidy
```

## Running Examples
```bash
# Run embedded example
cd examples/embedded && go run main.go

# Clean up example database
rm examples/embedded/myapp.db
```

## Dependencies
```bash
# Download dependencies
go mod download

# Update dependencies
go get -u ./...
go mod tidy
```

## Documentation
```bash
# View package documentation
go doc

# View specific function documentation
go doc Run
go doc Migrator
```
