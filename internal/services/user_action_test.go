package services

import (
	"testing"

	"github.com/alfafaa/alfafaa-blog/internal/dto"
	"github.com/alfafaa/alfafaa-blog/internal/models"
	"github.com/alfafaa/alfafaa-blog/tests/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestFollowUser_Success(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepository)
	mockArticleRepo := new(mocks.MockArticleRepository)

	service := NewUserService(mockUserRepo, mockArticleRepo)

	followerID := uuid.New()
	followingID := uuid.New()

	targetUser := &models.User{
		ID:       followingID,
		Username: "targetuser",
		Email:    "target@example.com",
		IsActive: true,
	}

	// Setup expectations
	mockUserRepo.On("FindByID", followingID).Return(targetUser, nil)
	mockUserRepo.On("IsFollowing", followerID, followingID).Return(false, nil)
	mockUserRepo.On("FollowUser", followerID, followingID).Return(nil)

	// Execute
	result, err := service.FollowUser(followerID.String(), followingID.String())

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.IsFollowing)

	mockUserRepo.AssertExpectations(t)
}

func TestFollowUser_AlreadyFollowing(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepository)
	mockArticleRepo := new(mocks.MockArticleRepository)

	service := NewUserService(mockUserRepo, mockArticleRepo)

	followerID := uuid.New()
	followingID := uuid.New()

	targetUser := &models.User{
		ID:       followingID,
		Username: "targetuser",
		Email:    "target@example.com",
		IsActive: true,
	}

	// Setup expectations
	mockUserRepo.On("FindByID", followingID).Return(targetUser, nil)
	mockUserRepo.On("IsFollowing", followerID, followingID).Return(true, nil)

	// Execute
	result, err := service.FollowUser(followerID.String(), followingID.String())

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.IsFollowing) // Still returns true

	mockUserRepo.AssertExpectations(t)
}

func TestFollowUser_CannotFollowSelf(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepository)
	mockArticleRepo := new(mocks.MockArticleRepository)

	service := NewUserService(mockUserRepo, mockArticleRepo)

	userID := uuid.New()

	// Execute
	result, err := service.FollowUser(userID.String(), userID.String())

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "cannot follow yourself")
}

func TestFollowUser_UserNotFound(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepository)
	mockArticleRepo := new(mocks.MockArticleRepository)

	service := NewUserService(mockUserRepo, mockArticleRepo)

	followerID := uuid.New()
	followingID := uuid.New()

	// Setup expectations
	mockUserRepo.On("FindByID", followingID).Return(nil, gorm.ErrRecordNotFound)

	// Execute
	result, err := service.FollowUser(followerID.String(), followingID.String())

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)

	mockUserRepo.AssertExpectations(t)
}

func TestUnfollowUser_Success(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepository)
	mockArticleRepo := new(mocks.MockArticleRepository)

	service := NewUserService(mockUserRepo, mockArticleRepo)

	followerID := uuid.New()
	followingID := uuid.New()

	// Setup expectations
	mockUserRepo.On("UnfollowUser", followerID, followingID).Return(nil)

	// Execute
	result, err := service.UnfollowUser(followerID.String(), followingID.String())

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.IsFollowing)

	mockUserRepo.AssertExpectations(t)
}

func TestUnfollowUser_CannotUnfollowSelf(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepository)
	mockArticleRepo := new(mocks.MockArticleRepository)

	service := NewUserService(mockUserRepo, mockArticleRepo)

	userID := uuid.New()

	// Execute
	result, err := service.UnfollowUser(userID.String(), userID.String())

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "cannot unfollow yourself")
}

func TestGetFollowers_Success(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepository)
	mockArticleRepo := new(mocks.MockArticleRepository)

	service := NewUserService(mockUserRepo, mockArticleRepo)

	userID := uuid.New()
	followers := []models.User{
		{ID: uuid.New(), Username: "follower1", FirstName: "John", LastName: "Doe"},
		{ID: uuid.New(), Username: "follower2", FirstName: "Jane", LastName: "Doe"},
	}

	query := &dto.PaginationQuery{Page: 1, PerPage: 20}

	// Setup expectations
	mockUserRepo.On("GetFollowers", userID, 20, 0).Return(followers, int64(2), nil)

	// Execute
	result, err := service.GetFollowers(userID.String(), query)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(2), result.Total)
	assert.Len(t, result.Users, 2)

	mockUserRepo.AssertExpectations(t)
}

func TestGetFollowing_Success(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepository)
	mockArticleRepo := new(mocks.MockArticleRepository)

	service := NewUserService(mockUserRepo, mockArticleRepo)

	userID := uuid.New()
	following := []models.User{
		{ID: uuid.New(), Username: "following1", FirstName: "John", LastName: "Doe"},
	}

	query := &dto.PaginationQuery{Page: 1, PerPage: 20}

	// Setup expectations
	mockUserRepo.On("GetFollowing", userID, 20, 0).Return(following, int64(1), nil)

	// Execute
	result, err := service.GetFollowing(userID.String(), query)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(1), result.Total)
	assert.Len(t, result.Users, 1)

	mockUserRepo.AssertExpectations(t)
}

func TestSetInterests_Success(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepository)
	mockArticleRepo := new(mocks.MockArticleRepository)

	service := NewUserService(mockUserRepo, mockArticleRepo)

	userID := uuid.New()
	cat1ID := uuid.New()
	cat2ID := uuid.New()

	req := &dto.SetInterestsRequest{
		CategoryIDs: []string{cat1ID.String(), cat2ID.String()},
	}

	categories := []models.Category{
		{ID: cat1ID, Name: "Technology", Slug: "technology"},
		{ID: cat2ID, Name: "Science", Slug: "science"},
	}

	// Setup expectations
	mockUserRepo.On("SetInterests", userID, []uuid.UUID{cat1ID, cat2ID}).Return(nil)
	mockUserRepo.On("GetInterests", userID).Return(categories, nil)

	// Execute
	result, err := service.SetInterests(userID.String(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)
	assert.Equal(t, "Technology", result[0].Name)
	assert.Equal(t, "Science", result[1].Name)

	mockUserRepo.AssertExpectations(t)
}

