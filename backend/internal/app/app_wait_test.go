package app_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Sem o worker de e-mail (como nos testes), esperar o encerramento volta na hora.
func TestEsperarOEncerramentoSemWorker(t *testing.T) {
	e := newEnv(t)
	begin := time.Now()
	assert.True(t, e.app.Wait(time.Minute))
	assert.Less(t, time.Since(begin), time.Second)
}
