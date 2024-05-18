package main

import (
	"bufio"
	"fmt"
	"net"
	"slices"
	"strconv"
	"strings"
)

type User struct {
	name string
	conn net.Conn
}

type Room struct {
	topic   string
	members []User
	new_msg *Message
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
	fmt.Println("Listening on localhost:9001 ...")

	// Create state
	rooms := []Room{}

	rooms = append(rooms, Room{topic: "random", members: []User{}, new_msg: nil})
	rooms = append(rooms, Room{topic: "games", members: []User{}, new_msg: nil})
	rooms = append(rooms, Room{topic: "nihongo", members: []User{}, new_msg: nil})

	// Take connections
	go serve_messages(&rooms)
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}

		go handle_conn(conn, &rooms)
	}
}

func serve_messages(rooms *[]Room) {
	for {
		for i := range len(*rooms) {
			if (*rooms)[i].new_msg == nil {
				continue
			}
			for _, user := range (*rooms)[i].members {
				msg := fmt.Sprintf("%s: %s\n", (*rooms)[i].new_msg.sender, (*rooms)[i].new_msg.body)
				user.conn.Write([]byte(msg))
			}
			(*rooms)[i].new_msg = nil
		}
	}
}

func handle_conn(conn net.Conn, rooms *[]Room) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	// Get username
	conn.Write([]byte("Enter a username: "))
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	// Session vars
	user := User{name: username, conn: conn}
	var current_room int = -1

	// Welcome msg
	conn.Write([]byte("Welcome! Available rooms are:\n"))
	for i, room := range *rooms {
		conn.Write([]byte(fmt.Sprintf("%d: %s\n", i+1, room.topic)))
	}
	conn.Write([]byte("Available commands are: join, say, leave\n"))

	// Take input loop
	for {
		// conn.Write([]byte("> "))
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		parts := strings.Split(input, " ")

		switch parts[0] {
		case "join", "enter":
			num, err := strconv.Atoi(parts[1])
			if current_room != -1 {
				conn.Write([]byte("already in a room\n"))
				continue
			} else if err != nil || num > len(*rooms) {
				conn.Write([]byte("invalid room\n"))
				continue
			}

			num -= 1
			conn.Write([]byte(fmt.Sprintf("Joining %s\n", (*rooms)[num].topic)))
			(*rooms)[num].members = append((*rooms)[num].members, user)
			current_room = num
		case "say", "send":
			if current_room == -1 {
				conn.Write([]byte("not in a room\n"))
				continue
			}

			msg := Message{sender: user.name, body: strings.Join(parts[1:], " ")}
			(*rooms)[current_room].new_msg = &msg
		case "leave", "exit", "quit":
			// Leave room first, then leave server
			if current_room != -1 {
				conn.Write([]byte(fmt.Sprintf("Leaving %s\n", (*rooms)[current_room].topic)))

				// Remove user from current room
				(*rooms)[current_room].members = slices.DeleteFunc((*rooms)[current_room].members, func(u User) bool {
					return user.name == u.name
				})

				current_room = -1
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
