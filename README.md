# Service Atlas Products
![Coverage](https://img.shields.io/badge/Coverage-74.4%25-brightgreen)

Service Atlas Products is the business semantic layer of the Service Atlas ecosystem. The core Service Atlas backend models technical service dependencies; this service organizes those dependencies into the language product teams use to describe business value.

At a high level, a **platform** is a collection of **products**. Each product can have **flows**, which describe how data moves through service nodes, and **capabilities**, which describe what the product can do.

## Core Concepts

- **Platform**: A top-level grouping of related products, such as a customer-facing application suite or an internal operations platform.
- **Product**: A cohesive business offering within a platform. Products belong to one platform and own their flows and capabilities.
- **Flow**: A named data path for a product. A flow represents how data moves through service nodes to support a business process.
- **Flow Step**: A directed edge inside a flow. Each step records a `current` service-node ID and a `next` service-node ID, and is validated against the Service Atlas engineering graph.
- **Capability**: Something a product can do, such as search, checkout, profile management, reporting, or notifications. Capabilities belong to a product.
- **Capability Step**: A link between a capability and a flow step. It captures the protocol and target, such as `HTTP` and `/api/boats`, that show where the capability is expressed in the data flow.

The relationship model is:

```text
Platform
  -> Product
      -> Flow
          -> Flow Step
      -> Capability
          -> Capability Step -> Flow Step
```

This allows the product layer to answer both product-oriented and graph-oriented questions:

- What products belong to this platform?
- What data flows support this product?
- Which service-node transitions make up this flow?
- What capabilities does this product provide?
- Which capabilities are present in a specific flow?

## How the API Works

The Bruno collection in `_bruno` is the working API reference. The local Bruno environment uses:

```text
baseUrl = http://localhost:8080
```

### Platforms

Platforms are the top-level containers.

| Action | Method | Path |
| --- | --- | --- |
| Create a platform | `POST` | `/platforms` |
| List all platforms | `GET` | `/platforms` |
| Get one platform | `GET` | `/platforms/{id}` |
| Update a platform | `PUT` | `/platforms/{id}` |
| Delete a platform | `DELETE` | `/platforms/{id}` |

Create request:

```json
{
  "name": "Platform 1",
  "description": "An Example Platform"
}
```

### Products

Products belong to platforms. A product is created with a `platform_id`, and products can be listed by platform.

| Action | Method | Path |
| --- | --- | --- |
| Create a product | `POST` | `/products` |
| Get one product | `GET` | `/products/{id}` |
| List products for a platform | `GET` | `/platforms/{id}/products` |
| Update a product | `PUT` | `/products/{id}` |
| Delete a product | `DELETE` | `/products/{id}` |

Create request:

```json
{
  "platform_id": 2,
  "name": "Product 1",
  "description": "Test Product"
}
```

### Flows

Flows belong to products. A flow names a data movement pattern, while its steps describe the service-node transitions that make up the path.

| Action | Method | Path |
| --- | --- | --- |
| Create a flow for a product | `POST` | `/products/{id}/flows` |
| List flows for a product | `GET` | `/products/{id}/flows` |
| Get one flow | `GET` | `/flows/{id}` |
| Update a flow | `PUT` | `/flows/{id}` |
| Delete a flow | `DELETE` | `/flows/{id}` |
| Add a flow step | `POST` | `/flows/{id}/steps` |
| List flow steps | `GET` | `/flows/{id}/steps` |
| Get a flow path | `GET` | `/flows/{id}/path` |
| Delete a flow step | `DELETE` | `/flow-steps/{id}` |

Create flow request:

```json
{
  "name": "Flow 1",
  "description": "flow 1 desc"
}
```

Create flow step request:

```json
{
  "flow_id": 1,
  "current": "cea050e5-ebc3-4d2b-aad2-877c47fa8961",
  "next": "75fce8b6-ca55-41dc-b5fb-1c12707887c0"
}
```

The `current` and `next` values are service-node IDs. Creating a flow step returns `422` when the required dependency relationship is not found in the Service Atlas backend.

### Capabilities

Capabilities belong to products and describe what the product can do. Capability steps connect a capability to specific flow steps, so the API can also list capabilities by flow.

| Action | Method | Path |
| --- | --- | --- |
| Create a capability | `POST` | `/capabilities` |
| Get one capability | `GET` | `/capabilities/{id}` |
| List capabilities for a product | `GET` | `/products/{id}/capabilities` |
| List capabilities for a flow | `GET` | `/flows/{id}/capabilities` |
| Update a capability | `PUT` | `/capabilities/{id}` |
| Delete a capability | `DELETE` | `/capabilities/{id}` |
| Add a capability step | `POST` | `/capability-steps` |
| List steps for a capability | `GET` | `/capabilities/{id}/steps` |
| Delete a capability step | `DELETE` | `/capability-steps/{id}` |

Create capability request:

```json
{
  "product_id": 1,
  "name": "a test capability",
  "description": "test"
}
```

Create capability step request:

```json
{
  "flow_step_id": 5,
  "capability_id": 1,
  "target": "/api/boats",
  "protocol": "HTTP"
}
```

## Example Workflow

1. Create a platform with `POST /platforms`.
2. Create a product in that platform with `POST /products`.
3. Create one or more flows for the product with `POST /products/{id}/flows`.
4. Add flow steps with `POST /flows/{id}/steps` to describe service-node transitions.
5. Create capabilities for the product with `POST /capabilities`.
6. Link capabilities to flow steps with `POST /capability-steps`.
7. Query the model from either direction:
   - `GET /products/{id}/flows` shows the flows for a product.
   - `GET /flows/{id}/path` shows the service-node path for a flow.
   - `GET /products/{id}/capabilities` shows what the product can do.
   - `GET /flows/{id}/capabilities` shows which capabilities appear in a flow.

## Architecture & Integration

Service Atlas Products runs as a standalone Go service with Postgres storage and integrates with the Service Atlas backend for graph validation.

- **Source of truth**: Platforms, products, flows, flow steps, capabilities, and capability steps are stored in Postgres.
- **Graph validation**: Flow steps are validated against the engineering graph so product-level flows correspond to real service relationships.
- **Frontend integration**: When `PRODUCTS_SERVICE_URL` is configured, the Service Atlas UI can enable product-layer features that connect engineering health to business impact.

## Development Setup

### Prerequisites

- Go 1.26+
- Postgres
- Liquibase
- sqlc
- Just
- Access to a running Service Atlas backend instance for flow-step validation

### Running the Service

1. Install dependencies:

   ```bash
   go mod download
   ```

2. Start the local database:

   ```bash
   just up
   ```

3. Apply migrations:

   ```bash
   just lbupdate
   ```

4. Start the server:

   ```bash
   go run main.go
   ```

The service listens on port `8080` by default, configurable with the `PORT` environment variable.

## Database & Tooling

- **Just**: Runs common project tasks. Use `just --list` to see available recipes.
- **Liquibase**: Manages database migrations from `migrations/changelog.yml`.
- **sqlc**: Generates type-safe Go database code from SQL queries.

Useful commands:

```bash
just test
just test-full
just generate_schema
just lbupdate
just lint
just vet
```

When making schema changes, update the Liquibase changelog first, regenerate `schema.sql`, and then regenerate sqlc output with:

```bash
just generate_schema
```

For changes that should not be included in the generated `schema.sql`, add `context: "!sqlc"` to the Liquibase changeset.

---

Inspired by [this RFC](https://github.com/service-atlas/services/discussions/179).
