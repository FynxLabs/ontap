package main

import (
	"encoding/base64"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-fuego/fuego"
	"github.com/go-fuego/fuego/option"
	"github.com/go-fuego/fuego/param"
)

// Global store variable for simplicity
var store *InMemoryStore

func main() {
	// Create a new store
	store = NewInMemoryStore()
	log.Println("Initialized in-memory store with sample data")

	// Get the port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create a new Fuego server with proper OpenAPI configuration
	log.Println("Creating Fuego server with OpenAPI configuration...")
	s := fuego.NewServer(
		fuego.WithAddr(":"+port), // Explicitly set the port
		fuego.WithEngineOptions(
			fuego.WithOpenAPIConfig(fuego.OpenAPIConfig{
				SwaggerURL:       "/swagger",
				SpecURL:          "/openapi.json",
				JSONFilePath:     "openapi.json",
				DisableSwaggerUI: false,
				DisableLocalSave: false,
			}),
		),
	)

	// Setup auth middleware based on environment variable
	authMode := os.Getenv("AUTH_MODE")
	if authMode == "basic" {
		username := os.Getenv("BASIC_USER")
		password := os.Getenv("BASIC_PASS")
		if username == "" || password == "" {
			log.Fatal("BASIC_USER and BASIC_PASS must be set when AUTH_MODE=basic")
		}

		log.Printf("Setting up basic auth middleware with username: %s", username)
		// Add basic auth middleware to all routes except OpenAPI
		fuego.Use(s, func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Skip auth for OpenAPI spec
				if r.URL.Path == "/openapi.json" || r.URL.Path == "/swagger" || strings.HasPrefix(r.URL.Path, "/swagger/") {
					log.Printf("Skipping auth for OpenAPI endpoint: %s", r.URL.Path)
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
		})
	} else {
		log.Println("No authentication required (AUTH_MODE not set to 'basic')")
	}

	// Add a health check endpoint
	fuego.Get(s, "/health", healthCheck,
		option.Summary("Health check endpoint"),
		option.Description("Returns a simple message to indicate that the server is running"),
	)

	// Define API routes
	log.Println("Defining API routes...")
	itemsGroup := fuego.Group(s, "/items",
		option.Tags("items"),
	)

	// List items
	fuego.Get(itemsGroup, "", listItems,
		option.Summary("List all items"),
		option.Description("Returns a list of all items"),
		option.OperationID("list-items"),
		option.QueryInt("limit", "Maximum number of items to return", param.Default(10)),
		option.QueryInt("offset", "Number of items to skip", param.Default(0)),
	)

	// Get item by ID
	fuego.Get(itemsGroup, "/{id}", getItem,
		option.Summary("Get an item by ID"),
		option.Description("Returns a single item by its ID"),
		option.OperationID("get-item"),
	)

	// Create item
	fuego.Post(itemsGroup, "", createItem,
		option.Summary("Create a new item"),
		option.Description("Creates a new item with the provided data"),
		option.OperationID("create-item"),
	)

	// Update item
	fuego.Put(itemsGroup, "/{id}", updateItem,
		option.Summary("Update an item"),
		option.Description("Updates an existing item with the provided data"),
		option.OperationID("update-item"),
	)

	// Delete item
	fuego.Delete(itemsGroup, "/{id}", deleteItem,
		option.Summary("Delete an item"),
		option.Description("Deletes an item by its ID"),
		option.OperationID("delete-item"),
	)

	// Start the server
	log.Printf("Starting server on port %s...", port)
	log.Printf("Server URL: http://localhost:%s", port)
	log.Printf("OpenAPI spec URL: http://localhost:%s/openapi.json", port)
	log.Printf("Swagger UI URL: http://localhost:%s/swagger", port)
	log.Printf("Health check URL: http://localhost:%s/health", port)

	// Run the server
	s.Run()
}

// Health check handler
func healthCheck(c fuego.ContextNoBody) (string, error) {
	log.Println("Handling request: GET /health")
	return "Server is up and running", nil
}

// Helper function for basic auth
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

// Handler implementations
func listItems(c fuego.ContextNoBody) ([]Item, error) {
	log.Println("Handling request: GET /items")
	// Get items from store with default values
	items := store.GetItems(10, 0)
	log.Printf("Returning %d items", len(items))
	return items, nil
}

func getItem(c fuego.ContextNoBody) (Item, error) {
	// Get ID from path parameter
	idStr := c.Request().PathValue("id")
	log.Printf("Handling request: GET /items/%s", idStr)

	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("Error: Invalid ID format: %s", idStr)
		return Item{}, &fuego.HTTPError{
			Status: http.StatusBadRequest,
			Title:  "Invalid ID format",
		}
	}

	// Get item from store
	item, exists := store.GetItem(id)
	if !exists {
		log.Printf("Error: Item not found with ID: %d", id)
		return Item{}, &fuego.HTTPError{
			Status: http.StatusNotFound,
			Title:  "Item not found",
		}
	}

	log.Printf("Returning item with ID: %d", id)
	return item, nil
}

func createItem(c fuego.ContextNoBody) (Item, error) {
	log.Println("Handling request: POST /items")

	// Create a new item with default values
	item := Item{
		Name:        "New Item",
		Description: "A new item created via API",
		CreatedAt:   time.Now(),
	}

	// Add to store
	item = store.AddItem(item)
	log.Printf("Created new item with ID: %d", item.ID)

	return item, nil
}

func updateItem(c fuego.ContextNoBody) (Item, error) {
	// Get ID from path parameter
	idStr := c.Request().PathValue("id")
	log.Printf("Handling request: PUT /items/%s", idStr)

	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("Error: Invalid ID format: %s", idStr)
		return Item{}, &fuego.HTTPError{
			Status: http.StatusBadRequest,
			Title:  "Invalid ID format",
		}
	}

	// Get item from store
	item, exists := store.GetItem(id)
	if !exists {
		log.Printf("Error: Item not found with ID: %d", id)
		return Item{}, &fuego.HTTPError{
			Status: http.StatusNotFound,
			Title:  "Item not found",
		}
	}

	// Update the item with default values
	item.Name = "Updated Item"
	item.Description = "An updated item via API"

	// Update in store
	updatedItem, exists := store.UpdateItem(id, item)
	if !exists {
		log.Printf("Error: Failed to update item with ID: %d", id)
		return Item{}, &fuego.HTTPError{
			Status: http.StatusNotFound,
			Title:  "Item not found",
		}
	}

	log.Printf("Updated item with ID: %d", id)
	return updatedItem, nil
}

func deleteItem(c fuego.ContextNoBody) (interface{}, error) {
	// Get ID from path parameter
	idStr := c.Request().PathValue("id")
	log.Printf("Handling request: DELETE /items/%s", idStr)

	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("Error: Invalid ID format: %s", idStr)
		return nil, &fuego.HTTPError{
			Status: http.StatusBadRequest,
			Title:  "Invalid ID format",
		}
	}

	// Delete from store
	success := store.DeleteItem(id)
	if !success {
		log.Printf("Error: Item not found with ID: %d", id)
		return nil, &fuego.HTTPError{
			Status: http.StatusNotFound,
			Title:  "Item not found",
		}
	}

	log.Printf("Deleted item with ID: %d", id)
	return nil, nil
}
