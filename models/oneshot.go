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
	Notes            string `json:"notes"`
	// Loaded relations
	Scenes   []OneShotScene  `json:"scenes,omitempty"`
	ActNPCs  []OneShotActNPC `json:"act_npcs,omitempty"`
	ActNotes []DmNote        `json:"act_notes,omitempty"`
}

type OneShotActNPC struct {
	ID        int64  `json:"id"`
	ActID     int64  `json:"act_id"`
	NPCID     *int64 `json:"npc_id,omitempty"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	Notes     string `json:"notes"`
	IsInline  bool   `json:"is_inline"`
	CreatedAt string `json:"created_at"`
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

// ─── Clue/Mystery Tracker ───

type Clue struct {
	ID           int64  `json:"id"`
	AdventureID  int64  `json:"adventure_id"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	ClueType     string `json:"clue_type"`
	IsRedHerring bool   `json:"is_red_herring"`
	IsRevealed   bool   `json:"is_revealed"`
	SortOrder    int    `json:"sort_order"`
	Notes        string `json:"notes"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
	// Loaded relations
	Dependencies []ClueDependency `json:"dependencies"`
	DependedBy   []ClueDependency `json:"depended_by"`
	NPCs         []ClueNPC        `json:"npcs"`
	Locations    []ClueLocation   `json:"locations"`
}

type ClueDependency struct {
	ID            int64  `json:"id"`
	ClueID        int64  `json:"clue_id"`
	DependsOnID   int64  `json:"depends_on_id"`
	DependsOnTitle string `json:"depends_on_title,omitempty"`
}

type ClueNPC struct {
	ID      int64  `json:"id"`
	ClueID  int64  `json:"clue_id"`
	NPCID   int64  `json:"npc_id"`
	NPCName string `json:"npc_name,omitempty"`
}

type ClueLocation struct {
	ID           int64  `json:"id"`
	ClueID       int64  `json:"clue_id"`
	LocationID   int64  `json:"location_id"`
	LocationName string `json:"location_name,omitempty"`
}

// ─── Pregenerated Characters ───

type PregeneratedCharacter struct {
	ID          int64  `json:"id"`
	UserID      int64  `json:"user_id"`
	Name        string `json:"name"`
	Race        string `json:"race"`
	Class       string `json:"class"`
	Subclass    string `json:"subclass"`
	Level       int    `json:"level"`
	Background  string `json:"background"`
	Alignment   string `json:"alignment"`
	Str         int    `json:"str"`
	Dex         int    `json:"dex"`
	Con         int    `json:"con"`
	Int         int    `json:"int"`
	Wis         int    `json:"wis"`
	Cha         int    `json:"cha"`
	HP          int    `json:"hp"`
	AC          int    `json:"ac"`
	Speed       int    `json:"speed"`
	Skills      string `json:"skills"`
	Equipment   string `json:"equipment"`
	Spells      string `json:"spells"`
	Features    string `json:"features"`
	Personality string `json:"personality"`
	Backstory   string `json:"backstory"`
	PortraitURL string `json:"portrait_url"`
	Notes       string `json:"notes"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type PartyBalance struct {
	Characters []PregeneratedCharacter `json:"characters"`
	Roles      map[string]int          `json:"roles"`
	Score      int                     `json:"score"`
	Rating     string                  `json:"rating"`
	Missing    []string                `json:"missing_roles"`
	Suggestion string                  `json:"suggestion"`
}

// ─── Prep Dashboard ───

type PrepChecklistItem struct {
	ID          int64  `json:"id"`
	AdventureID int64  `json:"adventure_id"`
	Item        string `json:"item"`
	Category    string `json:"category"`
	IsChecked   bool   `json:"is_checked"`
	SortOrder   int    `json:"sort_order"`
}

type PrepDashboardData struct {
	Adventure   OneShotAdventure        `json:"adventure"`
	Acts        []OneShotAct            `json:"acts"`
	Clues       []Clue                  `json:"clues"`
	Pregens     []PregeneratedCharacter `json:"pregens"`
	Checklist   []PrepChecklistItem     `json:"checklist"`
	Pacing      *SessionPacing          `json:"pacing,omitempty"`
	SessionID   *int64                  `json:"session_id,omitempty"`
}

// ─── DM Screen ───

type DmNote struct {
	ID          int64  `json:"id"`
	AdventureID int64  `json:"adventure_id"`
	UserID      int64  `json:"user_id"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type DmQuickRefSection struct {
	Title    string           `json:"title"`
	Icon     string           `json:"icon"`
	Entries  []DmQuickRefEntry `json:"entries"`
}

type DmQuickRefEntry struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Reference   string `json:"reference,omitempty"`
}
