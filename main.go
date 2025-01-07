package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	// "fmt"
	// "strings"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

// Room represents the user model for our CRUD operations
type Room struct {
	ID        int     `json:"id"`
	Name      string  `json:"name"`
	MaxPerson int     `json:"max_person"`
	Price     float64 `json:"price"`
	RoomCode  string  `json:"room_code"`
}

var db *sql.DB

func main() {
	// Initialize database connection
	var err error
	db, err = sql.Open("mysql", "root:root@tcp(127.0.0.1:9900)/hotel")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Initialize router
	router := mux.NewRouter()

	// Define API routes
	router.HandleFunc("/rooms", getRooms).Methods("GET")    // Fetch all rooms
	router.HandleFunc("/room/{id}", getRoom).Methods("GET") // Fetch a room by ID
	router.HandleFunc("/room", createRoom).Methods("POST")  // Create a new room
	// router.HandleFunc("/user/{id}", updateUser).Methods("PUT")    // Update a user by ID
	// router.HandleFunc("/user/{id}", deleteUser).Methods("DELETE") // Delete a user by ID

	// Start server on port 8000
	log.Fatal(http.ListenAndServe(":8080", router))
}

func getRooms(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var rooms []Room
	rows, err := db.Query("SELECT id, name, max_person, price, room_code FROM room")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Print(err.Error())
		json.NewEncoder(w).Encode(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var room Room
		if err := rows.Scan(&room.ID, &room.Name, &room.MaxPerson, &room.Price, &room.RoomCode); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Print(err.Error())
			json.NewEncoder(w).Encode(err.Error())
		}
		rooms = append(rooms, room)
	}
	json.NewEncoder(w).Encode(rooms)
}

func getRoom(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	id := params["id"]
	var room Room
	err := db.QueryRow("SELECT id, name, max_person, price, room_code FROM room WHERE id = ?", id).Scan(&room.ID, &room.Name, &room.MaxPerson, &room.Price, &room.RoomCode)
	if err != nil {
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		log.Print(err.Error())
		json.NewEncoder(w).Encode(err.Error())
	}
	json.NewEncoder(w).Encode(room)
}

func createRoom(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var room Room
	err := json.NewDecoder(r.Body).Decode(&room)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Print(err.Error())
		return
	}

	room.RoomCode = roomCodeGenerator(room.Name, room.MaxPerson)
	result, err := db.Exec("INSERT INTO room (name, max_person, price, room_code) VALUES (?, ?, ?, ?)", room.Name, room.MaxPerson, room.Price, room.RoomCode)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Print(err.Error())
		json.NewEncoder(w).Encode(err.Error())
	}
	id, err := result.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Print(err.Error())
		json.NewEncoder(w).Encode(err.Error())
	}
	room.ID = int(id)
	json.NewEncoder(w).Encode(room)
}

func roomCodeGenerator(name string, maxPerson int) string {
	var roomCode string
	switch {
	case strings.Contains(name, "Super Suite"):
		roomCode = "SS"
	case strings.Contains(name, "Suite"):
		roomCode = "S"
	case strings.Contains(name, "Business"):
		roomCode = "B"
	case strings.Contains(name, "Casual"):
		roomCode = "C"
	case strings.Contains(name, "Family"):
		roomCode = "F"
	default:
		roomCode = "U"
	}

	roomCode = roomCode + fmt.Sprintf("%03d", maxPerson)

	return roomCode
}
