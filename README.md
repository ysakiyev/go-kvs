# go-kvs

Client-server key-value storage.

Server has Write-Ahead Log(WAL), which stores all commands.
On startup, server reads WAL line-by-line and restores state to in-memory index.
In-memory index has key as a key and WAL offset as a value.

Set operation appends serialized command to WAL and adds/updates in-memory index.

Get operation reads key from in-memory index. If it exists, reads line of file from offset and deserializes. 

Run server:
- go run ./cmd/server/main.go

Run client:
- go run ./cmd/client/main.go

Commands:
- get {key}
- set {key} {val}
- del {key}