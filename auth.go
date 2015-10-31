package main

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/gorilla/context"
)

type key int

var CurrentPrincipal key = 0

func parseBasicAuth(data64 string) (email, password string, err error) {
	codec := base64.StdEncoding
	data, err := codec.DecodeString(data64)
	if err != nil {
		return
	}

	parts := strings.SplitN(string(data), ":", 2)
	if len(parts) != 2 {
	}
	return parts[0], parts[1], nil
}

type AuthFinder func(*Context, string) (interface{}, error)

func findUserByBearerAuth(ctx *Context, s string) (interface{}, error) {
	token, err := ctx.Server.ParseJWT(s)
	if err != nil {
		return nil, err
	}

	return ctx.Server.store.Users.FindByEmail(token.Claims["sub"].(string))
}

func findUserByBasicAuth(ctx *Context, token string) (user interface{}, err error) {
	email, password, err := parseBasicAuth(token)
	if err != nil {
		return nil, NewOAuthError("invalid_request", "invalid basic credentials")
	}
	return ctx.Server.store.Users.FindByCredentials(email, password)
}

func findClientByBasicAuth(ctx *Context, token string) (client interface{}, err error) {
	id, secret, err := parseBasicAuth(token)
	if err != nil {
		return nil, NewOAuthError("invalid_request", "invalid basic credentials")
	}
	return ctx.Server.store.Apps.FindByCredentials(id, secret)
}

func detectBasicUser(handler HandlerFunc) HandlerFunc {
	finders := map[string]AuthFinder{
		"basic": findUserByBasicAuth,
	}

	return detectAuthWithFinders(handler, finders)
}

func detectBearerUser(handler HandlerFunc) HandlerFunc {
	finders := map[string]AuthFinder{
		"bearer": findUserByBearerAuth,
	}

	return detectAuthWithFinders(handler, finders)
}

func detectUser(handler HandlerFunc) HandlerFunc {
	finders := map[string]AuthFinder{
		"bearer": findUserByBearerAuth,
		"basic":  findUserByBasicAuth,
	}

	return detectAuthWithFinders(handler, finders)
}

func detectClient(handler HandlerFunc) HandlerFunc {
	finders := map[string]AuthFinder{
		"basic": findClientByBasicAuth,
	}

	// TODO: Should be an error if both types of credentials are provided
	return detectAuthWithFinders(detectInlineClientAuth(handler), finders)
}

func parseInlineClientAuth(r *http.Request) (clientId, clientSecret string, err error) {
	err = r.ParseForm()
	if err != nil {
		return
	}

	clientId, err = readOneFormValue(r, "client_id")
	if err != nil {
		return
	}

	clientSecret, err = readOneFormValue(r, "client_secret")
	if err != nil {
		return
	}

	return
}

func detectInlineClientAuth(handler HandlerFunc) HandlerFunc {
	return HandlerFunc(func(ctx *Context, w http.ResponseWriter) error {
		r := ctx.Request
		if _, ok := context.GetOk(r, CurrentPrincipal); !ok {
			clientId, clientSecret, err := parseInlineClientAuth(r)

			if err == nil {
				client, err := ctx.Server.store.Apps.FindByCredentials(clientId, clientSecret)
				if err == nil {
					context.Set(r, CurrentPrincipal, client)
				}
			}
		}

		defer context.Clear(r)
		return handler(ctx, w)
	})
}

func detectAuthWithFinders(handler HandlerFunc, finders map[string]AuthFinder) HandlerFunc {
	return HandlerFunc(func(ctx *Context, w http.ResponseWriter) error {
		r := ctx.Request
		if _, ok := context.GetOk(r, CurrentPrincipal); !ok {
			auths := r.Header["Authorization"]
			if len(auths) > 1 {
				return NewOAuthError("invalid_request", "multiple authorization headers")
			} else if len(auths) == 1 {
				auth := auths[0]
				parts := strings.SplitN(auth, " ", 2)
				if len(parts) > 2 {
					return NewOAuthError("invalid_request", "invalid authorization header format")
				} else if len(parts) == 2 {
					scheme := strings.ToLower(parts[0])
					token := parts[1]

					finder := finders[scheme]
					if finder == nil {
						return NewOAuthError("invalid_request", "unsupported authorization scheme")
					}
					principal, err := finder(ctx, token)

					// you asked for auth, you get booted if you fail
					if err != nil {
						switch e := err.(type) {
						case *OAuthError:
							return e
						default:
							return NewOAuthError("access_denied", "invalid credentials")
						}
						return err
					}

					if principal != nil {
						context.Set(r, CurrentPrincipal, principal)
					}
				}
			}
		}

		defer context.Clear(r)
		return handler(ctx, w)
	})
}

// Doesn't check auth type since CurrentPrincipal can only be of a detected type
// anyway. If you wrap your handler in a detector for the wrong type, things
// will break.
func requireAuth(handler HandlerFunc) HandlerFunc {
	return HandlerFunc(func(ctx *Context, w http.ResponseWriter) error {
		r := ctx.Request
		if _, ok := context.GetOk(r, CurrentPrincipal); !ok {
			return NewOAuthError("access_denied", "authentication required")
		}

		return handler(ctx, w)
	})
}
