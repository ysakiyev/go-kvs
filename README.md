# go-kvs

Distributed key-value storage with leader-follower streaming replication.

## Features

- **Write-Ahead Log (WAL)**: All commands persisted to disk for durability
- **In-memory Index**: O(1) lookups with WAL offset values
- **Streaming Replication**: Real-time command streaming to followers via gRPC
- **Auto-Reconnect**: Followers automatically retry connection on failure
- **No Startup Order Dependency**: Start nodes in any order
- **Dynamic Follower Registration**: Followers connect themselves to leader
- **Low Latency**: Immediate replication on write (no polling)

## Architecture

```
┌─────────────────────────────┐
│  Leader (port 50051)        │
│  ┌──────────────────────┐   │
│  │ StreamManager        │   │
│  │ - follower1 → stream │   │
│  │ - follower2 → stream │   │
│  └──────────────────────┘   │
│         ↓ Broadcast         │
│   [WAL + In-memory Index]   │
└──────────┬──────────────────┘
           │
    ┌──────┴──────┬───────────┐
    ↓             ↓           ↓
┌─────────┐  ┌─────────┐  ┌─────────┐
│Follower1│  │Follower2│  │Follower3│
│ Stream  │  │ Stream  │  │ Stream  │
│  Recv() │  │  Recv() │  │  Recv() │
│  [WAL]  │  │  [WAL]  │  │  [WAL]  │
└─────────┘  └─────────┘  └─────────┘
```

## How It Works

### Storage Layer
- **Write-Ahead Log (WAL)**: Append-only log file storing serialized commands
- **In-memory Index**: Map of `key → WAL offset` for fast lookups
- **Crash Recovery**: On startup, replay WAL to rebuild in-memory index

### Replication Flow
1. **Follower connects**: Calls `StreamReplication()` RPC to leader
2. **Leader registers**: Adds follower to active streams map
3. **Client writes**: Leader applies to local WAL + index
4. **Broadcast**: Leader sends command to all follower streams
5. **Follower applies**: Deserializes and applies to local WAL + index

### Operations
- **SET**: Append to WAL → Update index → Broadcast to followers
- **GET**: Lookup key in index → Read from WAL at offset → Deserialize
- **DEL**: Mark as deleted in index → Broadcast to followers
- **KEYS**: Return all keys from in-memory index

## Quick Start

### Build
```bash
go build ./cmd/server/
go build ./cmd/client/
```

### Single Node Mode

**Terminal 1 - Server:**
```bash
./server --leader --port=50051
```

**Terminal 2 - Client:**
```bash
./client
> set name alice
> get name
alice
> keys
Keys (1):
  - name
> exit
```

### Replicated Mode (3 nodes)

**Start in any order - followers will auto-connect:**

**Terminal 1 - Leader:**
```bash
./server --node-id=leader --leader --port=50051
```

**Terminal 2 - Follower 1:**
```bash
./server --node-id=follower1 --port=50052 --leader-addr="localhost:50051"
# Output: Connected to leader, receiving stream
```

**Terminal 3 - Follower 2:**
```bash
./server --node-id=follower2 --port=50053 --leader-addr="localhost:50051"
# Output: Connected to leader, receiving stream
```

**Terminal 4 - Client:**
```bash
./client
> set x foo
> set y bar
> set z baz
> keys
Keys (3):
  - x
  - y
  - z
```

**Verify replication worked:**
```bash
# Check follower WAL files
cat wal-follower1.log  # Should contain x, y, z
cat wal-follower2.log  # Should contain x, y, z
cat wal-leader.log     # Should contain x, y, z
```

## Client Commands

| Command | Description | Example |
|---------|-------------|---------|
| `get {key}` | Retrieve value for key | `get username` |
| `set {key} {val}` | Store key-value pair | `set username alice` |
| `del {key}` | Delete key | `del username` |
| `keys` | List all stored keys | `keys` |
| `exit` | Close client | `exit` |

