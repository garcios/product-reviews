# Reviews REST API

This is a RESTful API written in Go for managing Review data. It provides standard CRUD operations and persists data into a PostgreSQL database.

## Running the API

You can start the requisite PostgreSQL database instance from the root folder if you have Docker installed:
```bash
cd ../../
podman compose up -d db
```

1. Navigate to this directory:
   ```bash
   cd api/reviews
   ```
2. Run the application:
   ```bash
   go run main.go
   ```

The server will automatically create the required `reviews` table and will start listening on `http://localhost:8082`.

---

## Endpoints

### 1. Create a Review
* **URL**: `/reviews`
* **Method**: `POST`
* **Request Body** (JSON):
  ```json
  {
    "productId": "p_123",
    "userId": "u_1",
    "body": "This product is amazing!",
    "rating": 5
  }
  ```
  *(Note: You can optionally provide an `"id"` and/or `"createdAt"`. If omitted, they are auto-generated.)*
* **Success Response** (`201 Created`)

---

### 2. Get All Reviews
* **URL**: `/reviews`
* **Method**: `GET`
* **Success Response** (`200 OK`)

---

### 3. Get Review by ID
* **URL**: `/reviews/{id}`
* **Method**: `GET`
* **Success Response** (`200 OK`)

---

### 4. Get Reviews by Product
* **URL**: `/products/{productId}/reviews`
* **Method**: `GET`
* **Success Response** (`200 OK`)

---

### 5. Get Reviews by User
* **URL**: `/users/{userId}/reviews`
* **Method**: `GET`
* **Success Response** (`200 OK`)

---

### 6. Update a Review
* **URL**: `/reviews/{id}`
* **Method**: `PUT`
* **Request Body** (JSON):
  ```json
  {
    "body": "Actually, it broke after a week.",
    "rating": 2
  }
  ```
* **Success Response** (`204 No Content`)

---

### 7. Delete a Review
* **URL**: `/reviews/{id}`
* **Method**: `DELETE`
* **Success Response** (`204 No Content`)
