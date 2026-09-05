package handlers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
	"villum/db"
	"villum/ent"
	"villum/ent/campaign"
	"villum/ent/campaignmember"
	"villum/ent/character"
	"villum/ent/characterproficiency"
	"villum/models"
)

func abilityMod(score int) int {
	return int(math.Floor(float64(score-10) / 2.0))
}

func computeMods(ch *models.Character) {
	ch.StrMod = abilityMod(ch.Str)
	ch.DexMod = abilityMod(ch.Dex)
	ch.ConMod = abilityMod(ch.Con)
	ch.IntMod = abilityMod(ch.Int)
	ch.WisMod = abilityMod(ch.Wis)
	ch.ChaMod = abilityMod(ch.Cha)
	if ch.Spellcasting != nil && ch.Spellcasting.Ability != "" {
		abilMod := 0
		switch ch.Spellcasting.Ability {
		case "str":
			abilMod = ch.StrMod
		case "dex":
			abilMod = ch.DexMod
		case "con":
			abilMod = ch.ConMod
		case "int":
			abilMod = ch.IntMod
		case "wis":
			abilMod = ch.WisMod
		case "cha":
			abilMod = ch.ChaMod
		}
		ch.SpellSaveDC = 8 + ch.ProficiencyBonus + abilMod
		ch.SpellAttackBonus = ch.ProficiencyBonus + abilMod
	}
}

func ListCharacters(c *gin.Context) {
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	type CharSummary struct {
		ID            int64  `json:"id"`
		UserID        int64  `json:"user_id"`
		Name          string `json:"name"`
		Race          string `json:"race"`
		Class         string `json:"class"`
		Level         int    `json:"level"`
		HPMax         int    `json:"hp_max"`
		HPCurrent     int    `json:"hp_current"`
		PortraitURL   string `json:"portrait_url,omitempty"`
		CampaignID    int64  `json:"campaign_id,omitempty"`
		RaceColor     string `json:"race_color,omitempty"`
		CharacterType string `json:"character_type"`
		CanEdit       bool   `json:"can_edit"`
	}
	chars := []CharSummary{}
	raceColors := GetRaceColorMap()

	if role == "admin" {
		query := c.DefaultQuery("q", "")
		if query != "" {
			rows, err := db.DB.Query(`
				SELECT c.id, c.user_id, c.name, c.race, c.class, c.level, c.hp_max, c.hp_current, COALESCE(c.portrait_url,''), COALESCE(c.character_type,'player')
				FROM characters c JOIN characters_fts fts ON c.id = fts.rowid
				WHERE characters_fts MATCH ? ORDER BY c.updated_at DESC`, query)
			if err != nil {
				WriteError(c, http.StatusInternalServerError, err)
				return
			}
			defer rows.Close()
			for rows.Next() {
				var ch CharSummary
				rows.Scan(&ch.ID, &ch.UserID, &ch.Name, &ch.Race, &ch.Class, &ch.Level, &ch.HPMax, &ch.HPCurrent, &ch.PortraitURL, &ch.CharacterType)
				ch.RaceColor = raceColors[ch.Race]
				ch.CanEdit = true
				chars = append(chars, ch)
			}
		} else {
			entChars, err := db.Client.Character.Query().Order(ent.Desc(character.FieldUpdatedAt)).All(c.Request.Context())
			if err != nil {
				WriteError(c, http.StatusInternalServerError, err)
				return
			}
			for _, e := range entChars {
				ch := CharSummary{ID: e.ID, UserID: e.UserID, Name: e.Name, Race: e.Race, Class: e.Class, Level: e.Level, HPMax: e.HpMax, HPCurrent: e.HpCurrent, PortraitURL: e.PortraitURL, CharacterType: e.CharacterType, CampaignID: e.CampaignID, CanEdit: canEditCharacter(c, e)}
				ch.RaceColor = raceColors[ch.Race]
				chars = append(chars, ch)
			}
		}
	} else {
		uid, _ := userID.(int64)
		entChars, err := db.Client.Character.Query().Where(character.UserID(uid)).Order(ent.Desc(character.FieldUpdatedAt)).All(c.Request.Context())
		if err != nil {
			WriteError(c, http.StatusInternalServerError, err)
			return
		}
		for _, e := range entChars {
			ch := CharSummary{ID: e.ID, UserID: e.UserID, Name: e.Name, Race: e.Race, Class: e.Class, Level: e.Level, HPMax: e.HpMax, HPCurrent: e.HpCurrent, PortraitURL: e.PortraitURL, CharacterType: e.CharacterType, CampaignID: e.CampaignID, CanEdit: canEditCharacter(c, e)}
			ch.RaceColor = raceColors[ch.Race]
			chars = append(chars, ch)
		}
	}
	c.JSON(http.StatusOK, chars)
}

