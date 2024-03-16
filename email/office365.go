package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
)

// Client is our type
type Client struct {
	httpClient  *http.Client
	AccessToken *string
}

type bodyType struct {
	ContentType string `json:"contentType"`
	Content     string `json:"content"`
}

type emailAddressType struct {
	Address string `json:"address"`
	Name    string `json:"name"`
}

type recipientType struct {
	EmailAddress emailAddressType `json:"emailAddress"`
}

type messageType struct {
	Subject      string          `json:"subject"`
	Body         bodyType        `json:"body"`
	ToRecipients []recipientType `json:"toRecipients"`
}

type message struct {
	Message messageType `json:"message"`
}

// Office365Client creates a new Microsoft Office365 client
func Office365Client(tenantID, clientID, clientSecret string) (*Client, error) {
	accessToken, err := initializeClient(tenantID, clientID, clientSecret)
	if err != nil {
		log.Printf("%+v", err)
		return nil, err
	}

	client := &Client{
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		AccessToken: accessToken,
	}

	return client, nil
}

func initializeClient(tenantID, clientID, clientSecret string) (*string, error) {
	// create confidential client
	cred, err := confidential.NewCredFromSecret(clientSecret)
	if err != nil {
		log.Printf("%+v", err)
		return nil, err
	}

	tenantUrl, err := url.JoinPath("https://login.microsoftonline.com", tenantID)
	if err != nil {
		log.Printf("%+v", err)
		return nil, err
	}

	confidentialClient, err := confidential.New(tenantUrl, clientID, cred)
	if err != nil {
		log.Printf("%+v", err)
		return nil, err
	}

	// acquire access token
	scopes := []string{"https://graph.microsoft.com/.default"}
	result, err := confidentialClient.AcquireTokenSilent(context.TODO(), scopes)
	if err != nil {
		// cache miss, authenticate
		result, err = confidentialClient.AcquireTokenByCredential(context.TODO(), scopes)
		if err != nil {
			log.Printf("%+v", err)
			return nil, err
		}
	}

	return &result.AccessToken, nil
}

// makeRequest is a helper function to wrap making REST calls to Microsoft Graph API
func (client *Client) makeRequest(method, url string, body io.Reader) ([]byte, error) {
	// create request
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		log.Printf("%+v", err)
		return nil, err
	}
	request.Header.Set("Accept", "application/json")

	// set content-type only on requests that send some content
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	// include auth token if available
	if client.AccessToken != nil {
		log.Println(*client.AccessToken)
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", *client.AccessToken))
	}

	// make request, get response
	var response *http.Response
	response, err = client.httpClient.Do(request)
	if err != nil {
		log.Printf("%+v", err)
		return nil, err
	}
	defer response.Body.Close()

	// error?
	if !(response.StatusCode >= 200 && response.StatusCode <= 299) {
		err = fmt.Errorf("%s call to %s returned status code %d ", method, url, response.StatusCode)
		log.Printf("%+v", err)
		return nil, err
	}

	// get body for caller, if there is something
	var data []byte
	if response.ContentLength != 0 {
		data, err = io.ReadAll(response.Body)
		if err != nil {
			log.Printf("%+v", err)
			return nil, err
		}
	}

	return data, nil
}

func (client *Client) Send(userID, subject, body, to string) error {
	msg := message{
		messageType{
			Subject: subject,
			Body:    bodyType{ContentType: "Text", Content: body},
			ToRecipients: []recipientType{
				{EmailAddress: emailAddressType{Address: to}},
			},
		},
	}

	m, err := json.Marshal(msg)
	if err != nil {
		log.Printf("%+v", err)
		return err
	}

	b, err := client.makeRequest("POST", fmt.Sprintf("https://graph.microsoft.com/v1.0/users/%s/sendMail", userID), bytes.NewReader(m))
	if err != nil {
		log.Printf("%+v", err)
		log.Printf("%+v", string(b))
		return err
	}

	return nil
}
