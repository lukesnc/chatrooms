package main

import (
	"bufio"
	"fmt"
	"net"
	"slices"
	"strconv"
	"strings"
	"time"
)

type User struct {
	name string
	conn net.Conn
}

type Room struct {
	topic   string
	members []User
	newMsg  *Message
}

type Message struct {
	sender string
	body   string
}

func main() {
	ln, err := net.Listen("tcp", ":9001")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Listening on", ln.Addr().String())

	// Create state
	rooms := []Room{}

	rooms = append(rooms, Room{topic: "random", members: []User{}, newMsg: nil})
	rooms = append(rooms, Room{topic: "games", members: []User{}, newMsg: nil})
	rooms = append(rooms, Room{topic: "nihongo", members: []User{}, newMsg: nil})

	// Take connections
	go serveMessages(&rooms)
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}

		fmt.Println("Connection from", conn.RemoteAddr().String())
		go handleConn(conn, &rooms)
	}
}

func serveMessages(rooms *[]Room) {
	for {
		for i := range len(*rooms) {
			if (*rooms)[i].newMsg == nil {
				continue
			}
			for _, user := range (*rooms)[i].members {
				msg := fmt.Sprintf("%s: %s\n", (*rooms)[i].newMsg.sender, (*rooms)[i].newMsg.body)
				user.conn.Write([]byte(msg))
			}
			(*rooms)[i].newMsg = nil
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func handleConn(conn net.Conn, rooms *[]Room) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	// Get username
	conn.Write([]byte("Enter a username: "))
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	// Session vars
	user := User{name: username, conn: conn}
	var currentRoom int = -1

	// Welcome msg
	welcomeMsg := "Welcome! Available rooms are:\n"
	for i, room := range *rooms {
		welcomeMsg += fmt.Sprintf("%d: %s\n", i+1, room.topic)
	}
	welcomeMsg += "Available commands are: join, say, leave\n"
	conn.Write([]byte(welcomeMsg))

	// Take input loop
	for {
		// conn.Write([]byte("> "))
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		parts := strings.Split(input, " ")

		switch parts[0] {
		case "join", "enter":
			num, err := strconv.Atoi(parts[1])
			if currentRoom != -1 {
				conn.Write([]byte("already in a room\n"))
				continue
			} else if err != nil || num > len(*rooms) {
				conn.Write([]byte("invalid room\n"))
				continue
			}

			num -= 1
			conn.Write([]byte(fmt.Sprintf("Joined %s\n", (*rooms)[num].topic)))
			(*rooms)[num].members = append((*rooms)[num].members, user)
			currentRoom = num
		case "say", "send":
			if currentRoom == -1 {
				conn.Write([]byte("not in a room\n"))
				continue
			}

			msg := Message{sender: user.name, body: strings.Join(parts[1:], " ")}
			(*rooms)[currentRoom].newMsg = &msg
		case "leave", "exit", "quit":
			// Leave room first, then leave server
			if currentRoom != -1 {
				conn.Write([]byte(fmt.Sprintf("Left %s\n", (*rooms)[currentRoom].topic)))

				// Remove user from current room
				(*rooms)[currentRoom].members = slices.DeleteFunc((*rooms)[currentRoom].members, func(u User) bool {
					return user.name == u.name
				})

				currentRoom = -1
			} else {
				conn.Write([]byte("Bye bye\n"))
				return
			}
		case "":
			continue
		default:
			conn.Write([]byte("unknown command\n"))
			continue
		}
	}
}
