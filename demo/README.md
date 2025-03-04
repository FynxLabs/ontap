# OnTap Demo API

A simple API for testing OnTap CLI features.

## Quick Start

1. Start the API servers:

   ```bash
   docker-compose up -d
   ```

2. Initialize OnTap with the demo config:

   ```bash
   ontap init --config config.yaml
   ```

3. Try these commands:

    ```bash
    # No auth API
    ontap demo-noauth list-items
    ontap demo-noauth get-item 1
    ontap demo-noauth create-item --data '{"name":"Test Item","description":"A test item"}'

    # Basic auth API
    ontap demo-basic list-items --auth="user:pass"
    ontap demo-basic get-item 1 --auth="user:pass"
    ontap demo-basic create-item --data '{"name":"Secure Item","description":"A secure item"}' --auth="user:pass"
    ```

## Available Endpoints

### Items API

- `GET /items` - List all items
  - Query parameters:
    - `limit` - Maximum number of items to return (default: 10)
    - `offset` - Number of items to skip (default: 0)

- `GET /items/{id}` - Get a specific item by ID

- `POST /items` - Create a new item
  - Request body:

    ```json
    {
      "name": "Item name",
      "description": "Item description"
    }
    ```

- `PUT /items/{id}` - Update an item
  - Request body:

    ```json
    {
      "name": "Updated name",
      "description": "Updated description"
    }
    ```

- `DELETE /items/{id}` - Delete an item

## Authentication

The demo includes two API servers:

1. **No Auth API** (<http://localhost:8080>)
   - No authentication required
   - Great for quick testing

2. **Basic Auth API** (<http://localhost:8081>)
   - Uses HTTP Basic Authentication
   - Username: `user`
   - Password: `pass`
   - Use with OnTap: `--auth="user:pass"`

## OpenAPI Specification

Both servers expose their OpenAPI specification at `/openapi.json`:

- No Auth API: <http://localhost:8080/openapi.json>
- Basic Auth API: <http://localhost:8081/openapi.json>

## Stopping the Servers

To stop the API servers:

```bash
docker-compose down
