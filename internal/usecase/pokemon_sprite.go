package usecase

import (
	"context"

	"github.com/vsrecorder/pokemon-sprites-api/internal/domain/entity"
	"github.com/vsrecorder/pokemon-sprites-api/internal/domain/repository"
)

type PokemonSpriteInterface interface {
	FindAll(
		ctx context.Context,
	) ([]*entity.PokemonSprite, error)
}

type PokemonSprite struct {
	repository repository.PokemonSpriteInterface
}

func NewPokemonSprite(
	repository repository.PokemonSpriteInterface,
) PokemonSpriteInterface {
	return &PokemonSprite{repository}
}

func (u *PokemonSprite) FindAll(
	ctx context.Context,
) ([]*entity.PokemonSprite, error) {
	return u.repository.FindAll(ctx)
}
