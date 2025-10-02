# networkHub

## ALPHA Version Note
This repository is under alpha revision, beware when using it. Please wait for the first stable release for production use.

## Introduction
networkHub is a Go SDK framework designed to streamline the process of launching custom VeChain networks and connecting to public networks (mainnet/testnet). It provides a simple client-based API for protocol and dapp development teams to configure, start, stop, and manage blockchain networks programmatically.

## Quick Start

### **Launch Local Custom Network** (Simplest Way):
```go
package main

import (
    "log"
    "github.com/vechain/networkhub/client"
    "github.com/vechain/networkhub/preset"
    "github.com/vechain/networkhub/thorbuilder"
)

func main() {
    // Step 1: Use a preset network configuration (3 nodes local network)
    network := preset.LocalThreeNodesNetwork()
    
    // Step 2: Configure thor builder for automatic binary management
    cfg := thorbuilder.DefaultConfig()
    network.ThorBuilder = cfg
    
    // Step 3: Create client and start network
    client, err := client.New(network)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Stop()
    
    // Step 4: Start the network
    err = client.Start()
    if err != nil {
        log.Fatal(err)
    }
    
    log.Println("✅ 3-node VeChain network started successfully!")
    log.Printf("🌐 First node API: %s", network.Nodes[0].GetHTTPAddr())
    
    // Your network is ready for use!
    // The thor binary is automatically downloaded and built
    // All nodes are configured with genesis, keys, and networking
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
    testnet, err := preset.NewTestnetNetwork()
    if err != nil {
        log.Fatal(err)
    }
    
    testnetClient, err := client.New(testnet)
    if err != nil {
        log.Fatal(err)
    }
    defer testnetClient.Stop()
    
    // Connect to VeChain mainnet (auto-starts)  
    mainnet, err := preset.NewMainnetNetwork()
    if err != nil {
        log.Fatal(err)
    }
    
    mainnetClient, err := client.New(mainnet)
    if err != nil {
        log.Fatal(err)
    }
    defer mainnetClient.Stop()
    
    log.Println("✅ Connected to VeChain public networks!")
    // Networks auto-start when connecting to public networks
}
```

## Key Features

- **🚀 Simple API**: Get a VeChain network running in just 4 lines of code
- **🔧 Automatic Thor Management**: Thor binary is automatically downloaded, built, and configured
- **🌐 Multiple Environments**: Support for both Local and Docker environments  
- **📦 Built-in Presets**: Pre-configured networks for common use cases
- **🏗️ Custom Networks**: Full control over genesis, nodes, and network parameters
- **🔗 Public Network Support**: Easy connection to VeChain mainnet and testnet
- **⚙️ Node Management**: Dynamically add and remove nodes from running networks
- **🏥 Health Monitoring**: Built-in network health checks and validation
- **🐳 Docker Support**: Run networks in Docker containers with proper networking
- **🔑 Key Management**: Automatic private key and genesis configuration

## Purpose and Scope
networkHub enables teams to quickly deploy custom networks and connect to public VeChain networks, facilitating development and testing in both isolated and live environments. The SDK approach provides full programmatic control over network lifecycle management.

## Architecture

The framework is built around a **Launcher** architecture that orchestrates node management across different environments:

- **Client**: High-level API for network management
- **Launcher**: Central orchestrator for network operations (previously called "Overseer")
- **Environments**: Support for Local process execution and Docker containers
- **Presets**: Pre-configured network templates for common scenarios
- **ThorBuilder**: Automatic Thor binary management and building

## Available Presets

### Local Networks
- `preset.LocalThreeNodesNetwork()` - 3-node local network with authority nodes
- `preset.LocalSixNodesNetwork()` - 6-node local network for larger testing scenarios

### Public Networks  
- `preset.NewTestnetNetwork()` - Connect to VeChain testnet
- `preset.NewMainnetNetwork()` - Connect to VeChain mainnet

## Environments

### Local Environment
Runs Thor nodes as local processes on your machine:
```go
network.Environment = environments.Local
```

### Docker Environment  
Runs Thor nodes in Docker containers with proper networking:
```go
network.Environment = environments.Docker
```

## Technical Requirements
- **Git**: For cloning the repository
- **Golang**: Version 1.19 or higher
- **Docker**: Required for Docker environment (optional for Local environment)

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
