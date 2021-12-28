package oneid

import (
	"log"
	"os"
	"sync"
	"testing"
)

func cleanEnvVars() {
	const msg = "failed unsetting env variable"

	if err := os.Unsetenv(serverIDKey); err != nil {
		log.Fatalln(msg, serverIDKey, "error:", err)
	}

	if err := os.Unsetenv(processIDKey); err != nil {
		log.Fatalln(msg, processIDKey, "error:", err)
	}

}

type EnvTestData struct {
	ServerID,
	ProcessID string
	IsError bool
}

// TestNewInt64ConfigMinValues tests NewInt64Config for using minimal values configuration
func TestNewInt64ConfigMinValues(t *testing.T) {

	c := NewInt64Config(0, 0, 0)

	if c.ProcessBits != minInt64ProcessBits {
		t.Error("ProcessBits is not set to minimum value")
	}

	if c.ServerBits != minInt64ServerBits {
		t.Error("ServerBits is not set to minimum value", c.ServerBits)
	}

	if c.SequenceBits < minInt64SequenceBits {
		t.Error("SequenceBits is not set to minimum value")
	}

}

// TestNewInt64ConfigSequanceBitsIsMax tests NewInt64Config SequenceBits is set to the maximum value
func TestNewInt64ConfigSequanceBitsIsMax(t *testing.T) {

	values := [3][4]int64{
		{0, 0, 0, totalInt64Bits - 2},
		{defaultInt64ProcessBits, defaultInt64ServerBits, defaultInt64SequenceBits, totalInt64Bits - defaultInt64ProcessBits - defaultInt64ServerBits},
		{minInt64ProcessBits, minInt64ServerBits, minInt64SequenceBits, totalInt64Bits - minInt64ProcessBits - minInt64ServerBits},
	}

	for _, v := range values {

		c := NewInt64Config(v[0], v[1], v[2])

		if c.SequenceBits != v[3] {
			t.Error("SequenceBits is not maxed out for values", v[0], v[1], v[2], "Expected:", v[3])

		}
	}

}

// TestNewInt64ConfigTotalBitsLength tests NewInt64Config total bits length
func estNewInt64ConfigTotalBitsLength(t *testing.T) {

	values := [5][3]int64{
		{0, 0, 0},
		{5, 10, 20},
		{10, 20, 30},
		{30, 50, 60},
		{108, 200, 300},
	}

	for _, v := range values {

		c := NewInt64Config(v[0], v[1], v[2])

		if c.ProcessBits+c.ServerBits+c.SequenceBits != totalInt64Bits {
			t.Error("Total bits is not equal to", totalInt64Bits)

		}
	}

}

// TestNewCustomInt6ZeroId tests CustomInt64 for any zero id
func TestNewCustomInt64ZeroId(t *testing.T) {

	// create config with default values
	c := NewInt64Config(defaultInt64ProcessBits, defaultInt64ServerBits, defaultInt64SequenceBits)

	for i := int64(0); i < 1024; i++ {
		if Int64(i, 0, &c) == 0 {
			t.Error("Zero Id found with serverID:", i)
		}
	}

}

// TestNewCustomInt64NonDuplicateId tests CustomInt64 for any duplicate id
func TestNewCustomInt64DuplicateId(t *testing.T) {

	// create config with default values
	c := NewInt64Config(defaultInt64ProcessBits, defaultInt64ServerBits, defaultInt64SequenceBits)

	var ids []int64

	for i := int64(0); i < 1024; i++ {
		id := Int64(i, 0, &c)

		for _, v := range ids {
			if id == v {
				t.Error("Duplicate Id found: ", id)
			}
		}

		ids = append(ids, id)
	}
}

// TestNewCustomInt64ForDuplicateIdMultipleThreads tests CustomInt64 for any duplicate id
func TestNewCustomInt64NonDuplicateIdMultipleThreads(t *testing.T) {

	ids := make(chan int64, 10_240)
	wg := &sync.WaitGroup{}
	wg.Add(10)

	// create config with default values
	c := NewInt64Config(defaultInt64ProcessBits, defaultInt64ServerBits, defaultInt64SequenceBits)

	for t := 0; t < 10; t++ {
		go func() {
			defer wg.Done()

			for i := int64(0); i < 1024; i++ {

				id := Int64(i, 0, &c)

				ids <- id
			}
		}()
	}

	wg.Wait()
	close(ids)

	seen := make(map[int64]struct{})

	for i := range ids {
		if _, ok := seen[i]; !ok {
			seen[i] = struct{}{}
		} else {
			t.Error("Duplicate Id found: ", i)
		}
	}

}

