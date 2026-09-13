package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"villum/db"
	"villum/models"
)

type CharacterSource interface {
	Parse(raw []byte) ([]models.ImportCharacter, error)
}
type CompendiumSource interface {
	Parse(raw []byte) ([]map[string]any, error)
}

type dndBeyondSource struct{}
type foundryActorSource struct{}
type fiveEToolsSource struct{}
type foundryPackSource struct{}

var characterSources = map[string]CharacterSource{"dndbeyond": dndBeyondSource{}, "foundry-actor": foundryActorSource{}}
var compendiumSources = map[string]CompendiumSource{"5etools": fiveEToolsSource{}, "foundry-pack": foundryPackSource{}}

func (dndBeyondSource) Parse(raw []byte) ([]models.ImportCharacter, error) {
	var top map[string]any
	if err := json.Unmarshal(raw, &top); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	var obj map[string]any
	if data, ok := top["data"].(map[string]any); ok {
		obj = data
	} else {
		obj = top
	}
	// If obj is empty and original wasn't object? Already handled.
	name, _ := obj["name"].(string)

	// stats array [{id,value}]
	str, dex, con, iv, wis, cha := 10, 10, 10, 10, 10, 10
	if stats, ok := obj["stats"].([]any); ok {
		for _, s := range stats {
			m, ok := s.(map[string]any)
			if !ok {
				continue
			}
			idF, _ := m["id"].(float64)
			valF, _ := m["value"].(float64)
			switch int(idF) {
			case 1:
				str = int(valF)
			case 2:
				dex = int(valF)
			case 3:
				con = int(valF)
			case 4:
				iv = int(valF)
			case 5:
				wis = int(valF)
			case 6:
				cha = int(valF)
			}
		}
	}
	// class and level
	className := ""
	level := 0
	if classes, ok := obj["classes"].([]any); ok {
		for _, c := range classes {
			cm, ok := c.(map[string]any)
			if !ok {
				continue
			}
			if def, ok := cm["definition"].(map[string]any); ok {
				if n, ok := def["name"].(string); ok && className == "" {
					className = n
				}
			}
			if lv, ok := cm["level"].(float64); ok {
				level += int(lv)
			}
		}
	}
	if level == 0 {
		level = 1
	}
	// HP
	hpMax := 10
	if v, ok := obj["overrideHitPoints"].(float64); ok && v > 0 {
		hpMax = int(v)
	} else {
		base, _ := obj["baseHitPoints"].(float64)
		bonus, _ := obj["bonusHitPoints"].(float64)
		if base+bonus > 0 {
			hpMax = int(base + bonus)
		} else if base > 0 {
			hpMax = int(base)
		}
	}
	removed, _ := obj["removedHitPoints"].(float64)
	hpCurrent := hpMax - int(removed)
	if hpCurrent < 0 {
		hpCurrent = 0
	}
	// AC
	ac := 10
	if v, ok := obj["armorClass"].(float64); ok && v != 0 {
		ac = int(v)
	} else if v, ok := obj["overrideArmorClass"].(float64); ok && v != 0 {
		ac = int(v)
	} else if v, ok := obj["overrideAC"].(float64); ok && v != 0 {
		ac = int(v)
	}
	// race
	race := ""
	if r, ok := obj["race"].(map[string]any); ok {
		if n, ok := r["fullName"].(string); ok && n != "" {
			race = n
		} else if n, ok := r["baseName"].(string); ok {
			race = n
		}
	}
	background := ""
	if bg, ok := obj["background"].(map[string]any); ok {
		if def, ok := bg["definition"].(map[string]any); ok {
			if n, ok := def["name"].(string); ok {
				background = n
			}
		}
	}
	alignment := ""
	if a, ok := obj["alignmentName"].(string); ok {
		alignment = a
	}
	// XP
	xp := 0
	if v, ok := obj["currentXp"].(float64); ok {
		xp = int(v)
	}
	// inventory
	var inv []models.InventoryItem
	if items, ok := obj["inventory"].([]any); ok {
		for _, it := range items {
			m, ok := it.(map[string]any)
			if !ok {
				continue
			}
			def, _ := m["definition"].(map[string]any)
			if def == nil {
				continue
			}
			filterType, _ := def["filterType"].(string)
			name2, _ := def["name"].(string)
			if name2 == "" {
				continue
			}
			category := "gear"
			if filterType == "Weapon" {
				category = "weapon"
			} else if filterType == "Armor" {
				category = "armor"
			}
			qty := 1
			if q, ok := m["quantity"].(float64); ok {
				qty = int(q)
			}
			damageDice := ""
			damageType := ""
			if dmg, ok := def["damage"].(map[string]any); ok {
				if ds, ok := dmg["diceString"].(string); ok {
					damageDice = ds
				}
				if dt, ok := dmg["damageType"].(string); ok {
					damageType = dt
				}
			}
			props := ""
			if parr, ok := def["properties"].([]any); ok {
				var parts []string
				for _, p := range parr {
					if pm, ok := p.(map[string]any); ok {
						if n, ok := pm["name"].(string); ok {
							parts = append(parts, n)
						}
					}
				}
				props = strings.Join(parts, ", ")
			}
			acBonus := 0
			if filterType == "Armor" {
				if v, ok := def["armorClass"].(float64); ok {
					acBonus = int(v)
				}
			}
			equipped, _ := m["equipped"].(bool)
			magic := false
			if v, ok := def["magic"].(bool); ok {
				magic = v
			}
			inv = append(inv, models.InventoryItem{
				Name:             name2,
				Quantity:         qty,
				Category:         category,
				DamageDice:       damageDice,
				DamageType:       damageType,
				WeaponProperties: props,
				ACBonus:          acBonus,
				IsEquipped:       equipped,
				IsMagical:        magic,
			})
		}
	}

	ic := models.ImportCharacter{
		Name:       name,
		Race:       race,
		Class:      className,
		Level:      level,
		XP:         xp,
		Background: background,
		Alignment:  alignment,
		Str:        str,
		Dex:        dex,
		Con:        con,
		Int:        iv,
		Wis:        wis,
		Cha:        cha,
		HPMax:      hpMax,
		HPCurrent:  hpCurrent,
		AC:         ac,
		Inventory:  inv,
	}
	return []models.ImportCharacter{ic}, nil
}

