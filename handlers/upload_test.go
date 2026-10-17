package handlers

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"mime/multipart"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/handlers/testutil"
)

var pngCounter int64

// uniqueTestPNG generates a unique PNG each call using an atomic counter
func uniqueTestPNG(t *testing.T) []byte {
	t.Helper()
	pngCounter++
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	seed := time.Now().UnixNano() + pngCounter
	for y := range 10 {
		for x := range 10 {
			idx := y*img.Stride + x*4
			img.Pix[idx+0] = uint8(seed + int64(x))
			img.Pix[idx+1] = uint8(seed>>8 + int64(y))
			img.Pix[idx+2] = uint8(seed>>16 + int64(x*y))
			img.Pix[idx+3] = 255
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode test PNG: %v", err)
	}
	return buf.Bytes()
}

func uploadTestImage(t *testing.T, r *gin.Engine, path string, ownerType, ownerID string, pngData []byte) *httptest.ResponseRecorder {
	t.Helper()
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	fw, err := w.CreateFormFile("image", "test.png")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	_, err = fw.Write(pngData)
	if err != nil {
		t.Fatalf("write png: %v", err)
	}
	if ownerType != "" {
		w.WriteField("owner_type", ownerType)
	}
	if ownerID != "" {
		w.WriteField("owner_id", ownerID)
	}
	w.Close()

	req := httptest.NewRequest("POST", path, &b)
	req.Header.Set("Content-Type", w.FormDataContentType())
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func testRouter() *gin.Engine {
	return testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/upload", HandleUpload)
		auth.GET("/uploads", GetUploads)
		auth.POST("/uploads/:id/crop", HandleCropUpload)
		auth.POST("/upload-links", CreateUploadLink)
		auth.DELETE("/upload-links/:id", DeleteUploadLink)
	})
}

func TestUploadAndGetUploads(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	SetMediaPath(t.TempDir())
	defer SetMediaPath("/app/media")

	r := testRouter()
	img1 := uniqueTestPNG(t)

	t.Run("upload image returns 200 with url and id", func(t *testing.T) {
		w := uploadTestImage(t, r, "/api/upload", "", "", img1)
		testutil.AssertStatus(t, w, 200)
		var data map[string]any
		testutil.ParseJSON(t, w, &data)
		if _, ok := data["url"]; !ok {
			t.Fatalf("response missing url: %s", w.Body.String())
		}
		if _, ok := data["id"]; !ok {
			t.Fatalf("response missing id: %s", w.Body.String())
		}
		if !strings.HasPrefix(data["url"].(string), "/media/") {
			t.Fatalf("url should start with /media/: %s", data["url"])
		}
	})

	t.Run("re-upload same file returns duplicate", func(t *testing.T) {
		w := uploadTestImage(t, r, "/api/upload", "", "", img1)
		testutil.AssertStatus(t, w, 200)
		var data map[string]any
		testutil.ParseJSON(t, w, &data)
		if data["duplicate"] != true {
			t.Fatalf("expected duplicate=true, got %+v", data)
		}
	})

	t.Run("upload invalid file type returns 400", func(t *testing.T) {
		var b bytes.Buffer
		w := multipart.NewWriter(&b)
		fw, _ := w.CreateFormFile("image", "test.txt")
		fw.Write([]byte("not an image"))
		w.Close()
		req := httptest.NewRequest("POST", "/api/upload", &b)
		req.Header.Set("Content-Type", w.FormDataContentType())
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		testutil.AssertStatus(t, rec, 400)
	})

	var uploadID int64
	t.Run("upload with different image has unique id", func(t *testing.T) {
		w := uploadTestImage(t, r, "/api/upload", "campaign", "1", uniqueTestPNG(t))
		testutil.AssertStatus(t, w, 200)
		var data map[string]any
		testutil.ParseJSON(t, w, &data)
		if data["duplicate"] == true {
			t.Fatalf("new image should not be duplicate: %s", w.Body.String())
		}
		id, ok := data["id"].(float64)
		if !ok {
			t.Fatalf("response missing id: %s", w.Body.String())
		}
		uploadID = int64(id)
	})

	t.Run("GetUploads filters by owner_type", func(t *testing.T) {
		w := testutil.Get(t, r, fmt.Sprintf("/api/uploads?owner_type=campaign&owner_id=1"))
		testutil.AssertStatus(t, w, 200)
		var uploads []map[string]any
		testutil.ParseJSON(t, w, &uploads)
		if len(uploads) == 0 {
			t.Fatal("expected at least one upload for campaign")
		}
	})

	t.Run("GetUploads returns empty for non-existent owner", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/uploads?owner_type=oneshot&owner_id=99999")
		testutil.AssertStatus(t, w, 200)
		var uploads []map[string]any
		testutil.ParseJSON(t, w, &uploads)
		if len(uploads) != 0 {
			t.Fatalf("expected empty list, got %d items", len(uploads))
		}
	})

	_ = uploadID // used by upload_links test
}

