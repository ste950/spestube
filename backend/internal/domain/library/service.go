package library

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/ste950/spestube/backend/internal/core/apperr"
)

// BaselineKeep è quanti video recenti entrano come nuovi quando si aggiunge un canale.
// Il feed ne porta quindici: farli entrare tutti riempirebbe Oggi di arretrati il giorno
// dopo, che è il contrario di un feed che finisce.
const BaselineKeep = 3

// Service raccoglie i casi d'uso del soggetto.
type Service struct {
	channels ChannelRepository
	videos   VideoRepository
	feeds    FeedRepository
	handles  HandleRepository
	policy   PolicySource
	now      func() time.Time
	loc      *time.Location
}

// Deps sono le dipendenze del servizio. Loc è il fuso in cui si legge l'orario di
// rilascio: "alle 7" vuol dire le 7 di chi guarda, non le 7 UTC.
type Deps struct {
	Channels ChannelRepository
	Videos   VideoRepository
	Feeds    FeedRepository
	Handles  HandleRepository
	Policy   PolicySource
	Now      func() time.Time
	Loc      *time.Location
}

func NewService(d Deps) *Service {
	return &Service{
		channels: d.Channels, videos: d.Videos, feeds: d.Feeds, handles: d.Handles,
		policy: d.Policy, now: d.Now, loc: d.Loc,
	}
}

func (s *Service) ListChannels(ctx context.Context) ([]Channel, error) {
	return s.channels.List(ctx)
}

func (s *Service) RemoveChannel(ctx context.Context, id string) error {
	return s.channels.Remove(ctx, id)
}

// AddChannel aggiunge un canale da link, handle o id, e ne scarica subito i video: il
// feed serve comunque per conoscere il nome del canale e per sapere se esiste.
func (s *Service) AddChannel(ctx context.Context, input string) (Channel, error) {
	ref, err := ParseChannelRef(input)
	if err != nil {
		return Channel{}, err
	}
	id := ref.ID
	if ref.Handle != "" {
		if id, err = s.handles.ChannelID(ctx, ref.Handle); err != nil {
			return Channel{}, err
		}
	}
	if _, ok, err := s.channels.Find(ctx, id); err != nil {
		return Channel{}, err
	} else if ok {
		return Channel{}, apperr.Conflict("channel_exists", "segui già questo canale")
	}

	feed, mode, err := s.recent(ctx, id)
	if err != nil {
		return Channel{}, err
	}
	now := s.now()
	ch := Channel{ID: id, Title: feed.ChannelTitle, Handle: ref.Handle, FeedMode: mode, AddedAt: now}
	if ch.Title == "" {
		ch.Title = id
	}
	if err := s.channels.Add(ctx, ch); err != nil {
		return Channel{}, err
	}
	if err := s.ingest(ctx, ch, feed, mode, now); err != nil {
		return Channel{}, err
	}
	return s.channels.Get(ctx, id)
}

// FetchAll scarica i video nuovi di tutti i canali. Un canale irraggiungibile non ferma
// il giro: l'errore resta sul canale e torna a chi ha chiamato per il log.
func (s *Service) FetchAll(ctx context.Context) error {
	channels, err := s.channels.List(ctx)
	if err != nil {
		return err
	}
	var errs []error
	for _, ch := range channels {
		if err := ctx.Err(); err != nil {
			return err
		}
		feed, mode, err := s.recent(ctx, ch.ID)
		now := s.now()
		if err != nil {
			kind := apperr.KindOf(err)
			if kind != apperr.KindUnavailable && kind != apperr.KindNotFound {
				return err
			}
			if markErr := s.channels.MarkFetched(ctx, ch.ID, ch.FeedMode, now, err.Error()); markErr != nil {
				return markErr
			}
			errs = append(errs, fmt.Errorf("canale %s: %w", ch.ID, err))
			continue
		}
		if err := s.ingest(ctx, ch, feed, mode, now); err != nil {
			return err
		}
	}
	return errors.Join(errs...)
}

// recent prova prima la playlist dei video lunghi e ripiega sul feed del canale solo se
// quella non esiste. Se non esiste nemmeno il secondo, è il canale a non esistere.
func (s *Service) recent(ctx context.Context, channelID string) (Feed, FeedMode, error) {
	feed, err := s.feeds.Recent(ctx, channelID, FeedLongVideos)
	if err == nil {
		return feed, FeedLongVideos, nil
	}
	if apperr.KindOf(err) != apperr.KindNotFound {
		return Feed{}, "", err
	}
	feed, err = s.feeds.Recent(ctx, channelID, FeedChannel)
	if err == nil {
		return feed, FeedChannel, nil
	}
	if apperr.KindOf(err) == apperr.KindNotFound {
		return Feed{}, "", apperr.NotFound("channel_not_found", "questo canale non esiste su YouTube")
	}
	return Feed{}, "", err
}

