package services

import (
	"fmt"
	"statio/config"
	"statio/internal/dto"
	"statio/internal/mappers"
	"statio/internal/models"
	"statio/internal/repositories"

	"strings"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	AppConfig *config.AppConfig
	userRepo  repositories.UserRepository
}

func NewUserService(
	AppConfig *config.AppConfig,
	userRepo repositories.UserRepository,
) *UserService {
	return &UserService{
		AppConfig: AppConfig,
		userRepo:  userRepo,
	}
}

func (s *UserService) GetUserByEmail(email string) (*models.User, error) {
	return s.userRepo.FindByEmail(email)
}

func (s *UserService) GetUserByUsername(username string) (*models.User, error) {
	return s.userRepo.FindByUsername(username)
}

func (s *UserService) GetUserByEmailOrUsername(identifier string) (*models.User, error) {
	return s.userRepo.FindByEmailOrUsername(identifier)
}

func (s *UserService) GetUserByInviteToken(inviteToken string) (*models.User, error) {
	return s.userRepo.FindByInviteToken(inviteToken)
}

func (s *UserService) GetUserByID(id string) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	resp := mappers.ToUserResponse(user)
	return resp, nil
}

func (s *UserService) GetAllUsers() ([]dto.UserResponse, error) {
	users, err := s.userRepo.FindAll()
	if err != nil {
		return nil, err
	}

	responses := make([]dto.UserResponse, 0, len(users))
	for _, user := range users {
		resp := mappers.ToUserResponse(&user)
		if resp != nil {
			responses = append(responses, *resp)
		}
	}

	return responses, nil
}

func (s *UserService) GetAllPaginated(
	search string,
	page, perPage int,
	sortBy, sortOrder string,
	filters map[string][]string,
) ([]dto.UserResponse, int64, error) {

	var total int64
	if err := s.userRepo.Count(search, filters, &total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	users, err := s.userRepo.FindPaginated(search, perPage, offset, sortBy, sortOrder, filters)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]dto.UserResponse, 0, len(users))
	for _, user := range users {
		resp := mappers.ToUserResponse(&user)
		if resp != nil {
			responses = append(responses, *resp)
		}
	}

	return responses, total, nil
}

func (s *UserService) CreateUser(req *dto.CreateUserRequest) error {
	var userExisting *models.User

	userExisting, _ = s.userRepo.FindByEmail(req.Email)
	if userExisting != nil {
		return fmt.Errorf("email is already taken")
	}

	userExisting, _ = s.userRepo.FindByUsername(req.Username)
	if userExisting != nil {
		return fmt.Errorf("username is already taken")
	}

	passwordHashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := mappers.ToUserModel(req)
	user.Password = string(passwordHashed)

	if err := s.userRepo.Create(user); err != nil {
		return err
	}

	return nil
}

func (s *UserService) UpdateUser(id string, req *dto.UpdateUserRequest) error {
	if req.Email != nil {
		userExisting, _ := s.userRepo.FindByEmail(*req.Email)
		if userExisting != nil && userExisting.ID != id {
			return fmt.Errorf("email is already in use by another user")
		}
	}

	if req.Username != nil {
		userExisting, _ := s.userRepo.FindByUsername(*req.Username)
		if userExisting != nil && userExisting.ID != id {
			return fmt.Errorf("username is already in use by another user")
		}
	}

	user, err := s.userRepo.FindByIDIncludePassword(id)
	if err != nil {
		return err
	}

	mappers.ApplyUserUpdateFromRequest(user, req)

	if req.Password != nil && strings.TrimSpace(*req.Password) != "" {
		passwordHashed, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		user.Password = string(passwordHashed)
	}

	if err := s.userRepo.Update(user); err != nil {
		return err
	}

	return nil
}

func (s *UserService) DeleteUser(id string) error {
	return s.userRepo.Delete(id)
}

func (s *UserService) GetUserInviteLink(id string) (*dto.UserInviteLinkResponse, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if user.InviteToken == nil {
		return nil, fmt.Errorf("user does not have an invite token")
	}

	inviteLink := fmt.Sprintf("%s/login?invite_token=%s", s.AppConfig.AppURL, *user.InviteToken)

	return &dto.UserInviteLinkResponse{
		InviteLink: inviteLink,
	}, nil
}
