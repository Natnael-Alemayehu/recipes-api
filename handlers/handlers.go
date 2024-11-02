package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/natnael-alemayehu/recipes-api/models"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type RecipesHandler struct {
	collection  *mongo.Collection
	ctx         context.Context
	redisClient *redis.Client
}

func NewRecipiesHandler(ctx context.Context, collection *mongo.Collection, redisClient *redis.Client) *RecipesHandler {
	return &RecipesHandler{
		collection:  collection,
		ctx:         ctx,
		redisClient: redisClient,
	}
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
func (handler *RecipesHandler) NewRecipeHander(c *gin.Context) {
	var recipe models.Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	recipe.ID = primitive.NewObjectID()
	recipe.PublishedAt = time.Now()
	_, err := handler.collection.InsertOne(handler.ctx, recipe)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error while inserting a new recipe",
		})
		return
	}
	log.Println("Removing recipe from redis")
	handler.redisClient.Del(handler.ctx, "recipes")
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
func (handler *RecipesHandler) ListRecipeHandler(c *gin.Context) {
	val, err := handler.redisClient.Get(handler.ctx, "recipes").Result()
	if err == redis.Nil {
		log.Printf("Request to MondoDB")
		cur, err := handler.collection.Find(handler.ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		defer cur.Close(handler.ctx)

		recipes := make([]models.Recipe, 0)
		for cur.Next(handler.ctx) {
			var recipe models.Recipe
			cur.Decode(&recipe)
			recipes = append(recipes, recipe)
		}
		data, _ := json.Marshal(recipes)
		handler.redisClient.Set(handler.ctx, "recipes", string(data), 0)
		c.JSON(http.StatusOK, recipes)
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	} else {
		log.Printf("Request to Redis")
		recipes := make([]models.Recipe, 0)
		json.Unmarshal([]byte(val), &recipes)
		c.JSON(http.StatusOK, recipes)
	}
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
func (handler *RecipesHandler) UpdateRecipeHandler(c *gin.Context) {
	id := c.Param("id")

	var recipe models.Recipe

	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	objectId, _ := primitive.ObjectIDFromHex(id)
	update := bson.D{{Key: "$set", Value: bson.D{
		{Key: "name", Value: recipe.Name},
		{Key: "tags", Value: recipe.Tags},
		{Key: "ingredients", Value: recipe.Ingredients},
		{Key: "instructions", Value: recipe.Instructions},
	}}}
	result, err := handler.collection.UpdateByID(handler.ctx, objectId, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}
	log.Println("Deleting from redis")
	handler.redisClient.Del(handler.ctx, "recipes")
	c.JSON(http.StatusOK, gin.H{
		"message":           "Recipe has been changed",
		"changed_documents": result.ModifiedCount,
	})
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
func (handler *RecipesHandler) DeleteRecipeHandler(c *gin.Context) {
	id := c.Param("id")

	objectId, _ := primitive.ObjectIDFromHex(id)

	filter := bson.D{{Key: "_id", Value: objectId}}

	var deleteDocument bson.M
	err := handler.collection.FindOneAndDelete(handler.ctx, filter).Decode(&deleteDocument)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error deleting document",
		})
		return
	}

	log.Println("Deleting from redis")
	handler.redisClient.Del(handler.ctx, "recipes")

	c.JSON(http.StatusOK, gin.H{
		"message": "Recipe has been deleted",
		"deleted": deleteDocument,
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
func (handler *RecipesHandler) ShowRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	objectId, _ := primitive.ObjectIDFromHex(id)

	filter := bson.D{{Key: "_id", Value: objectId}}

	var recipe models.Recipe
	err := handler.collection.FindOne(handler.ctx, filter).Decode(&recipe)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, recipe)
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
func (handler *RecipesHandler) SearchRecipeHandler(c *gin.Context) {
	tag := c.Query("tag")

	filter := bson.D{{Key: "tags", Value: tag}}
	cur, err := handler.collection.Find(handler.ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	defer cur.Close(handler.ctx)

	listOfRecipe := make([]models.Recipe, 0)
	for cur.Next(handler.ctx) {
		var recipe models.Recipe
		cur.Decode(&recipe)
		listOfRecipe = append(listOfRecipe, recipe)
	}
	c.JSON(http.StatusOK, listOfRecipe)
}