func (s *Service) ingest(ctx context.Context, ch Channel, feed Feed, mode FeedMode, now time.Time) error {
	entries := append([]FeedEntry(nil), feed.Entries...)
	sort.SliceStable(entries, func(i, j int) bool { return entries[i].PublishedAt.After(entries[j].PublishedAt) })

	first := ch.LastFetchedAt.IsZero()
	videos := make([]Video, 0, len(entries))
	for i, e := range entries {
		v := Video{
			ID: e.VideoID, ChannelID: ch.ID, Title: e.Title, PublishedAt: e.PublishedAt,
			FetchedAt: now, Source: mode, Status: StatusNew,
		}
		if first && i >= BaselineKeep {
			v.Status = StatusDiscarded
			v.Discard = DiscardBaseline
		}
		videos = append(videos, v)
	}
	if _, err := s.videos.AddMissing(ctx, videos); err != nil {
		return err
	}
	return s.channels.MarkFetched(ctx, ch.ID, mode, now, "")
}

// ReleaseResult conta l'esito di un rilascio.
type ReleaseResult struct {
	Released  int
	Discarded int
	Archived  int
}

// ReleaseDue è il rilascio giornaliero. Si può chiamare quante volte si vuole: fa
// qualcosa solo se dall'ultimo orario di rilascio sono rimasti video in attesa.
//
// La regola non ha stato: un video nuovo diventa visibile al primo orario di rilascio
// successivo al suo download. Così un riavvio non fa saltare né raddoppiare un rilascio,
// e un dispositivo spento alle 7 recupera appena si riaccende.
func (s *Service) ReleaseDue(ctx context.Context) error {
	pol, err := s.policy.Policy(ctx)
	if err != nil {
		return err
	}
	now := s.now()
	last := pol.ReleaseAt.Last(now, s.loc)

	// I video del rilascio precedente rimasti in Oggi escono di scena: Oggi mostra solo
	// l'ultimo arrivo, e ciò che si vuole tenere va in Da vedere con un motivo.
	if _, err := s.videos.ArchiveReleasedBefore(ctx, last, now); err != nil {
		return err
	}
	_, err = s.release(ctx, last, now, pol.Filters)
	return err
}

// ReleaseNow rilascia subito tutti i video in attesa, senza aspettare l'orario. Serve
// in sviluppo, per vedere il risultato di un download.
func (s *Service) ReleaseNow(ctx context.Context) (ReleaseResult, error) {
	pol, err := s.policy.Policy(ctx)
	if err != nil {
		return ReleaseResult{}, err
	}
	now := s.now()
	return s.release(ctx, now.Add(time.Second), now, pol.Filters)
}

func (s *Service) release(ctx context.Context, fetchedBefore, now time.Time, f Filters) (ReleaseResult, error) {
	pending, err := s.videos.ListNewFetchedBefore(ctx, fetchedBefore)
	if err != nil {
		return ReleaseResult{}, err
	}
	var res ReleaseResult
	for _, v := range pending {
		v = v.Release(now, f)
		if err := s.videos.Save(ctx, v); err != nil {
			return res, err
		}
		if v.Status == StatusReleased {
			res.Released++
		} else {
			res.Discarded++
		}
	}
	return res, nil
}

// ExpireQueue archivia i video rimasti in Da vedere oltre QueueLifetime.
func (s *Service) ExpireQueue(ctx context.Context) error {
	now := s.now()
	_, err := s.videos.ArchiveQueuedBefore(ctx, now.Add(-QueueLifetime), now)
	return err
}

// Today è il contenuto della schermata Oggi.
type Today struct {
	Videos        []Video
	NextReleaseAt time.Time
}

func (s *Service) Today(ctx context.Context) (Today, error) {
	pol, err := s.policy.Policy(ctx)
	if err != nil {
		return Today{}, err
	}
	videos, err := s.videos.ListReleased(ctx)
	if err != nil {
		return Today{}, err
	}
	return Today{Videos: videos, NextReleaseAt: pol.ReleaseAt.Next(s.now(), s.loc)}, nil
}

func (s *Service) Queue(ctx context.Context) ([]Video, error) {
	return s.videos.ListQueued(ctx)
}

// Video restituisce un video che si può aprire in Guarda.
func (s *Service) Video(ctx context.Context, id string) (Video, error) {
	v, err := s.videos.Get(ctx, id)
	if err != nil {
		return Video{}, err
	}
	if !v.Playable() {
		return Video{}, apperr.NotFound("video_not_found", "video inesistente")
	}
	return v, nil
}

func (s *Service) Enqueue(ctx context.Context, id, reason string) (Video, error) {
	return s.change(ctx, id, func(v Video) (Video, error) { return v.Enqueue(reason, s.now()) })
}

func (s *Service) MarkWatched(ctx context.Context, id string) (Video, error) {
	return s.change(ctx, id, func(v Video) (Video, error) { return v.MarkWatched(s.now()) })
}

func (s *Service) Discard(ctx context.Context, id string) (Video, error) {
	return s.change(ctx, id, func(v Video) (Video, error) { return v.DiscardByUser() })
}

func (s *Service) change(ctx context.Context, id string, step func(Video) (Video, error)) (Video, error) {
	v, err := s.videos.Get(ctx, id)
	if err != nil {
		return Video{}, err
	}
	if v, err = step(v); err != nil {
		return Video{}, err
	}
	if err := s.videos.Save(ctx, v); err != nil {
		return Video{}, err
	}
	return v, nil
}
