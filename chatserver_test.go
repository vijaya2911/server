package chatserver

import (
	"context"
	"net"
	"log"
	"time"
	"testing"
)

func expect(t *testing.T, ser interface{}) {
	if ser == nil {
		t.Errorf("Expected ChatServer reference, got nil")
	}
}
func clientTest(t *testing.T) {
	t.Logf("Starting client...")
	time.Sleep(1 * time.Second)
	conn, err := net.Dial("tcp","localhost:5555")
	if err != nil {
		t.Errorf("Expected Connection Success, got error: %s", err.Error())
	}
	log.Println("Client connected...")
	conn.Close()
}


func TestStart(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ser := NewChatServer("localhost", "5555", ctx)
	expect (t, ser)
	defer ser.Close()
	errc := ser.Start()
	clientTest(t)
	cancel()
	<-errc
}
