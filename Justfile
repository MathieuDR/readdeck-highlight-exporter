# Justfile for readdeck-highlight-exporter
# This file contains common development tasks

# Display all available commands by default
default:
    @just --list

    
# Run tests (all or specific package)
test *ARGS:
    #!/usr/bin/env bash
    # Parse args to separate package name from flags
    package=""
    flags=""
    
    for arg in {{ARGS}}; do
        if [[ "$arg" == -* ]]; then
            # This is a flag
            flags="$flags $arg"
        elif [ -z "$package" ]; then
            # This is the package name (first non-flag argument)
            package="$arg"
        fi
    done
    
    if [ -z "$package" ]; then
        # Run all tests if no package specified
        echo "Running all tests with flags: $flags"
        go test $flags ./...
    else
        # Convert package name to path
        package_path="./internal/$package"
        if [ -d "$package_path" ]; then
            echo "Running tests for package '$package' with flags: $flags"
            go test $flags "$package_path"
        else
            echo "Error: Package '$package' not found at path $package_path"
            exit 1
        fi
    fi

# Run integration tests with the appropriate environment flag
test-integration *FLAGS:
    #!/usr/bin/env bash
    echo "Running integration tests..."
    export RUN_INTEGRATION_TEST=true
    go test {{FLAGS}} ./internal/test
    result=$?
    # Reset the flag
    unset RUN_INTEGRATION_TEST
    # Return the original exit code
    exit $result


# Clean test artifacts directory
clean:
    #!/usr/bin/env bash
    echo "Cleaning test artifacts..."
    rm -rf ./test_artifacts
    echo "Done!"

# Build using Nix
build:
    nix build

# Build using Go directly
build-go:
    go build -o bin/highlight-exporter

# Set up a development environment
setup:
    go mod vendor
    go mod tidy
    go mod verify

# Format Go code
fmt:
    go fmt ./...

# Generate new command template
new-command NAME:
    #!/usr/bin/env bash
    echo "Creating new cobra command: {{NAME}}"
    cd cmd && go run github.com/spf13/cobra-cli@latest add {{NAME}}
