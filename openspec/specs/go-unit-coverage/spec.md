# go-unit-coverage Specification

## Purpose
Backend Go handler unit tests for all registered API routes.

## Requirements

### Requirement: Admin handler tests
The system SHALL have Go unit tests for all admin handler functions in `handlers/admin.go`.

#### Scenario: Admin list users returns 200
- **WHEN** authenticated admin user calls GET /api/admin/users
- **THEN** response status SHALL be 200 and body SHALL contain a JSON array of users

#### Scenario: Admin create user returns 201
- **WHEN** authenticated admin user calls POST /api/admin/users with valid user data
- **THEN** response status SHALL be 201 and body SHALL contain the created user

#### Scenario: Admin create user returns 400 for missing fields
- **WHEN** authenticated admin user calls POST /api/admin/users with missing required fields
- **THEN** response status SHALL be 400

#### Scenario: Admin update user returns 200
- **WHEN** authenticated admin user calls PUT /api/admin/users/:id with valid data
- **THEN** response status SHALL be 200

#### Scenario: Admin delete user returns 200
- **WHEN** authenticated admin user calls DELETE /api/admin/users/:id
- **THEN** response status SHALL be 200

#### Scenario: Admin reset password returns 200
- **WHEN** authenticated admin user calls PUT /api/admin/users/:id/password
- **THEN** response status SHALL be 200

#### Scenario: Admin email settings CRUD returns 200
- **WHEN** authenticated admin user calls GET /api/admin/email-settings
- **THEN** response status SHALL be 200

#### Scenario: Admin create compendium entry returns 201
- **WHEN** authenticated admin user calls POST /api/admin/compendium/:type with valid data
- **THEN** response status SHALL be 201

#### Scenario: Admin update compendium entry returns 200
- **WHEN** authenticated admin user calls PUT /api/admin/compendium/:type/:id with valid data
- **THEN** response status SHALL be 200

#### Scenario: Admin delete compendium entry returns 200
- **WHEN** authenticated admin user calls DELETE /api/admin/compendium/:type/:id
- **THEN** response status SHALL be 200

### Requirement: Calendar handler tests
The system SHALL have Go unit tests for calendar event handlers.

#### Scenario: List calendar events returns 200
- **WHEN** authenticated user calls GET /api/calendar
- **THEN** response status SHALL be 200

#### Scenario: Create calendar event returns 201
- **WHEN** authenticated user calls POST /api/calendar with valid event data
- **THEN** response status SHALL be 201

#### Scenario: Update calendar event returns 200
- **WHEN** authenticated user calls PUT /api/calendar/:id with valid data
- **THEN** response status SHALL be 200

#### Scenario: Delete calendar event returns 200
- **WHEN** authenticated user calls DELETE /api/calendar/:id
- **THEN** response status SHALL be 200

### Requirement: Timeline handler tests
The system SHALL have Go unit tests for timeline event handlers.

#### Scenario: List timeline events returns 200
- **WHEN** authenticated user calls GET /api/timeline
- **THEN** response status SHALL be 200

#### Scenario: Create timeline event returns 201
- **WHEN** authenticated user calls POST /api/timeline with valid data
- **THEN** response status SHALL be 201

#### Scenario: Update timeline event returns 200
- **WHEN** authenticated user calls PUT /api/timeline/:id with valid data
- **THEN** response status SHALL be 200

#### Scenario: Delete timeline event returns 200
- **WHEN** authenticated user calls DELETE /api/timeline/:id
- **THEN** response status SHALL be 200

### Requirement: Feat handler tests
The system SHALL have Go unit tests for feat handlers.

#### Scenario: List feats returns 200
- **WHEN** authenticated user calls GET /api/feats
- **THEN** response status SHALL be 200

#### Scenario: Create feat returns 201
- **WHEN** authenticated user calls POST /api/feats with valid data
- **THEN** response status SHALL be 201

#### Scenario: Update feat returns 200
- **WHEN** authenticated user calls PUT /api/feats/:id with valid data
- **THEN** response status SHALL be 200

#### Scenario: Delete feat returns 200
- **WHEN** authenticated user calls DELETE /api/feats/:id
- **THEN** response status SHALL be 200

### Requirement: Companion handler tests
The system SHALL have Go unit tests for companion handlers.

#### Scenario: List companions returns 200
- **WHEN** authenticated user calls GET /api/companions
- **THEN** response status SHALL be 200

#### Scenario: Create companion returns 201
- **WHEN** authenticated user calls POST /api/companions with valid data
- **THEN** response status SHALL be 201

#### Scenario: Update companion returns 200
- **WHEN** authenticated user calls PUT /api/companions/:id with valid data
- **THEN** response status SHALL be 200

#### Scenario: Delete companion returns 200
- **WHEN** authenticated user calls DELETE /api/companions/:id
- **THEN** response status SHALL be 200

