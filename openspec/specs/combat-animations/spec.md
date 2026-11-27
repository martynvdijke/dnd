# combat-animations Specification

## Purpose
GSAP-orchestrated animation sequences for the three highest-impact D&D emotional moments — crit celebrations, HP changes, and combat turn transitions.

## Requirements

### Requirement: Crit celebration animation
The system SHALL play an orchestrated animation sequence when a critical hit (max roll on any die d4+) or critical fail (nat 1 on any die d4+) is detected after dice settle. The sequence SHALL include a screen-edge color flash, a particle burst from the die center, and a text slam ("CRITICAL!" for success, "CRITICAL FAIL!" for fail), timed with the existing Web Audio crit chord.

#### Scenario: Nat 20 on d20 triggers crit celebration
- **WHEN** a d20 roll result of 20 is displayed after dice settle
- **THEN** the system plays a gold screen-edge flash, 12 gold particle sparks burst from the die center, "CRITICAL!" text slams in with back.ease, and the crit chord plays

#### Scenario: Nat 1 on d20 triggers crit fail animation
- **WHEN** a d20 roll result of 1 is displayed after dice settle
- **THEN** the system plays a red screen-edge flash, 12 red particle sparks burst from the die center, "CRITICAL FAIL!" text slams in, and the crit fail sound plays

#### Scenario: Max roll on non-d20 die triggers crit celebration
- **WHEN** a d6 roll result of 6, d8 result of 8, d10 result of 10, d12 result of 12, or d100 result of 100 is displayed after dice settle
- **THEN** the system plays a crit celebration sequence with a "Max on dX!" text label instead of "CRITICAL!"

#### Scenario: Non-crit roll does not trigger celebration
- **WHEN** a roll result is neither the max nor 1 for the die type
- **THEN** no crit celebration sequence plays

#### Scenario: Reduced motion skips crit animation
- **WHEN** the user's browser has `prefers-reduced-motion: reduce` set
- **THEN** no visual crit animation plays (no flash, no particles, no text slam), but the crit sound still plays if sound is enabled

### Requirement: HP damage and heal animation
The system SHALL animate HP changes on the character sheet with a floating damage/heal number, a color flash on the HP value, and an elastic-eased HP bar transition.

#### Scenario: Taking damage shows floating red number
- **WHEN** a character's HP decreases
- **THEN** a red "-N" number appears at the HP display, flies upward 40px, and fades out over 0.4s, while the HP number flashes red and the HP bar shrinks with elastic easing

#### Scenario: Healing shows floating green number
- **WHEN** a character's HP increases
- **THEN** a green "+N" number appears at the HP display, flies upward 40px, and fades out over 0.4s, while the HP number flashes green and the HP bar grows with elastic easing

#### Scenario: HP bar color shifts at low health
- **WHEN** a character's HP drops below 25% of max HP
- **THEN** the HP bar pulses red continuously via CSS animation (GSAP triggers the class, CSS owns the loop)

#### Scenario: Reduced motion skips HP animation
- **WHEN** the user's browser has `prefers-reduced-motion: reduce` set
- **THEN** HP updates instantly without floating numbers, color flashes, or bar animation, but the HP value and bar width still update to the correct values

### Requirement: Combat turn transition animation
The system SHALL animate combatant turn changes in the combat tracker with a slide-in and glow pulse on the newly active combatant's row.

#### Scenario: Next turn slides new combatant into focus
- **WHEN** `nextTurn()` advances to the next combatant in initiative order
- **THEN** the previous combatant's row dims to reduced opacity, the new active row slides in from 20px right with a gold left-border glow, and the combatant name pulses scale 1→1.05→1

#### Scenario: Monster turn adds red tint sweep
- **WHEN** the new active combatant is a monster (not a player character)
- **THEN** in addition to the standard turn transition, a subtle red tint sweeps across the active row

#### Scenario: Reduced motion skips turn animation
- **WHEN** the user's browser has `prefers-reduced-motion: reduce` set
- **THEN** turn changes update instantly without slide, glow, or pulse, but the active row is still visually marked (e.g., via CSS class for accessibility)

### Requirement: Animation cleanup
The system SHALL clean up all dynamically-created animation elements (particle sparks, floating numbers, crit overlay) after their animations complete, leaving no persistent DOM pollution.

#### Scenario: Particle sparks removed after animation
- **WHEN** a crit celebration particle burst animation completes
- **THEN** all 12 `.crit-spark` elements are removed from the DOM

#### Scenario: Floating damage number removed after animation
- **WHEN** an HP damage/heal floating number animation completes
- **THEN** the `.damage-float` element is removed from the DOM

#### Scenario: Crit overlay removed after animation
- **WHEN** a crit celebration sequence completes
- **THEN** the `.crit-overlay` and `.crit-text` elements are removed from the DOM
