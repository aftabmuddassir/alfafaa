package services

import (
	"testing"
	"time"

	"github.com/alfafaa/alfafaa-blog/internal/dto"
	"github.com/alfafaa/alfafaa-blog/internal/models"
	"github.com/alfafaa/alfafaa-blog/tests/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// helper to create engagement service with mocks
func newTestEngagementService() (EngagementService, *mocks.MockEngagementRepository, *mocks.MockArticleRepository, *mocks.MockCommentRepository, *mocks.MockUserRepository) {
	engagementRepo := new(mocks.MockEngagementRepository)
	articleRepo := new(mocks.MockArticleRepository)
	commentRepo := new(mocks.MockCommentRepository)
	userRepo := new(mocks.MockUserRepository)

	service := NewEngagementService(engagementRepo, articleRepo, commentRepo, userRepo)
	return service, engagementRepo, articleRepo, commentRepo, userRepo
}

// ==================== LIKES ====================

func TestLikeArticle_Success(t *testing.T) {
	service, engagementRepo, articleRepo, _, userRepo := newTestEngagementService()

	userID := uuid.New()
	articleID := uuid.New()
	authorID := uuid.New()

	article := &models.Article{ID: articleID, Slug: "test-article", AuthorID: authorID}
	user := &models.User{ID: userID, FirstName: "John", LastName: "Doe"}

	articleRepo.On("FindBySlug", "test-article").Return(article, nil)
	engagementRepo.On("HasLiked", userID, articleID).Return(false, nil)
	engagementRepo.On("CreateLike", mock.AnythingOfType("*models.Like")).Return(nil)
	userRepo.On("FindByID", userID).Return(user, nil)
	engagementRepo.On("CreateNotification", mock.AnythingOfType("*models.Notification")).Return(nil)
	engagementRepo.On("GetLikesCount", articleID).Return(int64(1), nil)

	result, err := service.LikeArticle(userID.String(), "test-article")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Liked)
	assert.Equal(t, 1, result.LikesCount)

	articleRepo.AssertExpectations(t)
	engagementRepo.AssertExpectations(t)
}

func TestLikeArticle_Idempotent(t *testing.T) {
	service, engagementRepo, articleRepo, _, _ := newTestEngagementService()

	userID := uuid.New()
	articleID := uuid.New()

	article := &models.Article{ID: articleID, Slug: "test-article", AuthorID: uuid.New()}

	articleRepo.On("FindBySlug", "test-article").Return(article, nil)
	engagementRepo.On("HasLiked", userID, articleID).Return(true, nil) // already liked
	engagementRepo.On("GetLikesCount", articleID).Return(int64(5), nil)

	result, err := service.LikeArticle(userID.String(), "test-article")

	assert.NoError(t, err)
	assert.True(t, result.Liked)
	assert.Equal(t, 5, result.LikesCount)

	// CreateLike should NOT have been called
	engagementRepo.AssertNotCalled(t, "CreateLike", mock.Anything)
}

func TestLikeArticle_NoSelfNotification(t *testing.T) {
	service, engagementRepo, articleRepo, _, _ := newTestEngagementService()

	userID := uuid.New()
	articleID := uuid.New()

	// User is the author of the article
	article := &models.Article{ID: articleID, Slug: "my-article", AuthorID: userID}

	articleRepo.On("FindBySlug", "my-article").Return(article, nil)
	engagementRepo.On("HasLiked", userID, articleID).Return(false, nil)
	engagementRepo.On("CreateLike", mock.AnythingOfType("*models.Like")).Return(nil)
	engagementRepo.On("GetLikesCount", articleID).Return(int64(1), nil)

	result, err := service.LikeArticle(userID.String(), "my-article")

	assert.NoError(t, err)
	assert.True(t, result.Liked)

	// No notification should be created for self-like
	engagementRepo.AssertNotCalled(t, "CreateNotification", mock.Anything)
}