// ListCampaignCharacters returns characters belonging to a campaign that the
// current user can access (campaign owner, DM, member, or admin), each
// annotated with `owned` so the UI can enforce read-only vs. full access.
func ListCampaignCharacters(c *gin.Context) {
	campaignID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		WriteError(c, http.StatusBadRequest, strErr("invalid campaign id"))
		return
	}
	ctx := c.Request.Context()

	// Access: admin, campaign owner, or campaign member may list.
	if _, err := db.Client.Campaign.Query().
		Where(campaign.ID(campaignID)).
		Only(ctx); ent.IsNotFound(err) {
		WriteNotFound(c, "campaign not found")
		return
	} else if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	if !isCampaignMember(c, campaignID) {
		WriteError(c, http.StatusForbidden, strErr("access denied"))
		return
	}

	type CampaignCharSummary struct {
		ID            int64  `json:"id"`
		UserID        int64  `json:"user_id"`
		Name          string `json:"name"`
		Race          string `json:"race"`
		Class         string `json:"class"`
		Level         int    `json:"level"`
		HPMax         int    `json:"hp_max"`
		HPCurrent     int    `json:"hp_current"`
		PortraitURL   string `json:"portrait_url,omitempty"`
		RaceColor     string `json:"race_color,omitempty"`
		CharacterType string `json:"character_type"`
		CampaignID    int64  `json:"campaign_id"`
		Owned         bool   `json:"owned"`
	}

	chars, err := db.Client.Character.Query().
		Where(character.CampaignID(campaignID)).
		Order(ent.Desc(character.FieldUpdatedAt)).
		All(ctx)
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}

	raceColors := GetRaceColorMap()
	out := make([]CampaignCharSummary, 0, len(chars))
	for _, e := range chars {
		out = append(out, CampaignCharSummary{
			ID:            e.ID,
			UserID:        e.UserID,
			Name:          e.Name,
			Race:          e.Race,
			Class:         e.Class,
			Level:         e.Level,
			HPMax:         e.HpMax,
			HPCurrent:     e.HpCurrent,
			PortraitURL:   e.PortraitURL,
			RaceColor:     raceColors[e.Race],
			CharacterType: e.CharacterType,
			CampaignID:    e.CampaignID,
			Owned:         canEditCharacter(c, e),
		})
	}
	c.JSON(http.StatusOK, out)
}

func ListAllCharacters(c *gin.Context) {
	role, _ := c.Get("role")
	if role != "dm" && role != "admin" {
		WriteError(c, http.StatusForbidden, strErr("dm or admin required"))
		return
	}

	type CharSummary struct {
		ID            int64  `json:"id"`
		UserID        int64  `json:"user_id"`
		Username      string `json:"username"`
		Name          string `json:"name"`
		Race          string `json:"race"`
		Class         string `json:"class"`
		Level         int    `json:"level"`
		HPMax         int    `json:"hp_max"`
		HPCurrent     int    `json:"hp_current"`
		PortraitURL   string `json:"portrait_url,omitempty"`
		RaceColor     string `json:"race_color,omitempty"`
		CharacterType string `json:"character_type"`
		CanEdit       bool   `json:"can_edit"`
	}
	chars := make([]CharSummary, 0)
	raceColors := GetRaceColorMap()

	rows, err := db.DB.Query(`
		SELECT c.id, c.user_id, u.username, c.name, c.race, c.class, c.level, c.hp_max, c.hp_current, COALESCE(c.portrait_url,''), COALESCE(c.character_type,'player')
		FROM characters c
		JOIN users u ON c.user_id = u.id
		ORDER BY c.updated_at DESC`)
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var ch CharSummary
		rows.Scan(&ch.ID, &ch.UserID, &ch.Username, &ch.Name, &ch.Race, &ch.Class, &ch.Level, &ch.HPMax, &ch.HPCurrent, &ch.PortraitURL, &ch.CharacterType)
		ch.RaceColor = raceColors[ch.Race]
		ch.CanEdit = true
		chars = append(chars, ch)
	}
	c.JSON(http.StatusOK, chars)
}

