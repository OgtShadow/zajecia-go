package main

import (
	"fmt"
)

type MyReader struct{}

func (r MyReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 'A'
	}
	return len(p), nil
}

func main() {
	buf := make([]byte, 64)
	n, err := MyReader{}.Read(buf)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Printf("read %d bytes: %s\n", n, string(buf[:n]))
}
