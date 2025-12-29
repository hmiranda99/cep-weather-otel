package validation

import "testing"

func TestIsValidZipcode(t *testing.T) {
    cases := []struct{
        in string
        ok bool
    }{
        {"01001000", true},
        {"01001-000", false},
        {"abcdef12", false},
        {"1234567", false},
        {"123456789", false},
    }
    for _, c := range cases {
        if IsValidZipcode(c.in) != c.ok {
            t.Fatalf("expected %v for %q", c.ok, c.in)
        }
    }
}
