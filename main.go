package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/natnael-alemayehu/recipes-api/docs"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Recipe struct {
	ID           primitive.ObjectID `json:"id" bson:"_id"`
	Name         string             `json:"name" bson:"name"`
	Tags         []string           `json:"tags" bson:"tags"`
	Ingredients  []string           `json:"ingredients" bson:"ingredients"`
	Instructions []string           `json:"instructions" bson:"instructions"`
	PublishedAt  time.Time          `json:"publishedAt" bson:"publishedAt"`
}

var (
	// client     *mongo.Client
	ctx        context.Context
	collection *mongo.Collection
)

func init() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal(err)
	}
	collection = client.Database(os.Getenv("MONGO_DATABASE")).Collection("recipes")
	log.Println("Connected to MongoDB")
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
	recipe.ID = primitive.NewObjectID()
	recipe.PublishedAt = time.Now()
	_, err := collection.InsertOne(ctx, recipe)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error while inserting a new recipe",
		})
		return
	}

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
	cur, err := collection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	defer cur.Close(ctx)

	recipes := make([]Recipe, 0)
	for cur.Next(ctx) {
		var recipe Recipe
		cur.Decode(&recipe)
		recipes = append(recipes, recipe)
	}
	c.JSON(http.StatusOK, recipes)
}

// var recipes []Recipe

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

	objectId, _ := primitive.ObjectIDFromHex(id)
	update := bson.D{{Key: "$set", Value: bson.D{
		{Key: "name", Value: recipe.Name},
		{Key: "tags", Value: recipe.Tags},
		{Key: "ingredients", Value: recipe.Ingredients},
		{Key: "instructions", Value: recipe.Instructions},
	}}}
	result, err := collection.UpdateByID(ctx, objectId, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}
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
func DeleteRecipeHandler(c *gin.Context) {
	id := c.Param("id")

	objectId, _ := primitive.ObjectIDFromHex(id)

	filter := bson.D{{Key: "_id", Value: objectId}}

	var deleteDocument bson.M
	err := collection.FindOneAndDelete(ctx, filter).Decode(&deleteDocument)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error deleting document",
		})
		return
	}

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
func ShowRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	objectId, _ := primitive.ObjectIDFromHex(id)

	filter := bson.D{{Key: "_id", Value: objectId}}

	var recipe Recipe
	err := collection.FindOne(ctx, filter).Decode(&recipe)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
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
func SearchRecipeHandler(c *gin.Context) {
	tag := c.Query("tag")

	filter := bson.D{{Key: "tags", Value: tag}}
	cur, err := collection.Find(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	defer cur.Close(ctx)

	listOfRecipe := make([]Recipe, 0)
	for cur.Next(ctx) {
		var recipe Recipe
		cur.Decode(&recipe)
		listOfRecipe = append(listOfRecipe, recipe)
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
