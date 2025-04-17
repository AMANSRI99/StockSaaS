package service

import (
	// Use your actual module path
	"errors"

	"github.com/AMANSRI99/StockSaaS/internal/app/model"
	"github.com/AMANSRI99/StockSaaS/internal/app/repository"

	"context"
	"fmt" // For potential validation errors
	"log"
	"time"

	"github.com/google/uuid"
)

// --- Interface Definition ---

// BasketService defines the interface for basket business logic operations.
type BasketService interface {
	CreateBasket(ctx context.Context, name string, stocks []model.Stock) (*model.Basket, error)
	ListAllBaskets(ctx context.Context) ([]model.Basket, error)
	GetBasketByID(ctx context.Context, id uuid.UUID) (*model.Basket, error)
	DeleteBasketByID(ctx context.Context, id uuid.UUID) error
	UpdateBasket(ctx context.Context, id uuid.UUID, name string, stocks []model.Stock) (*model.Basket, error)
}

// --- Implementation ---

// basketService implements the BasketService interface. Since it's an implementation hence the small b.
// It's unexported (starts with lowercase 'b') as users should interact via the interface.
type basketService struct {
	repo repository.BasketRepository // Dependency on repository interface
}

// NewBasketService creates a new service instance with its dependencies.
// It returns the interface type.
func NewBasketService(repo repository.BasketRepository) BasketService {
	return &basketService{
		repo: repo,
	}
}

