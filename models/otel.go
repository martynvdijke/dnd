package models

type OTelSettings struct {
	ID       int    `json:"id"`
	Endpoint string `json:"endpoint"`
	Enabled  bool   `json:"enabled"`
}
