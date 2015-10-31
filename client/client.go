package registrar

import (
	"bytes"
	"fmt"
	"io"
	"log"

	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/google/go-querystring/query"
)

var debug = true

const DefaultUserAgent = "registrar-go/0.1"

type Client struct {
	client *http.Client

	BaseURL   *url.URL
	UserAgent string

	Accounts *AccountsService
}

func NewClient(baseURL string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		panic(err)
	}

	c := &Client{client: httpClient, BaseURL: u, UserAgent: DefaultUserAgent}
	c.Accounts = &AccountsService{client: c}

	return c
}

func (c *Client) NewRequest(method, urlStr string, body interface{}) (*http.Request, error) {
	rel, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	u := c.BaseURL.ResolveReference(rel)

	var contentType string
	var buf io.ReadWriter
	if body != nil {
		if method == "GET" {
			v, err := query.Values(body)
			if err != nil {
				return nil, err
			}
			u.RawQuery = v.Encode()
		} else {
			buf = new(bytes.Buffer)
			err := json.NewEncoder(buf).Encode(body)
			if err != nil {
				return nil, err
			}
			contentType = "application/json"
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/json")
	if contentType != "" {
		req.Header.Add("Content-Type", contentType)
	}
	if c.UserAgent != "" {
		req.Header.Add("User-Agent", c.UserAgent)
	}
	return req, nil
}

func (c *Client) Call(method, urlStr string, body interface{}, v interface{}) error {
	req, err := c.NewRequest(method, urlStr, body)
	if err != nil {
		return err
	}

	err = c.Do(req, v)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) Do(req *http.Request, v interface{}) error {
	if debug {
		log.Printf("%v %v", req.Method, req.URL.String())
		if req.Body != nil {
			body, err := ioutil.ReadAll(req.Body)
			if err != nil {
				return err
			}
			req.Body = ioutil.NopCloser(bytes.NewBuffer(body))
			log.Println("===== begin request body =====")
			log.Print(string(body))
			log.Println("====== end request body ======")
		}

	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		if debug {
			log.Printf("request failed with status: %s", resp.Status)
		}
		return fmt.Errorf("request failed with status: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if debug {
		log.Println("===== begin response body =====")
		log.Print(string(body))
		log.Println("====== end response body ======")
	}

	if v != nil {
		err = json.Unmarshal(body, v)
	}

	return err
}
