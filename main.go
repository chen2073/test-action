package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Post struct {
	Message   string    `bson:"message"`
	Timestamp time.Time `bson:"timestamp"`
}

type ReqObj struct {
	Message string `json:"message"`
}

type ResObj struct {
	Status  string
	Message string
	Data    any
}

func openClient(connStr string) (*mongo.Client, error) {
	// Use the SetServerAPIOptions() method to set the version of the Stable API on the client
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.
		Client().
		ApplyURI(connStr).
		SetServerAPIOptions(serverAPI)

	// Create a new client and connect to the server
	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func testClient(client *mongo.Client) error {
	// Send a ping to confirm a successful connection
	return client.Database("admin").RunCommand(context.TODO(), bson.D{{"ping", 1}}).Err()
}

func closeClient(client *mongo.Client) error {
	return client.Disconnect(context.TODO())
}

func main() {
	connStr := os.Getenv("DB")

	client, err := openClient(connStr)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	err = testClient(client)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	defer closeClient(client)

	collection := client.Database("my-site").Collection("posts")

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		// AllowOrigins: "https://chen2073.click",
	}))

	app.Use(logger.New(logger.Config{
		Format: "[${ip}]:${port} ${status} - ${method} ${path}\n",
	}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("hello aaron, jake")
	})

	app.Get("/ping", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	app.Get("/ping_db", func(c *fiber.Ctx) error {
		err := testClient(client)
		if err != nil {
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		return c.SendStatus(fiber.StatusOK)
	})

	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	postRouter := app.Group("/post")

	postRouter.Get("/", func(c *fiber.Ctx) error {
		limit := c.QueryInt("limit", 10)
		offset := c.QueryInt("offset", 0)

		unsetStage := bson.D{{"$unset", bson.A{"_id"}}}
		sortStage := bson.D{{"$sort", bson.D{{"timestamp", -1}}}}
		limitStage := bson.D{{"$limit", limit}}
		offsetStage := bson.D{{"$skip", offset}}

		cursor, err := collection.Aggregate(context.TODO(), mongo.Pipeline{unsetStage, sortStage, limitStage, offsetStage})
		if err != nil {
			log.Fatal(err)
			return c.Status(fiber.StatusInternalServerError).JSON(ResObj{
				Status:  "fail",
				Message: "fail to retrieve records",
				Data:    nil,
			})
		}

		var results []Post
		err = cursor.All(context.TODO(), &results)
		if err != nil {
			log.Fatal(err)
			return c.Status(fiber.StatusInternalServerError).JSON(ResObj{
				Status:  "fail",
				Message: "fail to retrieve records",
				Data:    nil,
			})
		}

		return c.Status(fiber.StatusOK).JSON(ResObj{
			Status:  "success",
			Message: "records retrieved",
			Data:    results,
		})
	})

	postRouter.Post("/", func(c *fiber.Ctx) error {
		message := ReqObj{}
		err := c.BodyParser(&message)
		if err != nil {
			log.Fatal(err)
			return c.Status(fiber.StatusBadRequest).JSON(ResObj{
				Status:  "fail",
				Message: "invalid request body",
				Data:    nil,
			})
		}

		newPost := Post{
			Message:   message.Message,
			Timestamp: time.Now(),
		}

		result, err := collection.InsertOne(context.TODO(), newPost)
		if err != nil {
			log.Fatal(err)
			return c.Status(fiber.StatusInternalServerError).JSON(ResObj{
				Status:  "fail",
				Message: "fail to insert record",
				Data:    nil,
			})
		}

		return c.Status(fiber.StatusCreated).JSON(ResObj{
			Status:  "success",
			Message: fmt.Sprintf("Inserted document with _id: %v", result.InsertedID),
			Data:    nil,
		})
	})

	log.Fatal(app.Listen(":8000"))
}
