package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"

	"github.com/k0kubun/pp/v3"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"golang.org/x/xerrors"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/tg"
)

// RandInt64 returns random int64 from randSource.
func RandInt64(randSource io.Reader) (int64, error) {
	var buf [bin.Word * 2]byte
	if _, err := io.ReadFull(randSource, buf[:]); err != nil {
		return 0, err
	}
	b := &bin.Buffer{Buf: buf[:]}
	return b.Long()
}

func (t *Troll) tryTyping(ctx context.Context, resolved tg.InputPeerUser, log *zap.Logger) error {
	n, err := RandInt64(rand.Reader)
	if err != nil {
		return xerrors.Errorf("rand: %w", err)
	}

	switch {
	case n%2 == 0:
		log.Info("Ignoring")
		return nil
	case !t.limiter.Allow():
		log.Info("Rate limit")
		return nil
	}
	log.Info("Typing")

	s := t.sender.To(&resolved)
	if err := s.TypingAction().Typing(ctx); err != nil {
		return xerrors.Errorf("typing %d: %w", resolved.UserID, err)
	}
	return nil
}

func (t *Troll) OnUserTyping(ctx context.Context, e tg.Entities, update *tg.UpdateUserTyping) error {
	_, ok := t.checkUserID(update.UserID)
	if !ok {
		return nil
	}

	t.logger.Info(
		"Got user typing update",
		zap.String("action", fmt.Sprintf("%T", update.Action)),
	)
	_, err := t.sender.Self().StyledText(ctx, styling.Code(pp.Sprintln(update)))
	return err
}

func (t *Troll) OnUserPhoto(ctx context.Context, e tg.Entities, update *tg.UpdateUserPhoto) error {
	resolved, ok := t.checkUserID(update.UserID)
	if !ok {
		return nil
	}

	log := t.logger
	if photo, ok := update.Photo.AsNotEmpty(); ok {
		log = t.logger.With(
			zap.Int64("photo_id", photo.PhotoID),
		)
	}

	_, err := t.sender.Self().StyledText(ctx, styling.Code(pp.Sprintln(update)))
	return multierr.Append(err, t.tryTyping(ctx, resolved, log))
}

func (t *Troll) OnUserName(ctx context.Context, e tg.Entities, update *tg.UpdateUserName) error {
	resolved, ok := t.checkUserID(update.UserID)
	if !ok {
		return nil
	}

	log := t.logger.With(
		zap.String("username", update.Username),
		zap.String("first_name", update.FirstName),
		zap.String("last_name", update.LastName),
	)

	_, err := t.sender.Self().StyledText(ctx, styling.Code(pp.Sprintln(update)))
	return multierr.Append(err, t.tryTyping(ctx, resolved, log))
}

func (t *Troll) OnUserStatus(ctx context.Context, e tg.Entities, update *tg.UpdateUserStatus) error {
	resolved, ok := t.checkUserID(update.UserID)
	if !ok {
		return nil
	}

	log := t.logger.With(zap.String("status", fmt.Sprintf("%T", update.Status)))
	if _, ok := update.Status.(*tg.UserStatusOnline); !ok {
		return nil
	}

	return t.tryTyping(ctx, resolved, log)
}
