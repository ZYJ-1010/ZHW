package games

import (
	"errors"
	"testing"

	"github.com/lib/pq"
)

func TestIsUniqueViolation(t *testing.T) {
	if !isUniqueViolation(&pq.Error{Code: "23505"}) {
		t.Fatal("PostgreSQL unique violation must map to a business conflict")
	}
	if isUniqueViolation(errors.New("other database error")) {
		t.Fatal("non-unique database errors must not be remapped")
	}
}
