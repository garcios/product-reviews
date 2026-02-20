package resolvers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"products/internal/product/models"
	"sync"

	"github.com/vikstrous/dataloadgen"
)

type CtxKey string

const (
	dataloaderKey CtxKey = "productDataloader"
	ApiCounterKey CtxKey = "apiCounterLoader"
)

type ApiCounter struct {
	mu     sync.Mutex
	counts map[string]int
}

func (c *ApiCounter) Increment(endpoint string) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.counts == nil {
		c.counts = make(map[string]int)
	}
	c.counts[endpoint]++
}

func GetApiCounter(ctx context.Context) *ApiCounter {
	if counter, ok := ctx.Value(ApiCounterKey).(*ApiCounter); ok {
		return counter
	}
	return nil
}

// FetchProducts batches and requests products by their IDs
func FetchProducts(ctx context.Context, ids []string) ([]*models.Product, []error) {
	url := "http://localhost:8081/products?ids=" + strings.Join(ids, ",")
	fmt.Printf("[Products Subgraph] Making REST call to: %s\n", url)
	GetApiCounter(ctx).Increment("/products")
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
		counter := &ApiCounter{counts: make(map[string]int)}
		ctx := context.WithValue(r.Context(), ApiCounterKey, counter)

		loader := dataloadgen.NewLoader(FetchProducts)
		ctx = context.WithValue(ctx, dataloaderKey, loader)
		next.ServeHTTP(w, r.WithContext(ctx))

		for endpoint, count := range counter.counts {
			fmt.Printf("[Products Subgraph] GraphQL query completed: %s, count=%d\n", endpoint, count)
		}
	})
}

// CtxLoadProvider helper to retrieve the DataLoader safely inside Resolvers
func CtxLoadProvider(ctx context.Context) *dataloadgen.Loader[string, *models.Product] {
	return ctx.Value(dataloaderKey).(*dataloadgen.Loader[string, *models.Product])
}
