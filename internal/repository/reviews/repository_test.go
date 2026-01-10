package reviews_test

import (
	"context"
	"product_review_hub/internal/models"
	"product_review_hub/internal/repository/reviews"
	"product_review_hub/internal/testutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepository_Create(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	repo := reviews.NewRepository(tdb.DB)
	ctx := context.Background()

	t.Run("create review with all fields", func(t *testing.T) {
		tdb.Cleanup(t)

		productID := tdb.CreateTestProduct(t, "Test Product", nil, 99.99)

		tx, err := repo.BeginTx(ctx)
		require.NoError(t, err)
		defer tx.Rollback()

		comment := "Great product!"
		review, err := repo.Create(ctx, tx, models.CreateReviewParams{
			ProductID: productID,
			FirstName: "John",
			LastName:  "Doe",
			Rating:    5,
			Comment:   &comment,
		})
		require.NoError(t, err)

		assert.NotZero(t, review.ID)
		assert.Equal(t, productID, review.ProductID)
		assert.Equal(t, "John", review.FirstName)
		assert.Equal(t, "Doe", review.LastName)
		assert.Equal(t, 5, review.Rating)
		require.NotNil(t, review.Comment)
		assert.Equal(t, comment, *review.Comment)

		require.NoError(t, repo.CommitTx(tx))
	})

	t.Run("create review without comment", func(t *testing.T) {
		tdb.Cleanup(t)

		productID := tdb.CreateTestProduct(t, "Test Product", nil, 99.99)

		tx, err := repo.BeginTx(ctx)
		require.NoError(t, err)
		defer tx.Rollback()

		review, err := repo.Create(ctx, tx, models.CreateReviewParams{
			ProductID: productID,
			FirstName: "Jane",
			LastName:  "Doe",
			Rating:    4,
			Comment:   nil,
		})
		require.NoError(t, err)
		assert.Nil(t, review.Comment)

		require.NoError(t, repo.CommitTx(tx))
	})

	t.Run("create review with various ratings", func(t *testing.T) {
		tdb.Cleanup(t)

		productID := tdb.CreateTestProduct(t, "Test Product", nil, 99.99)

		tx, err := repo.BeginTx(ctx)
		require.NoError(t, err)
		defer tx.Rollback()

		for rating := 1; rating <= 5; rating++ {
			review, err := repo.Create(ctx, tx, models.CreateReviewParams{
				ProductID: productID,
				FirstName: "User",
				LastName:  "Test",
				Rating:    rating,
			})
			require.NoError(t, err, "rating %d", rating)
			assert.Equal(t, rating, review.Rating)
		}

		require.NoError(t, repo.CommitTx(tx))
	})
}

func TestRepository_GetByID(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	repo := reviews.NewRepository(tdb.DB)
	ctx := context.Background()

	t.Run("get existing review", func(t *testing.T) {
		tdb.Cleanup(t)

		productID := tdb.CreateTestProduct(t, "Test Product", nil, 99.99)
		comment := "Test comment"
		reviewID := tdb.CreateTestReview(t, productID, "John", "Doe", 5, &comment)

		tx, err := repo.BeginTx(ctx)
		require.NoError(t, err)
		defer tx.Rollback()

		review, err := repo.GetByID(ctx, tx, reviewID)
		require.NoError(t, err)

		assert.Equal(t, reviewID, review.ID)
		assert.Equal(t, productID, review.ProductID)
		assert.Equal(t, "John", review.FirstName)
		assert.Equal(t, "Doe", review.LastName)

		require.NoError(t, repo.CommitTx(tx))
	})

	t.Run("get non-existing review", func(t *testing.T) {
		tdb.Cleanup(t)

		tx, err := repo.BeginTx(ctx)
		require.NoError(t, err)
		defer tx.Rollback()

		_, err = repo.GetByID(ctx, tx, 999999)
		assert.Error(t, err)
		assert.ErrorIs(t, err, reviews.ErrNotFound)
	})
}

