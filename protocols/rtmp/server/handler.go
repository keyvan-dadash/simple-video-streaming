package server

//ConnHandler is structure that handle Conn
type ConnHandler struct {
	Conn *Conn
}

//NewConnHadler return handler for handle Conn
func NewConnHadler(c *Conn) *ConnHandler {
	return &ConnHandler{
		Conn: c,
	}
}

//Handle the Conn
func (ch *ConnHandler) Handle() {
	handshakeHandler := NewHandShakeConn(ch.Conn)
	handshakeHandler.HandShake()
}
