package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/spf13/viper"
	"github.com/sulavmhrzn/gatekeeper/internal/middleware"
)

type Config struct {
	Server struct {
		Port      int    `mapstructure:"port"`
		SecretKey string `mapstructure:"secret_key"`
	} `mapstructure:"server"`
	Backend struct {
		URL string `mapstructure:"url"`
	} `mapstructure:"backend"`
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

	origin, err := url.Parse(cfg.Backend.URL)
	if err != nil {
		log.Fatal("Invalid origin Url:", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(origin)

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("Backend error: %v", err)
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("Gatekeeper Error: Backend is unreachable"))
	}
	var handler http.Handler = proxy

	handler = middleware.RecoveryMiddleware(
		middleware.LoggingMiddleware(
			middleware.AuthMiddleware(handler, cfg.Server.SecretKey),
		),
	)

	log.Printf("Gatekeeper listening on :%d... (Proxying to :%s)", cfg.Server.Port, cfg.Backend.URL)
	if err := http.ListenAndServe(":8000", handler); err != nil {
		log.Fatal(err)
	}
}
