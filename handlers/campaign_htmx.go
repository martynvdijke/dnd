package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/models"
)

// ─── Campaign NPCs Section (HTMX) ───

type htmxCampaignNPCData struct {
	CampaignID int64
	NPCs       []models.CampaignNPC
}

func HtmxCampaignNPCsSection(c *gin.Context) {
	campaignID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	rows, err := db.DB.Query(`
		SELECT cn.id, cn.campaign_id, cn.npc_id, COALESCE(cn.role,''), COALESCE(cn.notes,''), COALESCE(cn.created_at,''),
		       COALESCE(n.name,''), COALESCE(n.race,''), COALESCE(n.class,'')
		FROM campaign_npcs cn
		JOIN npcs n ON n.id = cn.npc_id
		WHERE cn.campaign_id=? ORDER BY n.name`, campaignID)
	if err != nil {
		c.String(http.StatusInternalServerError, "query error")
		return
	}
	defer rows.Close()
	out := make([]models.CampaignNPC, 0)
	for rows.Next() {
		var npc models.CampaignNPC
		rows.Scan(&npc.ID, &npc.CampaignID, &npc.NPCID, &npc.Role, &npc.Notes, &npc.CreatedAt,
			&npc.NPCName, &npc.NPCRace, &npc.NPCClass)
		out = append(out, npc)
	}
	renderTemplate(c, "campaign_npcs_section.html", htmxCampaignNPCData{
		CampaignID: campaignID,
		NPCs:       out,
	})
}

// ─── Campaign Encounters Section (HTMX) ───

type htmxCampaignEncounterData struct {
	CampaignID int64
	Encounters []models.EncounterTemplate
}

func HtmxCampaignEncountersSection(c *gin.Context) {
	campaignID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	rows, err := db.DB.Query("SELECT id, campaign_id, user_id, name, description, environment, difficulty, xp_budget, total_xp, notes, created_at FROM encounter_templates WHERE campaign_id=? ORDER BY created_at DESC", campaignID)
	if err != nil {
		c.String(http.StatusInternalServerError, "query error")
		return
	}
	defer rows.Close()
	out := make([]models.EncounterTemplate, 0)
	for rows.Next() {
		var e models.EncounterTemplate
		rows.Scan(&e.ID, &e.CampaignID, &e.UserID, &e.Name, &e.Description, &e.Environment, &e.Difficulty, &e.XPBudget, &e.TotalXP, &e.Notes, &e.CreatedAt)
		// Load monsters for count display
		mrows, err := db.DB.Query("SELECT id, encounter_id, name, count, cr, xp, ac, hp, initiative_mod, source, notes, compendium_monster_id FROM encounter_monsters WHERE encounter_id=?", e.ID)
		if err == nil {
			e.Monsters = make([]models.EncounterMonster, 0)
			for mrows.Next() {
				var m models.EncounterMonster
				mrows.Scan(&m.ID, &m.EncounterID, &m.Name, &m.Count, &m.CR, &m.XP, &m.AC, &m.HP, &m.InitiativeMod, &m.Source, &m.Notes, &m.CompendiumMonsterID)
				e.Monsters = append(e.Monsters, m)
			}
			mrows.Close()
		}
		out = append(out, e)
	}
	renderTemplate(c, "campaign_encounters_section.html", htmxCampaignEncounterData{
		CampaignID: campaignID,
		Encounters: out,
	})
}

// ─── Campaign Encounter Monsters (HTMX) ───

type htmxEncounterMonsterData struct {
	CampaignID    int64
	EncounterID   int64
	EncounterName string
	Monsters      []models.EncounterMonster
}

