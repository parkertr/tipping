package eventhandlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/parkertr/tipping/internal/domain"
	"github.com/parkertr/tipping/internal/infrastructure/repository"
	"github.com/parkertr/tipping/pkg/events"
)

var ErrInvalidEventDataType = errors.New("invalid event data type: expected []byte")

// UserEventHandler handles user-related events.
type UserEventHandler struct {
	userRepo repository.UserRepository
}

// NewUserEventHandler creates a new user event handler.
func NewUserEventHandler(userRepo repository.UserRepository) *UserEventHandler {
	return &UserEventHandler{
		userRepo: userRepo,
	}
}

// Handle processes user-related events.
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

// handleUserRegistered processes UserRegistered events.
func (h *UserEventHandler) handleUserRegistered(ctx context.Context, event *events.Event) error {
	var data events.UserRegistered

	eventData, ok := event.Data.([]byte)
	if !ok {
		return ErrInvalidEventDataType
	}

	if err := json.Unmarshal(eventData, &data); err != nil {
		return fmt.Errorf("failed to unmarshal event data: %w", err)
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

	if err := h.userRepo.Create(ctx, user); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// handleUserProfileUpdated processes UserProfileUpdated events.
func (h *UserEventHandler) handleUserProfileUpdated(ctx context.Context, event *events.Event) error {
	var data events.UserProfileUpdated

	eventData, ok := event.Data.([]byte)
	if !ok {
		return ErrInvalidEventDataType
	}

	if err := json.Unmarshal(eventData, &data); err != nil {
		return fmt.Errorf("failed to unmarshal event data: %w", err)
	}

	user, err := h.userRepo.GetByID(ctx, data.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return nil // User not found, ignore event
	}

	user.Name = data.Name
	user.Picture = data.Picture
	user.UpdatedAt = data.UpdatedAt

	if err := h.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// handleUserDeactivated processes UserDeactivated events.
func (h *UserEventHandler) handleUserDeactivated(ctx context.Context, event *events.Event) error {
	var data events.UserDeactivated

	eventData, ok := event.Data.([]byte)
	if !ok {
		return ErrInvalidEventDataType
	}

	if err := json.Unmarshal(eventData, &data); err != nil {
		return fmt.Errorf("failed to unmarshal event data: %w", err)
	}

	user, err := h.userRepo.GetByID(ctx, data.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return nil // User not found, ignore event
	}

	user.IsActive = false
	user.UpdatedAt = data.UpdatedAt

	if err := h.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}
