# networkHub

## ALPHA Version Note
This repository is under alpha revision, beware when using it. Please wait for the first stable release for production use.

## Introduction
networkHub is a versatile framework designed to streamline the process of launching custom Vechain networks across various environments. It is tailored specifically for protocol and dapp development teams, providing a robust toolset to configure, start, stop, and manage blockchain networks with ease.

# Quick start
### **Launch Pre-configured Network via command line**:
  ```bash
  # Setup the preset network
  > go run cmd/main.go cmd preset local threeMasterNodesNetwork /Users/pedro/go/src/github.com/vechain/thor/bin/thor
  ...
  2024/06/07 17:31:43 INFO preset network config was successful... networkId=localthreeMaster
  
  # Start preset network
  > go run cmd/main.go cmd start localthreeMaster
  2024/06/07 17:31:57 INFO Registered preset network networkId=threeMasterNodesNetwork
  2024/06/07 17:31:57 INFO Registered preset network networkId=sixNodesNetwork
  2024/06/07 17:31:57 INFO Starting network... ID=localthreeMaster
  ...
  ```
### **Launch Pre-configured Network via command line using Docker**:
  ```bash
  # Setup the preset network
  > go run cmd/main.go cmd preset docker threeMasterNodesNetwork vechain/thor:latest
  ...
  2024/06/13 17:15:18 INFO Configuring network...
  2024/06/13 17:15:18 INFO Network saved to file filepath=/Users/pedro/go/src/github.com/vechain/networkhub/networks_db.json
  2024/06/13 17:15:18 INFO preset network config was successful... networkId=dockerthreeMaster
  
  # Start preset network
  > go run cmd/main.go cmd start dockerthreeMaster
  2024/06/13 17:16:05 INFO Network created networkName=dockerthreeMaster-network
  2024/06/13 17:16:06 INFO network started successfully...
  ...
  ```


### **Launch Pre-configured Network via API server**:
  ```bash
  # Start the networkhub api
  go run ./cmd/main.go api
  
  # Response
  api called
  Server started on :8080
  
  # Start a local pre-setup configured network 
  # Must specify the path to thor locally
  curl -X POST http://localhost:8080/preset/threeMasterNodesNetwork \
     -d '{"artifactPath":"/Users/pedro/go/src/github.com/vechain/thor/bin/thor", "environment":"local"}'
  
   # Response
  {"networkId": "localthreeMasterNodes"}
  
  # Start the network
  curl -X GET http://localhost:8080/start/localthreeMasterNodes
  
  # Response
  Network Started
  ```

## Purpose and Scope
networkHub enables teams to quickly deploy custom networks, facilitating development and testing in a controlled environment. This framework is especially beneficial for protocol and dapp teams looking to experiment with network configurations and behaviors without the overhead of setting up infrastructure from scratch.


## Technical Requirements
- **Git**: For cloning the repository.
- **Golang**: Version 1.19 or higher.

## Setup
To setup networkHub, follow these steps:

1. Clone the repository:
   ```bash
   git clone https://github.com/vechain/networkhub
   ```
2. Build the application:
   ```bash
   go build -o networkHub ./cmd/main.go
   ```
3. To run networkHub, execute:
   ```bash
   ./networkHub
   ```
   or
   ```bash
   go run ./cmd/main.go
   ```
   This will display the available command-line options.

### Running the API
To launch the framework in API mode, use the following command:
```bash
./networkHub api
```
or
```bash
go run ./cmd/main.go api
```

## Usage Examples
Below are some example `curl` requests to interact with the networkHub via its HTTP API. 
(Note: Replace `<api-url>` with the actual URL where networkHub API is hosted.)

- **Configure Network**:
  ```bash
  # Request
  curl -X POST <api-url>/config -d '{...}' # {...} is a network config json
  
  # Response
  {"networkId": "8b38927893d1fc841b7bedcbaebb03821000908cfa1ee07a09002bc0e0ed3caf"}
  ```
  

- **Start Network**:
  ```bash
  # Request
  curl -X POST <api-url>/start/8b38927893d1fc841b7bedcbaebb03821000908cfa1ee07a09002bc0e0ed3caf
  
  # Response
  Network Started
  ```

- **Stop Network**:
  ```bash
  # Request
  curl -X POST <api-url>/stop/8b38927893d1fc841b7bedcbaebb03821000908cfa1ee07a09002bc0e0ed3caf
  
   # Response
  Network Stopped
  ```

## Thorbuilder Package
The `thorbuilder` package is a key component of the networkHub framework that provides flexible configuration options for building Thor binaries from source. It supports both local builds and Docker image creation, with options for reusable builds and debug configurations.

### Features:
- **Configurable Build Process**: Customize download and build parameters through a structured Config system
- **Reusable Builds**: Option to reuse existing cloned repositories for faster subsequent builds
- **Debug Build Support**: Build with debug flags for development and testing
- **Docker Image Building**: Create Docker images from Thor source
- **Custom Genesis Support**: Fetch custom genesis files from URLs
- **Flexible Repository Sources**: Support for different branches and repository URLs

### Configuration Options:
```go
type Config struct {
    DownloadConfig *DownloadConfig
    BuildConfig    *BuildConfig
}

type DownloadConfig struct {
    RepoUrl    string  // Repository URL (default: https://github.com/vechain/thor)
    Branch     string  // Branch to clone (default: master)
    IsReusable bool    // Whether to reuse existing clone
}

type BuildConfig struct {
    ExistingPath string  // Path to existing Thor binary
    DebugBuild   bool    // Whether to build with debug flags
}
```

### Example Usage:

#### Basic Usage with Default Configuration:
```go
package main

import (
    "log"
    "github.com/vechain/networkhub/thorbuilder"
    "log/slog"
)

func main() {
    // Use default configuration
    cfg := thorbuilder.DefaultConfig()
    builder := thorbuilder.New(cfg)

    if err := builder.Download(); err != nil {
        log.Fatalf("Failed to download source: %v", err)
    }

    thorBinaryPath, err := builder.Build()
    if err != nil {
        log.Fatalf("Failed to build binary: %v", err)
    }

    slog.Info("Thor binary built successfully", "path", thorBinaryPath)
}
```

#### Custom Configuration Example:
```go
package main

import (
    "log"
    "github.com/vechain/networkhub/thorbuilder"
    "log/slog"
)

func main() {
    cfg := &thorbuilder.Config{
        DownloadConfig: &thorbuilder.DownloadConfig{
            RepoUrl:    "https://github.com/your-fork/thor",
            Branch:     "custom-feature",
            IsReusable: true,
        },
        BuildConfig: &thorbuilder.BuildConfig{
            DebugBuild: true,
        },
    }
    
    builder := thorbuilder.New(cfg)

    if err := builder.Download(); err != nil {
        log.Fatalf("Failed to download source: %v", err)
    }

    thorBinaryPath, err := builder.Build()
    if err != nil {
        log.Fatalf("Failed to build binary: %v", err)
    }

    slog.Info("Debug Thor binary built successfully", "path", thorBinaryPath)
}
```

#### Building a Docker Image:
```go
package main

import (
    "log"
    "github.com/vechain/networkhub/thorbuilder"
    "log/slog"
)

func main() {
    cfg := thorbuilder.DefaultConfig()
    builder := thorbuilder.New(cfg)

    imageTag, err := builder.BuildDockerImage()
    if err != nil {
        log.Fatalf("Failed to build Docker image: %v", err)
    }

    slog.Info("Docker image built successfully", "tag", imageTag)
}
```
