package notification

import "fmt"

func (r *ListNotificationsRequest) Validate() error {
	if r.GetUserId() == "" {
		return fmt.Errorf("user_id is required")
	}
	return nil
}

func (r *MarkReadRequest) Validate() error {
	if r.GetUserId() == "" {
		return fmt.Errorf("user_id is required")
	}
	return nil
}
