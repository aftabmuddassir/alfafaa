package repositories

import (
	"testing"

	"github.com/alfafaa/alfafaa-blog/internal/models"
	"github.com/alfafaa/alfafaa-blog/tests/helpers"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type UserRepositoryTestSuite struct {
	suite.Suite
	db   *gorm.DB
	repo UserRepository
}

func (suite *UserRepositoryTestSuite) SetupSuite() {
	suite.db = helpers.SetupTestDB()
	suite.repo = NewUserRepository(suite.db)
}

func (suite *UserRepositoryTestSuite) SetupTest() {
	helpers.CleanupTestDB(suite.db)
}

func TestUserRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(UserRepositoryTestSuite))
}

// Create Tests

func (suite *UserRepositoryTestSuite) TestCreate_Success() {
	user := &models.User{
		ID:           uuid.New(),
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		FirstName:    "Test",
		LastName:     "User",
		Role:         models.RoleReader,
		IsActive:     true,
	}

	err := suite.repo.Create(user)

	assert.NoError(suite.T(), err)
	assert.NotEqual(suite.T(), uuid.Nil, user.ID)
}

func (suite *UserRepositoryTestSuite) TestCreate_DuplicateEmail() {
	user1 := &models.User{
		ID:           uuid.New(),
		Username:     "user1",
		Email:        "same@example.com",
		PasswordHash: "hashedpassword",
		Role:         models.RoleReader,
		IsActive:     true,
	}
	user2 := &models.User{
		ID:           uuid.New(),
		Username:     "user2",
		Email:        "same@example.com",
		PasswordHash: "hashedpassword",
		Role:         models.RoleReader,
		IsActive:     true,
	}

	err := suite.repo.Create(user1)
	assert.NoError(suite.T(), err)

	err = suite.repo.Create(user2)
	assert.Error(suite.T(), err) // Should fail due to unique constraint
}

// FindByID Tests

func (suite *UserRepositoryTestSuite) TestFindByID_Success() {
	user := &models.User{
		ID:           uuid.New(),
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		Role:         models.RoleReader,
		IsActive:     true,
	}
	suite.repo.Create(user)

	found, err := suite.repo.FindByID(user.ID)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), found)
	assert.Equal(suite.T(), user.Email, found.Email)
	assert.Equal(suite.T(), user.Username, found.Username)
}

func (suite *UserRepositoryTestSuite) TestFindByID_NotFound() {
	found, err := suite.repo.FindByID(uuid.New())

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), found)
	assert.Equal(suite.T(), gorm.ErrRecordNotFound, err)
}

// FindByEmail Tests

func (suite *UserRepositoryTestSuite) TestFindByEmail_Success() {
	user := &models.User{
		ID:           uuid.New(),
		Username:     "testuser",
		Email:        "find@example.com",
		PasswordHash: "hashedpassword",
		Role:         models.RoleReader,
		IsActive:     true,
	}
	suite.repo.Create(user)

	found, err := suite.repo.FindByEmail("find@example.com")

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), found)
	assert.Equal(suite.T(), user.Username, found.Username)
}

func (suite *UserRepositoryTestSuite) TestFindByEmail_NotFound() {
	found, err := suite.repo.FindByEmail("nonexistent@example.com")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), found)
}

// FindByUsername Tests

func (suite *UserRepositoryTestSuite) TestFindByUsername_Success() {
	user := &models.User{
		ID:           uuid.New(),
		Username:     "uniqueuser",
		Email:        "unique@example.com",
		PasswordHash: "hashedpassword",
		Role:         models.RoleReader,
		IsActive:     true,
	}
	suite.repo.Create(user)

	found, err := suite.repo.FindByUsername("uniqueuser")

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), found)
	assert.Equal(suite.T(), user.Email, found.Email)
}

func (suite *UserRepositoryTestSuite) TestFindByUsername_NotFound() {
	found, err := suite.repo.FindByUsername("nonexistentuser")

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), found)
}

// FindAll Tests

func (suite *UserRepositoryTestSuite) TestFindAll_Success() {
	// Create multiple users
	users := []*models.User{
		{
			ID:           uuid.New(),
			Username:     "user1",
			Email:        "user1@example.com",
			PasswordHash: "hashedpassword",
			Role:         models.RoleReader,
			IsActive:     true,
		},
		{
			ID:           uuid.New(),
			Username:     "user2",
			Email:        "user2@example.com",
			PasswordHash: "hashedpassword",
			Role:         models.RoleAuthor,
			IsActive:     true,
		},
		{
			ID:           uuid.New(),
			Username:     "user3",
			Email:        "user3@example.com",
			PasswordHash: "hashedpassword",
			Role:         models.RoleReader,
			IsActive:     false,
		},
	}

	for _, u := range users {
		suite.repo.Create(u)
	}

	filters := UserFilters{Limit: 10}
	result, total, err := suite.repo.FindAll(filters)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 3)
	assert.Equal(suite.T(), int64(3), total)
}

