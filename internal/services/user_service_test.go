package services

import (
	"testing"
	"time"

	"github.com/alfafaa/alfafaa-blog/internal/dto"
	"github.com/alfafaa/alfafaa-blog/internal/models"
	"github.com/alfafaa/alfafaa-blog/internal/utils"
	"github.com/alfafaa/alfafaa-blog/tests/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type UserServiceTestSuite struct {
	suite.Suite
	userRepo    *mocks.MockUserRepository
	articleRepo *mocks.MockArticleRepository
	service     UserService
}

func (suite *UserServiceTestSuite) SetupTest() {
	suite.userRepo = new(mocks.MockUserRepository)
	suite.articleRepo = new(mocks.MockArticleRepository)
	suite.service = NewUserService(suite.userRepo, suite.articleRepo)
}

func TestUserServiceTestSuite(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}

// GetUser Tests

func (suite *UserServiceTestSuite) TestGetUser_Success() {
	userID := uuid.New()
	user := &models.User{
		ID:         userID,
		Username:   "testuser",
		Email:      "test@example.com",
		FirstName:  "Test",
		LastName:   "User",
		Role:       models.RoleAuthor,
		IsActive:   true,
		IsVerified: true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	suite.userRepo.On("FindByID", userID).Return(user, nil)
	suite.articleRepo.On("FindByAuthor", userID, mock.AnythingOfType("repositories.ArticleFilters")).
		Return([]models.Article{}, int64(5), nil)

	result, err := suite.service.GetUser(userID.String())

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), user.Email, result.Email)
	assert.Equal(suite.T(), user.Username, result.Username)
	assert.Equal(suite.T(), 5, result.ArticleCount)
	suite.userRepo.AssertExpectations(suite.T())
	suite.articleRepo.AssertExpectations(suite.T())
}

func (suite *UserServiceTestSuite) TestGetUser_InvalidUUID() {
	result, err := suite.service.GetUser("not-a-valid-uuid")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), utils.ErrBadRequest, err)
}

func (suite *UserServiceTestSuite) TestGetUser_NotFound() {
	userID := uuid.New()

	suite.userRepo.On("FindByID", userID).Return(nil, gorm.ErrRecordNotFound)

	result, err := suite.service.GetUser(userID.String())

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), utils.ErrNotFound, err)
	suite.userRepo.AssertExpectations(suite.T())
}

// GetUsers Tests

func (suite *UserServiceTestSuite) TestGetUsers_Success() {
	users := []models.User{
		{
			ID:       uuid.New(),
			Username: "user1",
			Email:    "user1@example.com",
			Role:     models.RoleReader,
			IsActive: true,
		},
		{
			ID:       uuid.New(),
			Username: "user2",
			Email:    "user2@example.com",
			Role:     models.RoleAuthor,
			IsActive: true,
		},
	}

	query := &dto.UserListQuery{
		PaginationQuery: dto.PaginationQuery{Page: 1, PerPage: 10},
	}

	suite.userRepo.On("FindAll", mock.AnythingOfType("repositories.UserFilters")).
		Return(users, int64(2), nil)

	result, total, err := suite.service.GetUsers(query)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 2)
	assert.Equal(suite.T(), int64(2), total)
	suite.userRepo.AssertExpectations(suite.T())
}

func (suite *UserServiceTestSuite) TestGetUsers_WithFilters() {
	users := []models.User{
		{
			ID:       uuid.New(),
			Username: "author1",
			Email:    "author1@example.com",
			Role:     models.RoleAuthor,
			IsActive: true,
		},
	}

	isActive := true
	query := &dto.UserListQuery{
		PaginationQuery: dto.PaginationQuery{Page: 1, PerPage: 10},
		Role:            "author",
		IsActive:        &isActive,
		Search:          "author",
	}

	suite.userRepo.On("FindAll", mock.AnythingOfType("repositories.UserFilters")).
		Return(users, int64(1), nil)

	result, total, err := suite.service.GetUsers(query)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 1)
	assert.Equal(suite.T(), int64(1), total)
	suite.userRepo.AssertExpectations(suite.T())
}

func (suite *UserServiceTestSuite) TestGetUsers_Empty() {
	query := &dto.UserListQuery{
		PaginationQuery: dto.PaginationQuery{Page: 1, PerPage: 10},
	}

	suite.userRepo.On("FindAll", mock.AnythingOfType("repositories.UserFilters")).
		Return([]models.User{}, int64(0), nil)

	result, total, err := suite.service.GetUsers(query)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 0)
	assert.Equal(suite.T(), int64(0), total)
	suite.userRepo.AssertExpectations(suite.T())
}

