package store

import (
	"time"

	"orbit/pkg/models"
)

func scanUser(
	id string,
	name, nickname, gender, occupation, preferredLanguage, timezone, communicationStyle *string,
	email, phone, deviceID *string,
	birthdate, onboardingCompletedAt *time.Time,
	whyHere []string,
	createdAt, updatedAt time.Time,
) models.User {
	if whyHere == nil {
		whyHere = []string{}
	}
	return models.User{
		ID:                  id,
		Name:                name,
		Nickname:            nickname,
		Gender:              gender,
		Birthdate:           birthdate,
		Occupation:          occupation,
		PreferredLanguage:   preferredLanguage,
		Timezone:            timezone,
		CommunicationStyle:  communicationStyle,
		WhyHere:             whyHere,
		Email:               email,
		Phone:               phone,
		DeviceID:            deviceID,
		OnboardingCompleted: onboardingCompletedAt != nil,
		CreatedAt:           createdAt,
		UpdatedAt:           updatedAt,
	}
}

func nullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
