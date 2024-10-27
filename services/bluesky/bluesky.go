package bluesky

// Import resty into your code and refer it as `resty`.
import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/joho/godotenv"
)

// Import resty into your code and refer it as `resty`.
func CreateNewSession() (NewSession, error) {
	env, envLoadErr := godotenv.Read(".env")
	if envLoadErr != nil {
		return NewSession{}, fmt.Errorf("failed to load environment variables to create session: %w", envLoadErr)
	}

	bskyPassword := env["BSKY_PASSWORD"]
	bskyHandle := env["BSKY_HANDLE"]

	client := resty.New().SetBaseURL("https://bsky.social")
	req := client.R().SetBody(map[string]string{"identifier": bskyHandle, "password": bskyPassword})

	var bskyResp NewSession

	resp, respErr := req.Post("xrpc/com.atproto.server.createSession")
	if respErr != nil {
		return bskyResp, fmt.Errorf("failed to create Bluesky session: %w", respErr)
	}

	unmarshalErr := json.Unmarshal(resp.Body(), &bskyResp)
	if unmarshalErr != nil {
		return bskyResp, fmt.Errorf("failed to unmarshal Bluesky response while creating session: %w", unmarshalErr)
	}

	if reflect.ValueOf(bskyResp).IsZero() {
		return bskyResp, fmt.Errorf("received a json response in an unknown format while creating session")
	}

	return bskyResp, nil
}

func SendPost(ctx context.Context, params PostParams) error {
	session, ok := ctx.Value("session").(NewSession)
	if !ok {
		var sessionCreateErr error
		session, sessionCreateErr = CreateNewSession()
		if sessionCreateErr != nil {
			return fmt.Errorf("failed to create new post: %w", sessionCreateErr)
		}
	}

	client := resty.New().SetAuthScheme("Bearer").SetBaseURL("https://bsky.social")
	client.SetAuthToken(session.AccessJwt)

	req := client.R().SetBody(createPostBody(params))

	resp, respErr := req.Post("xrpc/com.atproto.repo.createRecord")
	if respErr != nil {
		return fmt.Errorf("failed to make request to create new post: %w", respErr)
	}

	if resp.IsError() {
		return fmt.Errorf("received an error response while trying to create a new post: %v", string(resp.Body()))
	}

	var createPostResp RespCreatePost
	unmarshalErr := json.Unmarshal(resp.Body(), &createPostResp)
	if unmarshalErr != nil {
		return fmt.Errorf("received an invalid response while trying to create post: %w", unmarshalErr)
	}

	if reflect.ValueOf(createPostResp).IsZero() {
		return fmt.Errorf("received a json response in invalid format while creating post")
	}

	return nil
}

func createPostBody(params PostParams) ReqCreatePost {
	post := ReqCreatePost{
		Repo:       "rickrollrumble.bsky.social",
		Collection: "app.bsky.feed.post",
		Record: Record{
			Text:      fmt.Sprintf("%s %s", params.Text, params.Link),
			CreatedAt: time.Now().Format(time.RFC3339),
		},
	}

	if params.Link != "" {
		post.Record.Facets = append(post.Record.Facets, Facet{
			Index: Index{
				ByteStart: len(params.Text),
				ByteEnd:   len(params.Text) + len(params.Link) + 1,
			},
			Features: []Features{
				{
					Type: "app.bsky.richtext.facet#link",
					URI:  params.Link,
				},
			},
		})
	}

	if len(params.Images) > 0 {
		post.Record.Embed.Type = "app.bsky.embed.images"
		post.Record.Embed.Images = params.Images
	}

	y, _ := json.Marshal(post)
	j := string(y)
	fmt.Println(j)
	return post
}

func UploadImage(ctx context.Context, image []byte) (RespImageUpload, error) {
	session, ok := ctx.Value("session").(NewSession)
	if !ok {
		var sessionCreateErr error
		session, sessionCreateErr = CreateNewSession()
		if sessionCreateErr != nil {
			return RespImageUpload{}, fmt.Errorf("failed to create new post: %w", sessionCreateErr)
		}
	}

	client := resty.New().SetAuthScheme("Bearer").SetBaseURL("https://bsky.social")
	client.SetAuthToken(session.AccessJwt)

	req := client.R().SetBody(image).SetHeader("Content-Type", "image/png")

	resp, respErr := req.Post("xrpc/com.atproto.repo.uploadBlob")
	if respErr != nil {
		return RespImageUpload{}, fmt.Errorf("failed to upload image: %w", respErr)
	}

	if resp.IsError() {
		return RespImageUpload{}, fmt.Errorf("received an error response while trying to upload image")
	}

	baseResp := make(map[string]RespImageUpload)
	unmarshalErr := json.Unmarshal(resp.Body(), &baseResp)
	if unmarshalErr != nil {
		return RespImageUpload{}, fmt.Errorf("response received from image upload was not valid: %w", unmarshalErr)
	}

	if reflect.ValueOf(baseResp["blob"]).IsZero() {
		return RespImageUpload{}, fmt.Errorf("received an invalid response while trying to upload image")
	}

	return baseResp["blob"], nil
}
