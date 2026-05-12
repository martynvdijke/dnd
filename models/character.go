package models

type Character struct {
	ID                int64  `json:"id"`
	UserID            int64  `json:"user_id"`
	CampaignID        *int64 `json:"campaign_id,omitempty"`
	Name              string `json:"name"`
	Race              string `json:"race"`
	Class             string `json:"class"`
	Subclass          string `json:"subclass"`
	Level             int    `json:"level"`
	XP                int    `json:"xp"`
	Background        string `json:"background"`
	Alignment         string `json:"alignment"`
	Str               int    `json:"str"`
	Dex               int    `json:"dex"`
	Con               int    `json:"con"`
	Int               int    `json:"int"`
	Wis               int    `json:"wis"`
	Cha               int    `json:"cha"`
	AC                int    `json:"ac"`
	Initiative        int    `json:"initiative"`
	Speed             int    `json:"speed"`
	HPMax             int    `json:"hp_max"`
	HPCurrent         int    `json:"hp_current"`
	TempHP            int    `json:"temp_hp"`
	HitDice           string `json:"hit_dice"`
	HitDiceCurrent    int    `json:"hit_dice_current"`
	ProficiencyBonus  int    `json:"proficiency_bonus"`
	Inspiration       int    `json:"inspiration"`
	PassivePerception int    `json:"passive_perception"`
	PersonalityTraits string `json:"personality_traits"`
	Ideals            string `json:"ideals"`
	Bonds             string `json:"bonds"`
	Flaws             string `json:"flaws"`
	Appearance        string `json:"appearance"`
	Backstory         string `json:"backstory"`
	PortraitURL       string `json:"portrait_url"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`

	// Computed
	StrMod              int    `json:"str_mod"`
	DexMod              int    `json:"dex_mod"`
	ConMod              int    `json:"con_mod"`
	IntMod              int    `json:"int_mod"`
	WisMod              int    `json:"wis_mod"`
	ChaMod              int    `json:"cha_mod"`
	SpellSaveDC         int    `json:"spell_save_dc"`
	SpellAttackBonus    int    `json:"spell_attack_bonus"`
	DeathSavesSuccesses int    `json:"death_saves_successes"`
	DeathSavesFailures  int    `json:"death_saves_failures"`
	ConcentratingOn     string `json:"concentrating_on"`

	Proficiencies []Proficiency   `json:"proficiencies,omitempty"`
	Features      []Feature       `json:"features,omitempty"`
	Spellcasting  *Spellcasting   `json:"spellcasting,omitempty"`
	Spells        []Spell         `json:"spells,omitempty"`
	Inventory     []InventoryItem `json:"inventory,omitempty"`
	Currency      *Currency       `json:"currency,omitempty"`
	Classes       []CharClass     `json:"classes,omitempty"`
}

type Currency struct {
	CharacterID int64 `json:"character_id"`
	CP          int   `json:"cp"`
	SP          int   `json:"sp"`
	EP          int   `json:"ep"`
	GP          int   `json:"gp"`
	PP          int   `json:"pp"`
}

type Proficiency struct {
	ID          int64  `json:"id"`
	CharacterID int64  `json:"character_id"`
	Type        string `json:"type"`
	Name        string `json:"name"`
}

