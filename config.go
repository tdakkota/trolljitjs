package main

import (
	"context"

	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
	"golang.org/x/xerrors"

	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfigdotenv"
	"github.com/cristalhq/aconfig/aconfigyaml"
)

type Config struct {
	AppID       int    `json:"app_id" yaml:"app_id" env:"APP_ID"`
	AppHash     string `json:"app_hash" yaml:"app_hash" env:"APP_HASH"`
	SessionFile string `json:"session_file" yaml:"session_file" env:"SESSION_FILE"`
	Phone       string `json:"phone" yaml:"phone" env:"PHONE"`
	Password    string `json:"password" yaml:"password" env:"PASSWORD"`
	DC          int    `json:"dc" yaml:"dc" env:"DC" default:"2"`
	Test        bool   `json:"test" yaml:"test" env:"TEST" default:"false"`

	StatusLoop bool   `json:"status_loop" yaml:"status_loop" env:"STATUS_LOOP"`
	Username   string `json:"troll" yaml:"troll" env:"TROLL"`
	StickerSet string `json:"sticker_set" yaml:"sticker_set" env:"STICKER_SET"`
}

func (c Config) AsAuth(code auth.CodeAuthenticator) auth.UserAuthenticator {
	return configAuth{
		Config:            c,
		CodeAuthenticator: code,
	}
}

func loadConfig() (cfg Config, _ error) {
	loader := aconfig.LoaderFor(&cfg, aconfig.Config{
		Files: []string{"trolljitrs.yaml", "trolljitrs.yml"},
		FileDecoders: map[string]aconfig.FileDecoder{
			".yaml": aconfigyaml.New(),
			".yml":  aconfigyaml.New(),
			".env":  aconfigdotenv.New(),
		},
	})

	if err := loader.Load(); err != nil {
		return cfg, err
	}
	return cfg, nil
}

type configAuth struct {
	Config
	auth.CodeAuthenticator
}

func (c configAuth) Phone(ctx context.Context) (string, error) {
	return c.Config.Phone, nil
}

func (c configAuth) Password(ctx context.Context) (string, error) {
	return c.Config.Password, nil
}

func (c configAuth) AcceptTermsOfService(ctx context.Context, tos tg.HelpTermsOfService) error {
	return xerrors.New("unexpected call")
}

func (c configAuth) SignUp(ctx context.Context) (auth.UserInfo, error) {
	return auth.UserInfo{}, xerrors.New("unexpected call")
}
