# USE: http.Get() -> http.Do()

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type OriginalUser struct {
	ID        int     `json:"id"`
	Email     string  `json:"email"`
	Amount    float64 `json:"amount"`
	Profile   Profile `json:"profile"`
	Password  string  `json:"password"`
	Username  string  `json:"username"`
	CreatedAt string  `json:"createdAt"`
	CreatedBy string  `json:"createdBy"`
}

type Profile struct {
	Avatar     string `json:"avatar"`
	LastName   string `json:"lastName"`
	FirstName  string `json:"firstName"`
	StaticData string `json:"staticData"`
}

type ModifiedUser struct {
	ID        int      `json:"id"`
	Email     *string  `json:"email,omitempty"`
	Amount    float64  `json:"amount"`
	Profile   *Profile `json:"profile,omitempty"`
	Username  *string  `json:"username,omitempty"`
	CreatedAt string   `json:"createdAt"`
	CreatedBy string   `json:"createdBy"`
}

func handleUser(w http.ResponseWriter, r *http.Request) {
	// Создаем контекст с таймаутом 5 секунд
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	
	// Создаем HTTP клиент с контекстом
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	
	// Создаем запрос с контекстом
	req, err := http.NewRequestWithContext(ctx, "GET", "http://83.136.232.77:8091/users", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Выполняем запрос
	resp, err := client.Do(req)
	if err != nil {
		// Проверяем, была ли ошибка из-за таймаута контекста
		if ctx.Err() == context.DeadlineExceeded {
			http.Error(w, "request timeout", http.StatusGatewayTimeout)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		http.Error(w, "API request failed", resp.StatusCode)
		return
	}
	
	var origins []OriginalUser
	if err := json.NewDecoder(resp.Body).Decode(&origins); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	modifiedUsers := make([]ModifiedUser, 0, len(origins))
	for _, origin := range origins {
		select {
		case <-ctx.Done():
			// Если контекст отменен, прерываем обработку
			http.Error(w, "request cancelled", http.StatusRequestTimeout)
			return
		default:
			user := ModifiedUser{
				ID:        origin.ID,
				Amount:    origin.Amount,
				CreatedAt: origin.CreatedAt,
				CreatedBy: origin.CreatedBy,
			}
			
			if origin.Amount <= 50000 {
				email := origin.Email
				username := origin.Username
				profile := origin.Profile
				profile.StaticData = ""
				
				user.Email = &email
				user.Username = &username
				user.Profile = &profile
			}
			modifiedUsers = append(modifiedUsers, user)
		}
	}
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(modifiedUsers); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	http.HandleFunc("/users", handleUser)
	fmt.Println("Server running on :8082")
	server := &http.Server{
		Addr:         ":8082",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		fmt.Println("Server error:", err)
	}
}
```