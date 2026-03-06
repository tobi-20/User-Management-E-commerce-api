package helpers

import "time"

var RefreshTokenExpiry time.Time = time.Now().Add(7 * 24 * time.Hour)