func TestRepository_GetByIDAndProductID(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	repo := reviews.NewRepository(tdb.DB)
	ctx := context.Background()

	t.Run("get review with matching product", func(t *testing.T) {
		tdb.Cleanup(t)

		productID := tdb.CreateTestProduct(t, "Test Product", nil, 99.99)
		reviewID := tdb.CreateTestReview(t, productID, "John", "Doe", 5, nil)

		tx, err := repo.BeginTx(ctx)
		require.NoError(t, err)
		defer tx.Rollback()

		review, err := repo.GetByIDAndProductID(ctx, tx, reviewID, productID)
		require.NoError(t, err)
		assert.Equal(t, reviewID, review.ID)

		require.NoError(t, repo.CommitTx(tx))
	})

	t.Run("get review with non-matching product", func(t *testing.T) {
		tdb.Cleanup(t)

		productID1 := tdb.CreateTestProduct(t, "Product 1", nil, 99.99)
		productID2 := tdb.CreateTestProduct(t, "Product 2", nil, 49.99)
		reviewID := tdb.CreateTestReview(t, productID1, "John", "Doe", 5, nil)

		tx, err := repo.BeginTx(ctx)
		require.NoError(t, err)
		defer tx.Rollback()

		_, err = repo.GetByIDAndProductID(ctx, tx, reviewID, productID2)
		assert.Error(t, err)
		assert.ErrorIs(t, err, reviews.ErrNotFound)
	})
}

func TestRepository_ListByProductID(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	repo := reviews.NewRepository(tdb.DB)
	ctx := context.Background()

	t.Run("list empty reviews", func(t *testing.T) {
		tdb.Cleanup(t)

		productID := tdb.CreateTestProduct(t, "Test Product", nil, 99.99)

		tx, err := repo.BeginTx(ctx)
		require.NoError(t, err)
		defer tx.Rollback()

		reviewsList, err := repo.ListByProductID(ctx, tx, models.ListReviewsParams{
			ProductID: productID,
			Limit:     10,
			Offset:    0,
		})
		require.NoError(t, err)
		assert.Empty(t, reviewsList)

		require.NoError(t, repo.CommitTx(tx))
	})

	t.Run("list reviews with pagination", func(t *testing.T) {
		tdb.Cleanup(t)

		productID := tdb.CreateTestProduct(t, "Test Product", nil, 99.99)

		// Create 5 reviews
		for i := 0; i < 5; i++ {
			tdb.CreateTestReview(t, productID, "User", "Test", i%5+1, nil)
		}

		tx, err := repo.BeginTx(ctx)
		require.NoError(t, err)
		defer tx.Rollback()

		// Get first page
		reviewsList, err := repo.ListByProductID(ctx, tx, models.ListReviewsParams{
			ProductID: productID,
			Limit:     2,
			Offset:    0,
		})
		require.NoError(t, err)
		assert.Len(t, reviewsList, 2)

		// Get second page
		reviewsList, err = repo.ListByProductID(ctx, tx, models.ListReviewsParams{
			ProductID: productID,
			Limit:     2,
			Offset:    2,
		})
		require.NoError(t, err)
		assert.Len(t, reviewsList, 2)

		require.NoError(t, repo.CommitTx(tx))
	})

	t.Run("list reviews only for specific product", func(t *testing.T) {
		tdb.Cleanup(t)

		productID1 := tdb.CreateTestProduct(t, "Product 1", nil, 99.99)
		productID2 := tdb.CreateTestProduct(t, "Product 2", nil, 49.99)

		tdb.CreateTestReview(t, productID1, "User1", "First", 5, nil)
		tdb.CreateTestReview(t, productID1, "User2", "Second", 4, nil)
		tdb.CreateTestReview(t, productID2, "User3", "Third", 3, nil)

		tx, err := repo.BeginTx(ctx)
		require.NoError(t, err)
		defer tx.Rollback()

		reviewsList, err := repo.ListByProductID(ctx, tx, models.ListReviewsParams{
			ProductID: productID1,
			Limit:     10,
			Offset:    0,
		})
		require.NoError(t, err)
		assert.Len(t, reviewsList, 2)

		for _, r := range reviewsList {
			assert.Equal(t, productID1, r.ProductID)
		}

		require.NoError(t, repo.CommitTx(tx))
	})
}

