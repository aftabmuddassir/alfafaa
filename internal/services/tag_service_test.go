package services

import (
	"testing"
	"time"

	"github.com/alfafaa/alfafaa-blog/internal/dto"
	"github.com/alfafaa/alfafaa-blog/internal/models"
	"github.com/alfafaa/alfafaa-blog/internal/repositories"
	"github.com/alfafaa/alfafaa-blog/internal/utils"
	"github.com/alfafaa/alfafaa-blog/tests/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type TagServiceTestSuite struct {
	suite.Suite
	tagRepo     *mocks.MockTagRepository
	articleRepo *mocks.MockArticleRepository
	service     TagService
}

func (suite *TagServiceTestSuite) SetupTest() {
	suite.tagRepo = new(mocks.MockTagRepository)
	suite.articleRepo = new(mocks.MockArticleRepository)
	suite.service = NewTagService(suite.tagRepo, suite.articleRepo)
}

func TestTagServiceTestSuite(t *testing.T) {
	suite.Run(t, new(TagServiceTestSuite))
}

// CreateTag Tests

func (suite *TagServiceTestSuite) TestCreateTag_Success() {
	req := &dto.CreateTagRequest{
		Name:        "Go Programming",
		Description: "Articles about Go",
	}

	suite.tagRepo.On("ExistsBySlug", "go-programming").Return(false, nil)
	suite.tagRepo.On("Create", mock.AnythingOfType("*models.Tag")).Return(nil)

	result, err := suite.service.CreateTag(req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "Go Programming", result.Name)
	assert.Equal(suite.T(), "go-programming", result.Slug)
	assert.Equal(suite.T(), 0, result.UsageCount)
	suite.tagRepo.AssertExpectations(suite.T())
}

func (suite *TagServiceTestSuite) TestCreateTag_DuplicateSlug_GeneratesUnique() {
	req := &dto.CreateTagRequest{
		Name: "Go Programming",
	}

	// First slug exists, second one doesn't
	suite.tagRepo.On("ExistsBySlug", "go-programming").Return(true, nil).Once()
	suite.tagRepo.On("ExistsBySlug", "go-programming-1").Return(false, nil).Once()
	suite.tagRepo.On("Create", mock.AnythingOfType("*models.Tag")).Return(nil)

	result, err := suite.service.CreateTag(req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "go-programming-1", result.Slug)
	suite.tagRepo.AssertExpectations(suite.T())
}

func (suite *TagServiceTestSuite) TestCreateTag_MultipleDuplicates() {
	req := &dto.CreateTagRequest{
		Name: "JavaScript",
	}

	// First two slugs exist
	suite.tagRepo.On("ExistsBySlug", "javascript").Return(true, nil).Once()
	suite.tagRepo.On("ExistsBySlug", "javascript-1").Return(true, nil).Once()
	suite.tagRepo.On("ExistsBySlug", "javascript-2").Return(false, nil).Once()
	suite.tagRepo.On("Create", mock.AnythingOfType("*models.Tag")).Return(nil)

	result, err := suite.service.CreateTag(req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "javascript-2", result.Slug)
	suite.tagRepo.AssertExpectations(suite.T())
}

// GetTag Tests

func (suite *TagServiceTestSuite) TestGetTag_Success() {
	tagID := uuid.New()
	tag := &models.Tag{
		ID:          tagID,
		Name:        "Go Programming",
		Slug:        "go-programming",
		Description: "Articles about Go",
		UsageCount:  10,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	suite.tagRepo.On("FindBySlug", "go-programming").Return(tag, nil)

	result, err := suite.service.GetTag("go-programming")

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "Go Programming", result.Name)
	assert.Equal(suite.T(), 10, result.UsageCount)
	suite.tagRepo.AssertExpectations(suite.T())
}

func (suite *TagServiceTestSuite) TestGetTag_NotFound() {
	suite.tagRepo.On("FindBySlug", "nonexistent").Return(nil, gorm.ErrRecordNotFound)

	result, err := suite.service.GetTag("nonexistent")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), utils.ErrNotFound, err)
	suite.tagRepo.AssertExpectations(suite.T())
}

