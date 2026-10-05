package validate_test

import (
	"testing"

	authorspb "github.com/Eastwesser/event-horizon/services/authors/proto"
)

func TestUpsertProfileRequest_Validate(t *testing.T) {
	cases := []struct {
		name    string
		req     *authorspb.UpsertProfileRequest
		wantErr bool
	}{
		{"valid", &authorspb.UpsertProfileRequest{UserId: "u1", DisplayName: "Alice"}, false},
		{"missing_user", &authorspb.UpsertProfileRequest{DisplayName: "Alice"}, true},
		{"missing_name", &authorspb.UpsertProfileRequest{UserId: "u1"}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.req.Validate()
			if tc.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected: %v", err)
			}
		})
	}
}

func TestGetAuthorRequest_Validate(t *testing.T) {
	if err := (&authorspb.GetAuthorRequest{}).Validate(); err == nil {
		t.Fatal("expected error")
	}
	if err := (&authorspb.GetAuthorRequest{UserId: "u1"}).Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestListAuthorsRequest_Validate(t *testing.T) {
	if err := (&authorspb.ListAuthorsRequest{}).Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestSubmitApplicationRequest_Validate(t *testing.T) {
	valid := &authorspb.SubmitApplicationRequest{
		UserId: "u1", DisplayName: "A", Motivation: "why", ContactEmail: "a@b.c",
	}
	if err := valid.Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (&authorspb.SubmitApplicationRequest{UserId: "u1", DisplayName: "A", Motivation: "why"}).Validate(); err == nil {
		t.Fatal("expected contact_email error")
	}
}

func TestGetMyApplicationRequest_Validate(t *testing.T) {
	if err := (&authorspb.GetMyApplicationRequest{}).Validate(); err == nil {
		t.Fatal("expected error")
	}
}

func TestListApplicationsRequest_Validate(t *testing.T) {
	if err := (&authorspb.ListApplicationsRequest{Status: "pending"}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (&authorspb.ListApplicationsRequest{Status: "nope"}).Validate(); err == nil {
		t.Fatal("expected error")
	}
}

func TestApproveApplicationRequest_Validate(t *testing.T) {
	if err := (&authorspb.ApproveApplicationRequest{}).Validate(); err == nil {
		t.Fatal("expected error")
	}
	if err := (&authorspb.ApproveApplicationRequest{ApplicationId: "a", ReviewerId: "r"}).Validate(); err != nil {
		t.Fatal(err)
	}
}
