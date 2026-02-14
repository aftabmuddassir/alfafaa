package dto

import "time"

// --- Like DTOs ---

// LikeResponse represents the response for like operations
type LikeResponse struct {
	Liked      bool `json:"liked"`
	LikesCount int  `json:"likes_count"`
}

// --- Bookmark DTOs ---

// BookmarkResponse represents the response for bookmark operations
type BookmarkResponse struct {
	Bookmarked bool `json:"bookmarked"`
}

// --- Notification DTOs ---

// NotificationArticleResponse represents a minimal article in notification responses
type NotificationArticleResponse struct {
	ID    string `json:"id"`
	Slug  string `json:"slug"`
	Title string `json:"title"`
}

// NotificationResponse represents a notification in API responses
type NotificationResponse struct {
	ID        string                       `json:"id"`
	Type      string                       `json:"type"`
	Message   string                       `json:"message"`
	Actor     PublicUserResponse           `json:"actor"`
	Article   *NotificationArticleResponse `json:"article"`
	Read      bool                         `json:"read"`
	CreatedAt time.Time                    `json:"created_at"`
}

// UnreadCountResponse represents the unread notification count
type UnreadCountResponse struct {
	Count int64 `json:"count"`
}

// --- Updated Comment DTOs ---

// EngagementCommentResponse represents a comment in the new engagement-aware format
type EngagementCommentResponse struct {
	ID         string                      `json:"id"`
	ArticleID  string                      `json:"article_id"`
	User       PublicUserResponse          `json:"user"`
	Content    string                      `json:"content"`
	ParentID   *string                     `json:"parent_id"`
	Replies    []EngagementCommentResponse `json:"replies"`
	LikesCount int                        `json:"likes_count"`
	CreatedAt  time.Time                   `json:"created_at"`
	UpdatedAt  time.Time                   `json:"updated_at"`
}
