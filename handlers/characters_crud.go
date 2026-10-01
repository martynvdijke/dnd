package handlers

import (
	"encoding/json"
	"errors"
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
	"villum/ent/campaigncharacter"
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
		ID            int64                `json:"id"`
		UserID        int64                `json:"user_id"`
		Name          string               `json:"name"`
		Race          string               `json:"race"`
		Class         string               `json:"class"`
		Level         int                  `json:"level"`
		HPMax         int                  `json:"hp_max"`
		HPCurrent     int                  `json:"hp_current"`
		PortraitURL   string               `json:"portrait_url,omitempty"`
		RaceColor     string               `json:"race_color,omitempty"`
		CharacterType string               `json:"character_type"`
		CanEdit       bool                 `json:"can_edit"`
		Campaigns     []models.CampaignRef `json:"campaigns"`
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
				ch := CharSummary{ID: e.ID, UserID: e.UserID, Name: e.Name, Race: e.Race, Class: e.Class, Level: e.Level, HPMax: e.HpMax, HPCurrent: e.HpCurrent, PortraitURL: e.PortraitURL, CharacterType: e.CharacterType, CanEdit: canEditCharacter(c, e)}
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
			ch := CharSummary{ID: e.ID, UserID: e.UserID, Name: e.Name, Race: e.Race, Class: e.Class, Level: e.Level, HPMax: e.HpMax, HPCurrent: e.HpCurrent, PortraitURL: e.PortraitURL, CharacterType: e.CharacterType, CanEdit: canEditCharacter(c, e)}
			ch.RaceColor = raceColors[ch.Race]
			chars = append(chars, ch)
		}
	}
	ids := make([]int64, 0, len(chars))
	for _, ch := range chars {
		ids = append(ids, ch.ID)
	}
	refs, err := characterCampaignRefs(c.Request.Context(), ids)
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	for i := range chars {
		if refs[chars[i].ID] == nil {
			chars[i].Campaigns = []models.CampaignRef{}
		} else {
			chars[i].Campaigns = refs[chars[i].ID]
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
		ID            int64                `json:"id"`
		UserID        int64                `json:"user_id"`
		Name          string               `json:"name"`
		Race          string               `json:"race"`
		Class         string               `json:"class"`
		Level         int                  `json:"level"`
		HPMax         int                  `json:"hp_max"`
		HPCurrent     int                  `json:"hp_current"`
		PortraitURL   string               `json:"portrait_url,omitempty"`
		RaceColor     string               `json:"race_color,omitempty"`
		CharacterType string               `json:"character_type"`
		Campaigns     []models.CampaignRef `json:"campaigns"`
		Owned         bool                 `json:"owned"`
	}

	links, err := db.Client.CampaignCharacter.Query().
		Where(campaigncharacter.CampaignID(campaignID)).
		Select(campaigncharacter.FieldCharacterID).
		All(ctx)
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	if len(links) == 0 {
		c.JSON(http.StatusOK, []CampaignCharSummary{})
		return
	}
	charIDs := make([]int64, 0, len(links))
	for _, link := range links {
		charIDs = append(charIDs, link.CharacterID)
	}
	chars, err := db.Client.Character.Query().
		Where(character.IDIn(charIDs...)).
		Order(ent.Desc(character.FieldUpdatedAt)).
		All(ctx)
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	refs, err := characterCampaignRefs(ctx, charIDs)
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}

	raceColors := GetRaceColorMap()
	out := make([]CampaignCharSummary, 0, len(chars))
	for _, e := range chars {
		campaigns := refs[e.ID]
		if campaigns == nil {
			campaigns = []models.CampaignRef{}
		}
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
			Campaigns:     campaigns,
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
	ch.Campaigns = characterCampaigns(ctx, ch.ID)
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
	uid, _ := userID.(int64)

	var ch models.Character
	if !BindOr400(c, &ch) {
		return
	}

	created, err := createCharacterCore(c.Request.Context(), uid, ch)
	if err != nil {
		var se *statusError
		if errors.As(err, &se) {
			WriteError(c, se.Status, se)
			return
		}
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusCreated, created)
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

	if _, ok := raw["campaign_ids"]; ok {
		uidVal, _ := c.Get("user_id")
		uid, _ := uidVal.(int64)
		role, _ := c.Get("role")
		if role != "admin" && entChar.UserID != uid {
			WriteError(c, http.StatusForbidden, strErr("only the character owner can change campaign memberships"))
			return
		}
		for _, cid := range ch.CampaignIDs {
			if cid != 0 && !isCampaignMember(c, cid) {
				WriteError(c, http.StatusForbidden, strErr("cannot add character to a campaign you are not a member of"))
				return
			}
		}
		currentIDs, err := characterMemberCampaignIDs(c.Request.Context(), id)
		if err != nil {
			WriteError(c, http.StatusInternalServerError, err)
			return
		}
		want := make(map[int64]bool, len(ch.CampaignIDs))
		for _, cid := range ch.CampaignIDs {
			if cid != 0 {
				want[cid] = true
			}
		}
		for _, cid := range currentIDs {
			if !want[cid] && !isCampaignMember(c, cid) {
				WriteError(c, http.StatusForbidden, strErr("cannot remove a membership in a campaign you are not a member of"))
				return
			}
		}
		if err := replaceCharacterCampaigns(c.Request.Context(), id, ch.CampaignIDs); err != nil {
			WriteError(c, http.StatusInternalServerError, err)
			return
		}
	}

	_, err = upd.Save(c.Request.Context())
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	if _, ok := raw["condition_immunities"]; ok {
		db.DB.Exec("UPDATE characters SET condition_immunities=? WHERE id=?", ch.ConditionImmunities, id)
	}
	if _, ok := raw["damage_resistances"]; ok {
		db.DB.Exec("UPDATE characters SET damage_resistances=? WHERE id=?", ch.DamageResistances, id)
	}
	if _, ok := raw["damage_vulnerabilities"]; ok {
		db.DB.Exec("UPDATE characters SET damage_vulnerabilities=? WHERE id=?", ch.DamageVulnerabilities, id)
	}
	if _, ok := raw["damage_immunities"]; ok {
		db.DB.Exec("UPDATE characters SET damage_immunities=? WHERE id=?", ch.DamageImmunities, id)
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

	// Armor Class follows equipped armor; manual AC is kept when unarmored.
	recomputeCharacterAC(id)

	updated, err := db.Client.Character.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}
	updatedChar := entCharacterToModel(updated)
	updatedChar.Campaigns = characterCampaigns(c.Request.Context(), id)
	c.JSON(http.StatusOK, updatedChar)
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
	ch.Campaigns = characterCampaigns(ctx, ch.ID)
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
	ch.Campaigns = characterCampaigns(ctx, ch.ID)
	ch.Proficiencies = loadProficiencies(ctx, ch.ID)
	ch.Features = loadFeatures(ctx, ch.ID)
	ch.Spellcasting = loadSpellcasting(ctx, ch.ID)
	ch.Spells = loadSpells(ctx, ch.ID)
	ch.Inventory = loadInventory(ctx, ch.ID)
	ch.Currency = loadCurrency(ctx, ch.ID)
	computeMods(ch)

	if c.DefaultQuery("format", "text") == "html" {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, characterToHTML(ch))
		return
	}

	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.String(http.StatusOK, characterToText(ch))
}

