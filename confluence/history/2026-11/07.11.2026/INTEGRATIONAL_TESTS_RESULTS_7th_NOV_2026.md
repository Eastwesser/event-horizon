[denismatveev@c0der authors]$ go test -tags=integration ./internal/repository/... -run TestRevertApplication -count=1 -v
=== RUN   TestRevertApplication_SoftDeletesAuthor
2026/11/07 17:44:58 github.com/testcontainers/testcontainers-go - Connected to docker: 
  Server Version: 27.3.1
  API Version: 1.47
  Operating System: Arch Linux
  Total Memory: 9809 MB
  Testcontainers for Go Version: v0.37.0
  Resolved Docker Host: unix:///var/run/docker.sock
  Resolved Docker Socket Path: /var/run/docker.sock
  Test SessionID: 87460f07dbe6463b92006c86ad3dbc8ba836f54022218e75494a6c8e9889badb
  Test ProcessID: 45c591c9-654d-428c-b4e7-e76c43457a9b
2026/11/07 17:44:58 🐳 Creating container for image postgres:16-alpine
2026/11/07 17:44:58 🐳 Creating container for image testcontainers/ryuk:0.11.0
2026/11/07 17:44:58 ✅ Container created: d2d1c5b507a2
2026/11/07 17:44:58 🐳 Starting container: d2d1c5b507a2
2026/11/07 17:44:59 ✅ Container started: d2d1c5b507a2
2026/11/07 17:44:59 ⏳ Waiting for container id d2d1c5b507a2 image: testcontainers/ryuk:0.11.0. Waiting for: &{Port:8080/tcp timeout:<nil> PollInterval:100ms skipInternalCheck:false}
2026/11/07 17:45:00 🔔 Container is ready: d2d1c5b507a2
2026/11/07 17:45:00 ✅ Container created: 3d580d63c500
2026/11/07 17:45:00 🐳 Starting container: 3d580d63c500
2026/11/07 17:45:00 ✅ Container started: 3d580d63c500
2026/11/07 17:45:00 ⏳ Waiting for container id 3d580d63c500 image: postgres:16-alpine. Waiting for: &{Port:5432/tcp timeout:0x1adc9f427948 PollInterval:100ms skipInternalCheck:false}
2026/11/07 17:45:02 🔔 Container is ready: 3d580d63c500
2026/11/07 17:45:06 🐳 Stopping container: 3d580d63c500
2026/11/07 17:45:07 ✅ Container stopped: 3d580d63c500
2026/11/07 17:45:07 🐳 Terminating container: 3d580d63c500
2026/11/07 17:45:07 🚫 Container terminated: 3d580d63c500
--- PASS: TestRevertApplication_SoftDeletesAuthor (9.73s)
PASS
ok      github.com/Eastwesser/event-horizon/services/authors/internal/repository        9.805s
[denismatveev@c0der authors]$ cd ~/event_horizon
make test-integration
Integration tests (testcontainers; needs Docker OR *_TEST_DATABASE_URL)
ok      github.com/Eastwesser/event-horizon/services/billing/internal/repository        15.166s
ok      github.com/Eastwesser/event-horizon/services/shop/internal/repository   9.089s
ok      github.com/Eastwesser/event-horizon/services/inventory/internal/repository      70.563s
ok      github.com/Eastwesser/event-horizon/services/payment/internal/repository        7.530s
ok      github.com/Eastwesser/event-horizon/services/authors/internal/repository        16.112s
ok      github.com/Eastwesser/event-horizon/services/history/internal/repository        7.196s
ok      github.com/Eastwesser/event-horizon/services/analytics/internal/repository/clickhouse   11.669s
[denismatveev@c0der event_horizon]$ 