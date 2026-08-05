package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/registry"
)

// ─── Transfer Envelope Types ───

// TransferEnvelope is the top-level villum-transfer JSON format.
type TransferEnvelope struct {
	VillumTransfer TransferMeta     `json:"villum_transfer"`
	Entities       []TransferEntity `json:"entities"`
}

// TransferMeta describes the transfer file itself.
type TransferMeta struct {
	Version    int    `json:"version"`
	ExportedAt string `json:"exported_at"`
	Source     string `json:"source"`
	SourceURL  string `json:"source_url,omitempty"`
}

// TransferEntity is one exported entity row.
type TransferEntity struct {
	Type       string         `json:"type"`
	OriginalID int64          `json:"original_id"`
	Data       map[string]any `json:"data"`
}

// TransferImportResult is returned after an import attempt.
type TransferImportResult struct {
	Status   string                 `json:"status"` // "completed", "partial", "failed"
	ImportID int64                  `json:"import_id,omitempty"`
	Results  []TransferEntityResult `json:"results"`
	Error    string                 `json:"error,omitempty"`
	DryRun   bool                   `json:"dry_run"`
}

// TransferEntityResult records what happened to one entity during import.
type TransferEntityResult struct {
	Type       string `json:"type"`
	OriginalID int64  `json:"original_id"`
	NewID      int64  `json:"new_id,omitempty"`
	Status     string `json:"status"` // "imported", "skipped", "error"
	Error      string `json:"error,omitempty"`
}

// ─── Column Schema Cache ───

type columnInfo struct {
	Index   int    // ordinal position (0-based)
	Name    string // column name
	Type    string // declared type (e.g. "INTEGER", "TEXT")
	NotNull bool
	PK      bool // is primary key
}

type tableSchemaCache struct {
	mu   sync.RWMutex
	data map[string][]columnInfo
}

var schemaCache = &tableSchemaCache{data: make(map[string][]columnInfo)}

func loadTableSchema(tx *sql.Tx, table string) ([]columnInfo, error) {
	schemaCache.mu.RLock()
	cols, ok := schemaCache.data[table]
	schemaCache.mu.RUnlock()
	if ok {
		return cols, nil
	}
	rows, err := tx.Query(fmt.Sprintf("PRAGMA table_info(%q)", table))
	if err != nil {
		return nil, fmt.Errorf("pragma table_info %s: %w", table, err)
	}
	defer rows.Close()
	for rows.Next() {
		var ci columnInfo
		var cid int
		var nullable int
		var dflt, colType *string
		if err := rows.Scan(&cid, &ci.Name, &colType, &nullable, &dflt, &ci.PK); err != nil {
			return nil, fmt.Errorf("scan column info: %w", err)
		}
		ci.Index = cid
		ci.NotNull = nullable != 0
		if colType != nil {
			ci.Type = *colType
		}
		cols = append(cols, ci)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(cols) == 0 {
		return nil, fmt.Errorf("table %q has no columns or does not exist", table)
	}
	schemaCache.mu.Lock()
	schemaCache.data[table] = cols
	schemaCache.mu.Unlock()
	return cols, nil
}

// ─── Export Helpers ───

// queryEntityData fetches all columns from the given table for the given IDs.
func queryEntityData(tx *sql.Tx, info registry.EntityInfo, ids []int64) ([]map[string]any, error) {
	cols, err := loadTableSchema(tx, info.Table)
	if err != nil {
		return nil, err
	}
	colNames := make([]string, 0, len(cols))
	colPtrs := make([]func(any) (string, any), 0, len(cols))
	for _, c := range cols {
		colNames = append(colNames, c.Name)
		colPtrs = append(colPtrs, scannerForCol(c))
	}
	// Build SELECT col1,col2,... FROM table WHERE id IN (?,?,...)
	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}
	query := fmt.Sprintf("SELECT %s FROM %q WHERE id IN (%s)",
		strings.Join(colNames, ","), info.Table, strings.Join(placeholders, ","))
	rows, err := tx.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query %s: %w", info.Table, err)
	}
	defer rows.Close()

	var results []map[string]any
	for rows.Next() {
		scanTargets := make([]any, len(colPtrs))
		colValMap := make([]colVal, len(colPtrs))
		for i, fn := range colPtrs {
			name, ptr := fn(nil)
			colValMap[i] = colVal{name: name}
			scanTargets[i] = ptr
		}
		if err := rows.Scan(scanTargets...); err != nil {
			return nil, fmt.Errorf("scan %s: %w", info.Table, err)
		}
		row := make(map[string]any, len(cols))
		for i := range colPtrs {
			row[colValMap[i].name] = deref(scanTargets[i])
		}
		results = append(results, row)
	}
	return results, rows.Err()
}

