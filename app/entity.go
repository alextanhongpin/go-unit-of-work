package app

import "github.com/google/uuid"

type User struct {
	ID    uuid.UUID
	Email string
}

type UserDevice struct {
	ID       uuid.UUID
	UserID   uuid.UUID
	DeviceID string
}
