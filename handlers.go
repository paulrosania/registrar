package main

import (
	"errors"
	"fmt"
	"log"

	"encoding/json"
	"net/http"

	"github.com/gorilla/context"

	"github.com/paulrosania/go-validation"
	"github.com/paulrosania/registrar/storage"
)

type HandlerFunc func(*Context, http.ResponseWriter) error

func writeError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	log.Print(err)
	switch err := err.(type) {
	case *OAuthError:
		err.Write(w)
	default:
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "{\"error_description\": %q}\n", err)
	}
}

func writeJson(w http.ResponseWriter, o interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	buf, err := json.Marshal(o)
	if err != nil {
		writeError(w, NewOAuthError("internal_server_error", err.Error()))
		return
	}

	fmt.Fprintln(w, string(buf))
}

func readOneFormValue(r *http.Request, key string) (val string, err error) {
	val, err = readOneFormValueOptional(r, key)
	if val == "" && err == nil {
		return "", NewOAuthError("invalid_request", fmt.Sprintf("missing required parameter %q", key))
	}

	return
}

func readOneFormValueOptional(r *http.Request, key string) (val string, err error) {
	vals := r.PostForm[key]

	switch len(vals) {
	case 1:
		return vals[0], nil
	case 0:
		return "", nil
	default:
		return "", NewOAuthError("invalid_request", fmt.Sprintf("multiple parameters for %q", key))
	}
}

// GET /.well-known/openid-configuration
func OpenIdConfigurationHandler(ctx *Context, w http.ResponseWriter) error {
	baseUrl := ctx.Server.config.Server.BaseUrl
	cfg := map[string]interface{}{
		"issuer":                 baseUrl,
		"authorization_endpoint": baseUrl + "/auth",
		"token_endpoint":         baseUrl + "/token",
		"userinfo_endpoint":      baseUrl + "/userinfo",
		"revocation_endpoint":    baseUrl + "/revoke",
		"jwks_uri":               baseUrl + "/certs",
		"scopes_supported": []string{
			"openid",
			"email",
			"profile"},
		"response_types_supported": []string{
			"code",
			"token",
			"id_token",
			"code token",
			"code id_token",
			"code token id_token",
			"none"},
		"id_token_signing_alg_values_supported": []string{"RS256"},
		"token_endpoint_auth_methods_supported": []string{"client_secret_basic"},
		"subject_types_supported":               []string{"public"},
		"grant_types_supported": []string{
			"authorization_code",
			"implicit",
		},
		"claims_supported": []string{
			"aud",
			"email",
			"email_verified",
			"exp",
			"family_name",
			"given_name",
			"iat",
			"iss",
			"locale",
			"name",
			"picture",
			"sub",
		},
	}
	writeJson(w, cfg)
	return nil
}

// GET /userinfo
func UserinfoHandler(ctx *Context, w http.ResponseWriter) error {
	// TODO: return valid userinfo response (sub is required)
	if u, ok := context.GetOk(ctx.Request, CurrentPrincipal); ok {
		writeJson(w, u)
	} else {
		return errors.New("failed to retrieve user")
	}

	return nil
}

func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*") // TODO: this is a security flaw (needs to be fixed to wherever the proxy lives)
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Max-Age", "3600") // (seconds)
}

func OptionsHandler(ctx *Context, w http.ResponseWriter) error {
	setCORSHeaders(w)
	return nil
}

// POST /accounts
func CreateAccountHandler(ctx *Context, w http.ResponseWriter) error {
	setCORSHeaders(w)

	params := new(storage.UserParams)
	err := json.NewDecoder(ctx.Request.Body).Decode(params)
	if err != nil {
		return NewOAuthError("invalid_request", err.Error())
	}

	v := validation.NewMultiValidator()
	v.Assert(params.Email != "", "email", "must provide a email")
	v.Assert(params.Password != "", "password", "must provide a password")
	if v.Valid() {
		user, err := ctx.Server.store.Users.New(params)
		if err == storage.ErrNotUnique {
			err := NewOAuthError("invalid_request", "validation failed")
			err.Meta["fields"] = map[string][]string{
				"email": []string{"email address taken"},
			}
			return err
		} else if err != nil {
			return NewOAuthError("invalid_request", err.Error())
		}
		writeJson(w, user)
	} else {
		err := NewOAuthError("invalid_request", "validation failed")
		err.Meta["fields"] = v.Errors()
		return err
	}

	return nil
}

// GET /client
func ClientHandler(ctx *Context, w http.ResponseWriter) error {
	if c, ok := context.GetOk(ctx.Request, CurrentPrincipal); ok {
		writeJson(w, c)
	} else {
		return errors.New("failed to retrieve client")
	}

	return nil
}

// POST /clients
func NewClientHandler(ctx *Context, w http.ResponseWriter) error {
	r := ctx.Request

	err := r.ParseForm()
	if err != nil {
		return NewOAuthError("invalid_request", err.Error())
	}

	name, err := readOneFormValue(r, "name")
	if err != nil {
		return err
	}

	client, err := ctx.Server.store.Apps.New(name, "", "", "", "secret")
	if err != nil {
		return NewOAuthError("invalid_request", err.Error())
	}

	writeJson(w, client)
	return nil
}

