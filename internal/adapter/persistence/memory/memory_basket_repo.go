//

// -----------------------this has been replaced with postgres_basket_repo now.------------------------------------

// package memory

// import (
// 	"errors"

// 	"github.com/AMANSRI99/StockSaaS/internal/app/model"
// 	"github.com/AMANSRI99/StockSaaS/internal/app/repository"

// 	"context"
// 	"sync"

// 	"github.com/google/uuid"
// )

// // InMemoryBasketRepo is an in-memory implementation of BasketRepository.
// type InMemoryBasketRepo struct {
// 	// baskets   map[uuid.UUID]model.Basket
// 	basketsMu sync.RWMutex
// }

// // NewInMemoryBasketRepo creates a new in-memory repository instance.
// // It returns the interface type, hiding the implementation details.
// func NewInMemoryBasketRepo() repository.BasketRepository {
// 	return &InMemoryBasketRepo{
// 		baskets: make(map[uuid.UUID]model.Basket),
// 		// Mutex is zero-value ready
// 	}
// }

// // Save implements the BasketRepository interface.
// func (r *InMemoryBasketRepo) Save(ctx context.Context, basket *model.Basket) error {
// 	// Check context cancellation (good practice, though less critical for in-memory)
// 	select {
// 	case <-ctx.Done():
// 		return ctx.Err()
// 	default:
// 	}

// 	r.basketsMu.Lock()
// 	defer r.basketsMu.Unlock()

// 	// Basic check if ID already exists (Save could mean update too, but we'll keep it simple for now)
// 	if _, exists := r.baskets[basket.ID]; exists {
// 		// In a real DB, this might be an UPSERT or return an error.
// 		// For now, we just overwrite for simplicity if ID is reused somehow.
// 		// Or return an error:
// 		return errors.New("basket already exists")
// 	}
// 	r.baskets[basket.ID] = *basket // Store a copy
// 	return nil
// }

// // FindAll implements the BasketRepository interface.
// func (r *InMemoryBasketRepo) FindAll(ctx context.Context) ([]model.Basket, error) {
// 	select {
// 	case <-ctx.Done():
// 		return nil, ctx.Err()
// 	default:
// 	}

// 	r.basketsMu.RLock()
// 	defer r.basketsMu.RUnlock()

// 	allBaskets := make([]model.Basket, 0, len(r.baskets))
// 	for _, basket := range r.baskets {
// 		allBaskets = append(allBaskets, basket)
// 	}
// 	return allBaskets, nil
// }

// /*
// // FindByID implements the BasketRepository interface (Example for later)
// func (r *InMemoryBasketRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Basket, error) {
// 	select {
// 	case <-ctx.Done():
// 		return nil, ctx.Err()
// 	default:
// 	}

// 	r.basketsMu.RLock()
// 	defer r.basketsMu.RUnlock()

// 	basket, exists := r.baskets[id]
// 	if !exists {
// 		return nil, repository.ErrBasketNotFound // Use the defined error
// 	}
// 	return &basket, nil // Return pointer to a copy
// }
// 