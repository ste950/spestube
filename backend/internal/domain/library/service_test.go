package library_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/ste950/spestube/backend/internal/core/apperr"
	"github.com/ste950/spestube/backend/internal/core/clock"
	"github.com/ste950/spestube/backend/internal/domain/library"
	"github.com/ste950/spestube/backend/internal/domain/library/librarytest"
)

const (
	chanA = "UCaaaaaaaaaaaaaaaaaaaaaa"
	chanB = "UCbbbbbbbbbbbbbbbbbbbbbb"
)

// world è un servizio con i suoi fake e un orologio che il test sposta a mano.
type world struct {
	svc      *library.Service
	channels *librarytest.Channels
	videos   *librarytest.Videos
	feeds    *librarytest.Feeds
	handles  *librarytest.Handles
	policy   *library.Policy
	now      time.Time
	loc      *time.Location
}

func newWorld(t *testing.T) *world {
	t.Helper()
	loc, err := time.LoadLocation("Europe/Rome")
	if err != nil {
		t.Fatal(err)
	}
	w := &world{
		channels: librarytest.NewChannels(),
		videos:   librarytest.NewVideos(),
		feeds:    &librarytest.Feeds{LongVideos: map[string]library.Feed{}, Channel: map[string]library.Feed{}, Down: map[string]bool{}},
		handles:  &librarytest.Handles{IDs: map[string]string{}},
		policy:   &library.Policy{ReleaseAt: clock.TimeOfDay{Hour: 7}},
		now:      time.Date(2026, time.November, 15, 10, 0, 0, 0, loc),
		loc:      loc,
	}
	w.svc = library.NewService(library.Deps{
		Channels: w.channels, Videos: w.videos, Feeds: w.feeds, Handles: w.handles,
		Policy: policySource{w}, Now: func() time.Time { return w.now }, Loc: loc,
	})
	return w
}

// policySource legge la policy dal mondo a ogni chiamata: un test può cambiarla a metà.
type policySource struct{ w *world }

func (p policySource) Policy(context.Context) (library.Policy, error) { return *p.w.policy, nil }

func (w *world) at(day, hour, minute int) time.Time {
	return time.Date(2026, time.November, day, hour, minute, 0, 0, w.loc)
}

// feedOf costruisce un feed di n video, dal più recente (v1) al più vecchio.
func feedOf(title string, n int, newest time.Time) library.Feed {
	feed := library.Feed{ChannelTitle: title}
	for i := 1; i <= n; i++ {
		feed.Entries = append(feed.Entries, library.FeedEntry{
			VideoID:     fmt.Sprintf("%s-v%d", title, i),
			Title:       fmt.Sprintf("Video %d", i),
			PublishedAt: newest.Add(-time.Duration(i) * 24 * time.Hour),
		})
	}
	return feed
}

func codeOf(err error) string {
	var ae *apperr.Error
	if errors.As(err, &ae) {
		return ae.Code
	}
	if err != nil {
		return "errore non applicativo: " + err.Error()
	}
	return ""
}

func (w *world) statuses() map[library.Status]int {
	out := map[library.Status]int{}
	for _, v := range w.videos.ByID {
		out[v.Status]++
	}
	return out
}

