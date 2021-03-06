package signals

// NewMessageNotify sends an empty message into the channel if there is a space available for it.
func NewMessageNotify(c chan struct{}) {
	select {
	case c <- struct{}{}:
	default: // allows non blocking channel usage
	}
}
