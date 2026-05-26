package repository

import (
	"context"
	"fmt"
	"self/azs2/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, connString string) (*Repository, error) {
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, err
	}
	return &Repository{pool: pool}, nil
}

func (r *Repository) AddVehicle(ctx context.Context, plate string, limit float64, systemCardID string) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx, `INSERT INTO vehicles (plate_number, route_limit, system_card_id) VALUES ($1, $2, $3) RETURNING id`, plate, limit, systemCardID).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *Repository) AddProviderCard(ctx context.Context, cardID string, vehicleID int64, providerID, providerCardNumber string, balance, limit float64) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO fuel_cards (id, vehicle_id, provider_id, provider_card_number, balance, "limit") VALUES ($1, $2, $3, $4, $5, $6)`,
		cardID, vehicleID, providerID, providerCardNumber, balance, limit)
	return err
}

func (r *Repository) GetVehicles(ctx context.Context) ([]models.Vehicle, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, plate_number, fuel_level, route_limit, system_card_id, created_at, updated_at FROM vehicles ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vehicles []models.Vehicle
	for rows.Next() {
		var v models.Vehicle
		var systemCardID *string
		err := rows.Scan(&v.ID, &v.PlateNumber, &v.FuelLevel, &v.RouteLimit, &systemCardID, &v.CreatedAt, &v.UpdatedAt)
		if err != nil {
			return nil, err
		}
		v.SystemCardID = systemCardID
		vehicles = append(vehicles, v)
	}
	return vehicles, nil
}

func (r *Repository) GetVehicleByID(ctx context.Context, id int64) (*models.Vehicle, error) {
	var v models.Vehicle
	var systemCardID *string
	err := r.pool.QueryRow(ctx, `SELECT id, plate_number, fuel_level, route_limit, system_card_id, created_at, updated_at FROM vehicles WHERE id = $1`, id).Scan(&v.ID, &v.PlateNumber, &v.FuelLevel, &v.RouteLimit, &systemCardID, &v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		return nil, err
	}
	v.SystemCardID = systemCardID
	return &v, nil
}

func (r *Repository) UpdateVehicleLimit(ctx context.Context, id int64, limit float64) error {
	_, err := r.pool.Exec(ctx, `UPDATE vehicles SET route_limit = $1 WHERE id = $2`, limit, id)
	return err
}

func (r *Repository) GetProviderCardsByVehicle(ctx context.Context, vehicleID int64) ([]models.FuelCard, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, vehicle_id, provider_id, provider_card_number, balance, "limit" FROM fuel_cards WHERE vehicle_id = $1`, vehicleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []models.FuelCard
	for rows.Next() {
		var c models.FuelCard
		var providerCardNumber *string
		err := rows.Scan(&c.ID, &c.VehicleID, &c.ProviderID, &providerCardNumber, &c.Balance, &c.Limit)
		if err != nil {
			return nil, err
		}
		c.ProviderCardNumber = providerCardNumber
		cards = append(cards, c)
	}
	return cards, nil
}

func (r *Repository) UpdateProviderCardBalance(ctx context.Context, cardID string, balance float64) error {
	_, err := r.pool.Exec(ctx, `UPDATE fuel_cards SET balance = $1 WHERE id = $2`, balance, cardID)
	return err
}

