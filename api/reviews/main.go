package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/lib/pq"
)

type Review struct {
	ID        string `json:"id"`
	ProductID string `json:"productId"`
	UserID    string `json:"userId"`
	Body      string `json:"body"`
	Rating    int    `json:"rating"`
	CreatedAt string `json:"createdAt"`
}

var db *sql.DB

func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func main() {
	var err error

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	}

	db, err = sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %v\n", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatalf("Could not connect to database: %v\n", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS reviews (
			id VARCHAR(255) PRIMARY KEY,
			product_id VARCHAR(255) NOT NULL,
			user_id VARCHAR(255) NOT NULL,
			body TEXT NOT NULL,
			rating INT NOT NULL,
			created_at TIMESTAMP NOT NULL
		)
	`)
	if err != nil {
		log.Fatalf("Failed to create reviews table: %v\n", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("POST /reviews", createReview)
	mux.HandleFunc("GET /reviews", getAllReviews)
	mux.HandleFunc("GET /reviews/{id}", getReviewByID)
	// Additional querying endpoints
	mux.HandleFunc("GET /products/{productId}/reviews", getReviewsByProduct)
	mux.HandleFunc("GET /users/{userId}/reviews", getReviewsByUser)
	mux.HandleFunc("PUT /reviews/{id}", updateReview)
	mux.HandleFunc("DELETE /reviews/{id}", deleteReview)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	fmt.Printf("Reviews REST API server running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func createReview(w http.ResponseWriter, r *http.Request) {
	var review Review
	if err := json.NewDecoder(r.Body).Decode(&review); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if review.ID == "" {
		review.ID = generateID()
	}
	if review.CreatedAt == "" {
		review.CreatedAt = time.Now().Format(time.RFC3339)
	}

	_, err := db.Exec(
		"INSERT INTO reviews (id, product_id, user_id, body, rating, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
		review.ID, review.ProductID, review.UserID, review.Body, review.Rating, review.CreatedAt,
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to insert review: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(review)
}

func getAllReviews(w http.ResponseWriter, r *http.Request) {
	idsParam := r.URL.Query().Get("ids")
	productIdsParam := r.URL.Query().Get("productIds")
	userIdsParam := r.URL.Query().Get("userIds")

	var rows *sql.Rows
	var err error

	if idsParam != "" {
		ids := strings.Split(idsParam, ",")
		rows, err = db.Query("SELECT id, product_id, user_id, body, rating, created_at FROM reviews WHERE id = ANY($1)", pq.Array(ids))
	} else if productIdsParam != "" {
		ids := strings.Split(productIdsParam, ",")
		rows, err = db.Query("SELECT id, product_id, user_id, body, rating, created_at FROM reviews WHERE product_id = ANY($1)", pq.Array(ids))
	} else if userIdsParam != "" {
		ids := strings.Split(userIdsParam, ",")
		rows, err = db.Query("SELECT id, product_id, user_id, body, rating, created_at FROM reviews WHERE user_id = ANY($1)", pq.Array(ids))
	} else {
		rows, err = db.Query("SELECT id, product_id, user_id, body, rating, created_at FROM reviews")
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("failed to query reviews: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var reviewList []Review
	for rows.Next() {
		var rev Review
		var t time.Time
		if err := rows.Scan(&rev.ID, &rev.ProductID, &rev.UserID, &rev.Body, &rev.Rating, &t); err != nil {
			http.Error(w, fmt.Sprintf("failed to scan review: %v", err), http.StatusInternalServerError)
			return
		}
		rev.CreatedAt = t.Format(time.RFC3339)
		reviewList = append(reviewList, rev)
	}

	if reviewList == nil {
		reviewList = []Review{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reviewList)
}

func getReviewByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var rev Review
	var t time.Time
	err := db.QueryRow("SELECT id, product_id, user_id, body, rating, created_at FROM reviews WHERE id = $1", id).
		Scan(&rev.ID, &rev.ProductID, &rev.UserID, &rev.Body, &rev.Rating, &t)
	if err == sql.ErrNoRows {
		http.Error(w, "review not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, fmt.Sprintf("failed to query review: %v", err), http.StatusInternalServerError)
		return
	}
	rev.CreatedAt = t.Format(time.RFC3339)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rev)
}

func getReviewsByProduct(w http.ResponseWriter, r *http.Request) {
	productId := r.PathValue("productId")

	rows, err := db.Query("SELECT id, product_id, user_id, body, rating, created_at FROM reviews WHERE product_id = $1", productId)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to query reviews: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var reviewList []Review
	for rows.Next() {
		var rev Review
		var t time.Time
		if err := rows.Scan(&rev.ID, &rev.ProductID, &rev.UserID, &rev.Body, &rev.Rating, &t); err != nil {
			http.Error(w, fmt.Sprintf("failed to scan review: %v", err), http.StatusInternalServerError)
			return
		}
		rev.CreatedAt = t.Format(time.RFC3339)
		reviewList = append(reviewList, rev)
	}

	if reviewList == nil {
		reviewList = []Review{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reviewList)
}

func getReviewsByUser(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("userId")

	rows, err := db.Query("SELECT id, product_id, user_id, body, rating, created_at FROM reviews WHERE user_id = $1", userId)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to query reviews: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var reviewList []Review
	for rows.Next() {
		var rev Review
		var t time.Time
		if err := rows.Scan(&rev.ID, &rev.ProductID, &rev.UserID, &rev.Body, &rev.Rating, &t); err != nil {
			http.Error(w, fmt.Sprintf("failed to scan review: %v", err), http.StatusInternalServerError)
			return
		}
		rev.CreatedAt = t.Format(time.RFC3339)
		reviewList = append(reviewList, rev)
	}

	if reviewList == nil {
		reviewList = []Review{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reviewList)
}

func updateReview(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var updatedReview Review
	if err := json.NewDecoder(r.Body).Decode(&updatedReview); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res, err := db.Exec("UPDATE reviews SET body = $1, rating = $2 WHERE id = $3", updatedReview.Body, updatedReview.Rating, id)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to update review: %v", err), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to check rows affected: %v", err), http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "review not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func deleteReview(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	res, err := db.Exec("DELETE FROM reviews WHERE id = $1", id)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to delete review: %v", err), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to check rows affected: %v", err), http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "review not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