func GetCharacter(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	entChar, err := db.Client.Character.Query().Where(character.ID(id)).Only(c.Request.Context())
	if ent.IsNotFound(err) {
		WriteNotFound(c, "character not found")
		return
	}
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}

	ch := entCharacterToModel(entChar)

	// Authorization — admin, owner, or any campaign member may view;
	// edit rights are computed separately via canEditCharacter.
	role, _ := c.Get("role")
	uidVal, _ := c.Get("user_id")
	uid, _ := uidVal.(int64)
	if role != "admin" && entChar.UserID != uid && !isCampaignMemberOfCharacter(c, id) {
		WriteError(c, http.StatusForbidden, strErr("access denied"))
		return
	}

	// Load sub-resources
	ctx := c.Request.Context()
	ch.Proficiencies = loadProficiencies(ctx, ch.ID)
	ch.Features = loadFeatures(ctx, ch.ID)
	ch.Spellcasting = loadSpellcasting(ctx, ch.ID)
	ch.Spells = loadSpells(ctx, ch.ID)
	ch.Inventory = loadInventory(ctx, ch.ID)
	ch.Currency = loadCurrency(ctx, ch.ID)
	ch.Classes = loadCharClasses(ctx, ch.ID)
	computeMods(ch)
	ch.CanEdit = canEditCharacter(c, entChar)

	c.JSON(http.StatusOK, ch)
}

func CreateCharacter(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var ch models.Character
	if !BindOr400(c, &ch) {
		return
	}

	if strings.TrimSpace(ch.Name) == "" {
		WriteError(c, http.StatusBadRequest, strErr("name is required"))
		return
	}
	if ch.CharacterType == "" {
		ch.CharacterType = "player"
	}
	if ch.CharacterType != "player" && ch.CharacterType != "linked" {
		WriteError(c, http.StatusBadRequest, strErr("character_type must be 'player' or 'linked'"))
		return
	}
	for _, score := range []int{ch.Str, ch.Dex, ch.Con, ch.Int, ch.Wis, ch.Cha} {
		if score < 0 || score > 30 {
			WriteError(c, http.StatusBadRequest, strErr("ability scores must be between 0 and 30"))
			return
		}
	}

	// Set defaults
	if ch.Level < 1 {
		ch.Level = 1
	}
	if ch.AC < 1 {
		ch.AC = 10
	}
	if ch.Speed < 1 {
		ch.Speed = 30
	}
	if ch.ProficiencyBonus < 1 {
		if ch.Level >= 17 {
			ch.ProficiencyBonus = 6
		} else if ch.Level >= 13 {
			ch.ProficiencyBonus = 5
		} else if ch.Level >= 9 {
			ch.ProficiencyBonus = 4
		} else if ch.Level >= 5 {
			ch.ProficiencyBonus = 3
		} else {
			ch.ProficiencyBonus = 2
		}
	}
	if ch.HPMax < 1 {
		ch.HPMax = 10
		ch.HPCurrent = 10
	}
	if ch.HitDice == "" {
		ch.HitDice = "1d10"
		ch.HitDiceCurrent = 1
	}

	uid, _ := userID.(int64)
	now := time.Now().Format("2006-01-02 15:04:05")

	charCreate := db.Client.Character.Create().
		SetUserID(uid).
		SetName(ch.Name).
		SetRace(ch.Race).
		SetClass(ch.Class).
		SetSubclass(ch.Subclass).
		SetLevel(ch.Level).
		SetXp(ch.XP).
		SetBackground(ch.Background).
		SetAlignment(ch.Alignment).
		SetStr(ch.Str).
		SetDex(ch.Dex).
		SetCon(ch.Con).
		SetInt(ch.Int).
		SetWis(ch.Wis).
		SetCha(ch.Cha).
		SetAc(ch.AC).
		SetInitiative(ch.Initiative).
		SetSpeed(ch.Speed).
		SetHpMax(ch.HPMax).
		SetHpCurrent(ch.HPCurrent).
		SetTempHp(ch.TempHP).
		SetHitDice(ch.HitDice).
		SetHitDiceCurrent(ch.HitDiceCurrent).
		SetProficiencyBonus(ch.ProficiencyBonus).
		SetInspiration(ch.Inspiration).
		SetPassivePerception(ch.PassivePerception).
		SetDeathSavesSuccesses(ch.DeathSavesSuccesses).
		SetDeathSavesFailures(ch.DeathSavesFailures).
		SetConcentratingOn(ch.ConcentratingOn).
		SetPersonalityTraits(ch.PersonalityTraits).
		SetIdeals(ch.Ideals).
		SetBonds(ch.Bonds).
		SetFlaws(ch.Flaws).
		SetAppearance(ch.Appearance).
		SetBackstory(ch.Backstory).
		SetPortraitURL(ch.PortraitURL).
		SetCreatedAt(now).
		SetUpdatedAt(now).
		SetCharacterType(ch.CharacterType)

	if ch.CampaignID != nil {
		charCreate.SetCampaignID(*ch.CampaignID)
	}

	char, err := charCreate.Save(c.Request.Context())
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}

	id := char.ID

	// Create default currency entry
	db.Client.CharacterCurrency.Create().SetCharacterID(id).Save(c.Request.Context())

	ch.ID = id
	ch.UserID = uid
	db.DB.Exec("PRAGMA wal_checkpoint(PASSIVE)")
	c.JSON(http.StatusCreated, ch)
}

