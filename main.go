package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
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
    // Load environment variables from .env file
   if os.Getenv(("ENV")) != "production" {
		// Load .env file if not in production
		err := godotenv.Load(".env")
		if err != nil {
	 	 	log.Fatal("Error loading .env file", err)
		}
	}

    // Get MongoDB URI from environment variable
    mongoURI := os.Getenv("MONGODB_URI")
    if mongoURI == "" {
        log.Fatal("MONGODB_URI environment variable is not set")
    }

    // Connect to MongoDB
    var err error
    client, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
    if err != nil {
        log.Fatal(err)
    }

    // Check connection
    err = client.Ping(ctx, nil)
    if err != nil {
        log.Fatal(err)
    }

    // Set up collection
    collection = client.Database("urls").Collection("beta")
}

func main() {
    r := chi.NewRouter()

    r.Use(middleware.Logger)
    r.Use(cors.Handler(cors.Options{
        AllowedOrigins: []string{"*"},
        AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders: []string{"Content-Type"},
    }))

    r.Get("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Server is running..."))
    })

    r.Post("/short-it", createShortURLHandler)
    r.Get("/short/{key}", redirectHandler)

	PORT := os.Getenv("PORT")
    if PORT == "" {
        log.Fatal("PORT environment variable is not set")
    }

	 // Serve static files in production
    if os.Getenv("ENV") == "production" {
        r.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("./client/.next"))))
    }

    http.ListenAndServe("0.0.0.0:" + PORT, r)
}

func createShortURLHandler(w http.ResponseWriter, r *http.Request) {
    r.ParseForm()
    u := r.Form.Get("URL")

    if u == "" {
        w.WriteHeader(http.StatusBadRequest)
        w.Write([]byte("URL is required"))
        return
    }

    // Generate key
    key := shortuuid.New()

    // Insert into the database
    err := insertMapping(key, u)
    if err != nil {
        http.Error(w, "Failed to store URL", http.StatusInternalServerError)
        return
    }

    log.Println("URL mapped successfully")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(fmt.Sprintf("http://localhost:4000/short/%s", key)))
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
    key := chi.URLParam(r, "key")
    if key == "" {
        w.WriteHeader(http.StatusBadRequest)
        w.Write([]byte("Key field is empty"))
        return
    }

    // Fetch mapping from the database
    u, err := fetchMapping(key)
    if err != nil {
        http.Error(w, "Failed to fetch URL", http.StatusInternalServerError)
        return
    }

    if u == "" {
        w.WriteHeader(http.StatusNotFound)
        w.Write([]byte("URL not found"))
        return
    }

    http.Redirect(w, r, u, http.StatusFound)
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
