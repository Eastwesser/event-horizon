package authors

import (
	"fmt"
	"strings"
)

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

func (r *SubmitApplicationRequest) Validate() error {
	if r.GetUserId() == "" {
		return fmt.Errorf("user_id is required")
	}
	if strings.TrimSpace(r.GetDisplayName()) == "" {
		return fmt.Errorf("display_name is required")
	}
	if strings.TrimSpace(r.GetMotivation()) == "" {
		return fmt.Errorf("motivation is required")
	}
	if strings.TrimSpace(r.GetContactEmail()) == "" {
		return fmt.Errorf("contact_email is required")
	}
	return nil
}

func (r *GetMyApplicationRequest) Validate() error {
	if r.GetUserId() == "" {
		return fmt.Errorf("user_id is required")
	}
	return nil
}
