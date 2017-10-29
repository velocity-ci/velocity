package websocket

import "log"

type Manager struct {
	clients map[string]*Client
}

func NewManager() *Manager {
	return &Manager{
		clients: map[string]*Client{},
	}
}

func (m *Manager) GetClients() map[string]*Client {
	return m.clients
}

func (m *Manager) Save(c *Client) {
	m.clients[c.ID] = c
}

func (m *Manager) GetClientByID(clientID string) *Client {
	return m.clients[clientID]
}

func (m *Manager) EmitAll(subcription string, m interface{}) {
	for _, c := range m.clients {
		for _, s := range c.subcriptions {
			if s == subcription {
				err := c.ws.WriteJSON(m)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}
}
