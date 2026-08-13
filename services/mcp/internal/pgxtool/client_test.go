package pgxtool

import "testing"

func TestValidateSelect_AllowsSelect(t *testing.T) {
	if err := ValidateSelect("SELECT id FROM users LIMIT 10"); err != nil {
		t.Fatal(err)
	}
}

func TestValidateSelect_RejectsMutation(t *testing.T) {
	cases := []string{
		"DELETE FROM users",
		"SELECT * FROM users; DROP TABLE users",
		"UPDATE users SET role='admin'",
		"INSERT INTO users VALUES (1)",
	}
	for _, c := range cases {
		if err := ValidateSelect(c); err == nil {
			t.Fatalf("expected reject for %q", c)
		}
	}
}
