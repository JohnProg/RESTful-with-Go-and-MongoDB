package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"net/http"
)

var (
	session    *mgo.Session
	collection *mgo.Collection
)

type User struct {
	Id        bson.ObjectId `bson:"_id" json:"id"`
	FirstName string        `bson:"firstName" json:"firstName"`
	LastName  string        `bson:"lastName" json:"lastName"`
	Age       int           `bson:"age" json:"age"`
}

type UserResource struct {
	User User `json:"user"`
}

type UsersResource struct {
	Users []User `json:"users"`
}

type Status struct {
	Code    int  `json:"code"`
	IsValid bool `json:"isValid"`
}

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	var user User

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		panic(err)
	}
	// get a new id
	user.Id = bson.NewObjectId()
	//insert into document collection
	err = collection.Insert(&user)
	if err != nil {
		panic(err)
	}
	j, err := json.Marshal(UserResource{User: user})
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
}

func UsersHandler(w http.ResponseWriter, r *http.Request) {
	var users []User

	iter := collection.Find(nil).Iter()
	result := User{}
	for iter.Next(&result) {
		users = append(users, result)
	}

	w.Header().Set("Content-Type", "application/json")
	j, err := json.Marshal(UsersResource{Users: users})
	if err != nil {
		panic(err)
	}
	w.Write(j)
}

func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	// Get id from url
	vars := mux.Vars(r)
	id := bson.ObjectIdHex(vars["id"])

	// Decode the incoming user json
	var user User
	err = json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		panic(err)
	}

	// update on MogoDB
	// M -> map[string]interface{}
	err = collection.Update(bson.M{"_id": id},
		bson.M{"$set": bson.M{"firstName": user.FirstName,
			"lastName": user.LastName,
			"age":      user.Age,
		}})
	if err == nil {
		log.Printf("Updated User: %s", id, user.FirstName)
	} else {
		panic(err)
	}
	j, err := json.Marshal(UserResource{User: user})
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
}
func GetUserHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	// Get id from url
	vars := mux.Vars(r)
	id := bson.ObjectIdHex(vars["id"])

	var user User
	err = collection.Find(bson.M{"_id": id}).One(&user)
	if err != nil {
		panic(err)
	}
	j, err := json.Marshal(UserResource{User: user})
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
}

func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)
	id := vars["id"]

	// Remove from database
	err = collection.Remove(bson.M{"_id": bson.ObjectIdHex(id)})
	if err != nil {
		log.Printf("Could not find User %s to delete", id)
	}
	j, err := json.Marshal(Status{200, true})
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
}

func init() {
	log.Println("Starting mongodb session")
	var err error
	session, err = mgo.Dial("localhost")
	if err != nil {
		log.Fatalf("Error connecting to MongoDB")
	}

	collection = session.DB("usersDB2").C("users")
}
func main() {

	router := mux.NewRouter()

	router.HandleFunc("/api/users", UsersHandler).Methods("GET")
	router.HandleFunc("/api/users", CreateUserHandler).Methods("POST")
	router.HandleFunc("/api/users/{id}", GetUserHandler).Methods("GET")
	router.HandleFunc("/api/users/{id}", UpdateUserHandler).Methods("PUT")
	router.HandleFunc("/api/users/{id}", DeleteUserHandler).Methods("DELETE")
	http.Handle("/api/", router)

	defer session.Close()

	log.Println("Listening on 8080")
	http.ListenAndServe(":8080", nil)
}
