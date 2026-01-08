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

		comment := "Great product!"
		review, err := repo.Create(ctx, models.CreateReviewParams{
			ProductID: productID,
			Author:    "John Doe",
			Rating:    5,
			Comment:   &comment,
		})
		require.NoError(t, err)

		assert.NotZero(t, review.ID)
		assert.Equal(t, productID, review.ProductID)
		assert.Equal(t, "John Doe", review.Author)
		assert.Equal(t, 5, review.Rating)
		require.NotNil(t, review.Comment)
		assert.Equal(t, comment, *review.Comment)
	})

	t.Run("create review without comment", func(t *testing.T) {
		tdb.Cleanup(t)

		productID := tdb.CreateTestProduct(t, "Test Product", nil, 99.99)

		review, err := repo.Create(ctx, models.CreateReviewParams{
			ProductID: productID,
			Author:    "Jane Doe",
			Rating:    4,
			Comment:   nil,
		})
		require.NoError(t, err)
		assert.Nil(t, review.Comment)
	})

	t.Run("create review with various ratings", func(t *testing.T) {
		tdb.Cleanup(t)

		productID := tdb.CreateTestProduct(t, "Test Product", nil, 99.99)

		for rating := 1; rating <= 5; rating++ {
			review, err := repo.Create(ctx, models.CreateReviewParams{
				ProductID: productID,
				Author:    "User",
				Rating:    rating,
			})
			require.NoError(t, err, "rating %d", rating)
			assert.Equal(t, rating, review.Rating)
		}
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
		reviewID := tdb.CreateTestReview(t, productID, "Author", 5, &comment)

		review, err := repo.GetByID(ctx, reviewID)
		require.NoError(t, err)

		assert.Equal(t, reviewID, review.ID)
		assert.Equal(t, productID, review.ProductID)
		assert.Equal(t, "Author", review.Author)
	})

	t.Run("get non-existing review", func(t *testing.T) {
		tdb.Cleanup(t)

		_, err := repo.GetByID(ctx, 999999)
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
		reviewID := tdb.CreateTestReview(t, productID, "Author", 5, nil)

		review, err := repo.GetByIDAndProductID(ctx, reviewID, productID)
		require.NoError(t, err)
		assert.Equal(t, reviewID, review.ID)
	})

	t.Run("get review with non-matching product", func(t *testing.T) {
		tdb.Cleanup(t)

		productID1 := tdb.CreateTestProduct(t, "Product 1", nil, 99.99)
		productID2 := tdb.CreateTestProduct(t, "Product 2", nil, 49.99)
		reviewID := tdb.CreateTestReview(t, productID1, "Author", 5, nil)

		_, err := repo.GetByIDAndProductID(ctx, reviewID, productID2)
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

		reviewsList, err := repo.ListByProductID(ctx, models.ListReviewsParams{
			ProductID: productID,
			Limit:     10,
			Offset:    0,
		})
		require.NoError(t, err)
		assert.Empty(t, reviewsList)
	})

	t.Run("list reviews with pagination", func(t *testing.T) {
		tdb.Cleanup(t)

		productID := tdb.CreateTestProduct(t, "Test Product", nil, 99.99)

		// Create 5 reviews
		for i := 0; i < 5; i++ {
			tdb.CreateTestReview(t, productID, "User", i%5+1, nil)
		}

		// Get first page
		reviewsList, err := repo.ListByProductID(ctx, models.ListReviewsParams{
			ProductID: productID,
			Limit:     2,
			Offset:    0,
		})
		require.NoError(t, err)
		assert.Len(t, reviewsList, 2)

		// Get second page
		reviewsList, err = repo.ListByProductID(ctx, models.ListReviewsParams{
			ProductID: productID,
			Limit:     2,
			Offset:    2,
		})
		require.NoError(t, err)
		assert.Len(t, reviewsList, 2)
	})

	t.Run("list reviews only for specific product", func(t *testing.T) {
		tdb.Cleanup(t)

		productID1 := tdb.CreateTestProduct(t, "Product 1", nil, 99.99)
		productID2 := tdb.CreateTestProduct(t, "Product 2", nil, 49.99)

		tdb.CreateTestReview(t, productID1, "User1", 5, nil)
		tdb.CreateTestReview(t, productID1, "User2", 4, nil)
		tdb.CreateTestReview(t, productID2, "User3", 3, nil)

		reviewsList, err := repo.ListByProductID(ctx, models.ListReviewsParams{
			ProductID: productID1,
			Limit:     10,
			Offset:    0,
		})
		require.NoError(t, err)
		assert.Len(t, reviewsList, 2)

		for _, r := range reviewsList {
			assert.Equal(t, productID1, r.ProductID)
		}
	})
}