func TestAddChannelKeepsOnlyTheMostRecentVideos(t *testing.T) {
	w := newWorld(t)
	w.feeds.LongVideos[chanA] = feedOf("Officina", 15, w.now)

	ch, err := w.svc.AddChannel(context.Background(), "https://www.youtube.com/channel/"+chanA)
	if err != nil {
		t.Fatal(err)
	}
	if ch.Title != "Officina" || ch.FeedMode != library.FeedLongVideos || !ch.LastFetchedAt.Equal(w.now) {
		t.Errorf("canale = %+v", ch)
	}
	got := w.statuses()
	if got[library.StatusNew] != library.BaselineKeep || got[library.StatusDiscarded] != 15-library.BaselineKeep {
		t.Fatalf("stati = %v, attesi %d nuovi e il resto scartato", got, library.BaselineKeep)
	}
	// I nuovi sono i più recenti, non i primi tre in ordine di arrivo.
	for i := 1; i <= 15; i++ {
		v := w.videos.ByID[fmt.Sprintf("Officina-v%d", i)]
		wantNew := i <= library.BaselineKeep
		if (v.Status == library.StatusNew) != wantNew {
			t.Errorf("video %d: stato %s", i, v.Status)
		}
		if !wantNew && v.Discard != library.DiscardBaseline {
			t.Errorf("video %d: motivo dello scarto %q", i, v.Discard)
		}
		if v.Source != library.FeedLongVideos || !v.FetchedAt.Equal(w.now) {
			t.Errorf("video %d: feed %q, scaricato %v", i, v.Source, v.FetchedAt)
		}
	}
}

func TestAddChannelFallsBackToTheChannelFeed(t *testing.T) {
	w := newWorld(t)
	w.feeds.Channel[chanA] = feedOf("Solo Canale", 2, w.now)

	ch, err := w.svc.AddChannel(context.Background(), chanA)
	if err != nil {
		t.Fatal(err)
	}
	if ch.FeedMode != library.FeedChannel {
		t.Errorf("feed = %q, atteso il feed del canale", ch.FeedMode)
	}
	for _, v := range w.videos.ByID {
		if v.Source != library.FeedChannel {
			t.Errorf("video %s: feed %q", v.ID, v.Source)
		}
	}
}

func TestAddChannelErrors(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		setup    func(*world)
		wantCode string
	}{
		{"riferimento non riconosciuto", "veritasium", func(*world) {}, "channel_ref_unsupported"},
		{"canale che non esiste su YouTube", chanA, func(*world) {}, "channel_not_found"},
		{"YouTube non risponde", chanA, func(w *world) { w.feeds.Down[chanA] = true }, "youtube_unavailable"},
		{"canale già seguito", chanA, func(w *world) {
			w.channels.ByID[chanA] = library.Channel{ID: chanA}
			w.feeds.LongVideos[chanA] = feedOf("Officina", 1, w.now)
		}, "channel_exists"},
		{"handle senza chiave API", "@officina", func(w *world) { w.handles.NoKey = true }, "handle_needs_api_key"},
		{"handle inesistente", "@nessuno", func(*world) {}, "channel_not_found"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := newWorld(t)
			tc.setup(w)
			before := len(w.channels.ByID)
			_, err := w.svc.AddChannel(context.Background(), tc.input)
			if code := codeOf(err); code != tc.wantCode {
				t.Fatalf("codice = %q, atteso %q", code, tc.wantCode)
			}
			if len(w.channels.ByID) != before || len(w.videos.ByID) != 0 {
				t.Errorf("un'aggiunta fallita non deve lasciare tracce: canali %d, video %d", len(w.channels.ByID), len(w.videos.ByID))
			}
		})
	}
}

func TestAddChannelByHandle(t *testing.T) {
	w := newWorld(t)
	w.handles.IDs["@officina"] = chanA
	w.feeds.LongVideos[chanA] = feedOf("Officina", 1, w.now)

	ch, err := w.svc.AddChannel(context.Background(), "https://youtube.com/@officina")
	if err != nil {
		t.Fatal(err)
	}
	if ch.ID != chanA || ch.Handle != "@officina" {
		t.Errorf("canale = %+v", ch)
	}
}