func HtmxCampaignEncounterMonsters(c *gin.Context) {
	encounterID, _ := strconv.ParseInt(c.Param("eid"), 10, 64)
	campaignIDStr := c.Query("campaign_id")
	campaignID, _ := strconv.ParseInt(campaignIDStr, 10, 64)

	var encounterName string
	db.DB.QueryRow("SELECT name FROM encounter_templates WHERE id=?", encounterID).Scan(&encounterName)

	rows, err := db.DB.Query("SELECT id, encounter_id, name, count, cr, xp, ac, hp, initiative_mod, source, notes, compendium_monster_id FROM encounter_monsters WHERE encounter_id=? ORDER BY id", encounterID)
	if err != nil {
		c.String(http.StatusInternalServerError, "query error")
		return
	}
	defer rows.Close()
	out := make([]models.EncounterMonster, 0)
	for rows.Next() {
		var m models.EncounterMonster
		rows.Scan(&m.ID, &m.EncounterID, &m.Name, &m.Count, &m.CR, &m.XP, &m.AC, &m.HP, &m.InitiativeMod, &m.Source, &m.Notes, &m.CompendiumMonsterID)
		out = append(out, m)
	}
	renderTemplate(c, "campaign_encounter_monsters.html", htmxEncounterMonsterData{
		CampaignID:    campaignID,
		EncounterID:   encounterID,
		EncounterName: encounterName,
		Monsters:      out,
	})
}

// ─── Campaign Encounter Monsters List (HTMX) ───

type htmxCampaignEncounterMonsterListData struct {
	CampaignID int64
	Monsters   []models.EncounterMonster
}

func HtmxCampaignEncounterMonsterList(c *gin.Context) {
	encounterID, _ := strconv.ParseInt(c.Param("eid"), 10, 64)
	campaignIDStr := c.Query("campaign_id")
	campaignID, _ := strconv.ParseInt(campaignIDStr, 10, 64)

	rows, err := db.DB.Query("SELECT id, encounter_id, name, count, cr, xp, ac, hp, initiative_mod, source, notes, compendium_monster_id FROM encounter_monsters WHERE encounter_id=? ORDER BY id", encounterID)
	if err != nil {
		c.String(http.StatusInternalServerError, "query error")
		return
	}
	defer rows.Close()
	out := make([]models.EncounterMonster, 0)
	for rows.Next() {
		var m models.EncounterMonster
		rows.Scan(&m.ID, &m.EncounterID, &m.Name, &m.Count, &m.CR, &m.XP, &m.AC, &m.HP, &m.InitiativeMod, &m.Source, &m.Notes, &m.CompendiumMonsterID)
		out = append(out, m)
	}
	renderTemplate(c, "campaign_encounter_monsters.html", htmxEncounterMonsterData{
		CampaignID:  campaignID,
		EncounterID: encounterID,
		Monsters:    out,
	})
}

// ─── New Encounter Form (HTMX) ───

type htmxNewEncounterFormData struct {
	CampaignID int64
}

func HtmxNewEncounterForm(c *gin.Context) {
	campaignID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	renderTemplate(c, "encounter_form.html", htmxNewEncounterFormData{CampaignID: campaignID})
}

// ─── Create Encounter (HTMX) ───

func HtmxCreateEncounter(c *gin.Context) {
	userID, _ := c.Get("user_id")
	campaignID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	name := c.PostForm("name")
	if name == "" {
		c.String(http.StatusBadRequest, "name required")
		return
	}
	_, err := db.DB.Exec("INSERT INTO encounter_templates(campaign_id,user_id,name,description,environment,difficulty,xp_budget,total_xp,notes) VALUES(?,?,?,?,?,?,?,?,?)",
		campaignID, userID, name, c.PostForm("description"), c.PostForm("environment"), c.PostForm("difficulty"), 0, 0, c.PostForm("notes"))
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	// Re-render the encounters section
	c.Request.URL.RawQuery = ""
	HtmxCampaignEncountersSection(c)
}

// ─── Delete Encounter (HTMX) ───

func HtmxDeleteEncounter(c *gin.Context) {
	encounterID, _ := strconv.ParseInt(c.Param("eid"), 10, 64)
	db.DB.Exec("DELETE FROM encounter_templates WHERE id=?", encounterID)
	c.Request.URL.RawQuery = ""
	HtmxCampaignEncountersSection(c)
}

// ─── Add Encounter Monster (HTMX) ───

type htmxMonsterFormData struct {
	EncounterID int64
	CampaignID  int64
	MonsterID   int64
	Monster     *models.EncounterMonster
}

