package ccworker

import (
	"fmt"

	"github.com/adrianosela/botnet/lib/protocol"
)

// Send enqueues a command to be sent out through this bot's websocket conn
func (b *BotWorker) Send(cmd *protocol.Command) error {
	select {
	case b.cmdOutChan <- cmd:
		return nil
	default:
		return fmt.Errorf("could not send command to remote bot")
	}
}
