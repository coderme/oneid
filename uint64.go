package oneid

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	defaultUint64ProcessBits  uint64 = 5
	defaultUint64ServerBits   uint64 = 10
	defaultUint64SequenceBits uint64 = 24

	// min values.
	minUint64ProcessBits  uint64 = 1
	minUint64ServerBits   uint64 = 1
	minUint64SequenceBits uint64 = 24

	totalUint64Bits uint64 = defaultUint64ProcessBits + defaultUint64ServerBits + defaultUint64SequenceBits
)

// Uint64Config is cocurrently-safe stateful configuration
// for raceless uint64 id generatation.
type Uint64Config struct {
	Epoch,
	CustomEpoch,
	LastTime,
	Sequence,
	ProcessBits,
	ServerBits,
	SequenceBits uint64
	*sync.Mutex
}

// NewUint64Config makes reasonable Unt64Config from the arguments passed,
//
// sequenceBits is greedy parameter and will be set to the maximum value when possible.
//
// Examples:
//
// * Horizontal scaling:
//   processBits: 5, serverBits: 10, sequenceBits 24
//   This will support upto 32 processes and 1024 servers.
//
// * Vertical scaling:
//   processBits: 6, serverBits: 1, sequenceBits: 20
//   This will support upto 64 processes.
func NewUint64Config(serverBits, processBits, sequenceBits uint64) Uint64Config {
	if processBits < minUint64ProcessBits {
		processBits = minUint64ProcessBits
	}

	if serverBits < minUint64ServerBits {
		serverBits = minUint64ServerBits
	}

	if sequenceBits < minUint64SequenceBits {
		sequenceBits = minUint64SequenceBits
	}

	if processBits+serverBits+sequenceBits > totalUint64Bits {
		processBits = defaultUint64ProcessBits
		serverBits = defaultUint64ServerBits
		sequenceBits = defaultUint64SequenceBits
	} else if processBits+serverBits+sequenceBits < totalUint64Bits {
		// max out bits for sequenceBits
		sequenceBits = totalUint64Bits - processBits - serverBits
	}

	var (
		now         = time.Now()
		epoch       = uint64(now.Unix())
		customEpoch = uint64(now.Unix()) - epoch
	)

	return Uint64Config{
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

// DefaultUint64Config sets:
// processBits to 5, which supports upto 32 processes per server
// serverBits: 10,  which supports upto 1024 servers
// sequenceBits: 24, which supports upto 16,777,216 ids per time instance.
var DefaultUint64Config = NewUint64Config(defaultUint64ProcessBits, defaultUint64ServerBits, defaultUint64SequenceBits)

// Uint64 generates uint64 id using  using serverID, processID and config
// if processID is zero, then the system pid will be used.
func Uint64(serverID, processID uint64, c *Uint64Config) uint64 {
	if processID == 0 {
		processID = uint64(os.Getpid())
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

// EnvUnt64 generates an uint64 id from envirment variables
// SERVER_ID: unique numeric value represents this server
// PROCESS_ID: unique numeric value represents this process.
func EnvUint64(c *Uint64Config) (uint64, error) {
	// server ID`
	serverIDText := os.Getenv(serverIDKey)

	serverID, err := strconv.ParseUint(serverIDText, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parsing serverID from env("+serverIDKey+") -> %v", err)
	}

	// process ID
	processIDText := os.Getenv(processIDKey)

	processID, err := strconv.ParseUint(processIDText, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parsing processID from env("+processIDKey+") -> %v", err)
	}

	return Uint64(serverID, processID, c), nil
}
