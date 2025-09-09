package runner

import "testing"

func TestValidateContainerName(t *testing.T) {
	good := []string{"abc", "a-1_2.3", "Zzz_123"}
	bad := []string{"", "-bad", ".bad", "bad/thing", "a b"}
	for _, n := range good {
		if err := ValidateContainerName(n); err != nil {
			t.Fatalf("expected valid: %q: %v", n, err)
		}
	}
	for _, n := range bad {
		if err := ValidateContainerName(n); err == nil {
			t.Fatalf("expected invalid: %q", n)
		}
	}
}
