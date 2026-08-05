package models

type UmamiSettings struct {
	ID                  int    `json:"id"`
	Enabled             bool   `json:"enabled"`
	TrackerHostname     string `json:"tracker_hostname"`
	WebsiteID           string `json:"website_id"`
	ShareData           bool   `json:"share_data"`
	EnableAdminTracking bool   `json:"enable_admin_tracking"`
}
