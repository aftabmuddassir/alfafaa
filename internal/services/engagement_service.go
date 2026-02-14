package services

import (
	"errors"
	"fmt"

	"github.com/alfafaa/alfafaa-blog/internal/dto"
	"github.com/alfafaa/alfafaa-blog/internal/models"
	"github.com/alfafaa/alfafaa-blog/internal/repositories"
	"github.com/alfafaa/alfafaa-blog/internal/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EngagementService defines the interface for engagement operations
type EngagementService interface {
	// Likes
	LikeArticle(userID, slug string) (*dto.LikeResponse, error)
	UnlikeArticle(userID, slug string) (*dto.LikeResponse, error)
	GetLikeStatus(userID, slug string) (*dto.LikeResponse, error)

	// Bookmarks
	BookmarkArticle(userID, slug string) (*dto.BookmarkResponse, error)
	UnbookmarkArticle(userID, slug string) (*dto.BookmarkResponse, error)
	GetBookmarkedArticles(userID string, query *dto.PaginationQuery) ([]dto.ArticleListItemResponse, int64, error)

	// Comments
	GetComments(slug string, query *dto.PaginationQuery) ([]dto.EngagementCommentResponse, int64, error)
	CreateComment(userID, slug string, req *dto.CreateCommentRequest) (*dto.EngagementCommentResponse, error)
	UpdateComment(userID, slug, commentID string, req *dto.UpdateCommentRequest) (*dto.EngagementCommentResponse, error)
	DeleteComment(userID, slug, commentID string, isAdmin bool) error

	// Notifications
	GetNotifications(userID string, query *dto.PaginationQuery) ([]dto.NotificationResponse, int64, error)
	GetUnreadCount(userID string) (*dto.UnreadCountResponse, error)
	MarkNotificationAsRead(userID, notificationID string) error
	MarkAllNotificationsAsRead(userID string) error

	// Notification creation (called as side effects)
	CreateNotification(actorID uuid.UUID, recipientID uuid.UUID, notifType models.NotificationType, message string, articleID *uuid.UUID) error

	// Article engagement data
	GetArticleEngagement(articleID uuid.UUID, userID string) (likesCount int, commentsCount int, userLiked bool, userBookmarked bool)
}

type engagementService struct {
	engagementRepo repositories.EngagementRepository
	articleRepo    repositories.ArticleRepository
	commentRepo    repositories.CommentRepository
	userRepo       repositories.UserRepository
}

// NewEngagementService creates a new engagement service
func NewEngagementService(
	engagementRepo repositories.EngagementRepository,
	articleRepo repositories.ArticleRepository,
	commentRepo repositories.CommentRepository,
	userRepo repositories.UserRepository,
) EngagementService {
	return &engagementService{
		engagementRepo: engagementRepo,
		articleRepo:    articleRepo,
		commentRepo:    commentRepo,
		userRepo:       userRepo,
	}
}

// --- Likes ---

func (s *engagementService) LikeArticle(userID, slug string) (*dto.LikeResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, utils.ErrBadRequest
	}

	article, err := s.articleRepo.FindBySlug(slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound
		}
		return nil, utils.WrapError(err, "failed to find article")
	}

	// Idempotent: if already liked, just return current status
	liked, err := s.engagementRepo.HasLiked(userUUID, article.ID)
	if err != nil {
		return nil, utils.WrapError(err, "failed to check like status")
	}

	if !liked {
		like := &models.Like{
			UserID:    userUUID,
			ArticleID: article.ID,
		}
		if err := s.engagementRepo.CreateLike(like); err != nil {
			return nil, utils.WrapError(err, "failed to create like")
		}

		// Create notification (don't notify yourself)
		if article.AuthorID != userUUID {
			user, _ := s.userRepo.FindByID(userUUID)
			if user != nil {
				actorName := user.GetFullName()
				message := fmt.Sprintf("%s liked your article", actorName)
				_ = s.CreateNotification(userUUID, article.AuthorID, models.NotificationTypeLike, message, &article.ID)
			}
		}
	}

	count, _ := s.engagementRepo.GetLikesCount(article.ID)
	return &dto.LikeResponse{
		Liked:      true,
		LikesCount: int(count),
	}, nil
}