func UpdateCharacter(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	// Check ownership
	entChar, err := db.Client.Character.Query().Where(character.ID(id)).Select(character.FieldUserID, character.FieldCharacterType).Only(c.Request.Context())
	if ent.IsNotFound(err) {
		WriteNotFound(c, "character not found")
		return
	}
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	if !canEditCharacter(c, entChar) {
		WriteError(c, http.StatusForbidden, strErr("access denied"))
		return
	}

	// Read the raw body once so we can detect whether dm_notes was explicitly
	// sent. The character sheet PUT omits the field, and DM notes must only be
	// written by an admin or the campaign's DM.
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		WriteError(c, http.StatusBadRequest, strErr("invalid request body"))
		return
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		WriteError(c, http.StatusBadRequest, err)
		return
	}
	_, dmNotesSent := raw["dm_notes"]

	var ch models.Character
	if err := json.Unmarshal(body, &ch); err != nil {
		WriteError(c, http.StatusBadRequest, err)
		return
	}
	if ch.CharacterType != "" && ch.CharacterType != "player" && ch.CharacterType != "linked" {
		WriteError(c, http.StatusBadRequest, strErr("character_type must be 'player' or 'linked'"))
		return
	}
	if strings.TrimSpace(ch.Name) == "" || strings.TrimSpace(ch.Race) == "" || strings.TrimSpace(ch.Class) == "" {
		WriteError(c, http.StatusBadRequest, strErr("name, race, and class are required"))
		return
	}
	if ch.DeathSavesSuccesses < 0 || ch.DeathSavesSuccesses > 3 || ch.DeathSavesFailures < 0 || ch.DeathSavesFailures > 3 {
		WriteError(c, http.StatusBadRequest, strErr("death saves must be between 0 and 3"))
		return
	}

	upd := db.Client.Character.UpdateOneID(id).
		SetName(ch.Name).
		SetRace(ch.Race).
		SetClass(ch.Class).
		SetSubclass(ch.Subclass).
		SetLevel(ch.Level).
		SetXp(ch.XP).
		SetBackground(ch.Background).
		SetAlignment(ch.Alignment).
		SetStr(ch.Str).
		SetDex(ch.Dex).
		SetCon(ch.Con).
		SetInt(ch.Int).
		SetWis(ch.Wis).
		SetCha(ch.Cha).
		SetAc(ch.AC).
		SetInitiative(ch.Initiative).
		SetSpeed(ch.Speed).
		SetHpMax(ch.HPMax).
		SetHpCurrent(ch.HPCurrent).
		SetTempHp(ch.TempHP).
		SetHitDice(ch.HitDice).
		SetHitDiceCurrent(ch.HitDiceCurrent).
		SetProficiencyBonus(ch.ProficiencyBonus).
		SetInspiration(ch.Inspiration).
		SetPassivePerception(ch.PassivePerception).
		SetDeathSavesSuccesses(ch.DeathSavesSuccesses).
		SetDeathSavesFailures(ch.DeathSavesFailures).
		SetConcentratingOn(ch.ConcentratingOn).
		SetPersonalityTraits(ch.PersonalityTraits).
		SetIdeals(ch.Ideals).
		SetBonds(ch.Bonds).
		SetFlaws(ch.Flaws).
		SetAppearance(ch.Appearance).
		SetBackstory(ch.Backstory).
		SetPortraitURL(ch.PortraitURL).
		SetUpdatedAt(time.Now().Format("2006-01-02 15:04:05"))

	if ch.CampaignID != nil {
		upd.SetCampaignID(*ch.CampaignID)
	} else {
		upd.ClearCampaignID()
	}

	// DM notes: only persist when explicitly sent by an admin or the campaign DM.
	if dmNotesSent {
		role, _ := c.Get("role")
		if role == "admin" || isDMOfCharacter(c, id) {
			var note string
			if json.Unmarshal(raw["dm_notes"], &note) == nil {
				upd.SetDmNotes(note)
			}
		}
	}

	_, err = upd.Save(c.Request.Context())
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}

	// Auto-calc passive perception
	charStats, err := db.Client.Character.Query().Where(character.ID(id)).Select(character.FieldWis, character.FieldProficiencyBonus).Only(c.Request.Context())
	if err == nil {
		wisMod := abilityMod(charStats.Wis)
		pp := 10 + wisMod
		perceptionCount, _ := db.Client.CharacterProficiency.Query().
			Where(
				characterproficiency.CharacterID(id),
				characterproficiency.TypeEQ("skill"),
				characterproficiency.NameEqualFold("perception"),
			).
			Count(c.Request.Context())
		if perceptionCount > 0 {
			pp += charStats.ProficiencyBonus
		}
		db.Client.Character.UpdateOneID(id).SetPassivePerception(pp).Exec(c.Request.Context())
	}

	SendCharacterUpdate(id)
	SendPartyUpdate()

	updated, err := db.Client.Character.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}
	c.JSON(http.StatusOK, entCharacterToModel(updated))
}

