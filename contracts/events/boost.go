package events

import "encoding/json"

// BoostStarted is published by game after a successful StartBoost spend (not on reuse).
type BoostStarted struct {
	EventUUID string `json:"event_uuid"`
	EventType string `json:"event_type"` // "BoostStarted"
	UserUUID  string `json:"user_uuid"`
	GameID    string `json:"game_id"`
	BoostUUID string `json:"boost_uuid"`
	CostLamps int32  `json:"cost_lamps"`
}

func (e BoostStarted) Marshal() ([]byte, error) {
	e.EventType = "BoostStarted"
	return json.Marshal(e)
}

func UnmarshalBoostStarted(b []byte) (BoostStarted, error) {
	var e BoostStarted
	err := json.Unmarshal(b, &e)
	return e, err
}
