package mtid_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wagner/convenios-e-fomentos/backend/internal/mtid"
)

func request(t *testing.T, m *mtid.Memory) string {
	t.Helper()
	ctx := context.Background()
	file, err := m.Upload(ctx, "x.pdf", []byte("%PDF-1.7 teste"))
	require.NoError(t, err)
	order, err := m.RequestSignature(ctx, mtid.Request{SignerCPF: "12345678909", FileID: file})
	require.NoError(t, err)
	return order
}

func TestAprovacaoAutomaticaSimulaOSignatario(t *testing.T) {
	m := mtid.NewMemory()
	m.AutoApprove = 20 * time.Millisecond
	order := request(t, m)

	st, err := m.Status(context.Background(), order)
	require.NoError(t, err)
	assert.Equal(t, mtid.StatePending, st.State, "ainda não assinou")

	require.Eventually(t, func() bool {
		st, _ := m.Status(context.Background(), order)
		return st.State == mtid.StateApproved
	}, time.Second, 10*time.Millisecond)
	st, _ = m.Status(context.Background(), order)
	signed, err := m.Download(context.Background(), st.FileID)
	require.NoError(t, err)
	assert.Contains(t, string(signed), "%PDF-1.7 teste")
	assert.Contains(t, string(signed), "12345678909")
}

func TestPedidoCanceladoNaoE_AprovadoDepois(t *testing.T) {
	m := mtid.NewMemory()
	m.AutoApprove = 20 * time.Millisecond
	order := request(t, m)
	require.NoError(t, m.Cancel(context.Background(), order))

	time.Sleep(60 * time.Millisecond) // a aprovação agendada roda e não pode derrubar nem recriar o pedido
	assert.Equal(t, 0, m.Orders())
	_, err := m.Status(context.Background(), order)
	assert.ErrorIs(t, err, mtid.ErrNotFound)
}

func TestSemAprovacaoAutomaticaOPedidoFicaPendente(t *testing.T) {
	m := mtid.NewMemory()
	order := request(t, m)
	time.Sleep(30 * time.Millisecond)
	st, _ := m.Status(context.Background(), order)
	assert.Equal(t, mtid.StatePending, st.State)
}
