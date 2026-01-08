package reviews_test

import (
	"context"
	"product_review_hub/internal/models"
	"product_review_hub/internal/repository/reviews"
	"product_review_hub/internal/testutil"
	"testing"
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
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		if review.ID == 0 {
			t.Error("Create() review ID should not be 0")
		}
		if review.ProductID != productID {
			t.Errorf("Create() ProductID = %v, want %v", review.ProductID, productID)
		}
		if review.Author != "John Doe" {
			t.Errorf("Create() Author = %v, want %v", review.Author, "John Doe")
		}
		if review.Rating != 5 {
			t.Errorf("Create() Rating = %v, want %v", review.Rating, 5)
		}
		if review.Comment == nil || *review.Comment != comment {
			t.Errorf("Create() Comment = %v, want %v", review.Comment, comment)
		}
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
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		if review.Comment != nil {
			t.Errorf("Create() Comment = %v, want nil", review.Comment)
		}
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
			if err != nil {
				t.Fatalf("Create() error = %v for rating %d", err, rating)
			}
			if review.Rating != rating {
				t.Errorf("Create() Rating = %v, want %v", review.Rating, rating)
			}
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
		if err != nil {
			t.Fatalf("GetByID() error = %v", err)
		}

		if review.ID != reviewID {
			t.Errorf("GetByID() ID = %v, want %v", review.ID, reviewID)
		}
		if review.ProductID != productID {
			t.Errorf("GetByID() ProductID = %v, want %v", review.ProductID, productID)
		}
		if review.Author != "Author" {
			t.Errorf("GetByID() Author = %v, want %v", review.Author, "Author")
		}
	})

	t.Run("get non-existing review", func(t *testing.T) {
		tdb.Cleanup(t)

		_, err := repo.GetByID(ctx, 999999)
		if err == nil {
			t.Error("GetByID() expected error for non-existing review")
		}
		if err != reviews.ErrNotFound {
			t.Errorf("GetByID() error = %v, want %v", err, reviews.ErrNotFound)
		}
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
		if err != nil {
			t.Fatalf("GetByIDAndProductID() error = %v", err)
		}

		if review.ID != reviewID {
			t.Errorf("GetByIDAndProductID() ID = %v, want %v", review.ID, reviewID)
		}
	})

	t.Run("get review with non-matching product", func(t *testing.T) {
		tdb.Cleanup(t)

		productID1 := tdb.CreateTestProduct(t, "Product 1", nil, 99.99)
		productID2 := tdb.CreateTestProduct(t, "Product 2", nil, 49.99)
		reviewID := tdb.CreateTestReview(t, productID1, "Author", 5, nil)

		_, err := repo.GetByIDAndProductID(ctx, reviewID, productID2)
		if err == nil {
			t.Error("GetByIDAndProductID() expected error for non-matching product")
		}
		if err != reviews.ErrNotFound {
			t.Errorf("GetByIDAndProductID() error = %v, want %v", err, reviews.ErrNotFound)
		}
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
		if err != nil {
			t.Fatalf("ListByProductID() error = %v", err)
		}

		if len(reviewsList) != 0 {
			t.Errorf("ListByProductID() len = %v, want 0", len(reviewsList))
		}
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
		if err != nil {
			t.Fatalf("ListByProductID() error = %v", err)
		}
		if len(reviewsList) != 2 {
			t.Errorf("ListByProductID() len = %v, want 2", len(reviewsList))
		}

		// Get second page
		reviewsList, err = repo.ListByProductID(ctx, models.ListReviewsParams{
			ProductID: productID,
			Limit:     2,
			Offset:    2,
		})
		if err != nil {
			t.Fatalf("ListByProductID() error = %v", err)
		}
		if len(reviewsList) != 2 {
			t.Errorf("ListByProductID() len = %v, want 2", len(reviewsList))
		}
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
		if err != nil {
			t.Fatalf("ListByProductID() error = %v", err)
		}
		if len(reviewsList) != 2 {
			t.Errorf("ListByProductID() len = %v, want 2", len(reviewsList))
		}

		for _, r := range reviewsList {
			if r.ProductID != productID1 {
				t.Errorf("ListByProductID() review ProductID = %v, want %v", r.ProductID, productID1)
			}
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
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}

		if updated.Author != "Updated Author" {
			t.Errorf("Update() Author = %v, want %v", updated.Author, "Updated Author")
		}
		if updated.Rating != 5 {
			t.Errorf("Update() Rating = %v, want %v", updated.Rating, 5)
		}
		if updated.Comment == nil || *updated.Comment != newComment {
			t.Errorf("Update() Comment = %v, want %v", updated.Comment, newComment)
		}
	})

	t.Run("update non-existing review", func(t *testing.T) {
		tdb.Cleanup(t)

		_, err := repo.Update(ctx, 999999, models.UpdateReviewParams{
			Author: "Author",
			Rating: 5,
		})
		if err == nil {
			t.Error("Update() expected error for non-existing review")
		}
		if err != reviews.ErrNotFound {
			t.Errorf("Update() error = %v, want %v", err, reviews.ErrNotFound)
		}
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
		if err != nil {
			t.Fatalf("UpdateByIDAndProductID() error = %v", err)
		}

		if updated.Author != "New Author" {
			t.Errorf("UpdateByIDAndProductID() Author = %v, want %v", updated.Author, "New Author")
		}
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
		if err == nil {
			t.Error("UpdateByIDAndProductID() expected error for non-matching product")
		}
		if err != reviews.ErrNotFound {
			t.Errorf("UpdateByIDAndProductID() error = %v, want %v", err, reviews.ErrNotFound)
		}
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
		if err != nil {
			t.Fatalf("Delete() error = %v", err)
		}

		// Verify review is deleted
		_, err = repo.GetByID(ctx, reviewID)
		if err != reviews.ErrNotFound {
			t.Errorf("GetByID() after delete error = %v, want %v", err, reviews.ErrNotFound)
		}
	})

	t.Run("delete non-existing review", func(t *testing.T) {
		tdb.Cleanup(t)

		err := repo.Delete(ctx, 999999)
		if err == nil {
			t.Error("Delete() expected error for non-existing review")
		}
		if err != reviews.ErrNotFound {
			t.Errorf("Delete() error = %v, want %v", err, reviews.ErrNotFound)
		}
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
		if err != nil {
			t.Fatalf("DeleteByIDAndProductID() error = %v", err)
		}

		// Verify review is deleted
		_, err = repo.GetByID(ctx, reviewID)
		if err != reviews.ErrNotFound {
			t.Errorf("GetByID() after delete error = %v, want %v", err, reviews.ErrNotFound)
		}
	})

	t.Run("delete with non-matching product", func(t *testing.T) {
		tdb.Cleanup(t)

		productID1 := tdb.CreateTestProduct(t, "Product 1", nil, 99.99)
		productID2 := tdb.CreateTestProduct(t, "Product 2", nil, 49.99)
		reviewID := tdb.CreateTestReview(t, productID1, "Author", 5, nil)

		err := repo.DeleteByIDAndProductID(ctx, reviewID, productID2)
		if err == nil {
			t.Error("DeleteByIDAndProductID() expected error for non-matching product")
		}
		if err != reviews.ErrNotFound {
			t.Errorf("DeleteByIDAndProductID() error = %v, want %v", err, reviews.ErrNotFound)
		}

		// Verify review still exists
		_, err = repo.GetByID(ctx, reviewID)
		if err != nil {
			t.Errorf("GetByID() should find review: %v", err)
		}
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
		if err != nil {
			t.Fatalf("GetAverageRatingByProductID() error = %v", err)
		}

		if avgRating == nil {
			t.Fatal("GetAverageRatingByProductID() avgRating should not be nil")
		}

		expected := 4.0
		if *avgRating != expected {
			t.Errorf("GetAverageRatingByProductID() = %v, want %v", *avgRating, expected)
		}
	})

	t.Run("average rating without reviews", func(t *testing.T) {
		tdb.Cleanup(t)

		productID := tdb.CreateTestProduct(t, "Test Product", nil, 99.99)

		avgRating, err := repo.GetAverageRatingByProductID(ctx, productID)
		if err != nil {
			t.Fatalf("GetAverageRatingByProductID() error = %v", err)
		}

		if avgRating != nil {
			t.Errorf("GetAverageRatingByProductID() = %v, want nil for product without reviews", *avgRating)
		}
	})
}
