package app

import (
	"context"

	"github.com/alextanhongpin/uow"
	"github.com/google/uuid"
)

type userRepository interface {
	CreateUser(ctx context.Context, email string) (*User, error)
	CreateUserDevice(ctx context.Context, userID uuid.UUID, deviceID string) (*UserDevice, error)
}

type UserUseCase struct {
	uow  uow.UOW
	repo userRepository
}

func NewUserUseCase(u uow.UOW, repo userRepository) *UserUseCase {
	return &UserUseCase{
		uow:  u,
		repo: repo,
	}
}

func (uc *UserUseCase) Register(ctx context.Context, email, deviceID string) error {
	return uc.uow.RunInTx(ctx, func(ctx context.Context) error {
		user, err := uc.repo.CreateUser(ctx, email)
		if err != nil {
			return err
		}

		// NOTE: Uncomment this to fail this transaction.
		//if true {
		//return errors.New("intentional error")
		//}

		_, err = uc.repo.CreateUserDevice(ctx, user.ID, deviceID)
		return err
	})
}
