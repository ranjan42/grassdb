# grassdb - Distributed Key-Value Store

`grassdb` is an enterprise-grade distributed key-value store written in Go. It uses the Raft consensus algorithm to ensure data consistency and high availability across a cluster of nodes. Just like grass, it's designed to be resilient and "regrow" easily after failures.

---

## 🌟 Key Features

*   **Distributed Consensus**: Implements the Raft consensus algorithm (Leader Election, Log Replication).
*   **Strong Consistency**: Writes are directed to the leader and replicated to followers.
*   **High Availability**: The cluster continues to operate as long as a quorum (majority) of nodes are up.
*   **Persistence**: Uses a Write-Ahead Log (WAL) to persist data to disk, ensuring survivability across restarts.
*   **gRPC API**: Modern, high-performance API for all interactions (Client-to-Node and Node-to-Node).

---

## 🏗 Architecture

`grassdb` is built on a modular architecture:

1.  **API Layer (gRPC)**: Handles client requests (`Get`, `Set`) and internal Raft RPCs (`RequestVote`, `AppendEntries`).
2.  **Consensus Layer (Raft)**: Manages the distributed state machine. It handles leader election, heartbeat mechanism, and log replication.
3.  **Storage Layer**:
    *   **In-Memory Store**: Fast access to current state.
    *   **WAL (Write-Ahead Log)**: Appends every write operation to a disk file for durability.

---

## 🚀 Getting Started

### Prerequisites

*   Go 1.24+
*   Protobuf Compiler (`protoc`) with Go plugins

### Installation

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/ranjan42/grassdb.git
    cd grassdb
    ```

2.  **Install dependencies:**
    ```bash
    go mod tidy
    ```

3.  **Generate Protobuf code (if modifying .proto files):**
    ```bash
    protoc --go_out=. --go-grpc_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative proto/grassdb.proto
    ```

### Running a Local Cluster

We provide a helper script to launch a 3-node cluster locally.

1.  **Build the binary:**
    ```bash
    go build -o grassdb main.go
    ```

2.  **Start the cluster:**
    ```bash
    bash start_cluster.sh
    ```
    This will start 3 nodes on ports `:50051`, `:50052`, and `:50053`.

3.  **Verify Leader Election:**
    Check the logs (e.g., `node1.log`) to see which node became the leader:
    ```bash
    grep "Won election" node*.log
    ```

---

## 🔧 Usage

You can interact with the cluster using a gRPC client. Since `grassdb` uses standard gRPC, you can write a simple client in Go, Python, or efficient CLI tools like `grpcurl`.

**Note**: All write operations (`Set`) must be sent to the **Leader**. If you send a write to a Follower, it will return an error indicating who the Leader is (Redirect logic to be fully implemented).

### Example Client interaction

### Client CLI Tool

You can use the built-in CLI tool to interact with the cluster. It automatically handles leader discovery and retries.

1. **Build the CLI:**
   ```bash
   go build -o grass-cli cmd/grass-cli/main.go
   ```

2. **Set a value:**
   ```bash
   ./grass-cli set mykey myvalue
   ```

3. **Get a value:**
   ```bash
   ./grass-cli get mykey
   ```

4. **Custom Peers:**
   If running on different ports/hosts:
   ```bash
   ./grass-cli -peers=host1:50051,host2:50052 set foo bar
   ```

---

## 📚 Internals & Design

### Raft Implementation Details
*   **Leader Election**: Randomized election timeouts (300-600ms) to prevent split votes.
*   **Heartbeats**: Leader sends heartbeats every 100ms to maintain authority.
*   **Transport**: Persistent gRPC connections are established between peers to minimize connection overhead.

### Data Persistence
Each node maintains its own `distdb_<node_id>.wal` file. On startup, the node replays this WAL to restore its state before joining the cluster.

---

## 🤝 Contributing

Contributions are welcome! Please check the [Task List](task.md) for current roadmap items.

1.  Fork the repo
2.  Create your feature branch (`git checkout -b feature/amazing-feature`)
3.  Commit your changes (`git commit -m 'Add some amazing feature'`)
4.  Push to the branch (`git push origin feature/amazing-feature`)
5.  Open a Pull Request

---

## 📜 License

Distributed under the MIT License. See `LICENSE` for more information.
