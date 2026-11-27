# rule-effects Specification

## Purpose
TBD - created by archiving change add-character-rule-effects. Update Purpose after archive.
## Requirements
### Requirement: Damage defenses
When a character takes typed damage, the system SHALL apply the character's damage immunities, resistances, and vulnerabilities in that priority order: immunity reduces the damage to zero, resistance halves it (rounded down), and vulnerability doubles it. A character with no matching defense SHALL take the unmodified damage.

#### Scenario: Resistance halves damage
- **WHEN** a character resistant to fire takes 9 fire damage
- **THEN** the applied damage is 4

#### Scenario: Vulnerability doubles damage
- **WHEN** a character vulnerable to cold takes 7 cold damage
- **THEN** the applied damage is 14

#### Scenario: Immunity negates damage
- **WHEN** a character immune to poison takes 12 poison damage
- **THEN** the applied damage is 0

#### Scenario: Untyped or undefended damage is unchanged
- **WHEN** a character with no relevant defense takes 10 slashing damage
- **THEN** the applied damage is 10

### Requirement: Condition mechanical effects
The system SHALL derive mechanical effects from a character's active conditions and apply them to rolls: disadvantage or advantage on attack rolls, ability checks, and saving throws; automatic failure of Strength and Dexterity saving throws; speed reduced to zero; and the incapacitated state. The system SHALL also flag when attacks made against the character have advantage.

#### Scenario: Blinded attacks at disadvantage
- **WHEN** a blinded character makes an attack roll
- **THEN** the attack is rolled at disadvantage

#### Scenario: Paralyzed auto-fails Strength and Dexterity saves
- **WHEN** a paralyzed character makes a Strength or Dexterity saving throw
- **THEN** the save automatically fails without rolling

#### Scenario: Restrained has zero speed
- **WHEN** the active effects for a restrained character are computed
- **THEN** speed is reported as zero

#### Scenario: Prone grants attackers advantage
- **WHEN** a prone character is targeted by an attack
- **THEN** the effects report that attacks against the character have advantage

### Requirement: Exhaustion effects
The system SHALL apply exhaustion level effects: level 1 imposes disadvantage on ability checks, level 2 halves speed, level 3 imposes disadvantage on attack rolls and saving throws, level 4 halves the effective hit point maximum, level 5 reduces speed to zero, and level 6 causes death.

#### Scenario: Exhaustion 3 penalizes attacks and saves
- **WHEN** a character at exhaustion level 3 makes an attack roll or a saving throw
- **THEN** the roll is at disadvantage

#### Scenario: Exhaustion 4 halves hit point maximum
- **WHEN** a character at exhaustion level 4 has an effective hit point maximum computed
- **THEN** the value is half the stored maximum

### Requirement: Death saving throws
The system SHALL resolve a death saving throw for a character at 0 hit points by rolling a d20: a natural 20 restores 1 hit point, a natural 1 adds two failures, a result of 10 or higher adds one success, and a result below 10 adds one failure. Three successes SHALL mark the character stable; three failures SHALL mark the character dead. Damage taken while at 0 hit points that is greater than or equal to the effective hit point maximum SHALL cause immediate death.

#### Scenario: Natural 20 revives
- **WHEN** a character at 0 hit points rolls a natural 20 on a death saving throw
- **THEN** the character regains 1 hit point and the death save counters reset

#### Scenario: Natural 1 adds two failures
- **WHEN** a character at 0 hit points rolls a natural 1 on a death saving throw
- **THEN** two failures are added

#### Scenario: Three successes stabilize
- **WHEN** a character accumulates three death save successes
- **THEN** the character is reported as stable

#### Scenario: Three failures kill
- **WHEN** a character accumulates three death save failures
- **THEN** the character is reported as dead

#### Scenario: Massive damage kills instantly
- **WHEN** a character at 0 hit points takes damage greater than or equal to the effective hit point maximum
- **THEN** the character is marked dead with three failures

### Requirement: Active effects reporting
The system SHALL expose a character's computed active effects, including condition effects, exhaustion effects, damage defenses, effective hit point maximum, and speed, so that clients can display what modifies a roll.

#### Scenario: Effects endpoint reports modifiers
- **WHEN** a client requests the active effects for a character with conditions and exhaustion
- **THEN** the response includes the derived condition and exhaustion modifiers, defenses, effective hit point maximum, and speed