### Requirement: Downtime activity handler tests
The system SHALL have Go unit tests for downtime activity handlers.

#### Scenario: List downtime activities returns 200
- **WHEN** authenticated user calls GET /api/characters/:id/downtime
- **THEN** response status SHALL be 200

#### Scenario: Create downtime activity returns 201
- **WHEN** authenticated user calls POST /api/characters/:id/downtime with valid data
- **THEN** response status SHALL be 201

#### Scenario: Update downtime activity returns 200
- **WHEN** authenticated user calls PUT /api/downtime/:id with valid data
- **THEN** response status SHALL be 200

#### Scenario: Delete downtime activity returns 200
- **WHEN** authenticated user calls DELETE /api/downtime/:id
- **THEN** response status SHALL be 200

#### Scenario: Advance downtime day returns 200
- **WHEN** authenticated user calls POST /api/downtime/:id/advance
- **THEN** response status SHALL be 200

#### Scenario: Get downtime types returns 200
- **WHEN** authenticated user calls GET /api/downtime/types
- **THEN** response status SHALL be 200

### Requirement: Level planner handler tests
The system SHALL have Go unit tests for level-up planner handlers.

#### Scenario: Get level-up plan returns 200
- **WHEN** authenticated user calls GET /api/characters/:id/level-plan
- **THEN** response status SHALL be 200

#### Scenario: Save level-up plan returns 200
- **WHEN** authenticated user calls POST /api/characters/:id/level-plan with valid data
- **THEN** response status SHALL be 200

#### Scenario: Delete level-up plan returns 200
- **WHEN** authenticated user calls DELETE /api/characters/:id/level-plan
- **THEN** response status SHALL be 200

#### Scenario: Get level-up suggestions returns 200
- **WHEN** authenticated user calls GET /api/characters/:id/level-suggestions
- **THEN** response status SHALL be 200

### Requirement: Recap handler tests
The system SHALL have Go unit tests for campaign recap handlers.

#### Scenario: List campaign recaps returns 200
- **WHEN** authenticated user calls GET /api/campaigns/:id/recaps
- **THEN** response status SHALL be 200

#### Scenario: Create campaign recap returns 201
- **WHEN** authenticated user calls POST /api/campaigns/:id/recaps with valid data
- **THEN** response status SHALL be 201

#### Scenario: Get campaign recap returns 200
- **WHEN** authenticated user calls GET /api/recaps/:id
- **THEN** response status SHALL be 200

#### Scenario: Update campaign recap returns 200
- **WHEN** authenticated user calls PUT /api/recaps/:id with valid data
- **THEN** response status SHALL be 200

#### Scenario: Delete campaign recap returns 200
- **WHEN** authenticated user calls DELETE /api/recaps/:id
- **THEN** response status SHALL be 200

#### Scenario: Generate campaign recap returns 200
- **WHEN** authenticated user calls POST /api/campaigns/:id/recaps/generate
- **THEN** response status SHALL be 200

### Requirement: Resource handler tests
The system SHALL have Go unit tests for character resource handlers.

#### Scenario: List character resources returns 200
- **WHEN** authenticated user calls GET /api/characters/:id/resources
- **THEN** response status SHALL be 200

#### Scenario: Create character resource returns 201
- **WHEN** authenticated user calls POST /api/characters/:id/resources with valid data
- **THEN** response status SHALL be 201

#### Scenario: Update resource returns 200
- **WHEN** authenticated user calls PUT /api/resources/:id with valid data
- **THEN** response status SHALL be 200

#### Scenario: Delete resource returns 200
- **WHEN** authenticated user calls DELETE /api/resources/:id
- **THEN** response status SHALL be 200

#### Scenario: Recover resources on rest returns 200
- **WHEN** authenticated user calls POST /api/characters/:id/recover-resources
- **THEN** response status SHALL be 200

### Requirement: Combat log handler tests
The system SHALL have Go unit tests for combat log handlers.

#### Scenario: List combat log entries returns 200
- **WHEN** authenticated user calls GET /api/combat-log
- **THEN** response status SHALL be 200

#### Scenario: Create combat log entry returns 201
- **WHEN** authenticated user calls POST /api/combat-log with valid data
- **THEN** response status SHALL be 201

#### Scenario: Get combat log stats returns 200
- **WHEN** authenticated user calls GET /api/combat-log/stats
- **THEN** response status SHALL be 200

### Requirement: Wiki handler tests
The system SHALL have Go unit tests for wiki page handlers.

#### Scenario: List wiki pages returns 200
- **WHEN** authenticated user calls GET /api/campaigns/:id/wiki
- **THEN** response status SHALL be 200

