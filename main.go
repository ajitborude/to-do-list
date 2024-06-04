package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ToDo struct {
	ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	IsCompleted bool               `json:"isCompleted"`
	Data        string             `json:"data"`
}

var collection *mongo.Collection

func main() {

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error while loading env")
	}

	port := os.Getenv("PORT")
	MONGODB_URI := os.Getenv("MONGODB_URI")

	clientOptions := options.Client().ApplyURI(MONGODB_URI)
	client, err := mongo.Connect(context.Background(), clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	defer client.Disconnect(context.Background())

	err = client.Ping(context.Background(), nil)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected To Mongo Database")

	collection = client.Database("todos_db").Collection("todos")

	app := fiber.New()

	if port == "" {
		port = "4000"
	}

	app.Get("/api/todos", getToDos)
	app.Post("/api/todos", createToDo)
	app.Patch("/api/todos/:id", updateToDo)
	app.Delete("api/todos/:id", deleteToDo)

	log.Fatal(app.Listen(":" + port))
}

func getToDos(c *fiber.Ctx) error {
	var todos []ToDo
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		return err
	}

	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var todo ToDo
		if err := cursor.Decode(&todo); err != nil {
			return err
		}
		todos = append(todos, todo)
	}

	if todos == nil {
		todos = []ToDo{}
	}

	return c.Status(200).JSON(fiber.Map{"data": todos})
}

func createToDo(c *fiber.Ctx) error {
	todo := new(ToDo)

	if err := c.BodyParser(todo); err != nil {
		return err
	}

	if todo.Data == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Todo data is required !"})
	}

	insertResult, err := collection.InsertOne(context.Background(), todo)

	if err != nil {
		return err
	}

	todo.ID = insertResult.InsertedID.(primitive.ObjectID)

	return c.Status(201).JSON(todo)
}

func updateToDo(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ToDo ID!"})
	}
	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": bson.M{"isCompleted": true}}

	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}
	return c.Status(200).JSON(fiber.Map{"success": true})
}

func deleteToDo(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ToDo ID!"})
	}
	filter := bson.M{"_id": objectID}
	_, err = collection.DeleteOne(context.Background(), filter)
	if err != nil {
		return err
	}
	return c.Status(200).JSON(fiber.Map{"success": true})
}
