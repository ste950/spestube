package library

import (
	"testing"
	"time"
)

var t0 = time.Date(2026, time.November, 15, 10, 0, 0, 0, time.UTC)

func TestFiltersCheck(t *testing.T) {
	all := Filters{ExcludeShorts: true, ExcludeLive: true, MinDuration: 5 * time.Minute, ExcludedKeywords: []string{"Reaction", "  ", "drama"}}
	cases := []struct {
		name    string
		filters Filters
		video   Video
		want    DiscardReason
	}{
		{"video lungo dalla playlist dei video lunghi", all, Video{Title: "Antenne LoRa", Source: FeedLongVideos, Duration: 18 * time.Minute}, DiscardNone},
		{"durata sconosciuta non scarta", all, Video{Title: "Antenne LoRa", Source: FeedChannel}, DiscardNone},
		{"parola chiave, maiuscole diverse", all, Video{Title: "La mia REACTION al video", Duration: time.Hour}, DiscardKeyword},
		{"parola chiave dentro una parola", all, Video{Title: "Melodrama in tre atti", Duration: time.Hour}, DiscardKeyword},
		{"parola chiave vuota non scarta tutto", Filters{ExcludedKeywords: []string{"", "   "}}, Video{Title: "Qualsiasi"}, DiscardNone},
		{"diretta", all, Video{Title: "In diretta", IsLive: true, Duration: time.Hour}, DiscardLive},
		{"diretta con il filtro spento", Filters{}, Video{Title: "In diretta", IsLive: true}, DiscardNone},
		{"short dal feed del canale", all, Video{Title: "Trucco veloce", Source: FeedChannel, Duration: 45 * time.Second}, DiscardShort},
		{"short al limite dei tre minuti", all, Video{Title: "Trucco", Source: FeedChannel, Duration: ShortMaxDuration}, DiscardShort},
		{"appena oltre il limite degli short, sotto la durata minima", all, Video{Title: "Trucco", Source: FeedChannel, Duration: ShortMaxDuration + time.Second}, DiscardDuration},
		{"video corto dalla playlist dei video lunghi non è uno short", Filters{ExcludeShorts: true}, Video{Title: "Annuncio", Source: FeedLongVideos, Duration: time.Minute}, DiscardNone},
		{"short con il filtro spento", Filters{}, Video{Title: "Trucco", Source: FeedChannel, Duration: 30 * time.Second}, DiscardNone},
		{"sotto la durata minima", all, Video{Title: "Breve", Source: FeedLongVideos, Duration: 4*time.Minute + 59*time.Second}, DiscardDuration},
		{"esattamente la durata minima", all, Video{Title: "Giusto", Source: FeedLongVideos, Duration: 5 * time.Minute}, DiscardNone},
		{"nessun filtro", Filters{}, Video{Title: "reaction drama", Duration: time.Second}, DiscardNone},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.filters.Check(tc.video); got != tc.want {
				t.Errorf("Check = %q, atteso %q", got, tc.want)
			}
		})
	}
}

func TestRelease(t *testing.T) {
	v := Video{ID: "a", Title: "Un video", Status: StatusNew}

	released := v.Release(t0, Filters{})
	if released.Status != StatusReleased || !released.ReleasedAt.Equal(t0) || released.Discard != DiscardNone {
		t.Errorf("rilascio: %+v", released)
	}

	discarded := v.Release(t0, Filters{ExcludedKeywords: []string{"video"}})
	if discarded.Status != StatusDiscarded || discarded.Discard != DiscardKeyword || !discarded.ReleasedAt.IsZero() {
		t.Errorf("scarto: %+v", discarded)
	}
}