func TestRepository_Update(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	repo := reviews.NewRepository(tdb.DB)
	ctx := context.Background()

	t.Run("update existing review", func(t *testing.T) {
		tdb.Cleanup(t)

		productID := tdb.CreateTestProduct(t, "Test Product", nil, 99.99)
		reviewID := tdb.CreateTestReview(t, productID, "Original Author", 3, nil)

		newComment := "Updated comment"
		updated, err := repo.Update(ctx, reviewID, models.UpdateReviewParams{
			Author:  "Updated Author",
			Rating:  5,
			Comment: &newComment,
		})
		require.NoError(t, err)

		assert.Equal(t, "Updated Author", updated.Author)
		assert.Equal(t, 5, updated.Rating)
		require.NotNil(t, updated.Comment)
		assert.Equal(t, newComment, *updated.Comment)
	})

	t.Run("update non-existing review", func(t *testing.T) {
		tdb.Cleanup(t)

		_, err := repo.Update(ctx, 999999, models.UpdateReviewParams{
			Author: "Author",
			Rating: 5,
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
		reviewID := tdb.CreateTestReview(t, productID, "Author", 3, nil)

		updated, err := repo.UpdateByIDAndProductID(ctx, reviewID, productID, models.UpdateReviewParams{
			Author: "New Author",
			Rating: 5,
		})
		require.NoError(t, err)
		assert.Equal(t, "New Author", updated.Author)
	})

	t.Run("update with non-matching product", func(t *testing.T) {
		tdb.Cleanup(t)

		productID1 := tdb.CreateTestProduct(t, "Product 1", nil, 99.99)
		productID2 := tdb.CreateTestProduct(t, "Product 2", nil, 49.99)
		reviewID := tdb.CreateTestReview(t, productID1, "Author", 5, nil)

		_, err := repo.UpdateByIDAndProductID(ctx, reviewID, productID2, models.UpdateReviewParams{
			Author: "Author",
			Rating: 5,
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
		reviewID := tdb.CreateTestReview(t, productID, "Author", 5, nil)

		err := repo.Delete(ctx, reviewID)
		require.NoError(t, err)

		// Verify review is deleted
		_, err = repo.GetByID(ctx, reviewID)
		assert.ErrorIs(t, err, reviews.ErrNotFound)
	})

	t.Run("delete non-existing review", func(t *testing.T) {
		tdb.Cleanup(t)

		err := repo.Delete(ctx, 999999)
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
		reviewID := tdb.CreateTestReview(t, productID, "Author", 5, nil)

		err := repo.DeleteByIDAndProductID(ctx, reviewID, productID)
		require.NoError(t, err)

		// Verify review is deleted
		_, err = repo.GetByID(ctx, reviewID)
		assert.ErrorIs(t, err, reviews.ErrNotFound)
	})

	t.Run("delete with non-matching product", func(t *testing.T) {
		tdb.Cleanup(t)

		productID1 := tdb.CreateTestProduct(t, "Product 1", nil, 99.99)
		productID2 := tdb.CreateTestProduct(t, "Product 2", nil, 49.99)
		reviewID := tdb.CreateTestReview(t, productID1, "Author", 5, nil)

		err := repo.DeleteByIDAndProductID(ctx, reviewID, productID2)
		assert.Error(t, err)
		assert.ErrorIs(t, err, reviews.ErrNotFound)

		// Verify review still exists
		_, err = repo.GetByID(ctx, reviewID)
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
		tdb.CreateTestReview(t, productID, "User1", 5, nil)
		tdb.CreateTestReview(t, productID, "User2", 3, nil)
		tdb.CreateTestReview(t, productID, "User3", 4, nil)

		avgRating, err := repo.GetAverageRatingByProductID(ctx, productID)
		require.NoError(t, err)
		require.NotNil(t, avgRating)
		assert.Equal(t, 4.0, *avgRating)
	})

	t.Run("average rating without reviews", func(t *testing.T) {
		tdb.Cleanup(t)

		productID := tdb.CreateTestProduct(t, "Test Product", nil, 99.99)

		avgRating, err := repo.GetAverageRatingByProductID(ctx, productID)
		require.NoError(t, err)
		assert.Nil(t, avgRating)
	})
}
