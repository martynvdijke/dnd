package handlers

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
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
