package models

type OneShotAdventure struct {
	ID              int64  `json:"id"`
	UserID          int64  `json:"user_id"`
	CampaignID      *int64 `json:"campaign_id,omitempty"`
	Title           string `json:"title"`
	Premise         string `json:"premise"`
	Hook            string `json:"hook"`
	Template        string `json:"template"`
	EstimatedMinutes int   `json:"estimated_minutes"`
	Difficulty      string `json:"difficulty"`
	Notes           string `json:"notes"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
	// Loaded relations
	Acts       []OneShotAct                `json:"acts,omitempty"`
	NPCs       []OneShotAdventureNPC       `json:"npcs,omitempty"`
	Locations  []OneShotAdventureLocation  `json:"locations,omitempty"`
	Encounters []OneShotAdventureEncounter `json:"encounters,omitempty"`
}

type OneShotAct struct {
	ID               int64  `json:"id"`
	AdventureID      int64  `json:"adventure_id"`
	Number           int    `json:"number"`
	Title            string `json:"title"`
	Description      string `json:"description"`
	EstimatedMinutes int    `json:"estimated_minutes"`
	// Loaded relations
	Scenes []OneShotScene `json:"scenes,omitempty"`
}

type OneShotScene struct {
	ID               int64  `json:"id"`
	ActID            int64  `json:"act_id"`
	Number           int    `json:"number"`
	Title            string `json:"title"`
	Description      string `json:"description"`
	SceneType        string `json:"scene_type"`
	LocationID       *int64 `json:"location_id,omitempty"`
	EncounterID      *int64 `json:"encounter_id,omitempty"`
	EstimatedMinutes int    `json:"estimated_minutes"`
	Notes            string `json:"notes"`
	// Loaded relations
	LocationName  string `json:"location_name,omitempty"`
	EncounterName string `json:"encounter_name,omitempty"`
}

type OneShotAdventureNPC struct {
	ID          int64  `json:"id"`
	AdventureID int64  `json:"adventure_id"`
	NPCID       int64  `json:"npc_id"`
	Role        string `json:"role"`
	NPCName     string `json:"npc_name,omitempty"`
}

type OneShotAdventureLocation struct {
	ID           int64  `json:"id"`
	AdventureID  int64  `json:"adventure_id"`
	LocationID   int64  `json:"location_id"`
	LocationName string `json:"location_name,omitempty"`
}

type OneShotAdventureEncounter struct {
	ID            int64  `json:"id"`
	AdventureID   int64  `json:"adventure_id"`
	EncounterID   int64  `json:"encounter_id"`
	EncounterName string `json:"encounter_name,omitempty"`
}