type Feature struct {
	ID          int64  `json:"id"`
	CharacterID int64  `json:"character_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Source      string `json:"source"`
	LevelGained int    `json:"level_gained"`
}

type Spellcasting struct {
	CharacterID int64  `json:"character_id"`
	Ability     string `json:"ability"`
	SaveDC      int    `json:"save_dc"`
	AttackBonus int    `json:"attack_bonus"`
	Slots1Max   int    `json:"slots_1_max"`
	Slots1Used  int    `json:"slots_1_used"`
	Slots2Max   int    `json:"slots_2_max"`
	Slots2Used  int    `json:"slots_2_used"`
	Slots3Max   int    `json:"slots_3_max"`
	Slots3Used  int    `json:"slots_3_used"`
	Slots4Max   int    `json:"slots_4_max"`
	Slots4Used  int    `json:"slots_4_used"`
	Slots5Max   int    `json:"slots_5_max"`
	Slots5Used  int    `json:"slots_5_used"`
	Slots6Max   int    `json:"slots_6_max"`
	Slots6Used  int    `json:"slots_6_used"`
	Slots7Max   int    `json:"slots_7_max"`
	Slots7Used  int    `json:"slots_7_used"`
	Slots8Max   int    `json:"slots_8_max"`
	Slots8Used  int    `json:"slots_8_used"`
	Slots9Max   int    `json:"slots_9_max"`
	Slots9Used  int    `json:"slots_9_used"`
}

type Spell struct {
	ID             int64  `json:"id"`
	CharacterID    int64  `json:"character_id"`
	Name           string `json:"name"`
	Level          int    `json:"level"`
	School         string `json:"school"`
	CastingTime    string `json:"casting_time"`
	Range          string `json:"range"`
	Components     string `json:"components"`
	Duration       string `json:"duration"`
	Description    string `json:"description"`
	Prepared       bool   `json:"prepared"`
	AlwaysPrepared bool   `json:"always_prepared"`
	Source         string `json:"source"`
	Notes          string `json:"notes"`
}

type InventoryItem struct {
	ID               int64   `json:"id"`
	CharacterID      int64   `json:"character_id"`
	Name             string  `json:"name"`
	Quantity         int     `json:"quantity"`
	Weight           float64 `json:"weight"`
	Category         string  `json:"category"`
	DamageDice       string  `json:"damage_dice"`
	DamageType       string  `json:"damage_type"`
	WeaponProperties string  `json:"weapon_properties"`
	ACBonus          int     `json:"ac_bonus"`
	ArmorType        string  `json:"armor_type"`
	Description      string  `json:"description"`
	IsEquipped       bool    `json:"is_equipped"`
	IsMagical        bool    `json:"is_magical"`
	Attunement       bool    `json:"attunement"`
	Notes            string  `json:"notes"`
}

type DiceRoll struct {
	ID          int64  `json:"id"`
	UserID      int64  `json:"user_id"`
	CharacterID *int64 `json:"character_id,omitempty"`
	Expression  string `json:"expression"`
	Result      string `json:"result"`
	Total       int    `json:"total"`
	Timestamp   string `json:"timestamp"`
}

type Location struct {
	ID          int64    `json:"id"`
	UserID      int64    `json:"user_id"`
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Description string   `json:"description"`
	ParentID    *int64   `json:"parent_id,omitempty"`
	Latitude    *float64 `json:"latitude,omitempty"`
	Longitude   *float64 `json:"longitude,omitempty"`
	CreatedAt   string   `json:"created_at"`
}

type CharacterLocation struct {
	ID           int64  `json:"id"`
	CharacterID  int64  `json:"character_id"`
	LocationID   int64  `json:"location_id"`
	Relationship string `json:"relationship"`
	Notes        string `json:"notes"`
}

type NPC struct {
	ID          int64  `json:"id"`
	UserID      int64  `json:"user_id"`
	Name        string `json:"name"`
	Race        string `json:"race"`
	Class       string `json:"class"`
	Description string `json:"description"`
	Notes       string `json:"notes"`
	Str         int    `json:"str"`
	Dex         int    `json:"dex"`
	Con         int    `json:"con"`
	Int         int    `json:"int"`
	Wis         int    `json:"wis"`
	Cha         int    `json:"cha"`
	HPMax       int    `json:"hp_max"`
	HPCurrent   int    `json:"hp_current"`
	IsAlive     bool   `json:"is_alive"`
	CreatedAt   string `json:"created_at"`
}

type CharacterNPC struct {
	ID               int64  `json:"id"`
	CharacterID      int64  `json:"character_id"`
	NPCID            int64  `json:"npc_id"`
	Relationship     string `json:"relationship"`
	Notes            string `json:"notes"`
	InteractionCount int    `json:"interaction_count"`
	LastInteracted   string `json:"last_interacted"`
}

type Session struct {
	ID              int64  `json:"id"`
	CharacterID     int64  `json:"character_id"`
	SessionDate     string `json:"session_date"`
	Title           string `json:"title"`
	Notes           string `json:"notes"`
	XPEarned        int    `json:"xp_earned"`
	GoldEarned      int    `json:"gold_earned"`
	ImportantEvents string `json:"important_events"`
	CreatedAt       string `json:"created_at"`
}

type Quest struct {
	ID          int64  `json:"id"`
	CharacterID int64  `json:"character_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Objectives  string `json:"objectives"`
	Rewards     string `json:"rewards"`
	Notes       string `json:"notes"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type JournalEntry struct {
	ID          int64  `json:"id"`
	CharacterID int64  `json:"character_id"`
	Title       string `json:"title"`
	Entry       string `json:"entry"`
	EntryDate   string `json:"entry_date"`
	CreatedAt   string `json:"created_at"`
}

type GraphNode struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	Group  string `json:"group"`
	Color  string `json:"color"`
	Size   int    `json:"size"`
	CharID int64  `json:"char_id,omitempty"`
}

type GraphEdge struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Label  string `json:"label,omitempty"`
	Dashes bool   `json:"dashes"`
	Width  int    `json:"width"`
}

type GraphData struct {
	Nodes []GraphNode `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
}