func TestFetchAllAddsOnlyUnseenVideosAsNew(t *testing.T) {
	w := newWorld(t)
	ctx := context.Background()
	w.feeds.LongVideos[chanA] = feedOf("Officina", 5, w.now)
	if _, err := w.svc.AddChannel(ctx, chanA); err != nil {
		t.Fatal(err)
	}

	// Un'ora dopo esce un video: dopo il primo download ogni video mai visto è nuovo,
	// senza il tetto dell'aggiunta, e quelli già noti non cambiano stato.
	w.now = w.now.Add(time.Hour)
	feed := w.feeds.LongVideos[chanA]
	feed.Entries = append([]library.FeedEntry{{VideoID: "appena-uscito", Title: "Novità", PublishedAt: w.now}}, feed.Entries...)
	w.feeds.LongVideos[chanA] = feed

	if err := w.svc.FetchAll(ctx); err != nil {
		t.Fatal(err)
	}
	fresh := w.videos.ByID["appena-uscito"]
	if fresh.Status != library.StatusNew || !fresh.FetchedAt.Equal(w.now) {
		t.Errorf("video nuovo = %+v", fresh)
	}
	got := w.statuses()
	if got[library.StatusNew] != library.BaselineKeep+1 || got[library.StatusDiscarded] != 5-library.BaselineKeep {
		t.Errorf("stati = %v", got)
	}
	if old := w.videos.ByID["Officina-v1"]; !old.FetchedAt.Equal(w.now.Add(-time.Hour)) {
		t.Errorf("un video già noto è stato riscritto: %+v", old)
	}
}

func TestFetchAllContinuesPastAnUnreachableChannel(t *testing.T) {
	w := newWorld(t)
	ctx := context.Background()
	fetched := w.now.Add(-time.Hour)
	w.channels.ByID[chanA] = library.Channel{ID: chanA, FeedMode: library.FeedLongVideos, LastFetchedAt: fetched}
	w.channels.ByID[chanB] = library.Channel{ID: chanB, FeedMode: library.FeedLongVideos, LastFetchedAt: fetched}
	w.feeds.Down[chanA] = true
	w.feeds.LongVideos[chanB] = feedOf("Storia", 2, w.now)

	err := w.svc.FetchAll(ctx)
	if apperr.KindOf(err) != apperr.KindUnavailable {
		t.Fatalf("errore = %v, atteso Unavailable", err)
	}
	if got := w.statuses()[library.StatusNew]; got != 2 {
		t.Errorf("video del canale raggiungibile = %d, attesi 2", got)
	}
	down := w.channels.ByID[chanA]
	if down.LastFetchError == "" || !down.LastFetchedAt.Equal(fetched) {
		t.Errorf("canale irraggiungibile = %+v: l'errore va registrato, la data dell'ultima lettura riuscita no", down)
	}
	if up := w.channels.ByID[chanB]; up.LastFetchError != "" || !up.LastFetchedAt.Equal(w.now) {
		t.Errorf("canale raggiungibile = %+v", up)
	}
}

