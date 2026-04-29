package service

import (
	"context"
	"errors"
	"fmt"
	"pfo-vector/internal/repository"
)

type UserEventService struct {
	queries repository.Querier
}

func NewUserEventService(q *repository.Queries) *UserEventService {
	return &UserEventService{queries: q}
}

var (
	ErrEventFull          = errors.New("event full")
	ErrAlreadyRegistered  = errors.New("already registered")
	ErrRegistrationClosed = errors.New("registred closed")
	ErrUnknown            = errors.New("unknown error")
	ErrNotFound           = errors.New("event not found error")
	
)

func (s *UserEventService) RegisterUserToEvent(ctx context.Context, userId, EventId int32) error {

	args := repository.RegisterUserWithStatusParams{
		UserID:  userId,
		EventID: EventId,
	}

	status, err := s.queries.RegisterUserWithStatus(ctx, args)
	if err != nil {
		return fmt.Errorf("failed to register user: %w", err)
	}

	// DEBUG: логирование статуса
	fmt.Printf("DEBUG: RegisterUserWithStatus returned status='%s'\n", status)

	switch status {
	case "SUCCESS":
		return nil
	case "NO_VACANCY":
		return ErrEventFull
	case "NOT_FOUND":
		return ErrNotFound
	case "ALREADY_REGISTERED":
		return ErrAlreadyRegistered
	case "REGISTRATION_CLOSED":
		return ErrRegistrationClosed
	default:
		return fmt.Errorf("unknown status from database: '%s'", status)
	}
}
