package services

type BlueskyNewSession struct {
	Did             string `json:"did"`
	DidDoc          DidDoc `json:"didDoc"`
	Handle          string `json:"handle"`
	Email           string `json:"email"`
	EmailConfirmed  bool   `json:"emailConfirmed"`
	EmailAuthFactor bool   `json:"emailAuthFactor"`
	AccessJwt       string `json:"accessJwt"`
	RefreshJwt      string `json:"refreshJwt"`
	Active          bool   `json:"active"`
}
type VerificationMethod struct {
	ID                 string `json:"id"`
	Type               string `json:"type"`
	Controller         string `json:"controller"`
	PublicKeyMultibase string `json:"publicKeyMultibase"`
}
type Service struct {
	ID              string `json:"id"`
	Type            string `json:"type"`
	ServiceEndpoint string `json:"serviceEndpoint"`
}
type DidDoc struct {
	Context            []string             `json:"@context"`
	ID                 string               `json:"id"`
	AlsoKnownAs        []string             `json:"alsoKnownAs"`
	VerificationMethod []VerificationMethod `json:"verificationMethod"`
	Service            []Service            `json:"service"`
}

type BlueskyCreatePostResp struct {
	URI              string `json:"uri"`
	Cid              string `json:"cid"`
	Commit           Commit `json:"commit"`
	ValidationStatus string `json:"validationStatus"`
}
type Commit struct {
	Cid string `json:"cid"`
	Rev string `json:"rev"`
}

type BlueskyCreatePostReq struct {
	Repo       string `json:"repo"`
	Collection string `json:"collection"`
	Record     Record `json:"record"`
}
type Record struct {
	Text      string `json:"text"`
	CreatedAt string `json:"createdAt"`
}
