package sqlvalidator

import "testing"

func TestHasLimitForSelectAddsLimit(t *testing.T) {
	got, added := HasLimitForSelect("SELECT * FROM test")
	want := "SELECT * FROM test LIMIT 100"
	if !added || got != want {
		t.Errorf("expected %q with added=true, got %q and added=%v", want, got, added)
	}
}

func TestHasLimitForSelectWithSemicolon(t *testing.T) {
	got, added := HasLimitForSelect("SELECT * FROM test;")
	want := "SELECT * FROM test LIMIT 100;"
	if !added || got != want {
		t.Errorf("expected %q with added=true, got %q and added=%v", want, got, added)
	}
}

func TestHasLimitForSelectAlreadyHasLimit(t *testing.T) {
	query := "SELECT * FROM test LIMIT 10;"
	got, added := HasLimitForSelect(query)
	if added || got != query {
		t.Errorf("expected original query unchanged, got %q and added=%v", got, added)
	}
}

func TestHasLimitForSelectParameterLimit(t *testing.T) {
	query := "SELECT * FROM test LIMIT ?;"
	got, added := HasLimitForSelect(query)
	if added || got != query {
		t.Errorf("expected original query unchanged, got %q and added=%v", got, added)
	}
}
