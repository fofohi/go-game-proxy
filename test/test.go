package main

import "fmt"

type  test1 interface {
	read() int
}

type test2 struct {
	abc int
}

type test3 struct {
	abc int
}

func (t2 *test2) read() int{
	fmt.Println(t2.abc)
	return 0
}

func (t3 *test3)read() int{
	fmt.Println(t3.abc)
	return 0
}

func t4(test1 test1)  {
	test1.read()
}

func main()  {

	test2 := &test2{2}
	t4(test2)
}