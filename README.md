# Service Atlas Products

![Coverage](https://img.shields.io/badge/Coverage-74.4%25-brightgreen)

Service Atlas Products is the business semantic layer of the Service Atlas ecosystem. While the core Service Atlas backend models technical service dependencies between infrastructure nodes, this service organizes those dependencies into the language of products, teams, and business value.

By bridging the gap between technical service graphs and business offerings, this API allows teams to answer critical questions:
- What products belong to this platform?
- What business flows support a specific product?
- Which service-node transitions make up a flow?
- What business capabilities (e.g., "Checkout", "Search") are provided by a product?
- Which capabilities are affected by a specific technical service dependency?

## Core Concepts

- **Platform**: A top-level grouping of related products (e.g., "Retail Banking", "Internal Operations").
- **Product**: A cohesive business offering within a platform (e.g., "Mobile App", "Lending API").
- **Flow**: A named data path through Service Atlas nodes that supports a product's business process.
- **Flow Step**: A directed edge in a flow, linking two service-node IDs (`current` and `next`). Flow steps must correspond to real technical dependencies in the Service Atlas graph.
- **Capability**: A specific function a product provides (e.g., "User Login", "Order Fulfillment").
- **Capability Step**: A link between a capability and a flow step, specifying the `protocol` and `target` (e.g., `HTTP` / `/api/v1/orders`).

## Entity Relationship Model

```text
Platform
  -> Product
      -> Flow
          -> Flow Step
      -> Capability
          -> Capability Step -> Flow Step
```

## Example Workflow: "Boat Rental Service"

Imagine a "Boating Platform" with a "Boat Rental" product.

1.  **Platform**: Create "Boating Platform".
2.  **Product**: Create "Boat Rental" under the platform.
3.  **Flow**: Define a "Reserve Boat" flow.
4.  **Flow Step**: Add a step from `frontend-service` to `booking-service`.
    - *Note: This transition must exist in the Service Atlas technical graph as a 'data' interaction.*
5.  **Capability**: Define a "Search Boats" capability for the product.
6.  **Capability Step**: Link "Search Boats" to the flow step above, targeting `/api/boats` via `HTTP`.

## API Reference

The service listens on port `8080` by default.

### Platforms

| Action | Method | Path |
| --- | --- | --- |
| Create a platform | `POST` | `/platforms` |
| List all platforms | `GET` | `/platforms` |
| Get one platform | `GET` | `/platforms/{id}` |
| Update a platform | `PUT` | `/platforms/{id}` |
| Delete a platform | `DELETE` | `/platforms/{id}` |

**Create Request (`POST /platforms`):**
```json
{
  "name": "Boating Platform",
  "description": "All boat-related services"
}
```
*Required: `name`. Optional: `description`.*

---

### Products

| Action | Method | Path |
| --- | --- | --- |
| Create a product | `POST` | `/products` |
| Get one product | `GET` | `/products/{id}` |
| List products for a platform | `GET` | `/platforms/{platform_id}/products` |
| Update a product | `PUT` | `/products/{id}` |
| Delete a product | `DELETE` | `/products/{id}` |

**Create Request (`POST /products`):**
```json
{
  "platform_id": 1,
  "name": "Boat Rental",
  "description": "Consumer boat rental application"
}
```
*Required: `platform_id`, `name`. Optional: `description`.*

---

### Flows & Flow Steps

Flows describe business paths. **Flow steps are validated** against the Service Atlas backend. A step from `current` to `next` is valid only if a "data" interaction exists between those nodes in the engineering graph.

| Action | Method | Path |
| --- | --- | --- |
| Create a flow | `POST` | `/products/{product_id}/flows` |
| List flows for a product | `GET` | `/products/{product_id}/flows` |
| Get one flow | `GET` | `/flows/{id}` |
| Update a flow | `PUT` | `/flows/{id}` |
| Delete a flow | `DELETE` | `/flows/{id}` |
| Add a flow step | `POST` | `/flows/{flow_id}/steps` |
| List flow steps | `GET` | `/flows/{flow_id}/steps` |
| Get full flow path | `GET` | `/flows/{flow_id}/path` |
| Delete a flow step | `DELETE` | `/flow-steps/{id}` |

**Create Flow Step (`POST /flows/{id}/steps`):**
```json
{
  "current": "cea050e5-ebc3-4d2b-aad2-877c47fa8961",
  "next": "75fce8b6-ca55-41dc-b5fb-1c12707887c0"
}
```
*Required: `current`, `next` (Service-node UUIDs).*
*Returns `422 Unprocessable Entity` if the dependency is not found in Service Atlas.*

---

### Capabilities & Capability Steps

| Action | Method | Path |
| --- | --- | --- |
| Create a capability | `POST` | `/capabilities` |
| Get one capability | `GET` | `/capabilities/{id}` |
| List capabilities for a product | `GET` | `/products/{product_id}/capabilities` |
| List capabilities for a flow | `GET` | `/flows/{flow_id}/capabilities` |
| Update a capability | `PUT` | `/capabilities/{id}` |
| Delete a capability | `DELETE` | `/capabilities/{id}` |
| Add a capability step | `POST` | `/capability-steps` |
| List steps for a capability | `GET` | `/capabilities/{id}/steps` |
| Delete a capability step | `DELETE` | `/capability-steps/{id}` |

**Create Capability Step (`POST /capability-steps`):**
```json
{
  "capability_id": 1,
  "flow_step_id": 5,
  "protocol": "HTTP",
  "target": "/api/boats"
}
```
*Required: `capability_id`, `flow_step_id`, `protocol`, `target`.*

## Configuration & Environment

| Variable | Description | Default |
| --- | --- | --- |
| `ADDRESS` | The address/port the service binds to. | `:8080` |
| `SERVICE_URL` | **Backend Validation URL**. Used to verify flow steps against the Service Atlas technical graph. The service calls `{SERVICE_URL}/services/{current}/dependencies`. | (Required) |
| `DB_URL` | Postgres connection string. Used as a fallback if the Secrets Provider is not configured or fails to provide a URL. | - |
| `SECRETS_STRATEGY` | Strategy for the Secrets Provider (e.g., `env`, `vault`). | `env` |

## Development Setup

### Prerequisites
- Go 1.26+
- Postgres
- [Secrets Provider](https://github.com/service-atlas/secrets-provider) (Integrated for DB credentials)
- [Just](https://github.com/casey/just) (Command runner)
- [Liquibase](https://www.liquibase.org/) (Database migrations)
- [sqlc](https://sqlc.dev/) (Go code generation from SQL)

### Running Locally
1. **Install dependencies**: `go mod download`
2. **Start Database**: `just up` (requires Docker)
3. **Run Migrations**: `just lbupdate`
4. **Environment**:
    - Set `SERVICE_URL` to a running Service Atlas backend.
    - The service uses `secrets-provider` for database info. By default, it looks for `DB_URL`, `DB_USERNAME`, and `DB_PASSWORD` in the environment.
5. **Start Service**: `go run main.go`

### Tooling
- `just test`: Run standard tests.
- `just generate_schema`: Regenerate `schema.sql` and `sqlc` code after migration changes.
- `_bruno/`: Contains a Bruno collection for interactive API testing. Note: This README is the primary human-readable guide.

---
Inspired by [this RFC](https://github.com/service-atlas/services/discussions/179).
