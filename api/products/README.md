# Products REST API

This is a RESTful API written in Go for managing Product data. It provides standard CRUD operations and persists data into a PostgreSQL database.

## Running the API

You can start the requisite PostgreSQL database instance from the root folder if you have Docker installed:
```bash
cd ../../
podman compose up -d db
```

1. Navigate to this directory:
   ```bash
   cd api/products
   ```
2. Run the application:
   ```bash
   go run main.go
   ```

The server will automatically create the required `products` table and will start listening on `http://localhost:8081`.

---

## Endpoints

### 1. Create a Product
* **URL**: `/products`
* **Method**: `POST`
* **Request Body** (JSON):
  ```json
  {
    "name": "Mechanical Keyboard",
    "price": 10999
  }
  ```
  *(Note: You can optionally provide an `"id"`. If omitted, a random 16-character hex ID will be generated. Price represents integer cents or minimum currency division depending on your modeling).*
* **Success Response** (`201 Created`):
  ```json
  {
    "id": "1a2b3c4d5e6f7g8h",
    "name": "Mechanical Keyboard",
    "price": 10999
  }
  ```
* **Example curl**:
  ```bash
  curl -X POST -H "Content-Type: application/json" -d '{"name": "Mechanical Keyboard", "price": 10999}' http://localhost:8081/products
  ```

---

### 2. Get All Products
* **URL**: `/products` (Optional query parameter: `?ids=id1,id2,id3`)
* **Method**: `GET`
* **Success Response** (`200 OK`):
  ```json
  [
    {
      "id": "1a2b3c4d5e6f7g8h",
      "name": "Mechanical Keyboard",
      "price": 10999
    }
  ]
  ```
  *(Note: Returns an empty list `[]` if no products exist.)*
* **Example curl**:
  ```bash
  curl http://localhost:8081/products
  ```
* **Example curl (with ids)**:
  ```bash
  curl "http://localhost:8081/products?ids=1,5,7"
  ```

---

### 3. Get Product by ID
* **URL**: `/products/{id}`
* **Method**: `GET`
* **URL Params**: `id=[string]`
* **Success Response** (`200 OK`):
  ```json
  {
    "id": "1a2b3c4d5e6f7g8h",
    "name": "Mechanical Keyboard",
    "price": 10999
  }
  ```
* **Error Response** (`404 Not Found`):
  ```text
  product not found
  ```
* **Example curl**:
  ```bash
  curl http://localhost:8081/products/1a2b3c4d5e6f7g8h
  ```

---

### 4. Update a Product
* **URL**: `/products/{id}`
* **Method**: `PUT`
* **URL Params**: `id=[string]`
* **Request Body** (JSON):
  ```json
  {
    "name": "Wireless Mechanical Keyboard",
    "price": 12999
  }
  ```
* **Success Response** (`200 OK`):
  ```json
  {
    "id": "1a2b3c4d5e6f7g8h",
    "name": "Wireless Mechanical Keyboard",
    "price": 12999
  }
  ```
* **Error Response** (`404 Not Found`):
  ```text
  product not found
  ```
* **Example curl**:
  ```bash
  curl -X PUT -H "Content-Type: application/json" -d '{"name": "Wireless Mechanical Keyboard", "price": 12999}' http://localhost:8081/products/1a2b3c4d5e6f7g8h
  ```

---

### 5. Delete a Product
* **URL**: `/products/{id}`
* **Method**: `DELETE`
* **URL Params**: `id=[string]`
* **Success Response** (`204 No Content`)
* **Error Response** (`404 Not Found`):
  ```text
  product not found
  ```
* **Example curl**:
  ```bash
  curl -X DELETE http://localhost:8081/products/1a2b3c4d5e6f7g8h
  ```
