package store

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"orbit/pkg/models"
)

var (
	ErrOnboardingIncomplete = errors.New("onboarding incomplete")
	ErrDuplicateLoginID     = errors.New("duplicate login identifier")
)

const userSelectColumns = `
	id, name, nickname, gender, birthdate, occupation, preferred_language,
	timezone, communication_style, why_here,
	email, phone, device_id,
	onboarding_completed_at, created_at, updated_at
`

type OnboardingProfile struct {
	Name, Nickname, Gender, Occupation, PreferredLanguage string
	Timezone, CommunicationStyle                          string
	Email, Phone, DeviceID                                string
	Birthdate                                             time.Time
	WhyHere                                               []string
}

func scanUserFromRow(row pgx.Row) (models.User, error) {
	var (
		id                                                    string
		name, nickname, gender, occupation, preferredLanguage *string
		timezone, communicationStyle                          *string
		email, phone, deviceID                                *string
		birthdate, onboardingCompletedAt                      *time.Time
		whyHere                                               []string
		createdAt, updatedAt                                  time.Time
	)

	err := row.Scan(
		&id, &name, &nickname, &gender, &birthdate, &occupation, &preferredLanguage,
		&timezone, &communicationStyle, &whyHere,
		&email, &phone, &deviceID,
		&onboardingCompletedAt, &createdAt, &updatedAt,
	)
	if err != nil {
		return models.User{}, err
	}

	return scanUser(
		id, name, nickname, gender, occupation, preferredLanguage,
		timezone, communicationStyle,
		email, phone, deviceID,
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
			email = $11,
			phone = $12,
			device_id = $13,
			onboarding_completed_at = now(),
			updated_at = now()
		WHERE id = $1
		RETURNING `+userSelectColumns,
		userID, p.Name, nickname, p.Gender, p.Birthdate, p.Occupation, p.PreferredLanguage,
		p.Timezone, p.CommunicationStyle, p.WhyHere,
		nullIfEmpty(p.Email), nullIfEmpty(p.Phone), nullIfEmpty(p.DeviceID),
	)

	user, err := scanUserFromRow(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.User{}, ErrNotFound
	}
	if isUniqueViolation(err) {
		return models.User{}, ErrDuplicateLoginID
	}
	return user, err
}

func (s *Store) FindUserByLogin(ctx context.Context, email, phone, deviceID string) (models.User, error) {
	row := s.db.QueryRow(ctx, `
		SELECT `+userSelectColumns+`
		FROM users
		WHERE
			($1 <> '' AND email = $1)
			OR ($2 <> '' AND phone = $2)
			OR ($3 <> '' AND device_id = $3)
		LIMIT 1
	`, email, phone, deviceID)

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

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
