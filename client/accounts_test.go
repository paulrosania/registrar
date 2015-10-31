package registrar

import (
	"testing"

	"net/http"

	"github.com/jarcoal/httpmock"
)

var transport *httpmock.MockTransport
var client *Client

func init() {
	transport = httpmock.NewMockTransport()
	hc := &http.Client{Transport: transport}
	client = NewClient("http://api.example.com/auth/", hc)
	debug = false
}

type ResponderResult struct {
	Called bool
}

func RegisterResponder(method, url string, inner httpmock.Responder) *ResponderResult {
	result := new(ResponderResult)
	transport.RegisterResponder(method, url,
		func(req *http.Request) (*http.Response, error) {
			result.Called = true
			return inner(req)
		},
	)
	return result
}

func TestAccountsCurrent(t *testing.T) {
	defer transport.Reset()
	endpoint := RegisterResponder("GET", "http://api.example.com/auth/userinfo",
		func(req *http.Request) (*http.Response, error) {
			account := Account{
				Email: "paul@example.com",
			}
			return httpmock.NewJsonResponse(200, account)
		},
	)

	account, err := client.Accounts.Current(nil)
	if err != nil {
		t.Fatal(err)
	}

	if !endpoint.Called {
		t.Fatalf("never called API")
	}

	if account == nil {
		t.Fatalf("expected account, got nil")
	}

	if account.Email != "paul@example.com" {
		t.Errorf("expected email to be \"paul@example.com\", got %q", account.Email)
	}
}