func TestUploadLinks(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	SetMediaPath(t.TempDir())
	defer SetMediaPath("/app/media")

	r := testRouter()

	// Upload image first
	w := uploadTestImage(t, r, "/api/upload", "character", "1", uniqueTestPNG(t))
	testutil.AssertStatus(t, w, 200)
	var uploadResp map[string]any
	testutil.ParseJSON(t, w, &uploadResp)
	uploadID := int64(uploadResp["id"].(float64))

	t.Run("CreateUploadLink succeeds with valid data", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/upload-links", map[string]any{
			"upload_id":   uploadID,
			"entity_type": "campaign",
			"entity_id":   42,
			"field_name":  "map",
		})
		testutil.AssertStatus(t, w, 201)
		var data map[string]any
		testutil.ParseJSON(t, w, &data)
		if data["upload_id"] != float64(uploadID) {
			t.Fatalf("expected upload_id %d, got %v", uploadID, data["upload_id"])
		}
		if data["entity_type"] != "campaign" {
			t.Fatalf("expected entity_type campaign, got %v", data["entity_type"])
		}
	})

	t.Run("CreateUploadLink fails with missing fields", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/upload-links", map[string]any{
			"upload_id": uploadID,
		})
		testutil.AssertStatus(t, w, 400)
	})

	t.Run("CreateUploadLink fails with empty entity_type", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/upload-links", map[string]any{
			"upload_id":   uploadID,
			"entity_type": "",
			"entity_id":   42,
		})
		testutil.AssertStatus(t, w, 400)
	})

	t.Run("CreateUploadLink fails with invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/upload-links", bytes.NewReader([]byte("{not-json}")))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		testutil.AssertStatus(t, rec, 400)
	})

	t.Run("GetUploads filters by entity_type via upload_links JOIN", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/uploads?entity_type=campaign&entity_id=42")
		testutil.AssertStatus(t, w, 200)
		var uploads []map[string]any
		testutil.ParseJSON(t, w, &uploads)
		if len(uploads) == 0 {
			t.Fatal("expected at least one upload via upload_links")
		}
	})

	t.Run("DeleteUploadLink succeeds", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/upload-links", map[string]any{
			"upload_id":   uploadID,
			"entity_type": "oneshot",
			"entity_id":   7,
		})
		testutil.AssertStatus(t, w, 201)
		var link map[string]any
		testutil.ParseJSON(t, w, &link)
		linkID := int64(link["id"].(float64))

		w = testutil.Delete(t, r, fmt.Sprintf("/api/upload-links/%d", linkID))
		testutil.AssertStatus(t, w, 204)
	})

	t.Run("DeleteUploadLink with invalid id returns 400", func(t *testing.T) {
		w := testutil.Delete(t, r, "/api/upload-links/abc")
		testutil.AssertStatus(t, w, 400)
	})

	t.Run("GetUploads returns empty after link deleted", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/uploads?entity_type=oneshot&entity_id=7")
		testutil.AssertStatus(t, w, 200)
		var uploads []map[string]any
		testutil.ParseJSON(t, w, &uploads)
		if len(uploads) != 0 {
			t.Fatalf("expected empty list after link delete, got %d items", len(uploads))
		}
	})
}

