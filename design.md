# Book Inventory Service - Design Document

## 1. Background and Purpose

The service manages book inventory for a library. At its core is an Elasticsearch index named `books`, which supports the following operations:

- Adding, updating, and deleting individual books.
- Free-text search by title and author name.
- Filtering by price range and ebook availability.
- Statistical aggregations (count by ebook availability, median price, upper percentile, distinct author count).

---

## 2. Elasticsearch - Data Store Design

### 2.1 Index Mapping

| Field | ES Type | Additional Sub-field | Note |
|---|---|---|---|
| `title` | `text` | - | Free text, needed for title search in `/search` |
| `author_name` | `text` (custom analyzer, includes a synonym filter) | `author_name.keyword` | The text field supports free-text search and pen-name / real-name matching. The keyword sub-field supports the cardinality aggregation (distinct author count) and exact sorting |
| `price` | `float` | - | Needed for range queries and the percentiles aggregation |
| `ebook_available` | `boolean` | - | Optional in the incoming request, but always present in the indexed document: if not supplied by the client, the service sets it to `false` by default before indexing |
| `publish_date` | `date`, format `yyyy-MM-dd` | - | |

Example mapping as it will be applied in practice (JSON, for use in a future `PUT /books` call):

```json
{
  "mappings": {
    "properties": {
      "title": { "type": "text" },
      "author_name": {
        "type": "text",
        "analyzer": "author_synonym_analyzer",
        "fields": {
          "keyword": { "type": "keyword" }
        }
      },
      "price": { "type": "float" },
      "ebook_available": { "type": "boolean" },
      "publish_date": {
        "type": "date",
        "format": "yyyy-MM-dd"
      }
    }
  }
}
```

### 2.2 Key Design Decisions and Rationale

#### Keyword sub-field for author name

Needed to support the cardinality aggregation (`/store` route, distinct author count) and exact sorting. A `text` field is tokenized and therefore unsuitable for exact counting of unique values.

#### Default value for `ebook_available`

Rather than leaving the field missing when the client does not supply it, the service applies a default value of `false` at write time (in the API layer, before the document is sent to Elasticsearch, see Section 3.2). As a result, every indexed document always has `ebook_available` explicitly set to either `true` or `false`, and the field is never missing at the Elasticsearch level.

This keeps the query side simple: "books with an ebook" is a plain `term`/`match` on `true`, and "books without an ebook" is a plain `term`/`match` on `false`, with no need to also account for a missing-field case (see Section 2.3).

#### Case-sensitive search support

The default standard analyzer lowercases text, so a regular `match` query is not case-sensitive. To support case-sensitive search, the keyword sub-field (which is not tokenized or lowercased) is used together with a `term` query, which performs an exact match without analyzing the input. In contrast, a `match` query runs the search string through the same analyzer as the field, so it cannot distinguish case.

#### Synonyms (Mark Twain / Samuel Langhorne Clemens)

A search-time synonyms approach was chosen (synonyms applied via the `search_analyzer` at query time), rather than index-time synonyms.

Rationale: if another pen name is added for the same author in the future, the search-time approach allows updating the synonym list alone, without reindexing existing documents. With the index-time approach, documents already indexed are "frozen" with whatever synonym list was in effect at indexing time, and will not automatically pick up a newly added rule.

The index settings that define `author_synonym_analyzer`, referenced from the mapping in Section 2.1:

```json
{
  "settings": {
    "analysis": {
      "filter": {
        "author_synonym_filter": {
          "type": "synonym",
          "synonyms": [
            "<pen name>, <real name>"
          ]
        }
      },
      "analyzer": {
        "author_synonym_analyzer": {
          "type": "custom",
          "tokenizer": "standard",
          "filter": ["lowercase", "author_synonym_filter"]
        }
      }
    }
  }
}
```