// GetTags Tests

func (suite *TagServiceTestSuite) TestGetTags_Success() {
	tags := []models.Tag{
		{
			ID:         uuid.New(),
			Name:       "Go Programming",
			Slug:       "go-programming",
			UsageCount: 10,
		},
		{
			ID:         uuid.New(),
			Name:       "JavaScript",
			Slug:       "javascript",
			UsageCount: 20,
		},
	}

	query := &dto.TagListQuery{
		PaginationQuery: dto.PaginationQuery{Page: 1, PerPage: 10},
	}

	suite.tagRepo.On("FindAll", mock.AnythingOfType("repositories.TagFilters")).
		Return(tags, int64(2), nil)

	result, total, err := suite.service.GetTags(query)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 2)
	assert.Equal(suite.T(), int64(2), total)
	suite.tagRepo.AssertExpectations(suite.T())
}

func (suite *TagServiceTestSuite) TestGetTags_WithSearch() {
	tags := []models.Tag{
		{
			ID:         uuid.New(),
			Name:       "Go Programming",
			Slug:       "go-programming",
			UsageCount: 10,
		},
	}

	query := &dto.TagListQuery{
		PaginationQuery: dto.PaginationQuery{Page: 1, PerPage: 10},
		Search:          "go",
	}

	suite.tagRepo.On("FindAll", mock.MatchedBy(func(filters repositories.TagFilters) bool {
		return filters.Search == "go"
	})).Return(tags, int64(1), nil)

	result, total, err := suite.service.GetTags(query)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 1)
	assert.Equal(suite.T(), int64(1), total)
	suite.tagRepo.AssertExpectations(suite.T())
}

func (suite *TagServiceTestSuite) TestGetTags_Empty() {
	query := &dto.TagListQuery{
		PaginationQuery: dto.PaginationQuery{Page: 1, PerPage: 10},
	}

	suite.tagRepo.On("FindAll", mock.AnythingOfType("repositories.TagFilters")).
		Return([]models.Tag{}, int64(0), nil)

	result, total, err := suite.service.GetTags(query)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 0)
	assert.Equal(suite.T(), int64(0), total)
	suite.tagRepo.AssertExpectations(suite.T())
}

// GetPopularTags Tests

func (suite *TagServiceTestSuite) TestGetPopularTags_Success() {
	tags := []models.Tag{
		{
			ID:         uuid.New(),
			Name:       "JavaScript",
			Slug:       "javascript",
			UsageCount: 100,
		},
		{
			ID:         uuid.New(),
			Name:       "Go Programming",
			Slug:       "go-programming",
			UsageCount: 50,
		},
	}

	suite.tagRepo.On("FindPopular", 10).Return(tags, nil)

	result, err := suite.service.GetPopularTags(10)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 2)
	assert.Equal(suite.T(), 100, result[0].UsageCount)
	suite.tagRepo.AssertExpectations(suite.T())
}

func (suite *TagServiceTestSuite) TestGetPopularTags_DefaultLimit() {
	tags := []models.Tag{}

	// When limit <= 0, defaults to 10
	suite.tagRepo.On("FindPopular", 10).Return(tags, nil)

	result, err := suite.service.GetPopularTags(0)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	suite.tagRepo.AssertExpectations(suite.T())
}

func (suite *TagServiceTestSuite) TestGetPopularTags_MaxLimit() {
	tags := []models.Tag{}

	// When limit > 50, caps to 50
	suite.tagRepo.On("FindPopular", 50).Return(tags, nil)

	result, err := suite.service.GetPopularTags(100)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	suite.tagRepo.AssertExpectations(suite.T())
}

func (suite *TagServiceTestSuite) TestGetPopularTags_NegativeLimit() {
	tags := []models.Tag{}

	// When limit < 0, defaults to 10
	suite.tagRepo.On("FindPopular", 10).Return(tags, nil)

	result, err := suite.service.GetPopularTags(-5)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	suite.tagRepo.AssertExpectations(suite.T())
}