## Server Command-Line Flags

| Flag | Description | Required | Example |
|------|-------------|----------|---------|
| `--node-id` | Unique node identifier | No (default: node1) | `--node-id=leader` |
| `--leader` | Run as leader | No (default: false) | `--leader` |
| `--port` | Port to listen on | No (default: 50051) | `--port=50052` |
| `--leader-addr` | Leader address (follower only) | Yes for followers | `--leader-addr=localhost:50051` |

## Streaming Replication Details

### Connection Flow
1. **Follower → Leader**: Follower initiates gRPC `StreamReplication()` call
2. **Leader**: Creates buffered channel (100 commands) for follower
3. **Leader**: Registers channel in StreamManager
4. **Leader**: Goroutine sends commands from channel to gRPC stream
5. **Follower**: Loop receives commands via `stream.Recv()`
6. **Follower**: Applies each command to local KVS

### Failure Handling
- **Follower disconnects**: Leader detects, removes from active streams
- **Leader restarts**: Followers retry connection every 2 seconds
- **Network partition**: Followers log errors, keep retrying
- **Stream buffer full**: Command dropped with warning (configurable to block)

### Why This Design?

**Advantages over leader-initiated connections:**
- ✅ No startup order dependency
- ✅ Followers can join/leave dynamically
- ✅ Leader doesn't need to know follower addresses
- ✅ NAT/firewall friendly (outbound only from follower)
- ✅ Auto-reconnect built-in

**Trade-offs:**
- Follower misses writes during disconnect (no catch-up mechanism yet)
- Leader memory overhead: one channel + goroutine per follower

## Project Structure

```
go-kvs/
├── api/proto/              # Protocol Buffer definitions
│   ├── kvs.proto          # Client-server RPC
│   └── replication.proto  # Leader-follower streaming
├── cmd/
│   ├── client/            # Client CLI application
│   └── server/            # Server application
├── internal/
│   ├── client/            # Client gRPC wrapper
│   ├── config/            # Server configuration
│   ├── follower/          # Follower stream client
│   ├── replication/       # StreamManager for leader
│   └── server/            # gRPC server handlers
│       ├── server.go      # Client-facing handlers (Get/Set/Del/Keys)
│       ├── leader_stream.go  # Follower stream handler
│       └── middleware/    # Logging interceptor
└── pkg/kvs/               # Core KVS logic (WAL + Index)
    ├── kvs.go             # Main KVS implementation
    ├── command/           # Command serialization
    └── wal/               # Write-Ahead Log
```

## Troubleshooting

### Follower can't connect to leader
```
Error: Failed to dial leader, retrying in 2s...
```
**Solution**: Ensure leader is running and address is correct.

### Write fails with "not leader"
```
Error: not leader
```
**Solution**: Client must connect to leader (port 50051), not follower.

### Keys missing on follower
**Check**: Did follower connect before writes?
**View logs**: `cat wal-follower1.log`
**Fix**: Writes during disconnect are lost (catch-up not implemented yet).

### Build errors
```
protoc: command not found
```
**Solution**: Install Protocol Buffers compiler:
```bash
# macOS
brew install protobuf

# Ubuntu
apt-get install protobuf-compiler
```

## Future Enhancements

- [ ] **Raft Consensus**: Leader election, log consistency checks
- [ ] **Catch-up mechanism**: Followers replay missed commands
- [ ] **Snapshots**: Compact WAL, faster startup
- [ ] **Persistence of stream state**: Track last applied sequence
- [ ] **Configurable consistency**: Sync vs async replication
- [ ] **Monitoring**: Metrics, health checks, replication lag

## Technical Details

- **Language**: Go 1.19+
- **RPC Framework**: gRPC with server-side streaming
- **Serialization**: Go's `encoding/gob` for commands
- **Concurrency**: Goroutines for stream handling, sync.RWMutex for stream map
- **Logging**: zerolog (structured logging)