The actual synonym pairs (e.g. `"mark twain, samuel langhorne clemens"`) are not hardcoded in the index settings above; they will be pulled from an external data source and used to populate this synonyms list at index-creation/update time. The mechanism (a search-time synonym filter) stays the same regardless of where the pairs come from.

### 2.3 Required Queries and Aggregations

| Exercise Requirement | Query/Aggregation Type |
|---|---|
| Books with an ebook available | `term`/`match` on `ebook_available: true` |
| Books without an ebook available | `term`/`match` on `ebook_available: false` |
| Books priced above $50 | `range` query on `price` (`gt: 50`) |
| Books priced $10-50 or $100-200 | `bool` with `should` of two `range` clauses |
| Simultaneous count: how many with ebook, how many without | `terms` aggregation on `ebook_available` |
| Median price and upper percentile | `percentiles` aggregation on `price` (`percents: 50, 90`) |
| Number of distinct authors | `cardinality` aggregation on `author_name.keyword` |
| Search by pen name/real name | `match` on `author_name` (with synonym analyzer) |
| Case-sensitive search | `term` on `author_name.keyword` |

---

## 3. API Service (Go)

### 3.1 Framework and Architecture

#### Framework

`gin-gonic` is used, as recommended in the exercise. It provides built-in JSON and query-parameter binding, straightforward method-based routing, and middleware support, all of which are used below.

#### Routing

All four operations on the books resource (POST, PUT, GET, DELETE) are dispatched to the same path, `/books`, distinguished purely by HTTP method. `/search` and `/store` are separate GET routes. `/activity` (Section 4) will be a separate GET route as well.

#### Layered architecture (Handler → Service → Repository)

The service is split into three layers, each with a single responsibility:

- **Handler** (`internal/handler`): parses each HTTP request (JSON body binding for POST/PUT, query-parameter binding for GET/DELETE), checks that required fields/parameters are present at all, calls the matching Service method, and translates the Service's result or error into an HTTP response using the unified envelope (Section 3.2).
- **Service** (`internal/service`): holds the business rules from Section 3.2, e.g. defaulting `ebook_available` to `false`, parsing and validating `price_range`, requiring at least one parameter on `/search`. It calls the Repository layer to read/write data, and has no knowledge of HTTP (no `gin.Context`, no status codes).
- **Repository** (`internal/repository`): the only layer that talks to Elasticsearch or Redis directly. It translates Service-level calls (e.g. "create this book", "get the last 3 actions for this user") into the actual client calls (Elasticsearch index/update/get/delete/search requests, Redis `LPUSH`/`LTRIM`/`EXPIRE`/`LRANGE`).

Proposed folder structure:

```text
book_service/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── handler/
│   │   ├── books_handler.go
│   │   ├── search_handler.go
│   │   ├── store_handler.go
│   │   └── activity_handler.go
│   ├── service/
│   │   ├── book_service.go
│   │   ├── search_service.go
│   │   ├── store_service.go
│   │   └── activity_service.go
│   ├── repository/
│   │   ├── elasticsearch/
│   │   │   └── book_repository.go
│   │   └── redis/
│   │       └── activity_repository.go
│   ├── middleware/
│   │   └── activity_middleware.go
│   ├── model/
│   │   └── book.go
│   └── router/
│       └── router.go
├── go.mod
└── go.sum
```

#### Dependency injection

Each layer depends on the layer below it through an interface, not a concrete implementation. For example, a `BookRepository` interface, implemented by an Elasticsearch-backed struct. Dependencies are wired once, by hand, in `main.go`: the Elasticsearch and Redis clients are created first, wrapped in repository implementations, then injected into services via constructor functions (e.g. `NewBookService(repo BookRepository) *BookService`), then injected into handlers the same way, before the handlers are registered with the router.

This is plain constructor-based dependency injection, not a DI framework (such as Wire or Fx). Rationale: the dependency graph here is small and static (a handful of services, each with one or two repositories), so a framework would add complexity without a real benefit; manual wiring in `main.go` stays easy to read and is idiomatic for a Go project of this size.

