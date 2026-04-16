# Service Atlas Products

> [!WARNING]
> **Status:** Draft — Open for Discussion | **Version:** 0.2.0 | **Date:** 2026-03-30
> **Work in Progress:** This project is currently under active development and follows the specifications outlined in the RFC below.

Service Atlas Products is the business semantic layer of the Service Atlas ecosystem. While the core backend models the technical dependencies of services, this project maps those interactions to real-world business value through **Platforms**, **Products**, and **Flows**.

## Core Concepts

The project introduces a hierarchical model to organize the service graph from a product perspective:

- **Platform**: The top-level grouping of related business offerings (e.g., "Consumer App", "Internal Tools").
- **Product**: A cohesive set of capabilities within a platform (e.g., "Shopping Cart", "User Profile").
- **Flow**: A named, human-readable path of data moving through specific services. It answers: *"What is this chain of interactions called, and what business purpose does it serve?"*
- **Flow Step**: An individual link in a Flow, validated against `data` typed dependencies in the Service Atlas engineering graph.


## Architecture & Integration

Service Atlas Products operates as a standalone service with its own relational storage (**Postgres**), while integrating with the Service Atlas backend (**Neo4j**).

- **Source of Truth**: Business definitions (Platforms, Products, Flows) live here in Postgres.
- **Validation**: Every Flow Step is verified against the engineering graph in the backend to ensure the represented data interaction actually exists.
- **Frontend Integration**: When `PRODUCTS_SERVICE_URL` is configured, the Service Atlas UI enables the product-layer features, bridging engineering health with business impact.

## Development Setup

### Prerequisites
- **Go 1.26+**
- **Postgres** (Schema definitions coming soon)
- Access to a running **Service Atlas Backend** instance

### Running the Service
1. Clone the repository and install dependencies:
   ```bash
   go mod download
   ```
2. Start the server:
   ```bash
   go run main.go
   ```
   The service listens on port `8080` by default (configurable via `PORT` environment variable).

## Roadmap

- [ ] Implementation of Platform, Product, and Flow CRUD operations.
- [ ] Integration with Service Atlas Backend for `data` dependency validation.
- [ ] Risk propagation: Highlighting business flows affected by technical debt.
- [ ] Broken flow detection: Automated alerts when underlying service dependencies change.

---
*Inspired by [this RFC](https://github.com/service-atlas/services/discussions/179)*
