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

	"github.com/lib/pq"
)

type Product struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Price int    `json:"price"`
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
		CREATE TABLE IF NOT EXISTS products (
			id VARCHAR(255) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			price INT NOT NULL
		)
	`)
	if err != nil {
		log.Fatalf("Failed to create products table: %v\n", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("POST /products", createProduct)
	mux.HandleFunc("GET /products", getAllProducts)
	mux.HandleFunc("GET /products/{id}", getProductByID)
	mux.HandleFunc("PUT /products/{id}", updateProduct)
	mux.HandleFunc("DELETE /products/{id}", deleteProduct)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081" // Use a different default port from users
	}

	fmt.Printf("Products REST API server running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func createProduct(w http.ResponseWriter, r *http.Request) {
	var product Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if product.ID == "" {
		product.ID = generateID()
	}

	_, err := db.Exec("INSERT INTO products (id, name, price) VALUES ($1, $2, $3)", product.ID, product.Name, product.Price)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to insert product: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(product)
}

func getAllProducts(w http.ResponseWriter, r *http.Request) {
	idsParam := r.URL.Query().Get("ids")
	var rows *sql.Rows
	var err error

	if idsParam != "" {
		ids := strings.Split(idsParam, ",")
		rows, err = db.Query("SELECT id, name, price FROM products WHERE id = ANY($1)", pq.Array(ids))
	} else {
		rows, err = db.Query("SELECT id, name, price FROM products")
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("failed to query products: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var productList []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Price); err != nil {
			http.Error(w, fmt.Sprintf("failed to scan product: %v", err), http.StatusInternalServerError)
			return
		}
		productList = append(productList, p)
	}

	if productList == nil {
		productList = []Product{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(productList)
}

func getProductByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var product Product
	err := db.QueryRow("SELECT id, name, price FROM products WHERE id = $1", id).Scan(&product.ID, &product.Name, &product.Price)
	if err == sql.ErrNoRows {
		http.Error(w, "product not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, fmt.Sprintf("failed to query product: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}

func updateProduct(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var updatedProduct Product
	if err := json.NewDecoder(r.Body).Decode(&updatedProduct); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res, err := db.Exec("UPDATE products SET name = $1, price = $2 WHERE id = $3", updatedProduct.Name, updatedProduct.Price, id)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to update product: %v", err), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to check rows affected: %v", err), http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "product not found", http.StatusNotFound)
		return
	}

	updatedProduct.ID = id
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedProduct)
}

func deleteProduct(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	res, err := db.Exec("DELETE FROM products WHERE id = $1", id)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to delete product: %v", err), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to check rows affected: %v", err), http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "product not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
