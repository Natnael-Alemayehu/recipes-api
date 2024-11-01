package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/natnael-alemayehu/recipes-api/docs"
	"github.com/rs/xid"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Recipe struct {
	ID           string    `json:"id,omitempty"`
	Name         string    `json:"name"`
	Tags         []string  `json:"tags"`
	Ingredients  []string  `json:"ingredients"`
	Instructions []string  `json:"instructions"`
	PublishedAt  time.Time `json:"publishedAt,omitempty"`
}

var recipes []Recipe

func init() {
	recipes = make([]Recipe, 0)
	file, _ := os.ReadFile("recipes.json")
	_ = json.Unmarshal([]byte(file), &recipes)
}

// NewRecipeHandler godoc
//
//	@Summary		Creates a new Recipe
//	@Description	This endpoint creates a new recipe in the db
//	@Produce		json
//	@Tags			recipes
//	@Param			id body Recipe true "The new Recipe"
//	@Success		200	{object}	Recipe
//	@Failure		400 {object}	map[string]interface{} "error"
//	@Router			/recipes [post]
func NewRecipeHander(c *gin.Context) {
	var recipe Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	recipe.ID = xid.New().String()
	recipe.PublishedAt = time.Now()
	recipes = append(recipes, recipe)
	c.JSON(http.StatusOK, recipe)
}

// ListRecipeHandler godoc
//
//	@Summary		Lists all the Recipes in the db
//	@Description	Lists all the recipes
//	@Produce		json
//	@Tags			recipes
//	@Success		200	{object}	[]Recipe
//	@Router			/recipes [get]
func ListRecipeHandler(c *gin.Context) {
	c.JSON(http.StatusOK, recipes)
}

// UpdateRecipeHandler godoc
//
//	@Summary		Update a recipe
//	@Description	Update an existing recipe given an ID parameter
//	@Tags			recipes
//	@Produce		json
//	@Param			id		path		string	true	"Recipe ID"
//	@Param			recipe	body		Recipe	true	"Updated recipe data"
//	@Success		200		{object}	Recipe
//
//	@Failure		400		{object}	map[string]interface{}	"Invalid input"
//	@Failure		404		{object}	map[string]interface{}	"Recipe not found"
//
//	@Router			/recipes/{id} [put]
func UpdateRecipeHandler(c *gin.Context) {
	id := c.Param("id")

	var recipe Recipe

	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	index := -1
	for i := 0; i < len(recipes); i++ {
		if recipes[i].ID == id {
			index = i
		}
	}
	if index == -1 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Recipe not found",
		})
		return
	}
	recipe.ID = id
	recipes[index] = recipe
	c.JSON(http.StatusOK, recipe)
}

// DeleteRecipeHandler godoc
//
//	@Summary		Deletes a recipe
//	@Description	Deletes a recipe from the db given an id
//	@Produce		json
//	@Tags			recipes
//	@Param			id path string true "Recipe ID"
//	@Success		200	{object}	map[string]interface{} "Deleted Successfully"
//	@Router			/recipes/{id} [delete]
func DeleteRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	index := -1
	for i := 0; i < len(recipes); i++ {
		if recipes[i].ID == id {
			index = i
		}
	}

	if index == -1 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Recipe not found",
		})
		return
	}
	recipes = append(recipes[:index], recipes[index+1:]...)
	c.JSON(http.StatusOK, gin.H{
		"message": "Recipe has been deleted",
	})
}

// ShowRecipeHandler godoc
//
//	@Summary		Show a recipe
//	@Description	Show a specific recipe
//	@Tags			recipes
//	@Produce		json
//	@Param			id	path		string	true	"Recipe ID"
//	@Success		200	{object}	Recipe
//	@Failure		400	{object}	map[string]interface{}	"Invalid input"
//	@Failure		404	{object}	map[string]interface{}	"Recipe not found"
//	@Router			/recipes/{id} [get]
func ShowRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	index := -1
	for i := 0; i < len(recipes); i++ {
		if recipes[i].ID == id {
			index = i
		}
	}
	if index == -1 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Recipe not found",
		})
		return
	}
	c.JSON(http.StatusOK, recipes[index])
}

// ShowRecipeHandler godoc
// @Summary		Search a recipe from a tag
// @Description	This endpoint will search the db and return the recipes within that tag.
// @Tags			recipes
// @Produce		json
// @Param		tag	query string	true	"Tag to search recipes"
// @Success		200	{object}	Recipe
// @Failure		400	{object}	map[string]interface{}	"Invalid input"
// @Failure		404	{object}	map[string]interface{}	"Recipe not found"
// @Router			/recipes/search [get]
func SearchRecipeHandler(c *gin.Context) {
	tag := c.Query("tag")
	var listOfRecipe []Recipe

	for i := 0; i < len(recipes); i++ {
		for _, t := range recipes[i].Tags {
			if strings.EqualFold(t, tag) {
				listOfRecipe = append(listOfRecipe, recipes[i])
			}
		}
	}
	c.JSON(http.StatusOK, listOfRecipe)
}

// @title			Recipe API
// @version		1.0
// @description	This is an Recipe API
// @termsOfService	http://swagger.io/terms/
// @basePath		/
// @host			localhost:8080
func main() {
	router := gin.Default()
	docs.SwaggerInfo.BasePath = "/"

	router.POST("/recipes", NewRecipeHander)
	router.GET("/recipes", ListRecipeHandler)
	router.GET("/recipes/:id", ShowRecipeHandler)
	router.GET("/recipes/search", SearchRecipeHandler)
	router.PUT("/recipes/:id", UpdateRecipeHandler)
	router.DELETE("/recipes/:id", DeleteRecipeHandler)

	// Swagger endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	router.Run()
}
