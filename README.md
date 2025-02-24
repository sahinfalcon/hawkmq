# HawkMQ

## Overview
HawkMQ is a **high-performance, distributed, fault-tolerant message queue** written in Go.  
Designed for **low-latency, high-throughput messaging** with built-in replication, persistence, and leader election.

## Features (Planned)
- [x] **Single-node in-memory queue** (FIFO, pub/sub)
  - Thread-safe queue implementation
  - TCP-based client/server architecture
  - Basic pub/sub functionality
- [ ] **WAL (Write-Ahead Log) for persistence**
- [ ] **Multi-node replication using Raft**
- [ ] **gRPC/TCP networking**
- [ ] **Backpressure handling & monitoring**

## Getting Started

### Prerequisites
- Go 1.22+

### Setup
Clone the repo:
```sh
git clone https://github.com/sahinfalcon/hawkmq.git
cd hawkmq
```

Start the server:
```sh
go run cmd/server/main.go
```

Use the client to publish/consume messages:
```sh
# Publish a message
go run cmd/client/main.go publish "simon says"

# Consume a message
go run cmd/client/main.go consume
```

Run tests:
```sh
go test ./...
```

## Current Implementation
- Thread-safe in-memory queue using mutex for synchronization
- TCP-based client/server communication
- Basic PUBLISH/CONSUME commands
- Concurrent request handling
- Comprehensive test coverage including concurrency tests

## Roadmap

| Stage | Feature                         | Status        |
| ----- | ------------------------------- | ------------- |
| 1     | In-memory queue                 | ✅ Complete    |
| 2     | WAL-based persistence           | 🚧 In progress |
| 3     | Raft leader election & failover | ⏳ Planned     |
| 4     | Multi-node replication          | ⏳ Planned     |
| 5     | Performance optimizations       | ⏳ Planned     |

## License
MIT