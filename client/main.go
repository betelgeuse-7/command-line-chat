package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/gorilla/websocket"
)

func main() {
	host := flag.String("addr", "localhost:8000", "host of the ws server")
	name := flag.String("name", "user_noname", "username")
	flag.Parse()

	wsServer := url.URL{Scheme: "ws", Host: *host, Path: "/ws"}

	conn, res, err := websocket.DefaultDialer.Dial(wsServer.String(), nil)
	if err != nil {
		log.Fatalf("couldn't connect to websocket server: %s\n", err.Error())
	}
	if res.StatusCode == http.StatusUpgradeRequired {
		log.Fatalf("the server did not upgrade to ws protocol\n")
	}

	go handleNewMessage(conn)

	conn.WriteMessage(1, []byte(*name))

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		msg := scanner.Text()
		if len(msg) > 0 {
			err := conn.WriteMessage(1, []byte(scanner.Text()))
			if err != nil {
				fmt.Printf("couldn't send message: %s\n", err.Error())
			}
		}
		fmt.Print("~> ")
	}
}

func handleNewMessage(conn *websocket.Conn) {
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Fatalf("error while reading: %s\n", err.Error())
		}

		greenWriter := color.New(color.FgGreen).Add(color.Underline).Add(color.Bold)
		whiteWriter := color.New(color.FgWhite).Add(color.Bold)
		yellowWriter := color.New(color.FgYellow)

		msgSplitted := strings.Split(string(msg), ": ")
		if len(msgSplitted) == 2 {
			whiteWriter.Print(msgSplitted[0], ": ")
			greenWriter.Print(msgSplitted[1])
			fmt.Println("")
		} else {
			yellowWriter.Println(string(msg))
		}
	}
}
