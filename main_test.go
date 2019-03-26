package main

import (
	"testing"
)

func TestFindRoom(t *testing.T) {
	serv := NewServer()

	cl := &Client{}
	r, err := serv.findRoom(cl)
	if r != nil {
		t.Errorf("expected room to be nil")
	}

	if err == nil {
		t.Errorf("expected error to NOT nil")
	}

	r = serv.createRoom("test_room")
	if r == nil {
		t.Errorf("expected room to be successfully instantiated")
	}

}

func TestJoinRoom(t *testing.T) {
	serv := NewServer()

	cl := &Client{}
	err := serv.joinRoom("test_room", cl)
	if err != nil {
		t.Errorf("expected error to be nil")
	}

	r, err := serv.findRoom(cl)
	if r == nil {
		t.Errorf("expected room to NOT be nil")
	}

}

func TestChangeNick(t *testing.T) {
	const otu = "oldTestUser"
	const ntu = "newTestUser"
	serv := NewServer()

	err := serv.ChangeNick(otu, ntu)
	if err == nil {
		t.Errorf("expected error to NOT be nil")
	}

	cl := &Client{nick: otu}
	err = serv.joinRoom("test_room", cl)
	if err != nil {
		t.Errorf("expected error to be nil")
	}

	if cl.nick != otu {
		t.Errorf("client nick should be [%s]", otu)
	}

	err = serv.ChangeNick(otu, ntu)
	if err != nil {
		t.Errorf("expected error to be nil")
	}

	if cl.nick != ntu {
		t.Errorf("client nick should be [%s]", otu)
	}

}
