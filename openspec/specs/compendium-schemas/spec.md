# compendium-schemas Specification

## Purpose
Dynamic schema definitions for compendium entry types — define fields, types, labels, validation rules per type without code changes.

## Requirements

### Requirement: Dynamic schema definitions for compendium types
The system SHALL support defining compendium entry types dynamically through a schema system that stores field definitions in the database. Each schema defines the structure of a compendium type (e.g., "Magic Item", "Monster", "Spell") without requiring Go code changes.

#### Scenario: Define a new compendium type
- **WHEN** an admin submits a new schema definition with a unique type name and a set of field definitions (name, type, label, required flag, default value)
- **THEN** the schema is persisted and the new compendium type is immediately available for creating entries

#### Scenario: Schema contains typed fields
- **WHEN** an admin defines fields with types (string, text, integer, float, boolean, json, select, multi-select)
- **THEN** the system stores the field type and uses it for input validation and rendering

#### Scenario: System schemas are seeded
- **WHEN** the application starts for the first time
- **THEN** built-in schemas are seeded for legacy types: races, classes, spells, feats, backgrounds, equipment, monsters

### Requirement: List and inspect schemas
The system SHALL provide an API to list all compendium schemas and view individual schema details including all field definitions.

#### Scenario: List all schemas
- **WHEN** a user requests GET /api/compendium/schemas
- **THEN** the system returns a list of all defined schemas with their type name, field count, and entry count

#### Scenario: View schema detail
- **WHEN** a user requests GET /api/compendium/schemas/:type
- **THEN** the system returns the full schema with all field definitions

### Requirement: Update and delete schemas
Admins SHALL be able to update schema metadata and field definitions, and delete unused schemas.

#### Scenario: Update schema fields
- **WHEN** an admin updates a schema's field definitions
- **THEN** the schema is updated; entries already using the schema remain accessible with their existing data

#### Scenario: Delete unused schema
- **WHEN** an admin deletes a schema that has no entries
- **THEN** the schema is permanently removed

#### Scenario: Prevent deletion of non-empty schema
- **WHEN** an admin attempts to delete a schema that has entries
- **THEN** the system rejects the deletion with an error message indicating the entry count