// POST /token
func TokenHandler(ctx *Context, w http.ResponseWriter) error {
	r := ctx.Request

	err := r.ParseForm()
	if err != nil {
		return NewOAuthError("invalid_request", err.Error())
	}

	grantType, err := readOneFormValue(r, "grant_type")
	if err != nil {
		return err
	}

	switch grantType {
	case "authorization_code":
		return authorizationCodeGrantHandler(ctx, w)
	case "client_credentials":
		return clientCredentialsGrantHandler(ctx, w)
	case "password":
		return passwordGrantHandler(ctx, w)
	case "refresh_token":
		return refreshTokenGrantHandler(ctx, w)
	}

	return NewOAuthError("unsupported_grant_type", fmt.Sprintf("unsupported grant type %q", grantType))
}

func authorizationCodeGrantHandler(ctx *Context, w http.ResponseWriter) error {
	r := ctx.Request
	app := context.Get(r, CurrentPrincipal).(*storage.Application)

	code, err := readOneFormValue(r, "code")
	if err != nil {
		return err
	}

	redirectUri, err := readOneFormValue(r, "redirect_uri")
	if err != nil {
		return err
	}

	resp, err := ctx.Server.store.Apps.ExchangeAuthCode(app, code, redirectUri)
	if err != nil {
		return err
	}

	writeJson(w, resp)
	return nil
}

func clientCredentialsGrantHandler(ctx *Context, w http.ResponseWriter) error {
	r := ctx.Request
	app := context.Get(r, CurrentPrincipal).(*storage.Application)

	scope, err := readOneFormValueOptional(r, "scope")
	if err != nil {
		return err
	}

	token, err := ctx.Server.store.Apps.Authorize(app, scope)
	if err != nil {
		return err
	}

	writeJson(w, TokenResponse{
		AccessToken: token.Token,
		TokenType:   token.Type,
		ExpiresIn:   3600,
		Scope:       scope,
	})
	return nil
}

func passwordGrantHandler(ctx *Context, w http.ResponseWriter) error {
	r := ctx.Request
	app := context.Get(r, CurrentPrincipal).(*storage.Application)

	username, err := readOneFormValue(r, "username")
	if err != nil {
		return err
	}

	password, err := readOneFormValue(r, "password")
	if err != nil {
		return err
	}

	scope, err := readOneFormValueOptional(r, "scope")
	if err != nil {
		return err
	}

	user, err := ctx.Server.store.Users.FindByCredentials(username, password)
	if err != nil {
		return NewOAuthError("access_denied", "invalid username/password")
	}

	seconds := 3600
	issuer := ctx.Server.config.OpenID.Issuer
	accessToken := NewJWT(issuer, app.ClientId, user.Email, seconds)
	refreshToken, err := ctx.Server.store.Users.Authorize(user.Id, app.Id, scope, true)
	if err != nil {
		return NewOAuthError("internal_server_error", "could not authorize client")
	}

	signedAccessToken, err := ctx.Server.SignJWT(accessToken)
	if err != nil {
		return NewOAuthError("internal_server_error", "could not authorize client")
	}

	resp := &TokenResponse{
		AccessToken:  signedAccessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    seconds,
		TokenType:    "bearer",
	}

	writeJson(w, resp)
	return nil
}

func refreshTokenGrantHandler(ctx *Context, w http.ResponseWriter) error {
	r := ctx.Request
	app := context.Get(r, CurrentPrincipal).(*storage.Application)

	refreshToken, err := readOneFormValue(r, "refresh_token")
	if err != nil {
		return err
	}

	scope, err := readOneFormValueOptional(r, "scope")
	if err != nil {
		return err
	}

	// TODO: actually need to load up old refresh token, check scope against scope param
	// if scope param is empty, use refresh token scope
	// if scope param includes scopes not granted to refresh token, fail (invalid_request)
	// otherwise, use scope param
	user, err := ctx.Server.store.Users.FindByRefreshToken(app.Id, refreshToken)
	if err != nil {
		log.Printf("refresh failed: %s", err)
		return NewOAuthError("invalid_grant", "refresh token is invalid or expired")
	} else if user == nil {
		log.Println("refresh failed: user not found with matching valid refresh token for specified client")
		return NewOAuthError("invalid_grant", "refresh token is invalid or expired")
	}

	seconds := 3600
	issuer := ctx.Server.config.OpenID.Issuer
	accessToken := NewJWT(issuer, app.ClientId, user.Email, seconds)
	signedAccessToken, err := ctx.Server.SignJWT(accessToken)
	if err != nil {
		return NewOAuthError("internal_server_error", "could not authorize client")
	}

	resp := &TokenResponse{
		AccessToken:  signedAccessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    seconds,
		Scope:        scope, // FIXME: BUG!
		TokenType:    "bearer",
	}

	writeJson(w, resp)
	return nil
}
