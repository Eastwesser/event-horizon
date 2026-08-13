module github.com/Eastwesser/event-horizon/services/notification

go 1.25.7

replace github.com/Eastwesser/event-horizon/platform => ../../platform

replace github.com/Eastwesser/event-horizon/contracts/events => ../../contracts/events

require (
	github.com/Eastwesser/event-horizon/contracts/events v0.0.0
	github.com/Eastwesser/event-horizon/platform v0.0.0
)
