## ADDED Requirements

### Requirement: Attack resolution
The system SHALL resolve an attack as a d20 roll plus an attack bonus against a target's AC, reporting hit, miss, or critical hit. A natural 20 SHALL be a critical hit; a natural 1 SHALL miss. On a hit, the system SHALL roll the weapon's damage dice and report the breakdown.

#### Scenario: Hit and damage
- **WHEN** an attacker resolves an attack against a target with AC 15 and the d20+bonus meets or exceeds 15
- **THEN** the result is a hit with a damage roll from the weapon's damage dice

#### Scenario: Critical hit
- **WHEN** the attack roll is a natural 20
- **THEN** the result is flagged critical and damage dice are doubled

#### Scenario: Natural 1 misses
- **WHEN** the attack roll is a natural 1
- **THEN** the result is a miss even if the bonus would meet the AC

### Requirement: Attack bonus derivation
The system SHALL derive a weapon's attack bonus from the character's relevant ability modifier, proficiency bonus, and magic bonus when no explicit bonus is stored, and SHALL use a stored explicit bonus when present.

#### Scenario: Finesse uses Dexterity
- **WHEN** a finesse weapon has no explicit attack bonus
- **THEN** the derived bonus uses the Dexterity modifier

#### Scenario: Explicit override wins
- **WHEN** a weapon stores an explicit attack bonus
- **THEN** that value is used instead of the derived one

### Requirement: Authoritative damage and healing
The system SHALL provide a single HP mutation endpoint that absorbs damage into temporary HP first, clamps current HP between 0 and maximum, increments death saves when HP reaches 0, resets them on healing, and returns the resulting HP state.

#### Scenario: Temporary HP absorbs damage
- **WHEN** a character with 5 temporary HP takes 8 damage
- **THEN** temporary HP drops to 0 and current HP drops by 3

#### Scenario: Reaching zero starts death saves
- **WHEN** damage reduces a character to 0 HP
- **THEN** death saves begin tracking and the 0 HP state is reported

### Requirement: Saving throws against a DC
The system SHALL roll a saving throw against a DC and report success or failure, supporting advantage/disadvantage and optional half-damage-on-success.

#### Scenario: Failed save
- **WHEN** a character rolls a save and the total is below the DC
- **THEN** the result reports failure

### Requirement: Condition application
On a failed save or a hit effect, the system SHALL apply the specified condition using the existing condition model, and SHALL NOT apply a condition the target is immune to.

#### Scenario: Immunity respected
- **WHEN** an effect tries to apply a condition the target is immune to
- **THEN** no condition is added

### Requirement: Concentration on damage
When a concentrating character takes damage, the system SHALL compute the concentration save DC and roll the saving throw automatically; on failure it SHALL drop concentration and remove the concentration condition.

#### Scenario: Concentration breaks
- **WHEN** a concentrating character takes damage and fails the concentration save
- **THEN** concentration is cleared and the concentration condition is removed

#### Scenario: Concentration holds
- **WHEN** a concentrating character takes damage and succeeds on the save
- **THEN** concentration is retained

### Requirement: Combat automation audit trail
Every automated attack, save, damage application, condition, and concentration resolution SHALL be recorded in the combat log with its roll breakdown.

#### Scenario: Resolution logged
- **WHEN** an attack is resolved with `apply` set
- **THEN** a combat-log entry records the roll, outcome, and damage applied

### Requirement: Round-based condition ticking
Advancing the turn SHALL tick round-based conditions.

#### Scenario: Condition expires on turn advance
- **WHEN** the turn advances and a round-based condition's duration reaches zero
- **THEN** the condition is removed
