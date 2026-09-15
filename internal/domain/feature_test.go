package domain_test

import (
	"testing"

	"feature-flag-manager/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestStatusValid(t *testing.T) {
	for _, status := range []domain.Status{domain.StatusOpen, domain.StatusClosed, domain.StatusWhitelisted} {
		assert.True(t, status.Valid())
	}
	assert.False(t, domain.Status("unknown").Valid())
}