func (s *engagementService) UnlikeArticle(userID, slug string) (*dto.LikeResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, utils.ErrBadRequest
	}

	article, err := s.articleRepo.FindBySlug(slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound
		}
		return nil, utils.WrapError(err, "failed to find article")
	}

	if err := s.engagementRepo.DeleteLike(userUUID, article.ID); err != nil {
		return nil, utils.WrapError(err, "failed to remove like")
	}

	count, _ := s.engagementRepo.GetLikesCount(article.ID)
	return &dto.LikeResponse{
		Liked:      false,
		LikesCount: int(count),
	}, nil
}

func (s *engagementService) GetLikeStatus(userID, slug string) (*dto.LikeResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, utils.ErrBadRequest
	}

	article, err := s.articleRepo.FindBySlug(slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound
		}
		return nil, utils.WrapError(err, "failed to find article")
	}

	liked, err := s.engagementRepo.HasLiked(userUUID, article.ID)
	if err != nil {
		return nil, utils.WrapError(err, "failed to check like status")
	}

	count, _ := s.engagementRepo.GetLikesCount(article.ID)
	return &dto.LikeResponse{
		Liked:      liked,
		LikesCount: int(count),
	}, nil
}

// --- Bookmarks ---

func (s *engagementService) BookmarkArticle(userID, slug string) (*dto.BookmarkResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, utils.ErrBadRequest
	}

	article, err := s.articleRepo.FindBySlug(slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound
		}
		return nil, utils.WrapError(err, "failed to find article")
	}

	// Idempotent: if already bookmarked, just return
	bookmarked, err := s.engagementRepo.HasBookmarked(userUUID, article.ID)
	if err != nil {
		return nil, utils.WrapError(err, "failed to check bookmark status")
	}

	if !bookmarked {
		bookmark := &models.Bookmark{
			UserID:    userUUID,
			ArticleID: article.ID,
		}
		if err := s.engagementRepo.CreateBookmark(bookmark); err != nil {
			return nil, utils.WrapError(err, "failed to create bookmark")
		}
	}

	return &dto.BookmarkResponse{Bookmarked: true}, nil
}

func (s *engagementService) UnbookmarkArticle(userID, slug string) (*dto.BookmarkResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, utils.ErrBadRequest
	}

	article, err := s.articleRepo.FindBySlug(slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound
		}
		return nil, utils.WrapError(err, "failed to find article")
	}

	if err := s.engagementRepo.DeleteBookmark(userUUID, article.ID); err != nil {
		return nil, utils.WrapError(err, "failed to remove bookmark")
	}

	return &dto.BookmarkResponse{Bookmarked: false}, nil
}

func (s *engagementService) GetBookmarkedArticles(userID string, query *dto.PaginationQuery) ([]dto.ArticleListItemResponse, int64, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, 0, utils.ErrBadRequest
	}

	articles, total, err := s.engagementRepo.GetBookmarkedArticles(userUUID, query.GetPerPage(), query.GetOffset())
	if err != nil {
		return nil, 0, utils.WrapError(err, "failed to get bookmarked articles")
	}

	responses := make([]dto.ArticleListItemResponse, len(articles))
	for i, article := range articles {
		responses[i] = toArticleListItemResponse(&article)
	}

	return responses, total, nil
}

// --- Comments ---

func (s *engagementService) GetComments(slug string, query *dto.PaginationQuery) ([]dto.EngagementCommentResponse, int64, error) {
	article, err := s.articleRepo.FindBySlug(slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, utils.ErrNotFound
		}
		return nil, 0, utils.WrapError(err, "failed to find article")
	}

	filters := repositories.CommentFilters{
		IncludeUnapproved: true, // Show all comments in the new engagement model
		Limit:             query.GetPerPage(),
		Offset:            query.GetOffset(),
	}

	comments, total, err := s.commentRepo.FindByArticle(article.ID, filters)
	if err != nil {
		return nil, 0, utils.WrapError(err, "failed to get comments")
	}

	responses := make([]dto.EngagementCommentResponse, len(comments))
	for i, comment := range comments {
		responses[i] = s.toCommentResponse(&comment)
	}

	return responses, total, nil
}