// UpdateUser Tests

func (suite *UserServiceTestSuite) TestUpdateUser_Success_OwnProfile() {
	userID := uuid.New()
	user := &models.User{
		ID:        userID,
		Username:  "testuser",
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		Role:      models.RoleReader,
		IsActive:  true,
	}

	firstName := "Updated"
	lastName := "Name"
	req := &dto.UpdateUserRequest{
		FirstName: &firstName,
		LastName:  &lastName,
	}

	suite.userRepo.On("FindByID", userID).Return(user, nil)
	suite.userRepo.On("Update", mock.AnythingOfType("*models.User")).Return(nil)

	result, err := suite.service.UpdateUser(userID.String(), req, userID.String(), false)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), firstName, result.FirstName)
	assert.Equal(suite.T(), lastName, result.LastName)
	suite.userRepo.AssertExpectations(suite.T())
}

func (suite *UserServiceTestSuite) TestUpdateUser_Success_AsAdmin() {
	userID := uuid.New()
	adminID := uuid.New()
	user := &models.User{
		ID:        userID,
		Username:  "testuser",
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		Role:      models.RoleReader,
		IsActive:  true,
	}

	bio := "Updated bio"
	req := &dto.UpdateUserRequest{
		Bio: &bio,
	}

	suite.userRepo.On("FindByID", userID).Return(user, nil)
	suite.userRepo.On("Update", mock.AnythingOfType("*models.User")).Return(nil)

	result, err := suite.service.UpdateUser(userID.String(), req, adminID.String(), true)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), bio, result.Bio)
	suite.userRepo.AssertExpectations(suite.T())
}

func (suite *UserServiceTestSuite) TestUpdateUser_Forbidden_NotOwner() {
	userID := uuid.New()
	otherUserID := uuid.New()

	req := &dto.UpdateUserRequest{}

	result, err := suite.service.UpdateUser(userID.String(), req, otherUserID.String(), false)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), utils.ErrForbidden, err)
}

func (suite *UserServiceTestSuite) TestUpdateUser_InvalidUUID() {
	req := &dto.UpdateUserRequest{}

	result, err := suite.service.UpdateUser("not-a-valid-uuid", req, "some-id", false)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), utils.ErrBadRequest, err)
}

func (suite *UserServiceTestSuite) TestUpdateUser_NotFound() {
	userID := uuid.New()

	req := &dto.UpdateUserRequest{}

	suite.userRepo.On("FindByID", userID).Return(nil, gorm.ErrRecordNotFound)

	result, err := suite.service.UpdateUser(userID.String(), req, userID.String(), false)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), utils.ErrNotFound, err)
	suite.userRepo.AssertExpectations(suite.T())
}

// AdminUpdateUser Tests

func (suite *UserServiceTestSuite) TestAdminUpdateUser_Success() {
	userID := uuid.New()
	user := &models.User{
		ID:         userID,
		Username:   "testuser",
		Email:      "test@example.com",
		Role:       models.RoleReader,
		IsActive:   true,
		IsVerified: false,
	}

	newRole := "author"
	isActive := true
	isVerified := true
	req := &dto.AdminUpdateUserRequest{
		Role:       &newRole,
		IsActive:   &isActive,
		IsVerified: &isVerified,
	}

	suite.userRepo.On("FindByID", userID).Return(user, nil)
	suite.userRepo.On("Update", mock.AnythingOfType("*models.User")).Return(nil)

	result, err := suite.service.AdminUpdateUser(userID.String(), req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), newRole, result.Role)
	assert.True(suite.T(), result.IsActive)
	assert.True(suite.T(), result.IsVerified)
	suite.userRepo.AssertExpectations(suite.T())
}

func (suite *UserServiceTestSuite) TestAdminUpdateUser_InvalidRole() {
	userID := uuid.New()
	user := &models.User{
		ID:       userID,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     models.RoleReader,
		IsActive: true,
	}

	invalidRole := "superadmin"
	req := &dto.AdminUpdateUserRequest{
		Role: &invalidRole,
	}

	suite.userRepo.On("FindByID", userID).Return(user, nil)

	result, err := suite.service.AdminUpdateUser(userID.String(), req)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	appErr, ok := utils.IsAppError(err)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), "INVALID_ROLE", appErr.Code)
	suite.userRepo.AssertExpectations(suite.T())
}

