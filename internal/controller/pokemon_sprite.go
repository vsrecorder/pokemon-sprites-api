package controller

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vsrecorder/pokemon-sprites-api/internal/controller/presenter"
	"github.com/vsrecorder/pokemon-sprites-api/internal/domain/repository"
	"github.com/vsrecorder/pokemon-sprites-api/internal/usecase"
)

const (
	PokemonSpritePath = "/pokemon-sprites"
)

type PokemonSprite struct {
	router     *gin.Engine
	repository repository.PokemonSpriteInterface
	usecase    usecase.PokemonSpriteInterface
}

func NewPokemonSprite(
	router *gin.Engine,
	repository repository.PokemonSpriteInterface,
	usecase usecase.PokemonSpriteInterface,
) *PokemonSprite {
	return &PokemonSprite{router, repository, usecase}
}

func (c *PokemonSprite) RegisterRoute(relativePath string) {
	r := c.router.Group(relativePath + PokemonSpritePath)
	r.GET(
		"",
		c.GetAll,
	)
}

func (c *PokemonSprite) GetAll(ctx *gin.Context) {
	pokemonSprites, err := c.usecase.FindAll(context.Background())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
		ctx.Abort()
		return
	}

	res := presenter.NewPokemonSpriteGetAllResponse(pokemonSprites)

	ctx.JSON(http.StatusOK, res)
}
