# go-kvs

Distributed key-value storage with leader-follower replication.

## Features

- **Write-Ahead Log (WAL)**: All commands are persisted to disk
- **In-memory Index**: Fast lookups with WAL offset values
- **Streaming Replication**: Followers connect to leader and receive real-time command stream
- **Auto-Reconnect**: Followers automatically retry connection if leader restarts
- **No Startup Order**: Start leader and followers in any order
- **gRPC Streaming**: Server-side streaming for low-latency replication

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

**Start in any order (followers will auto-connect to leader):**

**Terminal 1 - Leader:**
```bash
go run ./cmd/server/main.go --node-id=leader --leader --port=50051
```

**Terminal 2 - Follower 1:**
```bash
go run ./cmd/server/main.go --node-id=follower1 --port=50052 --leader-addr="localhost:50051"
```

**Terminal 3 - Follower 2:**
```bash
go run ./cmd/server/main.go --node-id=follower2 --port=50053 --leader-addr="localhost:50051"
```

**Terminal 4 - Client (connects to leader):**
```bash
go run ./cmd/client/main.go
```

### How Streaming Replication Works

1. **Followers connect to leader** on startup (not vice versa)
2. **Leader maintains streaming connections** to all followers
3. **On write**: Leader immediately broadcasts command to all streams
4. **No startup order dependency**: Start nodes in any order
5. **Auto-reconnect**: Followers automatically retry if connection drops

## Commands

- `get {key}` - Retrieve value for key
- `set {key} {val}` - Store key-value pair
- `del {key}` - Delete key
- `keys` - List all keys
- `exit` - Close client