func (foundryActorSource) Parse(raw []byte) ([]models.ImportCharacter, error) {
	var top map[string]any
	if err := json.Unmarshal(raw, &top); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	var obj map[string]any
	if actor, ok := top["actor"].(map[string]any); ok {
		obj = actor
	} else {
		obj = top
	}
	name, _ := obj["name"].(string)
	system, _ := obj["system"].(map[string]any)
	abilities := map[string]int{"str": 10, "dex": 10, "con": 10, "int": 10, "wis": 10, "cha": 10}
	if system != nil {
		if ab, ok := system["abilities"].(map[string]any); ok {
			for k := range abilities {
				if v, ok := ab[k].(map[string]any); ok {
					if val, ok := v["value"].(float64); ok {
						abilities[k] = int(val)
					}
				}
			}
		}
	}
	hpMax, hpCur := 10, 10
	ac := 10
	speed := 30
	if system != nil {
		if attrs, ok := system["attributes"].(map[string]any); ok {
			if hp, ok := attrs["hp"].(map[string]any); ok {
				if v, ok := hp["max"].(float64); ok {
					hpMax = int(v)
				}
				if v, ok := hp["value"].(float64); ok {
					hpCur = int(v)
				}
			}
			if acVal, ok := attrs["ac"]; ok {
				switch vv := acVal.(type) {
				case map[string]any:
					if v, ok := vv["value"].(float64); ok {
						ac = int(v)
					} else if v, ok := vv["flat"].(float64); ok {
						ac = int(v)
					}
				case float64:
					ac = int(vv)
				}
			}
			if sp, ok := attrs["speed"].(map[string]any); ok {
				if v, ok := sp["value"].(string); ok {
					fmt.Sscanf(v, "%d", &speed)
				} else if v, ok := sp["value"].(float64); ok {
					speed = int(v)
				}
			}
		}
		if det, ok := system["details"].(map[string]any); ok {
			// level handled below
			_ = det
		}
	}
	level := 1
	if system != nil {
		if det, ok := system["details"].(map[string]any); ok {
			if v, ok := det["level"].(float64); ok {
				level = int(v)
			}
		}
	}
	// items
	var inv []models.InventoryItem
	if items, ok := obj["items"].([]any); ok {
		for _, it := range items {
			m, ok := it.(map[string]any)
			if !ok {
				continue
			}
			typ, _ := m["type"].(string)
			if typ != "weapon" && typ != "equipment" {
				continue
			}
			n, _ := m["name"].(string)
			if n == "" {
				continue
			}
			category := "gear"
			if typ == "weapon" {
				category = "weapon"
			} else if typ == "equipment" {
				category = "armor"
			}
			damageDice, damageType := "", ""
			props := ""
			qty := 1
			if sys, ok := m["system"].(map[string]any); ok {
				if dmg, ok := sys["damage"].(map[string]any); ok {
					if parts, ok := dmg["parts"].([]any); ok && len(parts) > 0 {
						if first, ok := parts[0].([]any); ok && len(first) >= 2 {
							if s, ok := first[0].(string); ok {
								damageDice = s
							}
							if s, ok := first[1].(string); ok {
								damageType = s
							}
						}
					}
				}
				if pp, ok := sys["properties"].([]any); ok {
					var ps []string
					for _, p := range pp {
						if s, ok := p.(string); ok {
							ps = append(ps, s)
						}
					}
					props = strings.Join(ps, ", ")
				}
				if q, ok := sys["quantity"].(float64); ok {
					qty = int(q)
				}
			}
			inv = append(inv, models.InventoryItem{
				Name:             n,
				Category:         category,
				DamageDice:       damageDice,
				DamageType:       damageType,
				WeaponProperties: props,
				Quantity:         qty,
			})
		}
	}
	// currency
	currency := models.Currency{}
	if system != nil {
		if cur, ok := system["currency"].(map[string]any); ok {
			if v, ok := cur["pp"].(float64); ok {
				currency.PP = int(v)
			}
			if v, ok := cur["gp"].(float64); ok {
				currency.GP = int(v)
			}
			if v, ok := cur["ep"].(float64); ok {
				currency.EP = int(v)
			}
			if v, ok := cur["sp"].(float64); ok {
				currency.SP = int(v)
			}
			if v, ok := cur["cp"].(float64); ok {
				currency.CP = int(v)
			}
		}
	}
	ic := models.ImportCharacter{
		Name:      name,
		Str:       abilities["str"],
		Dex:       abilities["dex"],
		Con:       abilities["con"],
		Int:       abilities["int"],
		Wis:       abilities["wis"],
		Cha:       abilities["cha"],
		HPMax:     hpMax,
		HPCurrent: hpCur,
		AC:        ac,
		Level:     level,
		Speed:     speed,
		Inventory: inv,
		Currency:  currency,
	}
	return []models.ImportCharacter{ic}, nil
}

