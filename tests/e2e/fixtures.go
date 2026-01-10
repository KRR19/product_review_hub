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
		Name:        "Product 产品 🎉",
		Description: "Product description with emoji 😀",
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

// ValidUpdateRequest returns a valid ProductUpdate request.
func (f *ProductFixtures) ValidUpdateRequest() api.ProductUpdate {
	return api.ProductUpdate{
		Name:        "Updated Product",
		Description: "An updated product description",
		Price:       149.99,
	}
}

// ValidUpdateRequestWithData returns a ProductUpdate with custom data.
func (f *ProductFixtures) ValidUpdateRequestWithData(name, desc string, price float32) api.ProductUpdate {
	return api.ProductUpdate{
		Name:        name,
		Description: desc,
		Price:       price,
	}
}

// UpdateRequestWithEmptyName returns a ProductUpdate request with empty name.
func (f *ProductFixtures) UpdateRequestWithEmptyName() api.ProductUpdate {
	return api.ProductUpdate{
		Name:        "",
		Description: "A test product description",
		Price:       99.99,
	}
}

// UpdateRequestWithZeroPrice returns a ProductUpdate request with zero price.
func (f *ProductFixtures) UpdateRequestWithZeroPrice() api.ProductUpdate {
	return api.ProductUpdate{
		Name:        "Test Product",
		Description: "A test product description",
		Price:       0,
	}
}

// UpdateRequestWithNegativePrice returns a ProductUpdate request with negative price.
func (f *ProductFixtures) UpdateRequestWithNegativePrice() api.ProductUpdate {
	return api.ProductUpdate{
		Name:        "Test Product",
		Description: "A test product description",
		Price:       -10.00,
	}
}

// ReviewFixtures provides test data for review-related tests.
type ReviewFixtures struct{}

// NewReviewFixtures creates a new ReviewFixtures instance.
func NewReviewFixtures() *ReviewFixtures {
	return &ReviewFixtures{}
}

// ValidCreateRequest returns a valid ReviewCreate request with minimal data.
func (f *ReviewFixtures) ValidCreateRequest() api.ReviewCreate {
	return api.ReviewCreate{
		Rating: 5,
	}
}

// ValidCreateRequestWithRating returns a valid ReviewCreate request with custom rating.
func (f *ReviewFixtures) ValidCreateRequestWithRating(rating int) api.ReviewCreate {
	return api.ReviewCreate{
		Rating: rating,
	}
}

// ValidCreateRequestFull returns a valid ReviewCreate request with all fields.
func (f *ReviewFixtures) ValidCreateRequestFull() api.ReviewCreate {
	firstName := "John"
	lastName := "Doe"
	comment := "Great product! Highly recommended."
	return api.ReviewCreate{
		Rating:    5,
		FirstName: &firstName,
		LastName:  &lastName,
		Comment:   &comment,
	}
}

// ValidCreateRequestWithAuthor returns a valid ReviewCreate request with first and last name.
func (f *ReviewFixtures) ValidCreateRequestWithAuthor(firstName, lastName string) api.ReviewCreate {
	return api.ReviewCreate{
		Rating:    4,
		FirstName: &firstName,
		LastName:  &lastName,
	}
}

// ValidCreateRequestWithComment returns a valid ReviewCreate request with comment.
func (f *ReviewFixtures) ValidCreateRequestWithComment(comment string) api.ReviewCreate {
	return api.ReviewCreate{
		Rating:  4,
		Comment: &comment,
	}
}

// CreateRequestWithZeroRating returns a ReviewCreate request with zero rating.
func (f *ReviewFixtures) CreateRequestWithZeroRating() api.ReviewCreate {
	return api.ReviewCreate{
		Rating: 0,
	}
}

// CreateRequestWithNegativeRating returns a ReviewCreate request with negative rating.
func (f *ReviewFixtures) CreateRequestWithNegativeRating() api.ReviewCreate {
	return api.ReviewCreate{
		Rating: -1,
	}
}

// CreateRequestWithRatingAboveMax returns a ReviewCreate request with rating above 5.
func (f *ReviewFixtures) CreateRequestWithRatingAboveMax() api.ReviewCreate {
	return api.ReviewCreate{
		Rating: 6,
	}
}

// ValidUpdateRequest returns a valid ReviewUpdate request.
func (f *ReviewFixtures) ValidUpdateRequest() api.ReviewUpdate {
	return api.ReviewUpdate{
		Rating: 4,
	}
}

// ValidUpdateRequestFull returns a valid ReviewUpdate request with all fields.
func (f *ReviewFixtures) ValidUpdateRequestFull() api.ReviewUpdate {
	firstName := "Jane"
	lastName := "Doe"
	comment := "Updated review - still great!"
	return api.ReviewUpdate{
		Rating:    4,
		FirstName: &firstName,
		LastName:  &lastName,
		Comment:   &comment,
	}
}

// UpdateRequestWithZeroRating returns a ReviewUpdate request with zero rating.
func (f *ReviewFixtures) UpdateRequestWithZeroRating() api.ReviewUpdate {
	return api.ReviewUpdate{
		Rating: 0,
	}
}

// UpdateRequestWithRatingAboveMax returns a ReviewUpdate request with rating above 5.
func (f *ReviewFixtures) UpdateRequestWithRatingAboveMax() api.ReviewUpdate {
	return api.ReviewUpdate{
		Rating: 6,
	}
}

// CreateRequestWithUnicode returns a ReviewCreate request with unicode characters.
func (f *ReviewFixtures) CreateRequestWithUnicode() api.ReviewCreate {
	firstName := "Ivan"
	lastName := "Petrov"
	comment := "Great product! 五星好评 🌟"
	return api.ReviewCreate{
		Rating:    5,
		FirstName: &firstName,
		LastName:  &lastName,
		Comment:   &comment,
	}
}

// CreateRequestWithSpecialChars returns a ReviewCreate request with special characters.
func (f *ReviewFixtures) CreateRequestWithSpecialChars() api.ReviewCreate {
	firstName := "John 'The Reviewer'"
	lastName := "Doe"
	comment := "Great <product> with \"quotes\" & special chars!"
	return api.ReviewCreate{
		Rating:    5,
		FirstName: &firstName,
		LastName:  &lastName,
		Comment:   &comment,
	}
}