func (suite *UserServiceTestSuite) TestAdminUpdateUser_NotFound() {
	userID := uuid.New()

	req := &dto.AdminUpdateUserRequest{}

	suite.userRepo.On("FindByID", userID).Return(nil, gorm.ErrRecordNotFound)

	result, err := suite.service.AdminUpdateUser(userID.String(), req)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), utils.ErrNotFound, err)
	suite.userRepo.AssertExpectations(suite.T())
}

// DeleteUser Tests

func (suite *UserServiceTestSuite) TestDeleteUser_Success() {
	userID := uuid.New()
	user := &models.User{
		ID:       userID,
		Username: "testuser",
		Email:    "test@example.com",
		IsActive: true,
	}

	suite.userRepo.On("FindByID", userID).Return(user, nil)
	suite.userRepo.On("Delete", userID).Return(nil)

	err := suite.service.DeleteUser(userID.String())

	assert.NoError(suite.T(), err)
	suite.userRepo.AssertExpectations(suite.T())
}

func (suite *UserServiceTestSuite) TestDeleteUser_InvalidUUID() {
	err := suite.service.DeleteUser("not-a-valid-uuid")

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), utils.ErrBadRequest, err)
}

func (suite *UserServiceTestSuite) TestDeleteUser_NotFound() {
	userID := uuid.New()

	suite.userRepo.On("FindByID", userID).Return(nil, gorm.ErrRecordNotFound)

	err := suite.service.DeleteUser(userID.String())

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), utils.ErrNotFound, err)
	suite.userRepo.AssertExpectations(suite.T())
}

// GetUserArticles Tests

func (suite *UserServiceTestSuite) TestGetUserArticles_Success() {
	userID := uuid.New()
	user := &models.User{
		ID:       userID,
		Username: "testuser",
		Email:    "test@example.com",
	}

	articles := []models.Article{
		{
			ID:      uuid.New(),
			Title:   "Article 1",
			Slug:    "article-1",
			Content: "Content 1",
			Status:  models.StatusPublished,
		},
		{
			ID:      uuid.New(),
			Title:   "Article 2",
			Slug:    "article-2",
			Content: "Content 2",
			Status:  models.StatusPublished,
		},
	}

	query := &dto.PaginationQuery{Page: 1, PerPage: 10}

	suite.userRepo.On("FindByID", userID).Return(user, nil)
	suite.articleRepo.On("FindByAuthor", userID, mock.AnythingOfType("repositories.ArticleFilters")).
		Return(articles, int64(2), nil)

	result, total, err := suite.service.GetUserArticles(userID.String(), query)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 2)
	assert.Equal(suite.T(), int64(2), total)
	suite.userRepo.AssertExpectations(suite.T())
	suite.articleRepo.AssertExpectations(suite.T())
}

func (suite *UserServiceTestSuite) TestGetUserArticles_InvalidUUID() {
	query := &dto.PaginationQuery{Page: 1, PerPage: 10}

	result, total, err := suite.service.GetUserArticles("not-a-valid-uuid", query)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), int64(0), total)
	assert.Equal(suite.T(), utils.ErrBadRequest, err)
}

func (suite *UserServiceTestSuite) TestGetUserArticles_UserNotFound() {
	userID := uuid.New()
	query := &dto.PaginationQuery{Page: 1, PerPage: 10}

	suite.userRepo.On("FindByID", userID).Return(nil, gorm.ErrRecordNotFound)

	result, total, err := suite.service.GetUserArticles(userID.String(), query)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), int64(0), total)
	assert.Equal(suite.T(), utils.ErrNotFound, err)
	suite.userRepo.AssertExpectations(suite.T())
}

func (suite *UserServiceTestSuite) TestGetUserArticles_NoArticles() {
	userID := uuid.New()
	user := &models.User{
		ID:       userID,
		Username: "testuser",
		Email:    "test@example.com",
	}

	query := &dto.PaginationQuery{Page: 1, PerPage: 10}

	suite.userRepo.On("FindByID", userID).Return(user, nil)
	suite.articleRepo.On("FindByAuthor", userID, mock.AnythingOfType("repositories.ArticleFilters")).
		Return([]models.Article{}, int64(0), nil)

	result, total, err := suite.service.GetUserArticles(userID.String(), query)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 0)
	assert.Equal(suite.T(), int64(0), total)
	suite.userRepo.AssertExpectations(suite.T())
	suite.articleRepo.AssertExpectations(suite.T())
}