func (fiveEToolsSource) Parse(raw []byte) ([]map[string]any, error) {
	var arr []map[string]any
	if err := json.Unmarshal(raw, &arr); err == nil {
		return arr, nil
	}
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	if data, ok := obj["data"]; ok {
		if a, ok := data.([]any); ok {
			out := make([]map[string]any, 0, len(a))
			for _, e := range a {
				if m, ok := e.(map[string]any); ok {
					out = append(out, m)
				} else {
					return nil, fmt.Errorf("invalid entry type")
				}
			}
			return out, nil
		}
	}
	// single key whose value is array
	if len(obj) == 1 {
		for _, v := range obj {
			if a, ok := v.([]any); ok {
				out := make([]map[string]any, 0, len(a))
				for _, e := range a {
					if m, ok := e.(map[string]any); ok {
						out = append(out, m)
					} else {
						return nil, fmt.Errorf("invalid entry type")
					}
				}
				return out, nil
			}
		}
	}
	return []map[string]any{obj}, nil
}

func (foundryPackSource) Parse(raw []byte) ([]map[string]any, error) {
	if !json.Valid(raw) {
		return nil, errors.New("Foundry LevelDB .db packs are not supported; export a JSON source pack instead")
	}
	var arr []map[string]any
	if err := json.Unmarshal(raw, &arr); err == nil {
		return arr, nil
	}
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	if entries, ok := obj["entries"]; ok {
		if a, ok := entries.([]any); ok {
			out := make([]map[string]any, 0, len(a))
			for _, e := range a {
				if m, ok := e.(map[string]any); ok {
					out = append(out, m)
				}
			}
			return out, nil
		}
	}
	if items, ok := obj["items"]; ok {
		if a, ok := items.([]any); ok {
			out := make([]map[string]any, 0, len(a))
			for _, e := range a {
				if m, ok := e.(map[string]any); ok {
					out = append(out, m)
				}
			}
			return out, nil
		}
	}
	return nil, fmt.Errorf("invalid JSON: expected array or entries/items object")
}

