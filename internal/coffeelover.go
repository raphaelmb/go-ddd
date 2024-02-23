package coffeeco

import (
	"github.com/google/uuid"
)

type CoffeeLover struct {
	ID           uuid.UUID
	FistName     string
	LastName     string
	EmailAddress string
}