// UpdateCharacterDMNotes persists the DM's private notes for a character.
// Only the campaign DM or an admin may write DM notes.
func UpdateCharacterDMNotes(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	role, _ := c.Get("role")
	if role != "admin" && !isDMOfCharacter(c, id) {
		WriteError(c, http.StatusForbidden, strErr("dm or admin required"))
		return
	}

	var req struct {
		DMNotes string `json:"dm_notes"`
	}
	if !BindOr400(c, &req) {
		return
	}

	_, err := db.Client.Character.UpdateOneID(id).SetDmNotes(req.DMNotes).Save(c.Request.Context())
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}

	SendCharacterUpdate(id)
	SendPartyUpdate()
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func DeleteCharacter(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	entChar, err := db.Client.Character.Query().Where(character.ID(id)).Select(character.FieldUserID, character.FieldCharacterType).Only(c.Request.Context())
	if ent.IsNotFound(err) {
		WriteNotFound(c, "character not found")
		return
	}
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	if !canEditCharacter(c, entChar) {
		WriteError(c, http.StatusForbidden, strErr("access denied"))
		return
	}

	// Delete child records first (Ent auto-migration creates FK constraints with NoAction,
	// so we must delete children manually before the parent character record).
	for _, table := range []string{
		"character_currency", "character_classes", "character_proficiencies",
		"character_locations", "character_npcs", "sessions", "quests", "journal",
		"rest_log", "character_conditions", "character_feats", "companions",
		"faction_reputation", "character_notes", "character_crafting",
		"character_resources", "downtime_activities", "level_up_plans",
		"character_spellcasting", "character_features", "spells", "inventory",
	} {
		db.DB.Exec("DELETE FROM "+table+" WHERE character_id = ?", id)
	}
	if err := db.Client.Character.DeleteOneID(id).Exec(c.Request.Context()); err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func ExportCharacter(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	entChar, err := db.Client.Character.Query().Where(character.ID(id)).Only(c.Request.Context())
	if ent.IsNotFound(err) {
		WriteNotFound(c, "character not found")
		return
	}
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}

	ch := entCharacterToModel(entChar)

	ctx := c.Request.Context()
	ch.Proficiencies = loadProficiencies(ctx, ch.ID)
	ch.Features = loadFeatures(ctx, ch.ID)
	ch.Spellcasting = loadSpellcasting(ctx, ch.ID)
	ch.Spells = loadSpells(ctx, ch.ID)
	ch.Inventory = loadInventory(ctx, ch.ID)
	ch.Currency = loadCurrency(ctx, ch.ID)
	computeMods(ch)

	format := c.DefaultQuery("format", "json")
	if format == "text" {
		c.String(http.StatusOK, characterToText(ch))
		return
	}
	c.JSON(http.StatusOK, ch)
}

