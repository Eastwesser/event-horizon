package history

import "fmt"

func (r *RecordEventRequest) Validate() error {
	if r.GetEventType() == "" {
		return fmt.Errorf("event_type is required")
	}
	return nil
}

func (r *ListEventsRequest) Validate() error { return nil }
