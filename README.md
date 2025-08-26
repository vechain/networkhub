## NetworkHub: Create and Run Custom Thor Networks

This guide shows multiple ways to spin up Thor networks with NetworkHub:
- Using built-in presets
- Manually with the Local environment (spawn native thor processes)
- Manually with the Docker environment (spawn thor in containers)
- Advanced: custom genesis, attaching/removing nodes at runtime, additional args

### Prerequisites
- Go 1.21+
- macOS/Linux
- Docker (only for the Docker environment)

### Conventions used below
- The imports assume you are inside a Go module and can import this repo packages.
- Many examples use `thorbuilder.DefaultConfig()` which honors environment variables:
  - `THOR_WORKING_DIR`: if set, builder uses an existing working directory and performs a local build
  - `THOR_BRANCH`: if set, builder downloads the Thor repo and checks out this branch; defaults to `master`

## 1) Using the presets

Presets give you ready-to-run topologies plus a sensible genesis. You can run them either in the Local or Docker environments.

### 1.1 Local environment with a preset
```go
package main

import (
    "log"
    "time"

    envlocal "github.com/vechain/networkhub/environments/local"
    "github.com/vechain/networkhub/preset"
    "github.com/vechain/networkhub/thorbuilder"
    "github.com/vechain/thor/v2/thorclient"
)

func main() {
    // Build or fetch thor based on env vars (THOR_WORKING_DIR / THOR_BRANCH)
    cfg := thorbuilder.DefaultConfig()

    // Choose a preset
    net := preset.LocalThreeMasterNodesNetwork()
    net.ThorBuilder = cfg // Local env will build thor and set the exec path

    // Create local environment, load, start
    env := envlocal.NewEnv()
    if _, err := env.LoadConfig(net); err != nil {
        log.Fatal(err)
    }
    if err := env.StartNetwork(); err != nil {
        log.Fatal(err)
    }
    defer env.StopNetwork()
}
```

Notes:
- If a node’s `APIAddr` is empty, Local automatically allocates an available port and binds on `0.0.0.0`.
- If a node’s P2P port is zero/missing, Local automatically allocates one.

### 1.2 Docker environment with a preset
```go
package main

import (
    "log"

    envdocker "github.com/vechain/networkhub/environments/docker"
    "github.com/vechain/networkhub/preset"
    "github.com/vechain/networkhub/thorbuilder"
)

func main() {
    // Build a Docker image from Thor based on env vars
    cfg := thorbuilder.DefaultConfig()

    net := preset.LocalSixNodesNetwork()
    net.ThorBuilder = cfg // Docker env will build a docker image and assign it to nodes

    env := envdocker.NewEnv()
    if _, err := env.LoadConfig(net); err != nil {
        log.Fatal(err)
    }
    if err := env.StartNetwork(); err != nil {
        log.Fatal(err)
    }
    defer env.StopNetwork()
}
```

Notes:
- Alternatively, you can set each node’s `ExecArtifact` to a prebuilt image tag (e.g. `vechain/thor:latest`).
- Docker env exposes each node’s API port on localhost; host ports are derived from node configuration during `LoadConfig`.

## 2) Advanced usage

### 2.1 Attaching a node at runtime (Local environment)
```go
// Assume env := envlocal.NewEnv() and env.StartNetwork() already called
// Reuse the same genesis that the network uses
g := env.Config().Nodes[0].GetGenesis()

newNode := &node.BaseNode{
    ID:        "nodeX",
    Type:      node.RegularNode,
    APICORS:   "*",
    Verbosity: 3,
    Key:       "<hex private key>",
    Genesis:   g,
}

// Optional: additional args passed to thor (e.g., enable specific tracers)
args := map[string]string{"api-allowed-tracers": "call"}

// If you didn’t provide net.ThorBuilder during LoadConfig, set ExecArtifact here
// newNode.SetExecArtifact("/path/to/thor")

if err := env.AttachNode(newNode, nil, args); err != nil {
    panic(err)
}
```

### 2.2 Removing a node at runtime (Local environment)
```go
if err := env.RemoveNode("nodeX"); err != nil {
    panic(err)
}
```

### 2.3 Accessing node HTTP endpoints
```go
for id, n := range env.Nodes() {
    // n implements node.Lifecycle, original configs are in env.Config().Nodes
    _ = id
}

// From the original configs:
for _, nc := range env.Config().Nodes {
    // GetHTTPAddr converts 0.0.0.0 bindings to 127.0.0.1 for convenience
    addr := nc.GetHTTPAddr()
    _ = addr
}
```

### 2.4 Port management notes (Local environment)
- API port and P2P port allocation use an in-process port manager keyed by network ID (`Environment + BaseID`).
- Ports are reserved on start and released on `StopNetwork()`.
- You can provide explicit `APIAddr` or `P2PListenPort` if you need fixed values; otherwise, leave empty/zero and let the manager allocate.

## Flow diagrams

