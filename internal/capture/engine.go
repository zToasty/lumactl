package capture

import (
	"bufio"
	"context"
	"encoding/binary"
	"io"
	"log/slog"
	"os/exec"
	"strings"
	"time"
)

type Engine struct {
	sampleRate int
}

func NewEngine() *Engine {
	return &Engine{
		sampleRate: 48000,
	}
}

// getMonitorSource узнает, куда сейчас идет системный звук, и возвращает его монитор
func (e *Engine) getMonitorSource() string {
	out, err := exec.Command("pactl", "get-default-sink").Output()
	if err != nil {
		slog.Error("failed to get default sink", "error", err)
		return ""
	}
	
	sink := strings.TrimSpace(string(out))
	if sink != "" {
		// В PipeWire/PulseAudio у каждого выхода есть "теневой" вход (.monitor)
		// Слушая его, мы слышим ровно то, что играет в колонках
		return sink + ".monitor"
	}
	return ""
}

func (e *Engine) Stream(ctx context.Context, chunkSize int) (<-chan []int16, error) {
	device := e.getMonitorSource()
	
	args := []string{
		"--record",
		"--format=s16le",
		"--rate=48000",
		"--channels=1",
		"--latency-msec=10",
		"--process-time-msec=10",
		"--stream-name=lumactl-capture",
	}

	if device != "" {
		slog.Info("Audio capture targeting", "device", device)
		args = append(args, "--device="+device)
	}

	cmd := exec.CommandContext(ctx, "parec", args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	// Читаем stderr в отдельной горутине для диагностики
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			slog.Warn("parec stderr", "msg", scanner.Text())
		}
	}()

	out := make(chan []int16)

	go func() {
		defer close(out)
		
		// Даем системе 100мс на инициализацию потока
		time.Sleep(100 * time.Millisecond)

		byteBuf := make([]byte, chunkSize*2)
		for {
			_, err := io.ReadFull(stdout, byteBuf)
			if err != nil {
				if err != io.EOF && ctx.Err() == nil {
					slog.Error("Capture read error", "error", err)
				}
				return
			}

			samples := make([]int16, chunkSize)
			for i := range chunkSize {
				samples[i] = int16(binary.LittleEndian.Uint16(byteBuf[i*2 : i*2+2]))
			}

			select {
			case <-ctx.Done():
				return
			case out <- samples:
			default:
			}
		}
	}()

	return out, nil
}
