# grassdb
A simple distributed database which like grass regrows easily
grassdb/
├── internal/
│   ├── server/        # gRPC server that handles Get/Set
│   │   └── server.go
│   ├── storage/       # WAL and in-memory key-value store
│   │   ├── wal.go
│   │   └── kvstore.go
│   └── raft/          # Raft consensus algorithm
│       ├── node.go
│       ├── log.go
│       ├── server.go
│       └── raft.go
└── proto/             # Protobuf definitions
    └── distdb.proto


## Features
✅ In-memory key-value store with persistence via WAL

✅ gRPC API to interact with the database (Set, Get)

✅ Basic Raft implementation:

Leader election (RequestVote)

Append entries placeholder (AppendEntries)

Term tracking and election timers


## How It Works
Startup

Each node initializes with a unique ID and peer list.

Key-value state is restored from a WAL file.

Client Interaction

A gRPC API allows clients to Set or Get values.

Raft Leader Election

If no leader heartbeat is received, a node starts an election.

Votes are requested via RequestVote RPC.

The leader replicates log entries via AppendEntries.

## Getting Started

1. Clone and build

git clone https://github.com/ranjan42/grassdb.git
cd grassdb
go mod tidy

2. Generate grpc code

protoc --go_out=. --go-grpc_out=. proto/grassdb.proto

3. Run the server

go run main.go

main.go should initialize the node, start the gRPC server, and connect to peers.

🛠️ To-Do
 Log replication and consistency checks

 Commit log entries to the key-value store only after quorum

 Cluster membership changes

 Snapshots and log compaction

 Metrics and observability

📚 Learn More
-> [The Raft Paper](https://chatgpt.com/c/683c1d9d-37ac-8000-892b-15e2bcef5979#:~:text=The%20Raft-,Paper,-gRPC%20in%20Go)

-> [gRPC in Go](https://grpc.io/docs/languages/go/)

📚 Learn More

-> [Write-Ahead Logs](https://en.wikipedia.org/wiki/Write-ahead_logging)

Authors
ranjan42 – Creator of PebbleDB