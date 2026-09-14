package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"task18/internal/calendar"
)

type Server struct {
	Svc *calendar.Service
}

type respOK struct {
	Result interface{} `json:"result"`
}
type respErr struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func parseUserID(r *http.Request) (int, error) {
	return strconv.Atoi(r.FormValue("user_id"))
}

func parseDate(r *http.Request) (time.Time, error) {
	s := r.FormValue("date")
	return time.Parse("2006-01-02", s)
}

func (s *Server) CreateEvent(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeJSON(w, 400, respErr{Error: err.Error()})
		return
	}
	uid, err := parseUserID(r)
	if err != nil {
		writeJSON(w, 400, respErr{Error: "invalid user_id"})
		return
	}
	d, err := parseDate(r)
	if err != nil {
		writeJSON(w, 400, respErr{Error: "invalid date"})
		return
	}
	ev := calendar.Event{UserID: uid, Date: d, Text: r.FormValue("event")}
	if err := s.Svc.Create(ev); err != nil {
		if err == calendar.ErrDuplicate {
			writeJSON(w, 503, respErr{Error: err.Error()})
			return
		}
		writeJSON(w, 500, respErr{Error: err.Error()})
		return
	}
	writeJSON(w, 200, respOK{Result: "created"})
}

func (s *Server) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeJSON(w, 400, respErr{Error: err.Error()})
		return
	}
	uid, err := parseUserID(r)
	if err != nil {
		writeJSON(w, 400, respErr{Error: "invalid user_id"})
		return
	}
	d, err := parseDate(r)
	if err != nil {
		writeJSON(w, 400, respErr{Error: "invalid date"})
		return
	}
	text := r.FormValue("event")
	newText := r.FormValue("new_event")
	if err := s.Svc.Update(uid, d, text, newText); err != nil {
		if err == calendar.ErrNotFound {
			writeJSON(w, 503, respErr{Error: err.Error()})
			return
		}
		writeJSON(w, 500, respErr{Error: err.Error()})
		return
	}
	writeJSON(w, 200, respOK{Result: "updated"})
}

func (s *Server) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeJSON(w, 400, respErr{Error: err.Error()})
		return
	}
	uid, err := parseUserID(r)
	if err != nil {
		writeJSON(w, 400, respErr{Error: "invalid user_id"})
		return
	}
	d, err := parseDate(r)
	if err != nil {
		writeJSON(w, 400, respErr{Error: "invalid date"})
		return
	}
	text := r.FormValue("event")
	if err := s.Svc.Delete(uid, d, text); err != nil {
		if err == calendar.ErrNotFound {
			writeJSON(w, 503, respErr{Error: err.Error()})
			return
		}
		writeJSON(w, 500, respErr{Error: err.Error()})
		return
	}
	writeJSON(w, 200, respOK{Result: "deleted"})
}

func (s *Server) EventsForDay(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeJSON(w, 400, respErr{Error: err.Error()})
		return
	}
	uid, err := parseUserID(r)
	if err != nil {
		writeJSON(w, 400, respErr{Error: "invalid user_id"})
		return
	}
	d, err := parseDate(r)
	if err != nil {
		writeJSON(w, 400, respErr{Error: "invalid date"})
		return
	}
	evs := s.Svc.EventsForDay(uid, d)
	writeJSON(w, 200, respOK{Result: evs})
}

func (s *Server) EventsForWeek(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeJSON(w, 400, respErr{Error: err.Error()})
		return
	}
	uid, err := parseUserID(r)
	if err != nil {
		writeJSON(w, 400, respErr{Error: "invalid user_id"})
		return
	}
	d, err := parseDate(r)
	if err != nil {
		writeJSON(w, 400, respErr{Error: "invalid date"})
		return
	}
	start := calendar.StartOfWeek(d)
	end := start.AddDate(0, 0, 6)
	evs := s.Svc.EventsForRange(uid, start, end)
	writeJSON(w, 200, respOK{Result: evs})
}

func (s *Server) EventsForMonth(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeJSON(w, 400, respErr{Error: err.Error()})
		return
	}
	uid, err := parseUserID(r)
	if err != nil {
		writeJSON(w, 400, respErr{Error: "invalid user_id"})
		return
	}
	d, err := parseDate(r)
	if err != nil {
		writeJSON(w, 400, respErr{Error: "invalid date"})
		return
	}
	start := calendar.StartOfMonth(d)
	end := start.AddDate(0, 1, -1)
	evs := s.Svc.EventsForRange(uid, start, end)
	writeJSON(w, 200, respOK{Result: evs})
}