// UpdateTag Tests

func (suite *TagServiceTestSuite) TestUpdateTag_Success() {
	tagID := uuid.New()
	tag := &models.Tag{
		ID:         tagID,
		Name:       "Go Programming",
		Slug:       "go-programming",
		UsageCount: 10,
	}

	newName := "Golang"
	newDescription := "Updated description"
	req := &dto.UpdateTagRequest{
		Name:        &newName,
		Description: &newDescription,
	}

	suite.tagRepo.On("FindByID", tagID).Return(tag, nil)
	suite.tagRepo.On("ExistsBySlug", "golang").Return(false, nil)
	suite.tagRepo.On("Update", mock.AnythingOfType("*models.Tag")).Return(nil)

	result, err := suite.service.UpdateTag(tagID.String(), req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "Golang", result.Name)
	assert.Equal(suite.T(), "golang", result.Slug)
	suite.tagRepo.AssertExpectations(suite.T())
}

func (suite *TagServiceTestSuite) TestUpdateTag_InvalidUUID() {
	req := &dto.UpdateTagRequest{}

	result, err := suite.service.UpdateTag("not-a-uuid", req)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), utils.ErrBadRequest, err)
}

func (suite *TagServiceTestSuite) TestUpdateTag_NotFound() {
	tagID := uuid.New()
	req := &dto.UpdateTagRequest{}

	suite.tagRepo.On("FindByID", tagID).Return(nil, gorm.ErrRecordNotFound)

	result, err := suite.service.UpdateTag(tagID.String(), req)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), utils.ErrNotFound, err)
	suite.tagRepo.AssertExpectations(suite.T())
}

func (suite *TagServiceTestSuite) TestUpdateTag_SlugConflict_KeepsOldSlug() {
	tagID := uuid.New()
	tag := &models.Tag{
		ID:   tagID,
		Name: "Go Programming",
		Slug: "go-programming",
	}

	newName := "JavaScript" // Slug for this already exists
	req := &dto.UpdateTagRequest{
		Name: &newName,
	}

	suite.tagRepo.On("FindByID", tagID).Return(tag, nil)
	suite.tagRepo.On("ExistsBySlug", "javascript").Return(true, nil)
	suite.tagRepo.On("Update", mock.AnythingOfType("*models.Tag")).Return(nil)

	result, err := suite.service.UpdateTag(tagID.String(), req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	// Name is updated but slug remains the same due to conflict
	assert.Equal(suite.T(), "JavaScript", result.Name)
	assert.Equal(suite.T(), "go-programming", result.Slug)
	suite.tagRepo.AssertExpectations(suite.T())
}

func (suite *TagServiceTestSuite) TestUpdateTag_OnlyDescription() {
	tagID := uuid.New()
	tag := &models.Tag{
		ID:          tagID,
		Name:        "Go Programming",
		Slug:        "go-programming",
		Description: "Old description",
	}

	newDescription := "New description"
	req := &dto.UpdateTagRequest{
		Description: &newDescription,
	}

	suite.tagRepo.On("FindByID", tagID).Return(tag, nil)
	suite.tagRepo.On("Update", mock.AnythingOfType("*models.Tag")).Return(nil)

	result, err := suite.service.UpdateTag(tagID.String(), req)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "New description", result.Description)
	suite.tagRepo.AssertExpectations(suite.T())
}

// DeleteTag Tests

func (suite *TagServiceTestSuite) TestDeleteTag_Success() {
	tagID := uuid.New()
	tag := &models.Tag{
		ID:         tagID,
		Name:       "Unused Tag",
		Slug:       "unused-tag",
		UsageCount: 0,
	}

	suite.tagRepo.On("FindByID", tagID).Return(tag, nil)
	suite.tagRepo.On("Delete", tagID).Return(nil)

	err := suite.service.DeleteTag(tagID.String())

	assert.NoError(suite.T(), err)
	suite.tagRepo.AssertExpectations(suite.T())
}

