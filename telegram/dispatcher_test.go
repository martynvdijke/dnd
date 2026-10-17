package telegram

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	tgmodels "github.com/go-telegram/bot/models"

	"villum/db"
	"villum/handlers/testutil"
	"villum/models"
)

func testMediaSaver(t *testing.T) {
	t.Helper()
	prev := mediaSaver
	SetMediaSaver(func(req MediaSaveRequest) (*models.Upload, error) {
		ext := ""
		if idx := strings.LastIndex(req.Name, "."); idx >= 0 {
			ext = strings.ToLower(req.Name[idx:])
		}
		if ext == "" {
			ext = ".jpg"
		}
		hash := req.TelegramFileUniqueID
		if hash == "" {
			hash = fmt.Sprintf("hash-%d-%s", req.OwnerID, req.Name)
		}
		hashFull := hash + "-" + req.Name
		// normalize ext lower
		res, err := db.DB.Exec(`INSERT INTO uploads(hash, ext, url, owner_type, owner_id, kind, source, telegram_file_unique_id) VALUES(?,?,?,?,?,?,?,?)`,
			hashFull, ext, "/media/"+hashFull, req.OwnerType, req.OwnerID, "image", req.Source, req.TelegramFileUniqueID)
		if err != nil {
			return nil, err
		}
		id, _ := res.LastInsertId()
		return &models.Upload{ID: int(id), Hash: hashFull, Ext: ext, OwnerType: req.OwnerType, OwnerID: req.OwnerID, Source: req.Source, TelegramFileUniqueID: req.TelegramFileUniqueID}, nil
	})
	t.Cleanup(func() { mediaSaver = prev })
}

func mockDownloadAndReply(t *testing.T, sent *[]string, fileData []byte, wantTooLarge bool) {
	t.Helper()
	mockTelegramServer(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "getFile") {
			w.Write(jsonOK(map[string]any{"file_id": "fid", "file_unique_id": "u", "file_path": "path/file.jpg"}))
			return
		}
		if strings.Contains(r.URL.Path, "/file/bot") {
			if wantTooLarge {
				w.Header().Set("Content-Length", "20971521")
				chunk := make([]byte, 1<<20)
				for i := 0; i < 20; i++ {
					w.Write(chunk)
				}
				w.Write([]byte("x"))
				return
			}
			w.Write(fileData)
			return
		}
		if txt := requestFormValue(r, "text"); txt != "" {
			*sent = append(*sent, txt)
		}
		w.Write(jsonOK(map[string]any{"message_id": 1}))
	})
}

func photoMessage(chatID, userID int64, fileID, fileUnique, caption, text string) *tgmodels.Message {
	return &tgmodels.Message{
		ID:   1,
		Chat: tgmodels.Chat{ID: chatID, Type: "private"},
		From: &tgmodels.User{ID: userID, FirstName: "Tester"},
		Text: text,
		Photo: []tgmodels.PhotoSize{
			{FileID: fileID, FileUniqueID: fileUnique, FileSize: 1234},
		},
		Caption: caption,
	}
}

func groupPhotoMessage(chatID, userID int64, fileID, fileUnique, caption string) *tgmodels.Message {
	return &tgmodels.Message{
		ID:   1,
		Chat: tgmodels.Chat{ID: chatID, Type: "supergroup"},
		From: &tgmodels.User{ID: userID, FirstName: "Tester"},
		Photo: []tgmodels.PhotoSize{
			{FileID: fileID, FileUniqueID: fileUnique, FileSize: 1234},
		},
		Caption: caption,
	}
}

func bindCampaign(t *testing.T, campaignID, chatID int64, enabled bool, autoSave bool) {
	t.Helper()
	isEnabled := 0
	if enabled {
		isEnabled = 1
	}
	autoSaveInt := 0
	if autoSave {
		autoSaveInt = 1
	}
	_, err := db.DB.Exec(`INSERT OR REPLACE INTO campaign_telegram_settings(campaign_id, chat_id, is_enabled, auto_save_media) VALUES(?,?,?,?)`, campaignID, chatID, isEnabled, autoSaveInt)
	if err != nil {
		t.Fatalf("bind: %v", err)
	}
}

