package enduser

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeProvider struct{ name string }

func (f fakeProvider) Name() string { return f.name }
func (f fakeProvider) Authenticate(_ context.Context, c map[string]string) (Identity, error) {
	return Identity{Provider: f.name, Subject: c["code"], EmailVerified: true}, nil
}

func TestRegistroDeProvedores(t *testing.T) {
	p := NewProviders()
	require.NoError(t, p.Register(fakeProvider{"github"}))
	assert.Error(t, p.Register(fakeProvider{"github"}), "duplicado")
	assert.Error(t, p.Register(fakeProvider{"local"}), "local é reservado")
	assert.Error(t, p.Register(fakeProvider{""}), "sem nome")

	got, ok := p.Get("github")
	require.True(t, ok)
	id, err := got.Authenticate(context.Background(), map[string]string{"code": "abc"})
	require.NoError(t, err)
	assert.Equal(t, Identity{Provider: "github", Subject: "abc", EmailVerified: true}, id)
	_, ok = p.Get("outro")
	assert.False(t, ok)
}
