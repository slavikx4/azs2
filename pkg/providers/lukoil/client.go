package lukoil

import (
	"context"
	"fmt"
	"math/rand"
	"self/azs2/internal/models"
	"time"
)

type Client struct {
	providerID   int
	providerName string
}

func NewClient() *Client {
	return &Client{
		providerID:   2,
		providerName: "lukoil",
	}
}

func (c *Client) GetProviderID() int {
	return c.providerID
}

func (c *Client) GetProviderName() string {
	return c.providerName
}

func (c *Client) CreateNewCard(ctx context.Context, vehicle models.Vehicle) (models.Card, error) {
	return models.Card{
		ProviderID:     c.providerID,
		ProviderCardID: rand.Intn(1000000),
		VehicleNumber:  vehicle.PlateNumber,
	}, nil
}

func (c *Client) GetTransactions(ctx context.Context) ([]models.Transaction, error) {
	if rand.Intn(10) < 2 {
		return []models.Transaction{
			{
				ID:           rand.Int63(),
				CardID:       fmt.Sprintf("CARD_%d", rand.Intn(1000)),
				Volume:       float64(rand.Intn(50) + 10),
				Price:        55.0 + float64(rand.Intn(50))/10,
				Sum:          0,
				Date:         time.Now(),
				ProviderTxID: fmt.Sprintf("LUK_%d", time.Now().Unix()),
			},
		}, nil
	}
	return []models.Transaction{}, nil
}

func (c *Client) GetQr(ctx context.Context, cardID int) (string, error) {
	return fmt.Sprintf("QR_LUKOIL_%d", cardID), nil
}
