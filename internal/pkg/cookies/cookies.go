package cookies

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/gorilla/securecookie"
)

var (
	hashKey  = securecookie.GenerateRandomKey(64)
	blockKey = securecookie.GenerateRandomKey(32)
	s        = securecookie.New(hashKey, blockKey)
)

func GenerateUserID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func SetUserCookie(w http.ResponseWriter, userID string) {
	value := map[string]string{
		"user_id": userID,
	}
	if encoded, err := s.Encode("user", value); err == nil {
		cookie := &http.Cookie{
			Name:     "user",
			Value:    encoded,
			Path:     "/",
			Expires:  time.Now().Add(24 * time.Hour),
			HttpOnly: true,
		}
		http.SetCookie(w, cookie)
	}
}

func GetUserID(r *http.Request) (string, error) {
	if cookie, err := r.Cookie("user"); err == nil {
		value := make(map[string]string)
		if err = s.Decode("user", cookie.Value, &value); err == nil {
			return value["user_id"], nil
		}
	}
	return "", http.ErrNoCookie
}
