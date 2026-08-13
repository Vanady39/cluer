package oidcauth

type Claims struct {
	Issuer               string   `json:"iss"`
	Subject              string   `json:"sub"`
	Expiration           int64    `json:"exp"`
	IssuedAt             int64    `json:"iat"`
	AuthTime             int64    `json:"auth_time"`
	SessionID            string   `json:"sid"`
	AuthContextReference string   `json:"acr"`
	Email                string   `json:"email"`
	EmailVerified        bool     `json:"email_verified"`
	Name                 string   `json:"name"`
	GivenName            string   `json:"given_name"`
	Picture              string   `json:"picture"`
	PreferredUsername    string   `json:"preferred_username"`
	Nickname             string   `json:"nickname"`
	Groups               []string `json:"groups"`
}
