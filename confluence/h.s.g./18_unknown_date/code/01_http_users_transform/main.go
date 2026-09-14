package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"
)

// OriginalUser — «толстый» ответ внешнего API.
type OriginalUser struct {
	ID      int     `json:"id"`
	Email   string  `json:"email"`
	Amount  float64 `json:"amount"`
	Profile struct {
		Avatar     string `json:"avatar"`
		LastName   string `json:"lastName"`
		FirstName  string `json:"firstName"`
		StaticData string `json:"staticData"`
	} `json:"profile"`
	Password  string `json:"password"`
	Username  string `json:"username"`
	CreatedAt string `json:"createdAt"`
	CreatedBy string `json:"createdBy"`
}

// ModifiedUser — публичный DTO: password/staticData выкинуты;
// PII (email/username/ФИО/avatar) только если Amount <= 50000.
type ModifiedUser struct {
	ID      int     `json:"id"`
	Email   string  `json:"email,omitempty"`
	Amount  float64 `json:"amount"`
	Profile struct {
		Avatar    string `json:"avatar,omitempty"`
		LastName  string `json:"lastName,omitempty"`
		FirstName string `json:"firstName,omitempty"`
	} `json:"profile"`
	Username  string `json:"username,omitempty"`
	CreatedAt string `json:"createdAt"`
	CreatedBy string `json:"createdBy"`
}

func transform(originals []OriginalUser) []ModifiedUser {
	out := make([]ModifiedUser, 0, len(originals))
	for _, o := range originals {
		m := ModifiedUser{
			ID: o.ID, Amount: o.Amount,
			CreatedAt: o.CreatedAt, CreatedBy: o.CreatedBy,
		}
		if o.Amount <= 50000 {
			m.Email = o.Email
			m.Username = o.Username
			m.Profile.FirstName = o.Profile.FirstName
			m.Profile.LastName = o.Profile.LastName
			m.Profile.Avatar = o.Profile.Avatar
		}
		out = append(out, m)
	}
	return out
}

func userHandler(upstream string) http.HandlerFunc {
	client := &http.Client{Timeout: 10 * time.Second}
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := client.Get(upstream)
		if err != nil {
			http.Error(w, "upstream failed", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		var originals []OriginalUser
		if err := json.NewDecoder(resp.Body).Decode(&originals); err != nil {
			http.Error(w, "decode failed", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(transform(originals))
	}
}

func main() {
	// Локальный «внешний» API — без реальной сети.
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u1 := OriginalUser{ID: 1, Email: "a@x", Amount: 1000, Username: "alice", Password: "secret", CreatedAt: "t1", CreatedBy: "sys"}
		u1.Profile.Avatar, u1.Profile.LastName, u1.Profile.FirstName, u1.Profile.StaticData = "ava", "A", "Alice", "static"
		u2 := OriginalUser{ID: 2, Email: "b@x", Amount: 90000, Username: "bob", Password: "secret", CreatedAt: "t2", CreatedBy: "sys"}
		u2.Profile.Avatar, u2.Profile.LastName, u2.Profile.FirstName, u2.Profile.StaticData = "ava2", "B", "Bob", "static"
		_ = json.NewEncoder(w).Encode([]OriginalUser{u1, u2})
	}))
	defer upstream.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/users", userHandler(upstream.URL))

	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/users")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	var got []ModifiedUser
	_ = json.NewDecoder(resp.Body).Decode(&got)
	b, _ := json.MarshalIndent(got, "", "  ")
	fmt.Println(string(b))
}
