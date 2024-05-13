package code

import "testing"

func TestMake(t *testing.T) {
	tests := []struct {
		op       Opcode
		operands []int
		expected []byte
	}{
		{OpConstant, []int{65534}, []byte{byte(OpConstant), 255, 254}},
		{OpAdd, []int{}, []byte{byte(OpAdd)}},
		{OpSetLocal, []int{255}, []byte{byte(OpSetLocal), 255}},
	}

	for _, ts := range tests {
		instruction := Make(ts.op, ts.operands...)

		if len(instruction) != len(ts.expected) {
			t.Errorf("instruction has wrong length. want=%d,got=%d", len(ts.expected), len(instruction))
		}

		for i, bt := range ts.expected {
			if instruction[i] != bt {
				t.Errorf("wrong byte at pos %d. want=%d, got=%d", i, bt, instruction[i])
			}
		}
	}
}

func TestInstructionsString(t *testing.T) {
	instructions := []Instructions{
		Make(OpAdd),
		Make(OpConstant, 2),
		Make(OpGetLocal, 1),
		Make(OpConstant, 65535),
	}
	expected := `0000 OpAdd
0001 OpConstant 2
0004 OpGetLocal 1
0006 OpConstant 65535
`

	concatted := Instructions{}
	for _, ins := range instructions {
		concatted = append(concatted, ins...)
	}

	if concatted.String() != expected {
		t.Errorf("instructions wrongly formatted.\nwant=%q\ngot=%q",
			expected, concatted.String())
	}
}

func TestReadOperands(t *testing.T) {
	tests := []struct {
		op        Opcode
		operands  []int
		bytesRead int
	}{
		{OpConstant, []int{65535}, 2},
		{OpSetLocal, []int{255}, 1},
	}

	for _, ts := range tests {
		instructions := Make(ts.op, ts.operands...)
		def, err := Lookup(byte(ts.op))
		if err != nil {
			t.Fatalf("definition not find: %q\n", err)
		}

		operandsRead, n := ReadOperands(def, instructions[1:])
		if n != ts.bytesRead {
			t.Fatalf("n wrong. want=%d, got=%d", ts.bytesRead, n)
		}
		for i, want := range ts.operands {
			if operandsRead[i] != want {
				t.Errorf("operand wrong. want=%d, got=%d", want, operandsRead[i])
			}
		}
	}
}
