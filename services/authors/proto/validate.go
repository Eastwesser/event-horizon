package authors

import "fmt"

func (r *UpsertProfileRequest) Validate() error {
	if r.GetUserId() == "" {
		return fmt.Errorf("user_id is required")
	}
	if r.GetDisplayName() == "" {
		return fmt.Errorf("display_name is required")
	}
	return nil
}

func (r *GetAuthorRequest) Validate() error {
	if r.GetUserId() == "" {
		return fmt.Errorf("user_id is required")
	}
	return nil
}

func (r *ListAuthorsRequest) Validate() error { return nil }
