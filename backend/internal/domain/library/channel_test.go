package library

import (
	"errors"
	"testing"

	"github.com/ste950/spestube/backend/internal/core/apperr"
)

func TestParseChannelRef(t *testing.T) {
	const id = "UCHnyfMqiRRG1u-2MsSQLbXA"
	cases := []struct {
		name     string
		in       string
		want     ChannelRef
		wantCode string
	}{
		{"id nudo", id, ChannelRef{ID: id}, ""},
		{"id con spazi attorno", "  " + id + "\n", ChannelRef{ID: id}, ""},
		{"link al canale", "https://www.youtube.com/channel/" + id, ChannelRef{ID: id}, ""},
		{"link con sezione", "https://www.youtube.com/channel/" + id + "/videos", ChannelRef{ID: id}, ""},
		{"link mobile senza schema", "m.youtube.com/channel/" + id, ChannelRef{ID: id}, ""},
		{"handle", "@veritasium", ChannelRef{Handle: "@veritasium"}, ""},
		{"link con handle", "https://www.youtube.com/@veritasium/videos", ChannelRef{Handle: "@veritasium"}, ""},
		{"link con handle e query", "youtube.com/@Storia.in_Chiaro?si=abc", ChannelRef{Handle: "@Storia.in_Chiaro"}, ""},
		{"vuoto", "   ", ChannelRef{}, "channel_ref_required"},
		{"handle troppo corto", "@ab", ChannelRef{}, "channel_ref_invalid"},
		{"handle con spazi", "@due parole", ChannelRef{}, "channel_ref_invalid"},
		{"vecchio link /c/", "https://www.youtube.com/c/Veritasium", ChannelRef{}, "channel_ref_unsupported"},
		{"vecchio link /user/", "https://www.youtube.com/user/1veritasium", ChannelRef{}, "channel_ref_unsupported"},
		{"link a un video", "https://www.youtube.com/watch?v=KfODG5REyss", ChannelRef{}, "channel_ref_unsupported"},
		{"altro sito", "https://example.com/channel/" + id, ChannelRef{}, "channel_ref_unsupported"},
		{"id troppo corto", "UC123", ChannelRef{}, "channel_ref_unsupported"},
		{"testo qualunque", "veritasium", ChannelRef{}, "channel_ref_unsupported"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseChannelRef(tc.in)
			if code := codeOf(err); code != tc.wantCode {
				t.Fatalf("ParseChannelRef(%q) codice = %q, atteso %q (%v)", tc.in, code, tc.wantCode, err)
			}
			if got != tc.want {
				t.Errorf("ParseChannelRef(%q) = %+v, atteso %+v", tc.in, got, tc.want)
			}
		})
	}
}

func codeOf(err error) string {
	var ae *apperr.Error
	if errors.As(err, &ae) {
		return ae.Code
	}
	if err != nil {
		return "errore non applicativo"
	}
	return ""
}