func TestGetInterests_Success(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepository)
	mockArticleRepo := new(mocks.MockArticleRepository)

	service := NewUserService(mockUserRepo, mockArticleRepo)

	userID := uuid.New()
	categories := []models.Category{
		{ID: uuid.New(), Name: "Technology", Slug: "technology"},
		{ID: uuid.New(), Name: "Science", Slug: "science"},
	}

	// Setup expectations
	mockUserRepo.On("GetInterests", userID).Return(categories, nil)

	// Execute
	result, err := service.GetInterests(userID.String())

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)

	mockUserRepo.AssertExpectations(t)
}

func TestGetPersonalizedFeed_Success(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepository)
	mockArticleRepo := new(mocks.MockArticleRepository)

	service := NewUserService(mockUserRepo, mockArticleRepo)

	userID := uuid.New()
	followingID := uuid.New()
	interestID := uuid.New()

	query := &dto.PaginationQuery{Page: 1, PerPage: 20}

	author := &models.User{
		ID:        followingID,
		Username:  "author1",
		FirstName: "John",
		LastName:  "Doe",
	}

	articles := []models.Article{
		{
			ID:       uuid.New(),
			Title:    "Article 1",
			Slug:     "article-1",
			Status:   models.StatusPublished,
			AuthorID: followingID,
			Author:   author,
		},
	}

	// Setup expectations
	mockUserRepo.On("GetFollowingIDs", userID).Return([]uuid.UUID{followingID}, nil)
	mockUserRepo.On("GetInterestIDs", userID).Return([]uuid.UUID{interestID}, nil)
	mockArticleRepo.On("FindForUser", userID, []uuid.UUID{followingID}, []uuid.UUID{interestID}, mock.AnythingOfType("repositories.ArticleFilters")).Return(articles, int64(1), nil)

	// Execute
	result, total, err := service.GetPersonalizedFeed(userID.String(), query)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(1), total)
	assert.Len(t, result, 1)
	assert.Equal(t, "Article 1", result[0].Title)

	mockUserRepo.AssertExpectations(t)
	mockArticleRepo.AssertExpectations(t)
}

func TestGetStaffPicks_Success(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepository)
	mockArticleRepo := new(mocks.MockArticleRepository)

	service := NewUserService(mockUserRepo, mockArticleRepo)

	query := &dto.PaginationQuery{Page: 1, PerPage: 20}

	authorID := uuid.New()
	author := &models.User{
		ID:        authorID,
		Username:  "author1",
		FirstName: "John",
		LastName:  "Doe",
	}

	articles := []models.Article{
		{
			ID:          uuid.New(),
			Title:       "Staff Pick Article",
			Slug:        "staff-pick-article",
			Status:      models.StatusPublished,
			IsStaffPick: true,
			AuthorID:    authorID,
			Author:      author,
		},
	}

	// Setup expectations
	mockArticleRepo.On("FindStaffPicks", mock.AnythingOfType("repositories.ArticleFilters")).Return(articles, int64(1), nil)

	// Execute
	result, total, err := service.GetStaffPicks(query)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(1), total)
	assert.Len(t, result, 1)
	assert.Equal(t, "Staff Pick Article", result[0].Title)

	mockArticleRepo.AssertExpectations(t)
}

func TestGetUserProfile_Success(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepository)
	mockArticleRepo := new(mocks.MockArticleRepository)

	service := NewUserService(mockUserRepo, mockArticleRepo)

	userID := uuid.New()
	currentUserID := uuid.New()

	user := &models.User{
		ID:        userID,
		Username:  "testuser",
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		IsActive:  true,
		Interests: []models.Category{
			{ID: uuid.New(), Name: "Technology", Slug: "technology"},
		},
	}

	// Setup expectations
	mockUserRepo.On("FindByIDWithRelations", userID).Return(user, nil)
	mockArticleRepo.On("FindByAuthor", userID, mock.AnythingOfType("repositories.ArticleFilters")).Return([]models.Article{}, int64(5), nil)
	mockUserRepo.On("GetFollowers", userID, 0, 0).Return([]models.User{}, int64(10), nil)
	mockUserRepo.On("GetFollowing", userID, 0, 0).Return([]models.User{}, int64(20), nil)
	mockUserRepo.On("IsFollowing", currentUserID, userID).Return(true, nil)

	// Execute
	result, err := service.GetUserProfile(userID.String(), currentUserID.String())

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "testuser", result.Username)
	assert.Equal(t, 5, result.ArticleCount)
	assert.Equal(t, 10, result.FollowerCount)
	assert.Equal(t, 20, result.FollowingCount)
	assert.True(t, result.IsFollowing)
	assert.Len(t, result.Interests, 1)

	mockUserRepo.AssertExpectations(t)
	mockArticleRepo.AssertExpectations(t)
}

func TestGetUserProfile_NotFound(t *testing.T) {
	mockUserRepo := new(mocks.MockUserRepository)
	mockArticleRepo := new(mocks.MockArticleRepository)

	service := NewUserService(mockUserRepo, mockArticleRepo)

	userID := uuid.New()

	// Setup expectations
	mockUserRepo.On("FindByIDWithRelations", userID).Return(nil, gorm.ErrRecordNotFound)

	// Execute
	result, err := service.GetUserProfile(userID.String(), "")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)

	mockUserRepo.AssertExpectations(t)
}
