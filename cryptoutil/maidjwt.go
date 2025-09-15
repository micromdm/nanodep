package cryptoutil

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const maidJWTServiceType = "com.apple.maid"

// NewMAIDJWT generates a new signed JWT using key and the claims server_uuid, iat, and jti.
// The claims may not be empty.
// The key should be the private key of the DEP name/MDM server.
// The server_uuid comes from the UUID of the MDM server/DEP name and
// can be acquired using the "AccountDetails" DEP API call.
func NewMAIDJWT(key interface{}, server_uuid string, iat time.Time, jti string) (string, error) {
	if server_uuid == "" || iat.IsZero() || jti == "" {
		return "", errors.New("empty claim parameter(s)")
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss":          server_uuid,
		"iat":          iat.Unix(),
		"jti":          jti,
		"service_type": maidJWTServiceType,
	})
	return token.SignedString(key)
}
