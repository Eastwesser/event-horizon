[denismatveev@c0der event_horizon]$ cd services/shop && GOWORK=off go test -tags=integration ./internal/repository/ -count=1 -timeout 5m -v
=== RUN   TestShop_PurchaseItemWithStock_AndVersionBump
2026/11/11 18:43:28 github.com/testcontainers/testcontainers-go - Connected to docker: 
  Server Version: 27.3.1
  API Version: 1.47
  Operating System: Arch Linux
  Total Memory: 9809 MB
  Testcontainers for Go Version: v0.37.0
  Resolved Docker Host: unix:///var/run/docker.sock
  Resolved Docker Socket Path: /var/run/docker.sock
  Test SessionID: 5d8e57081f4ce4de98fb52b14b3fe255e15274cad6e14b3df33d2c07627b9f66
  Test ProcessID: ca141e77-8b7f-41f2-a035-c84c0eabad89
2026/11/11 18:43:28 🐳 Creating container for image postgres:16-alpine
2026/11/11 18:43:31 🐳 Creating container for image testcontainers/ryuk:0.11.0
2026/11/11 18:43:32 ✅ Container created: 1f316bccda57
2026/11/11 18:43:32 🐳 Starting container: 1f316bccda57
2026/11/11 18:43:36 ✅ Container started: 1f316bccda57
2026/11/11 18:43:36 ⏳ Waiting for container id 1f316bccda57 image: testcontainers/ryuk:0.11.0. Waiting for: &{Port:8080/tcp timeout:<nil> PollInterval:100ms skipInternalCheck:false}
2026/11/11 18:43:37 🔔 Container is ready: 1f316bccda57
2026/11/11 18:43:37 ✅ Container created: c873f4b17cc0
2026/11/11 18:43:37 🐳 Starting container: c873f4b17cc0
2026/11/11 18:43:40 ✅ Container started: c873f4b17cc0
2026/11/11 18:43:40 ⏳ Waiting for container id c873f4b17cc0 image: postgres:16-alpine. Waiting for: &{Port:5432/tcp timeout:0x200f5957c8f8 PollInterval:100ms skipInternalCheck:false}
2026/11/11 18:43:50 🔔 Container is ready: c873f4b17cc0
2026/11/11 18:43:53 🐳 Stopping container: c873f4b17cc0
2026/11/11 18:43:57 ✅ Container stopped: c873f4b17cc0
2026/11/11 18:43:57 🐳 Terminating container: c873f4b17cc0
2026/11/11 18:43:57 🚫 Container terminated: c873f4b17cc0
--- PASS: TestShop_PurchaseItemWithStock_AndVersionBump (30.09s)
PASS
ok      github.com/Eastwesser/event-horizon/services/shop/internal/repository   30.317s
[denismatveev@c0der shop]$ 