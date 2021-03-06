// Filename: computer_test.go
// Contents: the tests for the Computer struct and methods

package armsim

import (
	"os"
	"testing"
)

func TestComputer(t *testing.T) {
	// Ensure Computer type exists with ram, registers, and etc.
	c := new(Computer)
	_ = c.ram
	_ = c.registers
	_ = c.cpu
}

func TestNewComputer(t *testing.T) {
	computer := NewComputer(32*1024, os.Stderr)

	if computer == nil {
		t.Fatal("Did not initialize Computer.")
	}

	// Test RAM
	if computer.ram == nil {
		t.Fatal("Did not initialize RAM.")
	}
	_, err := computer.ram.ReadByte(32*1024 - 1)
	if err != nil {
		t.Fatal("Did not initialize RAM to correct size.")
	}

	// Test Registers
	if computer.registers == nil {
		t.Fatal("Did not initialize Registers.")
	}
	_, err = computer.registers.ReadByte(99)
	if err != nil {
		t.Fatal("Did not initialize registers to correct size (too small).")
	}
	_, err = computer.registers.ReadByte(100)
	if err == nil {
		t.Fatal("Did not initialize registers to correct size (too big).")
	}

	// Test CPU
	if computer.cpu == nil {
		t.Fatal("Did not initialize CPU.")
	}
	if computer.cpu.ram != computer.ram {
		t.Fatal("Did not initialize CPU with correct RAM.")
	}
	if computer.cpu.registers != computer.registers {
		t.Fatal("Did not initialize CPU with correct registers.")
	}

	// Test Step Counter
	if computer.step_counter != 1 {
		t.Fatal("Did not initialize step_counter with 1.")
	}
}

func TestRun(t *testing.T) {
	// Setup
	c := NewComputer(32*1024, os.Stderr)
	c.EnableTracing()

	halt := make(chan bool, 1)
	finished := make(chan bool, 1)

	// Simulate loading program and PC
	c.ram.WriteWord(0x0, 0x67)
	c.ram.WriteWord(0x4, 0x67)
	c.ram.WriteWord(0x8, 0x67)
	c.registers.WriteWord(PC, 0x0)

	go c.Run(halt, finished)
	<-finished
	pc, _ := c.registers.ReadWord(PC)

	// Should be the last position + 8
	// (a fetch for the 0x0 value and an increment)
	if pc != 0x08+0x8 {
		t.Fatal("Did not run until 0x0 was fetched.")
	}

	// Simulate loading program and PC
	c.ram.WriteWord(0x80, 0x67)
	c.ram.WriteWord(0x84, 0x67)
	c.ram.WriteWord(0x88, 0x67)
	c.registers.WriteWord(PC, 0x80)

	c.Run(halt, finished)
	pc, _ = c.registers.ReadWord(PC)

	// Should be the last position + 8
	// (a fetch for the 0x0 value and an increment)
	if pc != 0x88+0x8 {
		t.Fatal("Did not run until 0x0 was fetched.")
	}
}

func TestStep(t *testing.T) {
	// Setup
	c := NewComputer(32*1024, os.Stderr)

	// Simulate loading program and PC
	c.ram.WriteWord(0x0, 0x67)
	c.ram.WriteWord(0x4, 0x67)
	c.ram.WriteWord(0x8, 0x67)
	c.registers.WriteWord(PC, 0x0)

	c.Step()
	pc, _ := c.registers.ReadWord(PC)
	// Test PC
	if pc != 0x04 {
		t.Fatal("Did not step to 0x4.")
	}
	// Test step_counter
	if c.step_counter != 2 {
		t.Fatal("Did not increment step_counter to 1.")
	}

	c.Step()
	// Test PC
	pc, _ = c.registers.ReadWord(PC)
	if pc != 0x08 {
		t.Fatal("Did not step to 0x8.")
	}
	// Test step_counter
	if c.step_counter != 3 {
		t.Fatal("Did not increment step_counter to 2.")
	}

	c.Step()
	// Test PC
	pc, _ = c.registers.ReadWord(PC)
	if pc != 0x0C {
		t.Fatal("Did not step to 0xC.")
	}
	// Test step_counter
	if c.step_counter != 4 {
		t.Fatal("Did not increment step_counter to 3.")
	}
}

/*
func TestTrace(t *testing.T) {
	// Setup
	c := NewComputer(32*1024, os.Stderr)

	// Simulate loading program and PC
	c.ram.WriteWord(0x0, 0x67)
	c.ram.WriteWord(0x4, 0x67)
	c.ram.WriteWord(0x8, 0x67)
	c.registers.WriteWord(PC, 0x0)

	output := c.Trace() + "\n"
	c.Step()
	output += c.Trace() + "\n"
	c.Step()
	output += c.Trace()
	c.Step()
	t.Log(output)

	// Open test file
	file, err := os.Open("../../test/trace_test.log")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	// Compare output and test file
	b := make([]byte, 1)
	for i := 0; i < len(output); i++ {
		file.Read(b)
		if output[i] != b[0] {
			t.Fatalf("Trace output incorrect. @ %d, %d != %d", i, output[i], b[0])
		}
	}
}
*/