func (suite *UserRepositoryTestSuite) TestFindAll_WithRoleFilter() {
	users := []*models.User{
		{
			ID:           uuid.New(),
			Username:     "reader1",
			Email:        "reader1@example.com",
			PasswordHash: "hashedpassword",
			Role:         models.RoleReader,
			IsActive:     true,
		},
		{
			ID:           uuid.New(),
			Username:     "author1",
			Email:        "author1@example.com",
			PasswordHash: "hashedpassword",
			Role:         models.RoleAuthor,
			IsActive:     true,
		},
	}

	for _, u := range users {
		suite.repo.Create(u)
	}

	filters := UserFilters{Role: string(models.RoleAuthor), Limit: 10}
	result, total, err := suite.repo.FindAll(filters)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 1)
	assert.Equal(suite.T(), int64(1), total)
	assert.Equal(suite.T(), models.RoleAuthor, result[0].Role)
}

func (suite *UserRepositoryTestSuite) TestFindAll_WithActiveFilter() {
	// Create active users
	activeUser := &models.User{
		ID:           uuid.New(),
		Username:     "active1",
		Email:        "active1@example.com",
		PasswordHash: "hashedpassword",
		Role:         models.RoleReader,
		IsActive:     true,
	}
	suite.repo.Create(activeUser)

	// Create inactive user by updating after creation (SQLite bool workaround)
	inactiveUser := &models.User{
		ID:           uuid.New(),
		Username:     "inactive1",
		Email:        "inactive1@example.com",
		PasswordHash: "hashedpassword",
		Role:         models.RoleReader,
		IsActive:     true,
	}
	suite.repo.Create(inactiveUser)
	// Explicitly set to inactive using raw update
	suite.db.Model(&models.User{}).Where("id = ?", inactiveUser.ID).Update("is_active", false)

	isActive := true
	filters := UserFilters{IsActive: &isActive, Limit: 10}
	result, total, err := suite.repo.FindAll(filters)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 1)
	assert.Equal(suite.T(), int64(1), total)
}

func (suite *UserRepositoryTestSuite) TestFindAll_WithPagination() {
	for i := 0; i < 5; i++ {
		suite.repo.Create(&models.User{
			ID:           uuid.New(),
			Username:     "paguser" + string(rune('0'+i)),
			Email:        "pag" + string(rune('0'+i)) + "@example.com",
			PasswordHash: "hashedpassword",
			Role:         models.RoleReader,
			IsActive:     true,
		})
	}

	filters := UserFilters{Limit: 2, Offset: 2}
	result, total, err := suite.repo.FindAll(filters)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 2)
	assert.Equal(suite.T(), int64(5), total)
}

// Update Tests

func (suite *UserRepositoryTestSuite) TestUpdate_Success() {
	user := &models.User{
		ID:           uuid.New(),
		Username:     "updateuser",
		Email:        "update@example.com",
		PasswordHash: "hashedpassword",
		FirstName:    "Original",
		Role:         models.RoleReader,
		IsActive:     true,
	}
	suite.repo.Create(user)

	user.FirstName = "Updated"
	user.LastName = "Name"
	err := suite.repo.Update(user)

	assert.NoError(suite.T(), err)

	found, _ := suite.repo.FindByID(user.ID)
	assert.Equal(suite.T(), "Updated", found.FirstName)
	assert.Equal(suite.T(), "Name", found.LastName)
}

// Delete Tests

func (suite *UserRepositoryTestSuite) TestDelete_Success() {
	user := &models.User{
		ID:           uuid.New(),
		Username:     "deleteuser",
		Email:        "delete@example.com",
		PasswordHash: "hashedpassword",
		Role:         models.RoleReader,
		IsActive:     true,
	}
	suite.repo.Create(user)

	err := suite.repo.Delete(user.ID)

	assert.NoError(suite.T(), err)

	// Should not find deleted user (soft delete)
	found, err := suite.repo.FindByID(user.ID)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), found)
}

// ExistsByEmail Tests

func (suite *UserRepositoryTestSuite) TestExistsByEmail_Exists() {
	user := &models.User{
		ID:           uuid.New(),
		Username:     "existsuser",
		Email:        "exists@example.com",
		PasswordHash: "hashedpassword",
		Role:         models.RoleReader,
		IsActive:     true,
	}
	suite.repo.Create(user)

	exists, err := suite.repo.ExistsByEmail("exists@example.com")

	assert.NoError(suite.T(), err)
	assert.True(suite.T(), exists)
}

func (suite *UserRepositoryTestSuite) TestExistsByEmail_NotExists() {
	exists, err := suite.repo.ExistsByEmail("notexists@example.com")

	assert.NoError(suite.T(), err)
	assert.False(suite.T(), exists)
}

// ExistsByUsername Tests

func (suite *UserRepositoryTestSuite) TestExistsByUsername_Exists() {
	user := &models.User{
		ID:           uuid.New(),
		Username:     "existinguser",
		Email:        "existing@example.com",
		PasswordHash: "hashedpassword",
		Role:         models.RoleReader,
		IsActive:     true,
	}
	suite.repo.Create(user)

	exists, err := suite.repo.ExistsByUsername("existinguser")

	assert.NoError(suite.T(), err)
	assert.True(suite.T(), exists)
}

func (suite *UserRepositoryTestSuite) TestExistsByUsername_NotExists() {
	exists, err := suite.repo.ExistsByUsername("nonexistinguser")

	assert.NoError(suite.T(), err)
	assert.False(suite.T(), exists)
}
