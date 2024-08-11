package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
	"github.com/lithammer/shortuuid/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	client     *mongo.Client
	collection *mongo.Collection
	ctx        = context.Background()
)

func init() {
	if os.Getenv("ENV") != "production" {
		err := godotenv.Load(".env")
		if err != nil {
			log.Fatal("Error loading .env file", err)
		}
	}

	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Fatal("MONGODB_URI environment variable is not set")
	}

	var err error
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	collection = client.Database("urls").Collection("beta")
}

func main() {
	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
		AllowHeaders: "Content-Type",
	}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Server is running...")
	})

	app.Post("/short-it", createShortURLHandler)
	app.Get("/short/:key", redirectHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "4000"
	}

	if os.Getenv("ENV") == "production" {
		app.Static("/", "./client/.next")
	}

	log.Fatal(app.Listen("0.0.0.0:" + port))
}

func createShortURLHandler(c *fiber.Ctx) error {
	u := c.FormValue("URL")

	if u == "" {
		return c.Status(fiber.StatusBadRequest).SendString("URL is required")
	}

	key := shortuuid.New()

	err := insertMapping(key, u)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to store URL")
	}

	log.Println("URL mapped successfully")
	return c.Status(fiber.StatusOK).SendString(fmt.Sprintf("http://localhost:4000/short/%s", key))
}

func redirectHandler(c *fiber.Ctx) error {
	key := c.Params("key")
	if key == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Key field is empty")
	}

	u, err := fetchMapping(key)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to fetch URL")
	}

	if u == "" {
		return c.Status(fiber.StatusNotFound).SendString("URL not found")
	}

	return c.Redirect(u, fiber.StatusFound)
}

func insertMapping(key string, url string) error {
	_, err := collection.UpdateOne(
		ctx,
		bson.M{"key": key},
		bson.M{"$set": bson.M{"url": url}},
		options.Update().SetUpsert(true),
	)
	return err
}

func fetchMapping(key string) (string, error) {
	var result struct {
		URL string `bson:"url"`
	}
	err := collection.FindOne(ctx, bson.M{"key": key}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", nil
		}
		return "", err
	}
	return result.URL, nil
}
