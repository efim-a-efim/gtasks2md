package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/tasks/v1"
)

// Authenticate handles authentication to Google APIs.
// Loads credentials from the given path, environment variable, or default local file.
// Manages token generation and refreshing.
func Authenticate(ctx context.Context, credentialsPath string) (*http.Client, error) {
	tokenPath := "token.json"

	// 1. Determine credentials path if not provided
	if credentialsPath == "" {
		credentialsPath = os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
		if credentialsPath == "" {
			credentialsPath = "credentials.json"
		}
	}

	// 2. Load credentials (client_secret.json equivalent)
	b, err := os.ReadFile(credentialsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("Credentials file not found at '%s'. Please provide a valid OAuth 2.0 Client ID JSON file", credentialsPath)
		}
		return nil, fmt.Errorf("unable to read credentials file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, tasks.TasksScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %v", err)
	}

	// 3. Try to load token from token.json
	tok, err := tokenFromFile(tokenPath)
	if err == nil && (tok.Valid() || tok.RefreshToken != "") {
		// Valid or refreshable token found
		return config.Client(ctx, tok), nil
	}

	// 4. Token not found or invalid, initiate OAuth flow
	tok, err = getTokenFromWeb(config)
	if err != nil {
		return nil, fmt.Errorf("unable to get token from web: %v", err)
	}

	err = saveToken(tokenPath, tok)
	if err != nil {
		fmt.Printf("Warning: unable to cache oauth token: %v\n", err)
	}

	return config.Client(ctx, tok), nil
}

// tokenFromFile retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// getTokenFromWeb requests a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	fmt.Print("Enter authorization code: ")
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, fmt.Errorf("unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve token from web: %v", err)
	}
	return tok, nil
}

// saveToken saves a token to a file path.
func saveToken(path string, token *oauth2.Token) error {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(token)
}
