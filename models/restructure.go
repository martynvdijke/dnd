package models

type Party struct {
	ID          int64  `json:"id"`
	UserID      int64  `json:"user_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
}

type OneShotItem struct {
	ID          int64   `json:"id"`
	AdventureID int64   `json:"adventure_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Quantity    int     `json:"quantity"`
	Weight      float64 `json:"weight"`
	PriceGP     float64 `json:"price_gp"`
	IsMagical   bool    `json:"is_magical"`
	Attunement  bool    `json:"attunement"`
	Notes       string  `json:"notes"`
	CreatedAt   string  `json:"created_at"`
}

type MonsterLibraryEntry struct {
	ID                  int64  `json:"id"`
	UserID              int64  `json:"user_id"`
	Name                string `json:"name"`
	AC                  int    `json:"ac"`
	HP                  int    `json:"hp"`
	Str                 int    `json:"str"`
	Dex                 int    `json:"dex"`
	Con                 int    `json:"con"`
	Int                 int    `json:"int"`
	Wis                 int    `json:"wis"`
	Cha                 int    `json:"cha"`
	CR                  string `json:"cr"`
	Source              string `json:"source"`
	IsFull              bool   `json:"is_full"`
	Saves               string `json:"saves"`
	Skills              string `json:"skills"`
	DamageVulnerabilities string `json:"damage_vulnerabilities"`
	DamageResistances   string `json:"damage_resistances"`
	DamageImmunities    string `json:"damage_immunities"`
	ConditionImmunities string `json:"condition_immunities"`
	Senses              string `json:"senses"`
	Languages           string `json:"languages"`
	SpecialAbilities    string `json:"special_abilities"`
	Actions             string `json:"actions"`
	LegendaryActions    string `json:"legendary_actions"`
	Description         string `json:"description"`
	CreatedAt           string `json:"created_at"`
}

type OneShotMonster struct {
	ID                  int64  `json:"id"`
	AdventureID         int64  `json:"adventure_id"`
	ActID               *int64 `json:"act_id,omitempty"`
	SceneID             *int64 `json:"scene_id,omitempty"`
	Name                string `json:"name"`
	AC                  int    `json:"ac"`
	HP                  int    `json:"hp"`
	Str                 int    `json:"str"`
	Dex                 int    `json:"dex"`
	Con                 int    `json:"con"`
	Int                 int    `json:"int"`
	Wis                 int    `json:"wis"`
	Cha                 int    `json:"cha"`
	CR                  string `json:"cr"`
	Source              string `json:"source"`
	IsFull              bool   `json:"is_full"`
	Saves               string `json:"saves"`
	Skills              string `json:"skills"`
	DamageVulnerabilities string `json:"damage_vulnerabilities"`
	DamageResistances   string `json:"damage_resistances"`
	DamageImmunities    string `json:"damage_immunities"`
	ConditionImmunities string `json:"condition_immunities"`
	Senses              string `json:"senses"`
	Languages           string `json:"languages"`
	SpecialAbilities    string `json:"special_abilities"`
	Actions             string `json:"actions"`
	LegendaryActions    string `json:"legendary_actions"`
	LibraryID           *int64 `json:"library_id,omitempty"`
	CreatedAt           string `json:"created_at"`
}

type NPCItemLink struct {
	ID               int64  `json:"id"`
	NPCID            int64  `json:"npc_id"`
	AdventureID      int64  `json:"adventure_id"`
	ItemID           int64  `json:"item_id"`
	RelationshipType string `json:"relationship_type"`
	Notes            string `json:"notes"`
	ItemName         string `json:"item_name,omitempty"`
	NPCName          string `json:"npc_name,omitempty"`
}

type OneShotPlayerCharacter struct {
	ID          int64  `json:"id"`
	AdventureID int64  `json:"adventure_id"`
	CharacterID int64  `json:"character_id"`
	Role        string `json:"role"`
	Notes       string `json:"notes"`
	CharName    string `json:"char_name,omitempty"`
	Username    string `json:"username,omitempty"`
}
