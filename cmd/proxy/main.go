package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/spf13/viper"
	"github.com/sulavmhrzn/gatekeeper/internal/middleware"
)

type Route struct {
	Path   string `mapstructure:"path"`
	Target string `mapstructure:"target"`
}
type Config struct {
	Server struct {
		Port      int    `mapstructure:"port"`
		SecretKey string `mapstructure:"secret_key"`
	} `mapstructure:"server"`
	Routes []Route `mapstructure:"routes"`
}

func LoadConfig() (Config, error) {
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	err = viper.Unmarshal(&cfg)
	return cfg, err
}

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("Could not load config: %v", err)
	}

	proxies := make(map[string]*httputil.ReverseProxy)

	for _, route := range cfg.Routes {
		targetURL, _ := url.Parse(route.Target)
		p := httputil.NewSingleHostReverseProxy(targetURL)
		routePath := route.Path
		originDirector := p.Director
		p.Director = func(r *http.Request) {
			originDirector(r)
			r.URL.Path = strings.TrimPrefix(r.URL.Path, routePath)
			if r.URL.Path == "" {
				r.URL.Path = "/"
			}
		}
		p.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("Backend (%s) error: %v", targetURL, err)
			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte("Gatekeeper Error: Backend is unreachable"))
		}
		proxies[route.Path] = p
		log.Printf("Route mapped: %s -> %s", route.Path, route.Target)
	}

	router := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for path, proxy := range proxies {

			if len(r.URL.Path) >= len(path) && r.URL.Path[:len(path)] == path {
				proxy.ServeHTTP(w, r)
				return
			}
		}
		http.Error(w, "Route not found", http.StatusNotFound)
	})

	handler := middleware.RecoveryMiddleware(
		middleware.LoggingMiddleware(
			middleware.AuthMiddleware(router, cfg.Server.SecretKey),
		),
	)

	log.Printf("Gatekeeper listening on :%d...", cfg.Server.Port)
	if err := http.ListenAndServe(":8000", handler); err != nil {
		log.Fatal(err)
	}
}
