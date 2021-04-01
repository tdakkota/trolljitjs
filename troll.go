package main

import (
	"context"
	"time"

	"go.uber.org/atomic"
	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/tg"
)

type Troll struct {
	domain     string
	resolvedID atomic.Int64

	raw    *tg.Client
	sender *message.Sender
	logger *zap.Logger
}

func NewTroll(domain string, raw *tg.Client) *Troll {
	return &Troll{
		domain: domain,
		raw:    raw,
		sender: message.NewSender(raw),
		logger: zap.NewNop(),
	}
}

// WithLogger sets logger to use.
func (t *Troll) WithLogger(logger *zap.Logger) *Troll {
	t.logger = logger
	return t
}

func (t *Troll) OnNewMessage(ctx context.Context, e tg.Entities, update *tg.UpdateNewMessage) error {
	msg, ok := update.Message.(*tg.Message)
	if !ok || msg.Out {
		return nil
	}

	u, ok := msg.GetPeerID().(*tg.PeerUser)
	if !ok || u.UserID != int(t.resolvedID.Load()) {
		return nil
	}

	t.logger.Info("Got message",
		zap.String("text", msg.Message),
		zap.Time("date", time.Unix(int64(msg.Date), 0)),
	)
	_, err := t.sender.Self().Revoke().Messages(ctx, msg.ID)
	return err
}

func (t *Troll) Run(ctx context.Context) error {
	p, err := t.sender.Resolve(t.domain, peer.OnlyUser).AsInputPeer(ctx)
	if err != nil {
		return xerrors.Errorf("resolve %q: %w", t.domain, err)
	}

	userPeer, ok := p.(*tg.InputPeerUser)
	if !ok {
		return xerrors.Errorf("unexpected type %T", p)
	}
	t.resolvedID.Store(int64(userPeer.UserID))

	ticker := time.NewTicker(2 * time.Minute)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			_, err := t.raw.AccountUpdateStatus(ctx, false)
			if err != nil {
				t.logger.Warn("Got error", zap.Error(err))
			}
		}
	}
}
