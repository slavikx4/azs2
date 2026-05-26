package dut

import (
	"context"
	"math/rand"
)

type Client struct{}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) GetDut(ctx context.Context, vehicleID int) (float64, error) {
	return float64(rand.Intn(100)), nil
}
