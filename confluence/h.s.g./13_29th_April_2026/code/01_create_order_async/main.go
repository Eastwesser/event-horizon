// CreateOrder: ответ ≤1s; долгий usecase продолжает в фоне; poll статуса.
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"time"
)

type Order struct{ Payload string }

type usecase interface {
	CreateOrder(order Order) error
}

type slowUC struct{}

func (slowUC) CreateOrder(Order) error {
	time.Sleep(1500 * time.Millisecond)
	return nil
}

type status int

const (
	statusPending status = iota
	statusOK
	statusFail
)

type store struct {
	mu   sync.Mutex
	data map[string]status
}

func (s *store) set(id string, st status) {
	s.mu.Lock()
	s.data[id] = st
	s.mu.Unlock()
}

func (s *store) get(id string) (status, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	st, ok := s.data[id]
	return st, ok
}

type Handler struct {
	uc    usecase
	store *store
}

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req struct {
		UUID    string `json:"uuid"`
		Payload string `json:"payload"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.UUID == "" || req.Payload == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "bad request"})
		return
	}
	if _, ok := h.store.get(req.UUID); ok {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"order_id": req.UUID, "message": "already accepted"})
		return
	}
	h.store.set(req.UUID, statusPending)

	done := make(chan error, 1)
	go func() {
		err := h.uc.CreateOrder(Order{Payload: req.Payload})
		if err != nil {
			h.store.set(req.UUID, statusFail)
		} else {
			h.store.set(req.UUID, statusOK)
		}
		done <- err
	}()

	select {
	case <-time.After(time.Second):
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"order_id": req.UUID,
			"message":  "Order processing",
		})
	case err := <-done:
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"order_id": req.UUID, "message": "Order created"})
	}
}

func (h *Handler) OrderStatus(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	st, ok := h.store.get(id)
	w.Header().Set("Content-Type", "application/json")
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
		return
	}
	names := map[status]string{statusPending: "pending", statusOK: "ok", statusFail: "fail"}
	_ = json.NewEncoder(w).Encode(map[string]string{"order_id": id, "status": names[st]})
}

func main() {
	h := &Handler{uc: slowUC{}, store: &store{data: map[string]status{}}}
	mux := http.NewServeMux()
	mux.HandleFunc("/orders", h.CreateOrder)
	mux.HandleFunc("/orders/status", h.OrderStatus)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/orders", "application/json",
		strings.NewReader(`{"uuid":"u-1","payload":"book"}`))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	var body map[string]string
	_ = json.NewDecoder(resp.Body).Decode(&body)
	fmt.Println("create:", resp.StatusCode, body)

	time.Sleep(1600 * time.Millisecond)
	st, err := http.Get(srv.URL + "/orders/status?id=u-1")
	if err != nil {
		panic(err)
	}
	defer st.Body.Close()
	_ = json.NewDecoder(st.Body).Decode(&body)
	fmt.Println("status:", body)
}