func suggestMapping(entries []map[string]any, fields []models.SchemaField) []FieldMapping {
	if len(entries) == 0 {
		return nil
	}
	first := entries[0]
	lowerKeys := map[string]string{}
	for k := range first {
		lowerKeys[strings.ToLower(k)] = k
	}
	var out []FieldMapping
	for _, f := range fields {
		if src, ok := lowerKeys[strings.ToLower(f.Name)]; ok {
			out = append(out, FieldMapping{Source: src, Target: f.Name})
		}
	}
	return out
}

type externalImportRequest struct {
	Source       string          `json:"source"`
	Kind         string          `json:"kind"`
	Payload      json.RawMessage `json:"payload"`
	URL          string          `json:"url"`
	DryRun       bool            `json:"dry_run"`
	DedupAction  string          `json:"dedup_action"`
	FieldMapping []FieldMapping  `json:"field_mapping"`
	SchemaID     int64           `json:"schema_id"`
	NameField    string          `json:"name_field"`
}

func fetchExternalURL(rawURL string) ([]byte, error) {
	u, err := url.Parse(rawURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" {
		return nil, errors.New("invalid url")
	}
	ips, err := net.LookupIP(u.Hostname())
	if err != nil || len(ips) == 0 {
		return nil, errors.New("could not resolve host")
	}
	for _, ip := range ips {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
			return nil, errors.New("url host must be publicly reachable")
		}
	}
	client := newAIClient(aiFetchTimeout)
	resp, err := client.Get(rawURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch failed: %s", resp.Status)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if err != nil {
		return nil, err
	}
	return data, nil
}

