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

// TestNewUint32ConfigMinValues tests NewUint32Config for using minimal values configuration.
func TestNewUint32ConfigMinValues(t *testing.T) {
	t.Parallel()

	c := NewUint32Config(0, 0, 0)
	if c.ProcessBits != minUint32ProcessBits {
		t.Error("ProcessBits is not set to minimum value")
	}

	if c.ServerBits != minUint32ServerBits {
		t.Error("ServerBits is not set to minimum value", c.ServerBits)
	}

	if c.SequenceBits < minUint32SequenceBits {
		t.Error("SequenceBits is not set to minimum value")
	}
}

// TestNewUint32ConfigSequanceBitsIsMax tests NewUint32Config SequenceBits is set to the maximum value.
func TestNewUint32ConfigSequanceBitsIsMax(t *testing.T) {
	t.Parallel()

	values := [3][4]uint32{
		{0, 0, 0, totalUint32Bits - 2},
		{
			defaultUint32ProcessBits, defaultUint32ServerBits, defaultUint32SequenceBits,
			totalUint32Bits - defaultUint32ProcessBits - defaultUint32ServerBits,
		},
		{
			minUint32ProcessBits, minUint32ServerBits, minUint32SequenceBits,
			totalUint32Bits - minUint32ProcessBits - minUint32ServerBits,
		},
	}

	for _, v := range values {
		c := NewUint32Config(v[0], v[1], v[2])

		if c.SequenceBits != v[3] {
			t.Error("SequenceBits is not maxed out for values", v[0], v[1], v[2], "Expected:", v[3])
		}
	}
}

// TestNewUint32ConfigTotalBitsLength tests NewUint32Config total bits length.
func estNewUint32ConfigTotalBitsLength(t *testing.T) {
	t.Parallel()

	values := [5][3]uint32{
		{0, 0, 0},
		{5, 10, 20},
		{10, 20, 30},
		{30, 50, 60},
		{108, 200, 300},
	}

	for _, v := range values {
		c := NewUint32Config(v[0], v[1], v[2])
		if c.ProcessBits+c.ServerBits+c.SequenceBits != totalUint32Bits {
			t.Error("Total bits is not equal to", totalUint32Bits)
		}
	}
}

// TestNewCustomInt6ZeroId tests CustomUint32 for any zero id.
func TestNewCustomUint32ZeroId(t *testing.T) {
	t.Parallel()

	// create config with default values
	c := NewUint32Config(defaultUint32ProcessBits, defaultUint32ServerBits, defaultUint32SequenceBits)

	for i := uint32(0); i < 1024; i++ {
		if Uint32(i, 0, &c) == 0 {
			t.Error("Zero Id found with serverID:", i)

			break
		}
	}
}

// TestNewCustomUint32NonDuplicateId tests CustomUint32 for any duplicate id.
func TestNewCustomUint32DuplicateId(t *testing.T) {
	t.Parallel()

	// create config with default values
	c := NewUint32Config(defaultUint32ProcessBits, defaultUint32ServerBits, defaultUint32SequenceBits)

	var ids []uint32

	for i := uint32(0); i < 1024; i++ {
		id := Uint32(i, 0, &c)

		for _, v := range ids {
			if id == v {
				t.Error("Duplicate Id found: ", id)

				return
			}
		}

		ids = append(ids, id)
	}
}

// TestNewCustomUint32ForDuplicateIdMultipleThreads tests CustomUint32 for any duplicate id.
func TestNewCustomUint32NonDuplicateIdMultipleThreads(t *testing.T) {
	t.Parallel()

	ids := make(chan uint32, 10_240)
	wg := &sync.WaitGroup{}

	wg.Add(10)

	// create config with default values
	//	c := NewUint32Config(defaultUint32ProcessBits, defaultUint32ServerBits, defaultUint32SequenceBits)

	for t := 0; t < 10; t++ {
		go func() {
			defer wg.Done()

			for i := uint32(0); i < 1024; i++ {
				id := Uint32(i, 0, &DefaultUint32Config)

				ids <- id
			}
		}()
	}

	wg.Wait()
	close(ids)

	seen := make(map[uint32]struct{})
	for i := range ids {
		if _, ok := seen[i]; !ok {
			seen[i] = struct{}{}

			continue
		}

		t.Error("Duplicate Id found: ", i)

		break
	}
}

