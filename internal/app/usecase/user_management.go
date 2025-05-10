package usecase

import (
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"spotify_recommender/internal/app/dto"
	"spotify_recommender/internal/domain/entity"
	"spotify_recommender/internal/domain/repository"
	"time"
)

type UserManagementUseCase struct {
	userRepository repository.UserRepository
}

func NewUserManagementUseCase(
	userRepository repository.UserRepository,
) *UserManagementUseCase {
	return &UserManagementUseCase{
		userRepository: userRepository,
	}
}

func (uc *UserManagementUseCase) Register(ctx context.Context, createDTO dto.CreateUserDTO) (*dto.UserDTO, error) {
	existingUser, err := uc.userRepository.GetByEmail(ctx, createDTO.Email)
	if err == nil && existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(createDTO.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := entity.NewUser(createDTO.Email, string(hashedPassword), createDTO.Name)

	// Устанавливаем предпочтения, если они предоставлены
	if createDTO.Preferences.FavoriteGenres != nil {
		user.Preferences.FavoriteGenres = createDTO.Preferences.FavoriteGenres
	}
	if createDTO.Preferences.DislikedGenres != nil {
		user.Preferences.DislikedGenres = createDTO.Preferences.DislikedGenres
	}
	if createDTO.Preferences.MinTempo > 0 {
		user.Preferences.MinTempo = createDTO.Preferences.MinTempo
	}
	if createDTO.Preferences.MaxTempo > 0 {
		user.Preferences.MaxTempo = createDTO.Preferences.MaxTempo
	}
	if createDTO.Preferences.PreferredMoods != nil {
		user.Preferences.PreferredMoods = createDTO.Preferences.PreferredMoods
	}

	err = uc.userRepository.Save(ctx, user)
	if err != nil {
		return nil, errors.New("Error with saving user")
	}

	userDTO := dto.UserFromEntity(user)

	return &userDTO, err
}

func (uc *UserManagementUseCase) Login(ctx context.Context, email string,
	password string) (*dto.UserDTO, error) {
	existingUser, err := uc.userRepository.GetByEmail(ctx, email)
	if err == nil && existingUser != nil {
		return nil, errors.New("no such an email")
	}
	err = bcrypt.CompareHashAndPassword([]byte(existingUser.PasswordHash), []byte(password))
	if err != nil {
		return nil, errors.New("wrong password")
	}

	existingUser.LastLoginAt = time.Now()
	err = uc.userRepository.Update(ctx, existingUser)
	if err != nil {
		return nil, err
	}
	userDTO := dto.UserFromEntity(existingUser)

	return &userDTO, nil
}

func (uc *UserManagementUseCase) GetUserProfile(
	ctx context.Context,
	userID string,
) (*dto.UserDTO, error) {
	user, err := uc.userRepository.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	userDTO := dto.UserFromEntity(user)

	return &userDTO, nil
}

func (uc *UserManagementUseCase) UpdatePreferences(
	ctx context.Context,
	userID string,
	preferencesDTO dto.UpdatePreferencesDTO,
) (*dto.UserDTO, error) {
	user, err := uc.userRepository.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	preferences := entity.Preferences{
		FavoriteGenres: preferencesDTO.FavoriteGenres,
		DislikedGenres: preferencesDTO.DislikedGenres,
		MinTempo:       preferencesDTO.MinTempo,
		MaxTempo:       preferencesDTO.MaxTempo,
		PreferredMoods: preferencesDTO.PreferredMoods,
	}

	err = uc.userRepository.UpdatePreferences(ctx, userID, preferences)
	if err != nil {
		return nil, err
	}

	user.Preferences = preferences

	userDTO := dto.UserFromEntity(user)

	return &userDTO, nil
}
