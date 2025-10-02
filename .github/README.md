# networkHub

## ALPHA Version Note
This repository is under alpha revision, beware when using it. Please wait for the first stable release for production use.

## Introduction
networkHub is a Go SDK framework designed to streamline the process of launching custom VeChain networks and connecting to public networks (mainnet/testnet). It provides a simple client-based API for protocol and dapp development teams to configure, start, stop, and manage blockchain networks programmatically.

## Quick Start

### **Launch Local Custom Network**:
```go
package main

import (
    "log"
    "github.com/vechain/networkhub/client"
    "github.com/vechain/networkhub/preset"
)

func main() {
    // Create a local three-node network
    network := preset.LocalThreeMasterNodesNetwork()
    
    // Set the thor binary path for all nodes
    for _, node := range network.Nodes {
        node.SetExecArtifact("/path/to/thor/binary")
    }
    
    // Create client and start network
    c, err := client.NewWithNetwork(network)
    if err != nil {
        log.Fatal(err)
    }
    defer c.Stop()
    
    log.Println("Network started successfully!")
    // Your network is ready for use
}
```

### **Connect to VeChain Public Networks**:
```go
package main

import (
    "log"
    "github.com/vechain/networkhub/client" 
    "github.com/vechain/networkhub/preset"
)

func main() {
    // Connect to VeChain testnet (auto-starts)
    testnet, _ := preset.NewTestnetNetwork("dev")
    testnetClient, err := client.NewWithNetwork(testnet)
    if err != nil {
        log.Fatal(err)
    }
    defer testnetClient.Stop()
    
    // Connect to VeChain mainnet (auto-starts)
    mainnet, _ := preset.NewMainnetNetwork("main")
    mainnetClient, err := client.NewWithNetwork(mainnet)
    if err != nil {
        log.Fatal(err)
    }
    defer mainnetClient.Stop()
    
    log.Println("Connected to public networks!")
}
```

## Purpose and Scope
networkHub enables teams to quickly deploy custom networks and connect to public VeChain networks, facilitating development and testing in both isolated and live environments. The SDK approach provides full programmatic control over network lifecycle management.

## Technical Requirements
- **Git**: For cloning the repository.
- **Golang**: Version 1.19 or higher.

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
