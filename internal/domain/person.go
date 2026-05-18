package domain

import "time"

type Person struct {
	ID          string
	DisplayName string
	BirthDate   *time.Time
	DeathDate   *time.Time
	Notes       string
}
