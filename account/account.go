// Package account owns whole-account operations: closing an account and
// revoking every credential that could still act as the user. The web form,
// the JSON API, and the admin user manager all call this one implementation so
// their behaviour cannot drift.
package account

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"

	"villum/db"
)

var (
	// ErrLastAdmin means closing the account would leave the instance without
	// an administrator able to manage it.
	ErrLastAdmin = errors.New("the last active admin cannot be deleted")
	// ErrNotFound means the requested account does not exist (or was already
	// closed and anonymised).
	ErrNotFound = errors.New("account not found")
	// ErrInvalid means the confirmation token was missing or malformed.
	ErrInvalid = errors.New("invalid account deletion request")
)

// Delete closes the account identified by userID in one transaction. The user
// row survives as an anonymous marker because characters, campaigns and shared
// content reference it; every direct identifier and credential is removed.
//
// The last admin is refused. The guard is part of the UPDATE statement so two
// concurrent admin deletions cannot both pass a separate count check.
func Delete(ctx context.Context, userID int64) error {
	if userID <= 0 {
		return ErrInvalid
	}

	tx, err := db.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var oldEmail string
	if err := tx.QueryRowContext(ctx,
		`SELECT COALESCE(email, '') FROM users WHERE id = ?`,
		userID,
	).Scan(&oldEmail); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	marker := "deleted-" + strconv.FormatInt(userID, 10)
	res, err := tx.ExecContext(ctx, `
		UPDATE users
		SET username = ?, password = '', display_name = 'Deleted user',
		    email = '', role = 'user', oidc_sub = NULL
		WHERE id = ?
		  AND (role <> 'admin'
		       OR (SELECT COUNT(*) FROM users WHERE role = 'admin') > 1)`,
		marker, userID)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		var role string
		if err := tx.QueryRowContext(ctx, `SELECT role FROM users WHERE id = ?`, userID).Scan(&role); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ErrNotFound
			}
			return err
		}
		if role == "admin" {
			return ErrLastAdmin
		}
		return ErrNotFound
	}

	// Telegram identifiers are keyed by Telegram's own user id in some tables,
	// so collect them before the identity rows disappear.
	tgIDs, err := queryInt64s(ctx, tx, `SELECT telegram_user_id FROM telegram_identities WHERE user_id = ?`, userID)
	if err != nil {
		return err
	}

	// Credentials and identity links. Deleting these first makes every other
	// step idempotent from the user's point of view: no device can still sign
	// in while the rest of the cleanup runs.
	statements := []struct {
		query string
		args  []any
	}{
		{`DELETE FROM auth_sessions WHERE user_id = ?`, []any{userID}},
		{`DELETE FROM api_tokens WHERE user_id = ?`, []any{userID}},
		{`DELETE FROM password_reset_tokens WHERE user_id = ?`, []any{userID}},
		{`DELETE FROM push_subscriptions WHERE user_id = ?`, []any{userID}},
		{`DELETE FROM push_mutes WHERE user_id = ?`, []any{userID}},
		{`DELETE FROM telegram_identities WHERE user_id = ?`, []any{userID}},
		{`DELETE FROM telegram_link_codes WHERE user_id = ?`, []any{userID}},
		{`DELETE FROM telegram_digest_subscriptions WHERE user_id = ?`, []any{userID}},
		{`DELETE FROM telegram_auth_tokens WHERE email = ?`, []any{oldEmail}},
		{`UPDATE campaign_telegram_settings SET bound_by_user_id = NULL WHERE bound_by_user_id = ?`, []any{userID}},
		// Membership in other people's campaigns, and personal working data
		// that has no value to anyone else.
		{`DELETE FROM campaign_members WHERE user_id = ?`, []any{userID}},
		{`DELETE FROM share_links WHERE created_by = ?`, []any{userID}},
		{`DELETE FROM ai_draft_sessions WHERE user_id = ?`, []any{userID}},
		{`DELETE FROM copilot_conversations WHERE user_id = ?`, []any{userID}},
		{`DELETE FROM dice_rolls WHERE user_id = ?`, []any{userID}},
		{`DELETE FROM session_rsvps WHERE user_id = ?`, []any{userID}},
		{`DELETE FROM external_import_logs WHERE user_id = ?`, []any{userID}},
		{`DELETE FROM transfer_import_logs WHERE user_id = ?`, []any{userID}},
		{`DELETE FROM compendium_import_logs WHERE user_id = ?`, []any{userID}},
		{`DELETE FROM ink_captures WHERE owner_user_id = ?`, []any{userID}},
		{`DELETE FROM campaign_invitations WHERE invited_by = ?`, []any{userID}},
		{`UPDATE campaign_invitations SET accepted_by = NULL WHERE accepted_by = ?`, []any{userID}},
	}
	for _, stmt := range statements {
		if _, err := tx.ExecContext(ctx, stmt.query, stmt.args...); err != nil {
			return err
		}
	}
	for _, tgID := range tgIDs {
		for _, q := range []string{
			`DELETE FROM telegram_character_claims WHERE telegram_user_id = ?`,
			`DELETE FROM telegram_pending_captures WHERE telegram_user_id = ?`,
			`DELETE FROM telegram_auth_tokens WHERE tg_user_id = ?`,
		} {
			if _, err := tx.ExecContext(ctx, q, tgID); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func queryInt64s(ctx context.Context, tx *sql.Tx, query string, args ...any) ([]int64, error) {
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []int64
	for rows.Next() {
		var v int64
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

// NormaliseConfirmation trims the typed confirmation token. Callers compare it
// with the account's current username before calling Delete.
func NormaliseConfirmation(s string) string {
	return strings.TrimSpace(s)
}
