package mtid

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

// Memory é um Client em memória, para testes: o "MTCidadão" é simulado por Approve, Reject e Expire.
type Memory struct {
	mu     sync.Mutex
	n      int
	files  map[string][]byte
	orders map[string]*memOrder
	// Err, se definido, faz toda chamada falhar com ele (simula o X-Via fora do ar).
	Err error
	// Now devolve a hora da assinatura; padrão time.Now.
	Now func() time.Time
	// AutoApprove, se positivo, aprova cada pedido sozinho depois desse tempo (simula o signatário no MTCidadão). Usado só no
	// desenvolvimento (MTID_SIMULATE); nos testes a aprovação é explícita, por Approve.
	AutoApprove time.Duration
}

type memOrder struct {
	req    Request
	status Status
}

// NewMemory cria o MT-ID simulado.
func NewMemory() *Memory {
	return &Memory{files: map[string][]byte{}, orders: map[string]*memOrder{}, Now: time.Now}
}

func (m *Memory) id(prefix string) string {
	m.n++
	return fmt.Sprintf("%s-%d", prefix, m.n)
}

func (m *Memory) Upload(_ context.Context, _ string, pdf []byte) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.Err != nil {
		return "", m.Err
	}
	id := m.id("file")
	m.files[id] = append([]byte(nil), pdf...)
	return id, nil
}

func (m *Memory) RequestSignature(_ context.Context, r Request) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.Err != nil {
		return "", m.Err
	}
	if _, ok := m.files[r.FileID]; !ok {
		return "", ErrNotFound
	}
	id := m.id("order")
	m.orders[id] = &memOrder{req: r, status: Status{State: StatePending}}
	if m.AutoApprove > 0 {
		time.AfterFunc(m.AutoApprove, func() { m.Approve(id) })
	}
	return id, nil
}

func (m *Memory) Status(_ context.Context, orderID string) (Status, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.Err != nil {
		return Status{}, m.Err
	}
	o, ok := m.orders[orderID]
	if !ok {
		return Status{}, ErrNotFound
	}
	return o.status, nil
}

func (m *Memory) Cancel(_ context.Context, orderID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.Err != nil {
		return m.Err
	}
	if _, ok := m.orders[orderID]; !ok {
		return ErrNotFound
	}
	delete(m.orders, orderID)
	return nil
}

func (m *Memory) Download(_ context.Context, fileID string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.Err != nil {
		return nil, m.Err
	}
	b, ok := m.files[fileID]
	if !ok {
		return nil, ErrNotFound
	}
	return append([]byte(nil), b...), nil
}

// Approve simula o signatário assinando: o arquivo assinado é o original com um carimbo no final.
// Pedido inexistente (cancelado) ou já decidido não muda.
func (m *Memory) Approve(orderID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	o, ok := m.orders[orderID]
	if !ok || o.status.State != StatePending {
		return
	}
	signed := append(append([]byte(nil), m.files[o.req.FileID]...), []byte("\n%% assinado por "+o.req.SignerCPF+"\n")...)
	id := m.id("file")
	m.files[id] = signed
	o.status = Status{State: StateApproved, SignedAt: m.Now(), FileID: id}
}

// Reject simula a recusa do signatário.
func (m *Memory) Reject(orderID string) { m.set(orderID, StateRejected) }

// Expire simula o vencimento do pedido.
func (m *Memory) Expire(orderID string) { m.set(orderID, StateExpired) }

func (m *Memory) set(orderID string, s State) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.orders[orderID].status = Status{State: s}
}

// Request devolve o pedido feito, para o teste conferir signatário, página e posição.
func (m *Memory) Request(orderID string) (Request, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	o, ok := m.orders[orderID]
	if !ok {
		return Request{}, false
	}
	return o.req, true
}

// Orders devolve quantos pedidos estão abertos (não cancelados).
func (m *Memory) Orders() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.orders)
}

// File devolve o conteúdo de um arquivo enviado ou gerado.
func (m *Memory) File(id string) ([]byte, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	b, ok := m.files[id]
	return b, ok
}

var _ Client = (*Memory)(nil)

// OrderIDs devolve os pedidos abertos (não cancelados), do mais antigo ao mais novo.
func (m *Memory) OrderIDs() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	ids := make([]string, 0, len(m.orders))
	for id := range m.orders {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return orderNum(ids[i]) < orderNum(ids[j]) })
	return ids
}

func orderNum(id string) int {
	var n int
	_, _ = fmt.Sscanf(id, "order-%d", &n)
	return n
}
