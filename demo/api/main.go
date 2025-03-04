package main

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// OpenAPI specification
const openAPISpec = `
{
  "openapi": "3.0.0",
  "info": {
    "title": "OnTap Demo API",
    "description": "A simple API for demonstrating OnTap CLI features",
    "version": "1.0.0"
  },
  "paths": {
    "/items": {
      "get": {
        "summary": "List all items",
        "description": "Returns a list of all items",
        "operationId": "list-items",
        "parameters": [
          {
            "name": "limit",
            "in": "query",
            "description": "Maximum number of items to return",
            "schema": {
              "type": "integer",
              "default": 10
            }
          },
          {
            "name": "offset",
            "in": "query",
            "description": "Number of items to skip",
            "schema": {
              "type": "integer",
              "default": 0
            }
          }
        ],
        "responses": {
          "200": {
            "description": "A list of items",
            "content": {
              "application/json": {
                "schema": {
                  "type": "array",
                  "items": {
                    "$ref": "#/components/schemas/Item"
                  }
                }
              }
            }
          }
        }
      },
      "post": {
        "summary": "Create a new item",
        "description": "Creates a new item with the provided data",
        "operationId": "create-item",
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "$ref": "#/components/schemas/ItemInput"
              }
            }
          }
        },
        "responses": {
          "201": {
            "description": "Item created",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/Item"
                }
              }
            }
          },
          "400": {
            "description": "Invalid input"
          }
        }
      }
    },
    "/items/{id}": {
      "get": {
        "summary": "Get an item by ID",
        "description": "Returns a single item by its ID",
        "operationId": "get-item",
        "parameters": [
          {
            "name": "id",
            "in": "path",
            "required": true,
            "schema": {
              "type": "integer"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "An item",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/Item"
                }
              }
            }
          },
          "404": {
            "description": "Item not found"
          }
        }
      },
      "put": {
        "summary": "Update an item",
        "description": "Updates an existing item with the provided data",
        "operationId": "update-item",
        "parameters": [
          {
            "name": "id",
            "in": "path",
            "required": true,
            "schema": {
              "type": "integer"
            }
          }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "$ref": "#/components/schemas/ItemInput"
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "Item updated",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/Item"
                }
              }
            }
          },
          "400": {
            "description": "Invalid input"
          },
          "404": {
            "description": "Item not found"
          }
        }
      },
      "delete": {
        "summary": "Delete an item",
        "description": "Deletes an item by its ID",
        "operationId": "delete-item",
        "parameters": [
          {
            "name": "id",
            "in": "path",
            "required": true,
            "schema": {
              "type": "integer"
            }
          }
        ],
        "responses": {
          "204": {
            "description": "Item deleted"
          },
          "404": {
            "description": "Item not found"
          }
        }
      }
    }
  },
  "components": {
    "schemas": {
      "Item": {
        "type": "object",
        "properties": {
          "id": {
            "type": "integer"
          },
          "name": {
            "type": "string"
          },
          "description": {
            "type": "string"
          },
          "created_at": {
            "type": "string",
            "format": "date-time"
          }
        },
        "required": ["id", "name"]
      },
      "ItemInput": {
        "type": "object",
        "properties": {
          "name": {
            "type": "string"
          },
          "description": {
            "type": "string"
          }
        },
        "required": ["name"]
      }
    }
  }
}
`