#### Scenario: Create wiki page returns 201
- **WHEN** authenticated user calls POST /api/campaigns/:id/wiki with valid data
- **THEN** response status SHALL be 201

#### Scenario: Get wiki page returns 200
- **WHEN** authenticated user calls GET /api/wiki/:id
- **THEN** response status SHALL be 200

#### Scenario: Update wiki page returns 200
- **WHEN** authenticated user calls PUT /api/wiki/:id with valid data
- **THEN** response status SHALL be 200

#### Scenario: Delete wiki page returns 200
- **WHEN** authenticated user calls DELETE /api/wiki/:id
- **THEN** response status SHALL be 200

### Requirement: Compendium admin handler tests (dynamic schemas)
The system SHALL have Go unit tests for dynamic compendium schema CRUD and entry management.

#### Scenario: List compendium schemas returns 200
- **WHEN** authenticated admin user calls GET /api/admin/compendium-schemas
- **THEN** response status SHALL be 200 and body SHALL contain an array of schemas

#### Scenario: Create compendium schema returns 201
- **WHEN** authenticated admin user calls POST /api/admin/compendium-schemas with valid schema data
- **THEN** response status SHALL be 201

#### Scenario: Get compendium schema returns 200
- **WHEN** authenticated admin user calls GET /api/admin/compendium-schemas/:id
- **THEN** response status SHALL be 200

#### Scenario: Update compendium schema returns 200
- **WHEN** authenticated admin user calls PUT /api/admin/compendium-schemas/:id with valid data
- **THEN** response status SHALL be 200

#### Scenario: Delete compendium schema returns 200
- **WHEN** authenticated admin user calls DELETE /api/admin/compendium-schemas/:id
- **THEN** response status SHALL be 200

#### Scenario: List compendium entries returns 200
- **WHEN** authenticated admin user calls GET /api/admin/compendium-schemas/:schemaId/entries
- **THEN** response status SHALL be 200

#### Scenario: Create compendium entry returns 201
- **WHEN** authenticated admin user calls POST /api/admin/compendium-schemas/:schemaId/entries with valid data
- **THEN** response status SHALL be 201

#### Scenario: Get compendium entry returns 200
- **WHEN** authenticated admin user calls GET /api/admin/compendium-entries/:id
- **THEN** response status SHALL be 200

#### Scenario: Update compendium entry returns 200
- **WHEN** authenticated admin user calls PUT /api/admin/compendium-entries/:id with valid data
- **THEN** response status SHALL be 200

#### Scenario: Delete compendium entry returns 200
- **WHEN** authenticated admin user calls DELETE /api/admin/compendium-entries/:id
- **THEN** response status SHALL be 200

#### Scenario: Batch delete entries returns 200
- **WHEN** authenticated admin user calls POST /api/admin/compendium-entries/batch-delete with IDs array
- **THEN** response status SHALL be 200

#### Scenario: Batch update entries returns 200
- **WHEN** authenticated admin user calls POST /api/admin/compendium-entries/batch-update with IDs and data
- **THEN** response status SHALL be 200

#### Scenario: Cross-schema search returns 200
- **WHEN** authenticated admin user calls GET /api/admin/compendium-search?q=query
- **THEN** response status SHALL be 200

#### Scenario: Import compendium entries returns 200
- **WHEN** authenticated admin user calls POST /api/admin/compendium-schemas/:schemaId/import
- **THEN** response status SHALL be 200

#### Scenario: List import logs returns 200
- **WHEN** authenticated admin user calls GET /api/admin/compendium-import-logs
- **THEN** response status SHALL be 200

#### Scenario: Export compendium entries returns 200
- **WHEN** authenticated admin user calls GET /api/admin/compendium-schemas/:schemaId/export
- **THEN** response status SHALL be 200

### Requirement: Monster management handler tests
The system SHALL have Go unit tests for custom monster library and campaign monster roster handlers.

#### Scenario: List monster library returns 200
- **WHEN** authenticated DM user calls appropriate monster library endpoint
- **THEN** response status SHALL be 200

#### Scenario: Create monster library entry returns 201
- **WHEN** authenticated DM user calls monster library create with valid data
- **THEN** response status SHALL be 201

#### Scenario: Update monster library entry returns 200
- **WHEN** authenticated DM user calls monster library update with valid data
- **THEN** response status SHALL be 200

#### Scenario: Delete monster library entry returns 200
- **WHEN** authenticated DM user calls monster library delete
- **THEN** response status SHALL be 200

#### Scenario: Campaign monster roster list returns 200
- **WHEN** authenticated DM user calls campaign monster roster endpoint
- **THEN** response status SHALL be 200

#### Scenario: Add monster to campaign roster returns 200
- **WHEN** authenticated DM user calls add to campaign roster endpoint
- **THEN** response status SHALL be 200

