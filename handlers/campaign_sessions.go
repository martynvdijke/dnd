package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/ent"
	"villum/ent/journalentry"
	"villum/ent/quest"
	"villum/ent/session"
	"villum/models"
)

func ListSessions(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	sessions, err := db.Client.Session.Query().Where(session.CharacterID(charID)).Order(ent.Desc(session.FieldSessionDate)).All(c.Request.Context())
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	var out = make([]models.Session, 0)
	for _, s := range sessions {
		out = append(out, models.Session{ID: s.ID, CharacterID: s.CharacterID, SessionDate: s.SessionDate, Title: s.Title, Notes: s.Notes, XPEarned: s.XpEarned, GoldEarned: s.GoldEarned, ImportantEvents: s.ImportantEvents, CreatedAt: s.CreatedAt})
	}
	WriteJSON(c, http.StatusOK, out)
}

func CreateSession(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if !canEditCharacterID(c, charID) {
		WriteError(c, http.StatusForbidden, errAccessDenied)
		return
	}
	var s models.Session
	if !BindOr400(c, &s) {
		return
	}
	result, err := db.Client.Session.Create().SetCharacterID(charID).SetSessionDate(s.SessionDate).SetTitle(s.Title).SetNotes(s.Notes).SetXpEarned(s.XPEarned).SetGoldEarned(s.GoldEarned).SetImportantEvents(s.ImportantEvents).Save(c.Request.Context())
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	WriteJSON(c, http.StatusCreated, gin.H{"id": result.ID})
}

func UpdateSession(c *gin.Context) {
	sid, _ := strconv.ParseInt(c.Param("sid"), 10, 64)
	sess, err := db.Client.Session.Get(c.Request.Context(), sid)
	if err != nil {
		WriteNotFound(c, "session not found")
		return
	}
	if !canEditCharacterID(c, sess.CharacterID) {
		WriteError(c, http.StatusForbidden, errAccessDenied)
		return
	}
	var s models.Session
	if !BindOr400(c, &s) {
		return
	}
	db.Client.Session.UpdateOneID(sid).SetSessionDate(s.SessionDate).SetTitle(s.Title).SetNotes(s.Notes).SetXpEarned(s.XPEarned).SetGoldEarned(s.GoldEarned).SetImportantEvents(s.ImportantEvents).Save(c.Request.Context())
	WriteJSON(c, http.StatusOK, gin.H{"ok": true})
}

func DeleteSession(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("sid"), 10, 64)
	sess, err := db.Client.Session.Get(c.Request.Context(), id)
	if err != nil {
		WriteNotFound(c, "session not found")
		return
	}
	if !canEditCharacterID(c, sess.CharacterID) {
		WriteError(c, http.StatusForbidden, errAccessDenied)
		return
	}
	db.Client.Session.DeleteOneID(id).Exec(c.Request.Context())
	WriteJSON(c, http.StatusOK, gin.H{"ok": true})
}

func ListQuests(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	quests, err := db.Client.Quest.Query().Where(quest.CharacterID(charID)).Order(quest.ByStatus(), quest.ByName()).All(c.Request.Context())
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	var out = make([]models.Quest, 0)
	for _, q := range quests {
		out = append(out, models.Quest{ID: q.ID, CharacterID: q.CharacterID, Name: q.Name, Description: q.Description, Status: q.Status, Objectives: q.Objectives, Rewards: q.Rewards, Notes: q.Notes, CreatedAt: q.CreatedAt, UpdatedAt: q.UpdatedAt})
	}
	WriteJSON(c, http.StatusOK, out)
}

func CreateQuest(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if !canEditCharacterID(c, charID) {
		WriteError(c, http.StatusForbidden, errAccessDenied)
		return
	}
	var q models.Quest
	if !BindOr400(c, &q) {
		return
	}
	if q.Status == "" {
		q.Status = "active"
	}
	result, err := db.Client.Quest.Create().SetCharacterID(charID).SetName(q.Name).SetDescription(q.Description).SetStatus(q.Status).SetObjectives(q.Objectives).SetRewards(q.Rewards).SetNotes(q.Notes).Save(c.Request.Context())
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	WriteJSON(c, http.StatusCreated, gin.H{"id": result.ID})
}

