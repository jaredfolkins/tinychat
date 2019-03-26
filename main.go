package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

const logName = "tinychat.log"
const DefaultRoom = "Gotham City"

// banner is a const displayed to the user as the connect to the system
const banner = `
--|Welcome|--------------------------------------------------------------------------------------

You are user [%s], Welcome to TinyChat. 

--|Help|-----------------------------------------------------------------------------------------

{no flag needed}
send a message to the room you are in
(example: hi freeze, i'm batman)

/help
prints this banner
(example: /help) 

/quit
quits the application
(example: /quit) 

/nick
sets your nickname
(example: /nick batman)

/room 
change chat room, only 1 room may be joined
(example: /room gotham)

/blast
blast a message to all connected clients 
(example: /blast the ice man cometh)

-------------------------------------------------------------------------------------------------
`

// helper logging function
func errl(err error, message string) {
	if err != nil {
		log.Println(err)
	} else {
		log.Printf("%s\n", message)
	}
}

// Client is a structure keeping the state of the user connected to the server
type Client struct {
	mu   sync.Mutex
	nick string
	Conn net.Conn
}

// Nick returns the nickname of the client
func (cl *Client) Nick() string {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	return cl.nick
}

// Write writes the output to a client
func (cl *Client) Write(s string) {
	cl.mu.Lock()
	defer cl.mu.Unlock()
	cl.Conn.Write([]byte(s))
}

// Serv is a pointer to our Server instance
var Serv *Server

// Server is the struct that keeps the state of the entire application
type Server struct {
	mu      sync.Mutex
	Rooms   map[string]*Room
	Clients map[string]*Client
}

// Room is the data strucutre used for a Chat Room, it keeps a map of all connected clients
type Room struct {
	mu      sync.Mutex
	Clients map[string]*Client
}

// CloseClient accpets a client pointer, closes the connection, and deletes it from the Clients map
func (s *Server) CloseClient(cl *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cl.Conn.Close()
	delete(s.Clients, cl.Nick())
}

// ChangeNick valides if the nick is in use
// if it isn't then the client's nickname is allowed to be changed
func (s *Server) ChangeNick(from, to string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// if the name we are changing TO exists, error
	if s.clientExists(to) {
		e := errors.New(fmt.Sprintf("user [%s] already exists\r\n", to))
		errl(e, "user already exists")
		return e
	}

	// the client should exist
	if s.clientExists(from) {
		// if the name we are changing FROM exists, proceed
		cl := s.Clients[from]
		r, err := s.findRoom(cl)
		if err != nil {
			errl(err, "Found a room!")
			return err
		}

		delete(r.Clients, from)
		delete(s.Clients, from)
		cl.nick = to
		r.Clients[to] = cl
		s.Clients[to] = cl
	} else {
		e := errors.New(fmt.Sprintf("user [%s] does not exists\r\n", to))
		errl(e, "user does not exists")
		return e
	}

	return nil
}

// Message sends the message to only the room the client is attached to
func (s *Server) Message(inputs []string, cl *Client) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	msg := fmt.Sprintf("[%s:%s]", time.Now().Format(time.RFC3339), cl.Nick())
	for _, v := range inputs {
		msg = fmt.Sprintf("%s %s", msg, v)
	}
	msg = msg + "\r\n"

	r, err := s.findRoom(cl)
	if err != nil {
		return err
	}

	if r != nil {
		for _, c := range r.Clients {
			c.Write(strings.TrimSpace(msg) + "\r\n")
		}
	}
	return nil
}

// Blast sends a message to every client connected to the server
// example: servide will be stopped for service in 45 minutes
func (s *Server) Blast(inputs []string, cl *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	msg := fmt.Sprintf("[%s:%s]", time.Now().Format(time.RFC3339), cl.Nick())
	for _, v := range inputs[1:] {
		msg = fmt.Sprintf("%s %s", msg, v)
	}
	msg = msg + "\r\n"

	for _, c := range s.Clients {
		c.Write(strings.TrimSpace(msg) + "\r\n")
	}
}

// JoinRoom is a public function for joining the room
func (s *Server) JoinRoom(roomname string, cl *Client) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.tryDeleteFromRoom(cl)

	err := s.joinRoom(roomname, cl)
	if err != nil {
		return err
	}

	return nil
}

// clientExists returns true if the client is found in the Server's Clients map
func (s *Server) clientExists(nick string) bool {
	if _, ok := s.Clients[nick]; ok {
		return ok
	}
	return false
}

// roomExists returns true if the room is found in the Server's Rooms map
func (s *Server) roomExists(roomname string) bool {
	if _, ok := s.Rooms[roomname]; ok {
		return ok
	}
	return false
}

// addClient accpets accepts a client and adds it to the Server's Client map
func (s *Server) addClient(cl *Client) error {
	if !s.clientExists(cl.Nick()) {
		s.Clients[cl.Nick()] = cl
		return nil
	}

	return errors.New("Client already exists")
}