func TestLoadELF(t *testing.T) {
	var c *Computer
	var checksum int32
	var pc uint32

	// Setup Computer
	c = NewComputer(32*1024, os.Stderr)

	// Load Non-existent Test File
	err := c.LoadELF("asdfasdfaitheirhasdifhadf")
	if err == nil {
		t.Fatalf("should have failed file error")
	}

	// Load Non-ELF Test File
	err = c.LoadELF("computer.go")
	if err == nil {
		t.Fatalf("should have failed with magic number err")
	}

	// Load Test File 1
	c = NewComputer(32*1024, os.Stderr)
	t.Log("Checksum of empty RAM: ", c.ram.Checksum())
	err = c.LoadELF("../../test/test1.exe")
	// Check Checksum
	checksum = c.ram.Checksum()
	if err != nil || checksum != 536861081 {
		t.Fatalf("Checksum did not match for test1.exe. Expected 536861081. Got %d", checksum)
	}
	// Check PC (should be entry point)
	pc, err = c.registers.ReadWord(PC)
	if pc != 0x138 {
		t.Fatalf("Did not set PC to correct entry point. Expected 0x138. Got %#x", pc)
	}

	// Clear Computer
	c = NewComputer(32*1024, os.Stderr)

	// Load Test File 2
	err = c.LoadELF("../../test/test2.exe")
	checksum = c.ram.Checksum()
	if err != nil || checksum != 536864433 {
		t.Fatalf("Checksum did not match for test2.exe. Expected 536864433. Got %d", checksum)
	}
	// Check PC (should be entry point)
	pc, err = c.registers.ReadWord(PC)
	if pc != 0x1038 {
		t.Fatalf("Did not set PC to correct entry point. Expected 0x1038. Got %#x", pc)
	}

	// Clear RAM
	c = NewComputer(32*1024, os.Stderr)

	// Load Test File 3
	err = c.LoadELF("../../test/test3.exe")
	checksum = c.ram.Checksum()
	if err != nil || checksum != 536861199 {
		t.Fatalf("Checksum did not match for test3.exe. Expected 536861199. Got %d", checksum)
	}
	// Check PC (should be entry point)
	pc, err = c.registers.ReadWord(PC)
	if pc != 0x1038 {
		t.Fatalf("Did not set PC to correct entry point. Expected 0x138. Got %#x", pc)
	}

	// Clear RAM
	c = NewComputer(8, os.Stderr)

	// Load Test 3 into insuffcient memory
	err = c.LoadELF("../../test/test3.exe")
	if err == nil {
		t.Fatal("Should have failed with insuffcient memory error.")
	}
}

func TestCompChecksum(t *testing.T) {
	computer := NewComputer(32*1024, os.Stderr)

	if computer.Checksum() != computer.ram.Checksum() {
		t.Fatal("Checksums didn't match.")
	}

	computer.ram.WriteWord(56, 0xFBC)
	if computer.Checksum() != computer.ram.Checksum() {
		t.Fatal("Checksums didn't match.")
	}

	computer.ram.WriteWord(0xFE, 0xabc)
	if computer.Checksum() != computer.ram.Checksum() {
		t.Fatal("Checksums didn't match.")
	}
}

func TestEnableTracing(t *testing.T) {
	c := NewComputer(32*1024, os.Stderr)

	c.EnableTracing()
	if c.traceFile == nil {
		t.Fatal("Tracing was not enabled.")
	}
}

func TestDisableTracing(t *testing.T) {
	c := NewComputer(32*1024, os.Stderr)

	c.DisableTracing()
	if c.traceFile != nil {
		t.Fatal("Tracing was not disabled.")
	}
	if f, err := os.Open("trace.log"); f == nil || err != nil {
		t.Fatal("Trace file was not written to disk.")
	}

	c.DisableTracing()
}

func TestReset(t *testing.T) {
	c := NewComputer(32*1024, os.Stderr)
	c.LoadELF("../../test/test1.exe")
	c.Run(nil, nil)

	c.Reset()

	file, err := os.Open("trace.log")
	if err != nil {
		t.Fatal("Trace file should be created.")
	}
	b := make([]byte, 1)
	file.Read(b)
	if b[0] != 0 {
		t.Fatal("Trace file should be empty.")
	}

	if c.step_counter != 1 {
		t.Fatal("Did not reset the step counter.")
	}

	for i := 0; i < 32*1024; i++ {
		b, _ := c.ram.ReadByte(uint32(i))
		if b != 0 {
			t.Fatal("Reset did not clear all the bits in RAM.")
		}
	}
	for i := 0; i < 12<<2; i++ {
		b, _ := c.registers.ReadByte(uint32(i))
		if b != 0 {
			t.Fatal("Reset did not clear all the bits in the registers.")
		}
	}
}
