## ADDED Requirements

### Requirement: Stateful roll mode toggle for character checks and saves
The character sheet SHALL provide a roll mode toggle supporting `normal`, `advantage`, and `disadvantage`. Selecting `Advantage` or `Disadvantage` SHALL set the active roll mode for subsequent ability checks, skill checks, saving throws, initiative rolls, and weapon attack rolls instead of immediately rolling a hardcoded Strength check.

#### Scenario: Player rolls a skill check with Advantage
- **WHEN** a player activates the `Advantage` toggle on the character sheet and clicks the `Stealth` skill button
- **THEN** the system submits the Stealth skill roll with `mode: "advantage"`

#### Scenario: Toggling active roll mode off returns to normal
- **WHEN** `Advantage` is currently active and the player clicks `Advantage` again (or clicks `Normal`)
- **THEN** the active roll mode resets to `normal`

### Requirement: Accurate saving throw proficiency calculation
The Combat tab SHALL add `proficiency_bonus` only to saving throws in which the character is proficient (determined from the character's class saving throw proficiencies and any explicit saving throw proficiency overrides) and SHALL visually distinguish proficient saving throws from non-proficient ones.

#### Scenario: Fighter saving throws apply proficiency only to STR and CON
- **WHEN** a Level 1 Fighter with `proficiency_bonus: 2` and `10` (`+0`) in all six abilities views the Combat tab
- **THEN** `STR` and `CON` saving throws display `+2` (marked proficient) and `DEX`, `INT`, `WIS`, and `CHA` saving throws display `+0`

### Requirement: Equipped weapon quick attacks on the Combat tab
The Combat tab SHALL display the character's equipped weapons (items with `equipped: true` and weapon category or `damage_dice`) and provide 1-click buttons to roll the weapon's attack roll and damage roll.

#### Scenario: Player rolls equipped weapon attack from Combat tab
- **WHEN** a character has an equipped weapon with `damage_dice: "1d8"` and views the Combat tab
- **THEN** the weapon is listed in the Combat tab quick-attacks panel and clicking its Attack or Damage button rolls and displays the result

### Requirement: Configurable Hit Dice count on Short Rest
The Combat tab SHALL allow the player to specify how many available Hit Dice (`1` up to `hit_dice_remaining`) to spend when triggering a Short Rest.

#### Scenario: Spending multiple Hit Dice on a Short Rest
- **WHEN** a character with 3 remaining Hit Dice sets the Short Rest Hit Dice count to `2` and clicks `Short Rest`
- **THEN** the client requests `POST /api/characters/:id/rest` with `type: "short"` and `hit_dice_count: 2`

### Requirement: Character sheet tab category filtering
The character sheet SHALL provide category filter pills (`All`, `Play`, `Build`, `World`, `Insights`) above the sheet tab bar that filter visible tabs by gameplay context while keeping all tab controls accessible.

#### Scenario: Filtering sheet tabs to Play mode
- **WHEN** a player selects the `Play` tab filter pill on the character sheet
- **THEN** live-play tabs (`Stats`, `Combat`, `Spells`, `Inventory`, `Resources`, `Companions`, `Dice`) remain visible and non-play tabs are hidden until `All` or another category is selected
