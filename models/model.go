package models

import "time"

type Card struct {
	ID        int64   `json:"id"`
	Useralias string  `json:"useralias"`
	Credit    float64 `json:"credit"`
	Vat       float64 `json:"vat"`
}

type LogRefill struct {
	ID           int64     `json:"id"`
	Date         time.Time `json:"date"`
	Credit       float64   `json:"credit"`
	CardId       int64     `json:"card_id"`
	Desc         string    `json:"description"`
	RefillType   uint8     `json:"refill_type"`
	AddedInvoice uint8     `json:"added_invoice"`
	AgentId      int64     `json:"agent_id"`
}

type LogPayment struct {
	ID              int64     `json:"id"`
	Date            time.Time `json:"date"`
	Payment         float64   `json:"payment"`
	CardId          int64     `json:"card_id"`
	RefillId        int64     `json:"id_logrefill"`
	Desc            string    `json:"description"`
	AddedRefill     int16     `json:"added_refill"`
	PaymentType     uint8     `json:"payment_type"`
	AddedCommission uint8     `json:"added_commission"`
	AgentId         int64     `json:"agent_id"`
}
