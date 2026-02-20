package resolvers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"sync"
	"users/internal/user/models"

	"github.com/vikstrous/dataloadgen"
)

type CtxKey string

const (
	dataloaderKey CtxKey = "userDataloader"
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

func FetchUsers(ctx context.Context, ids []string) ([]*models.User, []error) {
	url := "http://localhost:8080/users?ids=" + strings.Join(ids, ",")
	fmt.Printf("[Users Subgraph] Making REST call to: %s\n", url)
	GetApiCounter(ctx).Increment("/users")
	resp, err := http.Get(url)
	if err != nil {
		return nil, []error{fmt.Errorf("failed to fetch users: %v", err)}
	}
	defer resp.Body.Close()

	var apiUsers []models.User
	if err := json.NewDecoder(resp.Body).Decode(&apiUsers); err != nil {
		return nil, []error{fmt.Errorf("failed to decode users: %v", err)}
	}

	userMap := make(map[string]*models.User)
	for i := range apiUsers {
		userMap[apiUsers[i].ID] = &apiUsers[i]
	}

	var results []*models.User
	errors := make([]error, len(ids))

	for _, id := range ids {
		if u, found := userMap[id]; found {
			results = append(results, u)
		} else {
			results = append(results, nil)
		}
	}

	return results, errors
}

func DataLoaderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		counter := &ApiCounter{counts: make(map[string]int)}
		ctx := context.WithValue(r.Context(), ApiCounterKey, counter)

		loader := dataloadgen.NewLoader(FetchUsers)
		ctx = context.WithValue(ctx, dataloaderKey, loader)
		next.ServeHTTP(w, r.WithContext(ctx))

		for endpoint, count := range counter.counts {
			fmt.Printf("[Users Subgraph] GraphQL query completed: %s, count=%d\n", endpoint, count)
		}
	})
}

func CtxLoadProvider(ctx context.Context) *dataloadgen.Loader[string, *models.User] {
	return ctx.Value(dataloaderKey).(*dataloadgen.Loader[string, *models.User])
}