func TestEnqueue(t *testing.T) {
	cases := []struct {
		name     string
		status   Status
		reason   string
		wantCode string
	}{
		{"da Oggi con un motivo", StatusReleased, "mi serve per l'orto", ""},
		{"da archiviato torna in coda", StatusArchived, "ci ripenso", ""},
		{"senza motivo", StatusReleased, "", "queue_reason_required"},
		{"motivo di soli spazi", StatusReleased, "  \n ", "queue_reason_required"},
		{"video non ancora rilasciato", StatusNew, "curiosità", "video_not_queueable"},
		{"video scartato", StatusDiscarded, "curiosità", "video_not_queueable"},
		{"già in coda", StatusQueued, "di nuovo", "video_not_queueable"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Video{ID: "a", Status: tc.status, ArchivedAt: t0.Add(-time.Hour)}.Enqueue(tc.reason, t0)
			if code := codeOf(err); code != tc.wantCode {
				t.Fatalf("codice = %q, atteso %q", code, tc.wantCode)
			}
			if err != nil {
				return
			}
			if got.Status != StatusQueued || got.QueueReason != tc.reason || !got.QueuedAt.Equal(t0) || !got.ArchivedAt.IsZero() {
				t.Errorf("in coda: %+v", got)
			}
		})
	}
}

func TestEnqueueTrimsTheReason(t *testing.T) {
	got, err := Video{Status: StatusReleased}.Enqueue("  per il weekend ", t0)
	if err != nil || got.QueueReason != "per il weekend" {
		t.Errorf("motivo = %q, err = %v", got.QueueReason, err)
	}
}

func TestQueueExpiry(t *testing.T) {
	queued := Video{Status: StatusQueued, QueuedAt: t0}
	day := 24 * time.Hour
	if want := t0.Add(14 * day); !queued.QueueExpiresAt().Equal(want) {
		t.Errorf("scadenza = %v, attesa %v", queued.QueueExpiresAt(), want)
	}
	cases := []struct {
		name string
		now  time.Time
		want int
	}{
		{"appena entrato", t0, 14},
		{"dopo un secondo ne restano ancora 14", t0.Add(time.Second), 14},
		{"dopo tre giorni", t0.Add(3 * day), 11},
		{"a trenta ore dalla scadenza", t0.Add(14*day - 30*time.Hour), 2},
		{"a tre ore dalla scadenza: scade domani", t0.Add(14*day - 3*time.Hour), 1},
		{"alla scadenza", t0.Add(14 * day), 0},
		{"oltre la scadenza", t0.Add(20 * day), 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := queued.QueueDaysLeft(tc.now); got != tc.want {
				t.Errorf("giorni rimasti = %d, attesi %d", got, tc.want)
			}
		})
	}

	notQueued := Video{Status: StatusReleased, QueuedAt: t0}
	if !notQueued.QueueExpiresAt().IsZero() || notQueued.QueueDaysLeft(t0) != 0 {
		t.Error("un video non in coda non ha scadenza")
	}
}

func TestMarkWatchedAndDiscard(t *testing.T) {
	for _, s := range []Status{StatusReleased, StatusQueued, StatusArchived} {
		got, err := Video{Status: s}.MarkWatched(t0)
		if err != nil || got.Status != StatusWatched || !got.WatchedAt.Equal(t0) {
			t.Errorf("visto da %s: %+v, %v", s, got, err)
		}
	}
	// Segnare due volte come visto non sposta la data della prima.
	again, err := Video{Status: StatusWatched, WatchedAt: t0}.MarkWatched(t0.Add(time.Hour))
	if err != nil || !again.WatchedAt.Equal(t0) {
		t.Errorf("visto due volte: %+v, %v", again, err)
	}
	if _, err := (Video{Status: StatusNew}).MarkWatched(t0); codeOf(err) != "video_not_watchable" {
		t.Errorf("visto da new: %v", err)
	}

	got, err := Video{Status: StatusReleased}.DiscardByUser()
	if err != nil || got.Status != StatusDiscarded || got.Discard != DiscardUser {
		t.Errorf("scarto: %+v, %v", got, err)
	}
	if _, err := (Video{Status: StatusWatched}).DiscardByUser(); codeOf(err) != "video_not_discardable" {
		t.Errorf("scarto di un video visto: %v", err)
	}
}

func TestPlayable(t *testing.T) {
	want := map[Status]bool{
		StatusNew: false, StatusDiscarded: false,
		StatusReleased: true, StatusQueued: true, StatusWatched: true, StatusArchived: true,
	}
	for s, playable := range want {
		if got := (Video{Status: s}).Playable(); got != playable {
			t.Errorf("Playable(%s) = %v, atteso %v", s, got, playable)
		}
	}
}