func UpdateQuest(c *gin.Context) {
	qid, _ := strconv.ParseInt(c.Param("qid"), 10, 64)
	if !canEditResourceID(c, "quests", qid) {
		WriteError(c, http.StatusForbidden, errAccessDenied)
		return
	}
	var q models.Quest
	if !BindOr400(c, &q) {
		return
	}
	db.Client.Quest.UpdateOneID(qid).SetName(q.Name).SetDescription(q.Description).SetStatus(q.Status).SetObjectives(q.Objectives).SetRewards(q.Rewards).SetNotes(q.Notes).SetUpdatedAt(time.Now().Format("2006-01-02 15:04:05")).Save(c.Request.Context())
	WriteJSON(c, http.StatusOK, gin.H{"ok": true})
}

func DeleteQuest(c *gin.Context) {
	qid, _ := strconv.ParseInt(c.Param("qid"), 10, 64)
	q, err := db.Client.Quest.Get(c.Request.Context(), qid)
	if err != nil {
		WriteNotFound(c, "quest not found")
		return
	}
	if !canEditCharacterID(c, q.CharacterID) {
		WriteError(c, http.StatusForbidden, errAccessDenied)
		return
	}
	db.Client.Quest.DeleteOneID(qid).Exec(c.Request.Context())
	WriteJSON(c, http.StatusOK, gin.H{"ok": true})
}

func ListJournal(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	entries, err := db.Client.JournalEntry.Query().Where(journalentry.CharacterID(charID)).Order(ent.Desc(journalentry.FieldEntryDate)).All(c.Request.Context())
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	var out = make([]models.JournalEntry, 0)
	for _, j := range entries {
		out = append(out, models.JournalEntry{ID: j.ID, CharacterID: j.CharacterID, Title: j.Title, Entry: j.Entry, EntryDate: j.EntryDate, CreatedAt: j.CreatedAt})
	}
	WriteJSON(c, http.StatusOK, out)
}

func CreateJournalEntry(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if !canEditCharacterID(c, charID) {
		WriteError(c, http.StatusForbidden, errAccessDenied)
		return
	}
	var j models.JournalEntry
	if !BindOr400(c, &j) {
		return
	}
	result, err := db.Client.JournalEntry.Create().SetCharacterID(charID).SetTitle(j.Title).SetEntry(j.Entry).SetEntryDate(j.EntryDate).Save(c.Request.Context())
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	WriteJSON(c, http.StatusCreated, gin.H{"id": result.ID})
}

func UpdateJournalEntry(c *gin.Context) {
	jid, _ := strconv.ParseInt(c.Param("jid"), 10, 64)
	je, err := db.Client.JournalEntry.Get(c.Request.Context(), jid)
	if err != nil {
		WriteNotFound(c, "journal entry not found")
		return
	}
	if !canEditCharacterID(c, je.CharacterID) {
		WriteError(c, http.StatusForbidden, errAccessDenied)
		return
	}
	var j models.JournalEntry
	if !BindOr400(c, &j) {
		return
	}
	db.Client.JournalEntry.UpdateOneID(jid).SetTitle(j.Title).SetEntry(j.Entry).SetEntryDate(j.EntryDate).Save(c.Request.Context())
	WriteJSON(c, http.StatusOK, gin.H{"ok": true})
}

func DeleteJournalEntry(c *gin.Context) {
	jid, _ := strconv.ParseInt(c.Param("jid"), 10, 64)
	je, err := db.Client.JournalEntry.Get(c.Request.Context(), jid)
	if err != nil {
		WriteNotFound(c, "journal entry not found")
		return
	}
	if !canEditCharacterID(c, je.CharacterID) {
		WriteError(c, http.StatusForbidden, errAccessDenied)
		return
	}
	db.Client.JournalEntry.DeleteOneID(jid).Exec(c.Request.Context())
	WriteJSON(c, http.StatusOK, gin.H{"ok": true})
}
