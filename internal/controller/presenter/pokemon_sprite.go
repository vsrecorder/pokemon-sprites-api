package presenter

import (
	"github.com/vsrecorder/pokemon-sprites-api/internal/controller/dto"
	"github.com/vsrecorder/pokemon-sprites-api/internal/domain/entity"
)

func NewPokemonSpriteGetAllResponse(
	pokemonSprites []*entity.PokemonSprite,
) []*dto.PokemonSpriteResponse {
	ret := []*dto.PokemonSpriteResponse{}

	for _, pokemonSprite := range pokemonSprites {
		ret = append(ret, &dto.PokemonSpriteResponse{

			ID:   pokemonSprite.ID,
			Name: pokemonSprite.Name,
		})
	}

	return ret
}
