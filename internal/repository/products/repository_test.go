package products_test

import (
	"context"
	"product_review_hub/internal/models"
	"product_review_hub/internal/repository/products"
	"product_review_hub/internal/testutil"
	"testing"
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
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if product.ID == 0 {
					t.Error("Create() product ID should not be 0")
				}
				if product.Name != tt.params.Name {
					t.Errorf("Create() name = %v, want %v", product.Name, tt.params.Name)
				}
				if tt.params.Description != nil && (product.Description == nil || *product.Description != *tt.params.Description) {
					t.Errorf("Create() description = %v, want %v", product.Description, tt.params.Description)
				}
				if product.Price != tt.params.Price {
					t.Errorf("Create() price = %v, want %v", product.Price, tt.params.Price)
				}
				if product.CreatedAt.IsZero() {
					t.Error("Create() created_at should not be zero")
				}
				if product.UpdatedAt.IsZero() {
					t.Error("Create() updated_at should not be zero")
				}
			}
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
		if err != nil {
			t.Fatalf("GetByID() error = %v", err)
		}

		if product.ID != productID {
			t.Errorf("GetByID() ID = %v, want %v", product.ID, productID)
		}
		if product.Name != "Test Product" {
			t.Errorf("GetByID() Name = %v, want %v", product.Name, "Test Product")
		}
		if product.AverageRating != nil {
			t.Errorf("GetByID() AverageRating should be nil for product without reviews")
		}
	})

	t.Run("get product with reviews", func(t *testing.T) {
		tdb.Cleanup(t)

		desc := "Test Description"
		productID := tdb.CreateTestProduct(t, "Test Product", &desc, 99.99)
		tdb.CreateTestReview(t, productID, "User1", 5, nil)
		tdb.CreateTestReview(t, productID, "User2", 3, nil)

		product, err := repo.GetByID(ctx, productID)
		if err != nil {
			t.Fatalf("GetByID() error = %v", err)
		}

		if product.AverageRating == nil {
			t.Error("GetByID() AverageRating should not be nil")
		} else if *product.AverageRating != 4.0 {
			t.Errorf("GetByID() AverageRating = %v, want %v", *product.AverageRating, 4.0)
		}
	})

	t.Run("get non-existing product", func(t *testing.T) {
		tdb.Cleanup(t)

		_, err := repo.GetByID(ctx, 999999)
		if err == nil {
			t.Error("GetByID() expected error for non-existing product")
		}
		if err != products.ErrNotFound {
			t.Errorf("GetByID() error = %v, want %v", err, products.ErrNotFound)
		}
	})
}

func TestRepository_List(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	repo := products.NewRepository(tdb.DB)
	ctx := context.Background()

	t.Run("list empty products", func(t *testing.T) {
		tdb.Cleanup(t)

		productsList, err := repo.List(ctx, models.ListProductsParams{Limit: 10, Offset: 0})
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}

		if len(productsList) != 0 {
			t.Errorf("List() len = %v, want 0", len(productsList))
		}
	})

	t.Run("list products with pagination", func(t *testing.T) {
		tdb.Cleanup(t)

		// Create 5 products
		for i := 0; i < 5; i++ {
			tdb.CreateTestProduct(t, "Product", nil, float64(i*10))
		}

		// Get first page
		productsList, err := repo.List(ctx, models.ListProductsParams{Limit: 2, Offset: 0})
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}
		if len(productsList) != 2 {
			t.Errorf("List() len = %v, want 2", len(productsList))
		}

		// Get second page
		productsList, err = repo.List(ctx, models.ListProductsParams{Limit: 2, Offset: 2})
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}
		if len(productsList) != 2 {
			t.Errorf("List() len = %v, want 2", len(productsList))
		}

		// Get last page
		productsList, err = repo.List(ctx, models.ListProductsParams{Limit: 2, Offset: 4})
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}
		if len(productsList) != 1 {
			t.Errorf("List() len = %v, want 1", len(productsList))
		}
	})

	t.Run("list products with average rating", func(t *testing.T) {
		tdb.Cleanup(t)

		productID := tdb.CreateTestProduct(t, "Product", nil, 99.99)
		tdb.CreateTestReview(t, productID, "User1", 5, nil)
		tdb.CreateTestReview(t, productID, "User2", 3, nil)

		productsList, err := repo.List(ctx, models.ListProductsParams{Limit: 10, Offset: 0})
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}

		if len(productsList) != 1 {
			t.Fatalf("List() len = %v, want 1", len(productsList))
		}

		if productsList[0].AverageRating == nil {
			t.Error("List() AverageRating should not be nil")
		} else if *productsList[0].AverageRating != 4.0 {
			t.Errorf("List() AverageRating = %v, want %v", *productsList[0].AverageRating, 4.0)
		}
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
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}

		if updated.Name != "Updated Name" {
			t.Errorf("Update() Name = %v, want %v", updated.Name, "Updated Name")
		}
		if updated.Description == nil || *updated.Description != newDesc {
			t.Errorf("Update() Description = %v, want %v", updated.Description, newDesc)
		}
		if updated.Price != 75.00 {
			t.Errorf("Update() Price = %v, want %v", updated.Price, 75.00)
		}
	})

	t.Run("update non-existing product", func(t *testing.T) {
		tdb.Cleanup(t)

		_, err := repo.Update(ctx, 999999, models.UpdateProductParams{
			Name:  "Name",
			Price: 10.00,
		})
		if err == nil {
			t.Error("Update() expected error for non-existing product")
		}
		if err != products.ErrNotFound {
			t.Errorf("Update() error = %v, want %v", err, products.ErrNotFound)
		}
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
		if err != nil {
			t.Fatalf("Delete() error = %v", err)
		}

		// Verify product is deleted
		_, err = repo.GetByID(ctx, productID)
		if err != products.ErrNotFound {
			t.Errorf("GetByID() after delete error = %v, want %v", err, products.ErrNotFound)
		}
	})

	t.Run("delete non-existing product", func(t *testing.T) {
		tdb.Cleanup(t)

		err := repo.Delete(ctx, 999999)
		if err == nil {
			t.Error("Delete() expected error for non-existing product")
		}
		if err != products.ErrNotFound {
			t.Errorf("Delete() error = %v, want %v", err, products.ErrNotFound)
		}
	})

	t.Run("delete product cascades to reviews", func(t *testing.T) {
		tdb.Cleanup(t)

		productID := tdb.CreateTestProduct(t, "Product", nil, 10.00)
		tdb.CreateTestReview(t, productID, "User", 5, nil)

		err := repo.Delete(ctx, productID)
		if err != nil {
			t.Fatalf("Delete() error = %v", err)
		}

		// Verify reviews are also deleted
		var count int
		err = tdb.DB.Get(&count, "SELECT COUNT(*) FROM reviews WHERE product_id = $1", productID)
		if err != nil {
			t.Fatalf("failed to count reviews: %v", err)
		}
		if count != 0 {
			t.Errorf("Reviews count = %v, want 0 (should be deleted with product)", count)
		}
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
		if err != nil {
			t.Fatalf("Exists() error = %v", err)
		}
		if !exists {
			t.Error("Exists() = false, want true")
		}
	})

	t.Run("non-existing product", func(t *testing.T) {
		tdb.Cleanup(t)

		exists, err := repo.Exists(ctx, 999999)
		if err != nil {
			t.Fatalf("Exists() error = %v", err)
		}
		if exists {
			t.Error("Exists() = true, want false")
		}
	})
}
