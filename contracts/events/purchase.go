package events

import "encoding/json"

// PurchasePaid is published by shop after a successful purchase (OrderPaid analogue).
type PurchasePaid struct {
	EventUUID   string `json:"event_uuid"`
	EventType   string `json:"event_type"` // "PurchasePaid"
	PurchaseUUID string `json:"purchase_uuid"`
	UserUUID    string `json:"user_uuid"`
	ItemUUID    string `json:"item_uuid"`
	Price       int32  `json:"price"`
}

func (e PurchasePaid) Marshal() ([]byte, error) {
	e.EventType = "PurchasePaid"
	return json.Marshal(e)
}

func UnmarshalPurchasePaid(b []byte) (PurchasePaid, error) {
	var e PurchasePaid
	err := json.Unmarshal(b, &e)
	return e, err
}

// PurchaseFulfilled is published by fulfillment after assembly delay (ShipAssembled analogue).
type PurchaseFulfilled struct {
	EventUUID    string `json:"event_uuid"`
	EventType    string `json:"event_type"` // "PurchaseFulfilled"
	PurchaseUUID string `json:"purchase_uuid"`
	UserUUID     string `json:"user_uuid"`
	ItemUUID     string `json:"item_uuid"`
}

func (e PurchaseFulfilled) Marshal() ([]byte, error) {
	e.EventType = "PurchaseFulfilled"
	return json.Marshal(e)
}

func UnmarshalPurchaseFulfilled(b []byte) (PurchaseFulfilled, error) {
	var e PurchaseFulfilled
	err := json.Unmarshal(b, &e)
	return e, err
}