#### Scenario: Remove monster from campaign roster returns 200
- **WHEN** authenticated DM user calls remove from campaign roster endpoint
- **THEN** response status SHALL be 200

### Requirement: One-shot adventure handler tests (extended)
The system SHALL have Go unit tests for remaining one-shot handlers (acts, scenes, pacing, clues, prep checklist, DM notes).

#### Scenario: Create one-shot act returns 201
- **WHEN** authenticated user calls POST /api/oneshot-adventures/:id/acts with valid data
- **THEN** response status SHALL be 201

#### Scenario: Update one-shot act returns 200
- **WHEN** authenticated user calls PUT /api/oneshot-acts/:id with valid data
- **THEN** response status SHALL be 200

#### Scenario: Delete one-shot act returns 200
- **WHEN** authenticated user calls DELETE /api/oneshot-acts/:id
- **THEN** response status SHALL be 200

#### Scenario: Create one-shot scene returns 201
- **WHEN** authenticated user calls POST /api/oneshot-acts/:id/scenes with valid data
- **THEN** response status SHALL be 201

#### Scenario: Update one-shot scene returns 200
- **WHEN** authenticated user calls PUT /api/oneshot-scenes/:id with valid data
- **THEN** response status SHALL be 200

#### Scenario: Delete one-shot scene returns 200
- **WHEN** authenticated user calls DELETE /api/oneshot-scenes/:id
- **THEN** response status SHALL be 200

#### Scenario: Start pacing session returns 200
- **WHEN** authenticated user calls POST /api/oneshot-adventures/:id/pacing/start
- **THEN** response status SHALL be 200

#### Scenario: Get pacing session returns 200
- **WHEN** authenticated user calls GET /api/oneshot-adventures/:id/pacing
- **THEN** response status SHALL be 200

#### Scenario: Pause pacing session returns 200
- **WHEN** authenticated user calls POST /api/session-pacing/:id/pause
- **THEN** response status SHALL be 200

#### Scenario: Resume pacing session returns 200
- **WHEN** authenticated user calls POST /api/session-pacing/:id/resume
- **THEN** response status SHALL be 200

#### Scenario: Advance to next scene returns 200
- **WHEN** authenticated user calls POST /api/session-pacing/:id/next-scene
- **THEN** response status SHALL be 200

#### Scenario: Complete pacing session returns 200
- **WHEN** authenticated user calls POST /api/session-pacing/:id/complete
- **THEN** response status SHALL be 200

#### Scenario: Update pacing timers returns 200
- **WHEN** authenticated user calls POST /api/session-pacing/:id/tick
- **THEN** response status SHALL be 200

#### Scenario: List prep checklist returns 200
- **WHEN** authenticated user calls GET /api/oneshot-adventures/:id/checklist
- **THEN** response status SHALL be 200

#### Scenario: Create prep checklist item returns 201
- **WHEN** authenticated user calls POST /api/oneshot-adventures/:id/checklist with valid data
- **THEN** response status SHALL be 201

#### Scenario: Update prep checklist item returns 200
- **WHEN** authenticated user calls PUT /api/prep-checklist/:cid with valid data
- **THEN** response status SHALL be 200

#### Scenario: Delete prep checklist item returns 200
- **WHEN** authenticated user calls DELETE /api/prep-checklist/:cid
- **THEN** response status SHALL be 200

#### Scenario: List DM notes returns 200
- **WHEN** authenticated user calls GET /api/oneshot-adventures/:id/notes
- **THEN** response status SHALL be 200

#### Scenario: Create DM note returns 201
- **WHEN** authenticated user calls POST /api/oneshot-adventures/:id/notes with valid data
- **THEN** response status SHALL be 201

#### Scenario: Update DM note returns 200
- **WHEN** authenticated user calls PUT /api/dm-notes/:nid with valid data
- **THEN** response status SHALL be 200

#### Scenario: Delete DM note returns 200
- **WHEN** authenticated user calls DELETE /api/oneshot-adventures/:id/notes/:nid
- **THEN** response status SHALL be 200

### Requirement: Player compendium access handler tests
The system SHALL have Go unit tests for the public player compendium read-only API.

#### Scenario: List player-accessible schemes returns 200
- **WHEN** authenticated user calls GET /api/compendium/schemas
- **THEN** response status SHALL be 200

#### Scenario: List player-accessible entries returns 200
- **WHEN** authenticated user calls GET /api/compendium/schemas/:id/entries
- **THEN** response status SHALL be 200

#### Scenario: Get single player-accessible entry returns 200
- **WHEN** authenticated user calls GET /api/compendium/schemas/:schemaId/entries/:entryId
- **THEN** response status SHALL be 200

#### Scenario: Player compendium search returns 200
- **WHEN** authenticated user calls GET /api/compendium/search?q=query
- **THEN** response status SHALL be 200
