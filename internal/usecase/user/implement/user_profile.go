package implement

import (
	"context"
	"errors"

	"github.com/ducklawrence05/go-test-backend-api/config"
	"github.com/ducklawrence05/go-test-backend-api/internal/constants/errorcode"
	"github.com/ducklawrence05/go-test-backend-api/internal/entities"
	"github.com/ducklawrence05/go-test-backend-api/internal/usecase/repository"
	"github.com/ducklawrence05/go-test-backend-api/internal/usecase/uow"
	"github.com/ducklawrence05/go-test-backend-api/internal/usecase/user"
	"github.com/ducklawrence05/go-test-backend-api/pkg/utils/password"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

// implement
type userProfileManager struct {
	config   *config.Config
	uow      uow.UserManagerUow
	userRepo repository.UserRepository
}

func NewUserProfileManager(
	config *config.Config,
	uow uow.UserManagerUow,
	userRepo repository.UserRepository,
) user.UserProfileManager {
	return &userProfileManager{
		config:   config,
		uow:      uow,
		userRepo: userRepo,
	}
}

func (m *userProfileManager) GetMe(ctx context.Context, userID uuid.UUID) (*entities.User, error) {
	user, err := m.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorcode.ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

func (m *userProfileManager) UpdateMe(ctx context.Context, dto user.UpdateMeDto) (*entities.User, error) {
	// get user
	user, err := m.userRepo.GetByID(ctx, dto.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorcode.ErrUserNotFound
		}
		return nil, err
	}

	// check username unique
	if dto.UserName != "" {
		taken, err := m.userRepo.IsUserNameTaken(ctx, dto.UserName, dto.UserID)
		if err != nil {
			return nil, err
		}
		if taken {
			return nil, errorcode.ErrInvalidUserName
		}
		user.UserName = dto.UserName
	}

	// other fields
	if dto.FirstName != "" {
		user.FirstName = dto.FirstName
	}
	if dto.LastName != "" {
		user.LastName = dto.LastName
	}

	// update
	if err := m.userRepo.Update(ctx, user, map[string]any{
		"user_name":  user.UserName,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
	}); err != nil {
		return nil, err
	}

	return user, nil
}

func (m *userProfileManager) ChangePassword(ctx context.Context, dto user.ChangePasswordDto) error {
	// get user
	user, err := m.userRepo.GetByID(ctx, dto.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errorcode.ErrUserNotFound
		}
		return err
	}

	g, _ := errgroup.WithContext(ctx)

	hpChan := make(chan string, 1)

	// check old password
	g.Go(func() error {
		if !password.ComparePasswords(user.Password, []byte(dto.OldPassword)) {
			return errorcode.ErrInvalidPassword
		}
		return nil
	})

	// hash password
	g.Go(func() error {
		hp, err := password.HashPassword(dto.NewPassword)
		if err != nil {
			return err
		}
		hpChan <- hp
		return nil
	})

	if err := g.Wait(); err != nil {
		return err
	}

	// change password
	if err := m.userRepo.Update(ctx, user, map[string]any{
		"password": <-hpChan,
	}); err != nil {
		return err
	}

	return nil
}

func (m *userProfileManager) DeleteMe(ctx context.Context, userID uuid.UUID) error {
	return m.uow.Do(ctx, func(r uow.UserManagerRepoProvider) error {
		// check user exists
		if _, err := r.UserRepository().GetByID(ctx, userID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errorcode.ErrUserNotFound
			}
			return err
		}

		// hard delete all rt
		if err := r.RefreshTokenRepository().DeleteByUserID(ctx, userID); err != nil {
			return err
		}

		// soft delete user
		if err := r.UserRepository().DeleteByID(ctx, userID); err != nil {
			return err
		}

		return nil
	})
}