func PrintCharacter(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	entChar, err := db.Client.Character.Query().Where(character.ID(id)).Only(c.Request.Context())
	if ent.IsNotFound(err) {
		WriteNotFound(c, "character not found")
		return
	}
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}

	ch := entCharacterToModel(entChar)

	ctx := c.Request.Context()
	ch.Proficiencies = loadProficiencies(ctx, ch.ID)
	ch.Features = loadFeatures(ctx, ch.ID)
	ch.Spellcasting = loadSpellcasting(ctx, ch.ID)
	ch.Spells = loadSpells(ctx, ch.ID)
	ch.Inventory = loadInventory(ctx, ch.ID)
	ch.Currency = loadCurrency(ctx, ch.ID)
	computeMods(ch)

	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.String(http.StatusOK, characterToText(ch))
}

func isDMOfCharacter(c *gin.Context, characterID int64) bool {
	currentUID, _ := c.Get("user_id")
	uid, _ := currentUID.(int64)

	entChar, err := db.Client.Character.Query().
		Where(character.ID(characterID)).
		Select(character.FieldCampaignID).
		Only(c.Request.Context())
	if err != nil {
		return false
	}

	if entChar.CampaignID != 0 {
		count, err := db.Client.CampaignMember.Query().
			Where(
				campaignmember.CampaignIDEQ(entChar.CampaignID),
				campaignmember.UserIDEQ(uid),
				campaignmember.RoleEQ("dm"),
			).
			Count(c.Request.Context())
		return err == nil && count > 0
	}
	return false
}

// isCampaignMemberOfCharacter reports whether the requester is a member of the
// character's campaign (any role).
func isCampaignMemberOfCharacter(c *gin.Context, characterID int64) bool {
	currentUID, _ := c.Get("user_id")
	uid, _ := currentUID.(int64)
	if uid == 0 {
		return false
	}
	entChar, err := db.Client.Character.Query().
		Where(character.ID(characterID)).
		Select(character.FieldCampaignID).
		Only(c.Request.Context())
	if err != nil || entChar.CampaignID == 0 {
		return false
	}
	count, err := db.Client.CampaignMember.Query().
		Where(
			campaignmember.CampaignIDEQ(entChar.CampaignID),
			campaignmember.UserIDEQ(uid),
		).
		Count(c.Request.Context())
	return err == nil && count > 0
}

// canViewCharacter: admin, owner, or any member of the character's campaign.
func canViewCharacter(c *gin.Context, characterID int64) bool {
	role, _ := c.Get("role")
	if role == "admin" {
		return true
	}
	currentUID, _ := c.Get("user_id")
	uid, _ := currentUID.(int64)
	entChar, err := db.Client.Character.Query().
		Where(character.ID(characterID)).
		Select(character.FieldUserID).
		Only(c.Request.Context())
	if err != nil {
		return false
	}
	if entChar.UserID == uid {
		return true
	}
	return isCampaignMemberOfCharacter(c, characterID)
}

// canEditCharacter enforces character_type edit rules:
//   - player: owner, admin, or campaign DM may edit
//   - linked: only admin or campaign DM may edit (owner cannot)