func TestRepository_Update(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	repo := reviews.NewRepository(tdb.DB)
	ctx := context.Background()

	t.Run("update existing review", func(t *testing.T) {
		tdb.Cleanup(t)

		productID := tdb.CreateTestProduct(t, "Test Product", nil, 99.99)
		reviewID := tdb.CreateTestReview(t, productID, "Original", "Author", 3, nil)

		tx, err := repo.BeginTx(ctx)
		require.NoError(t, err)
		defer tx.Rollback()

		newComment := "Updated comment"
		updated, err := repo.Update(ctx, tx, reviewID, models.UpdateReviewParams{
			FirstName: "Updated",
			LastName:  "Author",
			Rating:    5,
			Comment:   &newComment,
		})
		require.NoError(t, err)

		assert.Equal(t, "Updated", updated.FirstName)
		assert.Equal(t, "Author", updated.LastName)
		assert.Equal(t, 5, updated.Rating)
		require.NotNil(t, updated.Comment)
		assert.Equal(t, newComment, *updated.Comment)

		require.NoError(t, repo.CommitTx(tx))
	})

	t.Run("update non-existing review", func(t *testing.T) {
		tdb.Cleanup(t)

		tx, err := repo.BeginTx(ctx)
		require.NoError(t, err)
		defer tx.Rollback()

		_, err = repo.Update(ctx, tx, 999999, models.UpdateReviewParams{
			FirstName: "John",
			LastName:  "Doe",
			Rating:    5,
		})
		assert.Error(t, err)
		assert.ErrorIs(t, err, reviews.ErrNotFound)
	})
}

func TestRepository_UpdateByIDAndProductID(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	repo := reviews.NewRepository(tdb.DB)
	ctx := context.Background()

	t.Run("update with matching product", func(t *testing.T) {
		tdb.Cleanup(t)

		productID := tdb.CreateTestProduct(t, "Test Product", nil, 99.99)
		reviewID := tdb.CreateTestReview(t, productID, "John", "Doe", 3, nil)

		tx, err := repo.BeginTx(ctx)
		require.NoError(t, err)
		defer tx.Rollback()

		updated, err := repo.UpdateByIDAndProductID(ctx, tx, reviewID, productID, models.UpdateReviewParams{
			FirstName: "New",
			LastName:  "Author",
			Rating:    5,
		})
		require.NoError(t, err)
		assert.Equal(t, "New", updated.FirstName)
		assert.Equal(t, "Author", updated.LastName)

		require.NoError(t, repo.CommitTx(tx))
	})

	t.Run("update with non-matching product", func(t *testing.T) {
		tdb.Cleanup(t)

		productID1 := tdb.CreateTestProduct(t, "Product 1", nil, 99.99)
		productID2 := tdb.CreateTestProduct(t, "Product 2", nil, 49.99)
		reviewID := tdb.CreateTestReview(t, productID1, "John", "Doe", 5, nil)

		tx, err := repo.BeginTx(ctx)
		require.NoError(t, err)
		defer tx.Rollback()

		_, err = repo.UpdateByIDAndProductID(ctx, tx, reviewID, productID2, models.UpdateReviewParams{
			FirstName: "John",
			LastName:  "Doe",
			Rating:    5,
		})
		assert.Error(t, err)
		assert.ErrorIs(t, err, reviews.ErrNotFound)
	})
}

func TestRepository_Delete(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	repo := reviews.NewRepository(tdb.DB)
	ctx := context.Background()

	t.Run("delete existing review", func(t *testing.T) {
		tdb.Cleanup(t)

		productID := tdb.CreateTestProduct(t, "Test Product", nil, 99.99)
		reviewID := tdb.CreateTestReview(t, productID, "John", "Doe", 5, nil)

		tx, err := repo.BeginTx(ctx)
		require.NoError(t, err)
		defer tx.Rollback()

		err = repo.Delete(ctx, tx, reviewID)
		require.NoError(t, err)

		require.NoError(t, repo.CommitTx(tx))

		// Verify review is deleted
		tx2, err := repo.BeginTx(ctx)
		require.NoError(t, err)
		defer tx2.Rollback()

		_, err = repo.GetByID(ctx, tx2, reviewID)
		assert.ErrorIs(t, err, reviews.ErrNotFound)
	})

	t.Run("delete non-existing review", func(t *testing.T) {
		tdb.Cleanup(t)

		tx, err := repo.BeginTx(ctx)
		require.NoError(t, err)
		defer tx.Rollback()

		err = repo.Delete(ctx, tx, 999999)
		assert.Error(t, err)
		assert.ErrorIs(t, err, reviews.ErrNotFound)
	})
}

