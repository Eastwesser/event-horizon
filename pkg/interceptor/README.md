# Shared interceptor reference (Week 1 / Kozirev)

Live copies used by services live under:

`services/{svc}/internal/interceptor/{logger,recovery}.go`

Keep those in sync if you change the pattern. A shared Go module was avoided because services pin different `google.golang.org/grpc` versions.