type Campaign struct {
	ID          int64  `json:"id"`
	UserID      int64  `json:"user_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	DMNotes     string `json:"dm_notes"`
	CreatedAt   string `json:"created_at"`
}

type RestLog struct {
	ID             int64  `json:"id"`
	CharacterID    int64  `json:"character_id"`
	RestType       string `json:"rest_type"`
	HPHealed       int    `json:"hp_healed"`
	SlotsRecovered string `json:"slots_recovered"`
	HitDiceSpent   int    `json:"hit_dice_spent"`
	Notes          string `json:"notes"`
	Timestamp      string `json:"timestamp"`
}

type CharClass struct {
	ID          int64  `json:"id"`
	CharacterID int64  `json:"character_id"`
	Class       string `json:"class"`
	Subclass    string `json:"subclass"`
	Level       int    `json:"level"`
	HitDice     string `json:"hit_dice"`
}

type EncounterTemplate struct {
	ID          int64               `json:"id"`
	CampaignID  *int64              `json:"campaign_id,omitempty"`
	UserID      int64               `json:"user_id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Environment string              `json:"environment"`
	Difficulty  string              `json:"difficulty"`
	XPBudget    int                 `json:"xp_budget"`
	TotalXP     int                 `json:"total_xp"`
	Notes       string              `json:"notes"`
	Monsters    []EncounterMonster  `json:"monsters,omitempty"`
	CreatedAt   string              `json:"created_at"`
}

type EncounterMonster struct {
	ID           int64  `json:"id"`
	EncounterID  int64  `json:"encounter_id"`
	Name         string `json:"name"`
	Count        int    `json:"count"`
	CR           string `json:"cr"`
	XP           int    `json:"xp"`
	AC           int    `json:"ac"`
	HP           int    `json:"hp"`
	InitiativeMod int   `json:"initiative_mod"`
	Source       string `json:"source"`
	Notes        string `json:"notes"`
}

type CalendarEvent struct {
	ID          int64  `json:"id"`
	CampaignID  int64  `json:"campaign_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	EventDate   string `json:"event_date"`
	EventType   string `json:"event_type"`
	Color       string `json:"color"`
	CreatedAt   string `json:"created_at"`
}

type Condition struct {
	ID            int64  `json:"id"`
	CharacterID   int64  `json:"character_id"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	Source        string `json:"source"`
	Duration      int    `json:"duration"`
	DurationType  string `json:"duration_type"`
	SavingThrow   string `json:"saving_throw"`
	SaveDC        int    `json:"save_dc"`
	Description   string `json:"description"`
	StartedAt     string `json:"started_at"`
}

type CharacterFeat struct {
	ID           int64  `json:"id"`
	CharacterID  int64  `json:"character_id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Prerequisites string `json:"prerequisites"`
	Source       string `json:"source"`
	LevelGained  int    `json:"level_gained"`
}

type Companion struct {
	ID          int64  `json:"id"`
	CharacterID int64  `json:"character_id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Race        string `json:"race"`
	HPMax       int    `json:"hp_max"`
	HPCurrent   int    `json:"hp_current"`
	AC          int    `json:"ac"`
	Str         int    `json:"str"`
	Dex         int    `json:"dex"`
	Con         int    `json:"con"`
	Int         int    `json:"int"`
	Wis         int    `json:"wis"`
	Cha         int    `json:"cha"`
	Speed       int    `json:"speed"`
	Abilities   string `json:"abilities"`
	Notes       string `json:"notes"`
	PortraitURL string `json:"portrait_url"`
	IsAlive     bool   `json:"is_alive"`
}

type Faction struct {
	ID           int64  `json:"id"`
	CampaignID   *int64 `json:"campaign_id,omitempty"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Type         string `json:"type"`
	Headquarters string `json:"headquarters"`
}

type FactionReputation struct {
	ID          int64  `json:"id"`
	CharacterID int64  `json:"character_id"`
	FactionID   int64  `json:"faction_id"`
	Standing    int    `json:"standing"`
	Rank        string `json:"rank"`
	Notes       string `json:"notes"`
}

type CharacterNote struct {
	ID          int64  `json:"id"`
	CharacterID int64  `json:"character_id"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	Visibility  string `json:"visibility"`
	Category    string `json:"category"`
}

type TimelineEvent struct {
	ID                int64  `json:"id"`
	CampaignID        int64  `json:"campaign_id"`
	Title             string `json:"title"`
	Description       string `json:"description"`
	EventDate         string `json:"event_date"`
	EventType         string `json:"event_type"`
	Importance        int    `json:"importance"`
	Icon              string `json:"icon"`
	LinkedEntityType  string `json:"linked_entity_type"`
	LinkedEntityID    *int64 `json:"linked_entity_id,omitempty"`
	CreatedAt         string `json:"created_at"`
}
