package main

import (
	"./protector"
	"bufio"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
)

func listenClient(address string) {
	fmt.Println("Launching server...")

	// Listen to all interfaces
	ln, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Println(err)
	}

	var prot *protector.Protector
	for {
		// Accept connections
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
		}

		go handleCon(conn, prot)
	}
}

func dialServer(port string) {
	conn, err := net.Dial("tcp", port)
	for conn == nil || err != nil {
		fmt.Println("Dialing failed with error", err)

		fmt.Println("Retrying to connect")
		conn, err = net.Dial("tcp", port)
	}
	fmt.Println("Dialing to " + port, conn)


	var initialSKey = getKey()
	var initialHash = getHash()

	var sKey string
	var prot = protector.New(initialHash)
	for i := 0; i < 3; i++ {
		var keyAndHash = initialSKey + ":" + initialHash
		if sKey == "" {
			fmt.Println("Text to send: " + keyAndHash)
			fmt.Fprintf(conn, keyAndHash + "\n")
			sKey = initialSKey
		} else {
			sKey = prot.GenSessionKey(sKey)
			fmt.Println("Text to send: " + sKey)
			fmt.Fprintf(conn, sKey + "\n")
		}

		message, _ := bufio.NewReader(conn).ReadString('\n')
		message = message[:len(message) - 1] // Remove '\n' sign
		fmt.Println("Message from server: " + message)

		sKey = prot.GenSessionKey(sKey)
		if message != sKey {
			panic("Keys don't match")
			break
		}
		fmt.Println("Keys are equal")
	}

	conn.Close()
}

func handleCon(conn net.Conn, prot *protector.Protector) {
	for {
		message, _ := bufio.NewReader(conn).ReadString('\n')
		if &message == nil || len(message) == 0 {
			fmt.Println("Connection was closed by client")
			conn.Close()
			return
		}

		message = message[:len(message) - 1] // Remove '\n' sign
		fmt.Println("Message Received: " + string(message))

		var clientSKey, clientHash string
		if strings.Contains(string(message), ":") {
			var keyAndHash = strings.Split(message, ":")
			clientSKey = keyAndHash[0]
			clientHash = keyAndHash[1]

			prot = protector.New(clientHash)
		} else {
			clientSKey = string(message)
		}

		var sKey = prot.GenSessionKey(clientSKey)
		fmt.Println("Text to send: " + sKey)
		fmt.Fprintf(conn, sKey + "\n")
	}
}

func getKey() string {
	var sb strings.Builder
	for i := 0; i < 10; i++ {
		var byte = byte('1' + rand.Intn(9))
		sb.WriteByte(byte)
	}

	return sb.String()
}

func getHash() string {
	var sb strings.Builder
	for i := 0; i < 5; i++ {
		var byte = byte('1' + rand.Intn(6))
		sb.WriteByte(byte)
	}

	return sb.String()
}

func validateIP(ip string) bool {
	if &ip == nil {
		return false
	}

	var splitValues = strings.Split(ip, ".")
	if len(splitValues) != 4 {
		return  false
	}

	for _, value := range splitValues {
		var integer, err = strconv.Atoi(value)
		if err != nil {
			return false
		}

		if (integer < 0) || (integer > math.MaxUint8) {
			return false
		}
	}

	return true
}

func main() {
	portPtr := flag.String("port", ":4200", "app ip*:port")
	instQuanPtr := flag.Uint("inst", 1, "instances quantity")
	flag.Parse()

	var mode string
	var address = strings.Split(string(*portPtr), ":")
	if address[0] != "" {
		var ipValid = validateIP(address[0])
		if !ipValid {
			panic("IP address is not valid")
		}
		mode = "client"
	} else {
		mode = "server"
	}

	if mode == "server" {
		listenClient(*portPtr)
	}

	if mode == "client" {
		var wg sync.WaitGroup
		for i := 0; i < int(*instQuanPtr); i++ {
			wg.Add(1)
			go dialServer(*portPtr)
		}
		wg.Wait()
	}

	fmt.Println("Program exited :(")
}