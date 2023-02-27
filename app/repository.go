package app

import (
	"context"
	"fmt"

	_ "embed"

	"github.com/alextanhongpin/uow"
	"github.com/google/uuid"
)

// Database queries.
var (
	//go:embed queries/create_users_table.sql
	createUserTableStmt string

	//go:embed queries/create_user_devices_table.sql
	createUserDeviceTableStmt string

	//go:embed queries/create_user.sql
	createUserStmt string

	//go:embed queries/create_user_device.sql
	createUserDeviceStmt string
)

type PostgresUserRepository struct {
	uow uow.UOW
}

func NewPostgresUserRepository(u uow.UOW) *PostgresUserRepository {
	return &PostgresUserRepository{
		uow: u,
	}
}

func (p *PostgresUserRepository) Migrate(ctx context.Context) error {
	return p.uow.RunInTx(ctx, func(ctx context.Context) error {
		tx := p.uow.DB(ctx)
		_, err := tx.ExecContext(ctx, createUserTableStmt)
		if err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx, createUserDeviceTableStmt)
		if err != nil {
			return err
		}

		return nil
	})
}

func (p *PostgresUserRepository) CreateUser(ctx context.Context, email string) (*User, error) {
	db := p.uow.DB(ctx)

	var u User
	err := db.
		QueryRowContext(ctx, createUserStmt, email).
		Scan(&u.ID, &u.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &u, nil
}

func (p *PostgresUserRepository) CreateUserDevice(ctx context.Context, userID uuid.UUID, deviceID string) (*UserDevice, error) {
	db := p.uow.DB(ctx)

	var u UserDevice
	err := db.
		QueryRowContext(ctx, createUserDeviceStmt, userID, deviceID).
		Scan(&u.ID, &u.UserID, &u.DeviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to create user device: %w", err)
	}

	return &u, nil
}
