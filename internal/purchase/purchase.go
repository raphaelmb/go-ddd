package purchase

import (
	"context"
	"errors"
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
	PaymentMeans       payment.Means
	timeOfPurchase     time.Time
	cardToken          *string
}

func (p *Purchase) validateandEnrich() error {
	if len(p.ProductsToPurchase) == 0 {
		return errors.New("purchase must consist of at least one prouduct")
	}

	p.total = *money.New(0, "USD")

	for _, v := range p.ProductsToPurchase {
		newTotal, _ := p.total.Add(&v.BasePrice)
		p.total = *newTotal
	}

	if p.total.IsZero() {
		return errors.New("purchase should not be zero")
	}

	p.id = uuid.New()
	p.timeOfPurchase = time.Now()

	return nil
}

type CardChargeService interface {
	ChangeCard(ctx context.Context, amount money.Money, cardToken string) error
}

type Service struct {
	cardService  CardChargeService
	purchaseRepo Repository
}

func (s Service) CompletePurchase(ctx context.Context, purchase *Purchase) error {
	if err := purchase.validateandEnrich(); err != nil {
		return err
	}

	switch purchase.PaymentMeans {
	case payment.MEANS_CARD:
		if err := s.cardService.ChangeCard(ctx, purchase.total, *purchase.cardToken); err != nil {
			return errors.New("card change failed")
		}
	case payment.MEANS_CASH:
	// TODO

	default:
		return errors.New("unknown payment type")
	}

	if err := s.purchaseRepo.Store(ctx, *purchase); err != nil {
		return errors.New("failed to store purchase")
	}

	return nil
}