func TestCropUpload(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	mediaDir := t.TempDir()
	SetMediaPath(mediaDir)
	defer SetMediaPath("/app/media")

	r := testRouter()
	pngData := uniqueTestPNG(t)

	w := uploadTestImage(t, r, "/api/upload", "character", "1", pngData)
	testutil.AssertStatus(t, w, 200)
	var uploadResp map[string]any
	testutil.ParseJSON(t, w, &uploadResp)
	uploadID := int64(uploadResp["id"].(float64))

	t.Run("crop valid upload returns 200", func(t *testing.T) {
		w := testutil.PostJSON(t, r, fmt.Sprintf("/api/uploads/%d/crop", uploadID), map[string]any{
			"x": 1, "y": 1, "width": 5, "height": 5,
			"target_width": 10, "target_height": 10,
		})
		testutil.AssertStatus(t, w, 200)
		var data map[string]any
		testutil.ParseJSON(t, w, &data)
		if _, ok := data["resized_url"]; !ok {
			t.Fatalf("response missing resized_url: %s", w.Body.String())
		}
		if _, ok := data["thumbnail_url"]; !ok {
			t.Fatalf("response missing thumbnail_url: %s", w.Body.String())
		}
	})

	t.Run("crop non-existent upload returns 404", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/uploads/99999/crop", map[string]any{
			"x": 0, "y": 0, "width": 10, "height": 10,
		})
		testutil.AssertStatus(t, w, 404)
	})

	t.Run("crop with invalid dimensions returns 400", func(t *testing.T) {
		w := testutil.PostJSON(t, r, fmt.Sprintf("/api/uploads/%d/crop", uploadID), map[string]any{
			"x": 0, "y": 0, "width": 0, "height": 0,
		})
		testutil.AssertStatus(t, w, 400)
	})

	t.Run("crop with out-of-bounds area returns 400", func(t *testing.T) {
		w := testutil.PostJSON(t, r, fmt.Sprintf("/api/uploads/%d/crop", uploadID), map[string]any{
			"x": 0, "y": 0, "width": 999, "height": 999,
		})
		testutil.AssertStatus(t, w, 400)
	})
}

func TestLargeUpload(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	SetMediaPath(t.TempDir())
	defer SetMediaPath("/app/media")

	r := testRouter()

	t.Run("small image uploads successfully", func(t *testing.T) {
		w := uploadTestImage(t, r, "/api/upload", "", "", uniqueTestPNG(t))
		testutil.AssertStatus(t, w, 200)
	})
}

// --- telegram-rich-media additions ---