func HtmxAddEncounterMonsterForm(c *gin.Context) {
	encounterID, _ := strconv.ParseInt(c.Param("eid"), 10, 64)
	campaignID, _ := strconv.ParseInt(c.Query("campaign_id"), 10, 64)
	renderTemplate(c, "encounter_monster_form.html", htmxMonsterFormData{
		EncounterID: encounterID,
		CampaignID:  campaignID,
	})
}

func HtmxCreateEncounterMonster(c *gin.Context) {
	encounterID, _ := strconv.ParseInt(c.Param("eid"), 10, 64)
	campaignID, _ := strconv.ParseInt(c.PostForm("campaign_id"), 10, 64)
	name := c.PostForm("name")
	if name == "" {
		c.String(http.StatusBadRequest, "name required")
		return
	}
	cr := c.PostForm("cr")
	source := c.PostForm("source")
	if source == "" {
		source = "homebrew"
	}
	_, err := db.DB.Exec("INSERT INTO encounter_monsters(encounter_id,name,count,cr,xp,ac,hp,initiative_mod,source,notes) VALUES(?,?,?,?,?,?,?,?,?,?)",
		encounterID, name, getIntParam(c, "count", 1), cr, getIntParam(c, "xp", 0), getIntParam(c, "ac", 10), getIntParam(c, "hp", 10), getIntParam(c, "initiative_mod", 0), source, c.PostForm("notes"))
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.Request.URL.RawQuery = "campaign_id=" + strconv.FormatInt(campaignID, 10)
	HtmxCampaignEncounterMonsters(c)
}

// ─── Edit Encounter Monster Form (HTMX) ───

func HtmxEditEncounterMonsterForm(c *gin.Context) {
	monsterID, _ := strconv.ParseInt(c.Param("mid"), 10, 64)
	encounterID, _ := strconv.ParseInt(c.Param("eid"), 10, 64)
	campaignID, _ := strconv.ParseInt(c.Query("campaign_id"), 10, 64)

	var m models.EncounterMonster
	err := db.DB.QueryRow("SELECT id, encounter_id, name, count, cr, xp, ac, hp, initiative_mod, source, notes, compendium_monster_id FROM encounter_monsters WHERE id=?", monsterID).
		Scan(&m.ID, &m.EncounterID, &m.Name, &m.Count, &m.CR, &m.XP, &m.AC, &m.HP, &m.InitiativeMod, &m.Source, &m.Notes, &m.CompendiumMonsterID)
	if err != nil {
		c.String(http.StatusNotFound, "not found")
		return
	}
	renderTemplate(c, "encounter_monster_form.html", htmxMonsterFormData{
		MonsterID:   monsterID,
		EncounterID: encounterID,
		CampaignID:  campaignID,
		Monster:     &m,
	})
}

func HtmxUpdateEncounterMonster(c *gin.Context) {
	monsterID, _ := strconv.ParseInt(c.Param("mid"), 10, 64)
	campaignID, _ := strconv.ParseInt(c.PostForm("campaign_id"), 10, 64)
	name := c.PostForm("name")
	if name == "" {
		c.String(http.StatusBadRequest, "name required")
		return
	}
	db.DB.Exec("UPDATE encounter_monsters SET name=?, count=?, cr=?, xp=?, ac=?, hp=?, initiative_mod=?, source=?, notes=? WHERE id=?",
		name, getIntParam(c, "count", 1), c.PostForm("cr"), getIntParam(c, "xp", 0), getIntParam(c, "ac", 10), getIntParam(c, "hp", 10), getIntParam(c, "initiative_mod", 0), c.PostForm("source"), c.PostForm("notes"), monsterID)
	c.Request.URL.RawQuery = "campaign_id=" + strconv.FormatInt(campaignID, 10)
	HtmxCampaignEncounterMonsters(c)
}

func HtmxDeleteEncounterMonster(c *gin.Context) {
	monsterID, _ := strconv.ParseInt(c.Param("mid"), 10, 64)
	campaignID, _ := strconv.ParseInt(c.Query("campaign_id"), 10, 64)
	db.DB.Exec("DELETE FROM encounter_monsters WHERE id=?", monsterID)
	c.Request.URL.RawQuery = "campaign_id=" + strconv.FormatInt(campaignID, 10)
	HtmxCampaignEncounterMonsters(c)
}

// ─── Campaign NPC Link Form (HTMX) ───

