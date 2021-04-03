package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strconv"

	"github.com/joho/godotenv"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/xerrors"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

func getLogLevel(or zapcore.Level) (zapcore.Level, error) {
	l, ok := os.LookupEnv("LOG_LEVEL")
	if !ok {
		return or, nil
	}

	var loglevel zapcore.Level
	if err := loglevel.UnmarshalText([]byte(l)); err == nil {
		return loglevel, nil
	}

	level, err := strconv.Atoi(l)
	if err != nil {
		return 0, err
	}

	return zapcore.Level(level), nil
}

func run(ctx context.Context) (err error) {
	envErr := godotenv.Load()

	level, err := getLogLevel(zapcore.DebugLevel)
	if err != nil {
		return xerrors.Errorf("get log level: %w", err)
	}

	logger, _ := zap.NewDevelopment(zap.IncreaseLevel(level))
	defer func() {
		multierr.AppendInto(&err, logger.Sync())
	}()

	if envErr != nil {
		logger.Info("Load environment file failed", zap.Error(envErr))
	}

	trollDomain := os.Getenv("TROLL")
	if trollDomain == "" {
		return xerrors.Errorf("empty TROLL env variable: %q", trollDomain)
	}
	stickerSet := os.Getenv("STICKER_SET")
	if stickerSet == "" {
		stickerSet = "wtfakkota"
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
	troll := NewTroll(trollDomain, stickerSet, raw).
		WithLogger(logger.Named("troll"))
	troll.Register(dispatcher)
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
