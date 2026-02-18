package repository

import (
	"context"

	"github.com/vsrecorder/pokemon-sprites-api/internal/domain/entity"
)

type PokemonSpriteInterface interface {
	FindAll(
		ctx context.Context,
	) ([]*entity.PokemonSprite, error)
}