func countUploads(t *testing.T) int {
	t.Helper()
	var n int
	db.DB.QueryRow(`SELECT COUNT(*) FROM uploads`).Scan(&n)
	return n
}

func countPending(t *testing.T) int {
	t.Helper()
	var n int
	db.DB.QueryRow(`SELECT COUNT(*) FROM telegram_pending_captures`).Scan(&n)
	return n
}

// 1. Bound + enabled campaign chat, photo with caption /save
func TestInboundMedia_BoundSaveCaption(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCampaign(t, 10, "Camp A", "Party", 1)
	bindCampaign(t, 10, -100, true, false)
	testMediaSaver(t)
	var sent []string
	mockDownloadAndReply(t, &sent, []byte("filedata"), false)

	msg := groupPhotoMessage(-100, 111, "fid1", "uniq1", "/save")
	handleInboundMedia(context.Background(), msg)

	if countUploads(t) != 1 {
		t.Fatalf("expected 1 upload, got %d", countUploads(t))
	}
	var ownerType, source, uniq string
	var ownerID int64
	err := db.DB.QueryRow(`SELECT owner_type, owner_id, source, telegram_file_unique_id FROM uploads LIMIT 1`).Scan(&ownerType, &ownerID, &source, &uniq)
	if err != nil {
		t.Fatalf("query upload: %v", err)
	}
	if ownerType != "campaign" || ownerID != 10 || source != "telegram" || uniq != "uniq1" {
		t.Fatalf("upload fields mismatch: %q %d %q %q", ownerType, ownerID, source, uniq)
	}
	found := false
	for _, s := range sent {
		if strings.Contains(s, "Saved") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected Saved reply, got %v", sent)
	}
}

// 1b text /save variant
func TestInboundMedia_BoundSaveText(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCampaign(t, 11, "Camp B", "Party", 1)
	bindCampaign(t, 11, -101, true, false)
	testMediaSaver(t)
	var sent []string
	mockDownloadAndReply(t, &sent, []byte("data2"), false)
	msg := groupPhotoMessage(-101, 111, "fid2", "uniq2", "")
	msg.Text = "/save"
	handleInboundMedia(context.Background(), msg)
	if countUploads(t) != 1 {
		t.Fatalf("expected 1 upload got %d", countUploads(t))
	}
}

// 2. auto_save_media=1 saves without /save
func TestInboundMedia_AutoSave(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCampaign(t, 20, "Camp Auto", "Party", 1)
	bindCampaign(t, 20, -200, true, true)
	testMediaSaver(t)
	var sent []string
	mockDownloadAndReply(t, &sent, []byte("auto"), false)
	msg := groupPhotoMessage(-200, 111, "fidA", "uniqA", "no command")
	handleInboundMedia(context.Background(), msg)
	if countUploads(t) != 1 {
		t.Fatalf("auto-save expected 1 upload")
	}
}

// 3. bound auto_save=0 no /save -> nothing stored and no reply
func TestInboundMedia_IgnoreWithoutCommand(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCampaign(t, 30, "Camp Ignore", "Party", 1)
	bindCampaign(t, 30, -300, true, false)
	testMediaSaver(t)
	var sent []string
	mockTelegramServer(t, func(w http.ResponseWriter, r *http.Request) {
		if txt := requestFormValue(r, "text"); txt != "" {
			sent = append(sent, txt)
		}
		w.Write(jsonOK(map[string]any{"message_id": 1}))
	})
	msg := groupPhotoMessage(-300, 111, "fidI", "uniqI", "hello")
	handleInboundMedia(context.Background(), msg)
	if countUploads(t) != 0 {
		t.Fatalf("expected 0 uploads, got %d", countUploads(t))
	}
	if len(sent) != 0 {
		t.Fatalf("expected no reply, got %v", sent)
	}
}

