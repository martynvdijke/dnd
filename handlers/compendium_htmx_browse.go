package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"villum/db"
	"villum/middleware"
	"villum/models"
)

func HtmxCompendiumSpellBrowse(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	class := strings.TrimSpace(c.Query("class"))
	level := strings.TrimSpace(c.Query("level"))
	school := strings.TrimSpace(c.Query("school"))
	source := strings.TrimSpace(c.Query("source"))
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}
	offset := (page - 1) * pageSize

	// Build query
	baseQuery := `SELECT id,name,level,school,casting_time,"range",components,duration,
		description,higher_levels,classes,source_page,
		COALESCE(system,''),COALESCE(source,''),COALESCE(publisher,'') FROM compendium_spells WHERE 1=1`
	countQuery := "SELECT COUNT(*) FROM compendium_spells WHERE 1=1"
	args := []any{}

	if q != "" {
		clause := " AND name LIKE ?"
		baseQuery += clause
		countQuery += clause
		args = append(args, "%"+q+"%")
	}
	if class != "" {
		clause := " AND classes LIKE ?"
		baseQuery += clause
		countQuery += clause
		args = append(args, "%\""+class+"\"%")
	}
	if level != "" {
		// Support comma-separated levels: "1,2,3" or range "1-3"
		if strings.Contains(level, "-") {
			parts := strings.SplitN(level, "-", 2)
			from, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
			to, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
			if err1 == nil && err2 == nil {
				clause := " AND level>=? AND level<=?"
				baseQuery += clause
				countQuery += clause
				args = append(args, from, to)
			}
		} else {
			levels := strings.Split(level, ",")
			placeholders := make([]string, len(levels))
			for i, l := range levels {
				lv, err := strconv.Atoi(strings.TrimSpace(l))
				if err == nil {
					placeholders[i] = "?"
					args = append(args, lv)
				}
			}
			validPlaceholders := []string{}
			for _, p := range placeholders {
				if p != "" {
					validPlaceholders = append(validPlaceholders, p)
				}
			}
			if len(validPlaceholders) > 0 {
				clause := " AND level IN (" + strings.Join(validPlaceholders, ",") + ")"
				baseQuery += clause
				countQuery += clause
			}
		}
	}
	if school != "" {
		clause := " AND school=?"
		baseQuery += clause
		countQuery += clause
		args = append(args, school)
	}
	if source != "" {
		switch source {
		case "srd":
			clause := " AND system='dnd5e' AND source='srd'"
			baseQuery += clause
			countQuery += clause
		case "homebrew":
			clause := " AND (source='homebrew' OR COALESCE(publisher,'')!='')"
			baseQuery += clause
			countQuery += clause
		case "imported":
			clause := " AND COALESCE(publisher,'')!='' AND source!='srd'"
			baseQuery += clause
			countQuery += clause
		}
	}

	// Get total count
	var totalCount int
	err := db.DB.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		totalCount = 0
	}

	// Get spells for current page
	baseQuery += " ORDER BY level, name LIMIT ? OFFSET ?"
	pageArgs := append(args, pageSize, offset)

	rows, err := db.DB.Query(baseQuery, pageArgs...)
	if err != nil {
		c.String(http.StatusInternalServerError, "query error: %v", err)
		return
	}
	defer rows.Close()

	spells := make([]models.CompendiumSpell, 0)
	for rows.Next() {
		var s models.CompendiumSpell
		err := rows.Scan(&s.ID, &s.Name, &s.Level, &s.School, &s.CastingTime, &s.Range,
			&s.Components, &s.Duration, &s.Description, &s.HigherLevels, &s.Classes,
			&s.SourcePage, &s.System, &s.Source, &s.Publisher)
		if err != nil {
			continue
		}
		spells = append(spells, s)
	}

	totalPages := (totalCount + pageSize - 1) / pageSize
	if totalPages < 1 {
		totalPages = 1
	}

	// Fetch distinct classes and schools for filter dropdowns
	classes := getDistinctSpellClasses()
	schools := getDistinctSpellSchools()

	// Get source counts
	srdCount := getSpellCountBySource("srd")
	homebrewCount := getSpellCountBySource("homebrew")
	importedCount := getSpellCountBySource("imported")

	data := htmxSpellBrowseData{
		Spells:     spells,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
		Query:      q,
		Class:      class,
		Level:      level,
		School:     school,
		Source:     source,
		Classes:    classes,
		Schools:    schools,
	}
	data.SourceCount.SRD = srdCount
	data.SourceCount.Homebrew = homebrewCount
	data.SourceCount.Imported = importedCount

	renderTemplate(c, "compendium_spell_browse", data)
}