type colVal struct {
	name string
	val  any
}

// scannerForCol returns a closure that allocates a scan target for the column type.
func scannerForCol(c columnInfo) func(initial any) (string, any) {
	return func(initial any) (string, any) {
		switch strings.ToUpper(c.Type) {
		case "INTEGER", "INT", "BIGINT", "SMALLINT", "TINYINT", "BOOL", "BOOLEAN":
			if initial != nil {
				return c.Name, initial
			}
			var v sql.NullInt64
			return c.Name, &v
		case "REAL", "FLOAT", "DOUBLE", "NUMERIC", "DECIMAL":
			if initial != nil {
				return c.Name, initial
			}
			var v sql.NullFloat64
			return c.Name, &v
		default:
			if initial != nil {
				return c.Name, initial
			}
			var v sql.NullString
			return c.Name, &v
		}
	}
}

func deref(ptr any) any {
	switch v := ptr.(type) {
	case *sql.NullInt64:
		if v.Valid {
			return v.Int64
		}
		return nil
	case *sql.NullFloat64:
		if v.Valid {
			return v.Float64
		}
		return nil
	case *sql.NullString:
		if v.Valid {
			return v.String
		}
		return nil
	case *sql.NullBool:
		if v.Valid {
			return v.Bool
		}
		return nil
	case *sql.NullTime:
		if v.Valid {
			return v.Time
		}
		return nil
	case *sql.NullInt32:
		if v.Valid {
			return v.Int32
		}
		return nil
	case *int64:
		return *v
	case *float64:
		return *v
	case *string:
		return *v
	case *bool:
		return *v
	default:
		return v
	}
}

// pkColumn returns the name of the primary key column for a table.
func pkColumn(cols []columnInfo) string {
	for _, c := range cols {
		if c.PK {
			return c.Name
		}
	}
	return "id" // fallback
}

// ─── Export ───

// exportRequest captures export parameters.
type exportRequest struct {
	Types      []string `json:"types"`
	CampaignID *int64   `json:"campaign_id,omitempty"`
}

