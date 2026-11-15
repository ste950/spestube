package library

import (
	"strings"
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

func TestMarkWatchedAndDiscard(t *testing.T) {
	for _, s := range []Status{StatusReleased, StatusArchived} {
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
		StatusReleased: true, StatusWatched: true, StatusArchived: true,
	}
	for s, playable := range want {
		if got := (Video{Status: s}).Playable(); got != playable {
			t.Errorf("Playable(%s) = %v, atteso %v", s, got, playable)
		}
	}
}

func TestNewCollectionFields(t *testing.T) {
	long := strings.Repeat("a", CollectionNameMax+1)
	cases := []struct {
		name, inName, inDesc string
		want                 CollectionFields
		wantCode             string
	}{
		{"nome e descrizione", "Orto", "  idee per la primavera  ", CollectionFields{"Orto", "idee per la primavera"}, ""},
		{"spazi interni ridotti a uno", "  Cucina   lenta ", "", CollectionFields{"Cucina lenta", ""}, ""},
		{"nome lungo al limite", strings.Repeat("à", CollectionNameMax), "", CollectionFields{strings.Repeat("à", CollectionNameMax), ""}, ""},
		{"nome vuoto", "   ", "", CollectionFields{}, "collection_name_required"},
		{"nome troppo lungo", long, "", CollectionFields{}, "collection_name_too_long"},
		{"descrizione troppo lunga", "Orto", strings.Repeat("x", CollectionDescriptionMax+1), CollectionFields{}, "collection_description_too_long"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NewCollectionFields(tc.inName, tc.inDesc)
			if code := codeOf(err); code != tc.wantCode {
				t.Fatalf("codice = %q, atteso %q", code, tc.wantCode)
			}
			if got != tc.want {
				t.Errorf("campi = %+v, attesi %+v", got, tc.want)
			}
		})
	}
}

func TestCheckOrder(t *testing.T) {
	current := []int64{1, 2, 3}
	cases := []struct {
		name   string
		wanted []int64
		ok     bool
	}{
		{"stesso ordine", []int64{1, 2, 3}, true},
		{"riordinate", []int64{3, 1, 2}, true},
		{"ne manca una", []int64{3, 1}, false},
		{"una di troppo", []int64{3, 1, 2, 4}, false},
		{"una sconosciuta", []int64{3, 1, 4}, false},
		{"una ripetuta", []int64{1, 1, 2}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := CheckOrder(current, tc.wanted)
			if (err == nil) != tc.ok {
				t.Errorf("CheckOrder(%v) = %v", tc.wanted, err)
			}
			if err != nil && codeOf(err) != "collection_order_invalid" {
				t.Errorf("codice = %q", codeOf(err))
			}
		})
	}
	if err := CheckOrder(nil, nil); err != nil {
		t.Errorf("nessuna raccolta: %v", err)
	}
}