func getDistinctSpellClasses() []string {
	classSet := make(map[string]bool)
	addClasses := func(classes string) {
		// Parse JSON array like ["Artificer","Bard","Cleric"]
		cls := strings.Trim(classes, "[]")
		parts := strings.Split(cls, ",")
		for _, p := range parts {
			name := strings.Trim(strings.TrimSpace(p), "\"")
			if name != "" {
				classSet[name] = true
			}
		}
	}
	rows, err := db.DB.Query("SELECT DISTINCT classes FROM compendium_spells WHERE classes != '' ORDER BY classes")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var classes string
			if rows.Scan(&classes) == nil {
				addClasses(classes)
			}
		}
	}
	entryRows, err := db.DB.Query("SELECT DISTINCT json_extract(data,'$.classes') FROM compendium_entries WHERE json_extract(data,'$.classes') IS NOT NULL")
	if err == nil {
		defer entryRows.Close()
		for entryRows.Next() {
			var classes string
			if entryRows.Scan(&classes) == nil {
				addClasses(classes)
			}
		}
	}

	result := make([]string, 0, len(classSet))
	for name := range classSet {
		result = append(result, name)
	}
	sort.Strings(result)
	return result
}

func getDistinctSpellSchools() []string {
	rows, err := db.DB.Query("SELECT DISTINCT school FROM compendium_spells WHERE school != '' ORDER BY school")
	if err != nil {
		return nil
	}
	defer rows.Close()

	var schools []string
	for rows.Next() {
		var school string
		rows.Scan(&school)
		schools = append(schools, school)
	}
	return schools
}

func getSpellCountBySource(sourceType string) int {
	var query string
	switch sourceType {
	case "srd":
		query = "SELECT COUNT(*) FROM compendium_spells WHERE system='dnd5e' AND source='srd'"
	case "homebrew":
		query = "SELECT COUNT(*) FROM compendium_spells WHERE source='homebrew' OR (COALESCE(publisher,'')!='' AND source!='srd')"
	case "imported":
		query = "SELECT COUNT(*) FROM compendium_spells WHERE COALESCE(publisher,'')!='' AND source!='srd'"
	default:
		return 0
	}
	var count int
	db.DB.QueryRow(query).Scan(&count)
	return count
}

// ─── Compendium Race Browse (HTMX) ───

type htmxRaceBrowseData struct {
	Races      []models.CompendiumRace
	TotalCount int
	Page       int
	PageSize   int
	TotalPages int
	Query      string
}

func HtmxCompendiumRaceBrowse(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}
	offset := (page - 1) * pageSize

	baseQuery := `SELECT id,name,description,speed,size,ability_bonuses,traits,languages,source_page,system,source,category,expansion,publisher FROM compendium_races WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM compendium_races WHERE 1=1`
	args := []any{}

	if q != "" {
		clause := " AND name LIKE ?"
		baseQuery += clause
		countQuery += clause
		args = append(args, "%"+q+"%")
	}

	var totalCount int
	err := db.DB.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		totalCount = 0
	}

	baseQuery += " ORDER BY name LIMIT ? OFFSET ?"
	pageArgs := append(args, pageSize, offset)

	rows, err := db.DB.Query(baseQuery, pageArgs...)
	if err != nil {
		c.String(http.StatusInternalServerError, "query error")
		return
	}
	defer rows.Close()

	races := make([]models.CompendiumRace, 0)
	for rows.Next() {
		var r models.CompendiumRace
		if err := rows.Scan(&r.ID, &r.Name, &r.Description, &r.Speed, &r.Size, &r.AbilityBonuses, &r.Traits, &r.Languages, &r.SourcePage, &r.System, &r.Source, &r.Category, &r.Expansion, &r.Publisher); err != nil {
			middleware.LogWarn("compendium", "scan failed, skipping race", "error", err)
			continue
		}
		races = append(races, r)
	}

	totalPages := (totalCount + pageSize - 1) / pageSize
	if totalPages < 1 {
		totalPages = 1
	}

	renderTemplate(c, "compendium_race_browse", htmxRaceBrowseData{
		Races:      races,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
		Query:      q,
	})
}

// ─── Compendium Class Browse (HTMX) ───

type htmxClassBrowseData struct {
	Classes    []models.CompendiumClass
	TotalCount int
	Page       int
	PageSize   int
	TotalPages int
	Query      string
}