The practical payoff of injecting dependencies at all, rather than having each service construct its own repository internally, is testability: a test can build a `BookService` with a fake `BookRepository` and exercise the business rules from Section 3.2 without a real Elasticsearch instance running.

#### Generic activity-tracking middleware

Every route accepts a `username` parameter that is not used by any business logic in Section 3. It exists solely so that, in Section 4, each request can be attributed to a user for activity logging. To avoid reworking every handler later, a single middleware is planned that will run around all routes except `/activity` itself (per the requirement that activity lookups not count as activity), capturing the HTTP method, the path, and the username for each request.

Like the handlers and services, the middleware depends on the `ActivityRepository` interface (Section 4), injected the same way in `main.go`, rather than reaching for a Redis client directly. The middleware will exist as a placeholder in Section 3; the actual Redis write logic is implemented in Section 4.

### 3.2 Key Design Decisions and Rationale

#### Parameter transport

JSON body for POST and PUT, query string for GET and DELETE. Rationale: GET and DELETE are not expected to carry a semantically meaningful body under the HTTP standard, and many frameworks and proxies ignore or mishandle one; a single `id` (and `username`) parameter fits naturally into a query string. POST and PUT carry multiple structured fields, which are a natural fit for a JSON body.

#### `price_range` format

A single string such as `"10-50"`, parsed server-side into a lower and upper bound before being passed to Elasticsearch as a `range` query. An invalid format (e.g. missing the separator, or non-numeric bounds) results in a 400 response.

#### Unified response envelope

Every response, success or failure, is wrapped in the same top-level structure:

```json
{
  "success": true,
  "data": { },
  "error": null
}
```

On failure, `success` is `false`, `data` is `null`, and `error` is an object (not a plain string) with two fields:

```json
{
  "success": false,
  "data": null,
  "error": {
    "text": "book not found",
    "status_code": 404
  }
}
```

`error.text` is a human-readable message, and `error.status_code` mirrors the HTTP status code of the response, so a client can determine the outcome from the JSON body alone, without inspecting the HTTP status line. The actual HTTP status code (200, 201, 400, 404) is still set on the response itself; the envelope only standardizes the body shape so that any client can check success/error the same way regardless of endpoint.

#### ID generation on creation

`POST /books` does not accept a client-supplied id. Elasticsearch is left to auto-generate the document `_id`, which is then read back from the indexing response and returned to the caller. This differs from the earlier bulk-indexing exercise, where ids were assigned manually.

#### Partial update on PUT

`PUT /books` updates only the `title` field, via Elasticsearch's Update API (a partial doc update), rather than reindexing the full document. This preserves all other fields that were not sent in the request.

#### Validation

Each endpoint validates its required fields; a missing required field results in a 400 response naming the missing field. `ebook_available` remains optional in the `POST /books` request itself; if omitted, the handler sets it to `false` before indexing, per the Section 2.2 decision, so the field is always explicitly present in Elasticsearch.

#### Not-found handling

GET, PUT, and DELETE on `/books` return 404 (using the same unified error envelope) when the given id does not exist.

#### `username` as a forward-looking field

`username` is accepted and validated as present on every route, but is not used for any business logic in Section 3. It is reserved for the activity-tracking feature built in Section 4, and is captured generically via the middleware described in 3.1 rather than individually in each handler.

### 3.3 Endpoint Specifications

#### POST /books

Request (JSON body):

```http
POST /books
Content-Type: application/json
```

```json
{
  "title": "The Great Gatsby",
  "author_name": "F. Scott Fitzgerald",
  "price": 15.99,
  "ebook_available": true,
  "publish_date": "1925-04-10",
  "username": "alice"
}
```

Response 201:

```json
{
  "success": true,
  "data": {
    "id": "abc123"
  },
  "error": null
}
```

#### PUT /books

Request (JSON body):

```http
PUT /books
Content-Type: application/json
```

