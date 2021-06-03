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

func configure(test bool, opts telegram.Options) (*telegram.Client, auth.UserAuthenticator, error) {
	if test {
		opts, err := telegram.OptionsFromEnvironment(opts)
		if err != nil {
			return nil, nil, xerrors.Errorf("from env: %w", err)
		}
		opts.DCList = dcs.Staging()

		client := telegram.NewClient(telegram.TestAppID, telegram.TestAppHash, opts)
		flow := auth.Test(rand.Reader, 2)
		return client, flow, nil
	}

	client, err := telegram.ClientFromEnvironment(opts)
	if err != nil {
		return nil, nil, xerrors.Errorf("from env: %w", err)
	}
	return client, auth.Env("", auth.CodeAuthenticatorFunc(codeAsk)), nil
}
