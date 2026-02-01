package services

import (
	"errors"

	"github.com/alfafaa/alfafaa-blog/internal/dto"
	"github.com/alfafaa/alfafaa-blog/internal/models"
	"github.com/alfafaa/alfafaa-blog/internal/repositories"
	"github.com/alfafaa/alfafaa-blog/internal/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserService defines the interface for user operations
type UserService interface {
	GetUser(id string) (*dto.UserDetailResponse, error)
	GetUserProfile(id string, currentUserID string) (*dto.UserProfileResponse, error)
	GetUsers(query *dto.UserListQuery) ([]dto.UserListItemResponse, int64, error)
	UpdateUser(id string, req *dto.UpdateUserRequest, currentUserID string, isAdmin bool) (*dto.UserDetailResponse, error)
	AdminUpdateUser(id string, req *dto.AdminUpdateUserRequest) (*dto.UserDetailResponse, error)
	DeleteUser(id string) error
	GetUserArticles(id string, query *dto.PaginationQuery) ([]dto.ArticleListItemResponse, int64, error)
	// Social graph methods
	FollowUser(followerID, followingID string) (*dto.FollowResponse, error)
	UnfollowUser(followerID, followingID string) (*dto.FollowResponse, error)
	GetFollowers(userID string, query *dto.PaginationQuery) (*dto.FollowListResponse, error)
	GetFollowing(userID string, query *dto.PaginationQuery) (*dto.FollowListResponse, error)
	// Interest methods
	SetInterests(userID string, req *dto.SetInterestsRequest) ([]dto.CategoryResponse, error)
	GetInterests(userID string) ([]dto.CategoryResponse, error)
	// Feed methods
	GetPersonalizedFeed(userID string, query *dto.PaginationQuery) ([]dto.ArticleListItemResponse, int64, error)
	GetStaffPicks(query *dto.PaginationQuery) ([]dto.ArticleListItemResponse, int64, error)
}

type userService struct {
	userRepo    repositories.UserRepository
	articleRepo repositories.ArticleRepository
}

// NewUserService creates a new user service
func NewUserService(userRepo repositories.UserRepository, articleRepo repositories.ArticleRepository) UserService {
	return &userService{
		userRepo:    userRepo,
		articleRepo: articleRepo,
	}
}

// GetUser retrieves a user by ID
func (s *userService) GetUser(id string) (*dto.UserDetailResponse, error) {
	userID, err := uuid.Parse(id)
	if err != nil {
		return nil, utils.ErrBadRequest
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound
		}
		return nil, utils.WrapError(err, "failed to find user")
	}

	// Get article count
	articles, count, _ := s.articleRepo.FindByAuthor(userID, repositories.ArticleFilters{
		Status: string(models.StatusPublished),
	})
	_ = articles

	return s.toDetailResponse(user, int(count)), nil
}

// GetUsers retrieves a list of users
func (s *userService) GetUsers(query *dto.UserListQuery) ([]dto.UserListItemResponse, int64, error) {
	filters := repositories.UserFilters{
		Role:   query.Role,
		Search: query.Search,
		Limit:  query.GetPerPage(),
		Offset: query.GetOffset(),
	}

	if query.IsActive != nil {
		filters.IsActive = query.IsActive
	}

	users, total, err := s.userRepo.FindAll(filters)
	if err != nil {
		return nil, 0, utils.WrapError(err, "failed to find users")
	}

	responses := make([]dto.UserListItemResponse, len(users))
	for i, user := range users {
		responses[i] = s.toListItemResponse(&user)
	}

	return responses, total, nil
}

// UpdateUser updates a user's profile
func (s *userService) UpdateUser(id string, req *dto.UpdateUserRequest, currentUserID string, isAdmin bool) (*dto.UserDetailResponse, error) {
	userID, err := uuid.Parse(id)
	if err != nil {
		return nil, utils.ErrBadRequest
	}

	// Check if user is updating their own profile or is an admin
	if id != currentUserID && !isAdmin {
		return nil, utils.ErrForbidden
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound
		}
		return nil, utils.WrapError(err, "failed to find user")
	}

	// Update fields
	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		user.LastName = *req.LastName
	}
	if req.Bio != nil {
		user.Bio = *req.Bio
	}
	if req.ProfileImageURL != nil {
		user.ProfileImageURL = req.ProfileImageURL
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, utils.WrapError(err, "failed to update user")
	}

	return s.toDetailResponse(user, 0), nil
}