```json
{
  "id": "abc123",
  "title": "New Title",
  "username": "alice"
}
```

Response 200:

```json
{
  "success": true,
  "data": {
    "id": "abc123",
    "title": "New Title"
  },
  "error": null
}
```

Response 404 if `id` does not exist (unified error envelope).

#### GET /books

Request (query string):

```http
GET /books?id=abc123&username=alice
```

Response 200:

```json
{
  "success": true,
  "data": {
    "id": "abc123",
    "title": "The Great Gatsby",
    "author_name": "F. Scott Fitzgerald",
    "price": 15.99,
    "ebook_available": true,
    "publish_date": "1925-04-10"
  },
  "error": null
}
```

Response 404 if `id` does not exist.

#### DELETE /books

Request (query string):

```http
DELETE /books?id=abc123&username=alice
```

Response 200:

```json
{
  "success": true,
  "data": {
    "id": "abc123",
    "deleted": true
  },
  "error": null
}
```

Response 404 if `id` does not exist.

#### GET /search

Request (query string):

```http
GET /search?title=harry&author_name=rowling&price_range=10-50&username=alice
```

`title`, `author_name`, and `price_range` are all optional, but at least one must be provided.

Response 200:

```json
{
  "success": true,
  "data": {
    "total": 5,
    "results": [
      {
        "id": "abc123",
        "title": "...",
        "author_name": "...",
        "price": 0,
        "ebook_available": true,
        "publish_date": "..."
      }
    ]
  },
  "error": null
}
```

Response 400 if `price_range` is malformed, or if none of the three optional parameters were provided.

#### GET /store

Request (query string):

```http
GET /store?username=alice
```

Response 200:

```json
{
  "success": true,
  "data": {
    "total_books": 101,
    "distinct_authors": 87
  },
  "error": null
}
```

#### Unified error shape (all endpoints)

```json
{
  "success": false,
  "data": null,
  "error": {
    "text": "book not found",
    "status_code": 404
  }
}
```

---

## 4. Caching Layer (Redis)

### 4.1 Purpose and Scope

Redis is used exclusively to power the `/activity` route: tracking, per user, the most recent actions (HTTP method + route) they performed.

### 4.2 Data Model

- **Key:** `activity:<username>`, one Redis LIST per user.
- **Value:** each list entry is a plain string of the form `"<METHOD> <PATH>"`, e.g. `"POST /books"`, `"GET /search"`.
- **Insertion order:** new entries are pushed to the head of the list (`LPUSH`), so the most recent action is always at index 0.
- **List size:** after each push, the list is trimmed to exactly the 3 most recent entries (`LTRIM key 0 2`).
- **Expiration:** a 30-day TTL is set (and refreshed) on the key on every write, so the key expires only for a user who has gone 30 straight days without any activity. This is a garbage-collection mechanism for abandoned users.

### 4.3 Key Design Decisions and Rationale

#### Redis client library

`go-redis`, a widely used and actively maintained Go client for Redis, is used for the actual `LPUSH`/`LTRIM`/`EXPIRE` calls.

#### Integration with the request-handling middleware

As described in Section 3.1, a single middleware wraps every route except `/activity` itself. On each request, after the underlying handler completes, the middleware performs the `LPUSH` + `LTRIM` + `EXPIRE` sequence against `activity:<username>`, using the request's own method and path. `/activity` is explicitly excluded from this middleware, per the requirement that looking up activity not itself count as an action.

### 4.4 Endpoint Specification: GET /activity

Request:

```http
GET /activity?username=alice
```

Response 200:

```json
{
  "success": true,
  "data": {
    "actions": ["GET /search", "POST /books", "DELETE /books"]
  },
  "error": null
}
```

If the user has no recorded activity (e.g. a username that has never made a request), the response is still a 200, with an empty list:

```json
{
  "success": true,
  "data": { "actions": [] },
  "error": null
}
```
