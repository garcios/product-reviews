package resolvers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"product-reviews/internal/generated"
	"product-reviews/internal/review/models"

	"github.com/vikstrous/dataloadgen"
)

type CtxKey string

const (
	ProductReviewsKey CtxKey = "productReviewsLoader"
	UserReviewsKey    CtxKey = "userReviewsLoader"
	ReviewKey         CtxKey = "reviewLoader"
)

// FetchReviews fetches individual reviews by their IDs
func FetchReviews(ctx context.Context, ids []string) ([]*models.Review, []error) {
	url := "http://localhost:8082/reviews?ids=" + strings.Join(ids, ",")
	fmt.Printf("[Reviews Subgraph] Making REST call to: %s\n", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, []error{fmt.Errorf("failed to fetch reviews: %v", err)}
	}
	defer resp.Body.Close()

	var apiReviews []models.Review
	if err := json.NewDecoder(resp.Body).Decode(&apiReviews); err != nil {
		return nil, []error{fmt.Errorf("failed to decode reviews: %v", err)}
	}

	revMap := make(map[string]*models.Review)
	for i := range apiReviews {
		revMap[apiReviews[i].ID] = &apiReviews[i]
	}

	var results []*models.Review
	errors := make([]error, len(ids))

	for _, id := range ids {
		if r, found := revMap[id]; found {
			results = append(results, r)
		} else {
			results = append(results, nil)
		}
	}

	return results, errors
}

// FetchProductReviews dynamically groups array returns associated to specific products
func FetchProductReviews(ctx context.Context, productIds []string) ([]*generated.Product, []error) {
	url := "http://localhost:8082/reviews?productIds=" + strings.Join(productIds, ",")
	fmt.Printf("[Reviews Subgraph] Making REST call to: %s\n", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, []error{fmt.Errorf("failed to fetch product reviews: %v", err)}
	}
	defer resp.Body.Close()

	var apiReviews []models.Review
	if err := json.NewDecoder(resp.Body).Decode(&apiReviews); err != nil {
		return nil, []error{fmt.Errorf("failed to decode reviews: %v", err)}
	}

	// Map them properly since a product maps to Multiple Reviews
	productReviewMap := make(map[string][]*models.Review)
	for i := range apiReviews {
		pID := apiReviews[i].ProductID
		productReviewMap[pID] = append(productReviewMap[pID], &apiReviews[i])
	}

	var results []*generated.Product
	errors := make([]error, len(productIds))

	for _, id := range productIds {
		// Dataloader protocol asserts array lengths remain constant! Ensure nil falls-back seamlessly to empty slice.
		revs := productReviewMap[id]
		if revs == nil {
			revs = []*models.Review{}
		}

		results = append(results, &generated.Product{
			ID:      id,
			Reviews: revs,
		})
	}

	return results, errors
}

// FetchUserReviews counts review arrays mapped exclusively towards Users
func FetchUserReviews(ctx context.Context, userIds []string) ([]*generated.User, []error) {
	url := "http://localhost:8082/reviews?userIds=" + strings.Join(userIds, ",")
	fmt.Printf("[Reviews Subgraph] Making REST call to: %s\n", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, []error{fmt.Errorf("failed to fetch user reviews: %v", err)}
	}
	defer resp.Body.Close()

	var apiReviews []models.Review
	if err := json.NewDecoder(resp.Body).Decode(&apiReviews); err != nil {
		return nil, []error{fmt.Errorf("failed to decode user reviews: %v", err)}
	}

	userReviewCounter := make(map[string]int)
	for i := range apiReviews {
		userReviewCounter[apiReviews[i].UserID]++
	}

	var results []*generated.User
	errors := make([]error, len(userIds))

	for _, id := range userIds {
		// Must return mapped User instance to Dataloader schema wrapper regardless of 0 counts bounds.
		count := userReviewCounter[id]
		results = append(results, &generated.User{
			ID:           id,
			TotalReviews: &count,
		})
	}

	return results, errors
}

func DataLoaderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reviewLoader := dataloadgen.NewLoader(FetchReviews)
		prodReviewsLoader := dataloadgen.NewLoader(FetchProductReviews)
		userReviewsLoader := dataloadgen.NewLoader(FetchUserReviews)

		ctx := context.WithValue(r.Context(), ReviewKey, reviewLoader)
		ctx = context.WithValue(ctx, ProductReviewsKey, prodReviewsLoader)
		ctx = context.WithValue(ctx, UserReviewsKey, userReviewsLoader)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Helpers
func CtxReviewProvider(ctx context.Context) *dataloadgen.Loader[string, *models.Review] {
	return ctx.Value(ReviewKey).(*dataloadgen.Loader[string, *models.Review])
}

func CtxProdReviewProvider(ctx context.Context) *dataloadgen.Loader[string, *generated.Product] {
	return ctx.Value(ProductReviewsKey).(*dataloadgen.Loader[string, *generated.Product])
}

func CtxUserReviewProvider(ctx context.Context) *dataloadgen.Loader[string, *generated.User] {
	return ctx.Value(UserReviewsKey).(*dataloadgen.Loader[string, *generated.User])
}
