package products_test

import (
	"context"
	"product_review_hub/internal/models"
	"product_review_hub/internal/repository/products"
	"product_review_hub/internal/testutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepository_Create(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	repo := products.NewRepository(tdb.DB)
	ctx := context.Background()

	tests := []struct {
		name    string
		params  models.CreateProductParams
		wantErr bool
	}{
		{
			name: "create product with all fields",
			params: models.CreateProductParams{
				Name:        "Test Product",
				Description: testutil.StringPtr("Test Description"),
				Price:       99.99,
			},
			wantErr: false,
		},
		{
			name: "create product without description",
			params: models.CreateProductParams{
				Name:        "Product Without Description",
				Description: nil,
				Price:       49.99,
			},
			wantErr: false,
		},
		{
			name: "create product with zero price",
			params: models.CreateProductParams{
				Name:        "Free Product",
				Description: testutil.StringPtr("This is free"),
				Price:       0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tdb.Cleanup(t)

			product, err := repo.Create(ctx, tt.params)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotZero(t, product.ID)
			assert.Equal(t, tt.params.Name, product.Name)
			if tt.params.Description != nil {
				require.NotNil(t, product.Description)
				assert.Equal(t, *tt.params.Description, *product.Description)
			}
			assert.Equal(t, tt.params.Price, product.Price)
			assert.False(t, product.CreatedAt.IsZero())
			assert.False(t, product.UpdatedAt.IsZero())
		})
	}
}

func TestRepository_GetByID(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	repo := products.NewRepository(tdb.DB)
	ctx := context.Background()

	t.Run("get existing product", func(t *testing.T) {
		tdb.Cleanup(t)

		desc := "Test Description"
		productID := tdb.CreateTestProduct(t, "Test Product", &desc, 99.99)

		product, err := repo.GetByID(ctx, productID)
		require.NoError(t, err)

		assert.Equal(t, productID, product.ID)
		assert.Equal(t, "Test Product", product.Name)
		assert.Nil(t, product.AverageRating)
	})

	t.Run("get product with reviews", func(t *testing.T) {
		tdb.Cleanup(t)

		desc := "Test Description"
		productID := tdb.CreateTestProduct(t, "Test Product", &desc, 99.99)
		tdb.CreateTestReview(t, productID, "User1", 5, nil)
		tdb.CreateTestReview(t, productID, "User2", 3, nil)

		product, err := repo.GetByID(ctx, productID)
		require.NoError(t, err)

		require.NotNil(t, product.AverageRating)
		assert.Equal(t, 4.0, *product.AverageRating)
	})

	t.Run("get non-existing product", func(t *testing.T) {
		tdb.Cleanup(t)

		_, err := repo.GetByID(ctx, 999999)
		assert.Error(t, err)
		assert.ErrorIs(t, err, products.ErrNotFound)
	})
}

func TestRepository_List(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	repo := products.NewRepository(tdb.DB)
	ctx := context.Background()

	t.Run("list empty products", func(t *testing.T) {
		tdb.Cleanup(t)

		productsList, err := repo.List(ctx, models.ListProductsParams{Limit: 10, Offset: 0})
		require.NoError(t, err)
		assert.Empty(t, productsList)
	})

	t.Run("list products with pagination", func(t *testing.T) {
		tdb.Cleanup(t)

		// Create 5 products
		for i := 0; i < 5; i++ {
			tdb.CreateTestProduct(t, "Product", nil, float64(i*10))
		}

		// Get first page
		productsList, err := repo.List(ctx, models.ListProductsParams{Limit: 2, Offset: 0})
		require.NoError(t, err)
		assert.Len(t, productsList, 2)

		// Get second page
		productsList, err = repo.List(ctx, models.ListProductsParams{Limit: 2, Offset: 2})
		require.NoError(t, err)
		assert.Len(t, productsList, 2)

		// Get last page
		productsList, err = repo.List(ctx, models.ListProductsParams{Limit: 2, Offset: 4})
		require.NoError(t, err)
		assert.Len(t, productsList, 1)
	})

	t.Run("list products with average rating", func(t *testing.T) {
		tdb.Cleanup(t)

		productID := tdb.CreateTestProduct(t, "Product", nil, 99.99)
		tdb.CreateTestReview(t, productID, "User1", 5, nil)
		tdb.CreateTestReview(t, productID, "User2", 3, nil)

		productsList, err := repo.List(ctx, models.ListProductsParams{Limit: 10, Offset: 0})
		require.NoError(t, err)
		require.Len(t, productsList, 1)

		require.NotNil(t, productsList[0].AverageRating)
		assert.Equal(t, 4.0, *productsList[0].AverageRating)
	})
}

func TestRepository_Update(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	repo := products.NewRepository(tdb.DB)
	ctx := context.Background()

	t.Run("update existing product", func(t *testing.T) {
		tdb.Cleanup(t)

		productID := tdb.CreateTestProduct(t, "Original Name", nil, 50.00)

		newDesc := "Updated Description"
		updated, err := repo.Update(ctx, productID, models.UpdateProductParams{
			Name:        "Updated Name",
			Description: &newDesc,
			Price:       75.00,
		})
		require.NoError(t, err)

		assert.Equal(t, "Updated Name", updated.Name)
		require.NotNil(t, updated.Description)
		assert.Equal(t, newDesc, *updated.Description)
		assert.Equal(t, 75.00, updated.Price)
	})

	t.Run("update non-existing product", func(t *testing.T) {
		tdb.Cleanup(t)

		_, err := repo.Update(ctx, 999999, models.UpdateProductParams{
			Name:  "Name",
			Price: 10.00,
		})
		assert.Error(t, err)
		assert.ErrorIs(t, err, products.ErrNotFound)
	})
}

func TestRepository_Delete(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	repo := products.NewRepository(tdb.DB)
	ctx := context.Background()

	t.Run("delete existing product", func(t *testing.T) {
		tdb.Cleanup(t)

		productID := tdb.CreateTestProduct(t, "To Delete", nil, 10.00)

		err := repo.Delete(ctx, productID)
		require.NoError(t, err)

		// Verify product is deleted
		_, err = repo.GetByID(ctx, productID)
		assert.ErrorIs(t, err, products.ErrNotFound)
	})

	t.Run("delete non-existing product", func(t *testing.T) {
		tdb.Cleanup(t)

		err := repo.Delete(ctx, 999999)
		assert.Error(t, err)
		assert.ErrorIs(t, err, products.ErrNotFound)
	})

	t.Run("delete product cascades to reviews", func(t *testing.T) {
		tdb.Cleanup(t)

		productID := tdb.CreateTestProduct(t, "Product", nil, 10.00)
		tdb.CreateTestReview(t, productID, "User", 5, nil)

		err := repo.Delete(ctx, productID)
		require.NoError(t, err)

		// Verify reviews are also deleted
		var count int
		err = tdb.DB.Get(&count, "SELECT COUNT(*) FROM reviews WHERE product_id = $1", productID)
		require.NoError(t, err)
		assert.Zero(t, count, "Reviews should be deleted with product")
	})
}

func TestRepository_Exists(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	repo := products.NewRepository(tdb.DB)
	ctx := context.Background()

	t.Run("existing product", func(t *testing.T) {
		tdb.Cleanup(t)

		productID := tdb.CreateTestProduct(t, "Product", nil, 10.00)

		exists, err := repo.Exists(ctx, productID)
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("non-existing product", func(t *testing.T) {
		tdb.Cleanup(t)

		exists, err := repo.Exists(ctx, 999999)
		require.NoError(t, err)
		assert.False(t, exists)
	})
}