type htmxCampaignNPCLinkData struct {
	CampaignID int64
	AllNPCs    []models.NPC
}

func HtmxCampaignNPCLinkForm(c *gin.Context) {
	campaignID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID, _ := c.Get("user_id")
	rows, err := db.DB.Query("SELECT id, name, race, class, description, notes FROM npcs WHERE user_id=? ORDER BY name", userID)
	if err != nil {
		c.String(http.StatusInternalServerError, "query error")
		return
	}
	defer rows.Close()
	var all []models.NPC
	for rows.Next() {
		var n models.NPC
		rows.Scan(&n.ID, &n.Name, &n.Race, &n.Class, &n.Description, &n.Notes)
		all = append(all, n)
	}
	renderTemplate(c, "campaign_npc_link_form.html", htmxCampaignNPCLinkData{
		CampaignID: campaignID,
		AllNPCs:    all,
	})
}

func HtmxCampaignLinkNPC(c *gin.Context) {
	campaignID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	npcID := c.PostForm("npc_id")
	role := c.PostForm("role")
	if npcID != "" {
		db.DB.Exec("INSERT INTO campaign_npcs(campaign_id, npc_id, role) VALUES(?,?,?)", campaignID, npcID, role)
	}
	HtmxCampaignNPCsSection(c)
}

func HtmxCampaignUnlinkNPC(c *gin.Context) {
	linkID, _ := strconv.ParseInt(c.Param("nid"), 10, 64)
	var campaignID int64
	db.DB.QueryRow("SELECT campaign_id FROM campaign_npcs WHERE id=?", linkID).Scan(&campaignID)
	db.DB.Exec("DELETE FROM campaign_npcs WHERE id=?", linkID)
	c.Request.URL.RawQuery = ""
	HtmxCampaignNPCsSection(c)
}

// ─── Campaign NPC Create Form (HTMX) ───

type htmxCampaignNPCCreateData struct {
	CampaignID int64
}

func HtmxCampaignNPCCreateForm(c *gin.Context) {
	campaignID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	renderTemplate(c, "campaign_npc_create_form.html", htmxCampaignNPCCreateData{CampaignID: campaignID})
}

func HtmxCampaignCreateAndLinkNPC(c *gin.Context) {
	campaignID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID, _ := c.Get("user_id")
	name := c.PostForm("name")
	if name == "" {
		c.String(http.StatusBadRequest, "name required")
		return
	}
	result, err := db.DB.Exec("INSERT INTO npcs(user_id,name,race,class,description,notes) VALUES(?,?,?,?,?,?)",
		userID, name, c.PostForm("race"), c.PostForm("class"), c.PostForm("description"), "")
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	npcID, _ := result.LastInsertId()
	db.DB.Exec("INSERT INTO campaign_npcs(campaign_id, npc_id, role, notes) VALUES(?,?,?,?)",
		campaignID, npcID, c.PostForm("role"), c.PostForm("notes"))
	HtmxCampaignNPCsSection(c)
}

// ─── Import Compendium Monster to Encounter (HTMX) ───