func (r *Repository) GetRefuelingsByVehicle(ctx context.Context, vehicleID int64) ([]models.Refueling, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, vehicle_id, system_card_id, provider_card_id, provider_id, station_id, station_name, address, liters, price_per_liter, total_amount, fuel_type, timestamp, status, provider_tx_id, created_at FROM refuelings WHERE vehicle_id = $1 ORDER BY timestamp DESC`, vehicleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var refuelings []models.Refueling
	for rows.Next() {
		var r models.Refueling
		var providerCardID, providerID, stationID, stationName, address, providerTxID *string
		err := rows.Scan(&r.ID, &r.VehicleID, &r.SystemCardID, &providerCardID, &providerID, &stationID, &stationName, &address, &r.Liters, &r.PricePerLiter, &r.TotalAmount, &r.FuelType, &r.Timestamp, &r.Status, &providerTxID, &r.CreatedAt)
		if err != nil {
			return nil, err
		}
		r.ProviderCardID = providerCardID
		r.ProviderID = providerID
		r.StationID = stationID
		r.StationName = stationName
		r.Address = address
		r.ProviderTxID = providerTxID
		refuelings = append(refuelings, r)
	}
	return refuelings, nil
}

func (r *Repository) GetRefuelingsByMonth(ctx context.Context, year, month int) ([]models.Refueling, error) {
	start := fmt.Sprintf("%d-%02d-01", year, month)
	end := fmt.Sprintf("%d-%02d-01", year, month+1)
	rows, err := r.pool.Query(ctx, `SELECT id, vehicle_id, system_card_id, provider_card_id, provider_id, station_id, station_name, address, liters, price_per_liter, total_amount, fuel_type, timestamp, status, provider_tx_id, created_at FROM refuelings WHERE timestamp >= $1 AND timestamp < $2 ORDER BY timestamp DESC`, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var refuelings []models.Refueling
	for rows.Next() {
		var r models.Refueling
		var providerCardID, providerID, stationID, stationName, address, providerTxID *string
		err := rows.Scan(&r.ID, &r.VehicleID, &r.SystemCardID, &providerCardID, &providerID, &stationID, &stationName, &address, &r.Liters, &r.PricePerLiter, &r.TotalAmount, &r.FuelType, &r.Timestamp, &r.Status, &providerTxID, &r.CreatedAt)
		if err != nil {
			return nil, err
		}
		r.ProviderCardID = providerCardID
		r.ProviderID = providerID
		r.StationID = stationID
		r.StationName = stationName
		r.Address = address
		r.ProviderTxID = providerTxID
		refuelings = append(refuelings, r)
	}
	return refuelings, nil
}

func (r *Repository) GetLatestTelemetry(ctx context.Context, vehicleID int64) (*models.TelemetryRecord, error) {
	var t models.TelemetryRecord
	var fuelLevel, mileage, latitude, longitude, speed *float64
	var engineOn *bool
	err := r.pool.QueryRow(ctx, `SELECT id, vehicle_id, fuel_level, mileage, latitude, longitude, speed, engine_on, timestamp FROM telemetry_records WHERE vehicle_id = $1 ORDER BY timestamp DESC LIMIT 1`, vehicleID).Scan(&t.ID, &t.VehicleID, &fuelLevel, &mileage, &latitude, &longitude, &speed, &engineOn, &t.Timestamp)
	if err != nil {
		return nil, err
	}
	t.FuelLevel = fuelLevel
	t.Mileage = mileage
	t.Latitude = latitude
	t.Longitude = longitude
	t.Speed = speed
	t.EngineOn = engineOn
	return &t, nil
}

func (r *Repository) InsertTelemetry(ctx context.Context, t *models.TelemetryRecord) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO telemetry_records (vehicle_id, fuel_level, mileage, latitude, longitude, speed, engine_on, timestamp) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		t.VehicleID, t.FuelLevel, t.Mileage, t.Latitude, t.Longitude, t.Speed, t.EngineOn, t.Timestamp)
	return err
}

func (r *Repository) UpdateVehicleFuel(ctx context.Context, vehicleID int64, fuelLevel float64) error {
	_, err := r.pool.Exec(ctx, `UPDATE vehicles SET fuel_level = $1 WHERE id = $2`, fuelLevel, vehicleID)
	return err
}

func (r *Repository) AddTransaction(ctx context.Context, tx models.Transaction) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO refuelings (vehicle_id, system_card_id, provider_card_id, liters, price_per_liter, total_amount, fuel_type, timestamp, status, provider_tx_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		tx.ID, tx.CardID, tx.ProviderTxID, tx.Volume, tx.Price, tx.Sum, "АИ-95", tx.Date, "confirmed", tx.ProviderTxID)
	return err
}
