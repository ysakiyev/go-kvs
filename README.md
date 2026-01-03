# go-kvs

Distributed key-value storage with leader-follower replication.

## Features

- **Write-Ahead Log (WAL)**: All commands are persisted to disk
- **In-memory Index**: Fast lookups with WAL offset values
- **Leader-Follower Replication**: Data automatically replicated across multiple nodes
- **gRPC Communication**: Client-server and server-server communication

## How It Works

Server has Write-Ahead Log (WAL) which stores all commands.
On startup, server reads WAL line-by-line and restores state to in-memory index.

**Set operation**: Appends command to WAL, updates in-memory index, replicates to followers.

**Get operation**: Reads key from in-memory index, fetches value from WAL at offset.

## Running Single Node

**Server:**
```bash
go run ./cmd/server/main.go --leader --port=50051
```

**Client:**
```bash
go run ./cmd/client/main.go
```

## Running with Replication (3 nodes)

**Leader:**
```bash
go run ./cmd/server/main.go --node-id=leader --leader --port=50051 --followers="localhost:50052,localhost:50053"
```

**Follower 1:**
```bash
go run ./cmd/server/main.go --node-id=follower1 --port=50052 --leader-addr="localhost:50051"
```

**Follower 2:**
```bash
go run ./cmd/server/main.go --node-id=follower2 --port=50053 --leader-addr="localhost:50051"
```

**Client (connects to leader):**
```bash
go run ./cmd/client/main.go
```

## Commands

- `get {key}` - Retrieve value for key
- `set {key} {val}` - Store key-value pair
- `del {key}` - Delete key
- `exit` - Close client