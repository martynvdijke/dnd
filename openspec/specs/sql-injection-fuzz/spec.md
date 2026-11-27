# sql-injection-fuzz Specification

## Purpose
Test that all SQL parameterized queries handle malicious input safely.

## Requirements

### Requirement: String field SQL injection resistance

Every string field in every POST/PUT endpoint SHALL be tested with a set of known SQL injection payloads to verify parameterized queries are used.

#### Scenario: Character name with SQL injection payload does not execute
- **WHEN** POST /api/characters is called with name="'; DROP TABLE characters; --"
- **THEN** the character SHALL be created with that literal name, no tables SHALL be dropped

#### Scenario: Campaign name with UNION injection returns 400 or 200
- **WHEN** POST /api/campaigns is called with name="' UNION SELECT * FROM users--"
- **THEN** the campaign SHALL be created with that literal name, no user data SHALL be leaked

#### Scenario: Search query with LIKE wildcard injection handles correctly
- **WHEN** GET /api/compendium/search is called with q="%' OR '1'='1"
- **THEN** response SHALL not return more results than an equivalent non-injection query

### Requirement: Numeric field injection resistance

Numeric fields in request bodies SHALL be tested for injection via string values.

#### Scenario: String value for numeric field returns 400
- **WHEN** POST /api/characters is called with str="'; DROP TABLE--"
- **THEN** response code SHALL be 400 (JSON parse error or validation error)

### Requirement: Large input handling

String fields SHALL be tested with oversized inputs to verify bounds checking.

#### Scenario: 10KB character name does not crash server
- **WHEN** POST /api/characters is called with name of length 10000
- **THEN** response code SHALL be 400 (rejected by validation) or 201 (truncated to field limit)

#### Scenario: Unicode BOM in search query is handled
- **WHEN** GET /api/compendium/search is called with q="\xEF\xBB\xBFfire"
- **THEN** response code SHALL be 200 (BOM SHALL be transparently handled)
