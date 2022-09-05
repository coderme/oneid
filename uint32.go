package oneid

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	defaultUint32ProcessBits  uint32 = 5
	defaultUint32ServerBits   uint32 = 10
	defaultUint32SequenceBits uint32 = 20
	// minUint32 values.
	minUint32ProcessBits  = 1
	minUint32ServerBits   = 1
	minUint32SequenceBits = 15

	totalUint32Bits = defaultUint32ProcessBits + defaultUint32ServerBits + defaultUint32SequenceBits

	// environment variables.
	serverIDKey  = "SERVER_ID"
	processIDKey = "PROCESS_ID"
)

// Uint32Config is cocurrently-safe stateful configuration
// for raceless uint32 id generatation.
type Uint32Config struct {
	Epoch,
	CustomEpoch,
	LastTime,
	Sequence,
	ProcessBits,
	ServerBits,
	SequenceBits uint32
	*sync.Mutex
}

// NewUint32Config makes reasonable Uint32Config from the arguments provided,
//
// sequenceBits is greedy parameter and will be set to the maximum value when possible.
//
// Examples:
//
// * Horizontal scaling:
//   processBits: 5, serverBits: 10, sequenceBits 12
//   This will support upto 32 processes and 1024 servers.
//
// * Vertical scaling:
//   processBits: 6, serverBits: 1, sequenceBits: 20
//   This will support upto 64 processes.
func NewUint32Config(serverBits, processBits, sequenceBits uint32) Uint32Config {
	if processBits < minUint32ProcessBits {
		processBits = minUint32ProcessBits
	}

	if serverBits < minUint32ServerBits {
		serverBits = minUint32ServerBits
	}

	if sequenceBits < minUint32SequenceBits {
		sequenceBits = minUint32ServerBits
	}

	if processBits+serverBits+sequenceBits > totalUint32Bits {
		processBits = defaultUint32ProcessBits
		serverBits = defaultUint32ServerBits
		sequenceBits = defaultUint32SequenceBits
	} else if processBits+serverBits+sequenceBits < totalUint32Bits {
		// max out bits for sequenceBits
		sequenceBits = totalUint32Bits - processBits - serverBits
	}

	var (
		now         = time.Now()
		epoch       = uint32(now.Unix())
		customEpoch = uint32(now.Unix()) - epoch
	)

	return Uint32Config{
		Epoch:        epoch,
		CustomEpoch:  customEpoch,
		LastTime:     0,
		Sequence:     0,
		ProcessBits:  processBits,
		ServerBits:   serverBits,
		SequenceBits: sequenceBits,
		Mutex:        &sync.Mutex{},
	}
}

// DefaultUint32Config sets:
// processBits to 5, which supports upto 32 processes per server
// serverBits: 10,  which supports upto 1024 servers
// sequenceBits: 12, which supports upto 4096 ids per time instance.
var DefaultUint32Config = NewUint32Config(defaultUint32ProcessBits, defaultUint32ServerBits, defaultUint32SequenceBits)

// Uint32 generates an uint32 id using serverID, processID and config
// if processID is zero, then the system pid will be used.
func Uint32(serverID, processID uint32, c *Uint32Config) uint32 {
	if serverID == 0 {
		serverID = 1
	}

	if processID == 0 {
		processID = uint32(os.Getpid())
	}

	c.Lock()
	defer c.Unlock()

	if c.CustomEpoch <= c.LastTime {
		c.Sequence++
		if c.Sequence == (2 << (c.SequenceBits - 1)) {
			c.Sequence = 0
			c.LastTime++
		}
	} else {
		c.Sequence = 0
		c.LastTime = c.CustomEpoch
	}

	return c.LastTime<<(c.ServerBits+c.ProcessBits+c.SequenceBits) |
		(serverID&(2<<(c.ServerBits-1)))<<(c.ProcessBits+c.SequenceBits) |
		(processID & (2 << (c.ProcessBits - 1)) << c.SequenceBits) |
		c.Sequence
}

// EnvUint32 generates an uint32 id from envirment variables
// SERVER_ID: unique numeric value represents this server
// PROCESS_ID: unique numeric value represents this process.
func EnvUint32(c *Uint32Config) (uint32, error) {
	// server ID
	serverIDText := os.Getenv(serverIDKey)

	serverID, err := strconv.ParseUint(serverIDText, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("parsing serverID from env("+serverIDKey+") -> %v", err)
	}

	if serverID == 0 {
		return 0, fmt.Errorf("serverID cannot be less than one, current value: %d", serverID)
	}

	// process ID
	processIDText := os.Getenv(processIDKey)

	processID, err := strconv.ParseUint(processIDText, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("parsing processID from env("+processIDKey+") -> %v", err)
	}

	return Uint32(uint32(serverID), uint32(processID), c), nil
}
