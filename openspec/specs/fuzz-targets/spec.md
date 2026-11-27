# fuzz-targets Specification

## Purpose
Go fuzz testing for input parsing endpoints (dice, import, search, CR values).

## Requirements

### Requirement: Dice expression parser fuzz target

The dice engine SHALL have a Go fuzz target that tests the expression parser with random inputs.

#### Scenario: Fuzz target handles any byte sequence without panic
- **WHEN** `f.Fuzz(func(t *testing.T, data string)` is called with random byte sequences
- **THEN** the dice expression parser SHALL NOT panic, hang, or crash on any input
- **THEN** the parser SHALL either return a valid result or a non-nil error

### Requirement: Character JSON import fuzz target

The character import endpoint SHALL have a Go fuzz target that tests JSON parsing with random inputs.

#### Scenario: Fuzz target handles malformed JSON without panic
- **WHEN** `f.Fuzz(func(t *testing.T, data []byte)` feeds random bytes to the import handler
- **THEN** the handler SHALL NOT panic, hang, or crash
- **THEN** the handler SHALL either return a valid response or an error response (4xx/5xx)

### Requirement: Compendium search fuzz target

The search endpoint SHALL have a Go fuzz target that tests query string handling.

#### Scenario: Fuzz target handles arbitrary search queries without panic
- **WHEN** `f.Fuzz(func(t *testing.T, query string)` feeds random strings to the search handler
- **THEN** the handler SHALL NOT panic, hang, or crash
- **THEN** the handler SHALL return either 200 or 400

### Requirement: Encounter CR value parser fuzz target

The encounter system SHALL have a Go fuzz target that tests CR string parsing.

#### Scenario: Fuzz target handles random CR values without panic
- **WHEN** `f.Fuzz(func(t *testing.T, cr string)` feeds random strings to the CR parser
- **THEN** the parser SHALL NOT panic or hang
- **THEN** the parser SHALL return either a valid XP value or an error