func HandleExternalImport(c *gin.Context) {
	var req externalImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var raw []byte
	if len(req.Payload) > 0 {
		raw = req.Payload
	} else if req.URL != "" {
		data, err := fetchExternalURL(req.URL)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		raw = data
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payload or url required"})
		return
	}
	userID, _ := c.Get("user_id")
	uid, _ := userID.(int64)

	if req.Kind == "character" {
		src, ok := characterSources[req.Source]
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "unknown source"})
			return
		}
		chars, err := src.Parse(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if req.DryRun {
			var rows []map[string]any
			for i, ch := range chars {
				status := "create"
				if ch.Name == "" {
					status = "error"
				}
				rows = append(rows, map[string]any{"index": i, "name": ch.Name, "status": status})
			}
			if rows == nil {
				rows = []map[string]any{}
			}
			c.JSON(http.StatusOK, gin.H{"source": req.Source, "kind": req.Kind, "dry_run": true, "created": 0, "skipped": 0, "duplicates": 0, "rows": rows})
			return
		}
		results := importCharacters(c.Request.Context(), uid, chars)
		var createdIDs []int64
		var rows []map[string]any
		created := 0
		skipped := 0
		for i, r := range results {
			if errMsg, ok := r["error"]; ok {
				rows = append(rows, map[string]any{"index": i, "name": chars[i].Name, "status": "error", "error": errMsg})
				skipped++
				continue
			}
			if id, ok := r["id"]; ok {
				var iid int64
				switch v := id.(type) {
				case int64:
					iid = v
				case int:
					iid = int64(v)
				case float64:
					iid = int64(v)
				}
				if iid != 0 {
					createdIDs = append(createdIDs, iid)
				}
				created++
				rows = append(rows, map[string]any{"index": i, "name": r["name"], "status": "create", "id": iid})
			}
		}
		if createdIDs == nil {
			createdIDs = []int64{}
		}
		if rows == nil {
			rows = []map[string]any{}
		}
		idsJSON, _ := json.Marshal(createdIDs)
		countsJSON, _ := json.Marshal(map[string]int{"created": created, "skipped": skipped})
		res, err := db.DB.Exec(`INSERT INTO external_import_logs(user_id, source, kind, status, created_ids, counts, filename) VALUES(?,?,?,?,?,?,?)`, uid, req.Source, "character", "completed", string(idsJSON), string(countsJSON), req.Source)
		var logID int64
		if err == nil {
			logID, _ = res.LastInsertId()
		}
		resp := gin.H{"source": req.Source, "kind": req.Kind, "dry_run": false, "created": created, "skipped": skipped, "duplicates": 0, "rows": rows}
		if logID != 0 {
			resp["log_id"] = logID
		}
		c.JSON(http.StatusOK, resp)
		return
	} else if req.Kind == "compendium" {
		role, _ := c.Get("role")
		if role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin required"})
			return
		}
		if req.SchemaID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "schema_id is required"})
			return
		}
		src, ok := compendiumSources[req.Source]
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "unknown source"})
			return
		}
		var fieldsJSON, displayName string
		var typeName string
		err := db.DB.QueryRow("SELECT type_name, display_name, fields FROM compendium_schemas WHERE id=?", req.SchemaID).Scan(&typeName, &displayName, &fieldsJSON)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "schema not found"})
			return
		}
		var schemaFields []models.SchemaField
		_ = json.Unmarshal([]byte(fieldsJSON), &schemaFields)
		entries, err := src.Parse(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		mapping := req.FieldMapping
		if len(mapping) == 0 {
			mapping = suggestMapping(entries, schemaFields)
		}
		dedup := req.DedupAction
		if dedup == "" {
			dedup = "skip"
		}
		opts := ImportOpts{
			SchemaID:          req.SchemaID,
			SchemaFields:      schemaFields,
			SchemaDisplayName: displayName,
			RawEntries:        entries,
			Mapping:           mapping,
			DuplicateAction:   dedup,
			DryRun:            req.DryRun,
			UseTx:             true,
			UserID:            uid,
			Filename:          req.Source,
			NameField:         req.NameField,
		}
		result, err := importCompendiumEntries(c.Request.Context(), db.DB, opts)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		// build rows per input index
		// map field errors and duplicates by index
		errByIdx := map[int]string{}
		for _, fe := range result.FieldErrors {
			errByIdx[fe.Index] = fe.Message
		}
		dupByIdx := map[int]models.CompendiumImportDuplicate{}
		for _, d := range result.Duplicates {
			dupByIdx[d.Index] = d
		}
		var rows []map[string]any
		for i := range entries {
			if msg, ok := errByIdx[i]; ok {
				rows = append(rows, map[string]any{"index": i, "status": "error", "error": msg})
				continue
			}
			if dup, ok := dupByIdx[i]; ok {
				rows = append(rows, map[string]any{"index": i, "status": dup.Resolved})
				continue
			}
			rows = append(rows, map[string]any{"index": i, "status": "create"})
		}
		if rows == nil {
			rows = []map[string]any{}
		}
		created := result.Inserted
		skipped := result.Skipped + result.Overwritten
		duplicates := len(result.Duplicates)
		if req.DryRun {
			c.JSON(http.StatusOK, gin.H{"source": req.Source, "kind": req.Kind, "dry_run": true, "created": 0, "skipped": skipped, "duplicates": duplicates, "rows": rows})
			return
		}
		createdIDs := result.CreatedIDs
		if createdIDs == nil {
			createdIDs = []int64{}
		}
		idsJSON, _ := json.Marshal(createdIDs)
		countsJSON, _ := json.Marshal(map[string]int{"created": created, "skipped": skipped})
		res, err := db.DB.Exec(`INSERT INTO external_import_logs(user_id, source, kind, status, created_ids, counts, filename) VALUES(?,?,?,?,?,?,?)`, uid, req.Source, "compendium", "completed", string(idsJSON), string(countsJSON), req.Source)
		var logID int64
		if err == nil {
			logID, _ = res.LastInsertId()
		}
		resp := gin.H{"source": req.Source, "kind": req.Kind, "dry_run": false, "created": created, "skipped": skipped, "duplicates": duplicates, "rows": rows}
		if logID != 0 {
			resp["log_id"] = logID
		}
		c.JSON(http.StatusOK, resp)
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": "invalid kind"})
}