func uploadBytes(t *testing.T, r *gin.Engine, filename string, data []byte) *httptest.ResponseRecorder {
	t.Helper()
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	fw, err := w.CreateFormFile("image", filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := fw.Write(data); err != nil {
		t.Fatalf("write bytes: %v", err)
	}
	w.Close()
	req := httptest.NewRequest("POST", "/api/upload", &b)
	req.Header.Set("Content-Type", w.FormDataContentType())
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestAudioVideoUpload(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	mediaDir := t.TempDir()
	SetMediaPath(mediaDir)
	defer SetMediaPath("/app/media")
	r := testRouter()

	// Minimal valid headers that http.DetectContentType recognises correctly
	mp3Data := append([]byte{0x49, 0x44, 0x33, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, bytes.Repeat([]byte{0x00}, 1024)...)
	mp4Data := append([]byte{0x00, 0x00, 0x00, 0x18, 0x66, 0x74, 0x79, 0x70, 0x6d, 0x70, 0x34, 0x32, 0x00, 0x00, 0x00, 0x00}, bytes.Repeat([]byte{0x00}, 1024)...)

	t.Run("MP3 upload 200 kind audio with URL", func(t *testing.T) {
		w := uploadBytes(t, r, "song.mp3", mp3Data)
		testutil.AssertStatus(t, w, 200)
		var data map[string]any
		testutil.ParseJSON(t, w, &data)
		if data["kind"] != "audio" {
			t.Fatalf("expected kind audio got %v", data["kind"])
		}
		if _, ok := data["url"].(string); !ok {
			t.Fatalf("missing url")
		}
		if !strings.HasPrefix(data["url"].(string), "/media/") {
			t.Fatalf("url should start with /media/")
		}
		// source defaults to upload, telegram id empty - verify via DB
		var source, tfuid string
		id := int64(data["id"].(float64))
		err := db.DB.QueryRow("SELECT COALESCE(source,''), COALESCE(telegram_file_unique_id,'') FROM uploads WHERE id=?", id).Scan(&source, &tfuid)
		if err != nil {
			t.Fatalf("db query: %v", err)
		}
		if source != "upload" {
			t.Fatalf("expected source upload got %q", source)
		}
		if tfuid != "" {
			t.Fatalf("expected empty telegram_file_unique_id got %q", tfuid)
		}
	})

	t.Run("MP4 upload 200 kind video", func(t *testing.T) {
		// use unique data to avoid dedupe with previous? mp4 already distinct, but ensure unique by appending byte
		uniq := append(mp4Data, byte(0x01))
		w := uploadBytes(t, r, "clip.mp4", uniq)
		testutil.AssertStatus(t, w, 200)
		var data map[string]any
		testutil.ParseJSON(t, w, &data)
		if data["kind"] != "video" {
			t.Fatalf("expected kind video got %v", data["kind"])
		}
	})

	t.Run("audio over 25MB returns 413", func(t *testing.T) {
		over := 25*1024*1024 + 1
		big := make([]byte, over)
		copy(big, mp3Data)
		w := uploadBytes(t, r, "big.mp3", big)
		testutil.AssertStatus(t, w, 413)
	})

	t.Run("video over 50MB returns 413", func(t *testing.T) {
		over := 50*1024*1024 + 1
		big := make([]byte, over)
		copy(big, mp4Data)
		w := uploadBytes(t, r, "big.mp4", big)
		testutil.AssertStatus(t, w, 413)
	})

	t.Run("content-type mismatch mp4 with text bytes returns 400", func(t *testing.T) {
		txt := []byte("hello world this is plain text not video")
		w := uploadBytes(t, r, "fake.mp4", txt)
		testutil.AssertStatus(t, w, 400)
	})

	t.Run("image upload still resizes and thumbnails", func(t *testing.T) {
		pngData := uniqueTestPNG(t)
		w := uploadTestImage(t, r, "/api/upload", "", "", pngData)
		testutil.AssertStatus(t, w, 200)
		var data map[string]any
		testutil.ParseJSON(t, w, &data)
		if data["kind"] != "image" {
			t.Fatalf("expected kind image got %v", data["kind"])
		}
		if thumb, _ := data["thumbnail_url"].(string); thumb == "" {
			t.Fatalf("expected thumbnail_url for image")
		}
	})

	t.Run("source defaults and telegram id empty for web upload", func(t *testing.T) {
		pngData := uniqueTestPNG(t)
		w := uploadTestImage(t, r, "/api/upload", "", "", pngData)
		testutil.AssertStatus(t, w, 200)
		var data map[string]any
		testutil.ParseJSON(t, w, &data)
		id := int64(data["id"].(float64))
		var source, tfuid string
		err := db.DB.QueryRow("SELECT COALESCE(source,''), COALESCE(telegram_file_unique_id,'') FROM uploads WHERE id=?", id).Scan(&source, &tfuid)
		if err != nil {
			t.Fatalf("db query: %v", err)
		}
		if source != "upload" {
			t.Fatalf("expected source upload got %q", source)
		}
		if tfuid != "" {
			t.Fatalf("expected empty telegram_file_unique_id")
		}
	})
}

func TestLoadMediaBytes(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	mediaDir := t.TempDir()
	SetMediaPath(mediaDir)
	defer SetMediaPath("/app/media")
	r := testRouter()

	t.Run("load existing upload", func(t *testing.T) {
		mp3Data := append([]byte{0x49, 0x44, 0x33, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, bytes.Repeat([]byte{0x01}, 512)...)
		w := uploadBytes(t, r, "load.mp3", mp3Data)
		testutil.AssertStatus(t, w, 200)
		var data map[string]any
		testutil.ParseJSON(t, w, &data)
		id := int64(data["id"].(float64))
		mf, err := LoadMediaBytes(id)
		if err != nil {
			t.Fatalf("LoadMediaBytes: %v", err)
		}
		if mf.Kind != "audio" {
			t.Fatalf("expected kind audio got %q", mf.Kind)
		}
		if len(mf.Data) == 0 {
			t.Fatalf("expected non-empty data")
		}
		if mf.Ext != ".mp3" {
			t.Fatalf("expected ext .mp3 got %q", mf.Ext)
		}
	})

	t.Run("missing upload returns error", func(t *testing.T) {
		_, err := LoadMediaBytes(999999)
		if err == nil {
			t.Fatalf("expected error for missing upload")
		}
	})
}
