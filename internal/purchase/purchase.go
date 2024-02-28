package purchase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/google/uuid"
	coffeeco "github.com/raphaelmb/go-ddd/internal"
	"github.com/raphaelmb/go-ddd/internal/loyalty"
	"github.com/raphaelmb/go-ddd/internal/payment"
	"github.com/raphaelmb/go-ddd/internal/store"
)

type StoreService interface {
	GetStoreSpecificDiscount(ctx context.Context, storeID uuid.UUID) (float32, error)
}

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
	ChargeCard(ctx context.Context, amount money.Money, cardToken string) error
}

type Service struct {
	cardService  CardChargeService
	purchaseRepo Repository
	storeStore   StoreService
}

func (s Service) CompletePurchase(ctx context.Context, storeID uuid.UUID, purchase *Purchase, coffeeBuxCard *loyalty.CoffeeBux) error {
	if err := purchase.validateandEnrich(); err != nil {
		return err
	}

	if err := s.calculateStoreSpecificDiscount(ctx, storeID, purchase); err != nil {
		return err
	}

	switch purchase.PaymentMeans {
	case payment.MEANS_CARD:
		if err := s.cardService.ChargeCard(ctx, purchase.total, *purchase.cardToken); err != nil {
			return errors.New("card change failed, cancelling purchase")
		}
	case payment.MEANS_CASH:
		// TODO

	case payment.MEANS_COFFEEBUX:
		if err := coffeeBuxCard.Pay(ctx, purchase.ProductsToPurchase); err != nil {
			return fmt.Errorf("failed to charge loyalty card: %w", err)
		}

	default:
		return errors.New("unknown payment type")
	}

	if err := s.purchaseRepo.Store(ctx, *purchase); err != nil {
		return errors.New("failed to store purchase")
	}

	if coffeeBuxCard != nil {
		coffeeBuxCard.AddStamp()
	}

	return nil
}

func (s Service) calculateStoreSpecificDiscount(ctx context.Context, storeID uuid.UUID, purchase *Purchase) error {
	discount, err := s.storeStore.GetStoreSpecificDiscount(ctx, storeID)
	if err != nil && err != store.ErrNoDiscount {
		return fmt.Errorf("failed to get discount: %w", err)
	}

	purchasePrice := purchase.total
	if discount > 0 {
		purchasePrice = *purchasePrice.Multiply(int64(100 - discount))
	}

	return nil
}
