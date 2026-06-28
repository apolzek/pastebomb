package repository

import (
	"errors"
	"gin-goinc-api/database"
	"gin-goinc-api/model"
	"gin-goinc-api/requests"
	"gin-goinc-api/responses"

	"gorm.io/gorm"
)

type UserRepository struct{}

func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

func (repo *UserRepository) GetAllUsers() ([]responses.UserResponse, error) {
	var userResponse []responses.UserResponse
	err := database.DB.Table("users").Find(&userResponse).Error
	return userResponse, err
}

func (repo *UserRepository) GetAllActiveUsers() ([]responses.UserResponse, error) {
	var userResponse []responses.UserResponse
	err := database.DB.Table("users").Where("is_active = ?", 1).Find(&userResponse).Error
	return userResponse, err
}

func (repo *UserRepository) GetAllActiveUsersPaginate(pageInt, perPageInt int) ([]responses.UserResponse, error) {
	var userResponse []responses.UserResponse
	err := database.DB.Table("users").Where("is_active = ?", 1).Offset((pageInt - 1) * perPageInt).Limit(perPageInt).Find(&userResponse).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return userResponse, err
}

func (repo *UserRepository) GetUserByIDorUsername(id any) (*responses.UserResponse, error) {
	user := new(responses.UserResponse)
	err := database.DB.Table("users").
		Where("id = ? OR username = ?", id, id).
		Where("is_active = ?", 1).
		Find(&user).
		Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return user, err
}

func (repo *UserRepository) GetUserByEmail(email string) (*model.User, error) {
	user := new(model.User)
	err := database.DB.Table("users").Where("email = ?", email).Find(&user).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return user, err
}

func (repo *UserRepository) GetUserData(userID interface{}) (*responses.UserResponse, error) {
	user := new(responses.UserResponse)

	var userConvertido int64
	switch v := userID.(type) {
	case float64:
		userConvertido = int64(v)
	case int:
		userConvertido = int64(v)
	case int64:
		userConvertido = v
	}
	err := database.DB.Table("users").Where("id = ?", userConvertido).Where("is_active = ?", 1).Find(&user).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return user, err
}

func (repo *UserRepository) UpdateUserData(userID int, userRequest *requests.UserRequest) error {
	user := new(model.User)

	err := database.DB.Table("users").Where("id = ?", userID).Find(&user).Error
	if err != nil {
		return err
	}

	if user.ID == nil {
		return gorm.ErrRecordNotFound
	}
	if user.Email != nil && *user.Email != userRequest.Email {

		existingUser, err := repo.GetUserByEmail(userRequest.Email)
		if err != nil {
			return err
		}

		if existingUser.ID != nil && *user.ID != *existingUser.ID {
			return errors.New("email already in use")
		}
	}

	user.Name = &userRequest.Name
	user.Username = &userRequest.Username
	user.Email = &userRequest.Email
	user.BornDate = &userRequest.BornDate
	return database.DB.Table("users").Where("id = ?", userID).Updates(&user).Error
}

func (repo *UserRepository) DeactivateUserByID(id int) error {
	return database.DB.Table("users").Where("id = ?", id).Update("is_active", 0).Error
}

func (r *UserRepository) CheckEmailExists(email string) bool {
	user := new(model.User)
	database.DB.Table("users").Where("email = ?", email).First(user)
	return user.Email != nil
}

func (r *UserRepository) CheckUsernameExists(username string) bool {
	user := new(model.User)
	database.DB.Table("users").Where("username = ?", username).First(user)
	return user.Email != nil
}

func (r *UserRepository) CreateUser(user *model.User) error {
	return database.DB.Table("users").Create(user).Error
}

func (r *UserRepository) CheckUserEmailExists(email string, userID int) bool {
	user := new(model.User)
	result := database.DB.Table("users").Where("email = ?", email).Find(user)
	return result.Error == nil && user.Email != nil && *user.ID != userID
}

func (r *UserRepository) UpdateUser(id int, user *model.User) error {
	return database.DB.Table("users").Where("id = ?", id).Updates(user).Error
}
