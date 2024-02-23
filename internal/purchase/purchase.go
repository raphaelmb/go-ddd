package purchase

import (
	"time"

	"github.com/Rhymond/go-money"
	"github.com/google/uuid"
	coffeeco "github.com/raphaelmb/go-ddd/internal"
	"github.com/raphaelmb/go-ddd/internal/payment"
	"github.com/raphaelmb/go-ddd/internal/store"
)

type Purchase struct {
	id                 uuid.UUID
	Store              store.Store
	ProductsToPurchase []coffeeco.Product
	total              money.Money
	PaymeantMeans      payment.Means
	timeOfPurchase     time.Time
	CardToken          *string
}
