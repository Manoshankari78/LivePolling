package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"live-polling-app/backend/models"
	"live-polling-app/backend/repositories"
	"live-polling-app/backend/utils"
)

type PollService struct {
	polls repositories.PollRepository
	votes repositories.VoteRepository
}

func NewPollService(polls repositories.PollRepository, votes repositories.VoteRepository) *PollService {
	return &PollService{polls: polls, votes: votes}
}

func buildOptions(values []string) []models.PollOption {
	options := make([]models.PollOption, len(values))
	for i, value := range values {
		options[i] = models.PollOption{ID: primitive.NewObjectID().Hex(), Text: strings.TrimSpace(value)}
	}
	return options
}

func (s *PollService) Create(ctx context.Context, creatorID primitive.ObjectID, question string, optionTexts []string, expiresAt *time.Time) (*models.Poll, error) {
	if err := utils.ValidatePoll(question, optionTexts); err != nil {
		return nil, err
	}
	if expiresAt != nil && !expiresAt.After(time.Now()) {
		return nil, errors.New("expiration must be in the future")
	}
	now := time.Now()
	poll := &models.Poll{ID: primitive.NewObjectID(), CreatorID: creatorID, Question: strings.TrimSpace(question), Options: buildOptions(optionTexts), Status: "active", CreatedAt: now, UpdatedAt: now, ExpiresAt: expiresAt}
	if err := s.polls.Create(ctx, poll); err != nil {
		return nil, err
	}
	return poll, nil
}

func (s *PollService) Get(ctx context.Context, id primitive.ObjectID) (*models.Poll, error) {
	poll, err := s.polls.FindByID(ctx, id)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if poll.Status == "active" && poll.ExpiresAt != nil && poll.ExpiresAt.Before(time.Now()) {
		poll.Status = "closed"
	}
	return poll, nil
}
func (s *PollService) List(ctx context.Context, creatorID primitive.ObjectID) ([]models.Poll, error) {
	return s.polls.ListByCreator(ctx, creatorID)
}
func (s *PollService) Update(ctx context.Context, id, creatorID primitive.ObjectID, question string, optionTexts []string, expiresAt *time.Time) (*models.Poll, error) {
	if err := utils.ValidatePoll(question, optionTexts); err != nil {
		return nil, err
	}
	if expiresAt != nil && !expiresAt.After(time.Now()) {
		return nil, errors.New("expiration must be in the future")
	}
	poll, err := s.polls.FindByID(ctx, id)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if poll.CreatorID != creatorID {
		return nil, ErrForbidden
	}
	if poll.Status != "active" {
		return nil, errors.New("closed polls cannot be edited")
	}
	if hasVotes(poll) {
		if len(optionTexts) != len(poll.Options) {
			return nil, errors.New("once voting starts, the number of options cannot change")
		}
		updatedOptions := make([]models.PollOption, len(poll.Options))
		for i, text := range optionTexts {
			updatedOptions[i] = poll.Options[i]
			updatedOptions[i].Text = strings.TrimSpace(text)
		}
		return s.polls.Update(ctx, id, creatorID, strings.TrimSpace(question), updatedOptions, expiresAt)
	}
	return s.polls.Update(ctx, id, creatorID, strings.TrimSpace(question), buildOptions(optionTexts), expiresAt)
}
func (s *PollService) Delete(ctx context.Context, id, creatorID primitive.ObjectID) error {
	poll, err := s.Get(ctx, id)
	if err != nil {
		return err
	}
	if poll.CreatorID != creatorID {
		return ErrForbidden
	}
	err = s.polls.Delete(ctx, id, creatorID)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return ErrNotFound
	}
	return err
}
func (s *PollService) Close(ctx context.Context, id, creatorID primitive.ObjectID) error {
	poll, err := s.Get(ctx, id)
	if err != nil {
		return err
	}
	if poll.CreatorID != creatorID {
		return ErrForbidden
	}
	if poll.Status != "active" {
		return ErrConflict
	}
	err = s.polls.Close(ctx, id, creatorID)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return ErrNotFound
	}
	return err
}
func (s *PollService) Vote(ctx context.Context, id primitive.ObjectID, optionID, voterID string) (*models.PollResults, error) {
	if len(voterID) < 16 || len(voterID) > 100 {
		return nil, errors.New("invalid voter id")
	}
	poll, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if poll.Status != "active" {
		return nil, ErrConflict
	}
	if poll.ExpiresAt != nil && poll.ExpiresAt.Before(time.Now()) {
		return nil, ErrConflict
	}
	validOption := false
	for _, option := range poll.Options {
		if option.ID == optionID {
			validOption = true
			break
		}
	}
	if !validOption {
		return nil, errors.New("invalid option")
	}
	vote := repositories.NewVote(id, optionID, voterID)
	if err := s.votes.Create(ctx, vote); err != nil {
		if repositories.IsDuplicateKey(err) {
			return nil, ErrConflict
		}
		return nil, err
	}
	if err := s.polls.IncrementOption(ctx, id, optionID); err != nil {
		return nil, err
	}
	updated, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	results := s.votes.Results(ctx, updated)
	return &results, nil
}
func (s *PollService) Results(ctx context.Context, id primitive.ObjectID) (*models.PollResults, error) {
	poll, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	results := s.votes.Results(ctx, poll)
	return &results, nil
}

func hasVotes(poll *models.Poll) bool {
	for _, option := range poll.Options {
		if option.Votes > 0 {
			return true
		}
	}
	return false
}
