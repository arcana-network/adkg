package common

type VerifierParams struct {
	ClientID string
	Domain   string
}

type GenericVerifierData struct {
	Token    string `json:"id_token"`
	Provider string `json:"provider"`
	UserID   string `json:"user_id"`
	AppID    string `json:"app_id"`
}
