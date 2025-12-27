package services

import (
	"errors"
	"fmt"
	"go-microservice/models"
	"sync"
)

type UserService struct {
	users map[int]*models.User
	mu    sync.RWMutex
}

func NewUserService() *UserService {
	return &UserService{
		users: make(map[int]*models.User),
	}
}

func (s *UserService) GetAll() []*models.User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	users := make([]*models.User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}

	return users
}

func (s *UserService) GetById(id int) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	user, exists := s.users[id]
	if !exists {
		return nil, errors.New("User not found")
	}

	return user, nil
}

func (s *UserService) Create(user models.User) (*models.User, error) {
	if err := user.Validate(); err != nil {
		return nil, err
	}
	user.ID = models.GenerateID()
	s.mu.Lock()
	defer s.mu.Unlock()
	userCopy := user
	s.users[user.ID] = &userCopy

	return &userCopy, nil
}

func (s *UserService) Update(id int, user models.User) (*models.User, error) {
	if err := user.Validate(); err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	_, exists := s.users[id]
	if !exists {
		return nil, fmt.Errorf("User with id=%d not found", id)
	}
	user.ID = id
	userCopy := user
	s.users[id] = &userCopy

	return &userCopy, nil
}

func (s *UserService) Delete(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, exists := s.users[id]
	if !exists {
		return fmt.Errorf("User with id=%d not found", id)
	}
	delete(s.users, id)
	return nil
}