func HandleExternalImportLogs(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid, _ := userID.(int64)
	role, _ := c.Get("role")
	var rows *http.Response
	_ = rows
	var dbRows interface {
		Next() bool
		Scan(...any) error
		Close() error
	}
	_ = dbRows
	var q string
	var args []any
	if role == "admin" {
		q = `SELECT id, user_id, source, kind, status, created_ids, counts, filename, created_at, rolled_back_at FROM external_import_logs ORDER BY created_at DESC LIMIT 50`
	} else {
		q = `SELECT id, user_id, source, kind, status, created_ids, counts, filename, created_at, rolled_back_at FROM external_import_logs WHERE user_id=? ORDER BY created_at DESC LIMIT 50`
		args = append(args, uid)
	}
	sqlRows, err := db.DB.Query(q, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer sqlRows.Close()
	type logOut struct {
		ID           int64          `json:"id"`
		UserID       int64          `json:"user_id"`
		Source       string         `json:"source"`
		Kind         string         `json:"kind"`
		Status       string         `json:"status"`
		CreatedIDs   []int64        `json:"created_ids"`
		Counts       map[string]any `json:"counts"`
		Filename     string         `json:"filename"`
		CreatedAt    string         `json:"created_at"`
		RolledBackAt *string        `json:"rolled_back_at"`
	}
	var out []logOut
	for sqlRows.Next() {
		var l logOut
		var idsJSON, countsJSON string
		var rolled *string
		if err := sqlRows.Scan(&l.ID, &l.UserID, &l.Source, &l.Kind, &l.Status, &idsJSON, &countsJSON, &l.Filename, &l.CreatedAt, &rolled); err != nil {
			continue
		}
		_ = json.Unmarshal([]byte(idsJSON), &l.CreatedIDs)
		if l.CreatedIDs == nil {
			l.CreatedIDs = []int64{}
		}
		_ = json.Unmarshal([]byte(countsJSON), &l.Counts)
		if l.Counts == nil {
			l.Counts = map[string]any{}
		}
		l.RolledBackAt = rolled
		out = append(out, l)
	}
	if out == nil {
		out = []logOut{}
	}
	c.JSON(http.StatusOK, out)
}

func HandleExternalImportRollback(c *gin.Context) {
	idStr := c.Param("id")
	logID, err := parseID(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid log id"})
		return
	}
	var status, kind, idsJSON string
	err = db.DB.QueryRow(`SELECT status, kind, created_ids FROM external_import_logs WHERE id=?`, logID).Scan(&status, &kind, &idsJSON)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "log not found"})
		return
	}
	if status == "rolled_back" {
		c.JSON(http.StatusOK, gin.H{"rolled_back": 0})
		return
	}
	var ids []int64
	_ = json.Unmarshal([]byte(idsJSON), &ids)
	if len(ids) > 0 {
		placeholders := make([]string, len(ids))
		args := make([]any, len(ids))
		for i, v := range ids {
			placeholders[i] = "?"
			args[i] = v
		}
		if kind == "character" {
			// delete dependents first to avoid FK constraint failure
			_, _ = db.DB.Exec("DELETE FROM inventory_items WHERE character_id IN ("+strings.Join(placeholders, ",")+")", args...)
			_, _ = db.DB.Exec("DELETE FROM character_currency WHERE character_id IN ("+strings.Join(placeholders, ",")+")", args...)
			_, _ = db.DB.Exec("DELETE FROM character_proficiencies WHERE character_id IN ("+strings.Join(placeholders, ",")+")", args...)
			_, _ = db.DB.Exec("DELETE FROM character_features WHERE character_id IN ("+strings.Join(placeholders, ",")+")", args...)
			_, _ = db.DB.Exec("DELETE FROM character_spellcasting WHERE character_id IN ("+strings.Join(placeholders, ",")+")", args...)
			_, _ = db.DB.Exec("DELETE FROM spells WHERE character_id IN ("+strings.Join(placeholders, ",")+")", args...)
			_, _ = db.DB.Exec("DELETE FROM characters WHERE id IN ("+strings.Join(placeholders, ",")+")", args...)
		} else if kind == "compendium" {
			_, _ = db.DB.Exec("DELETE FROM compendium_entries WHERE id IN ("+strings.Join(placeholders, ",")+")", args...)
		}
	}
	_, _ = db.DB.Exec(`UPDATE external_import_logs SET status='rolled_back', rolled_back_at=datetime('now') WHERE id=?`, logID)
	c.JSON(http.StatusOK, gin.H{"rolled_back": len(ids)})
}

func parseID(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}
