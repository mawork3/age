package age

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"time"
	"unsafe"
)

// Struct cast
type PinkStruct struct {
	A uint8
	B int
	C int64
}

type VioletStruct struct {
	A uint8
	B int64
	C int64
}

func m1() {
	pink := PinkStruct{
		A: 1,
		B: 42,
		C: 9000,
	}

	violet := *(*VioletStruct)(unsafe.Pointer(&pink))

	fmt.Println(violet.A)
	fmt.Println(violet.B)
	fmt.Println(violet.C)
}

func m2() {
	go heapHeapHeap()

	readAndHaveFun()
}

func unsafeStringToBytes(s *string) []byte {
	sh := (*reflect.StringHeader)(unsafe.Pointer(s))
	sliceHeader := &reflect.SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}

	// CHANGE:
	time.Sleep(1 * time.Nanosecond)

	return *(*[]byte)(unsafe.Pointer(sliceHeader))
}

func readAndHaveFun() {
	reader := bufio.NewReader(os.Stdin)
	count := 1
	var firstChar byte

	for {
		s, _ := reader.ReadString('\n')
		if len(s) == 0 {
			continue
		}
		firstChar = s[0]

		// HERE BE DRAGONS
		bytes := unsafeStringToBytes(&s)

		_, _ = reader.ReadString('\n')

		if len(bytes) > 0 && bytes[0] != firstChar {
			fmt.Printf("win! after %d iterations\n", count)
			os.Exit(0)
		}

		count++
	}
}

func heapHeapHeap() {
	var a *[]byte
	for {
		tmp := make([]byte, 1000000, 1000000)
		a = &tmp
		_ = a
	}
}
