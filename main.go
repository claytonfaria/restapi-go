package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/claytonfaria/restapi-go/helper"
	"github.com/claytonfaria/restapi-go/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/gorilla/mux"
)

func getBooks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var books []models.Book
	collection := helper.ConnectDB()

	// bson.M, we passed empty filter, so we want to get all data.
	cur, err := collection.Find(context.TODO(), bson.M{})

	if err != nil {
		helper.GetError(err, w)
		return
	}

	// Close cursor once finished
	defer cur.Close(context.TODO())

	for cur.Next(context.TODO()) {
		// create a value into which the single document can be decoded
		var book models.Book
		err := cur.Decode(&book)
		if err != nil {
			log.Fatal(err)
		}
		// add item to array
		books = append(books, book)
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}
	json.NewEncoder(w).Encode(books)
}

func getBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var book models.Book
	// get params with mux
	var params = mux.Vars(r)

	// string to primitive.ObjectID
	id, _ := primitive.ObjectIDFromHex(params["id"])

	collection := helper.ConnectDB()

	// Create filter for sorting data
	filter := bson.M{"_id": id}
	err := collection.FindOne(context.TODO(), filter).Decode(&book)

	if err != nil {
		helper.GetError(err, w)
		return
	}
	json.NewEncoder(w).Encode(book)
}

func createBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var book models.Book

	// Decode body request params
	_ = json.NewDecoder(r.Body).Decode(&book)

	// connect db
	collection := helper.ConnectDB()

	// insert book model
	result, err := collection.InsertOne(context.TODO(), book)

	if err != nil {
		helper.GetError(err, w)
		return
	}

	json.NewEncoder(w).Encode(result)
}

func updateBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-Type", "application/json")

	var params = mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	var book models.Book

	collection := helper.ConnectDB()
	filter := bson.M{"_id": id}

	_ = json.NewDecoder(r.Body).Decode(&book)

	// prepare update model
	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "isbn", Value: book.Isbn},
			{Key: "title", Value: book.Title},
			{Key: "author", Value: bson.D{
				{Key: "firstname", Value: book.Author.FirstName},
				{Key: "lastname", Value: book.Author.LastName},
			}},
		}},
	}

	err := collection.FindOneAndUpdate(context.TODO(), filter, update).Decode(&book)

	if err != nil {
		helper.GetError(err, w)
		return
	}

	book.ID = id

	json.NewEncoder(w).Encode(book)
}

func deleteBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-Type", "application/json")

	var params = mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(params["id"])

	collection := helper.ConnectDB()
	filter := bson.M{"_id": id}

	deleteResult, err := collection.DeleteOne(context.TODO(), filter)

	if err != nil {
		helper.GetError(err, w)
		return
	}

	json.NewEncoder(w).Encode(deleteResult)
}

func main() {
	fmt.Println("Server up and running")

	r := mux.NewRouter()

	// arrange the routes
	r.HandleFunc("/api/books", getBooks).Methods("GET")
	r.HandleFunc("/api/books/{id}", getBook).Methods("GET")
	r.HandleFunc("/api/books", createBook).Methods("POST")
	r.HandleFunc("/api/books/{id}", updateBook).Methods("PUT")
	r.HandleFunc("/api/books/{id}", deleteBook).Methods("DELETE")

	// Set Port
	log.Fatal(http.ListenAndServe(":3000", r))
}
