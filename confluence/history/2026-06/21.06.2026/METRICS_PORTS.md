[denismatveev@c0der event_horizon]$ sudo ss -tlnp | grep -E "9091|9092|9093|9094|9095|9096|9097|8222|8079"
[sudo] пароль для denismatveev: 
LISTEN 0      4096         0.0.0.0:8222       0.0.0.0:*    users:(("docker-proxy",pid=265708,fd=4))   
LISTEN 0      4096               *:8079             *:*    users:(("balancer-servic",pid=412054,fd=3))
LISTEN 0      4096            [::]:8222          [::]:*    users:(("docker-proxy",pid=265720,fd=4))   
LISTEN 0      4096               *:9096             *:*    users:(("gateway-service",pid=429264,fd=3))
LISTEN 0      4096               *:9097             *:*    users:(("gateway-service",pid=429853,fd=6))
LISTEN 0      4096               *:9095             *:*    users:(("gateway-service",pid=428850,fd=3))
[denismatveev@c0der event_horizon]$ docker-compose -f deployments/docker-compose.cluster.yml up -d auth billing game leaderboard
[+] Running 4/4
 ✔ Container deployments-game-1         Started                                                                                                                        5.1s 
 ✔ Container deployments-billing-1      Started                                                                                                                        4.9s 
 ✔ Container deployments-auth-1         Started                                                                                                                        7.8s 
 ✔ Container deployments-leaderboard-1  Started                                                                                                                        5.9s 
[denismatveev@c0der event_horizon]$ sudo ss -tlnp | grep -E "9091|9092|9093|9094|9095|9096|9097|8222|8079"
LISTEN 0      4096         0.0.0.0:8222       0.0.0.0:*    users:(("docker-proxy",pid=265708,fd=4))   
LISTEN 0      4096         0.0.0.0:9091       0.0.0.0:*    users:(("docker-proxy",pid=454823,fd=4))   
LISTEN 0      4096               *:8079             *:*    users:(("balancer-servic",pid=412054,fd=3))
LISTEN 0      4096            [::]:8222          [::]:*    users:(("docker-proxy",pid=265720,fd=4))   
LISTEN 0      4096               *:9096             *:*    users:(("gateway-service",pid=429264,fd=3))
LISTEN 0      4096               *:9097             *:*    users:(("gateway-service",pid=429853,fd=6))
LISTEN 0      4096            [::]:9091          [::]:*    users:(("docker-proxy",pid=454842,fd=4))   
LISTEN 0      4096               *:9095             *:*    users:(("gateway-service",pid=428850,fd=3))
[denismatveev@c0der event_horizon]$ curl http://localhost:9091/metrics 2>/dev/null | head -5
# HELP go_gc_duration_seconds A summary of the wall-time pause (stop-the-world) duration in garbage collection cycles.
# TYPE go_gc_duration_seconds summary
go_gc_duration_seconds{quantile="0"} 0
go_gc_duration_seconds{quantile="0.25"} 0
go_gc_duration_seconds{quantile="0.5"} 0
[denismatveev@c0der event_horizon]$ curl http://localhost:9093/metrics 2>/dev/null | head -5
[denismatveev@c0der event_horizon]$ curl http://localhost:9092/metrics 2>/dev/null | head -5
[denismatveev@c0der event_horizon]$ curl http://localhost:9094/metrics 2>/dev/null | head -5
[denismatveev@c0der event_horizon]$ curl http://localhost:9095/metrics 2>/dev/null | head -5
curl http://localhost:9096/metrics 2>/dev/null | head -5
curl http://localhost:9097/metrics 2>/dev/null | head -5
# HELP gateway_request_duration_seconds Duration of HTTP requests in seconds
# TYPE gateway_request_duration_seconds histogram
gateway_request_duration_seconds_bucket{method="GET",path="/health",le="0.001"} 3
gateway_request_duration_seconds_bucket{method="GET",path="/health",le="0.005"} 3
gateway_request_duration_seconds_bucket{method="GET",path="/health",le="0.01"} 3
# HELP gateway_request_duration_seconds Duration of HTTP requests in seconds
# TYPE gateway_request_duration_seconds histogram
gateway_request_duration_seconds_bucket{method="GET",path="/health",le="0.001"} 1
gateway_request_duration_seconds_bucket{method="GET",path="/health",le="0.005"} 1
gateway_request_duration_seconds_bucket{method="GET",path="/health",le="0.01"} 1
# HELP gateway_request_duration_seconds Duration of HTTP requests in seconds
# TYPE gateway_request_duration_seconds histogram
gateway_request_duration_seconds_bucket{method="GET",path="/health",le="0.001"} 1
gateway_request_duration_seconds_bucket{method="GET",path="/health",le="0.005"} 1
gateway_request_duration_seconds_bucket{method="GET",path="/health",le="0.01"} 1
[denismatveev@c0der event_horizon]$ curl http://localhost:8079/metrics 2>/dev/null | head -5
{"error":"Too many requests. Please curl http://localhost:8222/metrics 2>/dev/null | head -5ent_horizon]$ curl http://localhost:8222/metrics 2>/dev/null | head -5
[denismatveev@c0der event_horizon]$ 