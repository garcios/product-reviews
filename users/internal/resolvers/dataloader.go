package resolvers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"users/internal/user/models"

	"github.com/vikstrous/dataloadgen"
)

type CtxKey string

const dataloaderKey CtxKey = "userDataloader"

func FetchUsers(ctx context.Context, ids []string) ([]*models.User, []error) {
	url := "http://localhost:8080/users?ids=" + strings.Join(ids, ",")
	fmt.Printf("[Users Subgraph] Making REST call to: %s\n", url)
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
		loader := dataloadgen.NewLoader(FetchUsers)
		ctx := context.WithValue(r.Context(), dataloaderKey, loader)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func CtxLoadProvider(ctx context.Context) *dataloadgen.Loader[string, *models.User] {
	return ctx.Value(dataloaderKey).(*dataloadgen.Loader[string, *models.User])
}
