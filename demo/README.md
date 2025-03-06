# OnTap Demo API

A simple API for testing OnTap CLI features.

## Quick Start

The easiest way to start the demo is using the mise task:

```bash
# From the project root
mise run demo-start
```

This will:
1. Start the API servers
2. Wait for them to be ready
3. Download the OpenAPI spec to both the demo directory and the api subdirectory
4. Initialize OnTap with the demo config

Alternatively, you can start the demo manually:

1. Start the API servers:

   ```bash
   docker-compose up -d
   ```

2. Wait for the servers to be ready and download the OpenAPI spec:

   ```bash
   # Wait for the server to be ready
   curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/health

   # Download the OpenAPI spec to both locations
   curl -s http://localhost:8080/openapi.json -o openapi.json
   curl -s http://localhost:8080/openapi.json -o api/openapi.json
   ```

3. Initialize OnTap with the demo config:

   ```bash
   ontap init --config config.yaml
   ```

## Try These Commands

```bash
# Using URL-based OpenAPI spec
ontap --config ./config.yaml demo-noauth list-items
ontap --config ./config.yaml demo-noauth get-item 1
ontap --config ./config.yaml demo-noauth create-item --data '{"name":"Test Item","description":"A test item"}'

# Using local file-based OpenAPI spec
ontap --config ./config.yaml demo-noauth-file list-items
ontap --config ./config.yaml demo-noauth-file get-item 1
ontap --config ./config.yaml demo-noauth-file create-item --data '{"name":"Test Item","description":"A test item"}'

# Basic auth API
ontap --config ./config.yaml demo-basic list-items --auth="user:pass"
ontap --config ./config.yaml demo-basic get-item 1 --auth="user:pass"
ontap --config ./config.yaml demo-basic create-item --data '{"name":"Secure Item","description":"A secure item"}' --auth="user:pass"
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

The demo API servers generate and serve OpenAPI specifications in two ways:

1. **URL-based access**: Both servers expose their OpenAPI specification at `/openapi.json`:
   - No Auth API: <http://localhost:8080/openapi.json>
   - Basic Auth API: <http://localhost:8081/openapi.json>

2. **Local file-based access**: The `demo-start` task downloads the OpenAPI specification to local files:
   - Demo directory: `./openapi.json`
   - API directory: `./api/openapi.json`

OnTap supports both URL-based and file-based OpenAPI specifications, as demonstrated in the `config.yaml` file.

## Configuration Options

The `config.yaml` file demonstrates different ways to configure OnTap:

```yaml
apis:
  # Example using URL-based OpenAPI spec
  demo-noauth:
    apispec: http://localhost:8080/openapi.json  # URL to OpenAPI spec
    url: http://localhost:8080                   # Base URL for API requests
    cache_ttl: 1h0m0s                           # Cache spec for 1 hour

  # Example using local file-based OpenAPI spec
  demo-noauth-file:
    apispec: ./api/openapi.json                  # Path to local OpenAPI spec file
    url: http://localhost:8080                   # Base URL for API requests
    cache_ttl: 24h0m0s                          # Cache spec for 24 hours

  # Example with basic authentication
  demo-basic:
    apispec: http://localhost:8081/openapi.json  # URL to OpenAPI spec
    url: http://localhost:8081                   # Base URL for API requests
    auth: user:pass                              # Basic auth credentials
    cache_ttl: 1h0m0s                           # Cache spec for 1 hour
```

### Configuration Options Explained

- **apispec** (required): Path to the OpenAPI specification. Can be a URL or a local file path.
  - URL example: `http://localhost:8080/openapi.json`
  - File path example: `./api/openapi.json`
- **url** (required): Base URL for API requests.
- **auth** (optional): Authentication credentials in the format `username:password` for basic auth, or a token for bearer auth.
- **cache_ttl** (optional): Time-to-live for caching the OpenAPI spec. Default is 24 hours.
- **output** (optional): Default output format (json, yaml, csv, text, table).
- **headers** (optional): Default headers to include in all requests.

### URL vs. File-based Specs

- **URL-based specs** are great for development:
  - Always up-to-date with the latest API changes
  - No need to manually download the spec
  - Use the `refresh` command to update the cache: `ontap --config ./config.yaml refresh demo-noauth`

- **File-based specs** are useful for:
  - Offline development
  - Version control of the API spec
  - Faster startup (no need to download the spec)
  - Environments where the API server might not be available

## Swagger UI

Both servers also provide a Swagger UI for interactive API documentation:

- No Auth API: <http://localhost:8080/swagger>
- Basic Auth API: <http://localhost:8081/swagger>

## Health Check

The servers provide a simple health check endpoint:

- No Auth API: <http://localhost:8080/health>
- Basic Auth API: <http://localhost:8081/health>

## Stopping the Servers

To stop the API servers:

```bash
# Using mise
mise run demo-stop

# Or manually
docker-compose down
```

## Additional Demo Tasks

The project includes several mise tasks to help manage the demo:

```bash
# View the logs from the API servers
mise run demo-logs

# Restart the API servers
mise run demo-restart

# Check the status of the API servers
mise run demo-status
```