func TestLikeArticle_ArticleNotFound(t *testing.T) {
	service, _, articleRepo, _, _ := newTestEngagementService()

	articleRepo.On("FindBySlug", "nonexistent").Return(nil, gorm.ErrRecordNotFound)

	result, err := service.LikeArticle(uuid.New().String(), "nonexistent")

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestLikeArticle_InvalidUserID(t *testing.T) {
	service, _, _, _, _ := newTestEngagementService()

	result, err := service.LikeArticle("not-a-uuid", "test-article")

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestUnlikeArticle_Success(t *testing.T) {
	service, engagementRepo, articleRepo, _, _ := newTestEngagementService()

	userID := uuid.New()
	articleID := uuid.New()

	article := &models.Article{ID: articleID, Slug: "test-article"}

	articleRepo.On("FindBySlug", "test-article").Return(article, nil)
	engagementRepo.On("DeleteLike", userID, articleID).Return(nil)
	engagementRepo.On("GetLikesCount", articleID).Return(int64(0), nil)

	result, err := service.UnlikeArticle(userID.String(), "test-article")

	assert.NoError(t, err)
	assert.False(t, result.Liked)
	assert.Equal(t, 0, result.LikesCount)
}

func TestGetLikeStatus_Liked(t *testing.T) {
	service, engagementRepo, articleRepo, _, _ := newTestEngagementService()

	userID := uuid.New()
	articleID := uuid.New()

	article := &models.Article{ID: articleID, Slug: "test-article"}

	articleRepo.On("FindBySlug", "test-article").Return(article, nil)
	engagementRepo.On("HasLiked", userID, articleID).Return(true, nil)
	engagementRepo.On("GetLikesCount", articleID).Return(int64(10), nil)

	result, err := service.GetLikeStatus(userID.String(), "test-article")

	assert.NoError(t, err)
	assert.True(t, result.Liked)
	assert.Equal(t, 10, result.LikesCount)
}

func TestGetLikeStatus_NotLiked(t *testing.T) {
	service, engagementRepo, articleRepo, _, _ := newTestEngagementService()

	userID := uuid.New()
	articleID := uuid.New()

	article := &models.Article{ID: articleID, Slug: "test-article"}

	articleRepo.On("FindBySlug", "test-article").Return(article, nil)
	engagementRepo.On("HasLiked", userID, articleID).Return(false, nil)
	engagementRepo.On("GetLikesCount", articleID).Return(int64(10), nil)

	result, err := service.GetLikeStatus(userID.String(), "test-article")

	assert.NoError(t, err)
	assert.False(t, result.Liked)
	assert.Equal(t, 10, result.LikesCount)
}

// ==================== BOOKMARKS ====================

func TestBookmarkArticle_Success(t *testing.T) {
	service, engagementRepo, articleRepo, _, _ := newTestEngagementService()

	userID := uuid.New()
	articleID := uuid.New()

	article := &models.Article{ID: articleID, Slug: "test-article"}

	articleRepo.On("FindBySlug", "test-article").Return(article, nil)
	engagementRepo.On("HasBookmarked", userID, articleID).Return(false, nil)
	engagementRepo.On("CreateBookmark", mock.AnythingOfType("*models.Bookmark")).Return(nil)

	result, err := service.BookmarkArticle(userID.String(), "test-article")

	assert.NoError(t, err)
	assert.True(t, result.Bookmarked)
}

func TestBookmarkArticle_Idempotent(t *testing.T) {
	service, engagementRepo, articleRepo, _, _ := newTestEngagementService()

	userID := uuid.New()
	articleID := uuid.New()

	article := &models.Article{ID: articleID, Slug: "test-article"}

	articleRepo.On("FindBySlug", "test-article").Return(article, nil)
	engagementRepo.On("HasBookmarked", userID, articleID).Return(true, nil) // already bookmarked

	result, err := service.BookmarkArticle(userID.String(), "test-article")

	assert.NoError(t, err)
	assert.True(t, result.Bookmarked)

	// CreateBookmark should NOT have been called
	engagementRepo.AssertNotCalled(t, "CreateBookmark", mock.Anything)
}

func TestBookmarkArticle_ArticleNotFound(t *testing.T) {
	service, _, articleRepo, _, _ := newTestEngagementService()

	articleRepo.On("FindBySlug", "nonexistent").Return(nil, gorm.ErrRecordNotFound)

	result, err := service.BookmarkArticle(uuid.New().String(), "nonexistent")

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestUnbookmarkArticle_Success(t *testing.T) {
	service, engagementRepo, articleRepo, _, _ := newTestEngagementService()

	userID := uuid.New()
	articleID := uuid.New()

	article := &models.Article{ID: articleID, Slug: "test-article"}

	articleRepo.On("FindBySlug", "test-article").Return(article, nil)
	engagementRepo.On("DeleteBookmark", userID, articleID).Return(nil)

	result, err := service.UnbookmarkArticle(userID.String(), "test-article")

	assert.NoError(t, err)
	assert.False(t, result.Bookmarked)
}

func TestGetBookmarkedArticles_Success(t *testing.T) {
	service, engagementRepo, _, _, _ := newTestEngagementService()

	userID := uuid.New()
	author := &models.User{ID: uuid.New(), Username: "author1", FirstName: "Jane"}
	articles := []models.Article{
		{ID: uuid.New(), Title: "Bookmarked 1", Slug: "bookmarked-1", Status: models.StatusPublished, Author: author},
		{ID: uuid.New(), Title: "Bookmarked 2", Slug: "bookmarked-2", Status: models.StatusPublished, Author: author},
	}

	query := &dto.PaginationQuery{Page: 1, PerPage: 20}
	engagementRepo.On("GetBookmarkedArticles", userID, 20, 0).Return(articles, int64(2), nil)

	result, total, err := service.GetBookmarkedArticles(userID.String(), query)

	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, result, 2)
	assert.Equal(t, "Bookmarked 1", result[0].Title)
}

func TestGetBookmarkedArticles_InvalidUserID(t *testing.T) {
	service, _, _, _, _ := newTestEngagementService()

	query := &dto.PaginationQuery{Page: 1, PerPage: 20}
	result, total, err := service.GetBookmarkedArticles("not-a-uuid", query)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, int64(0), total)
}

// ==================== COMMENTS ====================

func TestGetComments_Success(t *testing.T) {
	service, _, articleRepo, commentRepo, _ := newTestEngagementService()

	articleID := uuid.New()
	article := &models.Article{ID: articleID, Slug: "test-article"}
	commenter := &models.User{ID: uuid.New(), Username: "commenter1", FirstName: "John"}

	comments := []models.Comment{
		{
			ID:        uuid.New(),
			ArticleID: articleID,
			UserID:    commenter.ID,
			Content:   "Great article!",
			User:      commenter,
			Replies:   []models.Comment{},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	query := &dto.PaginationQuery{Page: 1, PerPage: 20}

	articleRepo.On("FindBySlug", "test-article").Return(article, nil)
	commentRepo.On("FindByArticle", articleID, mock.AnythingOfType("repositories.CommentFilters")).Return(comments, int64(1), nil)

	result, total, err := service.GetComments("test-article", query)

	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, result, 1)
	assert.Equal(t, "Great article!", result[0].Content)
	assert.Equal(t, "commenter1", result[0].User.Username)
}

func TestGetComments_ArticleNotFound(t *testing.T) {
	service, _, articleRepo, _, _ := newTestEngagementService()

	articleRepo.On("FindBySlug", "nonexistent").Return(nil, gorm.ErrRecordNotFound)
	query := &dto.PaginationQuery{Page: 1, PerPage: 20}

	result, total, err := service.GetComments("nonexistent", query)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, int64(0), total)
}

func TestCreateComment_Success(t *testing.T) {
	service, engagementRepo, articleRepo, commentRepo, userRepo := newTestEngagementService()

	userID := uuid.New()
	articleID := uuid.New()
	authorID := uuid.New()
	commentID := uuid.New()

	article := &models.Article{ID: articleID, Slug: "test-article", AuthorID: authorID}
	user := &models.User{ID: userID, Username: "commenter", FirstName: "John", LastName: "Doe"}

	articleRepo.On("FindBySlug", "test-article").Return(article, nil)
	commentRepo.On("Create", mock.AnythingOfType("*models.Comment")).Run(func(args mock.Arguments) {
		// Simulate GORM assigning an ID
		comment := args.Get(0).(*models.Comment)
		comment.ID = commentID
	}).Return(nil)
	commentRepo.On("FindByID", commentID).Return(&models.Comment{
		ID:        commentID,
		ArticleID: articleID,
		UserID:    userID,
		Content:   "Nice post!",
		User:      user,
		Replies:   []models.Comment{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil)
	userRepo.On("FindByID", userID).Return(user, nil)
	engagementRepo.On("CreateNotification", mock.AnythingOfType("*models.Notification")).Return(nil)

	req := &dto.CreateCommentRequest{Content: "Nice post!"}
	result, err := service.CreateComment(userID.String(), "test-article", req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Nice post!", result.Content)
	assert.Equal(t, "commenter", result.User.Username)
}

func TestCreateComment_WithParentID(t *testing.T) {
	service, engagementRepo, articleRepo, commentRepo, userRepo := newTestEngagementService()

	userID := uuid.New()
	articleID := uuid.New()
	authorID := uuid.New()
	parentID := uuid.New()
	commentID := uuid.New()

	article := &models.Article{ID: articleID, Slug: "test-article", AuthorID: authorID}
	user := &models.User{ID: userID, Username: "replier", FirstName: "Jane"}

	articleRepo.On("FindBySlug", "test-article").Return(article, nil)
	commentRepo.On("Create", mock.AnythingOfType("*models.Comment")).Run(func(args mock.Arguments) {
		comment := args.Get(0).(*models.Comment)
		comment.ID = commentID
	}).Return(nil)
	commentRepo.On("FindByID", commentID).Return(&models.Comment{
		ID:        commentID,
		ArticleID: articleID,
		UserID:    userID,
		ParentID:  &parentID,
		Content:   "I agree!",
		User:      user,
		Replies:   []models.Comment{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil)
	userRepo.On("FindByID", userID).Return(user, nil)
	engagementRepo.On("CreateNotification", mock.AnythingOfType("*models.Notification")).Return(nil)

	parentIDStr := parentID.String()
	req := &dto.CreateCommentRequest{Content: "I agree!", ParentID: &parentIDStr}
	result, err := service.CreateComment(userID.String(), "test-article", req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.ParentID)
	assert.Equal(t, parentID.String(), *result.ParentID)
}

func TestCreateComment_NoSelfNotification(t *testing.T) {
	service, _, articleRepo, commentRepo, _ := newTestEngagementService()

	userID := uuid.New()
	articleID := uuid.New()
	commentID := uuid.New()

	// User is the article author - commenting on own article
	article := &models.Article{ID: articleID, Slug: "my-article", AuthorID: userID}
	user := &models.User{ID: userID, Username: "author", FirstName: "Self"}

	articleRepo.On("FindBySlug", "my-article").Return(article, nil)
	commentRepo.On("Create", mock.AnythingOfType("*models.Comment")).Run(func(args mock.Arguments) {
		args.Get(0).(*models.Comment).ID = commentID
	}).Return(nil)
	commentRepo.On("FindByID", commentID).Return(&models.Comment{
		ID:        commentID,
		ArticleID: articleID,
		UserID:    userID,
		Content:   "Self comment",
		User:      user,
		Replies:   []models.Comment{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil)

	req := &dto.CreateCommentRequest{Content: "Self comment"}
	result, err := service.CreateComment(userID.String(), "my-article", req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestUpdateComment_Success(t *testing.T) {
	service, _, _, commentRepo, _ := newTestEngagementService()

	userID := uuid.New()
	commentID := uuid.New()
	articleID := uuid.New()

	user := &models.User{ID: userID, Username: "commenter"}
	comment := &models.Comment{
		ID:        commentID,
		ArticleID: articleID,
		UserID:    userID,
		Content:   "Original",
		User:      user,
		Replies:   []models.Comment{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	commentRepo.On("FindByID", commentID).Return(comment, nil).Once()
	commentRepo.On("Update", mock.AnythingOfType("*models.Comment")).Return(nil)
	commentRepo.On("FindByID", commentID).Return(&models.Comment{
		ID:        commentID,
		ArticleID: articleID,
		UserID:    userID,
		Content:   "Updated content",
		User:      user,
		Replies:   []models.Comment{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil)

	req := &dto.UpdateCommentRequest{Content: "Updated content"}
	result, err := service.UpdateComment(userID.String(), "test-article", commentID.String(), req)

	assert.NoError(t, err)
	assert.Equal(t, "Updated content", result.Content)
}

func TestUpdateComment_Forbidden(t *testing.T) {
	service, _, _, commentRepo, _ := newTestEngagementService()

	userID := uuid.New()
	otherUserID := uuid.New()
	commentID := uuid.New()

	comment := &models.Comment{
		ID:      commentID,
		UserID:  otherUserID, // different user owns this comment
		Content: "Original",
	}

	commentRepo.On("FindByID", commentID).Return(comment, nil)

	req := &dto.UpdateCommentRequest{Content: "Trying to update"}
	result, err := service.UpdateComment(userID.String(), "test-article", commentID.String(), req)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestUpdateComment_NotFound(t *testing.T) {
	service, _, _, commentRepo, _ := newTestEngagementService()

	commentID := uuid.New()
	commentRepo.On("FindByID", commentID).Return(nil, gorm.ErrRecordNotFound)

	req := &dto.UpdateCommentRequest{Content: "Trying to update"}
	result, err := service.UpdateComment(uuid.New().String(), "test-article", commentID.String(), req)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestDeleteComment_ByOwner(t *testing.T) {
	service, _, _, commentRepo, _ := newTestEngagementService()

	userID := uuid.New()
	commentID := uuid.New()

	comment := &models.Comment{ID: commentID, UserID: userID, Content: "My comment"}

	commentRepo.On("FindByID", commentID).Return(comment, nil)
	commentRepo.On("Delete", commentID).Return(nil)

	err := service.DeleteComment(userID.String(), "test-article", commentID.String(), false)

	assert.NoError(t, err)
}

func TestDeleteComment_ByAdmin(t *testing.T) {
	service, _, _, commentRepo, _ := newTestEngagementService()

	adminID := uuid.New()
	ownerID := uuid.New()
	commentID := uuid.New()

	comment := &models.Comment{ID: commentID, UserID: ownerID, Content: "Someone's comment"}

	commentRepo.On("FindByID", commentID).Return(comment, nil)
	commentRepo.On("Delete", commentID).Return(nil)

	err := service.DeleteComment(adminID.String(), "test-article", commentID.String(), true)

	assert.NoError(t, err)
}

func TestDeleteComment_ForbiddenNonOwnerNonAdmin(t *testing.T) {
	service, _, _, commentRepo, _ := newTestEngagementService()

	otherUserID := uuid.New()
	ownerID := uuid.New()
	commentID := uuid.New()

	comment := &models.Comment{ID: commentID, UserID: ownerID, Content: "Someone's comment"}

	commentRepo.On("FindByID", commentID).Return(comment, nil)

	err := service.DeleteComment(otherUserID.String(), "test-article", commentID.String(), false)

	assert.Error(t, err)
	commentRepo.AssertNotCalled(t, "Delete", mock.Anything)
}

// ==================== NOTIFICATIONS ====================

func TestGetNotifications_Success(t *testing.T) {
	service, engagementRepo, _, _, _ := newTestEngagementService()

	userID := uuid.New()
	actorID := uuid.New()
	articleID := uuid.New()

	actor := &models.User{ID: actorID, Username: "actor", FirstName: "Jane", LastName: "Doe"}
	article := &models.Article{ID: articleID, Slug: "liked-article", Title: "Liked Article"}

	notifications := []models.Notification{
		{
			ID:        uuid.New(),
			UserID:    userID,
			ActorID:   actorID,
			Type:      models.NotificationTypeLike,
			Message:   "Jane Doe liked your article",
			ArticleID: &articleID,
			Read:      false,
			CreatedAt: time.Now(),
			Actor:     actor,
			Article:   article,
		},
	}

	query := &dto.PaginationQuery{Page: 1, PerPage: 20}
	engagementRepo.On("GetNotifications", userID, 20, 0).Return(notifications, int64(1), nil)

	result, total, err := service.GetNotifications(userID.String(), query)

	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, result, 1)
	assert.Equal(t, "like", result[0].Type)
	assert.Equal(t, "Jane Doe liked your article", result[0].Message)
	assert.Equal(t, "actor", result[0].Actor.Username)
	assert.NotNil(t, result[0].Article)
	assert.Equal(t, "Liked Article", result[0].Article.Title)
	assert.False(t, result[0].Read)
}

func TestGetNotifications_FollowTypeNoArticle(t *testing.T) {
	service, engagementRepo, _, _, _ := newTestEngagementService()

	userID := uuid.New()
	actorID := uuid.New()

	actor := &models.User{ID: actorID, Username: "follower", FirstName: "John"}

	notifications := []models.Notification{
		{
			ID:        uuid.New(),
			UserID:    userID,
			ActorID:   actorID,
			Type:      models.NotificationTypeFollow,
			Message:   "John started following you",
			ArticleID: nil, // follow notifications have no article
			Read:      false,
			CreatedAt: time.Now(),
			Actor:     actor,
			Article:   nil,
		},
	}

	query := &dto.PaginationQuery{Page: 1, PerPage: 20}
	engagementRepo.On("GetNotifications", userID, 20, 0).Return(notifications, int64(1), nil)

	result, _, err := service.GetNotifications(userID.String(), query)

	assert.NoError(t, err)
	assert.Equal(t, "follow", result[0].Type)
	assert.Nil(t, result[0].Article)
}

func TestGetUnreadCount_Success(t *testing.T) {
	service, engagementRepo, _, _, _ := newTestEngagementService()

	userID := uuid.New()
	engagementRepo.On("GetUnreadCount", userID).Return(int64(5), nil)

	result, err := service.GetUnreadCount(userID.String())

	assert.NoError(t, err)
	assert.Equal(t, int64(5), result.Count)
}

func TestGetUnreadCount_InvalidUserID(t *testing.T) {
	service, _, _, _, _ := newTestEngagementService()

	result, err := service.GetUnreadCount("bad-uuid")

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestMarkNotificationAsRead_Success(t *testing.T) {
	service, engagementRepo, _, _, _ := newTestEngagementService()

	userID := uuid.New()
	notifID := uuid.New()

	engagementRepo.On("MarkAsRead", notifID, userID).Return(nil)

	err := service.MarkNotificationAsRead(userID.String(), notifID.String())

	assert.NoError(t, err)
}

func TestMarkNotificationAsRead_NotFound(t *testing.T) {
	service, engagementRepo, _, _, _ := newTestEngagementService()

	userID := uuid.New()
	notifID := uuid.New()

	engagementRepo.On("MarkAsRead", notifID, userID).Return(gorm.ErrRecordNotFound)

	err := service.MarkNotificationAsRead(userID.String(), notifID.String())

	assert.Error(t, err)
}

func TestMarkAllNotificationsAsRead_Success(t *testing.T) {
	service, engagementRepo, _, _, _ := newTestEngagementService()

	userID := uuid.New()
	engagementRepo.On("MarkAllAsRead", userID).Return(nil)

	err := service.MarkAllNotificationsAsRead(userID.String())

	assert.NoError(t, err)
}

// ==================== NOTIFICATION CREATION ====================

func TestCreateNotification_Success(t *testing.T) {
	service, engagementRepo, _, _, _ := newTestEngagementService()

	actorID := uuid.New()
	recipientID := uuid.New()
	articleID := uuid.New()

	engagementRepo.On("CreateNotification", mock.AnythingOfType("*models.Notification")).Return(nil)

	err := service.CreateNotification(actorID, recipientID, models.NotificationTypeLike, "Test liked", &articleID)

	assert.NoError(t, err)
	engagementRepo.AssertCalled(t, "CreateNotification", mock.AnythingOfType("*models.Notification"))
}

func TestCreateNotification_SkipSelfNotification(t *testing.T) {
	service, engagementRepo, _, _, _ := newTestEngagementService()

	userID := uuid.New()
	articleID := uuid.New()

	// Actor and recipient are the same
	err := service.CreateNotification(userID, userID, models.NotificationTypeLike, "Self like", &articleID)

	assert.NoError(t, err)
	// Should NOT call CreateNotification on repo
	engagementRepo.AssertNotCalled(t, "CreateNotification", mock.Anything)
}

// ==================== ARTICLE ENGAGEMENT ====================

func TestGetArticleEngagement_WithAuthUser(t *testing.T) {
	service, engagementRepo, _, _, _ := newTestEngagementService()

	articleID := uuid.New()
	userID := uuid.New()

	engagementRepo.On("GetLikesCount", articleID).Return(int64(42), nil)
	engagementRepo.On("GetCommentsCount", articleID).Return(int64(7), nil)
	engagementRepo.On("HasLiked", userID, articleID).Return(true, nil)
	engagementRepo.On("HasBookmarked", userID, articleID).Return(false, nil)

	likesCount, commentsCount, userLiked, userBookmarked := service.GetArticleEngagement(articleID, userID.String())

	assert.Equal(t, 42, likesCount)
	assert.Equal(t, 7, commentsCount)
	assert.True(t, userLiked)
	assert.False(t, userBookmarked)
}

func TestGetArticleEngagement_NoAuthUser(t *testing.T) {
	service, engagementRepo, _, _, _ := newTestEngagementService()

	articleID := uuid.New()

	engagementRepo.On("GetLikesCount", articleID).Return(int64(10), nil)
	engagementRepo.On("GetCommentsCount", articleID).Return(int64(3), nil)

	likesCount, commentsCount, userLiked, userBookmarked := service.GetArticleEngagement(articleID, "")

	assert.Equal(t, 10, likesCount)
	assert.Equal(t, 3, commentsCount)
	assert.False(t, userLiked)
	assert.False(t, userBookmarked)

	// HasLiked/HasBookmarked should NOT be called with empty userID
	engagementRepo.AssertNotCalled(t, "HasLiked", mock.Anything, mock.Anything)
	engagementRepo.AssertNotCalled(t, "HasBookmarked", mock.Anything, mock.Anything)
}
