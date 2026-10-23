package scan_test

import (
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wagner/convenios-e-fomentos/backend/internal/scan"
)

func listen() (net.Listener, error) {
	var lc net.ListenConfig
	return lc.Listen(context.Background(), "tcp", "127.0.0.1:0")
}

// fakeClamd é um clamd mínimo: lê o comando INSTREAM e os blocos, e responde com reply (ou fecha a conexão se reply for vazio).
type fakeClamd struct {
	addr    string
	got     chan []byte
	command chan string
}

func startFake(t *testing.T, reply func(received []byte) string, maxBytes int) *fakeClamd {
	t.Helper()
	l, err := listen()
	require.NoError(t, err)
	t.Cleanup(func() { _ = l.Close() })
	f := &fakeClamd{addr: l.Addr().String(), got: make(chan []byte, 1), command: make(chan string, 1)}
	go func() {
		conn, err := l.Accept()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		cmd := make([]byte, len("zINSTREAM\x00"))
		if _, err := io.ReadFull(conn, cmd); err != nil {
			return
		}
		f.command <- string(cmd)
		var received []byte
		for {
			var hdr [4]byte
			if _, err := io.ReadFull(conn, hdr[:]); err != nil {
				return
			}
			n := binary.BigEndian.Uint32(hdr[:])
			if n == 0 {
				break
			}
			chunk := make([]byte, n)
			if _, err := io.ReadFull(conn, chunk); err != nil {
				return
			}
			received = append(received, chunk...)
			if maxBytes > 0 && len(received) > maxBytes {
				_, _ = conn.Write([]byte("INSTREAM size limit exceeded. ERROR\x00"))
				return
			}
		}
		f.got <- received
		if r := reply(received); r != "" {
			_, _ = conn.Write([]byte(r + "\x00"))
		}
	}()
	return f
}

func TestClamAVArquivoLimpo(t *testing.T) {
	// Mais de um bloco de 64 KiB: o conteúdo precisa chegar inteiro e em ordem.
	data := bytes.Repeat([]byte("0123456789abcdef"), 20_000) // 320 KB
	f := startFake(t, func([]byte) string { return "stream: OK" }, 0)
	res, err := scan.ClamAV{Addr: f.addr}.Scan(context.Background(), bytes.NewReader(data))
	require.NoError(t, err)
	assert.False(t, res.Infected)
	assert.Equal(t, "zINSTREAM\x00", <-f.command)
	assert.Equal(t, data, <-f.got)
}

func TestClamAVArquivoInfectado(t *testing.T) {
	f := startFake(t, func([]byte) string { return "stream: Eicar-Test-Signature FOUND" }, 0)
	res, err := scan.ClamAV{Addr: f.addr}.Scan(context.Background(), strings.NewReader("X5O!P%@AP"))
	require.NoError(t, err)
	assert.True(t, res.Infected)
	assert.Equal(t, "Eicar-Test-Signature", res.Signature)
}

func TestClamAVArquivoVazioEOK(t *testing.T) {
	f := startFake(t, func([]byte) string { return "stream: OK" }, 0)
	res, err := scan.ClamAV{Addr: f.addr}.Scan(context.Background(), strings.NewReader(""))
	require.NoError(t, err)
	assert.False(t, res.Infected)
	assert.Empty(t, <-f.got)
}

func TestClamAVFalhaFechada(t *testing.T) {
	t.Run("serviço fora do ar", func(t *testing.T) {
		l, err := listen()
		require.NoError(t, err)
		addr := l.Addr().String()
		require.NoError(t, l.Close())
		_, err = scan.ClamAV{Addr: addr, Timeout: time.Second}.Scan(context.Background(), strings.NewReader("x"))
		assert.ErrorIs(t, err, scan.ErrScan)
	})
	t.Run("resposta de erro do clamd", func(t *testing.T) {
		f := startFake(t, func([]byte) string { return "stream: Can't allocate memory ERROR" }, 0)
		_, err := scan.ClamAV{Addr: f.addr}.Scan(context.Background(), strings.NewReader("x"))
		assert.ErrorIs(t, err, scan.ErrScan, "erro do clamd nunca vira 'limpo'")
	})
	t.Run("limite de tamanho do clamd", func(t *testing.T) {
		f := startFake(t, func([]byte) string { return "" }, 100_000)
		_, err := scan.ClamAV{Addr: f.addr}.Scan(context.Background(), bytes.NewReader(make([]byte, 400_000)))
		assert.ErrorIs(t, err, scan.ErrScan)
	})
	t.Run("conexão fechada sem resposta", func(t *testing.T) {
		f := startFake(t, func([]byte) string { return "" }, 0)
		_, err := scan.ClamAV{Addr: f.addr}.Scan(context.Background(), strings.NewReader("x"))
		assert.ErrorIs(t, err, scan.ErrScan)
	})
	t.Run("resposta inesperada", func(t *testing.T) {
		f := startFake(t, func([]byte) string { return "lixo" }, 0)
		_, err := scan.ClamAV{Addr: f.addr}.Scan(context.Background(), strings.NewReader("x"))
		assert.ErrorIs(t, err, scan.ErrScan)
	})
	t.Run("sem resposta dentro do prazo", func(t *testing.T) {
		l, err := listen()
		require.NoError(t, err)
		t.Cleanup(func() { _ = l.Close() })
		go func() {
			c, err := l.Accept()
			if err == nil {
				time.Sleep(2 * time.Second)
				_ = c.Close()
			}
		}()
		start := time.Now()
		_, err = scan.ClamAV{Addr: l.Addr().String(), Timeout: 300 * time.Millisecond}.Scan(context.Background(), strings.NewReader("x"))
		assert.ErrorIs(t, err, scan.ErrScan)
		assert.Less(t, time.Since(start), 1500*time.Millisecond)
	})
	t.Run("contexto cancelado", func(t *testing.T) {
		l, err := listen()
		require.NoError(t, err)
		t.Cleanup(func() { _ = l.Close() })
		go func() {
			c, err := l.Accept()
			if err == nil {
				time.Sleep(3 * time.Second)
				_ = c.Close()
			}
		}()
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()
		start := time.Now()
		_, err = scan.ClamAV{Addr: l.Addr().String(), Timeout: 10 * time.Second}.Scan(ctx, strings.NewReader("x"))
		require.Error(t, err)
		assert.Less(t, time.Since(start), 2*time.Second)
	})
}