func (s *engagementService) CreateComment(userID, slug string, req *dto.CreateCommentRequest) (*dto.EngagementCommentResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, utils.ErrBadRequest
	}

	article, err := s.articleRepo.FindBySlug(slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound
		}
		return nil, utils.WrapError(err, "failed to find article")
	}

	comment := &models.Comment{
		ArticleID:  article.ID,
		UserID:     userUUID,
		Content:    req.Content,
		IsApproved: true, // Auto-approve in engagement model
	}

	if req.ParentID != nil {
		parentUUID, err := uuid.Parse(*req.ParentID)
		if err != nil {
			return nil, utils.NewAppError("INVALID_PARENT_ID", "Invalid parent comment ID", 400)
		}
		comment.ParentID = &parentUUID
	}

	if err := s.commentRepo.Create(comment); err != nil {
		return nil, utils.WrapError(err, "failed to create comment")
	}

	// Fetch the created comment with user preloaded
	createdComment, err := s.commentRepo.FindByID(comment.ID)
	if err != nil {
		return nil, utils.WrapError(err, "failed to fetch created comment")
	}

	// Create notification for the article author (don't notify yourself)
	if article.AuthorID != userUUID {
		user, _ := s.userRepo.FindByID(userUUID)
		if user != nil {
			actorName := user.GetFullName()
			message := fmt.Sprintf("%s commented on your article", actorName)
			_ = s.CreateNotification(userUUID, article.AuthorID, models.NotificationTypeComment, message, &article.ID)
		}
	}

	response := s.toCommentResponse(createdComment)
	return &response, nil
}

func (s *engagementService) UpdateComment(userID, slug, commentID string, req *dto.UpdateCommentRequest) (*dto.EngagementCommentResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, utils.ErrBadRequest
	}

	commentUUID, err := uuid.Parse(commentID)
	if err != nil {
		return nil, utils.ErrBadRequest
	}

	comment, err := s.commentRepo.FindByID(commentUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound
		}
		return nil, utils.WrapError(err, "failed to find comment")
	}

	// Only the comment owner can update
	if comment.UserID != userUUID {
		return nil, utils.ErrForbidden
	}

	comment.Content = req.Content
	if err := s.commentRepo.Update(comment); err != nil {
		return nil, utils.WrapError(err, "failed to update comment")
	}

	// Fetch updated comment
	updatedComment, err := s.commentRepo.FindByID(commentUUID)
	if err != nil {
		return nil, utils.WrapError(err, "failed to fetch updated comment")
	}

	response := s.toCommentResponse(updatedComment)
	return &response, nil
}

func (s *engagementService) DeleteComment(userID, slug, commentID string, isAdmin bool) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return utils.ErrBadRequest
	}

	commentUUID, err := uuid.Parse(commentID)
	if err != nil {
		return utils.ErrBadRequest
	}

	comment, err := s.commentRepo.FindByID(commentUUID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrNotFound
		}
		return utils.WrapError(err, "failed to find comment")
	}

	// Only the owner or admin can delete
	if comment.UserID != userUUID && !isAdmin {
		return utils.ErrForbidden
	}

	if err := s.commentRepo.Delete(commentUUID); err != nil {
		return utils.WrapError(err, "failed to delete comment")
	}

	return nil
}

// --- Notifications ---

func (s *engagementService) GetNotifications(userID string, query *dto.PaginationQuery) ([]dto.NotificationResponse, int64, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, 0, utils.ErrBadRequest
	}

	notifications, total, err := s.engagementRepo.GetNotifications(userUUID, query.GetPerPage(), query.GetOffset())
	if err != nil {
		return nil, 0, utils.WrapError(err, "failed to get notifications")
	}

	responses := make([]dto.NotificationResponse, len(notifications))
	for i, notif := range notifications {
		responses[i] = s.toNotificationResponse(&notif)
	}

	return responses, total, nil
}

func (s *engagementService) GetUnreadCount(userID string) (*dto.UnreadCountResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, utils.ErrBadRequest
	}

	count, err := s.engagementRepo.GetUnreadCount(userUUID)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get unread count")
	}

	return &dto.UnreadCountResponse{Count: count}, nil
}

