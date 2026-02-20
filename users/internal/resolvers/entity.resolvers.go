package resolvers

import (
	"context"

	"users/internal/generated"
	"users/internal/user/models"
)

func (r *entityResolver) FindUserByID(ctx context.Context, id string) (*models.User, error) {
	return CtxLoadProvider(ctx).Load(ctx, id)
}

func (r *Resolver) Entity() generated.EntityResolver { return &entityResolver{r} }

type entityResolver struct{ *Resolver }
