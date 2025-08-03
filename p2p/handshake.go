package p2p

// var ErrInvalidHandShake = errors.New("Invalid handshake")

type HandShakeFunc func(Peer) error

func NOHandShake(Peer) error {
	return nil
}
