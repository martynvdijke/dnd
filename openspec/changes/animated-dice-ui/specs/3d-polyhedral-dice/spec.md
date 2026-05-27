## ADDED Requirements

### Requirement: 3D CSS rendering for all standard polyhedral dice
The system SHALL render each standard polyhedral die type (d4, d6, d8, d10, d12, d20, d100) as a distinct 3D CSS shape with proper face-up orientation on its result value.

| Die | 3D Shape | Face Count | Description |
|-----|----------|-----------|-------------|
| d4  | Tetrahedron (triangular pyramid) | 4 | 4 equilateral triangle faces |
| d6  | Cube (existing) | 6 | 6 square faces with pips |
| d8  | Octahedron | 8 | 8 equilateral triangle faces |
| d10 | Pentagonal trapezohedron | 10 | 10 kite-shaped faces |
| d12 | Dodecahedron | 12 | 12 regular pentagon faces |
| d20 | Icosahedron | 20 | 20 equilateral triangle faces |
| d100| Pentagonal trapezohedron (like d10) | 10 | Same shape as d10, distinct color |

#### Scenario: d6 3D cube rendering
- **WHEN** a d6 roll result is displayed
- **THEN** it renders as a rotating 3D cube with pip marks (1-6) on each face
- **THEN** the cube rotates so the result value's face is oriented toward the viewer

#### Scenario: d20 3D icosahedron rendering
- **WHEN** a d20 roll result is displayed
- **THEN** it renders as a 3D icosahedron (20 triangular faces) with the result number displayed on the face oriented toward the viewer
- **THEN** natural 20 shows a gold glow (crit success effect)
- **THEN** natural 1 shows a red glow (crit fail effect)

#### Scenario: d8 3D octahedron rendering
- **WHEN** a d8 roll result is displayed
- **THEN** it renders as a 3D octahedron (8 triangular faces, like two square pyramids base-to-base) with the result number on the visible face

#### Scenario: d4 3D tetrahedron rendering
- **WHEN** a d4 roll result is displayed
- **THEN** it renders as a 3D tetrahedron (4 triangular faces) with the result displayed on or near the visible face

#### Scenario: d10 and d100 3D rendering
- **WHEN** a d10 or d100 roll result is displayed
- **THEN** it renders as a 3D pentagonal trapezohedron (10 kite-shaped faces) with the result number on the visible face
- **THEN** d100 uses a distinct color scheme from d10

#### Scenario: d12 3D dodecahedron rendering
- **WHEN** a d12 roll result is displayed
- **THEN** it renders as a 3D dodecahedron (12 pentagonal faces) with the result number on the visible face

### Requirement: Dice color themes per type
Each die type SHALL use a distinct color theme matching the existing pattern.

#### Scenario: Color themes applied
- **WHEN** dice are rendered
- **THEN** each die type uses its designated color scheme as defined in the CSS classes `.d4` through `.d100`

### Requirement: Rolling animation for all dice types
All dice SHALL show a rolling/tumbling animation while the API request is in flight, then settle to display the result value.

#### Scenario: Rolling animation plays
- **WHEN** user clicks "Roll the Bones"
- **THEN** dice appear with a CSS rolling animation in the 3D container
- **WHEN** the API response arrives (after ~500-900ms)
- **THEN** the dice settle to their final orientation showing the result values
- **THEN** the total text appears with a pop-in animation

#### Scenario: Reduced motion respected
- **WHEN** user has `prefers-reduced-motion: reduce` set
- **THEN** dice appear in their final state immediately with no rolling animation

### Requirement: Die face markers
Each die face SHALL show an appropriate marker indicating its value:

| Die | Marker Type |
|-----|------------|
| d4  | Numeral on/near each vertex |
| d6  | Pips (dots) arranged in standard pattern (existing) |
| d8  | Numeral on each face |
| d10 | Numeral (0-9 or 1-10) on each face |
| d12 | Numeral on each face |
| d20 | Numeral on each face |
| d100| Numeral (00, 10, 20…90) on each face |

#### Scenario: d6 pip arrangement
- **WHEN** a d6 face shows value 4
- **THEN** it displays 4 pips arranged in the standard 4 corners pattern (as defined in `D6_PIPS`)

#### Scenario: Numeric face markers
- **WHEN** a d20 face shows value 20
- **THEN** it displays the numeral "20" centered on the face

### Requirement: Multiple dice of the same type render side by side
When an expression produces multiple dice of the same type, the system SHALL render each die individually, arranged in a flex row.

#### Scenario: 3d8 rendered
- **WHEN** user rolls `3d8`
- **THEN** three separate d8 3D shapes are displayed side by side in the dice container
- **THEN** each shows its individual result value

### Requirement: Crit effects on d20
Natural 20 and natural 1 on a d20 SHALL trigger visual effects.

#### Scenario: Critical hit glow
- **WHEN** a d20 shows a natural 20
- **THEN** the die wrapper gets the `dice-crit-success` class with gold glow box-shadow effect

#### Scenario: Critical fail glow
- **WHEN** a d20 shows a natural 1
- **THEN** the die wrapper gets the `dice-crit-fail` class with red glow box-shadow effect