func HtmxImportCompendiumMonsterToEncounter(c *gin.Context) {
	encounterID, _ := strconv.ParseInt(c.Param("eid"), 10, 64)
	compendiumID, _ := strconv.ParseInt(c.PostForm("compendium_monster_id"), 10, 64)
	campaignID, _ := strconv.ParseInt(c.PostForm("campaign_id"), 10, 64)
	count := getIntParam(c, "count", 1)

	// Fetch compendium monster
	var name, cr, source string
	var ac, hp int
	err := db.DB.QueryRow("SELECT name, ac, hp, cr, source FROM compendium_monsters WHERE id=?", compendiumID).
		Scan(&name, &ac, &hp, &cr, &source)
	if err != nil {
		c.String(http.StatusNotFound, "compendium monster not found")
		return
	}

	_, err = db.DB.Exec("INSERT INTO encounter_monsters(encounter_id, name, count, cr, ac, hp, source, notes, compendium_monster_id) VALUES(?,?,?,?,?,?,?,?,?)",
		encounterID, name, count, cr, ac, hp, source, "", compendiumID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.Request.URL.RawQuery = "campaign_id=" + strconv.FormatInt(campaignID, 10)
	HtmxCampaignEncounterMonsters(c)
}

// ─── Import Compendium Monster to One-Shot (HTMX) ───

func HtmxImportCompendiumMonsterToOneShot(c *gin.Context) {
	adventureID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	compendiumID, _ := strconv.ParseInt(c.PostForm("compendium_monster_id"), 10, 64)

	var cm models.CompendiumMonster
	var isFull int
	err := db.DB.QueryRow("SELECT id,name,type,size,ac,hp,str,dex,con,int_,wis,cha,cr,source,is_full,saves,skills,damage_vulnerabilities,damage_resistances,damage_immunities,condition_immunities,senses,languages,special_abilities,actions,legendary_actions,description FROM compendium_monsters WHERE id=?", compendiumID).
		Scan(&cm.ID, &cm.Name, &cm.Type, &cm.Size, &cm.AC, &cm.HP,
			&cm.Str, &cm.Dex, &cm.Con, &cm.Int, &cm.Wis, &cm.Cha,
			&cm.CR, &cm.Source, &isFull,
			&cm.Saves, &cm.Skills, &cm.DamageVulnerabilities, &cm.DamageResistances, &cm.DamageImmunities, &cm.ConditionImmunities,
			&cm.Senses, &cm.Languages, &cm.SpecialAbilities, &cm.Actions, &cm.LegendaryActions, &cm.Description)
	if err != nil {
		c.String(http.StatusNotFound, "compendium monster not found")
		return
	}

	isFullVal := 0
	if isFull == 1 {
		isFullVal = 1
	}

	_, err = db.DB.Exec(`INSERT INTO oneshot_monsters(adventure_id, name, ac, hp, str, dex, con, int_, wis, cha, cr, source, is_full,
		saves, skills, damage_vulnerabilities, damage_resistances, damage_immunities, condition_immunities, senses, languages,
		special_abilities, actions, legendary_actions, compendium_monster_id) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		adventureID, cm.Name, cm.AC, cm.HP, cm.Str, cm.Dex, cm.Con, cm.Int, cm.Wis, cm.Cha,
		cm.CR, cm.Source, isFullVal,
		cm.Saves, cm.Skills, cm.DamageVulnerabilities, cm.DamageResistances, cm.DamageImmunities, cm.ConditionImmunities,
		cm.Senses, cm.Languages, cm.SpecialAbilities, cm.Actions, cm.LegendaryActions, compendiumID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	// Re-render the monsters section
	HtmxOneShotMonsters(c)
}

// ─── Campaign Monster Roster (HTMX) ───

type htmxCampaignRosterData struct {
	CampaignID int64
	Monsters   []models.CampaignMonsterRoster
	CanAdd     bool
}

type htmxCampaignRosterAddData struct {
	CampaignID        int64
	CompendiumMonster models.CompendiumMonster
}

// HtmxCampaignMonsterRoster lists all monsters in a campaign's roster
func HtmxCampaignMonsterRoster(c *gin.Context) {
	campaignID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	rows, err := db.DB.Query(`
		SELECT id, campaign_id, COALESCE(compendium_monster_id,0), COALESCE(library_monster_id,0),
		       name, ac, hp, str, dex, con, int_, wis, cha, cr, is_full,
		       saves, skills, damage_vulnerabilities, damage_resistances, damage_immunities,
		       condition_immunities, senses, languages, special_abilities, actions, legendary_actions,
		       description, source, notes, created_at
		FROM campaign_monster_roster
		WHERE campaign_id=? ORDER BY name`, campaignID)
	if err != nil {
		c.String(http.StatusInternalServerError, "query error")
		return
	}
	defer rows.Close()

	out := make([]models.CampaignMonsterRoster, 0)
	for rows.Next() {
		var m models.CampaignMonsterRoster
		var isFull int
		rows.Scan(&m.ID, &m.CampaignID, &m.CompendiumMonsterID, &m.LibraryMonsterID,
			&m.Name, &m.AC, &m.HP, &m.Str, &m.Dex, &m.Con, &m.Int, &m.Wis, &m.Cha,
			&m.CR, &isFull,
			&m.Saves, &m.Skills, &m.DamageVulnerabilities, &m.DamageResistances, &m.DamageImmunities,
			&m.ConditionImmunities, &m.Senses, &m.Languages, &m.SpecialAbilities, &m.Actions, &m.LegendaryActions,
			&m.Description, &m.Source, &m.Notes, &m.CreatedAt)
		m.IsFull = isFull == 1
		if m.CompendiumMonsterID != nil && *m.CompendiumMonsterID == 0 {
			m.CompendiumMonsterID = nil
		}
		if m.LibraryMonsterID != nil && *m.LibraryMonsterID == 0 {
			m.LibraryMonsterID = nil
		}
		out = append(out, m)
	}

	renderTemplate(c, "campaign_monster_roster.html", htmxCampaignRosterData{
		CampaignID: campaignID,
		Monsters:   out,
		CanAdd:     true,
	})
}

// HtmxAddCampaignMonster adds a compendium or library monster to the campaign roster
func HtmxAddCampaignMonster(c *gin.Context) {
	campaignID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	compendiumID, _ := strconv.ParseInt(c.PostForm("compendium_monster_id"), 10, 64)
	libraryID, _ := strconv.ParseInt(c.PostForm("library_monster_id"), 10, 64)

	var table, idCol string
	var id int64
	if compendiumID > 0 {
		table = "compendium_monsters"
		idCol = "id"
		id = compendiumID
	} else if libraryID > 0 {
		table = "monster_library"
		idCol = "id"
		id = libraryID
	} else {
		c.String(http.StatusBadRequest, "must provide compendium_monster_id or library_monster_id")
		return
	}

	// Fetch monster fields
	var name, cr, source, saves, skills, dv, dr, di, ci, senses, lang, sa, actions, la, desc string
	var ac, hp, str, dex, con, int_, wis, cha, isFull int
	err := db.DB.QueryRow(fmt.Sprintf(`
		SELECT name, ac, hp, str, dex, con, int_, wis, cha, cr, source, is_full,
		       saves, skills, damage_vulnerabilities, damage_resistances, damage_immunities,
		       condition_immunities, senses, languages, special_abilities, actions, legendary_actions, description
		FROM %s WHERE %s=?`, table, idCol), id).
		Scan(&name, &ac, &hp, &str, &dex, &con, &int_, &wis, &cha,
			&cr, &source, &isFull,
			&saves, &skills, &dv, &dr, &di, &ci,
			&senses, &lang, &sa, &actions, &la, &desc)
	if err != nil {
		c.String(http.StatusNotFound, "monster not found")
		return
	}

	var libMonID interface{}
	if libraryID > 0 {
		libMonID = libraryID
	} else {
		libMonID = nil
	}

	var compMonID interface{}
	if compendiumID > 0 {
		compMonID = compendiumID
	} else {
		compMonID = nil
	}

	_, err = db.DB.Exec(`
		INSERT INTO campaign_monster_roster(campaign_id, compendium_monster_id, library_monster_id,
			name, ac, hp, str, dex, con, int_, wis, cha, cr, is_full,
			saves, skills, damage_vulnerabilities, damage_resistances, damage_immunities,
			condition_immunities, senses, languages, special_abilities, actions, legendary_actions,
			description, source, notes)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,'')`,
		campaignID, compMonID, libMonID,
		name, ac, hp, str, dex, con, int_, wis, cha, cr, isFull,
		saves, skills, dv, dr, di, ci,
		senses, lang, sa, actions, la, desc, source)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	HtmxCampaignMonsterRoster(c)
}

// HtmxRemoveCampaignMonster removes a monster from the campaign roster
func HtmxRemoveCampaignMonster(c *gin.Context) {
	campaignID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	rosterID, _ := strconv.ParseInt(c.Param("rid"), 10, 64)

	_, err := db.DB.Exec("DELETE FROM campaign_monster_roster WHERE id=? AND campaign_id=?", rosterID, campaignID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	HtmxCampaignMonsterRoster(c)
}

// NullInt64 helper
func nullInt64(ni sql.NullInt64) *int64 {
	if ni.Valid {
		return &ni.Int64
	}
	return nil
}

// NullString helper
func nullString(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}
