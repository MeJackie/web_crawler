package buffer

import (
	"fmt"
	"testing"
)

func TestNewBuffer(t *testing.T) {
	size := uint32(10)
	buf, err := NewBuffer(size)
	if err != nil {
		t.Fatalf("An error occurs when new a buffer : %s (size: %d)",
		err, size)
	}

	if buf == nil {
		t.Fatal("Couldn't create buffer!")
	}

	if size != buf.Cap() {
		t.Fatalf("Inconsistent buffer cap: expected: %d, actual: %d",
			size, buf.Cap())
	}

	buf, err = NewBuffer(0)
	if err == nil {
		t.Fatal("No error when new a buffer with zero size!")
	}
}

func TestMyBufferPut(t *testing.T) {
	size := uint32(10)
	buf, err := NewBuffer(size)
	if err != nil {
		t.Fatalf("An error occurs when new a buffer : %s (size: %d)",
			err, size)
	}

	//生成测试数据
	data := make([]uint32, size)
	for i := uint32(0); i < size; i++ {
		data[i] = i
	}

	var count uint32
	var datum uint32
	for _, datum := range data {
		ok, err := buf.Put(datum)
		if err != nil {
			t.Fatalf("An error occurs when putting a datum to the buffer: %s (datum: %d)",
				err, datum)
		}

		if !ok {
			t.Fatalf("couldn'd put a datum to the buffer! (buffer: %d)",
				datum)
		}

		count++
		if buf.Len() != count {
			t.Fatalf("Inconsistent buffer len: excepted: %d, actual: %d",
				count, buf.Len())
		}
	}

	datum = size
	ok, err := buf.Put(datum)
	if err != nil {
		t.Fatalf("An error occurs when putting a datum to the buffer: %s (datum: %d)",
			err, datum)
	}

	if ok {
		t.Fatalf("It still can put datum to the full buffer! (datum: %d)",
			datum)
	}

	buf.Close()
	_, err = buf.Put(datum)
	if err == nil {
		t.Fatalf("It still can put datum to the closed buffer! (datum: %d)",
			datum)
	}
}

func TestMyBufferPutInParallel(t *testing.T) {

	size := uint32(22)
	bufferSize := uint32(20)
	buf, err := NewBuffer(bufferSize)
	if err != nil {
		t.Fatalf("An error occurs when new a buffer : %s (size: %d)",
			err, size)
	}

	//生成测试数据
	data := make([]uint32, size)
	for i := uint32(0); i < size; i++ {
		data[i] = i
	}

	testingFunc := func(datum interface{}, t *testing.T) func(t *testing.T) {
		return func(t *testing.T) {
			t.Parallel()
			ok, err := buf.Put(datum)
			if err != nil {
				t.Fatalf("An error occurs when putting a datum to the buffer: %s (datum: %d)",
					err, datum)
			}

			if !ok && buf.Len() < buf.Cap() {
				t.Fatalf("Couldn't put datum to the buffer! (datum: %d)",
					datum)
			}
		}
	}

	t.Run("Put in parallel(1)", func(t *testing.T) {
		for _, datum := range data[:size/2]{
			t.Run(fmt.Sprintf("Datum=%d", datum),testingFunc(datum, t))
		}
	})

	t.Run("Put in parallel(2)", func(t *testing.T) {
		for _, datum := range data[size/2:] {
			t.Run(fmt.Sprintf("Datum=%d", datum), testingFunc(datum, t))
		}
	})
	if buf.Len() != buf.Cap() {
		t.Fatalf("Inconsistent buffer len: expected: %d, actual: %d",
			buf.Cap(), buf.Len())
	}
}