// TestUint32ZeroServerID calls Uint32() with server id = 0
// checks for returning a non zero id.
func TestUint32ZeroServerID(t *testing.T) {
	t.Parallel()

	if Uint32(1, 0, &DefaultUint32Config) == 0 {
		t.Error("ID equals zero")
	}
}

// TestUint32ForNonUniqueIdOnSameProcessAndServer tests Uint32() serially for any duplicate ids generated
// using same serverID.
func TestUint32ForNonUniqueIdsOnSameProcessAndServer(t *testing.T) {
	t.Parallel()

	var ids []uint32

	for c := 0; c < 100_000; c++ {
		id := Uint32(1, 0, &DefaultUint32Config)

		for _, v := range ids {
			if v == id {
				t.Error("Duplicate Id found id:", id)

				return
			}
		}

		ids = append(ids, id)
	}
}

// TestUint32ForNonUniqueIdOnSameProcessAndServerAcrossMultipleThreads tests Uint32() concurrently for
// any duplicate id generated  using same serverID.
func TestUint32ForDuplicateIdOnSameProcessAndServerAcrossMultipleThreads(t *testing.T) {
	t.Parallel()

	ids := make(chan uint32, 100_000)
	wg := &sync.WaitGroup{}

	wg.Add(10)

	for p := 0; p < 10; p++ {
		go func() {
			defer wg.Done()

			for c := 0; c < 10_000; c++ {
				ids <- Uint32(1, 0, &DefaultUint32Config)
			}
		}()
	}

	wg.Wait()

	close(ids)

	seen := make(map[uint32]struct{})
	for i := range ids {
		if _, ok := seen[i]; !ok {
			seen[i] = struct{}{}

			continue
		}

		t.Error("Duplicate Id found:", i)

		break
	}
}

// TestUint32ForNonUniqueIdOnDifferentServerIDs tests Uint32() serially for any duplicate ids generated
// using different serverIDs upto the maximum 1024.
func TestUint32ForNonUniqueIdOnDifferentServerIDs(t *testing.T) {
	t.Parallel()

	var ids []uint32

	for c := uint32(1); c < 1025; c++ {
		id := Uint32(1, 0, &DefaultUint32Config)

		for _, v := range ids {
			if v == id {
				t.Error("Duplicate Id found with serverID:", c, "id:", id)

				return
			}
		}

		ids = append(ids, id)
	}
}

// TestUint32ForNonUniqueIdOnDifferentServerIDsAcrossMultipleThreads tests Uint32() concurrently
// for any duplicate ids generated using different serverIDs upto the maximum 1024.
func TestUint32ForNonUniqueIdOnDifferentServerIDsAcrossMultipleThreads(t *testing.T) {
	t.Parallel()

	ids := make(chan uint32, 10_240)
	wg := &sync.WaitGroup{}

	wg.Add(10)

	for p := 0; p < 10; p++ {
		go func() {
			defer wg.Done()

			for c := 0; c < 1024; c++ {
				ids <- Uint32(1, 0, &DefaultUint32Config)
			}
		}()
	}

	wg.Wait()

	close(ids)

	seen := make(map[uint32]struct{})
	for i := range ids {
		if _, ok := seen[i]; !ok {
			seen[i] = struct{}{}

			continue
		}

		t.Error("Duplicate Id found:", i)

		break
	}
}

// TestEnvUint32 calls EnvUuint32 with custom env variables.
func TestEnvUint32(t *testing.T) {
	t.Parallel()
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

		id, err := EnvUint32(&DefaultUint32Config)
		if err == nil && v.IsError {
			t.Error("expected error found none.",
				"ServerID:", v.ServerID,
				"ProcessID:", v.ProcessID,
				"generated ID:", id,
			)
		}
	}
}

// BenchmarkUint32 benchmarks a Uint32(1).
func BenchmarkUint32(b *testing.B) {
	for c := 0; c < b.N; c++ {
		_ = Uint32(1, 0, &DefaultUint32Config)
	}
}
