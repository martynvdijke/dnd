package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
	"villum/db"
	"villum/registry"
)

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
