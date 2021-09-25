package oneid

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	defaultInt64ProcessBits  int64 = 5
	defaultInt64ServerBits   int64 = 10
	defaultInt64SequenceBits int64 = 12
	// minInt64 values
	minInt64ProcessBits  = 1
	minInt64ServerBits   = 1
	minInt64SequenceBits = 12

	totalInt64Bits = defaultInt64ProcessBits + defaultInt64ServerBits + defaultInt64SequenceBits

	// enviroment variables
	serverIDKey  = "SERVER_ID"
	processIDKey = "PROCESS_ID"
)

// Int64Config is cocurrently-safe stateful configuration
// for raceless int64 id generatation
type Int64Config struct {
	Epoch,
	CustomEpoch,
	LastTime,
	Sequence,
	ProcessBits,
	ServerBits,
	SequenceBits int64
	*sync.Mutex
}

// NewInt64Config makes reasonable Int64Config from the arguments provided,
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
func NewInt64Config(serverBits, processBits, sequenceBits int64) Int64Config {

	const ()

	if processBits < minInt64ProcessBits {
		processBits = minInt64ProcessBits
	}

	if serverBits < minInt64ServerBits {
		serverBits = minInt64ServerBits
	}

	if sequenceBits < minInt64SequenceBits {
		sequenceBits = minInt64ServerBits
	}

	if processBits+serverBits+sequenceBits > totalInt64Bits {
		processBits = defaultInt64ProcessBits
		serverBits = defaultInt64ServerBits
		sequenceBits = defaultInt64SequenceBits
	} else if processBits+serverBits+sequenceBits < totalInt64Bits {
		// max out bits for sequenceBits
		sequenceBits = totalInt64Bits - processBits - serverBits
	}

	var (
		now         = time.Now()
		epoch       = now.Unix()
		customEpoch = now.Unix() - epoch
	)

	return Int64Config{
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

var (
	// DefaultInt64Config sets:
	// processBits to 5, which supports upto 32 processes per server
	// serverBits: 10,  which supports upto 1024 servers
	// sequenceBits: 12, which supports upto 4096 ids per time instance
	DefaultInt64Config = NewInt64Config(defaultInt64ProcessBits, defaultInt64ServerBits, defaultInt64SequenceBits)
)

// Int64 generates an int64 id using serverID, processID and config
// if processID is zero, then the system pid will be used
func Int64(serverID, processID int64, c *Int64Config) int64 {

	if serverID <= 0 {
		serverID = 1
	}

	if processID <= 0 {
		processID = int64(os.Getpid())
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

// EnvInt64 generates an int64 id from envirment variables
// SERVER_ID: unique numeric value represents this server
// PROCESS_ID: unique numeric value represents this process
func EnvInt64(c *Int64Config) (int64, error) {

	// server ID
	serverIDText := os.Getenv(serverIDKey)

	serverID, err := strconv.ParseInt(serverIDText, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parsing serverID from env("+serverIDKey+") -> %v", err)
	}

	if serverID <= 0 {
		return 0, fmt.Errorf("serverID cannot be less than one, current value: %d", serverID)
	}

	// process ID
	processIDText := os.Getenv(processIDKey)

	processID, err := strconv.ParseInt(processIDText, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parsing processID from env("+processIDKey+") -> %v", err)
	}

	if processID < 0 {
		return 0, fmt.Errorf("processID cannot be a negative, curretn value: %d", processID)
	}

	return Int64(serverID, processID, c), nil
}

