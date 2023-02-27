package main_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/alextanhongpin/go-unit-of-work/app"
	"github.com/alextanhongpin/go-unit-of-work/internal"
	"github.com/alextanhongpin/uow"
	_ "github.com/lib/pq"
)

var ErrRollback = errors.New("test: rollback")

var db *sql.DB

func TestMain(m *testing.M) {
	// Start the database.
	var kill func()
	db, kill = internal.NewPostgresTestDB()

	// Run migration.
	repo := app.NewPostgresUserRepository(uow.New(db))
	err := repo.Migrate(context.Background())
	if err != nil {
		log.Fatalf("failed to run migration: %v", err)
	}

	code := m.Run()

	// This cannot be deferred because os.Exit does not care
	// about defer.
	kill()

	os.Exit(code)
}

func TestUserUseCase(t *testing.T) {
	// Arrange.
	u := uow.New(db)
	repo := app.NewPostgresUserRepository(u)
	uc := app.NewUserUseCase(u, repo)

	// In the happy path, create user succeeds.
	t.Run("happy path", func(t *testing.T) {
		email := "john.appleseed@gmail.com"
		deviceID := "ios-123"

		// Act.
		// Wrap the operation in a transaction and rollback after the test complete.
		err := u.RunInTx(context.Background(), func(ctx context.Context) error {
			// Even though Register method calls `RunInTx`, it will not create a new
			// transaction.
			if err := uc.Register(ctx, email, deviceID); err != nil {
				t.Errorf("failed to register user: %v", err)
				return err
			}

			// Assert.
			// Both user and user device should be created successfully.
			userExists, err := checkUserExists(ctx, email)
			if err != nil {
				return err
			}

			if !userExists {
				t.Error("expected user to be created, but failed")
				return ErrRollback
			}

			emailExists, err := checkUserDeviceExists(ctx, deviceID)
			if err != nil {
				return err
			}

			if !emailExists {
				t.Error("expected user device to be created, but failed")
				return ErrRollback
			}

			// Rollback the operation, instead of truncating the table for every tests.
			return ErrRollback
		})

		// The returned error must be ErrRollback.
		if err != nil && !errors.Is(err, ErrRollback) {
			t.Fatalf("failed to rollback: %v", err)
		}
	})
}

func checkUserExists(ctx context.Context, email string) (bool, error) {
	// Get the *UnitOfWork instance from context.
	u := uow.MustValue(ctx)

	// Get the *sql.Tx from the context.
	tx := u.DB(ctx)

	var exists bool
	err := tx.
		QueryRowContext(ctx, `select exists (select 1 from users where email = $1)`, email).
		Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user exists: %w", err)
	}

	return true, nil
}

func checkUserDeviceExists(ctx context.Context, deviceID string) (bool, error) {
	// Get the *UnitOfWork instance from context.
	u := uow.MustValue(ctx)

	// Get the *sql.Tx from the context.
	tx := u.DB(ctx)

	var exists bool
	err := tx.
		QueryRowContext(ctx, `select exists (select 1 from user_devices where device_id = $1)`, deviceID).
		Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user devices exists: %w", err)
	}

	return true, nil
}
