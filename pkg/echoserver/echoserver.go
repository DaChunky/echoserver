package echoserver

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/dachunky/echoserver/pkg/logging"
)

// EchoServer provides a TCP connection listener on a specific port
type EchoServer interface {
	// StartListening on the specified port
	StartListening() error
	// IsHealthy returns nil, if the server is running
	IsHealthy() error
	// Stop the server
	Stop()
}

type echoServer struct {
	port     int
	listener net.Listener
	// A wait group is needed to shut down the server correctly without any open connections
	wg sync.WaitGroup
	// Provide a possibility to stop the server
	quit chan bool
	// healthy status of the server
	isRunning bool
}

// NewEchoServer creates a new echo server instance
func NewEchoServer(port int) EchoServer {
	ret := new(echoServer)
	ret.port = port
	ret.quit = make(chan bool, 1)
	ret.isRunning = false
	return ret
}

func (es *echoServer) StartListening() error {
	// define the connection string. an empty ip address results in listening to all incoming requests on the port
	connStr := fmt.Sprintf(":%d", es.port)
	// listen to all incoming connections on the specified port
	l, err := net.Listen("tcp", connStr)
	if err != nil {
		logging.LogFmt(logging.LOG_FATAL, "[START] failed to start listening on [%s]: %x", connStr, err)
		return err
	}
	logging.LogFmt(logging.LOG_MAIN, "[START] start listening on: '%s'", connStr)
	es.listener = l
	// increase wait group count to 2.
	//   - for the listener to shut down
	//   - for the connection handler to escape
	es.wg.Add(1)
	es.isRunning = true
	logging.Log(logging.LOG_INFO, "[START] GO handling incoming requests")
	go es.handleIncomingConnections()
	return nil
}

func (es *echoServer) IsHealthy() error {
	if es.listener == nil {
		return fmt.Errorf("listener is not active")
	}
	if !es.isRunning {
		return fmt.Errorf("server is not running anymore")
	}
	return nil
}

func (es *echoServer) Stop() {
	if es.listener == nil {
		return
	}
	// send quit command to the quit channel
	logging.Log(logging.LOG_INFO, "[STOP] send quit command")
	es.quit <- true
	// stop listening
	logging.Log(logging.LOG_INFO, "[STOP] stop listening")
	es.listener.Close()
	// wait until all connection usage is done
	logging.Log(logging.LOG_INFO, "[STOP] wait for connection closing")
	es.wg.Wait()
	// clear the channel
	logging.Log(logging.LOG_INFO, "[STOP] clear quit channel")
	close(es.quit)
}

func readUntilEOF(reader *bufio.Reader) ([]byte, error) {
	ret := make([]byte, 0)
	block := 1024
	zeroLengthRetry := 0
	for {
		buf := make([]byte, block)
		n, err := reader.Read(buf)
		ret = append(ret, buf[:n]...)
		if err != nil {
			if err == io.EOF {
				if n < 1 {
					zeroLengthRetry = zeroLengthRetry + 1
					if zeroLengthRetry < 10 {
						time.Sleep(100 * time.Millisecond)
						continue
					}
					logging.LogFmt(logging.LOG_DEBUG, "[ReadUntilEOF] read EOF with zero content [%d] times", zeroLengthRetry)
					return nil, fmt.Errorf("got [%d] times EOF but no data", zeroLengthRetry)
				}
				logging.LogFmt(logging.LOG_DEBUG, "[ReadUntilEOF] read EOF [%d bytes]", n)
				return ret, nil
			}
			logging.Log(logging.LOG_DEBUG, "[ReadUntilEOF] got error not equal to EOF")
			return nil, err
		}
		logging.LogFmt(logging.LOG_DEBUG, "[ReadUntilEOF] read [%d] bytes from stream", n)
		if n < block {
			return ret, nil
		}
	}
}

func (es *echoServer) serveConnection(conn net.Conn) {
	defer func() {
		conn.Close()
		es.wg.Done()
	}()
	buf, err := readUntilEOF(bufio.NewReader(conn))
	if (err != nil) || (len(buf) < 1) {
		logging.LogFmt(logging.LOG_ERROR, "[SERVE] failed to read client request: %s", conn.RemoteAddr().String())
		return
	}
	response := string(buf)
	response = fmt.Sprintf("I hear you %s: %s", conn.RemoteAddr().String(), response)
	_, err = conn.Write([]byte(response))
	if err != nil {
		logging.LogFmt(logging.LOG_ERROR, "[SERVE] failed to server client: %s", conn.RemoteAddr().String())
	}
}

func (es *echoServer) handleIncomingConnections() {
	// release wait group counter, when stopping handling requests
	defer func() {
		es.isRunning = false
		es.wg.Done()
	}()
	for {
		conn, err := es.listener.Accept()
		if err != nil {
			select {
			case <-es.quit:
				logging.Log(logging.LOG_MAIN, "[HANDLE] retreive quit command")
				return
			default:
				logging.Log(logging.LOG_ERROR, "[HANDLE] error on retrieving incoming request")
			}
			continue
		}
		logging.LogFmt(logging.LOG_MAIN, "[HANDLE] start serving client: '%s'", conn.RemoteAddr().String())
		es.wg.Add(1)
		go es.serveConnection(conn)
	}
}