func isDMOfCharacter(c *gin.Context, characterID int64) bool {
	currentUID, _ := c.Get("user_id")
	uid, _ := currentUID.(int64)
	if uid == 0 {
		return false
	}

	// The character is run/played by the caller if they own any campaign the
	// character is a member of, or hold the "dm" role in one.
	count, err := db.Client.CampaignCharacter.Query().
		Where(
			campaigncharacter.CharacterID(characterID),
			campaigncharacter.HasCampaignWith(
				campaign.Or(
					campaign.UserID(uid),
					campaign.HasMembersWith(
						campaignmember.UserID(uid),
						campaignmember.RoleEQ("dm"),
					),
				),
			),
		).
		Count(c.Request.Context())
	return err == nil && count > 0
}

// isCampaignMemberOfCharacter reports whether the requester is a member of any
// campaign the character belongs to (any role).
func isCampaignMemberOfCharacter(c *gin.Context, characterID int64) bool {
	currentUID, _ := c.Get("user_id")
	uid, _ := currentUID.(int64)
	if uid == 0 {
		return false
	}
	count, err := db.Client.CampaignCharacter.Query().
		Where(
			campaigncharacter.CharacterID(characterID),
			campaigncharacter.HasCampaignWith(
				campaign.Or(
					campaign.UserID(uid),
					campaign.HasMembersWith(campaignmember.UserID(uid)),
				),
			),
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