// AdminUpdateUser updates a user with admin privileges
func (s *userService) AdminUpdateUser(id string, req *dto.AdminUpdateUserRequest) (*dto.UserDetailResponse, error) {
	userID, err := uuid.Parse(id)
	if err != nil {
		return nil, utils.ErrBadRequest
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound
		}
		return nil, utils.WrapError(err, "failed to find user")
	}

	// Update fields
	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		user.LastName = *req.LastName
	}
	if req.Bio != nil {
		user.Bio = *req.Bio
	}
	if req.ProfileImageURL != nil {
		user.ProfileImageURL = req.ProfileImageURL
	}
	if req.Role != nil {
		role := models.UserRole(*req.Role)
		if !role.IsValid() {
			return nil, utils.NewAppError("INVALID_ROLE", "Invalid role specified", 400)
		}
		user.Role = role
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}
	if req.IsVerified != nil {
		user.IsVerified = *req.IsVerified
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, utils.WrapError(err, "failed to update user")
	}

	return s.toDetailResponse(user, 0), nil
}

// DeleteUser deletes a user
func (s *userService) DeleteUser(id string) error {
	userID, err := uuid.Parse(id)
	if err != nil {
		return utils.ErrBadRequest
	}

	// Check if user exists
	_, err = s.userRepo.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrNotFound
		}
		return utils.WrapError(err, "failed to find user")
	}

	if err := s.userRepo.Delete(userID); err != nil {
		return utils.WrapError(err, "failed to delete user")
	}

	return nil
}

// GetUserArticles retrieves articles by a user
func (s *userService) GetUserArticles(id string, query *dto.PaginationQuery) ([]dto.ArticleListItemResponse, int64, error) {
	userID, err := uuid.Parse(id)
	if err != nil {
		return nil, 0, utils.ErrBadRequest
	}

	// Check if user exists
	_, err = s.userRepo.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, utils.ErrNotFound
		}
		return nil, 0, utils.WrapError(err, "failed to find user")
	}

	filters := repositories.ArticleFilters{
		Status: string(models.StatusPublished),
		Limit:  query.GetPerPage(),
		Offset: query.GetOffset(),
		Sort:   query.GetSort(),
	}

	articles, total, err := s.articleRepo.FindByAuthor(userID, filters)
	if err != nil {
		return nil, 0, utils.WrapError(err, "failed to find articles")
	}

	responses := make([]dto.ArticleListItemResponse, len(articles))
	for i, article := range articles {
		responses[i] = toArticleListItemResponse(&article)
	}

	return responses, total, nil
}

// toDetailResponse converts a user model to a detail response DTO
func (s *userService) toDetailResponse(user *models.User, articleCount int) *dto.UserDetailResponse {
	return &dto.UserDetailResponse{
		ID:              user.ID.String(),
		Username:        user.Username,
		Email:           user.Email,
		FirstName:       user.FirstName,
		LastName:        user.LastName,
		Bio:             user.Bio,
		ProfileImageURL: user.ProfileImageURL,
		Role:            string(user.Role),
		IsVerified:      user.IsVerified,
		IsActive:        user.IsActive,
		ArticleCount:    articleCount,
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
	}
}

// toListItemResponse converts a user model to a list item response DTO
func (s *userService) toListItemResponse(user *models.User) dto.UserListItemResponse {
	return dto.UserListItemResponse{
		ID:              user.ID.String(),
		Username:        user.Username,
		Email:           user.Email,
		FirstName:       user.FirstName,
		LastName:        user.LastName,
		ProfileImageURL: user.ProfileImageURL,
		Role:            string(user.Role),
		IsVerified:      user.IsVerified,
		IsActive:        user.IsActive,
		CreatedAt:       user.CreatedAt,
	}
}

// toArticleListItemResponse converts an article model to a list item response DTO
func toArticleListItemResponse(article *models.Article) dto.ArticleListItemResponse {
	response := dto.ArticleListItemResponse{
		ID:                 article.ID.String(),
		Title:              article.Title,
		Slug:               article.Slug,
		Excerpt:            article.Excerpt,
		FeaturedImageURL:   article.FeaturedImageURL,
		Status:             string(article.Status),
		PublishedAt:        article.PublishedAt,
		ViewCount:          article.ViewCount,
		ReadingTimeMinutes: article.ReadingTimeMinutes,
		CreatedAt:          article.CreatedAt,
	}

	if article.Author != nil {
		response.Author = dto.PublicUserResponse{
			ID:              article.Author.ID.String(),
			Username:        article.Author.Username,
			FirstName:       article.Author.FirstName,
			LastName:        article.Author.LastName,
			Bio:             article.Author.Bio,
			ProfileImageURL: article.Author.ProfileImageURL,
		}
	}

	for _, cat := range article.Categories {
		response.Categories = append(response.Categories, dto.CategoryResponse{
			ID:   cat.ID.String(),
			Name: cat.Name,
			Slug: cat.Slug,
		})
	}

	for _, tag := range article.Tags {
		response.Tags = append(response.Tags, dto.TagResponse{
			ID:   tag.ID.String(),
			Name: tag.Name,
			Slug: tag.Slug,
		})
	}

	return response
}