// CreateBasket contains the business logic for creating a new basket.
func (s *basketService) CreateBasket(ctx context.Context, name string, stocks []model.Stock) (*model.Basket, error) {
	// 1. Input Validation (Could be more extensive business rules here)
	if name == "" {
		return nil, fmt.Errorf("basket name cannot be empty") // Return specific errors if needed
	}
	if len(stocks) == 0 {
		return nil, fmt.Errorf("basket must contain at least one stock")
	}
	for i, stock := range stocks {
		if stock.Symbol == "" || stock.Quantity <= 0 {
			return nil, fmt.Errorf("invalid data for stock #%d: symbol and positive quantity required", i+1)
		}
		// Add more business rules? e.g., check if stock symbol exists in a master list?
	}

	// 2. Create the domain model object
	newBasket := model.Basket{
		ID:        uuid.New(), // Service is responsible for generating ID
		Name:      name,
		Stocks:    stocks,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	// 3. Persist using the repository
	log.Printf("Service: Attempting to save basket ID %s", newBasket.ID)
	err := s.repo.Save(ctx, &newBasket)
	if err != nil {
		log.Printf("Service: Error saving basket ID %s: %v", newBasket.ID, err)
		// Don't expose raw repository errors directly? Maybe wrap them.
		return nil, fmt.Errorf("failed to save basket: %w", err) // Wrap error
	}

	log.Printf("Service: Successfully saved basket ID %s", newBasket.ID)
	// 4. Return the created model
	return &newBasket, nil
}

// ListAllBaskets retrieves all baskets using the repository.
func (s *basketService) ListAllBaskets(ctx context.Context) ([]model.Basket, error) {
	log.Printf("Service: Attempting to find all baskets")
	baskets, err := s.repo.FindAll(ctx)
	if err != nil {
		log.Printf("Service: Error finding all baskets: %v", err)
		return nil, fmt.Errorf("failed to retrieve baskets: %w", err) // Wrap error
	}

	// The service could potentially filter or enrich data here if needed.

	// Ensure we return an empty slice, not nil, if no baskets found
	if baskets == nil {
		baskets = []model.Basket{}
	}

	log.Printf("Service: Found %d baskets", len(baskets))
	return baskets, nil
}

// GetBasketByID retrieves a single basket. (Example for later)
func (s *basketService) GetBasketByID(ctx context.Context, id uuid.UUID) (*model.Basket, error) {
	log.Printf("Service: Attempting to find basket by ID %s", id)
	basket, err := s.repo.FindByID(ctx, id)
	if err != nil {
		log.Printf("Service: Error finding basket ID %s: %v", id, err)
		if err == repository.ErrBasketNotFound { // Handle specific repo errors
			return nil, err // Or return a service-specific NotFoundError
		}
		return nil, fmt.Errorf("failed to retrieve basket %s: %w", id, err)
	}
	log.Printf("Service: Found basket ID %s", id)
	return basket, nil
}

// DeleteBasketByID handles the business logic for deleting a basket.
func (s *basketService) DeleteBasketByID(ctx context.Context, id uuid.UUID) error {
	log.Printf("Service: Attempting to delete basket by ID %s", id)

	// Optional: You might want to check if it exists first using FindByID,
	// but the repository DeleteByID check for RowsAffected handles the NotFound case.
	// Depending on requirements, you might add other business logic here
	// before deletion (e.g., checking if the basket is 'active').

	err := s.repo.DeleteByID(ctx, id) // Call the repository method
	if err != nil {
		log.Printf("Service: Error deleting basket ID %s: %v", id, err)
		// Passing up specific known errors like NotFound
		if errors.Is(err, repository.ErrBasketNotFound) {
			return err
		}
		// Wrap other errors
		return fmt.Errorf("failed to delete basket %s: %w", id, err)
	}

	log.Printf("Service: Successfully deleted basket ID %s", id)
	return nil // Success
}

// UpdateBasket handles the business logic for updating an existing basket.
func (s *basketService) UpdateBasket(ctx context.Context, id uuid.UUID, name string, stocks []model.Stock) (*model.Basket, error) {
	log.Printf("Service: Attempting to update basket ID %s", id)

	// 1. Input Validation
	if name == "" {
		return nil, fmt.Errorf("basket name cannot be empty")
	}
	// Allow empty stocks for PUT replace semantics (will delete all items)
	// if len(stocks) == 0 { return nil, fmt.Errorf("basket must contain at least one stock") }
	for i, stock := range stocks {
		if stock.Symbol == "" || stock.Quantity <= 0 {
			return nil, fmt.Errorf("invalid data for stock #%d: symbol and positive quantity required", i+1)
		}
	}

	// 2. Optional but recommended: Check if basket exists first using FindByID
	// This retrieves CreatedAt and confirms existence before complex update.
	existingBasket, err := s.repo.FindByID(ctx, id)
	if err != nil {
		log.Printf("Service: Basket %s not found for update: %v", id, err)
		// Pass up NotFound or wrap other errors
		return nil, err // Repo FindByID already wraps non-NotFound errors
	}

	// 3. Prepare the updated model object
	updatedBasket := model.Basket{
		ID:        id,                       // Use the ID from the path parameter
		Name:      name,                     // Use the new name
		Stocks:    stocks,                   // Use the new list of stocks
		CreatedAt: existingBasket.CreatedAt, // Preserve original creation time
		// UpdatedAt will be set by the database trigger via repo.Update
	}

	// 4. Call the repository to persist changes
	err = s.repo.Update(ctx, &updatedBasket)
	if err != nil {
		log.Printf("Service: Error updating basket ID %s in repository: %v", id, err)
		// Pass up specific known errors like NotFound (though caught above ideally)
		if errors.Is(err, repository.ErrBasketNotFound) {
			return nil, err
		}
		// Wrap other errors
		return nil, fmt.Errorf("failed to update basket %s: %w", id, err)
	}

	// 5. Return the updated basket representation
	// Note: updatedBasket.UpdatedAt here won't reflect the trigger's change yet.
	// If that's needed, uncomment the FindByID call below.
	log.Printf("Service: Successfully updated basket ID %s (returning intended state)", id)
	// Optionally, fetch again to get DB-generated UpdatedAt:
	// return s.repo.FindByID(ctx, id)
	return &updatedBasket, nil // Return the state we intended to save
}
