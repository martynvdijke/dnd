package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"villum/crypto"
	"villum/db"
	"villum/ent/aiendpoint"
	"villum/handlers/testutil"
)

func inkRouter(userID int64, role string) *gin.Engine {
	return testutil.NewRouterWithUser(func(auth *gin.RouterGroup) {
		RegisterCampaignRoutes(auth)
	}, userID, role)
}

func inkMultipart(t *testing.T, fields map[string]string, pages [][]byte, texts []string) (*bytes.Buffer, string) {
	t.Helper()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	for k, v := range fields {
		w.WriteField(k, v)
	}
	for _, text := range texts {
		w.WriteField("text", text)
	}
	for i, page := range pages {
		fw, err := w.CreateFormFile("pages", "page.png")
		if err != nil {
			t.Fatalf("create page part: %v", err)
		}
		if _, err := fw.Write(page); err != nil {
			t.Fatalf("write page: %v", err)
		}
		_ = i
	}
	w.Close()
	return &body, w.FormDataContentType()
}

func postInkCapture(t *testing.T, r *gin.Engine, fields map[string]string, pages [][]byte, texts []string) *httptest.ResponseRecorder {
	t.Helper()
	body, ct := inkMultipart(t, fields, pages, texts)
	req := httptest.NewRequest("POST", "/api/ink/captures", body)
	req.Header.Set("Content-Type", ct)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func parseInkCapture(t *testing.T, rec *httptest.ResponseRecorder) inkCapture {
	t.Helper()
	var resp struct {
		Capture    inkCapture `json:"capture"`
		Duplicates int        `json:"duplicates"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse capture: %v (%s)", err, rec.Body.String())
	}
	return resp.Capture
}

func TestInkCaptureLifecycle(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	SetMediaPath(t.TempDir())
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCampaign(t, 1, "InkCamp", "Party", 1)

	r := inkRouter(1, "admin")
	page := uniqueTestPNG(t)
	rec := postInkCapture(t, r, map[string]string{
		"campaign_id": "1",
		"title":       "Session 12",
		"session_key": "2026-10-20",
	}, [][]byte{page, uniqueTestPNG(t)}, []string{"The party entered the crypt."})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create: want 201 got %d %s", rec.Code, rec.Body.String())
	}
	cap := parseInkCapture(t, rec)
	if len(cap.Pages) != 2 {
		t.Fatalf("want 2 pages got %d", len(cap.Pages))
	}
	if cap.Status != "transcribed" || cap.OCRSource != "boox" {
		t.Errorf("sidecar should mark transcribed/boox, got %s/%s", cap.Status, cap.OCRSource)
	}
	if cap.Pages[0].Text == "" || cap.Pages[0].URL == "" {
		t.Errorf("first page missing text or url: %+v", cap.Pages[0])
	}

	list := testutil.Get(t, r, "/api/ink/captures?campaign_id=1")
	if list.Code != http.StatusOK {
		t.Fatalf("list: %d %s", list.Code, list.Body.String())
	}
	var captures []inkCapture
	if err := json.Unmarshal(list.Body.Bytes(), &captures); err != nil {
		t.Fatalf("parse list: %v", err)
	}
	if len(captures) != 1 {
		t.Fatalf("want 1 capture got %d", len(captures))
	}

	// Manual correction marks the capture reviewed.
	req := httptest.NewRequest("PATCH", "/api/ink/captures/"+inkID(cap.ID), bytes.NewReader([]byte(`{"pages":[{"id":`+inkID(cap.Pages[0].ID)+`,"text":"Corrected text"}]}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("patch: %d %s", w.Code, w.Body.String())
	}
	var updated inkCapture
	if err := json.Unmarshal(w.Body.Bytes(), &updated); err != nil {
		t.Fatalf("parse patched capture: %v (%s)", err, w.Body.String())
	}
	if updated.Status != "reviewed" {
		t.Errorf("want reviewed got %s", updated.Status)
	}

	promote := testutil.PostJSON(t, r, "/api/ink/captures/"+inkID(cap.ID)+"/promote", map[string]any{"target": "recap"})
	if promote.Code != http.StatusCreated {
		t.Fatalf("promote: %d %s", promote.Code, promote.Body.String())
	}
	var recapID int64
	if err := db.DB.QueryRow("SELECT recap_id FROM ink_captures WHERE id=?", cap.ID).Scan(&recapID); err != nil || recapID == 0 {
		t.Fatalf("capture not linked to recap: %v", err)
	}
	var isSent bool
	if err := db.DB.QueryRow("SELECT is_sent FROM campaign_recaps WHERE id=?", recapID).Scan(&isSent); err != nil {
		t.Fatalf("recap missing: %v", err)
	}
	if isSent {
		t.Error("promoted recap must stay unpublished")
	}

	second := testutil.PostJSON(t, r, "/api/ink/captures/"+inkID(cap.ID)+"/promote", map[string]any{"target": "recap"})
	if second.Code != http.StatusConflict {
		t.Errorf("double promote: want 409 got %d", second.Code)
	}
}

func TestInkCaptureDuplicatePage(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	SetMediaPath(t.TempDir())
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCampaign(t, 1, "InkCamp", "Party", 1)

	r := inkRouter(1, "admin")
	page := uniqueTestPNG(t)
	rec := postInkCapture(t, r, map[string]string{"campaign_id": "1"}, [][]byte{page, page}, nil)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Capture    inkCapture `json:"capture"`
		Duplicates int        `json:"duplicates"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if len(resp.Capture.Pages) != 1 || resp.Duplicates != 1 {
		t.Errorf("want 1 page and 1 duplicate, got %d/%d", len(resp.Capture.Pages), resp.Duplicates)
	}
}

func TestInkCaptureHiddenFromOtherMembers(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	SetMediaPath(t.TempDir())
	testutil.SeedUser(t, 1, "dm", "user")
	testutil.SeedUser(t, 2, "player", "user")
	testutil.SeedCampaign(t, 1, "InkCamp", "Party", 1)
	testutil.SeedCampaignMember(t, 1, 2, "player")

	dmRouter := inkRouter(1, "user")
	rec := postInkCapture(t, dmRouter, map[string]string{"campaign_id": "1"}, [][]byte{uniqueTestPNG(t)}, nil)
	if rec.Code != http.StatusCreated {
		t.Fatalf("dm create: %d %s", rec.Code, rec.Body.String())
	}
	cap := parseInkCapture(t, rec)

	player := inkRouter(2, "user")
	if got := testutil.Get(t, player, "/api/ink/captures?campaign_id=1"); got.Code != http.StatusOK {
		t.Fatalf("list: %d", got.Code)
	} else {
		var captures []inkCapture
		json.Unmarshal(got.Body.Bytes(), &captures)
		if len(captures) != 0 {
			t.Errorf("other member must not see the DM capture, got %d", len(captures))
		}
	}
	if got := testutil.Get(t, player, "/api/ink/captures/"+inkID(cap.ID)); got.Code != http.StatusForbidden {
		t.Errorf("read: want 403 got %d", got.Code)
	}
	if got := testutil.PostJSON(t, player, "/api/ink/captures/"+inkID(cap.ID)+"/promote", map[string]any{"target": "journal"}); got.Code != http.StatusForbidden {
		t.Errorf("promote: want 403 got %d", got.Code)
	}
}

func TestInkPlayerJournalPromotion(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	SetMediaPath(t.TempDir())
	testutil.SeedUser(t, 1, "dm", "user")
	testutil.SeedUser(t, 2, "player", "user")
	testutil.SeedCampaign(t, 1, "InkCamp", "Party", 1)
	testutil.SeedCampaignMember(t, 1, 2, "player")
	testutil.SeedCharacterInCampaign(t, 7, 2, 1, "Aria", "Elf", "Wizard")

	player := inkRouter(2, "user")
	rec := postInkCapture(t, player, map[string]string{
		"campaign_id":  "1",
		"character_id": "7",
		"title":        "Aria's notes",
	}, [][]byte{uniqueTestPNG(t)}, []string{"I found a hidden door."})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", rec.Code, rec.Body.String())
	}
	cap := parseInkCapture(t, rec)

	denied := testutil.PostJSON(t, player, "/api/ink/captures/"+inkID(cap.ID)+"/promote", map[string]any{"target": "recap"})
	if denied.Code != http.StatusForbidden {
		t.Errorf("player recap promote: want 403 got %d", denied.Code)
	}

	promote := testutil.PostJSON(t, player, "/api/ink/captures/"+inkID(cap.ID)+"/promote", map[string]any{"target": "journal"})
	if promote.Code != http.StatusCreated {
		t.Fatalf("journal promote: %d %s", promote.Code, promote.Body.String())
	}
	var journalID int64
	if err := db.DB.QueryRow("SELECT journal_entry_id FROM ink_captures WHERE id=?", cap.ID).Scan(&journalID); err != nil || journalID == 0 {
		t.Fatalf("capture not linked to journal: %v", err)
	}
	var charID int64
	var entry string
	if err := db.DB.QueryRow("SELECT character_id, entry FROM journal WHERE id=?", journalID).Scan(&charID, &entry); err != nil {
		t.Fatalf("journal row missing: %v", err)
	}
	if charID != 7 || entry == "" {
		t.Errorf("unexpected journal row char=%d entry=%q", charID, entry)
	}
}

func TestInkOCRWithoutVisionEndpoint(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	SetMediaPath(t.TempDir())
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCampaign(t, 1, "InkCamp", "Party", 1)

	r := inkRouter(1, "admin")
	rec := postInkCapture(t, r, map[string]string{"campaign_id": "1"}, [][]byte{uniqueTestPNG(t)}, nil)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", rec.Code, rec.Body.String())
	}
	cap := parseInkCapture(t, rec)

	ocr := testutil.PostJSON(t, r, "/api/ink/captures/"+inkID(cap.ID)+"/ocr", map[string]any{})
	if ocr.Code != http.StatusConflict {
		t.Fatalf("ocr without vision: want 409 got %d %s", ocr.Code, ocr.Body.String())
	}
	var text string
	if err := db.DB.QueryRow("SELECT ocr_text FROM ink_pages WHERE capture_id=?", cap.ID).Scan(&text); err != nil {
		t.Fatal(err)
	}
	if text != "" {
		t.Errorf("failed OCR must not write text, got %q", text)
	}
	var status string
	db.DB.QueryRow("SELECT status FROM ink_captures WHERE id=?", cap.ID).Scan(&status)
	if status != "uploaded" {
		t.Errorf("want uploaded got %s", status)
	}
}

func TestInkCaptureRequiresImage(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	SetMediaPath(t.TempDir())
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCampaign(t, 1, "InkCamp", "Party", 1)

	r := inkRouter(1, "admin")
	body, ct := inkMultipart(t, map[string]string{"campaign_id": "1"}, nil, nil)
	req := httptest.NewRequest("POST", "/api/ink/captures", body)
	req.Header.Set("Content-Type", ct)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("no pages: want 400 got %d", w.Code)
	}
}

func inkID(v int64) string {
	return strconv.FormatInt(v, 10)
}

func TestInkVisionTranscription(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	SetMediaPath(t.TempDir())
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCampaign(t, 1, "InkCamp", "Party", 1)

	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["model"] != "vision-test" {
			t.Errorf("unexpected model %v", body["model"])
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"choices":[{"message":{"content":"The runes read: open the west gate."}}]}`))
	}))
	defer srv.Close()

	enc, err := crypto.Encrypt("test-key")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Client.AIEndpoint.Create().
		SetName("vision-test").SetType(aiendpoint.TypeVision).SetBaseURL(srv.URL).
		SetEncryptedAPIKey(enc).SetModel("vision-test").SetEnabled(true).
		SetCreatedAt(time.Now().UTC().Format(time.RFC3339)).
		SetUpdatedAt(time.Now().UTC().Format(time.RFC3339)).
		Save(context.Background()); err != nil {
		t.Fatalf("seed endpoint: %v", err)
	}

	r := inkRouter(1, "admin")
	rec := postInkCapture(t, r, map[string]string{"campaign_id": "1"}, [][]byte{uniqueTestPNG(t)}, nil)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", rec.Code, rec.Body.String())
	}
	cap := parseInkCapture(t, rec)

	ocr := testutil.PostJSON(t, r, "/api/ink/captures/"+inkID(cap.ID)+"/ocr", map[string]any{})
	if ocr.Code != http.StatusOK {
		t.Fatalf("ocr: %d %s", ocr.Code, ocr.Body.String())
	}
	var transcribed inkCapture
	json.Unmarshal(ocr.Body.Bytes(), &transcribed)
	if len(transcribed.Pages) != 1 || !strings.Contains(transcribed.Pages[0].Text, "west gate") {
		t.Fatalf("missing transcription: %+v", transcribed.Pages)
	}
	if transcribed.Pages[0].OCRSource != "vision" || transcribed.Status != "transcribed" {
		t.Errorf("want vision/transcribed got %s/%s", transcribed.Pages[0].OCRSource, transcribed.Status)
	}

	// A sidecar page is never sent to the model.
	before := atomic.LoadInt32(&calls)
	withText := postInkCapture(t, r, map[string]string{"campaign_id": "1"}, [][]byte{uniqueTestPNG(t)}, []string{"Boox text"})
	cap2 := parseInkCapture(t, withText)
	again := testutil.PostJSON(t, r, "/api/ink/captures/"+inkID(cap2.ID)+"/ocr", map[string]any{})
	if again.Code != http.StatusOK {
		t.Fatalf("sidecar ocr: %d", again.Code)
	}
	if atomic.LoadInt32(&calls) != before {
		t.Errorf("sidecar text must not call the vision endpoint")
	}
}
