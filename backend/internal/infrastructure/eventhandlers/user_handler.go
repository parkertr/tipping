package eventhandlers

import (
	"context"
	"encoding/json"

	"github.com/parkertr2/footy-tipping/internal/domain"
	"github.com/parkertr2/footy-tipping/internal/infrastructure/repository"
	"github.com/parkertr2/footy-tipping/pkg/events"
)

// UserEventHandler handles user-related events
type UserEventHandler struct {
	userRepo repository.UserRepository
}

// NewUserEventHandler creates a new user event handler
func NewUserEventHandler(userRepo repository.UserRepository) *UserEventHandler {
	return &UserEventHandler{
		userRepo: userRepo,
	}
}

// Handle processes user-related events
func (h *UserEventHandler) Handle(ctx context.Context, event *events.Event) error {
	switch event.Type {
	case "UserRegistered":
		return h.handleUserRegistered(ctx, event)
	case "UserProfileUpdated":
		return h.handleUserProfileUpdated(ctx, event)
	case "UserDeactivated":
		return h.handleUserDeactivated(ctx, event)
	default:
		return nil
	}
}

// handleUserRegistered processes UserRegistered events
func (h *UserEventHandler) handleUserRegistered(ctx context.Context, event *events.Event) error {
	var data events.UserRegistered
	if err := json.Unmarshal(event.Data.([]byte), &data); err != nil {
		return err
	}

	user := &domain.User{
		ID:        data.ID,
		GoogleID:  data.GoogleID,
		Email:     data.Email,
		Name:      data.Name,
		Picture:   data.Picture,
		CreatedAt: data.CreatedAt,
		UpdatedAt: data.CreatedAt,
		IsActive:  true,
		Stats: domain.UserStats{
			TotalPoints:        0,
			CorrectPredictions: 0,
			TotalPredictions:   0,
			CurrentRank:        0,
		},
	}

	return h.userRepo.Create(ctx, user)
}

// handleUserProfileUpdated processes UserProfileUpdated events
func (h *UserEventHandler) handleUserProfileUpdated(ctx context.Context, event *events.Event) error {
	var data events.UserProfileUpdated
	if err := json.Unmarshal(event.Data.([]byte), &data); err != nil {
		return err
	}

	user, err := h.userRepo.GetByID(ctx, data.UserID)
	if err != nil {
		return err
	}
	if user == nil {
		return nil // User not found, ignore event
	}

	user.Name = data.Name
	user.Picture = data.Picture
	user.UpdatedAt = data.UpdatedAt

	return h.userRepo.Update(ctx, user)
}

// handleUserDeactivated processes UserDeactivated events
func (h *UserEventHandler) handleUserDeactivated(ctx context.Context, event *events.Event) error {
	var data events.UserDeactivated
	if err := json.Unmarshal(event.Data.([]byte), &data); err != nil {
		return err
	}

	user, err := h.userRepo.GetByID(ctx, data.UserID)
	if err != nil {
		return err
	}
	if user == nil {
		return nil // User not found, ignore event
	}

	user.IsActive = false
	user.UpdatedAt = data.UpdatedAt

	return h.userRepo.Update(ctx, user)
}
