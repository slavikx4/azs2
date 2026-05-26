package main

import (
	"log"
	"self/azs2/internal/controller"
)

func main() {
	log.Println("Starting FuelControl service...")
	controller.Start()
}
