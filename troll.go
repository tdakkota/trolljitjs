package main

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/tg"
)

type Troll struct {
	domain, stickerSet string

	resolved    *tg.InputPeerUser
	resolvedMux sync.RWMutex
	sticker     *tg.Document
	stickerMux  sync.RWMutex

	raw    *tg.Client
	sender *message.Sender
	logger *zap.Logger
}

func NewTroll(domain, stickerSet string, raw *tg.Client) *Troll {
	return &Troll{
		domain:     domain,
		stickerSet: stickerSet,
		raw:        raw,
		sender:     message.NewSender(raw),
		logger:     zap.NewNop(),
	}
}

// WithLogger sets logger to use.
func (t *Troll) WithLogger(logger *zap.Logger) *Troll {
	t.logger = logger
	return t
}

func (t *Troll) checkUserID(id int) (tg.InputPeerUser, bool) {
	t.resolvedMux.RLock()
	if t.resolved == nil {
		t.resolvedMux.RUnlock()
		return tg.InputPeerUser{}, false
	}
	resolved := *t.resolved
	t.resolvedMux.RUnlock()

	if resolved.UserID != id {
		return tg.InputPeerUser{}, false
	}

	return resolved, true
}

func (t *Troll) checkSticker() (tg.Document, bool) {
	t.stickerMux.RLock()
	if t.sticker == nil {
		t.stickerMux.RUnlock()
		return tg.Document{}, false
	}
	sticker := *t.sticker
	t.stickerMux.RUnlock()

	return sticker, true
}

func (t *Troll) getSticker(ctx context.Context) error {
	set, err := t.raw.MessagesGetStickerSet(ctx, &tg.InputStickerSetShortName{
		ShortName: t.stickerSet,
	})
	if err != nil {
		return xerrors.Errorf("get sticker set %q", t.stickerSet)
	}

	if len(set.Documents) < 1 {
		return xerrors.Errorf("sticker set is empty %v", set)
	}

	sticker, ok := set.Documents[len(set.Documents)-1].AsNotEmpty()
	if !ok {
		return xerrors.Errorf("last sticker is empty document %v", set)
	}

	t.stickerMux.Lock()
	t.sticker = sticker
	t.stickerMux.Unlock()
	return nil
}

func (t *Troll) getUser(ctx context.Context) error {
	p, err := t.sender.Resolve(t.domain, peer.OnlyUser).AsInputPeer(ctx)
	if err != nil {
		return xerrors.Errorf("resolve %q: %w", t.domain, err)
	}

	userPeer, ok := p.(*tg.InputPeerUser)
	if !ok {
		return xerrors.Errorf("unexpected type %T", p)
	}
	t.resolvedMux.Lock()
	t.resolved = userPeer
	t.resolvedMux.Unlock()

	return nil
}

func (t *Troll) setup(ctx context.Context) error {
	if err := t.getUser(ctx); err != nil {
		return xerrors.Errorf("get user: %w", err)
	}

	if err := t.getSticker(ctx); err != nil {
		t.logger.Warn("Get sticker failed", zap.Error(err))
	}

	return nil
}

func (t *Troll) Run(ctx context.Context) error {
	if err := t.setup(ctx); err != nil {
		return xerrors.Errorf("setup: %w", err)
	}

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
