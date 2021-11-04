package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

func main() {
	server := newServer()
	server.setUp()
	server.start()
}

type Port string

func (p *Port) set(port string) {
	*p = Port(port)
}
func (p Port) get() string {
	return string(p)
}

type Server struct {
	config struct {
		port Port
	}
	clients []Client
}

func newServer() *Server {
	return &Server{}
}

func (s *Server) setUp() {
	port := flag.String("port", ":8000", "port for the server to listen to")
	flag.Parse()

	if !strings.HasPrefix(*port, ":") {
		*port = ":" + *port
	}
	s.config.port.set(*port)
}

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (s *Server) start() {
	fmt.Println("------------------------")
	log.Printf("Server started listening on localhost%s\n", s.config.port)
	fmt.Println("------------------------")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello \n")
	})
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, "couldn't upgrade", http.StatusUpgradeRequired)
			return
		}
		client := newClient(conn)
		client.giveRandomId()
		s.clients = append(s.clients, *client)
		go s.handleWsConn(client)
	})

	http.ListenAndServe(s.config.port.get(), nil)
}

func (s *Server) handleWsConn(client *Client) {
	defer client.WsConn.Close()
	for {
		_, msg, err := client.WsConn.ReadMessage()
		if err != nil {
			serverLog := fmt.Sprintf("%s#%s disconnected", client.Name, client.Id)
			log.Println(serverLog)
			s.broadcastMessage(serverLog, client.Id)
			break
		}
		if len(client.Name) == 0 {
			client.Name = string(msg)
			continue
		}
		serverLog := fmt.Sprintf("%s#%s said: %s\n", client.Name, client.Id, string(msg))
		log.Println(serverLog)
		s.broadcastMessage(serverLog, client.Id)
	}
}

func (s *Server) broadcastMessage(msg, exceptUserWithId string) {
	for _, c := range s.clients {
		if c.Id != exceptUserWithId {
			c.WsConn.WriteMessage(1, []byte(msg))
		}
	}
}

type Client struct {
	Id     string
	Name   string
	WsConn *websocket.Conn
}

func newClient(conn *websocket.Conn) *Client {
	return &Client{
		WsConn: conn,
	}
}

func (c *Client) giveRandomId() {
	id := uuid.New()
	c.Id = id.String()
}
