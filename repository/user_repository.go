package repository

import (
	"errors"
	"fmt"
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

	// [FIX] Isolate the validation snippet in a separate file
	var userConvertido int
	switch v := userID.(type) {
	case float64:
		userConvertido = int(v)
	case int:
		userConvertido = v
	}
	err := database.DB.Table("users").Where("id = ?", userConvertido).Where("is_active = ?", 1).Find(&user).Error
	fmt.Println(err)
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
			fmt.Println("Erro ao buscar usuário por email:", err)
			return err
		}

		if existingUser.ID != nil && *user.ID != *existingUser.ID {
			fmt.Println("Email já está em uso por outro usuário")
			return errors.New("Email already in use")
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
