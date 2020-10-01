// bare bones client implementation
package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/teirm/go_derpy_fs/common"
)

const (
	defaultAddress string = "127.0.0.1"
	defaultPort    string = "0"
)

type ClientConfig struct {
	ip          *string
	port        *string
	account     *string
	op          *string
	file        *string
	interactive *bool
}

type ClientState struct {
	conn     net.Conn
	diskIo   chan common.ClientData
	netWrite chan common.ClientData
	netRead  chan common.ResponseData
}

// Create a header
func serializeHeader(header common.Header) string {
	sizeStr := strconv.FormatUint(header.Size, 64)
	s := []string{header.Operation, header.Account, header.FileName, sizeStr}
	return strings.Join(s, ":")
}

// create conection to server
func connect(ip string, port string) (net.Conn, error) {
	address := ip + ":" + port
	return net.Dial("tcp", address)
}

// performOperation
func performOperation(config ClientConfig, client ClientState) error {

	account := *config.account
	fileName := *config.file
	switch *config.op {
	case "CREATE":
		doCreate(account, client)
	case "READ":
		doRead(account, fileName, client)
	case "WRITE":
		doWrite(account, fileName, client)
	case "DELETE":
		doDelete(account, fileName, client)
	case "LIST":
		doList(account, client)
	}
	return nil
}

// do a create operation for a new account
func doCreate(account string, client ClientState) {
	header := common.Header{"CREATE", account, "", 0}
	client.netWrite <- common.ClientData{header, "", client.conn}
}

// do a read operation
func doRead(account string, fileName string, client ClientState) {
	header := common.Header{"READ", account, fileName, 0}
	client.netWrite <- common.ClientData{header, "", client.conn}
}

// do a write operation
func doWrite(account string, fileName string, client ClientState) {
	header := common.Header{"WRITE", account, fileName, 0}
	client.diskIo <- common.ClientData{header, "", client.conn}
}

// do a delete operation
func doDelete(account string, fileName string, client ClientState) {
	header := common.Header{"DELETE", account, fileName, 0}
	client.netWrite <- common.ClientData{header, "", client.conn}
}

// do a list operation
func doList(account string, client ClientState) {
	header := common.Header{"LIST", account, "", 0}
	client.netWrite <- common.ClientData{header, "", client.conn}
}

// Basic sanity checking on configuration
func validateConfig(config ClientConfig) error {
	if *config.account == "" {
		return fmt.Errorf("invalid account name: %s", *config.account)
	}

	if err := common.CheckOperation(*config.op); err != nil {
		return err
	}

	return nil
}

// initialize and start client
func startClient(config ClientConfig) error {
	var client ClientState
	var err error

	// TODO: connecting so early might be problematic
	// if disk is slow. Maybe connect closer to when
	// doing network IO
	client.conn, err = connect(*config.ip, *config.port)
	if err != nil {
		log.Printf("unable to connect to server: %v\n", err)
		return err
	}

	// default to non-interactive worker count
	netWorkers := 1
	diskWorkers := 1
	respWorkers := 1
	if *config.interactive == true {
		netWorkers = 3
		diskWorkers = 3
		respWorkers = 3
	}

	client.diskIo = make(chan common.ClientData)
	client.netWrite = make(chan common.ClientData)
	client.netRead = make(chan common.ResponseData)

	for i := 0; i < netWorkers; i++ {
		go func(cli ClientState) {
			for data := range cli.netWrite {
				// TODO: write client message
			}
		}(client)
	}

	for i := 0; i < diskWorkers; i++ {
		go func(cli ClientState) {
			for data := range cli.diskIo {
				// TODO: do disk IO
			}
		}(client)
	}

	for i := 0; i < respWorkers; i++ {
		go func(cli ClientState) {
			for data := range cli.netRead {
				// TODO: read server responses
			}
		}(client)
	}

	return nil
}

func main() {
	var config ClientConfig
	config.ip = flag.String("address", defaultAddress, "address to connect to")
	config.port = flag.String("port", defaultPort, "port to connect to")
	config.account = flag.String("account", "", "account to access")
	config.op = flag.String("op", "NOOP", "operation to perform")
	config.file = flag.String("file-name", "", "file to read or write into")
	config.interactive = flag.Bool("interactive", false, "start an interactice session")

	flag.Parse()

	err := startClient(config)
	if err != nil {
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}