func TestRepository_DeleteByIDAndProductID(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	repo := reviews.NewRepository(tdb.DB)
	ctx := context.Background()

	t.Run("delete with matching product", func(t *testing.T) {
		tdb.Cleanup(t)

		productID := tdb.CreateTestProduct(t, "Test Product", nil, 99.99)
		reviewID := tdb.CreateTestReview(t, productID, "John", "Doe", 5, nil)

		tx, err := repo.BeginTx(ctx)
		require.NoError(t, err)
		defer tx.Rollback()

		err = repo.DeleteByIDAndProductID(ctx, tx, reviewID, productID)
		require.NoError(t, err)

		require.NoError(t, repo.CommitTx(tx))

		// Verify review is deleted
		tx2, err := repo.BeginTx(ctx)
		require.NoError(t, err)
		defer tx2.Rollback()

		_, err = repo.GetByID(ctx, tx2, reviewID)
		assert.ErrorIs(t, err, reviews.ErrNotFound)
	})

	t.Run("delete with non-matching product", func(t *testing.T) {
		tdb.Cleanup(t)

		productID1 := tdb.CreateTestProduct(t, "Product 1", nil, 99.99)
		productID2 := tdb.CreateTestProduct(t, "Product 2", nil, 49.99)
		reviewID := tdb.CreateTestReview(t, productID1, "John", "Doe", 5, nil)

		tx, err := repo.BeginTx(ctx)
		require.NoError(t, err)
		defer tx.Rollback()

		err = repo.DeleteByIDAndProductID(ctx, tx, reviewID, productID2)
		assert.Error(t, err)
		assert.ErrorIs(t, err, reviews.ErrNotFound)

		require.NoError(t, repo.CommitTx(tx))

		// Verify review still exists
		tx2, err := repo.BeginTx(ctx)
		require.NoError(t, err)
		defer tx2.Rollback()

		_, err = repo.GetByID(ctx, tx2, reviewID)
		assert.NoError(t, err)
	})
}

func TestRepository_GetAverageRatingByProductID(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	repo := reviews.NewRepository(tdb.DB)
	ctx := context.Background()

	t.Run("average rating with reviews", func(t *testing.T) {
		tdb.Cleanup(t)

		productID := tdb.CreateTestProduct(t, "Test Product", nil, 99.99)
		tdb.CreateTestReview(t, productID, "User1", "First", 5, nil)
		tdb.CreateTestReview(t, productID, "User2", "Second", 3, nil)
		tdb.CreateTestReview(t, productID, "User3", "Third", 4, nil)

		tx, err := repo.BeginTx(ctx)
		require.NoError(t, err)
		defer tx.Rollback()

		avgRating, err := repo.GetAverageRatingByProductID(ctx, tx, productID)
		require.NoError(t, err)
		require.NotNil(t, avgRating)
		assert.Equal(t, 4.0, *avgRating)

		require.NoError(t, repo.CommitTx(tx))
	})

	t.Run("average rating without reviews", func(t *testing.T) {
		tdb.Cleanup(t)

		productID := tdb.CreateTestProduct(t, "Test Product", nil, 99.99)

		tx, err := repo.BeginTx(ctx)
		require.NoError(t, err)
		defer tx.Rollback()

		avgRating, err := repo.GetAverageRatingByProductID(ctx, tx, productID)
		require.NoError(t, err)
		assert.Nil(t, avgRating)

		require.NoError(t, repo.CommitTx(tx))
	})
}