func (s *engagementService) MarkNotificationAsRead(userID, notificationID string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return utils.ErrBadRequest
	}

	notifUUID, err := uuid.Parse(notificationID)
	if err != nil {
		return utils.ErrBadRequest
	}

	if err := s.engagementRepo.MarkAsRead(notifUUID, userUUID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrNotFound
		}
		return utils.WrapError(err, "failed to mark notification as read")
	}

	return nil
}

func (s *engagementService) MarkAllNotificationsAsRead(userID string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return utils.ErrBadRequest
	}

	if err := s.engagementRepo.MarkAllAsRead(userUUID); err != nil {
		return utils.WrapError(err, "failed to mark all notifications as read")
	}

	return nil
}

// --- Notification creation ---

func (s *engagementService) CreateNotification(actorID uuid.UUID, recipientID uuid.UUID, notifType models.NotificationType, message string, articleID *uuid.UUID) error {
	// Don't create notifications when actor is the recipient
	if actorID == recipientID {
		return nil
	}

	notification := &models.Notification{
		UserID:    recipientID,
		ActorID:   actorID,
		Type:      notifType,
		Message:   message,
		ArticleID: articleID,
	}

	return s.engagementRepo.CreateNotification(notification)
}

// --- Article engagement data ---

func (s *engagementService) GetArticleEngagement(articleID uuid.UUID, userID string) (likesCount int, commentsCount int, userLiked bool, userBookmarked bool) {
	lCount, _ := s.engagementRepo.GetLikesCount(articleID)
	cCount, _ := s.engagementRepo.GetCommentsCount(articleID)
	likesCount = int(lCount)
	commentsCount = int(cCount)

	if userID != "" {
		userUUID, err := uuid.Parse(userID)
		if err == nil {
			userLiked, _ = s.engagementRepo.HasLiked(userUUID, articleID)
			userBookmarked, _ = s.engagementRepo.HasBookmarked(userUUID, articleID)
		}
	}

	return
}

// --- Helper methods ---

func (s *engagementService) toCommentResponse(comment *models.Comment) dto.EngagementCommentResponse {
	response := dto.EngagementCommentResponse{
		ID:         comment.ID.String(),
		ArticleID:  comment.ArticleID.String(),
		Content:    comment.Content,
		LikesCount: comment.LikesCount,
		CreatedAt:  comment.CreatedAt,
		UpdatedAt:  comment.UpdatedAt,
	}

	if comment.ParentID != nil {
		parentStr := comment.ParentID.String()
		response.ParentID = &parentStr
	}

	if comment.User != nil {
		response.User = dto.PublicUserResponse{
			ID:              comment.User.ID.String(),
			Username:        comment.User.Username,
			FirstName:       comment.User.FirstName,
			LastName:        comment.User.LastName,
			Bio:             comment.User.Bio,
			ProfileImageURL: comment.User.ProfileImageURL,
		}
	}

	// Convert replies
	response.Replies = make([]dto.EngagementCommentResponse, len(comment.Replies))
	for i, reply := range comment.Replies {
		response.Replies[i] = s.toCommentResponse(&reply)
	}

	return response
}

func (s *engagementService) toNotificationResponse(notif *models.Notification) dto.NotificationResponse {
	response := dto.NotificationResponse{
		ID:        notif.ID.String(),
		Type:      string(notif.Type),
		Message:   notif.Message,
		Read:      notif.Read,
		CreatedAt: notif.CreatedAt,
	}

	if notif.Actor != nil {
		response.Actor = dto.PublicUserResponse{
			ID:              notif.Actor.ID.String(),
			Username:        notif.Actor.Username,
			FirstName:       notif.Actor.FirstName,
			LastName:        notif.Actor.LastName,
			Bio:             notif.Actor.Bio,
			ProfileImageURL: notif.Actor.ProfileImageURL,
		}
	}

	if notif.Article != nil {
		response.Article = &dto.NotificationArticleResponse{
			ID:    notif.Article.ID.String(),
			Slug:  notif.Article.Slug,
			Title: notif.Article.Title,
		}
	}

	return response
}