// TestInt64ZeroServerID calls Int64() with server id = 0
// checks for returning a non zero id
func TestInt64ZeroServerID(t *testing.T) {

	if Int64(1, 0, &DefaultInt64Config) == 0 {
		t.Error("ID equals zero")
	}

}

// TestInt64ForNonUniqueIdOnSameProcessAndServer tests Int64() serially for any duplicate ids generated
// using same serverID
func TestInt64ForNonUniqueIdsOnSameProcessAndServer(t *testing.T) {
	var ids []int64

	for c := 0; c < 100_000; c++ {
		id := Int64(1, 0, &DefaultInt64Config)

		for _, v := range ids {
			if v == id {
				t.Error("Duplicate Id found id:", id)
			}
		}
		ids = append(ids, id)
	}
}

// TestInt64ForNonUniqueIdOnSameProcessAndServerAcrossMultipleThreads tests Int64() concurrently for
// any duplicate id generated  using same serverID
func TestInt64ForDuplicateIdOnSameProcessAndServerAcrossMultipleThreads(t *testing.T) {
	ids := make(chan int64, 100_000)
	wg := &sync.WaitGroup{}

	wg.Add(10)
	for p := 0; p < 10; p++ {
		go func() {
			defer wg.Done()

			for c := 0; c < 10_000; c++ {
				ids <- Int64(1, 0, &DefaultInt64Config)
			}
		}()
	}

	wg.Wait()

	close(ids)
	seen := make(map[int64]struct{})

	for i := range ids {
		if _, ok := seen[i]; !ok {
			seen[i] = struct{}{}
		} else {
			t.Error("Duplicate Id found:", i)
		}
	}
}

// TestInt64ForNonUniqueIdOnDifferentServerIDs tests Int64() serially for any duplicate ids generated
// using different serverIDs upto the maximum 1024
func TestInt64ForNonUniqueIdOnDifferentServerIDs(t *testing.T) {
	var ids []int64

	for c := int64(1); c < 1025; c++ {
		id := Int64(1, 0, &DefaultInt64Config)

		for _, v := range ids {
			if v == id {
				t.Error("Duplicate Id found with serverID:", c, "id:", id)
			}
		}
		ids = append(ids, id)
	}
}

// TestInt64ForNonUniqueIdOnDifferentServerIDsAcrossMultipleThreads tests Int64() concurrently for any duplicate ids generated
// using different serverIDs upto the maximum 1024
func TestInt64ForNonUniqueIdOnDifferentServerIDsAcrossMultipleThreads(t *testing.T) {

	ids := make(chan int64, 10_240)
	wg := &sync.WaitGroup{}

	wg.Add(10)
	for p := 0; p < 10; p++ {
		go func() {
			defer wg.Done()

			for c := 0; c < 1024; c++ {
				ids <- Int64(1, 0, &DefaultInt64Config)
			}
		}()
	}

	wg.Wait()

	close(ids)
	seen := make(map[int64]struct{})

	for i := range ids {
		if _, ok := seen[i]; !ok {
			seen[i] = struct{}{}
		} else {
			t.Error("Duplicate Id found:", i)
		}
	}
}

// TestEnvInt64 calls EnvUint64 with custom env variables
func TestEnvInt64(t *testing.T) {
	cleanEnvVars()

	data := []EnvTestData{
		{
			ServerID:  "",
			ProcessID: "",
			IsError:   true,
		},
		{
			ServerID:  "0",
			ProcessID: "0",
			IsError:   true,
		},
		{
			ServerID:  "1",
			ProcessID: "0",
			IsError:   false,
		},
		{
			ServerID:  "-1",
			ProcessID: "-1",
			IsError:   true,
		},
		{
			ServerID:  " ",
			ProcessID: " ",
			IsError:   true,
		},
		{
			ServerID:  "100_000",
			ProcessID: "100_000",
			IsError:   true,
		},
		{
			ServerID:  "1",
			ProcessID: "1",
			IsError:   false,
		},
		{
			ServerID:  "100",
			ProcessID: "100",
			IsError:   false,
		},
	}

	for _, v := range data {
		err := os.Setenv(serverIDKey, v.ServerID)
		if err != nil {
			log.Fatalln("failed to set env", serverIDKey, "to", v.ServerID)
		}

		err = os.Setenv(processIDKey, v.ProcessID)
		if err != nil {
			log.Fatalln("failed to set env", processIDKey, "to", v.ProcessID)
		}
