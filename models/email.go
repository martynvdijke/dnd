package models

type EmailSettings struct {
	ID       int    `json:"id"`
	SMTPHost string `json:"smtp_host"`
	SMTPPort int    `json:"smtp_port"`
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
	FromAddr string `json:"from_addr"`
	Enabled  bool   `json:"enabled"`
}
