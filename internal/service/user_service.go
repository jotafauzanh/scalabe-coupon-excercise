package service

import (
	"github.com/jotafauzanh/scalabe-coupon-excercise/internal/model"
	"github.com/jotafauzanh/scalabe-coupon-excercise/internal/repository"
)

type UserService struct {
	Repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{Repo: repo}
}

func (s *UserService) CreateUser(user *model.User) error {
	return s.Repo.Create(user)
}

func (s *UserService) GetAllUsers() ([]model.User, error) {
	return s.Repo.FindAll()
}

func (s *UserService) GetUserByID(id uint) (model.User, error) {
	return s.Repo.FindByID(id)
}
