package infrastructure

import (
	"context"

	"github.com/vsrecorder/pokemon-sprites-api/internal/domain/entity"
	"github.com/vsrecorder/pokemon-sprites-api/internal/domain/repository"
	"github.com/vsrecorder/pokemon-sprites-api/internal/infrastructure/model"
	"gorm.io/gorm"
)

type PokemonSprite struct {
	db *gorm.DB
}

func NewPokemonSprite(
	db *gorm.DB,
) repository.PokemonSpriteInterface {
	return &PokemonSprite{db}
}

func (i *PokemonSprite) FindAll(
	ctx context.Context,
) ([]*entity.PokemonSprite, error) {
	var pokemonSprites []*model.PokemonSprite
	if err := i.db.WithContext(ctx).Find(&pokemonSprites).Error; err != nil {
		return nil, err
	}

	var result []*entity.PokemonSprite
	for _, pokemonSprite := range pokemonSprites {
		result = append(result, &entity.PokemonSprite{
			ID:   pokemonSprite.ID,
			Name: pokemonSprite.Name,
		})
	}

	return result, nil
}
