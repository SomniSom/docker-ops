package locale

import "testing"

func TestParseLocaleToken(t *testing.T) {
	tests := map[string]Code{
		"ru_RU.UTF-8": Ru,
		"ru":          Ru,
		"ru:en":       Ru,
		"en_US.UTF-8": En,
		"C":           En,
		"C.UTF-8":     En,
	}
	for in, want := range tests {
		if got := parseLocaleToken(in); got != want {
			t.Errorf("parseLocaleToken(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestBootstrapDQLang(t *testing.T) {
	t.Setenv("DQ_LANG", "ru")
	t.Setenv("LANG", "en_US.UTF-8")
	BootstrapFromArgs(nil)
	if Current() != Ru {
		t.Fatalf("expected Ru, got %v", Current())
	}
	Set("en")
}

func TestBootstrapArgsLang(t *testing.T) {
	t.Setenv("DQ_LANG", "")
	BootstrapFromArgs([]string{"validate", "--lang", "ru"})
	if Current() != Ru {
		t.Fatalf("expected Ru from argv, got %v", Current())
	}
	Set("en")
}

func TestTUnknownKey(t *testing.T) {
	Set("en")
	if got := T("no.such.key"); got != "no.such.key" {
		t.Fatalf("got %q", got)
	}
}

func TestCatalogParity(t *testing.T) {
	for k := range catalogEn {
		if _, ok := catalogRu[k]; !ok {
			t.Errorf("catalogRu missing key %q", k)
		}
	}
}

func TestRussianStrings(t *testing.T) {
	Set("ru")
	if T("root.short") == catalogEn["root.short"] {
		t.Fatal("expected Russian root.short")
	}
	Set("en")
}
