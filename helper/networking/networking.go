package networking

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
)

type AsyncTcpServer struct {
	Address    string
	Port       int
	allSockets map[net.Conn]int
	mutex      sync.Mutex
}

func NewAsyncTcpServer(address string, port int) *AsyncTcpServer {
	return &AsyncTcpServer{
		Address:    address,
		Port:       port,
		allSockets: make(map[net.Conn]int),
	}
}

func (server *AsyncTcpServer) GetConnections() map[net.Conn]int {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	return server.allSockets
}

func (server *AsyncTcpServer) GetConnection(socketID int) net.Conn {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	for conn, id := range server.allSockets {
		if id == socketID {
			return conn
		}
	}
	return nil
}

func (server *AsyncTcpServer) Listen() {
	address := fmt.Sprintf("%s:%d", server.Address, server.Port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Println("Error creating listener:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Server is listening on", address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		go server.handleConnection(conn)
	}
}

func (server *AsyncTcpServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	fmt.Println("New connection from", conn.RemoteAddr())
	server.mutex.Lock()
	socketID := GetSocketID(conn)
	server.allSockets[conn] = socketID
	server.mutex.Unlock()

	for {
		data := make([]byte, 1024)
		n, err := conn.Read(data)
		if err != nil {
			fmt.Println("Error reading from connection:", err)
			break
		}

		fmt.Printf("Received from %s: %s\n", conn.RemoteAddr(), string(data[:n]))
	}

	server.mutex.Lock()
	delete(server.allSockets, conn)
	server.mutex.Unlock()
	fmt.Println("Connection closed from", conn.RemoteAddr())
}

func (server *AsyncTcpServer) Close() {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	for conn := range server.allSockets {
		_ = conn.Close()
	}
	server.allSockets = make(map[net.Conn]int)
}

func (server *AsyncTcpServer) Send(conn net.Conn, data []byte) error {
	_, err := conn.Write(data)
	return err
}

func (server *AsyncTcpServer) Receive(conn net.Conn) ([]byte, error) {
	data := make([]byte, 1024)
	n, err := conn.Read(data)
	if err != nil {
		return nil, err
	}
	return data[:n], nil
}

type AsyncTcpClient struct {
	Address string
	Port    int
	conn    net.Conn
}

func NewAsyncTcpClient(address string, port int) *AsyncTcpClient {
	return &AsyncTcpClient{
		Address: address,
		Port:    port,
		conn:    nil,
	}
}

func (client *AsyncTcpClient) Connect() error {
	if net.ParseIP(client.Address) == nil {
		ips, err := net.LookupHost(client.Address)
		if err != nil || len(ips) == 0 {
			return fmt.Errorf("DNS resolution error: %w", err)
		}
		client.Address = ips[0]
	}

	address := fmt.Sprintf("%s:%d", client.Address, client.Port)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return fmt.Errorf("error connecting to server: %w", err)
	}
	client.conn = conn
	fmt.Printf("Connected to server %s\n", address)
	return nil
}

func (client *AsyncTcpClient) GetConnection() net.Conn {
	return client.conn
}

func (client *AsyncTcpClient) Close() {
	if client.conn != nil {
		_ = client.conn.Close()
	}
}

func (client *AsyncTcpClient) Send(data []byte) error {
	_, err := client.conn.Write(data)
	return err
}

func (client *AsyncTcpClient) Receive() ([]byte, error) {
	data := make([]byte, 1024)
	n, err := client.conn.Read(data)
	if err != nil {
		return nil, err
	}
	return data[:n], nil
}

// Networking utilities

func GetSocketID(conn net.Conn) int {
	if conn == nil {
		return 0
	}

	localAddr := conn.LocalAddr().String()
	remoteAddr := conn.RemoteAddr().String()
	combined := localAddr + "_" + remoteAddr

	hash := 5381
	for _, char := range combined {
		hash = ((hash << 5) + hash) + int(char)
	}

	return abs(hash) % 1000000
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func SendJSON(conn net.Conn, data interface{}) error {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %w", err)
	}
	_, err = conn.Write(jsonBytes)
	return err
}
