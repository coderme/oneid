package oneid

import (
	"log"
	"os"
	"sync"
	"testing"
)

// TestNewUint64ConfigMinValues tests NewUint64Config for using minimal values configuration.
func TestNewUint64ConfigMinValues(t *testing.T) {
	t.Parallel()

	c := NewUint64Config(0, 0, 0)

	if c.ProcessBits != minUint64ProcessBits {
		t.Error("ProcessBits is not set to minimum value")
	}

	if c.ServerBits != minUint64ServerBits {
		t.Error("ServerBits is not set to minimum value", c.ServerBits)
	}

	if c.SequenceBits < minUint64SequenceBits {
		t.Error("SequenceBits is not set to minimum value")
	}
}

// TestNewUint64ConfigSequanceBitsIsMax tests NewUint64Config SequenceBits is set to the maximum value.
func TestNewUint64ConfigSequanceBitsIsMax(t *testing.T) {
	t.Parallel()

	values := [3][4]uint64{
		{0, 0, 0, totalUint64Bits - 2},
		{
			defaultUint64ProcessBits, defaultUint64ServerBits,
			defaultUint64SequenceBits, totalUint64Bits - defaultUint64ProcessBits - defaultUint64ServerBits,
		},
		{
			minUint64ProcessBits, minUint64ServerBits,
			minUint64SequenceBits, totalUint64Bits - minUint64ProcessBits - minUint64ServerBits,
		},
	}

	for _, v := range values {
		c := NewUint64Config(v[0], v[1], v[2])

		if c.SequenceBits != v[3] {
			t.Error("SequenceBits is not maxed out for values", v[0], v[1], v[2], "Expected:", v[3])
		}
	}
}

// TestNewUint64ConfigTotalBitsLength tests NewUint64Config total bits length.
func estNewUint64ConfigTotalBitsLength(t *testing.T) {
	t.Parallel()

	values := [5][3]uint64{
		{0, 0, 0},
		{5, 10, 20},
		{10, 20, 30},
		{30, 50, 60},
		{108, 200, 300},
	}

	for _, v := range values {
		c := NewUint64Config(v[0], v[1], v[2])

		if c.ProcessBits+c.ServerBits+c.SequenceBits != totalUint64Bits {
			t.Error("Total bits is not equal to", totalUint64Bits)
		}
	}
}

// TestNewCustomUint6ZeroId tests CustomUint64 for any zero id.
func TestNewCustomUint64ZeroId(t *testing.T) {
	t.Parallel()

	// create config with default values
	c := NewUint64Config(defaultUint64ProcessBits, defaultUint64ServerBits, defaultUint64SequenceBits)

	for i := uint64(0); i < 10_000; i++ {
		if Uint64(i, 0, &c) == 0 {
			t.Error("Zero Id found with serverID:", i)
		}
	}
}

// TestNewCustomUint64NonDuplicateId tests CustomUint64 for any duplicate id.
func TestNewCustomUint64DuplicateId(t *testing.T) {
	t.Parallel()

	// create config with default values
	c := NewUint64Config(defaultUint64ProcessBits, defaultUint64ServerBits, defaultUint64SequenceBits)

	var ids []uint64

	for i := uint64(0); i < 10_000; i++ {
		id := Uint64(i, 0, &c)

		for _, v := range ids {
			if id == v {
				t.Error("Duplicate Id found: ", id)
			}
		}

		ids = append(ids, id)
	}
}

// TestNewCustomUint64ForDuplicateIdMultipleThreads tests CustomUint64 for any duplicate id.
func TestNewCustomUint64NonDuplicateIdMultipleThreads(t *testing.T) {
	t.Parallel()

	ids := make(chan uint64, 100_000_000)
	wg := &sync.WaitGroup{}
	wg.Add(10_000)

	for t := 0; t < 10_000; t++ {
		go func() {
			defer wg.Done()

			for i := uint64(0); i < 10_000; i++ {
				id := Uint64(i, 0, &DefaultUint64Config)

				ids <- id
			}
		}()
	}

	wg.Wait()
	close(ids)

	seen := make(map[uint64]struct{})

	for i := range ids {
		if _, ok := seen[i]; !ok {
			seen[i] = struct{}{}
		} else {
			t.Error("Duplicate Id found: ", i)
		}
	}
}

