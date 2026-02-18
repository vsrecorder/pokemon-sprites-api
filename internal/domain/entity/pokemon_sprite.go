package entity

type PokemonSprite struct {
	ID   string
	Name string
}

func NewPokemonSprite(
	id string,
	name string,
) *PokemonSprite {
	return &PokemonSprite{
		ID:   id,
		Name: name,
	}
}
