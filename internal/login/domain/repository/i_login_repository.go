package repository

import "hub-user-service/internal/login/domain/model"

type ILoginRepository interface {
	GetUserByEmail(email string) (*model.User, error)
}
