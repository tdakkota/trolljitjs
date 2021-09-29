package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"golang.org/x/net/proxy"
	"golang.org/x/xerrors"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/dcs"
	"github.com/gotd/td/tg"
)

func codeAsk(context.Context, *tg.AuthSentCode) (string, error) {
	fmt.Print("code:")
	code, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return "", err
	}
	code = strings.ReplaceAll(code, "\n", "")
	return code, nil
}

func configure(
	cfg Config,
	dispatcher telegram.UpdateHandler,
	logger *zap.Logger,
) (*telegram.Client, auth.UserAuthenticator, error) {
	dir, _ := filepath.Split(cfg.SessionFile)
	if dir != "" {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return nil, nil, xerrors.Errorf("session dir creation: %w", err)
		}
	}

	opts := telegram.Options{
		DC: cfg.DC,
		Resolver: dcs.Plain(dcs.PlainOptions{
			Dial: proxy.Dial,
		}),
		SessionStorage: &session.FileStorage{
			Path: cfg.SessionFile,
		},
		Logger:        logger,
		UpdateHandler: dispatcher,
		Middlewares: []telegram.Middleware{
			telegram.MiddlewareFunc(backoffRetry),
		},
	}

	if cfg.Test {
		opts.DCList = dcs.Test()
		client := telegram.NewClient(telegram.TestAppID, telegram.TestAppHash, opts)
		flow := auth.Test(rand.Reader, cfg.DC)
		return client, flow, nil
	}

	client := telegram.NewClient(cfg.AppID, cfg.AppHash, opts)
	return client, cfg.AsAuth(auth.CodeAuthenticatorFunc(codeAsk)), nil
}
