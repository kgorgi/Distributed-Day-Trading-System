package quotemock

import (
	"bufio"
	"net"
)

func main() {

	ln, _ := net.Listen("tcp", ":4443")

	for {
		conn, _ := ln.Accept()
		bufio.NewReader(conn).ReadString('\n')
		conn.Write([]byte("11.52,ABC,quoteMock,123456,KEY\n"))
		conn.Close()
	}
}
