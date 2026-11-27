# prepared-statement-caching Specification

## Purpose
Cache prepared statements for frequent ent ORM query shapes.

## Requirements

### Requirement: Enable modernc driver statement cache
The system SHALL enable the modernc.org/sqlite driver's built-in statement cache via DSN parameter.

#### Scenario: Statement cache enabled in DSN
- **WHEN** the database connection string is constructed
- **THEN** it SHALL include the `_pragma` parameter to enable statement caching (if supported by modernc)
- **OR** if the driver does not expose this, a connection-level cache SHALL be implemented

### Requirement: Add connection-level prepared statement cache
The system SHALL implement a prepared statement cache with LRU eviction (max 1000 statements) if the driver does not provide built-in caching.

#### Scenario: Prepared statement cache created
- **WHEN** the database connection is initialized
- **THEN** a prepared statement cache SHALL be initialized with a maximum size of 1000 entries

#### Scenario: Cache hit reuses prepared statement
- **WHEN** an SQL query shape matches a previously cached prepared statement
- **THEN** the cached statement SHALL be reused instead of preparing a new one

#### Scenario: Cache eviction on overflow
- **WHEN** the cache exceeds 1000 entries
- **THEN** the least recently used entry SHALL be evicted

### Requirement: Cache statistics logging
The system SHALL log prepared statement cache hit/miss statistics every 60 minutes.

#### Scenario: Cache stats logged
- **WHEN** 60 minutes have elapsed
- **THEN** a log entry SHALL contain cached statement count, hit count, and miss count
