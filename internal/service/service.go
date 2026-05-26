package service

import (
	"context"
	"fmt"
	"math/rand"
	"self/azs2/internal/models"
	"self/azs2/internal/repository"
	"time"
)

type Service struct {
	repo      *repository.Repository
	providers map[string]models.ProviderClient
	dutClient models.DutClient
}

func NewService(repo *repository.Repository, dutClient models.DutClient, providers ...models.ProviderClient) *Service {
	providersMap := make(map[string]models.ProviderClient)
	for _, p := range providers {
		providersMap[p.GetProviderName()] = p
	}
	return &Service{
		repo:      repo,
		providers: providersMap,
		dutClient: dutClient,
	}
}

func (s *Service) AddVehicle(ctx context.Context, plate string, limit float64) error {
	systemCardID := fmt.Sprintf("%06d", rand.Intn(1000000))

	vehicleID, err := s.repo.AddVehicle(ctx, plate, limit, systemCardID)
	if err != nil {
		return err
	}

	providerConfigs := []struct {
		id   string
		name string
	}{
		{"gazprom", "Газпромнефть"},
		{"lukoil", "Лукойл"},
		{"rosneft", "Роснефть"},
	}

	for _, pc := range providerConfigs {
		provider, ok := s.providers[pc.id]
		if !ok {
			continue
		}
		card, err := provider.CreateNewCard(ctx, models.Vehicle{PlateNumber: plate, RouteLimit: limit})
		if err != nil {
			continue
		}
		cardID := fmt.Sprintf("FC-%s-%s", pc.id, systemCardID)
		providerCardNumber := fmt.Sprintf("%s%d", pc.id, card.ProviderCardID%10000)
		s.repo.AddProviderCard(ctx, cardID, vehicleID, pc.id, providerCardNumber, limit, limit)
	}

	return nil
}

func (s *Service) GetVehicles(ctx context.Context) ([]models.Vehicle, error) {
	return s.repo.GetVehicles(ctx)
}

func (s *Service) UpdateVehicleLimit(ctx context.Context, id int64, newLimit float64) error {
	cards, err := s.repo.GetProviderCardsByVehicle(ctx, id)
	if err != nil {
		return err
	}

	for _, card := range cards {
		if card.Balance > newLimit {
			s.repo.UpdateProviderCardBalance(ctx, card.ID, newLimit)
		}
	}

	return s.repo.UpdateVehicleLimit(ctx, id, newLimit)
}

func (s *Service) UpdateCardBalance(ctx context.Context, vehicleID int64, newBalance float64) error {
	cards, err := s.repo.GetProviderCardsByVehicle(ctx, vehicleID)
	if err != nil {
		return err
	}

	for _, card := range cards {
		if card.ProviderID == "lukoil" {
			return s.repo.UpdateProviderCardBalance(ctx, card.ID, newBalance)
		}
	}
	return fmt.Errorf("card not found")
}

func (s *Service) GetVehicleRefuelings(ctx context.Context, vehicleID int64) ([]models.Refueling, error) {
	return s.repo.GetRefuelingsByVehicle(ctx, vehicleID)
}

func (s *Service) GetMonthlyReport(ctx context.Context, year, month int) ([]models.Refueling, error) {
	return s.repo.GetRefuelingsByMonth(ctx, year, month)
}

func (s *Service) GetTelemetry(ctx context.Context, vehicleID int64) (*models.TelemetryRecord, error) {
	return s.repo.GetLatestTelemetry(ctx, vehicleID)
}

func (s *Service) StartFuelWorker(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for range ticker.C {
			vehicles, _ := s.repo.GetVehicles(ctx)
			for _, v := range vehicles {
				if rand.Intn(10) < 3 {
					newFuel := v.FuelLevel - float64(rand.Intn(3)+1)
					if newFuel < 20 {
						newFuel = 20
					}
					s.repo.UpdateVehicleFuel(ctx, v.ID, newFuel)
				}
			}
		}
	}()
}

func (s *Service) StartTransactionWorker(ctx context.Context) {
	ticker := time.NewTicker(45 * time.Second)
	go func() {
		for range ticker.C {
			for _, provider := range s.providers {
				transactions, _ := provider.GetTransactions(ctx)
				for _, tx := range transactions {
					s.repo.AddTransaction(ctx, tx)
				}
			}
		}
	}()
}

func (s *Service) StartTelemetryWorker(ctx context.Context) {
	ticker := time.NewTicker(20 * time.Second)
	go func() {
		for range ticker.C {
			vehicles, _ := s.repo.GetVehicles(ctx)
			for _, v := range vehicles {
				dutValue, _ := s.dutClient.GetDut(ctx, int(v.ID))
				fuelLevel := dutValue
				telemetry := &models.TelemetryRecord{
					VehicleID: v.ID,
					FuelLevel: &fuelLevel,
					Mileage:   floatPtr(10000 + float64(rand.Intn(50000))),
					Latitude:  floatPtr(55.75 + (float64(rand.Intn(100)) / 1000)),
					Longitude: floatPtr(37.61 + (float64(rand.Intn(100)) / 1000)),
					Speed:     floatPtr(float64(rand.Intn(90))),
					EngineOn:  boolPtr(rand.Intn(2) == 1),
					Timestamp: time.Now(),
				}
				s.repo.InsertTelemetry(ctx, telemetry)
			}
		}
	}()
}
func (s *Service) GetQr(ctx context.Context, vehicleNumber string, providerID int) (string, error) {
	vehicles, err := s.repo.GetVehicles(ctx)
	if err != nil {
		return "", err
	}
	var targetVehicle *models.Vehicle
	for i := range vehicles {
		if vehicles[i].PlateNumber == vehicleNumber {
			targetVehicle = &vehicles[i]
			break
		}
	}
	if targetVehicle == nil {
		return "", fmt.Errorf("vehicle not found")
	}

	providerName := ""
	switch providerID {
	case 1:
		providerName = "gazprom"
	case 2:
		providerName = "lukoil"
	case 3:
		providerName = "rosneft"
	}

	provider, ok := s.providers[providerName]
	if !ok {
		return "", fmt.Errorf("provider not found")
	}
	return provider.GetQr(ctx, 0)
}

func (s *Service) GetProviderCardsByVehicle(ctx context.Context, vehicleID int64) ([]models.FuelCard, error) {
	return s.repo.GetProviderCardsByVehicle(ctx, vehicleID)
}

func floatPtr(f float64) *float64 { return &f }
func boolPtr(b bool) *bool        { return &b }
