## MODIFIED Requirements

### Requirement: Attack resolution
The system SHALL resolve an attack as a d20 roll plus an attack bonus against a target's AC, reporting hit, miss, or critical hit. A natural 20 SHALL be a critical hit; a natural 1 SHALL miss. The attack roll SHALL incorporate advantage or disadvantage derived from the attacker's conditions and exhaustion, from conditions on the target that grant attackers advantage, and from the situation (flanking), combined with any explicitly requested advantage. The system SHALL apply cover and range modifiers: half cover adds 2 to the target's effective AC, three-quarters cover adds 5, total cover prevents the attack, a target beyond the weapon's normal range but within long range imposes disadvantage, and a target beyond long range cannot be attacked. The system SHALL report the cover bonus and the situational advantage so clients can explain the result. On a hit, the system SHALL roll the weapon's damage dice and report the breakdown. The system SHALL support attacks initiated either by a player character (`attacker_id`) or by a monster/NPC combat entry supplying an explicit `attacker_name`, `attack_bonus`, and `damage_dice`.

#### Scenario: Hit and damage
- **WHEN** an attacker resolves an attack against a target with AC 15 and the d20+bonus meets or exceeds 15
- **THEN** the result is a hit with a damage roll from the weapon's damage dice

#### Scenario: Monster combatant resolves an attack against a target
- **WHEN** a monster combatant resolves an attack with `attacker_name: "Goblin"`, `attack_bonus: 4`, and `damage_dice: "1d6+2"` against a target combat entry
- **THEN** the system resolves the d20+4 roll against the target's AC, rolls `1d6+2` on a hit, and logs the outcome

#### Scenario: Critical hit
- **WHEN** the attack roll is a natural 20
- **THEN** the result is flagged critical and damage dice are doubled

#### Scenario: Natural 1 misses
- **WHEN** the attack roll is a natural 1
- **THEN** the result is a miss even if the bonus would meet the AC

#### Scenario: Condition disadvantage applies
- **WHEN** a poisoned character makes an attack
- **THEN** the attack roll is resolved at disadvantage

#### Scenario: Half cover raises AC
- **WHEN** an attacker targets a creature behind half cover
- **THEN** the target's effective AC is increased by 2 for that attack

#### Scenario: Total cover blocks the attack
- **WHEN** an attacker targets a creature behind total cover
- **THEN** the attack cannot be made and no damage is dealt

#### Scenario: Flanking grants advantage
- **WHEN** an attacker flanks the target
- **THEN** the attack roll is resolved at advantage

#### Scenario: Long range imposes disadvantage
- **WHEN** the target is beyond the weapon's normal range but within its long range
- **THEN** the attack roll is resolved at disadvantage

#### Scenario: Beyond long range cannot be attacked
- **WHEN** the target is beyond the weapon's long range
- **THEN** the attack cannot be made and no damage is dealt
