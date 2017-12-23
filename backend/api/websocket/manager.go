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

func (m *Manager) Remove(c *Client) {
	delete(m.clients, c.ID)
}

func (m *Manager) GetClientByID(clientID string) *Client {
	return m.clients[clientID]
}

func (m *Manager) EmitAll(message *PhoenixMessage) {
	clientCount := 0
	for _, c := range m.clients {
		for _, s := range c.subscriptions {
			if s == message.Topic {
				err := c.ws.WriteJSON(message)
				clientCount++
				if err != nil {
					log.Println(err)
				}
			}
		}
	}
	log.Printf("Emitted %s to %d clients", message.Topic, clientCount)
}