// GetUserProfile retrieves a user profile with social info
func (s *userService) GetUserProfile(id string, currentUserID string) (*dto.UserProfileResponse, error) {
	userID, err := uuid.Parse(id)
	if err != nil {
		return nil, utils.ErrBadRequest
	}

	user, err := s.userRepo.FindByIDWithRelations(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound
		}
		return nil, utils.WrapError(err, "failed to find user")
	}

	// Get article count
	_, articleCount, _ := s.articleRepo.FindByAuthor(userID, repositories.ArticleFilters{
		Status: string(models.StatusPublished),
	})

	// Get follower/following counts
	_, followerCount, _ := s.userRepo.GetFollowers(userID, 0, 0)
	_, followingCount, _ := s.userRepo.GetFollowing(userID, 0, 0)

	// Check if current user is following this user
	var isFollowing bool
	if currentUserID != "" && currentUserID != id {
		currentUID, err := uuid.Parse(currentUserID)
		if err == nil {
			isFollowing, _ = s.userRepo.IsFollowing(currentUID, userID)
		}
	}

	// Convert interests to DTO
	interests := make([]dto.CategoryResponse, len(user.Interests))
	for i, cat := range user.Interests {
		interests[i] = dto.CategoryResponse{
			ID:   cat.ID.String(),
			Name: cat.Name,
			Slug: cat.Slug,
		}
	}

	return &dto.UserProfileResponse{
		ID:              user.ID.String(),
		Username:        user.Username,
		FirstName:       user.FirstName,
		LastName:        user.LastName,
		Bio:             user.Bio,
		ProfileImageURL: user.ProfileImageURL,
		Role:            string(user.Role),
		IsVerified:      user.IsVerified,
		IsActive:        user.IsActive,
		ArticleCount:    int(articleCount),
		FollowerCount:   int(followerCount),
		FollowingCount:  int(followingCount),
		IsFollowing:     isFollowing,
		Interests:       interests,
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
	}, nil
}

// FollowUser follows another user
func (s *userService) FollowUser(followerID, followingID string) (*dto.FollowResponse, error) {
	followerUUID, err := uuid.Parse(followerID)
	if err != nil {
		return nil, utils.ErrBadRequest
	}

	followingUUID, err := uuid.Parse(followingID)
	if err != nil {
		return nil, utils.ErrBadRequest
	}

	// Cannot follow yourself
	if followerUUID == followingUUID {
		return nil, utils.NewAppError("INVALID_OPERATION", "You cannot follow yourself", 400)
	}

	// Check if user to follow exists
	_, err = s.userRepo.FindByID(followingUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound
		}
		return nil, utils.WrapError(err, "failed to find user")
	}

	// Check if already following
	isFollowing, err := s.userRepo.IsFollowing(followerUUID, followingUUID)
	if err != nil {
		return nil, utils.WrapError(err, "failed to check follow status")
	}

	if isFollowing {
		return &dto.FollowResponse{IsFollowing: true}, nil
	}

	// Create follow relationship
	if err := s.userRepo.FollowUser(followerUUID, followingUUID); err != nil {
		return nil, utils.WrapError(err, "failed to follow user")
	}

	return &dto.FollowResponse{IsFollowing: true}, nil
}

// UnfollowUser unfollows another user
func (s *userService) UnfollowUser(followerID, followingID string) (*dto.FollowResponse, error) {
	followerUUID, err := uuid.Parse(followerID)
	if err != nil {
		return nil, utils.ErrBadRequest
	}

	followingUUID, err := uuid.Parse(followingID)
	if err != nil {
		return nil, utils.ErrBadRequest
	}

	// Cannot unfollow yourself
	if followerUUID == followingUUID {
		return nil, utils.NewAppError("INVALID_OPERATION", "You cannot unfollow yourself", 400)
	}

	// Remove follow relationship
	if err := s.userRepo.UnfollowUser(followerUUID, followingUUID); err != nil {
		return nil, utils.WrapError(err, "failed to unfollow user")
	}

	return &dto.FollowResponse{IsFollowing: false}, nil
}

