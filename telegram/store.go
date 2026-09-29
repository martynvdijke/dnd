package telegram

import (
	"crypto/sha256"
	"encoding/hex"
	"villum/db"
)

func hashCode(code string) string {
	h := sha256.Sum256([]byte(code))
	return hex.EncodeToString(h[:])
}
func InsertLinkCode(userID int64, hash, expiresAt string) error {
	_, err := db.DB.Exec("INSERT INTO telegram_link_codes(user_id,code_hash,expires_at) VALUES(?,?,?)", userID, hash, expiresAt)
	return err
}
func ConsumeLinkCode(hash string) (int64, bool) {
	res, err := db.DB.Exec("UPDATE telegram_link_codes SET used_at=datetime('now') WHERE code_hash=? AND used_at IS NULL AND datetime('now') <= expires_at", hash)
	if err != nil {
		return 0, false
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		return 0, false
	}
	var userID int64
	if err := db.DB.QueryRow("SELECT user_id FROM telegram_link_codes WHERE code_hash=?", hash).Scan(&userID); err != nil {
		return 0, false
	}
	return userID, true
}
func LookupIdentityByTelegramID(tgID int64) (int64, string, int, bool) {
	var uid int64
	var username string
	var dm int
	err := db.DB.QueryRow("SELECT user_id,telegram_username,dm_enabled FROM telegram_identities WHERE telegram_user_id=?", tgID).Scan(&uid, &username, &dm)
	if err != nil {
		return 0, "", 0, false
	}
	return uid, username, dm, true
}
func UpsertIdentity(userID, tgUserID, chatID int64, username string) error {
	_, err := db.DB.Exec(`INSERT INTO telegram_identities(user_id,telegram_user_id,telegram_chat_id,telegram_username) VALUES(?,?,?,?)
		ON CONFLICT(telegram_user_id) DO UPDATE SET user_id=excluded.user_id, telegram_chat_id=excluded.telegram_chat_id, telegram_username=excluded.telegram_username`, userID, tgUserID, chatID, username)
	return err
}
func DeleteIdentityByUserID(userID int64) error {
	_, err := db.DB.Exec("DELETE FROM telegram_identities WHERE user_id=?", userID)
	return err
}
func DeleteIdentityByTelegramID(tgID int64) error {
	_, err := db.DB.Exec("DELETE FROM telegram_identities WHERE telegram_user_id=?", tgID)
	return err
}
func GetIdentityByUserID(userID int64) (int64, string, int, bool) {
	var tgID int64
	var username string
	var dm int
	err := db.DB.QueryRow("SELECT telegram_user_id,telegram_username,dm_enabled FROM telegram_identities WHERE user_id=?", userID).Scan(&tgID, &username, &dm)
	if err != nil {
		return 0, "", 0, false
	}
	return tgID, username, dm, true
}
func SetDMEnabled(userID int64, enabled bool) error {
	v := 0
	if enabled {
		v = 1
	}
	_, err := db.DB.Exec("UPDATE telegram_identities SET dm_enabled=? WHERE user_id=?", v, userID)
	return err
}
