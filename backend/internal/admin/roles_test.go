package admin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEscapeLike(t *testing.T) {
	assert.Equal(t, `100\%\_\\`, escapeLike(`100%_\`))
}
