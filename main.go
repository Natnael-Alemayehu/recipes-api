package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-contrib/sessions"
	redisStore "github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/natnael-alemayehu/recipes-api/docs"
	"github.com/natnael-alemayehu/recipes-api/handlers"
	"github.com/natnael-alemayehu/recipes-api/middleware"
	"github.com/redis/go-redis/v9"
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

var (
	recipeHandler *handlers.RecipesHandler
	authHandler   *handlers.AuthHandler
)

func init() {
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal(err)
	}
	collection := client.Database(os.Getenv("MONGO_DATABASE")).Collection("recipes")
	log.Println("Connected to MongoDB")

	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	fmt.Println(redisClient.Ping(ctx))
	recipeHandler = handlers.RecipiesHandler(ctx, collection, redisClient)

	collectionUsers := client.Database(os.Getenv("MONGO_DATABASE")).Collection("users")
	authHandler = handlers.NewAuthHandler(ctx, collectionUsers)
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

	store, _ := redisStore.NewStore(10, "tcp", os.Getenv("REDIS_URI"), "", []byte("string"))
	router.Use(sessions.Sessions("recipe-api", store))

	authorized := router.Group("/")

	router.GET("/recipes", recipeHandler.ListRecipeHandler)

	router.POST("/signin", authHandler.SignInHandler)
	router.POST("/refresh", authHandler.RefreshHandler)
	router.POST("/signup", authHandler.SignUpHandler)

	authorized.Use(middleware.AuthMiddleware())
	{
		authorized.POST("/recipes", recipeHandler.NewRecipeHander)
		authorized.GET("/recipes/:id", recipeHandler.ShowRecipeHandler)
		authorized.GET("/recipes/search", recipeHandler.SearchRecipeHandler)
		authorized.PUT("/recipes/:id", recipeHandler.UpdateRecipeHandler)
		authorized.DELETE("/recipe/:id", recipeHandler.DeleteRecipeHandler)

		// Swagger endpoint
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	}

	router.Run()
}