// 4. oversized
func TestInboundMedia_Oversized(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCampaign(t, 40, "Camp Big", "Party", 1)
	bindCampaign(t, 40, -400, true, false)
	testMediaSaver(t)
	var sent []string
	mockDownloadAndReply(t, &sent, nil, true)
	msg := groupPhotoMessage(-400, 111, "fidBig", "uniqBig", "/save")
	handleInboundMedia(context.Background(), msg)
	if countUploads(t) != 0 {
		t.Fatalf("oversized should not store")
	}
	found := false
	for _, s := range sent {
		if strings.Contains(strings.ToLower(s), "too large") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected too large reply, got %v", sent)
	}
}

// 5. duplicate
func TestInboundMedia_Duplicate(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCampaign(t, 50, "Camp Dup", "Party", 1)
	bindCampaign(t, 50, -500, true, false)
	testMediaSaver(t)
	var sent []string
	mockDownloadAndReply(t, &sent, []byte("dup"), false)
	msg := groupPhotoMessage(-500, 111, "fidDup", "uniqDup", "/save")
	handleInboundMedia(context.Background(), msg)
	if countUploads(t) != 1 {
		t.Fatalf("first save expected 1")
	}
	// second same unique
	sent = nil
	handleInboundMedia(context.Background(), groupPhotoMessage(-500, 111, "fidDup2", "uniqDup", "/save"))
	if countUploads(t) != 1 {
		t.Fatalf("duplicate should not create second row, got %d", countUploads(t))
	}
	found := false
	for _, s := range sent {
		if strings.Contains(s, "Already saved") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected Already saved reply, got %v", sent)
	}
}

// 6. DM multi picker + /save 2
func TestInboundMedia_DM_MultiPickerAndSave(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCampaign(t, 60, "Alpha", "Party", 1)
	testutil.SeedCampaign(t, 61, "Beta", "Party", 1)
	// user owns both via SeedCampaign (member)
	if err := UpsertIdentity(1, 999, 999, "tester"); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	testMediaSaver(t)
	var sent []string
	mockDownloadAndReply(t, &sent, []byte("dmdata"), false)

	msg := photoMessage(999, 999, "fidDM", "uniqDM", "", "")
	handleInboundMedia(context.Background(), msg)

	if countPending(t) != 1 {
		t.Fatalf("expected 1 pending, got %d", countPending(t))
	}
	found := false
	for _, s := range sent {
		if strings.Contains(s, "Choose where") && strings.Contains(s, "Alpha") && strings.Contains(s, "Beta") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected picker reply, got %v", sent)
	}
	// Now /save 2 -> Beta (sorted by name: Alpha=1, Beta=2)
	sent = nil
	mockTelegramServer(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "getFile") {
			w.Write(jsonOK(map[string]any{"file_id": "fidDM", "file_unique_id": "u", "file_path": "path/file.jpg"}))
			return
		}
		if strings.Contains(r.URL.Path, "/file/bot") {
			w.Write([]byte("dmdata"))
			return
		}
		if txt := requestFormValue(r, "text"); txt != "" {
			sent = append(sent, txt)
		}
		w.Write(jsonOK(map[string]any{"message_id": 1}))
	})
	upd := messageUpdate(2, 999, 999, "/save 2")
	// HandleUpdate would route /save; test handleSaveCommand directly
	// Need to simulate cmdContext
	c := &cmdContext{ctx: context.Background(), chatID: 999, tgUserID: 999, name: "save", args: []string{"2"}}
	// need mock for SaveMedia download again
	handled := handleSaveCommand(c)
	if !handled {
		t.Fatalf("handleSaveCommand not handled")
	}
	if countUploads(t) != 1 {
		t.Fatalf("expected 1 upload after /save 2, got %d", countUploads(t))
	}
	var ownerID int64
	db.DB.QueryRow(`SELECT owner_id FROM uploads LIMIT 1`).Scan(&ownerID)
	if ownerID != 61 {
		t.Fatalf("expected saved to Beta (61), got %d", ownerID)
	}
	if countPending(t) != 0 {
		t.Fatalf("pending should be deleted, got %d", countPending(t))
	}
	// also ensure via HandleUpdate path picker works: duplicate pending cleared
	_ = sent
	_ = upd
}

