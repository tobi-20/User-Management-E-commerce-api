package globals

import "time"

const SitePath = "http://localhost:8080"

var RefreshTokenExpiry = 7 * 24 * time.Hour
var MaxVerifyTime = 15 * time.Minute
var	MaxLifetime = 30 * 24 * time.Hour

type EmailSendParams struct {
	Path      string
	Selector  string
	Verifier  string
	EmailAddr string
}
