package services

// Import resty into your code and refer it as `resty`.
import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"
)

// Import resty into your code and refer it as `resty`.
func CreateNewSession() (BlueskyNewSession, error) {
	client := resty.New().SetBaseURL("https://bsky.social")
	req := client.R().SetBody(map[string]string{"identifier": "rickrollrumble.bsky.social", "password": "L!V48r$#1MCGWer4"})

	var bskyResp BlueskyNewSession

	resp, respErr := req.Post("xrpc/com.atproto.server.createSession")
	if respErr != nil {
		return bskyResp, fmt.Errorf("failed to create Bluesky session: %w", respErr)
	}

	unmarshalErr := json.Unmarshal(resp.Body(), &bskyResp)
	if unmarshalErr != nil {
		return bskyResp, fmt.Errorf("failed to unmarshal Bluesky response while creating session: %w", unmarshalErr)
	}

	return bskyResp, nil
}

func SendPost(ctx context.Context, body interface{}) error {
	session, ok := ctx.Value("session").(BlueskyNewSession)
	if !ok {
		var sessionCreateErr error
		session, sessionCreateErr = CreateNewSession()
		if sessionCreateErr != nil {
			return fmt.Errorf("failed to create new post: %w", sessionCreateErr)
		}
	}

	client := resty.New().SetAuthScheme("Bearer")
	client.SetAuthToken(session.AccessJwt)

	req := client.R().SetBody(body)
	resp, respErr := req.Post("xrpc/com.atproto.repo.createRecord")
	if respErr != nil {
		return fmt.Errorf("failed to make request to create new post: %w", respErr)
	}

	if resp.IsError() {
		return fmt.Errorf("received an error response while trying to create a new post: %v", string(resp.Body()))
	}

	var createPostResp BlueskyCreatePostResp
	unmarshalErr := json.Unmarshal(resp.Body(), &createPostResp)
	if unmarshalErr != nil {
		return fmt.Errorf("received an invalid response while trying to create post: %w", unmarshalErr)
	}

	return nil
}