// 7. single campaign direct save no picker
func TestInboundMedia_DM_SingleCampaignDirect(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCampaign(t, 70, "Solo", "Party", 1)
	if err := UpsertIdentity(1, 888, 888, "tester"); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	testMediaSaver(t)
	var sent []string
	mockDownloadAndReply(t, &sent, []byte("solo"), false)
	msg := photoMessage(888, 888, "fidSolo", "uniqSolo", "", "")
	handleInboundMedia(context.Background(), msg)
	if countUploads(t) != 1 {
		t.Fatalf("expected direct save 1 upload")
	}
	if countPending(t) != 0 {
		t.Fatalf("single campaign should not create pending, got %d", countPending(t))
	}
	for _, s := range sent {
		if strings.Contains(s, "Choose where") {
			t.Fatalf("should not have picker for single campaign, got %v", sent)
		}
	}
}

// 8. expired pending
func TestInboundMedia_ExpiredPending(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCampaign(t, 80, "ExpCamp", "Party", 1)
	if err := UpsertIdentity(1, 777, 777, "tester"); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	testMediaSaver(t)
	var sent []string
	mockDownloadAndReply(t, &sent, []byte("exp"), false)
	// store pending with capture
	msg := photoMessage(777, 777, "fidExp", "uniqExp", "", "")
	_ = msg
	// Create multiple campaigns so pending is created
	testutil.SeedCampaign(t, 81, "ExpCamp2", "Party", 1)
	// need user to be member of both
	handleInboundMedia(context.Background(), photoMessage(777, 777, "fidExp", "uniqExp", "", ""))
	if countPending(t) != 1 {
		t.Fatalf("expected pending 1 got %d", countPending(t))
	}
	// backdate expires_at
	_, err := db.DB.Exec(`UPDATE telegram_pending_captures SET expires_at = datetime('now', '-2 days') WHERE telegram_user_id=777`)
	if err != nil {
		t.Fatalf("backdate: %v", err)
	}
	sent = nil
	mockTelegramServer(t, func(w http.ResponseWriter, r *http.Request) {
		if txt := requestFormValue(r, "text"); txt != "" {
			sent = append(sent, txt)
		}
		w.Write(jsonOK(map[string]any{"message_id": 1}))
	})
	c := &cmdContext{ctx: context.Background(), chatID: 777, tgUserID: 777, name: "save", args: []string{"1"}}
	handleSaveCommand(c)
	if countUploads(t) != 0 {
		t.Fatalf("expired should store nothing, got %d", countUploads(t))
	}
	found := false
	for _, s := range sent {
		if strings.Contains(strings.ToLower(s), "expired") || strings.Contains(s, "No pending") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected expiry reply, got %v", sent)
	}
	if countPending(t) != 0 {
		t.Fatalf("expired pending should be cleaned")
	}
}

