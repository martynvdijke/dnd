package db

import (
	"log"
)

// CampaignEventSettings holds per-campaign Google Calendar integration settings.
type CampaignEventSettings struct {
	ID               int    `json:"id"`
	CampaignID       int    `json:"campaign_id"`
	Slug             string `json:"slug"`
	DisplayName      string `json:"display_name"`
	CalendarID       string `json:"calendar_id"`
	Tags             string `json:"tags"`
	ColorLabels      string `json:"color_labels"`
	FilterMode       string `json:"filter_mode"`
	CacheTTLSeconds  int    `json:"cache_ttl_seconds"`
	CredentialsJSON  string `json:"credentials_json"`
	AuthMethod       string `json:"auth_method"`
	OAuthClientID    string `json:"oauth_client_id"`
	OAuthClientSecret string `json:"oauth_client_secret"`
	OAuthRefreshToken string `json:"oauth_refresh_token"`
	IsActive         bool   `json:"is_active"`
}

// ListCampaignEventSettings returns all campaign event settings.
func ListCampaignEventSettings() []CampaignEventSettings {
	rows, err := DB.Query(`
		SELECT id, campaign_id, slug, display_name, calendar_id, tags,
			COALESCE(color_labels,''), COALESCE(filter_mode,'text'),
			cache_ttl_seconds, COALESCE(credentials_json,''),
			COALESCE(auth_method,'service_account'), COALESCE(oauth_client_id,''),
			COALESCE(oauth_client_secret,''), COALESCE(oauth_refresh_token,''),
			is_active
		FROM campaign_event_settings ORDER BY display_name`)
	if err != nil {
		log.Printf("campaign_event_settings: list error: %v", err)
		return nil
	}
	defer rows.Close()

	var settings []CampaignEventSettings
	for rows.Next() {
		var s CampaignEventSettings
		if err := rows.Scan(&s.ID, &s.CampaignID, &s.Slug, &s.DisplayName,
			&s.CalendarID, &s.Tags, &s.ColorLabels, &s.FilterMode,
			&s.CacheTTLSeconds, &s.CredentialsJSON, &s.AuthMethod,
			&s.OAuthClientID, &s.OAuthClientSecret, &s.OAuthRefreshToken,
			&s.IsActive); err != nil {
			log.Printf("campaign_event_settings: scan error: %v", err)
			continue
		}
		settings = append(settings, s)
	}
	return settings
}

// GetCampaignEventSettingsBySlug returns a campaign event setting by its slug.
func GetCampaignEventSettingsBySlug(slug string) *CampaignEventSettings {
	var s CampaignEventSettings
	err := DB.QueryRow(`
		SELECT id, campaign_id, slug, display_name, calendar_id, tags,
			COALESCE(color_labels,''), COALESCE(filter_mode,'text'),
			cache_ttl_seconds, COALESCE(credentials_json,''),
			COALESCE(auth_method,'service_account'), COALESCE(oauth_client_id,''),
			COALESCE(oauth_client_secret,''), COALESCE(oauth_refresh_token,''),
			is_active
		FROM campaign_event_settings WHERE slug=?`, slug).
		Scan(&s.ID, &s.CampaignID, &s.Slug, &s.DisplayName,
			&s.CalendarID, &s.Tags, &s.ColorLabels, &s.FilterMode,
			&s.CacheTTLSeconds, &s.CredentialsJSON, &s.AuthMethod,
			&s.OAuthClientID, &s.OAuthClientSecret, &s.OAuthRefreshToken,
			&s.IsActive)
	if err != nil {
		return nil
	}
	return &s
}

// GetCampaignEventSettingsByID returns a campaign event setting by its ID.
func GetCampaignEventSettingsByID(id int) *CampaignEventSettings {
	var s CampaignEventSettings
	err := DB.QueryRow(`
		SELECT id, campaign_id, slug, display_name, calendar_id, tags,
			COALESCE(color_labels,''), COALESCE(filter_mode,'text'),
			cache_ttl_seconds, COALESCE(credentials_json,''),
			COALESCE(auth_method,'service_account'), COALESCE(oauth_client_id,''),
			COALESCE(oauth_client_secret,''), COALESCE(oauth_refresh_token,''),
			is_active
		FROM campaign_event_settings WHERE id=?`, id).
		Scan(&s.ID, &s.CampaignID, &s.Slug, &s.DisplayName,
			&s.CalendarID, &s.Tags, &s.ColorLabels, &s.FilterMode,
			&s.CacheTTLSeconds, &s.CredentialsJSON, &s.AuthMethod,
			&s.OAuthClientID, &s.OAuthClientSecret, &s.OAuthRefreshToken,
			&s.IsActive)
	if err != nil {
		return nil
	}
	return &s
}

// SaveCampaignEventSettings upserts a campaign event setting.
func SaveCampaignEventSettings(s CampaignEventSettings) (int, error) {
	if s.CacheTTLSeconds <= 0 {
		s.CacheTTLSeconds = 300
	}
	if s.AuthMethod == "" {
		s.AuthMethod = "service_account"
	}
	if s.FilterMode == "" {
		s.FilterMode = "text"
	}

	if s.ID > 0 {
		_, err := DB.Exec(`
			UPDATE campaign_event_settings SET
				campaign_id=?, slug=?, display_name=?, calendar_id=?, tags=?,
				color_labels=?, filter_mode=?, cache_ttl_seconds=?,
				credentials_json=?, auth_method=?, oauth_client_id=?,
				oauth_client_secret=?, oauth_refresh_token=?, is_active=?,
				updated_at=datetime('now')
			WHERE id=?`,
			s.CampaignID, s.Slug, s.DisplayName, s.CalendarID, s.Tags,
			s.ColorLabels, s.FilterMode, s.CacheTTLSeconds,
			s.CredentialsJSON, s.AuthMethod, s.OAuthClientID,
			s.OAuthClientSecret, s.OAuthRefreshToken, s.IsActive,
			s.ID)
		if err != nil {
			log.Printf("campaign_event_settings: update error: %v", err)
			return s.ID, err
		}
		return s.ID, nil
	}

	result, err := DB.Exec(`
		INSERT INTO campaign_event_settings(campaign_id, slug, display_name, calendar_id, tags, color_labels, filter_mode, cache_ttl_seconds, credentials_json, auth_method, oauth_client_id, oauth_client_secret, oauth_refresh_token, is_active)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		s.CampaignID, s.Slug, s.DisplayName, s.CalendarID, s.Tags,
		s.ColorLabels, s.FilterMode, s.CacheTTLSeconds,
		s.CredentialsJSON, s.AuthMethod, s.OAuthClientID,
		s.OAuthClientSecret, s.OAuthRefreshToken, s.IsActive)
	if err != nil {
		log.Printf("campaign_event_settings: insert error: %v", err)
		return 0, err
	}
	id, _ := result.LastInsertId()
	return int(id), nil
}

// DeleteCampaignEventSettings deletes a campaign event setting by ID.
func DeleteCampaignEventSettings(id int) error {
	_, err := DB.Exec("DELETE FROM campaign_event_settings WHERE id=?", id)
	if err != nil {
		log.Printf("campaign_event_settings: delete error: %v", err)
	}
	return err
}
