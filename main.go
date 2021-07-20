package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strconv"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

func run(ctx context.Context) error {
	cfg, err := loadConfig()
	if err != nil {
		return xerrors.Errorf("load config: %w", err)
	}

	logger, err := createLogger()
	if err != nil {
		return xerrors.Errorf("create logger: %w", err)
	}
	defer func() {
		_ = logger.Sync()
	}()

	dispatcher := tg.NewUpdateDispatcher()
	client, flow, err := configure(cfg, dispatcher, logger)
	if err != nil {
		return xerrors.Errorf("configure: %w", err)
	}

	raw := tg.NewClient(client)
	troll := NewTroll(cfg.Username, cfg.StickerSet, raw).
		WithLogger(logger.Named("troll"))
	troll.Register(dispatcher)
	return client.Run(ctx, func(ctx context.Context) error {
		if err := client.Auth().IfNecessary(ctx, auth.NewFlow(flow, auth.SendCodeOptions{})); err != nil {
			return xerrors.Errorf("auth flow: %w", err)
		}

		if cfg.Test {
			username := "bot" + strconv.Itoa(rand.Intn(999999)) // #nosec G404
			_, err := raw.AccountUpdateUsername(ctx, username)
			if err != nil && !tgerr.Is(err, tg.ErrUsernameNotModified) {
				return xerrors.Errorf("update username: %w", err)
			}

			logger.Info("New username", zap.String("username", username))
		}

		return troll.Run(ctx, cfg.StatusLoop)
	})
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := run(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(2)
	}
}