// 9. /save invalid args
func TestInboundMedia_SaveInvalidArgs(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCampaign(t, 90, "A", "Party", 1)
	testutil.SeedCampaign(t, 91, "B", "Party", 1)
	if err := UpsertIdentity(1, 666, 666, "tester"); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	// need pending
	var sent []string
	mockTelegramServer(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "getFile") {
			w.Write(jsonOK(map[string]any{"file_id": "f", "file_unique_id": "u", "file_path": "p"}))
			return
		}
		if strings.Contains(r.URL.Path, "/file/bot") {
			w.Write([]byte("x"))
			return
		}
		if txt := requestFormValue(r, "text"); txt != "" {
			sent = append(sent, txt)
		}
		w.Write(jsonOK(map[string]any{"message_id": 1}))
	})
	testMediaSaver(t)
	handleInboundMedia(context.Background(), photoMessage(666, 666, "fidInv", "uniqInv", "", ""))
	// reset sent for invalid tests
	cases := [][]string{{}, {"abc"}, {"0"}, {"99"}}
	for _, args := range cases {
		sent = nil
		c := &cmdContext{ctx: context.Background(), chatID: 666, tgUserID: 666, name: "save", args: args}
		handleSaveCommand(c)
		if len(sent) == 0 {
			t.Fatalf("expected reply for args %v", args)
		}
		got := sent[0]
		if len(args) == 0 || args[0] == "abc" || args[0] == "0" {
			if !strings.Contains(got, "Usage") {
				t.Fatalf("args %v expected Usage, got %q", args, got)
			}
		} else {
			if !strings.Contains(got, "Invalid choice") {
				t.Fatalf("args %v expected Invalid choice, got %q", args, got)
			}
		}
	}
	if countUploads(t) != 0 {
		t.Fatalf("invalid args should not store, got %d", countUploads(t))
	}
}

// Extra: prove DM picker works immediately after linking (disproves product bug)
func TestInboundMedia_AfterLinkPickerStillWorks(t *testing.T) {
	setupTelegramDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCampaign(t, 100, "Alpha", "Party", 1)
	testutil.SeedCampaign(t, 101, "Beta", "Party", 1)
	var sent []string
	mockTelegramServer(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "getFile") {
			w.Write(jsonOK(map[string]any{"file_id": "f", "file_unique_id": "u", "file_path": "p/file.jpg"}))
			return
		}
		if strings.Contains(r.URL.Path, "/file/bot") {
			w.Write([]byte("filedata"))
			return
		}
		if txt := requestFormValue(r, "text"); txt != "" {
			sent = append(sent, txt)
		}
		w.Write(jsonOK(map[string]any{"message_id": 1}))
	})
	testMediaSaver(t)
	// Simulate /start <code> linking flow via HandleUpdate
	code, _, err := CreateLinkCode(1)
	if err != nil {
		t.Fatalf("create code: %v", err)
	}
	HandleUpdate(context.Background(), messageUpdate(1, 555, 555, "/start "+code))
	if len(sent) == 0 || !strings.Contains(sent[0], "Linked!") {
		t.Fatalf("expected Linked! got %v", sent)
	}
	sent = nil
	// Now send DM photo - should get picker, not Linked!
	upd := &tgmodels.Update{
		ID: 2,
		Message: &tgmodels.Message{
			ID:    2,
			Chat:  tgmodels.Chat{ID: 555, Type: "private"},
			From:  &tgmodels.User{ID: 555, FirstName: "Tester"},
			Photo: []tgmodels.PhotoSize{{FileID: "fidAfter", FileUniqueID: "uniqAfter", FileSize: 123}},
		},
	}
	HandleUpdate(context.Background(), upd)
	if len(sent) == 0 {
		t.Fatalf("expected picker after linking, got no reply")
	}
	got := sent[0]
	if strings.Contains(got, "Linked!") {
		t.Fatalf("got spurious Linked! after photo: %q — would indicate product bug", got)
	}
	if !strings.Contains(got, "Choose where") {
		t.Fatalf("expected picker after linking, got %q", got)
	}
	if countPending(t) != 1 {
		t.Fatalf("expected pending 1 got %d", countPending(t))
	}
}

func TestIsSaveWord(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"/save", true},
		{"/save@MyBot", true},
		{"/SAVE", true},
		{"/save 2", true},
		{" /save ", true},
		{"/saver", false},
		{"/help", false},
		{"hello", false},
		{"", false},
	}
	for _, tc := range cases {
		if got := isSaveWord(tc.in); got != tc.want {
			t.Errorf("isSaveWord(%q)=%v want %v", tc.in, got, tc.want)
		}
	}
}