func HtmxCompendiumClassBrowse(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}
	offset := (page - 1) * pageSize

	baseQuery := `SELECT id,name,description,hit_die,primary_ability,saving_throws,proficiencies,spellcasting_ability,source_page,system,source,category,expansion,publisher FROM compendium_classes WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM compendium_classes WHERE 1=1`
	args := []any{}

	if q != "" {
		clause := " AND name LIKE ?"
		baseQuery += clause
		countQuery += clause
		args = append(args, "%"+q+"%")
	}

	var totalCount int
	err := db.DB.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		totalCount = 0
	}

	baseQuery += " ORDER BY name LIMIT ? OFFSET ?"
	pageArgs := append(args, pageSize, offset)

	rows, err := db.DB.Query(baseQuery, pageArgs...)
	if err != nil {
		c.String(http.StatusInternalServerError, "query error")
		return
	}
	defer rows.Close()

	classes := make([]models.CompendiumClass, 0)
	for rows.Next() {
		var cl models.CompendiumClass
		if err := rows.Scan(&cl.ID, &cl.Name, &cl.Description, &cl.HitDie, &cl.PrimaryAbility, &cl.SavingThrows, &cl.Proficiencies, &cl.SpellcastingAbility, &cl.SourcePage, &cl.System, &cl.Source, &cl.Category, &cl.Expansion, &cl.Publisher); err != nil {
			middleware.LogWarn("compendium", "scan failed, skipping class", "error", err)
			continue
		}
		classes = append(classes, cl)
	}

	totalPages := (totalCount + pageSize - 1) / pageSize
	if totalPages < 1 {
		totalPages = 1
	}

	renderTemplate(c, "compendium_class_browse", htmxClassBrowseData{
		Classes:    classes,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
		Query:      q,
	})
}

// ─── Compendium Equipment Browse (HTMX) ───

type htmxEquipmentBrowseData struct {
	Items      []models.CompendiumEquipment
	TotalCount int
	Page       int
	PageSize   int
	TotalPages int
	Query      string
	Category   string
	Categories []string
}

func getDistinctEquipmentCategories() []string {
	rows, err := db.DB.Query("SELECT DISTINCT category FROM compendium_equipment WHERE category != '' ORDER BY category")
	if err != nil {
		return nil
	}
	defer rows.Close()
	var categories []string
	for rows.Next() {
		var cat string
		rows.Scan(&cat)
		categories = append(categories, cat)
	}
	return categories
}

func HtmxCompendiumEquipmentBrowse(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	category := strings.TrimSpace(c.Query("category"))
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}
	offset := (page - 1) * pageSize

	baseQuery := `SELECT id,name,category,cost,weight,description,source_page,system,source,item_type,item_rarity,publisher FROM compendium_equipment WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM compendium_equipment WHERE 1=1`
	args := []any{}

	if q != "" {
		clause := " AND name LIKE ?"
		baseQuery += clause
		countQuery += clause
		args = append(args, "%"+q+"%")
	}
	if category != "" {
		clause := " AND category=?"
		baseQuery += clause
		countQuery += clause
		args = append(args, category)
	}

	var totalCount int
	err := db.DB.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		totalCount = 0
	}

	baseQuery += " ORDER BY name LIMIT ? OFFSET ?"
	pageArgs := append(args, pageSize, offset)

	rows, err := db.DB.Query(baseQuery, pageArgs...)
	if err != nil {
		c.String(http.StatusInternalServerError, "query error")
		return
	}
	defer rows.Close()

	items := make([]models.CompendiumEquipment, 0)
	for rows.Next() {
		var e models.CompendiumEquipment
		if err := rows.Scan(&e.ID, &e.Name, &e.Category, &e.Cost, &e.Weight, &e.Description, &e.SourcePage, &e.System, &e.Source, &e.ItemType, &e.ItemRarity, &e.Publisher); err != nil {
			middleware.LogWarn("compendium", "scan failed, skipping equipment", "error", err)
			continue
		}
		items = append(items, e)
	}

	totalPages := (totalCount + pageSize - 1) / pageSize
	if totalPages < 1 {
		totalPages = 1
	}

	categories := getDistinctEquipmentCategories()

	renderTemplate(c, "compendium_equipment_browse", htmxEquipmentBrowseData{
		Items:      items,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
		Query:      q,
		Category:   category,
		Categories: categories,
	})
}

// ─── Compendium Card (HTMX partial) ───
