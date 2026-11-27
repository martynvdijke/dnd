# concurrent-safety-tests Specification

## Purpose
Parallel request tests for shared-state endpoints (combat, inventory, spell slots).

## Requirements

### Requirement: Combat tracker concurrent access

The combat tracker SHALL have tests verifying safe concurrent access.

#### Scenario: Multiple concurrent initiative rolls do not corrupt state
- **WHEN** 10 concurrent POST /api/combat/initiative requests are made for different characters
- **THEN** all requests SHALL return 200 and all created entries SHALL be retrievable

#### Scenario: Concurrent next-turn calls are serialized
- **WHEN** 5 concurrent POST /api/combat/next-turn calls are made
- **THEN** each call SHALL return 200 and the turn counter SHALL advance by exactly 5

#### Scenario: Concurrent HP updates on same entry do not produce negative HP
- **WHEN** 10 concurrent PUT /api/combat/:id requests decrement HP by varying amounts
- **THEN** hp_current SHALL NEVER be negative (bounded by 0)

### Requirement: Inventory concurrent access

The inventory system SHALL have tests verifying safe concurrent updates.

#### Scenario: Concurrent inventory quantity updates do not lose data
- **WHEN** 5 concurrent PUT /api/inventory/:id requests modify quantity by different amounts
- **THEN** final quantity SHALL equal initial + sum of all deltas

### Requirement: Spell slot concurrent access

The spellcasting system SHALL have tests verifying safe concurrent slot usage.

#### Scenario: Concurrent spell slot decrements stay within bounds
- **WHEN** 10 concurrent requests decrement spell slots
- **THEN** slots_used SHALL NEVER exceed slots_max for any level
