package chatserver

import (
	"bufio"
	"fmt"
	"log"
	"net"
//"runtime/debug"
	"runtime/pprof"
	"runtime/debug"
	"context"
	"os"
)

type server struct {
	addr string
	l net.Listener
	errc chan error
	ctx context.Context
	cancel context.CancelFunc
}

func NewChatServer(host string, port string, myCtx context.Context) *server {
	myaddr := host+":"+port
	ser := server{addr: myaddr, errc: make(chan error)}
	ser.ctx, ser.cancel = context.WithCancel(myCtx)
	return &ser
}

func (s *server) Close() {
	s.l.Close()
}
func (s *server) Start() <-chan error {
	go s.mainRoutine()
	return s.errc
}
func (s *server) mainRoutine () {
	var err error
	log.Println("Starting ChatServer...")
	s.l, err = net.Listen("tcp", s.addr)
	if err != nil {
		s.errc <- err
		return
	}
	log.Printf("Listening on %s", s.addr)
	defer  s.endGame()
	go s.broadcaster()
	for {
		conn, err := s.l.Accept()
		if err != nil {
			log.Printf("Accept Error: %s", err.Error())
			s.errc <- err
			return
		}
		go s.handleConn(conn)
		select {
		case <-s.ctx.Done():
			s.errc <-s.ctx.Err()
			return
		default:
		}
	}
}

func (s *server) endGame() {
	s.l.Close()
	pprof.Lookup("goroutine").WriteTo(os.Stdout, 1)
	debug.PrintStack()
}

type client chan<- string

var (
	entering = make(chan client)
	leaving = make (chan client)
	messages = make(chan string)
)

func (s *server) broadcaster() {
	clients := make(map[client]bool)
	for {
		select {
		case cli := <-entering:
			clients[cli] = true

		case cli := <-leaving:
			delete(clients, cli)
			close(cli)

		case msg := <-messages:
			for cli := range clients {
				cli <- msg
			}
		case <-s.ctx.Done():
			fmt.Println("broadcaster: Received Cancel")
			for cli := range clients {
				delete(clients, cli)
				close(cli)
			}
			return
		}
	}
}

func (s *server) handleConn(conn net.Conn) {
	fmt.Printf("connection accepted from %s", conn.RemoteAddr().String())
	ch := make(chan string)
	go s.clientWriter(conn, ch)
	input := bufio.NewScanner(conn)
	who := conn.RemoteAddr().String()
	entering <- ch
	ch <- "you are " + who
	messages <- who + ": "+"has joined"
	for input.Scan() {
		messages <- who + ": " + input.Text()
	}
	messages <- who + ": "+"has left"
	leaving <- ch
}

func (s *server) clientWriter (conn net.Conn, ch <-chan string) {
	defer conn.Close()
	for {
		select {
			case msg := <-ch:
				fmt.Fprintln(conn, msg)
			case <-s.ctx.Done():
				fmt.Println("clientWriter: Received cancel")
				s.errc <- s.ctx.Err()
				return
		}
	}
}


