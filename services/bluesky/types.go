package bluesky

type NewSession struct {
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

type RespCreatePost struct {
	URI              string `json:"uri"`
	Cid              string `json:"cid"`
	Commit           Commit `json:"commit"`
	ValidationStatus string `json:"validationStatus"`
}
type Commit struct {
	Cid string `json:"cid"`
	Rev string `json:"rev"`
}

type ReqCreatePost struct {
	Repo       string `json:"repo"`
	Collection string `json:"collection"`
	Record     Record `json:"record"`
}
type Index struct {
	ByteStart int `json:"byteStart"`
	ByteEnd   int `json:"byteEnd"`
}
type Features struct {
	Type string `json:"$type"`
	URI  string `json:"uri"`
}
type Facet struct {
	Index    Index      `json:"index"`
	Features []Features `json:"features"`
}

type ImageDetails struct {
	Alt   string          `json:"alt"`
	Image RespImageUpload `json:"image"`
}

type Embed struct {
	Type   string         `json:"$type"`
	Images []ImageDetails `json:"images"`
}

type Record struct {
	Text      string  `json:"text"`
	CreatedAt string  `json:"createdAt"`
	Facets    []Facet `json:"facets,omitempty"`
	Embed     Embed   `json:"embed,omitempty"`
}

type PostParams struct {
	Text   string
	Link   string
	Images []ImageDetails
}

type RespImageUpload struct {
	Type     string `json:"$type"`
	Ref      Ref    `json:"ref"`
	MimeType string `json:"mimeType"`
	Size     int    `json:"size"`
}
type Ref struct {
	Link string `json:"$link"`
}