func (suite *TagServiceTestSuite) TestDeleteTag_InvalidUUID() {
	err := suite.service.DeleteTag("not-a-uuid")

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), utils.ErrBadRequest, err)
}

func (suite *TagServiceTestSuite) TestDeleteTag_NotFound() {
	tagID := uuid.New()

	suite.tagRepo.On("FindByID", tagID).Return(nil, gorm.ErrRecordNotFound)

	err := suite.service.DeleteTag(tagID.String())

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), utils.ErrNotFound, err)
	suite.tagRepo.AssertExpectations(suite.T())
}

func (suite *TagServiceTestSuite) TestDeleteTag_InUse() {
	tagID := uuid.New()
	tag := &models.Tag{
		ID:         tagID,
		Name:       "Popular Tag",
		Slug:       "popular-tag",
		UsageCount: 10,
	}

	suite.tagRepo.On("FindByID", tagID).Return(tag, nil)

	err := suite.service.DeleteTag(tagID.String())

	assert.Error(suite.T(), err)
	appErr, ok := utils.IsAppError(err)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), "TAG_IN_USE", appErr.Code)
	suite.tagRepo.AssertExpectations(suite.T())
}

// GetTagArticles Tests

func (suite *TagServiceTestSuite) TestGetTagArticles_Success() {
	tagID := uuid.New()
	authorID := uuid.New()
	tag := &models.Tag{
		ID:         tagID,
		Name:       "Go Programming",
		Slug:       "go-programming",
		UsageCount: 2,
	}

	articles := []models.Article{
		{
			ID:       uuid.New(),
			Title:    "Go Basics",
			Slug:     "go-basics",
			Status:   models.StatusPublished,
			AuthorID: authorID,
			Author: &models.User{
				ID:       authorID,
				Username: "author1",
			},
		},
		{
			ID:       uuid.New(),
			Title:    "Advanced Go",
			Slug:     "advanced-go",
			Status:   models.StatusPublished,
			AuthorID: authorID,
			Author: &models.User{
				ID:       authorID,
				Username: "author1",
			},
		},
	}

	query := &dto.PaginationQuery{Page: 1, PerPage: 10}

	suite.tagRepo.On("FindBySlug", "go-programming").Return(tag, nil)
	suite.articleRepo.On("FindByTag", tagID, mock.AnythingOfType("repositories.ArticleFilters")).
		Return(articles, int64(2), nil)

	result, total, err := suite.service.GetTagArticles("go-programming", query)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 2)
	assert.Equal(suite.T(), int64(2), total)
	suite.tagRepo.AssertExpectations(suite.T())
	suite.articleRepo.AssertExpectations(suite.T())
}

func (suite *TagServiceTestSuite) TestGetTagArticles_TagNotFound() {
	query := &dto.PaginationQuery{Page: 1, PerPage: 10}

	suite.tagRepo.On("FindBySlug", "nonexistent").Return(nil, gorm.ErrRecordNotFound)

	result, total, err := suite.service.GetTagArticles("nonexistent", query)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.Equal(suite.T(), int64(0), total)
	assert.Equal(suite.T(), utils.ErrNotFound, err)
	suite.tagRepo.AssertExpectations(suite.T())
}

func (suite *TagServiceTestSuite) TestGetTagArticles_NoArticles() {
	tagID := uuid.New()
	tag := &models.Tag{
		ID:         tagID,
		Name:       "New Tag",
		Slug:       "new-tag",
		UsageCount: 0,
	}

	query := &dto.PaginationQuery{Page: 1, PerPage: 10}

	suite.tagRepo.On("FindBySlug", "new-tag").Return(tag, nil)
	suite.articleRepo.On("FindByTag", tagID, mock.AnythingOfType("repositories.ArticleFilters")).
		Return([]models.Article{}, int64(0), nil)

	result, total, err := suite.service.GetTagArticles("new-tag", query)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 0)
	assert.Equal(suite.T(), int64(0), total)
	suite.tagRepo.AssertExpectations(suite.T())
	suite.articleRepo.AssertExpectations(suite.T())
}
