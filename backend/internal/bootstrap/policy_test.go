package bootstrap

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/ste950/spestube/backend/internal/core/clock"
	"github.com/ste950/spestube/backend/internal/domain/library"
	"github.com/ste950/spestube/backend/internal/domain/settings"
	"github.com/ste950/spestube/backend/internal/domain/settings/settingstest"
)

// Il rilascio legge orario e filtri dalle impostazioni a ogni giro: un cambio fatto
// dall'app deve valere senza riavviare il processo.
func TestPolicyFollowsTheSettings(t *testing.T) {
	ctx := context.Background()
	svc := settings.NewService(settingstest.NewRepository(), settings.Defaults(clock.TimeOfDay{Hour: 7}), time.Now)
	policy := policyFromSettings{svc}

	got, err := policy.Policy(ctx)
	if err != nil {
		t.Fatal(err)
	}
	want := library.Policy{
		ReleaseAt: clock.TimeOfDay{Hour: 7},
		Filters:   library.Filters{ExcludeShorts: true, ExcludeLive: true},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("default = %+v, attesa %+v", got, want)
	}

	if _, err := svc.Update(ctx, settings.Settings{
		ReleaseTime: clock.TimeOfDay{Hour: 19, Minute: 30}, MinDurationMinutes: 5,
		ExcludedKeywords: []string{"reaction"},
	}); err != nil {
		t.Fatal(err)
	}
	got, err = policy.Policy(ctx)
	if err != nil {
		t.Fatal(err)
	}
	want = library.Policy{
		ReleaseAt: clock.TimeOfDay{Hour: 19, Minute: 30},
		Filters:   library.Filters{MinDuration: 5 * time.Minute, ExcludedKeywords: []string{"reaction"}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("dopo il salvataggio = %+v, attesa %+v", got, want)
	}
}
