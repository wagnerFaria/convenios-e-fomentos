package scan

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

const chunkSize = 64 << 10

// ErrScan indica que a verificação não pôde ser feita (o arquivo não foi considerado limpo).
var ErrScan = errors.New("scan: verificação indisponível")

// ClamAV fala com o clamd pelo protocolo INSTREAM (TCP). O limite de tamanho do clamd (StreamMaxLength, padrão 25 MB)
// precisa cobrir o maior arquivo aceito (200 MB); acima do limite, o clamd recusa e a verificação falha (fail-closed).
type ClamAV struct {
	// Addr é host:porta do clamd.
	Addr string
	// Timeout vale para a conexão e para cada leitura/escrita (padrão 60 s).
	Timeout time.Duration
}

func (c ClamAV) timeout() time.Duration {
	if c.Timeout > 0 {
		return c.Timeout
	}
	return 60 * time.Second
}

// Scan envia o conteúdo em blocos e interpreta a resposta do clamd.
func (c ClamAV) Scan(ctx context.Context, r io.Reader) (Result, error) {
	d := net.Dialer{Timeout: c.timeout()}
	conn, err := d.DialContext(ctx, "tcp", c.Addr)
	if err != nil {
		return Result{}, fmt.Errorf("%w: %w", ErrScan, err)
	}
	defer func() { _ = conn.Close() }()
	// Cancelar o contexto interrompe a verificação em andamento.
	stop := context.AfterFunc(ctx, func() { _ = conn.Close() })
	defer stop()

	arm := func() { _ = conn.SetDeadline(time.Now().Add(c.timeout())) }
	arm()
	if _, err := conn.Write([]byte("zINSTREAM\x00")); err != nil {
		return Result{}, fmt.Errorf("%w: %w", ErrScan, err)
	}
	buf := make([]byte, chunkSize)
	var hdr [4]byte
	for {
		n, rerr := r.Read(buf)
		if n > 0 {
			binary.BigEndian.PutUint32(hdr[:], uint32(n)) //nolint:gosec // n <= chunkSize
			arm()
			if _, err := conn.Write(append(hdr[:], buf[:n]...)); err != nil {
				// O clamd pode ter respondido (ex.: limite de tamanho) e fechado: a resposta explica.
				if res, perr := readReply(conn); perr != nil || res.Infected {
					return res, perr
				}
				return Result{}, fmt.Errorf("%w: %w", ErrScan, err)
			}
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			return Result{}, fmt.Errorf("lendo o arquivo: %w", rerr)
		}
	}
	arm()
	if _, err := conn.Write([]byte{0, 0, 0, 0}); err != nil {
		return Result{}, fmt.Errorf("%w: %w", ErrScan, err)
	}
	return readReply(conn)
}

// readReply lê a resposta "stream: OK", "stream: <assinatura> FOUND" ou "... ERROR".
func readReply(conn net.Conn) (Result, error) {
	b, err := io.ReadAll(io.LimitReader(conn, 4096))
	if err != nil && len(b) == 0 {
		return Result{}, fmt.Errorf("%w: %w", ErrScan, err)
	}
	reply := strings.TrimRight(strings.TrimSpace(string(b)), "\x00")
	switch {
	case strings.HasSuffix(reply, " OK"):
		return Result{}, nil
	case strings.HasSuffix(reply, " FOUND"):
		sig := strings.TrimSuffix(strings.TrimPrefix(reply, "stream: "), " FOUND")
		return Result{Infected: true, Signature: sig}, nil
	default:
		return Result{}, fmt.Errorf("%w: resposta do clamd: %q", ErrScan, reply)
	}
}
