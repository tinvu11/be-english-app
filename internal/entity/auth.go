package entity

// AuthIdentity is the trusted identity extracted from a Firebase ID token.
type AuthIdentity struct {
	UID     string
	Email   string
	Name    string
	Picture string
}
