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

// ─── Session Pacing ───

type SessionPacing struct {
	ID              int64  `json:"id"`
	AdventureID     int64  `json:"adventure_id"`
	CurrentActID    *int64 `json:"current_act_id,omitempty"`
	CurrentSceneID  *int64 `json:"current_scene_id,omitempty"`
	Status          string `json:"status"`
	ElapsedSeconds  int    `json:"elapsed_seconds"`
	StartedAt       string `json:"started_at"`
	CompletedAt     string `json:"completed_at,omitempty"`
	AdventureTitle  string `json:"adventure_title,omitempty"`
	ActTitle        string `json:"act_title,omitempty"`
	SceneTitle      string `json:"scene_title,omitempty"`
	SceneEstimated  int    `json:"scene_estimated_minutes,omitempty"`
	ActNumber       int    `json:"act_number,omitempty"`
	SceneNumber     int    `json:"scene_number,omitempty"`
	TotalActs       int    `json:"total_acts,omitempty"`
	TotalScenes     int    `json:"total_scenes,omitempty"`
	CompletedActs   int    `json:"completed_acts,omitempty"`
	CompletedScenes int    `json:"completed_scenes,omitempty"`
	SceneTimings    []SceneTiming `json:"scene_timings,omitempty"`
}

type SceneTiming struct {
	ID             int64  `json:"id"`
	SessionID      int64  `json:"session_id"`
	SceneID        int64  `json:"scene_id"`
	ElapsedSeconds int    `json:"elapsed_seconds"`
	Status         string `json:"status"`
	StartedAt      string `json:"started_at"`
	CompletedAt    string `json:"completed_at,omitempty"`
	SceneTitle     string `json:"scene_title,omitempty"`
	SceneType      string `json:"scene_type,omitempty"`
	EstimatedMin   int    `json:"estimated_minutes,omitempty"`
}
