package solarmon

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Register struct {
	id               uint64
	gain             int64
	name             string
	bytesCnt         uint64
	unit             string
	raw              []byte
	value            string
	vtype            string
	lastReadDuration time.Duration
	lastRead         time.Time
	lastErr          error
	TsType           string
	MName            string
	Mutex            *sync.Mutex
}

func NewRegister() *Register {
	return &Register{Mutex: &sync.Mutex{}}
}

func (r *Register) ReadHR(mbus *ModbusRTU) error {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	r.lastErr = nil
	t0 := time.Now()
	r.raw, r.lastErr = mbus.Read(uint16(r.id), uint16(r.bytesCnt))
	t1 := time.Now()

	r.lastReadDuration = t1.Sub(t0)
	r.lastRead = t1

	if r.lastErr != nil {
		return r.lastErr
	}

	r.lastErr = r.ParseResult()
	return r.lastErr
}

func (r *Register) ParseResult() error {
	var v string
	fgain := float64(r.gain)
	switch r.vtype {
	case "I32":
		tmp := int32(binary.BigEndian.Uint32(r.raw))
		v = fmt.Sprint(float64(tmp) / fgain)
	case "U32":
		tmp := binary.BigEndian.Uint32(r.raw)
		if tmp == (^uint32(0)) {
			tmp = 0
		}
		v = fmt.Sprint(float64(tmp) / fgain)
	case "I16":
		tmp := int16(binary.BigEndian.Uint16(r.raw))
		v = fmt.Sprint(float64(tmp) / fgain)
	case "U16":
		tmp := binary.BigEndian.Uint16(r.raw)
		v = fmt.Sprint(float64(tmp) / fgain)
	default:
		v = fmt.Sprint(r.raw)
	}
	r.value = v
	return nil
}

func (r *Register) String() string {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	u := r.unit
	if r.unit == "_" {
		u = ""
	}

	if r.lastErr != nil {
		return fmt.Sprintf("%s ## %s", r.lastRead.Format("2006-01-02 15:04:05"), r.lastErr)
	}

	return fmt.Sprintf(
		"%15s| %10s |%v| %v", r.name, r.value+u, r.id, r.lastReadDuration.Round(100*time.Microsecond),
	)
}

func GetRegistersToRead(cfg *Config) ([]*Register, error) {
	var registersDesc []string
	registersDesc = append(registersDesc, cfg.ReadRegistersFromCli...)

	if cfg.ReadRegistersFromFile != "" {
		content, err := ioutil.ReadFile(cfg.ReadRegistersFromFile)
		if err != nil {
			return nil, err
		}

		for _, line := range strings.Split(string(content), "\n") {
			if len(line) < 1 || line[:1] == "#" {
				continue
			}

			registersDesc = append(registersDesc, line)
		}
	}

	if len(registersDesc) == 0 {
		return nil, errors.New("please specify some registers")
	}

	var registers []*Register
	var err error

	for _, r := range registersDesc {
		//register_id:total_bytes_toread:short_name:gain:register_type:unit
		rinfo := strings.Split(r, ":")

		if len(rinfo) < 6 {
			continue
		}

		r := NewRegister()

		r.TsType = cfg.DefaultTsType
		r.MName = cfg.DefaultMName
		r.id, err = strconv.ParseUint(rinfo[0], 10, 16)

		if err != nil {
			return nil, err
		}

		r.bytesCnt, err = strconv.ParseUint(rinfo[1], 10, 16)

		if err != nil {
			return nil, err
		}

		r.name = rinfo[2]
		r.gain, err = strconv.ParseInt(rinfo[3], 10, 32)

		if err != nil {
			return nil, err
		}

		r.vtype = rinfo[4]
		r.unit = rinfo[5]

		if len(rinfo) >= 7 {
			r.TsType = rinfo[6]
		}

		if len(rinfo) >= 8 {
			r.MName = rinfo[7]
		}

		registers = append(registers, r)
	}

	return registers, nil
}
