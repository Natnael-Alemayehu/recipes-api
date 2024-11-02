package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/natnael-alemayehu/recipes-api/docs"
	"github.com/natnael-alemayehu/recipes-api/handlers"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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

var recipeHandler *handlers.RecipesHandler

func init() {
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal(err)
	}
	collection := client.Database(os.Getenv("MONGO_DATABASE")).Collection("recipes")
	recipeHandler = handlers.NewRecipiesHandler(ctx, collection)
	log.Println("Connected to MongoDB")
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

	router.POST("/recipes", recipeHandler.NewRecipeHander)
	router.GET("/recipes", recipeHandler.ListRecipeHandler)
	router.GET("/recipes/:id", recipeHandler.ShowRecipeHandler)
	router.GET("/recipes/search", recipeHandler.SearchRecipeHandler)
	router.PUT("/recipes/:id", recipeHandler.UpdateRecipeHandler)
	router.DELETE("/recipes/:id", recipeHandler.DeleteRecipeHandler)

	// Swagger endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	router.Run()
}
