// Package cookies обрабатывает cookie запросов для приложения
package cookies

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/gorilla/securecookie"
)

var (
	// hash ключ для работы с cookie
	hashKey = securecookie.GenerateRandomKey(64)
	// blockKey - дополнительный ключ для шифрования cookie
	blockKey = securecookie.GenerateRandomKey(32)
	// s - SecureCookie для работы с пользовательскими cookie в приложении
	s = securecookie.New(hashKey, blockKey)
)

// GenerateUserID генерирует значение для ID пользователя
func GenerateUserID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		logrus.Error("UserID for session did not generate")
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

// GetEncodedValue кодирует значение ID пользователя в securecookie
func GetEncodedValue(userID string) (string, error) {
	value := map[string]string{
		"user_id": userID,
	}
	return s.Encode("user", value)
}

// SetUserCookie устанавливает значение ID пользователя в securecookie
func SetUserCookie(w http.ResponseWriter, userID string) {
	encoded, err := GetEncodedValue(userID)
	if err == nil {
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

// GetUserID получает значение ID пользователя из securecookie
func GetUserID(r *http.Request) (string, error) {
	if cookie, err := r.Cookie("user"); err == nil {
		value := make(map[string]string)
		if err = s.Decode("user", cookie.Value, &value); err == nil {
			return value["user_id"], nil
		}
	}
	return "", http.ErrNoCookie
}
