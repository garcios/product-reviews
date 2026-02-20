# Users REST API

This is a simple RESTful API written in Go for managing User data. It provides standard CRUD (Create, Read, Update, Delete) operations.

## Running the API

1. Navigate to this directory:
   ```bash
   cd api/users
   ```
2. Run the application:
   ```bash
   go run main.go
   ```

The server will start listening on `http://localhost:8080`.

---

## Endpoints

### 1. Create a User
* **URL**: `/users`
* **Method**: `POST`
* **Request Body** (JSON):
  ```json
  {
    "username": "johndoe"
  }
  ```
  *(Note: You can optionally provide an `"id"`. If omitted, a random 16-character hex ID will be generated.)*
* **Success Response** (`201 Created`):
  ```json
  {
    "id": "1a2b3c4d5e6f7g8h",
    "username": "johndoe"
  }
  ```
* **Example curl**:
  ```bash
  curl -X POST -H "Content-Type: application/json" -d '{"username": "johndoe"}' http://localhost:8080/users
  ```

---

### 2. Get All Users
* **URL**: `/users`
* **Method**: `GET`
* **Success Response** (`200 OK`):
  ```json
  [
    {
      "id": "1a2b3c4d5e6f7g8h",
      "username": "johndoe"
    }
  ]
  ```
  *(Note: Returns an empty list `[]` if no users exist.)*
* **Example curl**:
  ```bash
  curl http://localhost:8080/users
  ```

---

### 3. Get User by ID
* **URL**: `/users/{id}`
* **Method**: `GET`
* **URL Params**: `id=[string]`
* **Success Response** (`200 OK`):
  ```json
  {
    "id": "1a2b3c4d5e6f7g8h",
    "username": "johndoe"
  }
  ```
* **Error Response** (`404 Not Found`):
  ```text
  user not found
  ```
* **Example curl**:
  ```bash
  curl http://localhost:8080/users/1a2b3c4d5e6f7g8h
  ```

---

### 4. Update a User
* **URL**: `/users/{id}`
* **Method**: `PUT`
* **URL Params**: `id=[string]`
* **Request Body** (JSON):
  ```json
  {
    "username": "janedoe"
  }
  ```
* **Success Response** (`200 OK`):
  ```json
  {
    "id": "1a2b3c4d5e6f7g8h",
    "username": "janedoe"
  }
  ```
* **Error Response** (`404 Not Found`):
  ```text
  user not found
  ```
* **Example curl**:
  ```bash
  curl -X PUT -H "Content-Type: application/json" -d '{"username": "janedoe"}' http://localhost:8080/users/1a2b3c4d5e6f7g8h
  ```

---

### 5. Delete a User
* **URL**: `/users/{id}`
* **Method**: `DELETE`
* **URL Params**: `id=[string]`
* **Success Response** (`204 No Content`)
* **Error Response** (`404 Not Found`):
  ```text
  user not found
  ```
* **Example curl**:
  ```bash
  curl -X DELETE http://localhost:8080/users/1a2b3c4d5e6f7g8h
  ```
