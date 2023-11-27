# go-kvs
Persistent key-value storage.

Client-server key-value storage.

Server has Write-Ahead Log(WAL), which stores all commands.
On startup, server reads WAL line-by-line and restores state to in-memory index.
In-memory index has key as a key and WAL offset as a value.

Set operation appends serialized command to WAL and adds/updates in-memory index.

Get operation reads key from in-memory index. If it exists, reads line of file from offset and deserializes. 

Del operation reads key from in-memory index. If it exists, appends serialized command and deletes key from in-memory index.

Commands:
- get {key}
- set {key} {val}
- del {key}