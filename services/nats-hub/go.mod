module github.com/Eastwesser/event-horizon/services/nats-hub

go 1.25.7

replace github.com/Eastwesser/event-horizon/platform => ../../platform

require (
	github.com/Eastwesser/event-horizon/platform v0.0.0
	github.com/nats-io/nats.go v1.52.0
)

require (
	github.com/klauspost/compress v1.19.1 // indirect
	github.com/nats-io/nkeys v0.4.15 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	golang.org/x/crypto v0.54.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
)
