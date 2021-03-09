package main

import (
	"fmt"
	"os"
)

var (
	x = make([]byte, 2, 3)
)

func main() {
	appends(x)
	b2, _ := os.ReadFile("./test")
	fmt.Println(b2)
}
func appends(x []byte) int {
	y := append(x, byte(1), byte(2), byte(1), byte(2), byte(1), byte(2))
	os.WriteFile("./test", append(y, []byte("test123")...), 0666)
	return 0
}