### Scenario 1: Network not started → Attach node → Start network
```mermaid
sequenceDiagram
  autonumber
  participant Caller
  participant Local as Local env
  participant Network as Network cfg
  participant Builder as ThorBuilder
  participant Ports as Ports Manager
  participant NodeCfg as Node Config
  participant LocalNode as Local Node (process)

  Caller->>Local: AttachNode(n, buildCfg, additionalArgs)
  Local->>Local: validateAttachable()
  alt invalid
    Local-->>Caller: error
  else valid
    opt buildCfg provided
      Local->>Builder: Download()+Build()
      Builder-->>Local: exec artifact path
      Local->>NodeCfg: SetExecArtifact(path)
    end
    Local->>NodeCfg: AddAdditionalArg(k,v) (for each)
    Local->>Ports: Allocate API port (if APIAddr empty)
    Ports-->>Local: port
    Local->>NodeCfg: SetAPIAddr("0.0.0.0:port")
    Local->>Ports: Allocate P2P port (if 0)
    Ports-->>Local: p2p port
    Local->>NodeCfg: SetP2PListenPort(p2p)
    Local->>Local: checkNode() (exec/data/config dirs)
    Local->>Network: Append n to Nodes
    Local->>Local: Register n in l.localNodes
    Local-->>Caller: attached (no process started)
  end

  Caller->>Local: StartNetwork()
  Local->>Local: enodes() for all nodes
  loop each node
    Local->>LocalNode: Start()
  end
  Local->>Network: HealthCheck all
  Local-->>Caller: ready
```

### Scenario 2: Network started → Attach node
```mermaid
sequenceDiagram
  autonumber
  participant Caller
  participant Local as Local env
  participant Builder as ThorBuilder
  participant Ports as Ports Manager
  participant NodeCfg as Node Config
  participant LocalNode as Local Node (process)

  Caller->>Local: AttachNode(n, buildCfg, additionalArgs)
  Local->>Local: validateAttachable()
  alt invalid
    Local-->>Caller: error
  else valid
    opt buildCfg provided
      Local->>Builder: Download()+Build()
      Builder-->>Local: exec artifact path
      Local->>NodeCfg: SetExecArtifact(path)
    end
    Local->>NodeCfg: AddAdditionalArg(k,v) (for each)
    Local->>Ports: Allocate API port (if APIAddr empty)
    Ports-->>Local: port
    Local->>NodeCfg: SetAPIAddr("0.0.0.0:port")
    Local->>Ports: Allocate P2P port (if 0)
    Ports-->>Local: p2p port
    Local->>NodeCfg: SetP2PListenPort(p2p)
    Local->>Local: checkNode()
    Local->>LocalNode: NewLocalNode(n, enodes)
    Local->>LocalNode: Start()
    Local->>NodeCfg: HealthCheck(0, 30s)
    Local-->>Caller: node attached and healthy
  end
```

### Corner cases and error flows
```mermaid
sequenceDiagram
  autonumber
  participant Caller
  participant Local as Local env
  participant Builder as ThorBuilder
  participant Ports as Ports Manager
  participant NodeCfg as Node Config

  Caller->>Local: AttachNode(n, buildCfg, args)
  alt network not loaded
    Local-->>Caller: error("network configuration is not loaded")
  else duplicate node ID
    Local-->>Caller: error("node already exists")
  else builder error
    Local->>Builder: Download()+Build()
    Builder-->>Local: error
    Local-->>Caller: error(from builder)
  else empty network ID on port alloc
    Local->>Ports: Allocate(l.id="")
    Ports-->>Local: error("networkID must not be empty")
    Local-->>Caller: error
  else exec artifact missing
    Local->>Local: checkNode()
    Local-->>Caller: error("exec artifact path ... does not exist")
  else process start failure
    Local->>LocalNode: Start()
    LocalNode-->>Local: error
    Local-->>Caller: error("unable to start node ...")
  else health check timeout
    Local->>NodeCfg: HealthCheck(0, 30s)
    NodeCfg-->>Local: error(timeout)
    Local-->>Caller: error("timeout waiting for node ...")
  end

  Note over Local,Ports: StopNetwork()
  Local->>Local: Stop each node
  alt stop error
    Local-->>Caller: error("unable to stop node ...")
  else ok
    Local->>Ports: ReleaseAll(l.id)
    Ports-->>Local: ok
    Local-->>Caller: stopped and ports released
  end
```

### Starting a network (Local environment)
```mermaid
sequenceDiagram
  autonumber
  participant Caller
  participant Local as Local env
  participant Network as Network cfg
  participant LocalNode as Local Node (process)

  Caller->>Local: StartNetwork()
  Local->>Local: enodes() for all nodes
  loop each node in Network.Nodes
    Local->>LocalNode: NewLocalNode(nodeCfg, enodes)
    Local->>LocalNode: Start()
    Note right of LocalNode: cleanup previous process, ensure dirs,\nwrite keys & genesis, exec thor with args
  end
  Local->>Network: HealthCheck all (0, 30s)
  Local-->>Caller: network ready
```

### Starting a network (Docker environment)
```mermaid
sequenceDiagram
  autonumber
  participant Caller
  participant Docker as Docker env
  participant IpMgr as IP Manager
  participant DockerCLI as Docker client/daemon

  Caller->>Docker: StartNetwork()
  Docker->>Docker: checkOrCreateNetwork(networkID, subnet)
  Docker->>Docker: build enodes list (using IpMgr.GetNodeIP for prior nodes)
  loop each node in Network.Nodes
    Docker->>IpMgr: NextIP(nodeID)
    Docker->>DockerCLI: Ensure image (inspect or pull)
    Docker->>DockerCLI: ContainerCreate(config, hostConfig, networkConfig)
    Docker->>DockerCLI: ContainerStart(id)
  end
  Docker-->>Caller: network started
```
