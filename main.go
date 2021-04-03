package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strconv"

	"go.uber.org/multierr"
	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

func run(ctx context.Context) (err error) {
	logger, _ := zap.NewDevelopment()
	defer func() {
		multierr.AppendInto(&err, logger.Sync())
	}()

	trollDomain := os.Getenv("TROLL")
	if trollDomain == "" {
		return xerrors.Errorf("empty TROLL env variable: %q", trollDomain)
	}
	_, test := os.LookupEnv("TEST")

	dispatcher := tg.NewUpdateDispatcher()
	client, flow, err := configure(test, telegram.Options{
		Logger:        logger,
		UpdateHandler: dispatcher,
	})
	if err != nil {
		return xerrors.Errorf("configure: %w", err)
	}

	raw := tg.NewClient(waitInvoker{client})
	troll := NewTroll(trollDomain, raw).
		WithLogger(logger.Named("troll"))
	dispatcher.OnNewMessage(troll.OnNewMessage)
	return client.Run(ctx, func(ctx context.Context) error {
		if err := client.AuthIfNecessary(ctx, telegram.NewAuth(flow, telegram.SendCodeOptions{})); err != nil {
			return xerrors.Errorf("auth flow: %w", err)
		}

		if test {
			username := "bot" + strconv.Itoa(rand.Intn(999999))
			_, err := raw.AccountUpdateUsername(ctx, username)
			if err != nil && !tgerr.Is(err, tg.ErrUsernameNotModified) {
				return xerrors.Errorf("update username: %w", err)
			}

			logger.Info("New username", zap.String("username", username))
		}

		return troll.Run(ctx)
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