func (s *Server) createRoom(roomname string) *Room {
	r := &Room{
		Clients: make(map[string]*Client),
	}
	s.Rooms[roomname] = r
	return r
}

// joinRoom is a helper function that doesn't lock
func (s *Server) joinRoom(roomname string, cl *Client) error {
	var r *Room
	if !s.roomExists(roomname) {
		r = s.createRoom(roomname)
	} else {
		r = s.Rooms[roomname]
	}

	r.Clients[cl.Nick()] = cl
	err := s.addClient(cl)
	if err != nil {
		return err
	}
	return nil
}

// tryDeleteFromRoom will scan all the rooms and delete any reference of the client from them
func (s *Server) tryDeleteFromRoom(cl *Client) {
	r, _ := s.findRoom(cl)
	if r != nil {
		delete(r.Clients, cl.Nick())
	}
}

// findRoom scans the rooms in the server instance for the client
func (s *Server) findRoom(cl *Client) (*Room, error) {
	for _, r := range s.Rooms {
		if _, ok := r.Clients[cl.Nick()]; ok {
			return r, nil
		}
	}
	st := fmt.Sprintf("%s does not have a room", cl.Nick())
	return nil, errors.New(st)
}

// clientRun is the method that a client runs while it waits for, and then processes, input
func clientRun(cl *Client, buf *bufio.Reader) {
	for {

		cmd, err := buf.ReadString('\n')
		if err != nil {
			fmt.Printf("Client disconnected.\n")
			break
		}

		// split up the inputs
		inputs := strings.Fields(cmd)

		// if command is empty, do not process
		if len(inputs) == 0 {
			cl.Write("Command not recognized\r\n")
		} else {
			switch inputs[0] {
			case "/help":
				out := fmt.Sprintf(banner, cl.Nick())
				cl.Write(out)
			case "/quit":
				Serv.CloseClient(cl)
			case "/blast":
				Serv.Blast(inputs, cl)
			case "/room":
				if len(inputs) >= 2 {
					var roomname string
					for _, v := range inputs[1:] {
						roomname = fmt.Sprintf("%s%s", roomname, v)
					}
					Serv.JoinRoom(strings.ToLower(roomname), cl)
					resp := fmt.Sprintf("Joining room %s\r\n", strings.ToLower(roomname))
					cl.Write(resp)
				} else {
					resp := fmt.Sprintf("Unable to join room\r\n")
					cl.Write(resp)
				}
			case "/nick":
				if len(inputs) >= 2 {
					from := cl.Nick()
					to := inputs[1]
					err := Serv.ChangeNick(from, to)
					resp := fmt.Sprintf("Nick changed from [%s] to [%s]\r\n", from, to)
					if err != nil {
						cl.Write(err.Error())
					} else {
						cl.Write(resp)
					}
				} else {
					resp := fmt.Sprintf("Nick unchanged and is currently [%s] \r\n", cl.Nick())
					cl.Write(resp)
				}
			default:
				err := Serv.Message(inputs, cl)
				errl(err, "Message sent to room successfully")
			}
		}

	}
}

// initClient is a helper function that sets up the client
// TODO handle the errors, derp
func initClient(conn net.Conn) {
	buf := bufio.NewReader(conn)
	uname := fmt.Sprintf("%s%d", "user", time.Now().UnixNano())
	cl := &Client{nick: uname, Conn: conn}
	err := Serv.JoinRoom(DefaultRoom, cl)
	errl(err, "Joined room")
	cl.Write(fmt.Sprintf(banner, uname))
	clientRun(cl, buf)
}

func NewServer() *Server {
	return &Server{
		Clients: make(map[string]*Client),
		Rooms:   make(map[string]*Room),
	}

}
func main() {
	// working directory
	cwd, err := os.Getwd()
	if err != nil {
		panic("unable to detect current working directory")
	}

	// env variables
	tcLog := os.Getenv("TCLogPath")
	if len(tcLog) == 0 {
		tcLog = path.Join(cwd, logName)
	} else {
		tcLog = path.Join(tcLog, logName)
	}

	tcPort := os.Getenv("TCPort")
	if len(tcPort) == 0 {
		tcPort = "8091"
	}

	tcHost := os.Getenv("TCHost")
	if len(tcHost) == 0 {
		tcHost = "localhost"
	}

	// logfile
	f, err := os.OpenFile(tcLog, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)
	log.Printf("Application Starting %s\n", time.Now().Format(time.RFC3339))

	// instantiate server
	Serv = NewServer()

	uri := fmt.Sprintf("%s:%s", tcHost, tcPort)
	ln, err := net.Listen("tcp", uri)
	errl(err, "Server is ready.")

	for {
		conn, err := ln.Accept()
		errl(err, "Client connected successfully")
		go initClient(conn)
	}
}
