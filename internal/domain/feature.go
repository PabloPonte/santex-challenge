package domain

import "time"

type Status string

const (
	StatusOpen        Status = "open"
	StatusClosed      Status = "closed"
	StatusWhitelisted Status = "whitelisted"
)

func (s Status) Valid() bool {
	return s == StatusOpen || s == StatusClosed || s == StatusWhitelisted
}

type Feature struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      Status    `json:"status"`
	StatusDate  time.Time `json:"statusDate"`
	Whitelist   []string  `json:"whitelist"`
}

type Evaluation struct {
	FeatureName string `json:"featureName"`
	UserID      string `json:"userId"`
	Enabled     bool   `json:"enabled"`
	Status      Status `json:"status"`
}
