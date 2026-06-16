package presenter

import (
	"fmt"
	"strings"

	"github.com/vsrecorder/pokemon-sprites-api/internal/controller/dto"
	"github.com/vsrecorder/pokemon-sprites-api/internal/domain/entity"
)

func NewPokemonSpriteGetAllResponse(
	pokemonSprites []*entity.PokemonSprite,
) []*dto.PokemonSpriteResponse {
	ret := []*dto.PokemonSpriteResponse{}

	for _, pokemonSprite := range pokemonSprites {
		imageURL := fmt.Sprintf("https://xx8nnpgt.user.webaccel.jp/images/pokemon-sprites/%s.png", strings.TrimLeft(pokemonSprite.ID, "0"))

		ret = append(ret, &dto.PokemonSpriteResponse{

			ID:       pokemonSprite.ID,
			Name:     pokemonSprite.Name,
			ImageURL: imageURL,
		})
	}

	return ret
}
