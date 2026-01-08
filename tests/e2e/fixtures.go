package e2e

import (
	"product_review_hub/internal/api"
)

// ProductFixtures provides test data for product-related tests.
type ProductFixtures struct{}

// NewProductFixtures creates a new ProductFixtures instance.
func NewProductFixtures() *ProductFixtures {
	return &ProductFixtures{}
}

// ValidCreateRequest returns a valid ProductCreate request.
func (f *ProductFixtures) ValidCreateRequest() api.ProductCreate {
	return api.ProductCreate{
		Name:        "Test Product",
		Description: "A high-quality test product",
		Price:       99.99,
	}
}

// ValidCreateRequestWithName returns a valid ProductCreate request with custom name.
func (f *ProductFixtures) ValidCreateRequestWithName(name string) api.ProductCreate {
	return api.ProductCreate{
		Name:        name,
		Description: "A high-quality test product",
		Price:       99.99,
	}
}

// ValidCreateRequestWithPrice returns a valid ProductCreate request with custom price.
func (f *ProductFixtures) ValidCreateRequestWithPrice(price float32) api.ProductCreate {
	return api.ProductCreate{
		Name:        "Test Product",
		Description: "A high-quality test product",
		Price:       price,
	}
}

// CreateRequestWithEmptyName returns a ProductCreate request with empty name.
func (f *ProductFixtures) CreateRequestWithEmptyName() api.ProductCreate {
	return api.ProductCreate{
		Name:        "",
		Description: "A test product description",
		Price:       99.99,
	}
}

// CreateRequestWithZeroPrice returns a ProductCreate request with zero price.
func (f *ProductFixtures) CreateRequestWithZeroPrice() api.ProductCreate {
	return api.ProductCreate{
		Name:        "Test Product",
		Description: "A test product description",
		Price:       0,
	}
}

// CreateRequestWithNegativePrice returns a ProductCreate request with negative price.
func (f *ProductFixtures) CreateRequestWithNegativePrice() api.ProductCreate {
	return api.ProductCreate{
		Name:        "Test Product",
		Description: "A test product description",
		Price:       -10.00,
	}
}

// CreateRequestWithLongName returns a ProductCreate request with a very long name.
func (f *ProductFixtures) CreateRequestWithLongName() api.ProductCreate {
	longName := make([]byte, 300)
	for i := range longName {
		longName[i] = 'a'
	}
	return api.ProductCreate{
		Name:        string(longName),
		Description: "A test product description",
		Price:       99.99,
	}
}

// CreateRequestWithSpecialChars returns a ProductCreate request with special characters.
func (f *ProductFixtures) CreateRequestWithSpecialChars() api.ProductCreate {
	return api.ProductCreate{
		Name:        "Product with 'quotes' & <special> \"chars\"",
		Description: "Description with \n newlines \t and tabs",
		Price:       99.99,
	}
}

// CreateRequestWithUnicode returns a ProductCreate request with unicode characters.
func (f *ProductFixtures) CreateRequestWithUnicode() api.ProductCreate {
	return api.ProductCreate{
		Name:        "Продукт 产品 🎉",
		Description: "Описание продукта с эмодзи 😀",
		Price:       199.99,
	}
}

// CreateRequestWithMinPrice returns a ProductCreate request with minimum valid price.
func (f *ProductFixtures) CreateRequestWithMinPrice() api.ProductCreate {
	return api.ProductCreate{
		Name:        "Cheap Product",
		Description: "A very affordable product",
		Price:       0.01,
	}
}

// CreateRequestWithMaxPrice returns a ProductCreate request with a large price.
// Note: DB has DECIMAL(10,2) constraint, so max is 99999999.99
func (f *ProductFixtures) CreateRequestWithMaxPrice() api.ProductCreate {
	return api.ProductCreate{
		Name:        "Expensive Product",
		Description: "A luxury product",
		Price:       9999999.99,
	}
}
