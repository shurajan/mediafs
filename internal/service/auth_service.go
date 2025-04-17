package service

import (
	"encoding/json"
	"os"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthData struct {
	PasswordHash string    `json:"password_hash"`
	Token        string    `json:"token"`
	LastAuthTime time.Time `json:"last_auth_time"`
}

type AuthService struct {
	path string
	data *AuthData
}

func NewAuthService(path string) *AuthService {
	return &AuthService{
		path: path,
		data: &AuthData{},
	}
}

func (a *AuthService) Load() error {
	content, err := os.ReadFile(a.path)
	if err != nil {
		return err
	}
	return json.Unmarshal(content, a.data)
}

func (a *AuthService) Save() error {
	bytes, err := json.MarshalIndent(a.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(a.path, bytes, 0600)
}

func (a *AuthService) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	a.data.PasswordHash = string(hash)
	a.data.Token = ""
	a.data.LastAuthTime = time.Time{}
	return a.Save()
}

func (a *AuthService) CheckPassword(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(a.data.PasswordHash), []byte(password)) == nil
}

func (a *AuthService) GenerateToken() string {
	token := uuid.NewString()
	a.data.Token = token
	a.data.LastAuthTime = time.Now()
	a.Save()
	return token
}

func (a *AuthService) CheckToken(token string) bool {
	return token == a.data.Token
}
