package daemon

import (
	"encoding/json"
	"io"
	"log/slog"
	"net"
	"os"

	"github.com/zToasty/lumactl/internal/effects"
)

// SocketPath — путь к файлу сокета.
const SocketPath = "/tmp/lumactl.sock"

// Command представляет структуру JSON-запроса от клиента
type Command struct {
	Action string          `json:"action"` // например, "switch" или "stop"
	Effect string          `json:"effect"` // имя эффекта из реестра
	Params json.RawMessage `json:"params"` // дополнительные аргументы в будущем
}

type Server struct {
	manager *effects.Manager
}

func NewServer(m *effects.Manager) *Server {
	return &Server{manager: m}
}

// Start запускает прослушивание сокета
func (s *Server) Start() error {
	// Если остался файл от прошлого сокета - удаляем его
	if _, err := os.Stat(SocketPath); err == nil {
		os.Remove(SocketPath)
	}

	listener, err := net.Listen("unix", SocketPath)
	if err != nil {
		return err
	}
	defer listener.Close()

	os.Chmod(SocketPath, 0666)
	slog.Info("Daemon server started", "socket", SocketPath)

	for {
		conn, err := listener.Accept()
		if err != nil {
			slog.Error("Failed to accept connection", "error", err)
			continue
		}

		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	decoder := json.NewDecoder(conn)
	for {
		var cmd Command
		if err := decoder.Decode(&cmd); err != nil {
			if err != io.EOF {
				slog.Debug("Failed to decode command", "error", err)
			}
			break
		}

		s.processCommand(cmd)
	}
}

func (s *Server) processCommand(cmd Command) {
	switch cmd.Action {
	case "switch":
		slog.Info("Received switch command", "effect", cmd.Effect)
		if err := s.manager.SwitchEffect(cmd.Effect, cmd.Params); err != nil {
			slog.Error("Failed to switch effect", "effect", cmd.Effect, "error", err)
		}
	case "stop":
		slog.Info("Received stop command")
		s.manager.Stop()
	default:
		slog.Warn("Received unknown action", "action", cmd.Action)
	}
}
