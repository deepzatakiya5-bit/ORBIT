package store

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"orbit/pkg/models"
)

var (
	ErrOnboardingIncomplete = errors.New("onboarding incomplete")
)

const userSelectColumns = `
	id, name, nickname, gender, birthdate, occupation, preferred_language,
	timezone, communication_style, why_here,
	onboarding_completed_at, created_at, updated_at
`

type OnboardingProfile struct {
	Name, Nickname, Gender, Occupation, PreferredLanguage string
	Timezone, CommunicationStyle                          string
	Birthdate                                             time.Time
	WhyHere                                               []string
}

func scanUserFromRow(row pgx.Row) (models.User, error) {
	var (
		id                                                                     string
		name, nickname, gender, occupation, preferredLanguage                  *string
		timezone, communicationStyle                                           *string
		birthdate, onboardingCompletedAt                                       *time.Time
		whyHere                                                                []string
		createdAt, updatedAt                                                   time.Time
	)

	err := row.Scan(
		&id, &name, &nickname, &gender, &birthdate, &occupation, &preferredLanguage,
		&timezone, &communicationStyle, &whyHere,
		&onboardingCompletedAt, &createdAt, &updatedAt,
	)
	if err != nil {
		return models.User{}, err
	}

	return scanUser(
		id, name, nickname, gender, occupation, preferredLanguage,
		timezone, communicationStyle,
		birthdate, onboardingCompletedAt, whyHere, createdAt, updatedAt,
	), nil
}

func (s *Store) CreateUser(ctx context.Context) (models.User, error) {
	row := s.db.QueryRow(ctx, `
		INSERT INTO users DEFAULT VALUES
		RETURNING `+userSelectColumns,
	)
	return scanUserFromRow(row)
}

func (s *Store) GetUser(ctx context.Context, userID uuid.UUID) (models.User, error) {
	row := s.db.QueryRow(ctx, `
		SELECT `+userSelectColumns+`
		FROM users
		WHERE id = $1
	`, userID)

	user, err := scanUserFromRow(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.User{}, ErrNotFound
	}
	return user, err
}

func (s *Store) CompleteOnboarding(ctx context.Context, userID uuid.UUID, p OnboardingProfile) (models.User, error) {
	nickname := p.Nickname
	if nickname == "" {
		nickname = p.Name
	}

	row := s.db.QueryRow(ctx, `
		UPDATE users
		SET
			name = $2,
			nickname = $3,
			gender = $4,
			birthdate = $5,
			occupation = $6,
			preferred_language = $7,
			timezone = $8,
			communication_style = $9,
			why_here = $10,
			onboarding_completed_at = now(),
			updated_at = now()
		WHERE id = $1
		RETURNING `+userSelectColumns,
		userID, p.Name, nickname, p.Gender, p.Birthdate, p.Occupation, p.PreferredLanguage,
		p.Timezone, p.CommunicationStyle, p.WhyHere,
	)

	user, err := scanUserFromRow(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.User{}, ErrNotFound
	}
	return user, err
}

func (s *Store) UserExists(ctx context.Context, userID uuid.UUID) (bool, error) {
	var exists bool
	err := s.db.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)
	`, userID).Scan(&exists)
	return exists, err
}

func RequireOnboarded(user models.User) error {
	if !user.OnboardingCompleted {
		return ErrOnboardingIncomplete
	}
	return nil
}
