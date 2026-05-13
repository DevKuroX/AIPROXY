package router

import (
	"context"
	"errors"
	"sync"

	"github.com/DevKuroX/AIPROXY/internal/models"
)

var (
	aliasStore AliasStore
	aliasCache sync.Map
)

type AliasStore interface {
	GetModelAliasByAlias(ctx context.Context, aliasName string) (*models.ModelAlias, error)
}

func SetAliasStore(store AliasStore) {
	aliasStore = store
}

func ResolveAlias(ctx context.Context, aliasName string) (string, error) {
	if cached, ok := aliasCache.Load(aliasName); ok {
		return cached.(string), nil
	}

	if aliasStore == nil {
		return aliasName, nil
	}

	alias, err := aliasStore.GetModelAliasByAlias(ctx, aliasName)
	if err != nil {
		if errors.Is(err, ErrAliasNotFound) {
			return aliasName, nil
		}
		return "", err
	}

	aliasCache.Store(aliasName, alias.TargetModel)
	return alias.TargetModel, nil
}

var ErrAliasNotFound = errors.New("alias not found")
