package resolvers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"products/internal/product/models"

	"github.com/vikstrous/dataloadgen"
)

type CtxKey string

const dataloaderKey CtxKey = "productDataloader"

// FetchProducts batches and requests products by their IDs
func FetchProducts(ctx context.Context, ids []string) ([]*models.Product, []error) {
	url := "http://localhost:8081/products?ids=" + strings.Join(ids, ",")
	fmt.Printf("[Products Subgraph] Making REST call to: %s\n", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, []error{fmt.Errorf("failed to fetch products: %v", err)}
	}
	defer resp.Body.Close()

	var apiProducts []models.Product
	if err := json.NewDecoder(resp.Body).Decode(&apiProducts); err != nil {
		return nil, []error{fmt.Errorf("failed to decode products: %v", err)}
	}

	productMap := make(map[string]*models.Product)
	for i := range apiProducts {
		productMap[apiProducts[i].ID] = &apiProducts[i]
	}

	var results []*models.Product
	errors := make([]error, len(ids))

	for _, id := range ids {
		if prod, found := productMap[id]; found {
			results = append(results, prod)
		} else {
			results = append(results, nil)
		}
	}

	return results, errors
}

// DataLoaderMiddleware wraps handlers and injects the dataloader instance into context
func DataLoaderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loader := dataloadgen.NewLoader(FetchProducts)
		ctx := context.WithValue(r.Context(), dataloaderKey, loader)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// CtxLoadProvider helper to retrieve the DataLoader safely inside Resolvers
func CtxLoadProvider(ctx context.Context) *dataloadgen.Loader[string, *models.Product] {
	return ctx.Value(dataloaderKey).(*dataloadgen.Loader[string, *models.Product])
}