func TestFetchAllBaselinesAChannelNeverFetched(t *testing.T) {
	w := newWorld(t)
	// Un canale entrato senza download (un import, o un primo tentativo fallito) riceve
	// lo stesso trattamento dell'aggiunta: solo i più recenti sono nuovi.
	w.channels.ByID[chanA] = library.Channel{ID: chanA, FeedMode: library.FeedLongVideos}
	w.feeds.LongVideos[chanA] = feedOf("Officina", 10, w.now)

	if err := w.svc.FetchAll(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := w.statuses(); got[library.StatusNew] != library.BaselineKeep {
		t.Errorf("stati = %v", got)
	}
}

func TestReleaseDue(t *testing.T) {
	w := newWorld(t)
	ctx := context.Background()
	newVideo := func(id string, fetched time.Time) library.Video {
		return library.Video{ID: id, Title: id, Status: library.StatusNew, FetchedAt: fetched, Source: library.FeedLongVideos}
	}
	w.videos.ByID["prima-delle-7"] = newVideo("prima-delle-7", w.at(15, 6, 30))
	w.videos.ByID["dopo-le-7"] = newVideo("dopo-le-7", w.at(15, 8, 0))
	w.videos.ByID["ieri-sera"] = newVideo("ieri-sera", w.at(14, 22, 0))
	w.videos.ByID["reaction"] = newVideo("reaction", w.at(15, 6, 0))
	w.policy.Filters = library.Filters{ExcludedKeywords: []string{"reaction"}}

	// Alle 6:59 l'ultimo orario di rilascio è le 7 di ieri, e tutti e quattro i video
	// sono arrivati dopo: non esce niente.
	w.now = w.at(15, 6, 59)
	if err := w.svc.ReleaseDue(ctx); err != nil {
		t.Fatal(err)
	}
	if got := w.statuses(); got[library.StatusNew] != 4 {
		t.Fatalf("alle 6:59 stati = %v, attesi 4 nuovi", got)
	}

	// Alle 7:00 escono i video scaricati prima delle 7; quello delle 8 non c'è ancora.
	w.now = w.at(15, 7, 0)
	if err := w.svc.ReleaseDue(ctx); err != nil {
		t.Fatal(err)
	}
	for id, want := range map[string]library.Status{
		"prima-delle-7": library.StatusReleased,
		"ieri-sera":     library.StatusReleased,
		"reaction":      library.StatusDiscarded,
	} {
		if got := w.videos.ByID[id].Status; got != want {
			t.Errorf("alle 7:00 %s = %s, atteso %s", id, got, want)
		}
	}
	if got := w.videos.ByID["reaction"].Discard; got != library.DiscardKeyword {
		t.Errorf("motivo dello scarto = %q", got)
	}

	// Durante il giorno il video delle 8 arriva e resta in attesa, a ogni giro.
	w.now = w.at(15, 8, 1)
	if err := w.svc.ReleaseDue(ctx); err != nil {
		t.Fatal(err)
	}
	w.now = w.at(15, 23, 59)
	if err := w.svc.ReleaseDue(ctx); err != nil {
		t.Fatal(err)
	}
	if got := w.videos.ByID["dopo-le-7"].Status; got != library.StatusNew {
		t.Errorf("il video delle 8 è uscito prima del rilascio successivo: %s", got)
	}
	if got := w.videos.ByID["prima-delle-7"].Status; got != library.StatusReleased {
		t.Errorf("il rilascio di oggi è sparito durante il giorno: %s", got)
	}

	// Il giorno dopo alle 7: esce il video delle 8, e ciò che è rimasto in Oggi da ieri
	// viene archiviato.
	w.now = w.at(16, 7, 0)
	if err := w.svc.ReleaseDue(ctx); err != nil {
		t.Fatal(err)
	}
	if got := w.videos.ByID["dopo-le-7"].Status; got != library.StatusReleased {
		t.Errorf("il video delle 8 = %s, atteso rilasciato il giorno dopo", got)
	}
	for _, id := range []string{"prima-delle-7", "ieri-sera"} {
		v := w.videos.ByID[id]
		if v.Status != library.StatusArchived || !v.ArchivedAt.Equal(w.now) {
			t.Errorf("%s = %s, atteso archiviato al rilascio successivo", id, v.Status)
		}
	}
}

func TestReleaseDueCatchesUpAfterDowntime(t *testing.T) {
	w := newWorld(t)
	w.videos.ByID["a"] = library.Video{ID: "a", Title: "a", Status: library.StatusNew, FetchedAt: w.at(14, 20, 0)}

	// Il dispositivo era spento alle 7 e si riaccende alle 11: il rilascio avviene
	// allora, e porta l'ora vera, così non viene archiviato al giro successivo.
	w.now = w.at(15, 11, 0)
	if err := w.svc.ReleaseDue(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := w.svc.ReleaseDue(context.Background()); err != nil {
		t.Fatal(err)
	}
	v := w.videos.ByID["a"]
	if v.Status != library.StatusReleased || !v.ReleasedAt.Equal(w.now) {
		t.Errorf("video = %+v", v)
	}
}

func TestReleaseDueKeepsWatchedVideos(t *testing.T) {
	w := newWorld(t)
	yesterday := w.at(14, 7, 0)
	w.videos.ByID["visto"] = library.Video{ID: "visto", Status: library.StatusWatched, ReleasedAt: yesterday}
	w.videos.ByID["salvato"] = library.Video{
		ID: "salvato", Status: library.StatusReleased, ReleasedAt: yesterday, CollectionIDs: []int64{1},
	}

	if err := w.svc.ReleaseDue(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := w.videos.ByID["visto"].Status; got != library.StatusWatched {
		t.Errorf("visto = %s: il rilascio archivia solo ciò che è rimasto in Oggi", got)
	}
	// Un video salvato esce da Oggi come gli altri: resta raggiungibile dalla raccolta.
	saved := w.videos.ByID["salvato"]
	if saved.Status != library.StatusArchived || len(saved.CollectionIDs) != 1 {
		t.Errorf("salvato = %+v", saved)
	}
}

func TestReleaseNowIgnoresTheSchedule(t *testing.T) {
	w := newWorld(t)
	w.videos.ByID["appena-scaricato"] = library.Video{ID: "appena-scaricato", Title: "x", Status: library.StatusNew, FetchedAt: w.now}
	w.videos.ByID["di-ieri"] = library.Video{ID: "di-ieri", Status: library.StatusReleased, ReleasedAt: w.at(14, 7, 0)}

	res, err := w.svc.ReleaseNow(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if res.Released != 1 || res.Discarded != 0 {
		t.Errorf("esito = %+v", res)
	}
	if got := w.statuses(); got[library.StatusReleased] != 2 {
		t.Errorf("stati = %v: il rilascio forzato non archivia", got)
	}
}

func TestTodayReportsTheNextRelease(t *testing.T) {
	w := newWorld(t)
	w.videos.ByID["a"] = library.Video{ID: "a", Status: library.StatusReleased, ReleasedAt: w.now}
	w.videos.ByID["b"] = library.Video{ID: "b", Status: library.StatusNew, FetchedAt: w.now}

	today, err := w.svc.Today(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(today.Videos) != 1 || today.Videos[0].ID != "a" {
		t.Errorf("video = %+v: in Oggi solo quelli rilasciati", today.Videos)
	}
	if want := w.at(16, 7, 0); !today.NextReleaseAt.Equal(want) {
		t.Errorf("prossimo rilascio = %v, atteso %v", today.NextReleaseAt, want)
	}
}

func TestVideoHidesWhatWasNeverShown(t *testing.T) {
	w := newWorld(t)
	ctx := context.Background()
	w.videos.ByID["nuovo"] = library.Video{ID: "nuovo", Status: library.StatusNew}
	w.videos.ByID["uscito"] = library.Video{ID: "uscito", Status: library.StatusReleased}

	if _, err := w.svc.Video(ctx, "nuovo"); codeOf(err) != "video_not_found" {
		t.Errorf("video non rilasciato: %v", err)
	}
	if _, err := w.svc.Video(ctx, "manca"); codeOf(err) != "video_not_found" {
		t.Errorf("video inesistente: %v", err)
	}
	if v, err := w.svc.Video(ctx, "uscito"); err != nil || v.ID != "uscito" {
		t.Errorf("video rilasciato: %+v, %v", v, err)
	}
}

func TestVideoActionsAreSaved(t *testing.T) {
	w := newWorld(t)
	ctx := context.Background()
	w.videos.ByID["b"] = library.Video{ID: "b", Status: library.StatusReleased}
	w.videos.ByID["c"] = library.Video{ID: "c", Status: library.StatusReleased}

	if _, err := w.svc.MarkWatched(ctx, "b"); err != nil {
		t.Fatal(err)
	}
	if _, err := w.svc.Discard(ctx, "c"); err != nil {
		t.Fatal(err)
	}
	if got := w.statuses(); got[library.StatusWatched] != 1 || got[library.StatusDiscarded] != 1 {
		t.Errorf("stati = %v", got)
	}
	if _, err := w.svc.Discard(ctx, "b"); codeOf(err) != "video_not_discardable" {
		t.Errorf("scarto di un video visto: %v", err)
	}
}
