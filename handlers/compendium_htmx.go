package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/models"
)

// ─── Compendium Monster Browser (HTMX) ───

type htmxCompendiumMonsterListData struct {
	Monsters    []models.CompendiumMonster
	EncounterID int64
	CampaignID  int64
	AdventureID int64
}

type htmxCompendiumMonsterDetailData struct {
	Monster     *models.CompendiumMonster
	EncounterID int64
	CampaignID  int64
	AdventureID int64
}

func HtmxCompendiumMonsterBrowser(c *gin.Context) {
	renderTemplate(c, "compendium_monster_browser.html", nil)
}

func HtmxCompendiumMonsterSearch(c *gin.Context) {
	query := "SELECT id,name,type,size,ac,hp,str,dex,con,int_,wis,cha,cr,source,is_full,saves,skills,damage_vulnerabilities,damage_resistances,damage_immunities,condition_immunities,senses,languages,special_abilities,actions,legendary_actions,description FROM compendium_monsters WHERE 1=1"
	args := []interface{}{}

	if q := c.Query("q"); q != "" {
		query += " AND name LIKE ?"
		args = append(args, "%"+q+"%")
	}
	if cr := c.Query("cr"); cr != "" {
		query += " AND cr=?"
		args = append(args, cr)
	}
	if t := c.Query("type"); t != "" {
		query += " AND type LIKE ?"
		args = append(args, "%"+t+"%")
	}
	query += " ORDER BY name LIMIT 50"

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		c.String(http.StatusInternalServerError, "query error")
		return
	}
	defer rows.Close()

	out := make([]models.CompendiumMonster, 0)
	for rows.Next() {
		var m models.CompendiumMonster
		var isFull int
		rows.Scan(&m.ID, &m.Name, &m.Type, &m.Size, &m.AC, &m.HP,
			&m.Str, &m.Dex, &m.Con, &m.Int, &m.Wis, &m.Cha,
			&m.CR, &m.Source, &isFull,
			&m.Saves, &m.Skills, &m.DamageVulnerabilities, &m.DamageResistances, &m.DamageImmunities, &m.ConditionImmunities,
			&m.Senses, &m.Languages, &m.SpecialAbilities, &m.Actions, &m.LegendaryActions, &m.Description)
		m.IsFull = isFull == 1
		out = append(out, m)
	}

	encounterID, _ := strconv.ParseInt(c.Query("encounter_id"), 10, 64)
	campaignID, _ := strconv.ParseInt(c.Query("campaign_id"), 10, 64)

	renderTemplate(c, "compendium_monster_list_item.html", htmxCompendiumMonsterListData{
		Monsters:    out,
		EncounterID: encounterID,
		CampaignID:  campaignID,
	})
}

func HtmxCompendiumMonsterDetail(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	encounterID, _ := strconv.ParseInt(c.Query("encounter_id"), 10, 64)
	campaignID, _ := strconv.ParseInt(c.Query("campaign_id"), 10, 64)
	adventureID, _ := strconv.ParseInt(c.Query("adventure_id"), 10, 64)

	var m models.CompendiumMonster
	var isFull int
	err := db.DB.QueryRow("SELECT id,name,type,size,ac,hp,str,dex,con,int_,wis,cha,cr,source,is_full,saves,skills,damage_vulnerabilities,damage_resistances,damage_immunities,condition_immunities,senses,languages,special_abilities,actions,legendary_actions,description FROM compendium_monsters WHERE id=?", id).
		Scan(&m.ID, &m.Name, &m.Type, &m.Size, &m.AC, &m.HP,
			&m.Str, &m.Dex, &m.Con, &m.Int, &m.Wis, &m.Cha,
			&m.CR, &m.Source, &isFull,
			&m.Saves, &m.Skills, &m.DamageVulnerabilities, &m.DamageResistances, &m.DamageImmunities, &m.ConditionImmunities,
			&m.Senses, &m.Languages, &m.SpecialAbilities, &m.Actions, &m.LegendaryActions, &m.Description)
	if err != nil {
		c.String(http.StatusNotFound, "monster not found")
		return
	}
	m.IsFull = isFull == 1

	renderTemplate(c, "compendium_monster_detail.html", htmxCompendiumMonsterDetailData{
		Monster:     &m,
		EncounterID: encounterID,
		CampaignID:  campaignID,
		AdventureID: adventureID,
	})
}

// ─── Compendium Monster Picker (HTMX) ───

func HtmxCompendiumMonsterPickerForEncounter(c *gin.Context) {
	encounterID, _ := strconv.ParseInt(c.Param("eid"), 10, 64)
	campaignID, _ := strconv.ParseInt(c.Query("campaign_id"), 10, 64)
	renderTemplate(c, "compendium_monster_browser.html", htmxCompendiumMonsterListData{
		EncounterID: encounterID,
		CampaignID:  campaignID,
	})
}

func HtmxCompendiumMonsterPickerForOneShot(c *gin.Context) {
	adventureID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	renderTemplate(c, "compendium_monster_browser.html", htmxCompendiumMonsterListData{
		AdventureID: adventureID,
	})
}
