package replication

import (
	"net"

	"github.com/codecrafters-io/redis-starter-go/internal/resp"
)

type Manager struct {
	replicas []net.Conn
}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) AddReplica(conn net.Conn) {
	m.replicas = append(m.replicas, conn)
}

func (m *Manager) Propagate(args []string) {
	encoded := []byte(resp.Array(args))

	for _, replica := range m.replicas {
		replica.Write(encoded)
	}
}
