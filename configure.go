package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"strings"

	"golang.org/x/xerrors"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/dcs"
)

func codeAsk(ctx context.Context) (string, error) {
	fmt.Print("code:")
	code, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return "", err
	}
	code = strings.ReplaceAll(code, "\n", "")
	return code, nil
}

func configure(test bool, opts telegram.Options) (*telegram.Client, telegram.UserAuthenticator, error) {
	if test {
		opts, err := telegram.OptionsFromEnvironment(opts)
		if err != nil {
			return nil, nil, xerrors.Errorf("from env: %w", err)
		}
		opts.DCList = dcs.StagingDCs()

		client := telegram.NewClient(telegram.TestAppID, telegram.TestAppHash, opts)
		flow := telegram.TestAuth(rand.Reader, 2)
		return client, flow, nil
	}

	client, err := telegram.ClientFromEnvironment(opts)
	if err != nil {
		return nil, nil, xerrors.Errorf("from env: %w", err)
	}
	return client, telegram.EnvAuth("", telegram.CodeAuthenticatorFunc(codeAsk)), nil
}
