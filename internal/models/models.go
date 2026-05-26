package models

import (
	"context"
	"time"
)

type FuelType string
type CardStatus string
type RefuelStatus string

const (
	FuelAI92 FuelType = "АИ-92"
	FuelAI95 FuelType = "АИ-95"
	FuelAI98 FuelType = "АИ-98"
	FuelDT   FuelType = "ДТ"
)

type Vehicle struct {
	ID           int64     `json:"id" db:"id"`
	PlateNumber  string    `json:"plate" db:"plate_number"`
	FuelLevel    float64   `json:"fuel_level" db:"fuel_level"`
	RouteLimit   float64   `json:"route_limit" db:"route_limit"`
	SystemCardID *string   `json:"system_card_id" db:"system_card_id"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type FuelCard struct {
	ID                 string     `json:"id" db:"id"`
	VehicleID          int64      `json:"vehicle_id" db:"vehicle_id"`
	ProviderID         string     `json:"provider_id" db:"provider_id"`
	ProviderCardNumber *string    `json:"provider_card_number" db:"provider_card_number"`
	Balance            float64    `json:"balance" db:"balance"`
	Limit              float64    `json:"limit" db:"limit"`
	Status             CardStatus `json:"status" db:"status"`
	IsActive           bool       `json:"is_active" db:"is_active"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
}

type Refueling struct {
	ID             int64        `json:"id" db:"id"`
	VehicleID      int64        `json:"vehicle_id" db:"vehicle_id"`
	SystemCardID   string       `json:"system_card_id" db:"system_card_id"`
	ProviderCardID *string      `json:"provider_card_id" db:"provider_card_id"`
	ProviderID     *string      `json:"provider_id" db:"provider_id"`
	StationID      *string      `json:"station_id" db:"station_id"`
	StationName    *string      `json:"station_name" db:"station_name"`
	Address        *string      `json:"address" db:"address"`
	Liters         float64      `json:"liters" db:"liters"`
	PricePerLiter  float64      `json:"price_per_liter" db:"price_per_liter"`
	TotalAmount    float64      `json:"total_amount" db:"total_amount"`
	FuelType       FuelType     `json:"fuel_type" db:"fuel_type"`
	Timestamp      time.Time    `json:"timestamp" db:"timestamp"`
	Status         RefuelStatus `json:"status" db:"status"`
	ProviderTxID   *string      `json:"provider_tx_id" db:"provider_tx_id"`
	CreatedAt      time.Time    `json:"created_at" db:"created_at"`
}

type TelemetryRecord struct {
	ID        int64     `json:"id" db:"id"`
	VehicleID int64     `json:"vehicle_id" db:"vehicle_id"`
	FuelLevel *float64  `json:"fuel_level" db:"fuel_level"`
	Mileage   *float64  `json:"mileage" db:"mileage"`
	Latitude  *float64  `json:"latitude" db:"latitude"`
	Longitude *float64  `json:"longitude" db:"longitude"`
	Speed     *float64  `json:"speed" db:"speed"`
	EngineOn  *bool     `json:"engine_on" db:"engine_on"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
}

type Transaction struct {
	ID           int64     `json:"id"`
	CardID       string    `json:"card_id"`
	Volume       float64   `json:"volume"`
	Price        float64   `json:"price"`
	Sum          float64   `json:"sum"`
	Date         time.Time `json:"date"`
	ProviderTxID string    `json:"provider_tx_id"`
}

type Card struct {
	ID             int    `json:"id"`
	ProviderID     int    `json:"provider_id"`
	ProviderCardID int    `json:"provider_card_id"`
	VehicleNumber  string `json:"vehicle_number"`
}

type ProviderClient interface {
	GetProviderID() int
	GetProviderName() string
	CreateNewCard(ctx context.Context, vehicle Vehicle) (Card, error)
	GetTransactions(ctx context.Context) ([]Transaction, error)
	GetQr(ctx context.Context, cardID int) (string, error)
}

type DutClient interface {
	GetDut(ctx context.Context, vehicleID int) (float64, error)
}
