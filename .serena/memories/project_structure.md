# Migration-Go Project Structure

## Root Directory (Core Library)
```
migration-go/
├── migration.go         # Core migration library implementation
├── migration_test.go    # Unit tests for core functionality
├── version.go          # Helper function to get current version
├── testdata/           # Test migration files (for unit tests only)
│   ├── 000_test_baseline.sql
│   └── 001_add_user_profile.sql
├── go.mod              # Go module definition
├── go.sum              # Dependency checksums
└── README.md           # Main documentation
```

**Purpose**: Contains only the core library code and its unit tests. No example code in root.

## Examples Directory
```
examples/
├── basic/              # Comprehensive usage examples
│   ├── main.go        # Runnable program showing all features
│   ├── README.md      # Detailed explanations
│   └── migrations/    # Sample migration files
│       ├── 000_test_baseline.sql
│       └── 001_add_user_profile.sql
│
└── embedded/          # Production-ready application example
    ├── main.go        # Complete startup flow example
    ├── README.md      # Step-by-step guide
    └── migrations/    # Application migration files
        ├── 000_initial_schema.sql
        └── 001_add_comments.sql
```

**Purpose**: Complete, runnable examples for users to learn from and copy.

## Design Principles

1. **Clean Separation**: Root contains only library code, examples are isolated
2. **Self-Contained Examples**: Each example is a complete, runnable project
3. **Minimal Root**: Keep root directory focused on the core library
4. **Testable**: Core library has comprehensive unit tests
5. **Documented**: Each directory has its own README

## File Organization Rules

### Root Directory Should Contain:
✅ Core library implementation (`*.go`)
✅ Unit tests (`*_test.go`)
✅ Test fixtures (`testdata/`)
✅ Documentation (`README.md`)
✅ Module files (`go.mod`, `go.sum`)

### Root Directory Should NOT Contain:
❌ Example applications
❌ Tutorial code
❌ Demo programs
❌ Multiple documentation files

### Examples Directory Should Contain:
✅ Complete, runnable programs
✅ Each example in its own subdirectory
✅ Each example with its own README
✅ Sample migration files specific to each example

## Running Tests

```bash
# Test core library only
go test -v

# Test all packages including examples
go test -v ./...

# Test specific package
go test -v ./examples/basic
```

## Running Examples

```bash
# Basic usage examples
cd examples/basic && go run main.go

# Embedded migrations example
cd examples/embedded && go run main.go
```
