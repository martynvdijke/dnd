# quick-edit-buttons Specification

## Purpose
+/- increment/decrement button pairs for quickly adjusting HP, coins, ability scores, XP, exhaustion, death saves, hit dice, inspiration, and proficiency bonus.

## Requirements

### Requirement: Ability scores have +/- stepper buttons
Each ability score (STR, DEX, CON, INT, WIS, CHA) SHALL display with a decrement (-) button, the current numeric value (tappable to type), and an increment (+) button arranged horizontally.

#### Scenario: Increment ability score
- **WHEN** the user taps the "+" button next to an ability score
- **THEN** the value SHALL increase by 1 and auto-save via the existing `autoSaveField` mechanism

#### Scenario: Decrement ability score
- **WHEN** the user taps the "-" button next to an ability score
- **THEN** the value SHALL decrease by 1 (minimum 1) and auto-save

#### Scenario: Tap number to type precise value
- **WHEN** the user taps the displayed ability score number
- **THEN** it SHALL transform into a text input for precise editing (existing behavior)

### Requirement: HP has large stepper controls
Hit Points current, max, and temp HP SHALL display with +/- buttons that adjust by 1 per tap, with larger styling than other steppers.

#### Scenario: Increment HP current
- **WHEN** the user taps "+" on current HP
- **THEN** HP current SHALL increase by 1 (not exceeding HP max) and auto-save

#### Scenario: Decrement HP current
- **WHEN** the user taps "-" on current HP
- **THEN** HP current SHALL decrease by 1 (minimum 0) and auto-save

### Requirement: Currency has inline stepper buttons
Each coin type (CP, SP, EP, GP, PP) in the details section SHALL display with +/- buttons next to the value for quick adjustment during loot distribution.

#### Scenario: Add coins
- **WHEN** the user taps "+" on a coin type
- **THEN** that coin count SHALL increase by 1 and auto-save

#### Scenario: Remove coins
- **WHEN** the user taps "-" on a coin type
- **THEN** that coin count SHALL decrease by 1 (minimum 0) and auto-save

### Requirement: Exhaustion has +/- buttons
The exhaustion level display SHALL keep its existing +/- buttons (already implemented) but SHALL render them at 44x44px minimum.

#### Scenario: Exhaustion buttons meet touch targets
- **WHEN** exhaustion buttons are displayed on mobile
- **THEN** each SHALL be at least 44x44px

### Requirement: Death saves have +/- buttons
Death save successes and failures SHALL display with +/- buttons for quick adjustment.

#### Scenario: Increment death save success
- **WHEN** the user taps "+" on death save successes
- **THEN** successes SHALL increase by 1 (max 3) and auto-save

#### Scenario: Increment death save failure
- **WHEN** the user taps "+" on death save failures
- **THEN** failures SHALL increase by 1 (max 3) and auto-save

### Requirement: Hit dice and XP have stepper controls
Hit dice total/used, XP, inspiration, and proficiency bonus SHALL render with +/- buttons.

#### Scenario: Increment XP
- **WHEN** the user taps "+" on XP
- **THEN** XP SHALL increase by 1 and auto-save

#### Scenario: Decrement XP
- **WHEN** the user taps "-" on XP (min 0)
- **THEN** XP SHALL decrease by 1 and auto-save

### Requirement: Stepper buttons show visual feedback on tap
When a stepper button is tapped, it SHALL briefly animate (scale bounce or background flash) to confirm the action.

#### Scenario: Visual feedback on increment
- **WHEN** the user taps any "+" stepper button
- **THEN** the button SHALL briefly flash or scale (CSS animation, ~150ms) to acknowledge the tap