func main() {
	// Create a new store
	store := NewInMemoryStore()

	// Create a new router
	mux := http.NewServeMux()

	// Setup auth middleware based on environment variable
	var handler http.Handler = mux
	authMode := os.Getenv("AUTH_MODE")
	if authMode == "basic" {
		username := os.Getenv("BASIC_USER")
		password := os.Getenv("BASIC_PASS")
		if username == "" || password == "" {
			log.Fatal("BASIC_USER and BASIC_PASS must be set when AUTH_MODE=basic")
		}
		handler = basicAuthMiddleware(username, password)(mux)
	}

	// Serve OpenAPI spec
	mux.HandleFunc("/openapi.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(openAPISpec))
	})

	// List items
	mux.HandleFunc("GET /items", func(w http.ResponseWriter, r *http.Request) {
		// Parse query parameters
		limitStr := r.URL.Query().Get("limit")
		offsetStr := r.URL.Query().Get("offset")

		limit := 10
		if limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
				limit = l
			}
		}

		offset := 0
		if offsetStr != "" {
			if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
				offset = o
			}
		}

		// Get items from store
		items := store.GetItems(limit, offset)

		// Return JSON response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(items)
	})

	// Get item by ID
	mux.HandleFunc("GET /items/{id}", func(w http.ResponseWriter, r *http.Request) {
		// Parse ID from path
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid ID format", http.StatusBadRequest)
			return
		}

		// Get item from store
		item, exists := store.GetItem(id)
		if !exists {
			http.Error(w, "Item not found", http.StatusNotFound)
			return
		}

		// Return JSON response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(item)
	})

	// Create item
	mux.HandleFunc("POST /items", func(w http.ResponseWriter, r *http.Request) {
		// Parse request body
		var input ItemInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate input
		if input.Name == "" {
			http.Error(w, "Name is required", http.StatusBadRequest)
			return
		}

		// Create item
		item := Item{
			Name:        input.Name,
			Description: input.Description,
		}

		// Add to store
		item = store.AddItem(item)

		// Return JSON response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(item)
	})

	// Update item
	mux.HandleFunc("PUT /items/{id}", func(w http.ResponseWriter, r *http.Request) {
		// Parse ID from path
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid ID format", http.StatusBadRequest)
			return
		}

		// Parse request body
		var input ItemInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate input
		if input.Name == "" {
			http.Error(w, "Name is required", http.StatusBadRequest)
			return
		}

		// Create item
		item := Item{
			Name:        input.Name,
			Description: input.Description,
		}

		// Update in store
		updatedItem, exists := store.UpdateItem(id, item)
		if !exists {
			http.Error(w, "Item not found", http.StatusNotFound)
			return
		}

		// Return JSON response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(updatedItem)
	})

	// Delete item
	mux.HandleFunc("DELETE /items/{id}", func(w http.ResponseWriter, r *http.Request) {
		// Parse ID from path
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid ID format", http.StatusBadRequest)
			return
		}

		// Delete from store
		success := store.DeleteItem(id)
		if !success {
			http.Error(w, "Item not found", http.StatusNotFound)
			return
		}

		// Return no content
		w.WriteHeader(http.StatusNoContent)
	})

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on :%s", port)
	log.Printf("OpenAPI spec available at http://localhost:%s/openapi.json", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}

// basicAuthMiddleware creates a middleware for basic authentication
func basicAuthMiddleware(username, password string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth for OpenAPI spec
			if r.URL.Path == "/openapi.json" {
				next.ServeHTTP(w, r)
				return
			}

			// Get the Authorization header
			auth := r.Header.Get("Authorization")
			if auth == "" {
				w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("Unauthorized"))
				return
			}

			// Check if the credentials are valid
			if !isValidBasicAuth(auth, username, password) {
				w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("Unauthorized"))
				return
			}

			// Call the next handler
			next.ServeHTTP(w, r)
		})
	}
}

// isValidBasicAuth checks if the provided basic auth credentials are valid
func isValidBasicAuth(auth, username, password string) bool {
	// Check if it's a Basic auth header
	if !strings.HasPrefix(auth, "Basic ") {
		return false
	}

	// Decode the base64 credentials
	credentials, err := base64.StdEncoding.DecodeString(auth[6:])
	if err != nil {
		return false
	}

	// Split the credentials into username and password
	parts := strings.SplitN(string(credentials), ":", 2)
	if len(parts) != 2 {
		return false
	}

	// Check if the credentials match
	return parts[0] == username && parts[1] == password
}
