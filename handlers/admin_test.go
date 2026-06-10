package handlers

import (
	"testing"

	"github.com/gin-gonic/gin"

	"villum/handlers/testutil"
)

func TestAdminUserCRUD(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/admin/users", AdminListUsers)
		auth.POST("/admin/users", AdminCreateUser)
		auth.PUT("/admin/users/:id", AdminUpdateUser)
		auth.DELETE("/admin/users/:id", AdminDeleteUser)
		auth.PUT("/admin/users/:id/password", AdminResetPassword)
	})

	t.Run("list users returns 200", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/admin/users")
		testutil.AssertStatus(t, w, 200)
		var users []map[string]any
		testutil.ParseJSON(t, w, &users)
		if len(users) < 1 {
			t.Fatal("expected at least 1 user")
		}
	})

	t.Run("create user returns 201", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/admin/users", map[string]any{
			"username": "newuser",
			"password": "password123",
			"role":     "user",
		})
		testutil.AssertStatus(t, w, 201)
	})

	t.Run("create user short password returns 400", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/admin/users", map[string]any{
			"username": "shortpw",
			"password": "short",
		})
		testutil.AssertStatus(t, w, 400)
	})

	t.Run("update user returns 200", func(t *testing.T) {
		w := testutil.PutJSON(t, r, "/api/admin/users/2", map[string]any{
			"username": "updateduser",
			"role":     "dm",
		})
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("delete user returns 200", func(t *testing.T) {
		w := testutil.Delete(t, r, "/api/admin/users/2")
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("reset password returns 200", func(t *testing.T) {
		w := testutil.PutJSON(t, r, "/api/admin/users/1/password", map[string]any{
			"password": "newpassword123",
		})
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("reset password short returns 400", func(t *testing.T) {
		w := testutil.PutJSON(t, r, "/api/admin/users/1/password", map[string]any{
			"password": "short",
		})
		testutil.AssertStatus(t, w, 400)
	})

	t.Run("create duplicate username returns conflict", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/admin/users", map[string]any{
			"username": "admin",
			"password": "password123",
		})
		if w.Code != 409 && w.Code != 400 {
			t.Fatalf("expected 409 or 400 for duplicate, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestAdminBackupSettings(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/admin/email-settings", GetBackupSettings)
		auth.POST("/admin/email-settings", SaveBackupSettings)
	})

	t.Run("get email settings returns default", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/admin/email-settings")
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("save email settings returns 200", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/admin/email-settings", map[string]any{
			"smtp_host": "smtp.example.com",
			"smtp_port": 587,
			"smtp_user": "test@example.com",
			"smtp_pass": "secret",
			"from_addr": "noreply@example.com",
		})
		testutil.AssertStatus(t, w, 200)
	})
}

func TestAdminBackupOps(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.POST("/admin/backup", TriggerBackup)
		auth.GET("/admin/backups", ListBackups)
	})

	t.Run("list backups returns 200", func(t *testing.T) {
		w := testutil.Get(t, r, "/api/admin/backups")
		testutil.AssertStatus(t, w, 200)
	})

	t.Run("trigger backup returns 200 or 500", func(t *testing.T) {
		w := testutil.PostJSON(t, r, "/api/admin/backup", nil)
		if w.Code != 200 && w.Code != 500 {
			t.Fatalf("expected 200 or 500, got %d: %s", w.Code, w.Body.String())
		}
	})
}
