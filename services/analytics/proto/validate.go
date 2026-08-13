package analytics

import "fmt"

func (r *RecordEventRequest) Validate() error {
	if r.GetEventType() == "" {
		return fmt.Errorf("event_type is required")
	}
	return nil
}

func (r *GetDAURequest) Validate() error       { return nil }
func (r *GetMAURequest) Validate() error       { return nil }
func (r *GetRetentionRequest) Validate() error { return nil }
