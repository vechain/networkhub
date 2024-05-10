# networkHub

## Introduction
networkHub is a versatile framework designed to streamline the process of launching custom Vechain networks across various environments. It is tailored specifically for protocol and dapp development teams, providing a robust toolset to configure, start, stop, and manage blockchain networks with ease.

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

- **Launch Pre-configured Network (ThreeMasterNodeNetwork)**:
  ```bash
  # Request
  curl -X GET <api-url>/preset/threeMasterNodesNetwork
  
   # Response
  {"networkId": "localthreeMasterNodes"}
  
  # Request
  curl -X GET <api-url>/start/localthreeMasterNodes
  
  # Response
  Network Started
  ```

## Project Structure
- **Entrypoints**: Interface to interact with the framework. Currently implemented as an HTTP API server.
- **Actions**: Domain-specific language (DSL) that allows users to configure, start, stop, and request information on networks.
- **Environments**: Interface for running networks in different environments, with the Local environment currently implemented.

```