// doExport is the shared export implementation used by both POST and campaign handlers.
func doExport(c *gin.Context, req exportRequest) {
	userID := c.GetInt64("user_id")
	admin := c.GetString("role") == "admin"

	types := req.Types
	if len(types) == 0 {
		for _, info := range registry.Entities {
			if info.Transferable {
				types = append(types, info.Type)
			}
		}
		sort.Strings(types)
	}

	tx, err := db.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback()

	env := TransferEnvelope{
		VillumTransfer: TransferMeta{
			Version:    1,
			ExportedAt: time.Now().UTC().Format(time.RFC3339),
			Source:     "villum",
		},
	}

	for _, t := range types {
		info, ok := registry.Entities[t]
		if !ok || !info.Transferable {
			continue
		}
		ids, err := collectVisibleIDs(tx, info, userID, admin)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("%s: %v", t, err)})
			return
		}
		if req.CampaignID != nil {
			ids = filterByCampaign(tx, info, ids, *req.CampaignID)
		}
		if len(ids) == 0 {
			continue
		}
		rows, err := queryEntityData(tx, info, ids)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("%s: %v", t, err)})
			return
		}
		for _, row := range rows {
			originalID, _ := row["id"].(int64)
			env.Entities = append(env.Entities, TransferEntity{
				Type:       t,
				OriginalID: originalID,
				Data:       row,
			})
		}
	}
	tx.Commit()

	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="villum-export-%s.json"`, time.Now().Format("20060102-150405")))
	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, env)
}

// HandleExportTransfer exports entity types, optionally scoped by campaign.
// POST /api/transfer/export
func HandleExportTransfer(c *gin.Context) {
	var req exportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	doExport(c, req)
}

// HandleExportTransferCampaign exports all entities associated with a campaign.
// GET /api/transfer/export/campaign/:id
func HandleExportTransferCampaign(c *gin.Context) {
	campaignID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid campaign id"})
		return
	}
	doExport(c, exportRequest{CampaignID: &campaignID})
}

// collectVisibleIDs returns all IDs the user can see for a given entity type.
func collectVisibleIDs(q registry.Queryer, info registry.EntityInfo, userID int64, admin bool) ([]int64, error) {
	if admin {
		rows, err := q.Query(fmt.Sprintf("SELECT id FROM %q", info.Table))
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		var ids []int64
		for rows.Next() {
			var id int64
			if err := rows.Scan(&id); err != nil {
				return nil, err
			}
			ids = append(ids, id)
		}
		return ids, rows.Err()
	}
	var query string
	var args []any
	if info.Type == "campaign" {
		// campaigns table uses user_id or campaign_members, not campaign_id.
		query = `SELECT id FROM campaigns WHERE user_id = ? UNION SELECT campaign_id FROM campaign_members WHERE user_id = ? AND role = 'dm'`
		args = []any{userID, userID}
	} else {
		where, n := registry.VisibilityWhere(info.Ownership)
		args = make([]any, n)
		for i := range args {
			args[i] = userID
		}
		query = fmt.Sprintf("SELECT id FROM %q WHERE %s", info.Table, where)
	}
	rows, err := q.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// filterByCampaign filters entity IDs to those associated with the given campaign.
func filterByCampaign(tx *sql.Tx, info registry.EntityInfo, ids []int64, campaignID int64) []int64 {
	if len(ids) == 0 {
		return nil
	}
	// Determine which column to check for campaign association.
	campaignCol := campaignColumn(info.Type)
	if campaignCol == "" {
		// Entity type doesn't have campaign association; include all.
		return ids
	}
	placeholders := make([]string, len(ids))
	args := make([]any, 0, len(ids)+1)
	args = append(args, campaignID)
	for i, id := range ids {
		placeholders[i] = "?"
		args = append(args, id)
	}
	query := fmt.Sprintf("SELECT id FROM %q WHERE %s = ? AND id IN (%s)",
		info.Table, campaignCol, strings.Join(placeholders, ","))
	rows, err := tx.Query(query, args...)
	if err != nil {
		return ids // fail open on error
	}
	defer rows.Close()
	var filtered []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			continue
		}
		filtered = append(filtered, id)
	}
	return filtered
}

// campaignColumn returns the SQL column name used to associate entities with a campaign.
func campaignColumn(entityType string) string {
	switch entityType {
	case "campaign":
		return "id"
	case "character", "encounter", "shop", "faction", "adventure":
		return "campaign_id"
	case "timeline":
		return "campaign_id"
	case "location":
		return "" // locations are user-level, not campaign-scoped
	default:
		return ""
	}
}

// ─── Import ───

// importOrder defines the dependency order: entities are inserted in this sequence.
var importOrder = []string{
	"campaign",
	"location",
	"npc",
	"monster",
	"adventure",
	"shop",
	"faction",
	"encounter",
	"character",
	"quest",
	"journal",
	"timeline",
}

// fkColumns maps entity type -> list of FK column names that need ID remapping.
var fkColumns = map[string][]struct {
	Col    string // column name in this entity's table
	Ref    string // referenced entity type
	Prefix string // if set, remap only if value starts with this (for string-encoded refs)
}{
	"character": {{Col: "campaign_id", Ref: "campaign"}},
	"encounter": {{Col: "campaign_id", Ref: "campaign"}},
	"shop":      {{Col: "campaign_id", Ref: "campaign"}, {Col: "oneshot_adventure_id", Ref: "adventure"}},
	"faction":   {{Col: "campaign_id", Ref: "campaign"}},
	"adventure": {{Col: "campaign_id", Ref: "campaign"}},
	"quest":     {{Col: "character_id", Ref: "character"}},
	"journal":   {{Col: "character_id", Ref: "character"}},
	"timeline":  {{Col: "campaign_id", Ref: "campaign"}},
}

// HandleImportTransfer imports a transfer envelope file.
// POST /api/transfer/import?dry_run=true
func HandleImportTransfer(c *gin.Context) {
	userID := c.GetInt64("user_id")
	dryRun := c.DefaultQuery("dry_run", "false") == "true"

	var env TransferEnvelope
	if err := c.ShouldBindJSON(&env); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transfer file: " + err.Error()})
		return
	}

	if len(env.Entities) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no entities to import"})
		return
	}

	// Group entities by type.
	grouped := make(map[string][]TransferEntity)
	for _, e := range env.Entities {
		if _, ok := registry.Entities[e.Type]; !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unknown entity type: %s", e.Type)})
			return
		}
		if !registry.Entities[e.Type].Transferable {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("entity type %s is not transferable", e.Type)})
			return
		}
		grouped[e.Type] = append(grouped[e.Type], e)
	}

	// Pre-validate all data is non-nil.
	for t, ents := range grouped {
		for _, e := range ents {
			if e.Data == nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%s id=%d: data is null", t, e.OriginalID)})
				return
			}
		}
	}

	if dryRun {
		c.JSON(http.StatusOK, TransferImportResult{
			Status:  "ok",
			DryRun:  true,
			Results: []TransferEntityResult{},
		})
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback()

	// idMapping: entity type -> old ID -> new ID
	idMapping := make(map[string]map[int64]int64)
	var results []TransferEntityResult
	hasError := false

	for _, t := range importOrder {
		ents, ok := grouped[t]
		if !ok || len(ents) == 0 {
			continue
		}
		info := registry.Entities[t]
		cols, err := loadTableSchema(tx, info.Table)
		if err != nil {
			hasError = true
			results = append(results, TransferEntityResult{
				Type: t, Status: "error", Error: err.Error(),
			})
			continue
		}
		targetCols := make([]columnInfo, 0, len(cols))
		for _, c := range cols {
			if c.PK || c.Name == "created_at" || c.Name == "updated_at" {
				continue // skip PK and timestamp columns (have DEFAULT values)
			}
			targetCols = append(targetCols, c)
		}
		if len(targetCols) == 0 {
			continue
		}
		colNames := make([]string, len(targetCols))
		for i, c := range targetCols {
			colNames[i] = c.Name
		}
		placeholders := make([]string, len(targetCols))
		for i := range placeholders {
			placeholders[i] = "?"
		}
		insertSQL := fmt.Sprintf("INSERT INTO %q (%s) VALUES (%s)",
			info.Table, strings.Join(colNames, ","), strings.Join(placeholders, ","))

		for _, ent := range ents {
			data := copyData(ent.Data)

			// Remap FK references.
			for _, fk := range fkColumns[t] {
				remapFK(data, fk.Col, fk.Ref, idMapping)
			}

			// If the entity has a user_id column, set to importing user.
			if hasColumn(targetCols, "user_id") {
				data["user_id"] = userID
			}

			// Build value slice matching targetCols order.
			vals := make([]any, len(targetCols))
			valid := true
			for i, c := range targetCols {
				val, exists := data[c.Name]
				if !exists || val == nil {
					if c.NotNull {
						// Use zero value for missing non-null fields.
						vals[i] = zeroValue(c.Type)
					} else {
						vals[i] = nil
					}
				} else {
					vals[i] = val
				}
			}
			if !valid {
				continue
			}

			newID, err := execInsert(tx, insertSQL, vals)
			if err != nil {
				hasError = true
				results = append(results, TransferEntityResult{
					Type: t, OriginalID: ent.OriginalID, Status: "error",
					Error: err.Error(),
				})
				continue
			}
			if idMapping[t] == nil {
				idMapping[t] = make(map[int64]int64)
			}
			idMapping[t][ent.OriginalID] = newID
			results = append(results, TransferEntityResult{
				Type: t, OriginalID: ent.OriginalID, NewID: newID, Status: "imported",
			})
		}
	}

	importedCount := 0
	for _, r := range results {
		if r.Status == "imported" {
			importedCount++
		}
	}
	logStatus := "completed"
	if hasError {
		logStatus = "failed"
	}
	logCounts := map[string]int{"imported": importedCount, "total": len(env.Entities)}
	logCountsJSON, _ := json.Marshal(logCounts)
	logResult, err := tx.Exec(`INSERT INTO transfer_import_logs(user_id, file_name, doc_type, counts, status)
		VALUES(?, ?, 'villum-transfer', ?, ?)`, userID, "imported", string(logCountsJSON), logStatus)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "log: " + err.Error()})
		return
	}
	logID, _ := logResult.LastInsertId()

	if hasError {
		tx.Commit() // commit what succeeded, mark as failed
		c.JSON(http.StatusOK, TransferImportResult{
			Status:   "failed",
			ImportID: logID,
			Results:  results,
		})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "commit: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, TransferImportResult{
		Status:   "completed",
		ImportID: logID,
		Results:  results,
	})
}

// ─── Audit Log ───

// HandleListTransferLogs lists the import logs for the current user (admin sees all).
// GET /api/transfer/logs
func HandleListTransferLogs(c *gin.Context) {
	userID := c.GetInt64("user_id")
	admin := c.GetString("role") == "admin"

	var rows *sql.Rows
	var err error
	if admin {
		rows, err = db.DB.Query(`SELECT id, user_id, file_name, doc_type, counts, status, error, created_at
			FROM transfer_import_logs ORDER BY created_at DESC LIMIT 100`)
	} else {
		rows, err = db.DB.Query(`SELECT id, user_id, file_name, doc_type, counts, status, error, created_at
			FROM transfer_import_logs WHERE user_id=? ORDER BY created_at DESC LIMIT 50`, userID)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	type ImportLogEntry struct {
		ID        int64           `json:"id"`
		UserID    int64           `json:"user_id"`
		FileName  string          `json:"file_name"`
		DocType   string          `json:"doc_type"`
		Counts    json.RawMessage `json:"counts"`
		Status    string          `json:"status"`
		Error     string          `json:"error"`
		CreatedAt string          `json:"created_at"`
	}
	logs := make([]ImportLogEntry, 0)
	for rows.Next() {
		var l ImportLogEntry
		var countsStr string
		if err := rows.Scan(&l.ID, &l.UserID, &l.FileName, &l.DocType, &countsStr, &l.Status, &l.Error, &l.CreatedAt); err != nil {
			continue
		}
		l.Counts = json.RawMessage(countsStr)
		logs = append(logs, l)
	}
	c.JSON(http.StatusOK, logs)
}

// ─── Route Registration ───

// RegisterTransferRoutes registers transfer export/import endpoints.
func RegisterTransferRoutes(auth *gin.RouterGroup) {
	auth.POST("/transfer/export", HandleExportTransfer)
	auth.GET("/transfer/export/campaign/:id", HandleExportTransferCampaign)
	auth.POST("/transfer/import", HandleImportTransfer)
	auth.GET("/transfer/logs", HandleListTransferLogs)
}

// ─── Internal Helpers ───

// copyData makes a shallow copy of a map.
func copyData(m map[string]any) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

// hasColumn checks if a column name exists in the slice.
func hasColumn(cols []columnInfo, name string) bool {
	for _, c := range cols {
		if c.Name == name {
			return true
		}
	}
	return false
}

// zeroValue returns a Go zero value for the given SQL type name.
func zeroValue(sqlType string) any {
	switch strings.ToUpper(sqlType) {
	case "INTEGER", "INT", "BIGINT", "SMALLINT", "TINYINT", "BOOL", "BOOLEAN":
		return int64(0)
	case "REAL", "FLOAT", "DOUBLE":
		return 0.0
	default:
		return ""
	}
}

// execInsert runs an INSERT and returns the new row ID.
func execInsert(tx *sql.Tx, sql string, args []any) (int64, error) {
	res, err := tx.Exec(sql, args...)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

// remapFK replaces a foreign key value in data from old ID to new ID using idMapping.
func remapFK(data map[string]any, col, refType string, idMapping map[string]map[int64]int64) {
	val, ok := data[col]
	if !ok || val == nil {
		return
	}
	oldID, ok := toInt64(val)
	if !ok || oldID == 0 {
		return
	}
	mapped, ok := idMapping[refType]
	if !ok {
		return
	}
	newID, ok := mapped[oldID]
	if !ok {
		// Referenced entity wasn't imported; leave as-is.
		return
	}
	data[col] = newID
}

// toInt64 attempts to convert a value to int64.
func toInt64(v any) (int64, bool) {
	switch n := v.(type) {
	case int64:
		return n, true
	case float64:
		return int64(n), true
	case int:
		return int64(n), true
	case json.Number:
		i, err := n.Int64()
		return i, err == nil
	default:
		s := fmt.Sprintf("%v", v)
		i, err := strconv.ParseInt(s, 10, 64)
		return i, err == nil
	}
}
