package models

type CompendiumEntry struct {
	System string `json:"system"`
	Source string `json:"source"`
}

type CompendiumRace struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	Speed          int    `json:"speed"`
	Size           string `json:"size"`
	AbilityBonuses string `json:"ability_bonuses"`
	Traits         string `json:"traits"`
	Languages      string `json:"languages"`
	SourcePage     string `json:"source_page"`
	System         string `json:"system,omitempty"`
	Source         string `json:"source,omitempty"`
}

type CompendiumClass struct {
	ID                  int64  `json:"id"`
	Name                string `json:"name"`
	Description         string `json:"description"`
	HitDie              int    `json:"hit_die"`
	PrimaryAbility      string `json:"primary_ability"`
	SavingThrows        string `json:"saving_throws"`
	Proficiencies       string `json:"proficiencies"`
	SpellcastingAbility string `json:"spellcasting_ability"`
	SourcePage          string `json:"source_page"`
	System              string `json:"system,omitempty"`
	Source              string `json:"source,omitempty"`
}

type CompendiumSpell struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Level        int    `json:"level"`
	School       string `json:"school"`
	CastingTime  string `json:"casting_time"`
	Range        string `json:"range"`
	Components   string `json:"components"`
	Duration     string `json:"duration"`
	Description  string `json:"description"`
	HigherLevels string `json:"higher_levels"`
	Classes      string `json:"classes"`
	SourcePage   string `json:"source_page"`
	System       string `json:"system,omitempty"`
	Source       string `json:"source,omitempty"`
}

type CompendiumFeat struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Prerequisites string `json:"prerequisites"`
	SourcePage    string `json:"source_page"`
	System        string `json:"system,omitempty"`
	Source        string `json:"source,omitempty"`
}

type CompendiumBackground struct {
	ID                 int64  `json:"id"`
	Name               string `json:"name"`
	Description        string `json:"description"`
	FeatureName        string `json:"feature_name"`
	FeatureDescription string `json:"feature_description"`
	Proficiencies      string `json:"proficiencies"`
	SourcePage         string `json:"source_page"`
	System             string `json:"system,omitempty"`
	Source             string `json:"source,omitempty"`
}

type CompendiumEquipment struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Category    string  `json:"category"`
	Cost        string  `json:"cost"`
	Weight      float64 `json:"weight"`
	Description string  `json:"description"`
	SourcePage  string  `json:"source_page"`
	System      string  `json:"system,omitempty"`
	Source      string  `json:"source,omitempty"`
}

type ImportCharacter struct {
	Name       string `json:"name"`
	Race       string `json:"race"`
	Class      string `json:"class"`
	Subclass   string `json:"subclass"`
	Level      int    `json:"level"`
	XP         int    `json:"xp"`
	Background string `json:"background"`
	Alignment  string `json:"alignment"`
	Str        int    `json:"str"`
	Dex        int    `json:"dex"`
	Con        int    `json:"con"`
	Int        int    `json:"int"`
	Wis        int    `json:"wis"`
	Cha        int    `json:"cha"`
	AC         int    `json:"ac"`
	Initiative int    `json:"initiative"`
	Speed      int    `json:"speed"`
	HPMax      int    `json:"hp_max"`
	HPCurrent  int    `json:"hp_current"`
	TempHP     int    `json:"temp_hp"`
	HitDice    string `json:"hit_dice"`

	PersonalityTraits string `json:"personality_traits"`
	Ideals            string `json:"ideals"`
	Bonds             string `json:"bonds"`
	Flaws             string `json:"flaws"`
	Appearance        string `json:"appearance"`
	Backstory         string `json:"backstory"`

	Currency      Currency        `json:"currency"`
	Proficiencies []Proficiency   `json:"proficiencies"`
	Features      []Feature       `json:"features"`
	Spellcasting  *Spellcasting   `json:"spellcasting"`
	Spells        []Spell         `json:"spells"`
	Inventory     []InventoryItem `json:"inventory"`
}
