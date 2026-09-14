package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
)

type Employee struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

var employees = []Employee{
	{1, "Alice", "ACTIVE"},
	{2, "Bob", "INACTIVE"},
	{3, "Charlie", "ACTIVE"},
}

func getEmployeesHandler(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	var result []Employee
	if status != "" {
		for _, e := range employees {
			if e.Status == status {
				result = append(result, e)
			}
		}
	} else {
		result = employees
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

func main() {
	srv := httptest.NewServer(http.HandlerFunc(getEmployeesHandler))
	defer srv.Close()
	resp, _ := http.Get(srv.URL + "?status=ACTIVE")
	defer resp.Body.Close()
	var got []Employee
	_ = json.NewDecoder(resp.Body).Decode(&got)
	fmt.Println(got)
}
