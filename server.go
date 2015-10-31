package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"crypto/rsa"
	"io/ioutil"
	"net/http"

	"github.com/BurntSushi/toml"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"

	"github.com/paulrosania/registrar/storage"
)

type Server struct {
	http.Server
	router *mux.Router

	config *Config
	store  *storage.Client

	verificationKey *rsa.PublicKey
	signingKey      *rsa.PrivateKey
}

func NewServer(cfg *Config) *Server {
	r := mux.NewRouter()

	store, err := storage.NewClient(&storage.Config{
		Database: cfg.Database,
	})
	if err != nil {
		panic(err)
	}

	s := &Server{
		Server: http.Server{
			Addr:    cfg.Server.Bind,
			Handler: r,
		},
		router: r,
		config: cfg,
		store:  store,
	}

	err = s.loadSigningKey()
	if err != nil {
		panic(err)
	}

	err = s.loadVerificationKey()
	if err != nil {
		panic(err)
	}

	s.loadRoutes()

	return s
}

func (s *Server) loadSigningKey() error {
	buf, err := ioutil.ReadFile(s.config.JWT.PrivateKey)
	if err != nil {
		return err
	}

	s.signingKey, err = jwt.ParseRSAPrivateKeyFromPEM(buf)
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) loadVerificationKey() error {
	buf, err := ioutil.ReadFile(s.config.JWT.PublicKey)
	if err != nil {
		return err
	}

	s.verificationKey, err = jwt.ParseRSAPublicKeyFromPEM(buf)
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) loadRoutes() {
	// Logged-out endpoints
	s.handleFunc("/accounts", CreateAccountHandler).Methods("POST")
	s.handleFunc("/accounts", OptionsHandler).Methods("OPTIONS")
	s.handleFunc("/clients", NewClientHandler).Methods("POST")
	s.handleFunc("/.well-known/openid-configuration", OpenIdConfigurationHandler).Methods("GET")

	// Logged-in client endpoints
	s.handleFunc("/client", detectClient(requireAuth(ClientHandler))).Methods("GET")
	s.handleFunc("/token", detectClient(requireAuth(TokenHandler))).Methods("POST")

	// Logged-in user endpoints
	s.handleFunc("/authorize", detectUser(requireAuth(UserinfoHandler))).Methods("GET")
	s.handleFunc("/authorize", detectUser(requireAuth(UserinfoHandler))).Methods("POST")
	s.handleFunc("/userinfo", detectUser(requireAuth(UserinfoHandler))).Methods("GET")
	s.handleFunc("/logout", detectUser(requireAuth(UserinfoHandler))).Methods("POST")
}

func (s *Server) handleFunc(path string, f HandlerFunc) *mux.Route {
	r := s.router

	return r.HandleFunc(path, func(w http.ResponseWriter, req *http.Request) {
		log.Printf("[recv] %s %s [%s]", req.Method, req.URL.Path, req.RemoteAddr)

		ctx := NewContext(context.Background(), s, req)

		if err := f(ctx, w); err != nil {
			if oe, ok := err.(*OAuthError); ok {
				log.Print("Return error: ", oe.Error())
				oe.Write(w)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
	})
}

func (s *Server) SignJWT(token *jwt.Token) (string, error) {
	return token.SignedString(s.signingKey)
}

func (s *Server) ParseJWT(data string) (*jwt.Token, error) {
	return jwt.Parse(data, func(token *jwt.Token) (interface{}, error) {
		// validate "alg" is what we expect:
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return s.verificationKey, nil
	})
}

var configFilePath = flag.String("c", "/etc/registrar/registrar.ini", "path to config file")

func main() {
	flag.Parse()
	var cfg Config
	_, err := toml.DecodeFile(*configFilePath, &cfg)
	if err != nil {
		log.Fatal("Failed reading config file:", err)
	}

	if cfg.Log.Path != "" {
		lf, err := os.Create(cfg.Log.Path)
		if err != nil {
			log.Fatal("Failed opening log file:", err)
		}
		log.SetOutput(lf)
	}

	s := NewServer(&cfg)
	defer s.store.Close()

	log.Println("Registrar server listening on", cfg.Server.Bind)
	err = s.ListenAndServe()
	if err != nil {
		log.Fatalf("Failed to start: %s", err)
	}
}