// TestUint64ZeroServerID calls Uint64() with server id = 0.
// checks for returning a non zero id
func TestUint64ZeroServerID(t *testing.T) {
	t.Parallel()

	if Uint64(1, 0, &DefaultUint64Config) == 0 {
		t.Error("ID equals zero")
	}
}

// TestUint64ForNonUniqueIdOnSameProcessAndServer tests Uint64() serially for any duplicate ids generated
// using same serverID.
func TestUint64ForNonUniqueIdsOnSameProcessAndServer(t *testing.T) {
	t.Parallel()

	var ids []uint64

	for c := 0; c < 100_000; c++ {
		id := Uint64(1, 0, &DefaultUint64Config)

		for _, v := range ids {
			if v == id {
				t.Error("Duplicate Id found id:", id)
			}
		}
		ids = append(ids, id)
	}
}

// TestUint64ForNonUniqueIdOnSameProcessAndServerAcrossMultipleThreads tests Uint64() concurrently for
// any duplicate id generated  using same serverID.
func TestUint64ForDuplicateIdOnSameProcessAndServerAcrossMultipleThreads(t *testing.T) {
	t.Parallel()

	ids := make(chan uint64, 100_000_000)
	wg := &sync.WaitGroup{}

	wg.Add(10_000)
	for p := 0; p < 10_000; p++ {
		go func() {
			defer wg.Done()

			for c := 0; c < 10_000; c++ {
				ids <- Uint64(1, 0, &DefaultUint64Config)
			}
		}()
	}

	wg.Wait()

	close(ids)

	seen := make(map[uint64]struct{})

	for i := range ids {
		if _, ok := seen[i]; !ok {
			seen[i] = struct{}{}
		} else {
			t.Error("Duplicate Id found:", i)
		}
	}
}

// TestUint64ForNonUniqueIdOnDifferentServerIDs tests Uint64() serially for any duplicate ids generated
// using different serverIDs upto the maximum 1024.
func TestUint64ForNonUniqueIdOnDifferentServerIDs(t *testing.T) {
	t.Parallel()

	var ids []uint64

	for c := uint64(1); c < 1025; c++ {
		id := Uint64(1, 0, &DefaultUint64Config)

		for _, v := range ids {
			if v == id {
				t.Error("Duplicate Id found with serverID:", c, "id:", id)
			}
		}

		ids = append(ids, id)
	}
}

// TestUint64ForNonUniqueIdOnDifferentServerIDsAcrossMultipleThreads tests Uint64()
// concurrently for any duplicate ids generated
// using different serverIDs upto the maximum 1024.
func TestUint64ForNonUniqueIdOnDifferentServerIDsAcrossMultipleThreads(t *testing.T) {
	t.Parallel()

	ids := make(chan uint64, 10_240_000)
	wg := &sync.WaitGroup{}

	wg.Add(10_000)

	for p := 0; p < 10_000; p++ {
		go func() {
			defer wg.Done()

			for c := uint64(0); c < 1024; c++ {
				ids <- Uint64(1, 0, &DefaultUint64Config)
			}
		}()
	}

	wg.Wait()
	close(ids)

	seen := make(map[uint64]struct{})

	for i := range ids {
		if _, ok := seen[i]; !ok {
			seen[i] = struct{}{}
		} else {
			t.Error("Duplicate Id found:", i)
		}
	}
}

// TestEnvUint64 calls EnvUint64 with custom env variables.
func TestEnvUint64(t *testing.T) {
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

		id, err := EnvUint64(&DefaultUint64Config)
		if err == nil && v.IsError {
			t.Error("expected error found none.",
				"ServerID:", v.ServerID,
				"ProcessID:", v.ProcessID,
				"generated ID:", id,
			)
		}

		if err != nil && !v.IsError {
			t.Error("expected no error, found one, error:", err)
		}
	}
}

// BenchmarkUint64 benchmarks a Uint64(1).
func BenchmarkUint64(b *testing.B) {
	for c := 0; c < b.N; c++ {
		_ = Uint64(1, 0, &DefaultUint64Config)
	}
}
