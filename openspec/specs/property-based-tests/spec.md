# property-based-tests Specification

## Purpose
Generative (rapid/quickcheck) tests for core domain computations (stats, dice, XP, HP, slots).

## Requirements

### Requirement: Character ability score invariants

The system SHALL have property-based tests verifying character ability score computations.

#### Scenario: Ability modifier formula holds for all valid scores
- **WHEN** ability scores from 1 to 30 are generated via rapid
- **THEN** modifier SHALL equal floor((score-10)/2)

#### Scenario: Passive perception formula is correct
- **WHEN** wisdom modifier and proficiency bonus are generated
- **THEN** passive_perception SHALL equal 10 + wis_mod + (proficiency_bonus if proficient)

#### Scenario: HP auto-calc for single class matches expected distribution
- **WHEN** a Fighter (d10) at level L with CON modifier C is auto-calc'd
- **THEN** hp_max SHALL equal 10 + C + sum_{i=2..L} (dice_roll + C) where dice_roll is in [5..10]

### Requirement: Spell slot computation invariants

The system SHALL have property-based tests for spell slot recovery and usage.

#### Scenario: Long rest resets all used slots to zero
- **WHEN** a character with random spell slot usage is long rested
- **THEN** all slots_*_used SHALL be 0

#### Scenario: Total prepared spells never exceeds available slots
- **WHEN** a character has random level and spellcasting ability
- **THEN** count of prepared spells SHALL NOT exceed class_prep_limit + ability_mod

### Requirement: Dice expression invariants

The system SHALL have property-based tests for the dice engine.

#### Scenario: Dice roll result is within expected bounds
- **WHEN** NdX+K expression is generated (N in 1..10, X in {4,6,8,10,12,20,100}, K in -10..10)
- **THEN** total SHALL be between N*1+K and N*X+K

#### Scenario: Advantage always picks higher roll
- **WHEN** two d20 rolls are simulated with advantage
- **THEN** the chosen result SHALL be the maximum of the two

#### Scenario: Disadvantage always picks lower roll
- **WHEN** two d20 rolls are simulated with disadvantage
- **THEN** the chosen result SHALL be the minimum of the two

### Requirement: Encounter XP calculation invariants

The system SHALL have property-based tests for encounter XP math.

#### Scenario: Total XP equals sum of monster XP
- **WHEN** encounter has M monsters each with XP X_i
- **THEN** total_xp SHALL equal sum(X_i)

#### Scenario: Difficulty multiplier matches party size rules
- **WHEN** party has P members (P in 1..8) and monster count is M
- **THEN** the XP multiplier SHALL match the 5e rules for that party/monster ratio
