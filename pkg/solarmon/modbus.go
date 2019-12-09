package solarmon

import (
	"fmt"
	"os"
	"runtime"

	"github.com/goburrow/modbus"
)

func NewModbusRTU(cfg *Config) (*ModbusRTU, error) {
	portName := cfg.TTYFile
	if runtime.GOOS == "windows" {
		portName = "\\\\.\\" + cfg.TTYFile
	}

	handler := modbus.NewRTUClientHandler(portName)
	handler.BaudRate = cfg.BaudRate
	handler.DataBits = cfg.DataBits
	handler.Parity = cfg.ModbusParity
	handler.StopBits = cfg.StopBits
	handler.SlaveId = byte(cfg.SlaveID)
	handler.Timeout = cfg.Timeout
	handler.IdleTimeout = cfg.Timeout * 2

	return &ModbusRTU{handler: handler, connected: false}, nil
}

type ModbusRTU struct {
	handler   *modbus.RTUClientHandler
	client    modbus.Client
	connected bool
}

func (m *ModbusRTU) Client() (modbus.Client, error) {
	var err error

	if runtime.GOOS == "linux" {
		if _, err := os.Stat(m.handler.Address); err != nil {
			m.connected = false
			err = fmt.Errorf("TTYFile %s is missing(make sure rs485 to USB adapter is connected)", m.handler.Address)
			return nil, err
		}
	}

	if m.client == nil {
		m.client = modbus.NewClient(m.handler)
	}

	if !m.connected {
		err = m.Reconnect()

		m.connected = true
		if err != nil {
			err = fmt.Errorf("error while connecting to %s: %v", m.handler.Address, err)
			m.connected = false
		}
	}

	return m.client, err
}

func (m *ModbusRTU) Read(id uint16, cnt uint16) ([]byte, error) {
	var err error
	var r []byte

	c, err := m.Client()

	if err != nil {
		return nil, fmt.Errorf("unable to read hregister %v: %v", id, err)
	}

	r, err = c.ReadHoldingRegisters(id, cnt)

	if err != nil {
		err = fmt.Errorf("unable to read hregister %v: %v", id, err)
		m.connected = false
	}

	return r, err
	//return []byte{0x01, 0x00, 0x00, 0x00}, nil
}

func (m *ModbusRTU) Reconnect() error {
	m.handler.Close()
	return m.handler.Connect()
}

func (m *ModbusRTU) Close() {
	m.handler.Close()
}