// GetFollowers returns the followers of a user
func (s *userService) GetFollowers(userID string, query *dto.PaginationQuery) (*dto.FollowListResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, utils.ErrBadRequest
	}

	users, total, err := s.userRepo.GetFollowers(userUUID, query.GetPerPage(), query.GetOffset())
	if err != nil {
		return nil, utils.WrapError(err, "failed to get followers")
	}

	userResponses := make([]dto.PublicUserResponse, len(users))
	for i, user := range users {
		userResponses[i] = dto.PublicUserResponse{
			ID:              user.ID.String(),
			Username:        user.Username,
			FirstName:       user.FirstName,
			LastName:        user.LastName,
			Bio:             user.Bio,
			ProfileImageURL: user.ProfileImageURL,
		}
	}

	return &dto.FollowListResponse{
		Users: userResponses,
		Total: total,
	}, nil
}

// GetFollowing returns the users that a user follows
func (s *userService) GetFollowing(userID string, query *dto.PaginationQuery) (*dto.FollowListResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, utils.ErrBadRequest
	}

	users, total, err := s.userRepo.GetFollowing(userUUID, query.GetPerPage(), query.GetOffset())
	if err != nil {
		return nil, utils.WrapError(err, "failed to get following")
	}

	userResponses := make([]dto.PublicUserResponse, len(users))
	for i, user := range users {
		userResponses[i] = dto.PublicUserResponse{
			ID:              user.ID.String(),
			Username:        user.Username,
			FirstName:       user.FirstName,
			LastName:        user.LastName,
			Bio:             user.Bio,
			ProfileImageURL: user.ProfileImageURL,
		}
	}

	return &dto.FollowListResponse{
		Users: userResponses,
		Total: total,
	}, nil
}

// SetInterests sets the user's interests (categories)
func (s *userService) SetInterests(userID string, req *dto.SetInterestsRequest) ([]dto.CategoryResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, utils.ErrBadRequest
	}

	// Parse category IDs
	categoryIDs := make([]uuid.UUID, len(req.CategoryIDs))
	for i, id := range req.CategoryIDs {
		catUUID, err := uuid.Parse(id)
		if err != nil {
			return nil, utils.NewAppError("INVALID_CATEGORY_ID", "Invalid category ID: "+id, 400)
		}
		categoryIDs[i] = catUUID
	}

	// Set interests
	if err := s.userRepo.SetInterests(userUUID, categoryIDs); err != nil {
		return nil, utils.WrapError(err, "failed to set interests")
	}

	// Return updated interests
	return s.GetInterests(userID)
}

// GetInterests returns the user's interests (categories)
func (s *userService) GetInterests(userID string) ([]dto.CategoryResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, utils.ErrBadRequest
	}

	categories, err := s.userRepo.GetInterests(userUUID)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get interests")
	}

	responses := make([]dto.CategoryResponse, len(categories))
	for i, cat := range categories {
		responses[i] = dto.CategoryResponse{
			ID:   cat.ID.String(),
			Name: cat.Name,
			Slug: cat.Slug,
		}
	}

	return responses, nil
}

// GetPersonalizedFeed returns articles personalized for the user
func (s *userService) GetPersonalizedFeed(userID string, query *dto.PaginationQuery) ([]dto.ArticleListItemResponse, int64, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, 0, utils.ErrBadRequest
	}

	// Get user's following and interests
	followingIDs, err := s.userRepo.GetFollowingIDs(userUUID)
	if err != nil {
		return nil, 0, utils.WrapError(err, "failed to get following")
	}

	interestIDs, err := s.userRepo.GetInterestIDs(userUUID)
	if err != nil {
		return nil, 0, utils.WrapError(err, "failed to get interests")
	}

	filters := repositories.ArticleFilters{
		Status: string(models.StatusPublished),
		Limit:  query.GetPerPage(),
		Offset: query.GetOffset(),
		Sort:   query.GetSort(),
	}

	articles, total, err := s.articleRepo.FindForUser(userUUID, followingIDs, interestIDs, filters)
	if err != nil {
		return nil, 0, utils.WrapError(err, "failed to get personalized feed")
	}

	responses := make([]dto.ArticleListItemResponse, len(articles))
	for i, article := range articles {
		responses[i] = toArticleListItemResponse(&article)
	}

	return responses, total, nil
}

// GetStaffPicks returns articles marked as staff picks
func (s *userService) GetStaffPicks(query *dto.PaginationQuery) ([]dto.ArticleListItemResponse, int64, error) {
	filters := repositories.ArticleFilters{
		Status: string(models.StatusPublished),
		Limit:  query.GetPerPage(),
		Offset: query.GetOffset(),
		Sort:   query.GetSort(),
	}

	articles, total, err := s.articleRepo.FindStaffPicks(filters)
	if err != nil {
		return nil, 0, utils.WrapError(err, "failed to get staff picks")
	}

	responses := make([]dto.ArticleListItemResponse, len(articles))
	for i, article := range articles {
		responses[i] = toArticleListItemResponse(&article)
	}

	return responses, total, nil
}
