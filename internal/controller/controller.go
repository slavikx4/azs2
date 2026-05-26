package controller

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"self/azs2/config"
	"self/azs2/internal/models"
	"self/azs2/internal/repository"
	"self/azs2/internal/service"
	"self/azs2/pkg/dut"
	"self/azs2/pkg/providers/gazprom"
	"self/azs2/pkg/providers/lukoil"
	"self/azs2/pkg/providers/rosneft"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

func Start() {
	log.Println("Starting FuelControl service...")

	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found, using environment variables")
	}

	var cfg config.Config
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}
	log.Println("Configuration loaded successfully")

	ctx := context.Background()
	repo, err := repository.New(ctx, cfg.DB.ConnectionString)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Database connection established")

	dutClient := dut.NewClient()
	gazpromClient := gazprom.NewClient()
	lukoilClient := lukoil.NewClient()
	rosneftClient := rosneft.NewClient()
	log.Println("Provider clients initialized")

	serv := service.NewService(repo, dutClient, gazpromClient, lukoilClient, rosneftClient)
	log.Println("Service layer initialized")

	serv.StartFuelWorker(ctx)
	serv.StartTransactionWorker(ctx)
	serv.StartTelemetryWorker(ctx)
	log.Println("Background workers started")

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/vehicles/add", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("POST /api/vehicles/add called from %s", r.RemoteAddr)
		var req struct {
			Plate string  `json:"plate"`
			Limit float64 `json:"limit"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("Error decoding request body: %v", err)
			http.Error(w, "Invalid request", 400)
			return
		}
		log.Printf("Adding vehicle: plate=%s, limit=%.2f", req.Plate, req.Limit)
		if err := serv.AddVehicle(r.Context(), req.Plate, req.Limit); err != nil {
			log.Printf("Error adding vehicle: %v", err)
			http.Error(w, err.Error(), 500)
			return
		}
		log.Printf("Vehicle added successfully: %s", req.Plate)
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]string{"message": "vehicle added"})
	})

	mux.HandleFunc("GET /api/vehicles", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("GET /api/vehicles called from %s", r.RemoteAddr)
		vehicles, err := serv.GetVehicles(r.Context())
		if err != nil {
			log.Printf("Error getting vehicles: %v", err)
			http.Error(w, err.Error(), 500)
			return
		}
		log.Printf("Returning %d vehicles", len(vehicles))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(vehicles)
	})

	mux.HandleFunc("GET /api/vehicles/{id}/cards", func(w http.ResponseWriter, r *http.Request) {
		idStr := strings.TrimPrefix(r.URL.Path, "/api/vehicles/")
		idStr = strings.TrimSuffix(idStr, "/cards")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			log.Printf("Error parsing vehicle ID: %v", err)
			http.Error(w, "Invalid vehicle ID", 400)
			return
		}
		log.Printf("GET /api/vehicles/%d/cards called from %s", id, r.RemoteAddr)

		cards, err := serv.GetProviderCardsByVehicle(r.Context(), id)
		if err != nil {
			log.Printf("Error getting cards for vehicle %d: %v", id, err)
			http.Error(w, err.Error(), 500)
			return
		}
		if cards == nil {
			cards = []models.FuelCard{}
		}
		log.Printf("Returning %d cards for vehicle %d", len(cards), id)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cards)
	})

	mux.HandleFunc("PUT /api/vehicles/{id}/limit", func(w http.ResponseWriter, r *http.Request) {
		idStr := strings.TrimPrefix(r.URL.Path, "/api/vehicles/")
		idStr = strings.TrimSuffix(idStr, "/limit")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			log.Printf("Error parsing vehicle ID: %v", err)
			http.Error(w, "Invalid vehicle ID", 400)
			return
		}
		log.Printf("PUT /api/vehicles/%d/limit called from %s", id, r.RemoteAddr)

		var req struct {
			Limit float64 `json:"limit"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("Error decoding request body: %v", err)
			http.Error(w, "Invalid request", 400)
			return
		}
		log.Printf("Updating limit for vehicle %d to %.2f", id, req.Limit)

		if err := serv.UpdateVehicleLimit(r.Context(), id, req.Limit); err != nil {
			log.Printf("Error updating limit: %v", err)
			http.Error(w, err.Error(), 500)
			return
		}
		log.Printf("Limit updated successfully for vehicle %d", id)
		json.NewEncoder(w).Encode(map[string]string{"message": "limit updated"})
	})

	mux.HandleFunc("PUT /api/vehicles/{id}/balance", func(w http.ResponseWriter, r *http.Request) {
		idStr := strings.TrimPrefix(r.URL.Path, "/api/vehicles/")
		idStr = strings.TrimSuffix(idStr, "/balance")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			log.Printf("Error parsing vehicle ID: %v", err)
			http.Error(w, "Invalid vehicle ID", 400)
			return
		}
		log.Printf("PUT /api/vehicles/%d/balance called from %s", id, r.RemoteAddr)

		var req struct {
			Balance float64 `json:"balance"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("Error decoding request body: %v", err)
			http.Error(w, "Invalid request", 400)
			return
		}
		log.Printf("Updating balance for vehicle %d to %.2f", id, req.Balance)

		if err := serv.UpdateCardBalance(r.Context(), id, req.Balance); err != nil {
			log.Printf("Error updating balance: %v", err)
			http.Error(w, err.Error(), 500)
			return
		}
		log.Printf("Balance updated successfully for vehicle %d", id)
		json.NewEncoder(w).Encode(map[string]string{"message": "balance updated"})
	})

	mux.HandleFunc("GET /api/vehicles/{id}/refuelings", func(w http.ResponseWriter, r *http.Request) {
		idStr := strings.TrimPrefix(r.URL.Path, "/api/vehicles/")
		idStr = strings.TrimSuffix(idStr, "/refuelings")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			log.Printf("Error parsing vehicle ID: %v", err)
			http.Error(w, "Invalid vehicle ID", 400)
			return
		}
		log.Printf("GET /api/vehicles/%d/refuelings called from %s", id, r.RemoteAddr)

		refuelings, err := serv.GetVehicleRefuelings(r.Context(), id)
		if err != nil {
			log.Printf("Error getting refuelings for vehicle %d: %v", id, err)
			http.Error(w, err.Error(), 500)
			return
		}
		log.Printf("Returning %d refuelings for vehicle %d", len(refuelings), id)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(refuelings)
	})

	mux.HandleFunc("GET /api/reports/monthly", func(w http.ResponseWriter, r *http.Request) {
		yearStr := r.URL.Query().Get("year")
		monthStr := r.URL.Query().Get("month")
		log.Printf("GET /api/reports/monthly called with year=%s, month=%s from %s", yearStr, monthStr, r.RemoteAddr)

		year, err := strconv.Atoi(yearStr)
		if err != nil {
			log.Printf("Error parsing year: %v", err)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]models.Refueling{})
			return
		}
		month, err := strconv.Atoi(monthStr)
		if err != nil {
			log.Printf("Error parsing month: %v", err)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]models.Refueling{})
			return
		}

		refuelings, err := serv.GetMonthlyReport(r.Context(), year, month)
		if err != nil {
			log.Printf("Error getting monthly report: %v", err)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]models.Refueling{})
			return
		}
		if refuelings == nil {
			refuelings = []models.Refueling{}
		}
		log.Printf("Returning %d refuelings for %d-%02d", len(refuelings), year, month)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(refuelings)
	})

	mux.HandleFunc("GET /api/vehicles/{id}/telemetry", func(w http.ResponseWriter, r *http.Request) {
		idStr := strings.TrimPrefix(r.URL.Path, "/api/vehicles/")
		idStr = strings.TrimSuffix(idStr, "/telemetry")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			log.Printf("Error parsing vehicle ID: %v", err)
			http.Error(w, "Invalid vehicle ID", 400)
			return
		}
		log.Printf("GET /api/vehicles/%d/telemetry called from %s", id, r.RemoteAddr)

		telemetry, err := serv.GetTelemetry(r.Context(), id)
		if err != nil {
			log.Printf("Error getting telemetry for vehicle %d: %v", id, err)
			http.Error(w, err.Error(), 500)
			return
		}
		log.Printf("Returning telemetry for vehicle %d", id)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(telemetry)
	})

	mux.HandleFunc("GET /api/qr", func(w http.ResponseWriter, r *http.Request) {
		vehicleNumber := r.URL.Query().Get("vehicle")
		providerIDStr := r.URL.Query().Get("provider_id")
		log.Printf("GET /api/qr called with vehicle=%s, provider_id=%s from %s", vehicleNumber, providerIDStr, r.RemoteAddr)

		providerID, err := strconv.Atoi(providerIDStr)
		if err != nil {
			log.Printf("Error parsing provider ID: %v", err)
			http.Error(w, "Invalid provider ID", 400)
			return
		}

		qr, err := serv.GetQr(r.Context(), vehicleNumber, providerID)
		if err != nil {
			log.Printf("Error getting QR: %v", err)
			http.Error(w, err.Error(), 500)
			return
		}
		log.Printf("Returning QR for vehicle %s, provider %d", vehicleNumber, providerID)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"qr": qr})
	})

	mux.Handle("/", http.FileServer(http.Dir("./front")))
	log.Println("Static file server configured for ./front")

	port := ":8080"
	log.Printf("Server starting on %s", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
