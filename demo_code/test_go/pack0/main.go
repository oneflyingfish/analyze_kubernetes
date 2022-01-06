package pack0

import "fmt"

var Value int = -1

func init() {
	Value = 0
	fmt.Printf("value is init in pack0 with %d\n", Value)
}

func init() {
	fmt.Println(("ttttttt"))
}

func PrintValue() {
	fmt.Printf("value is %d\n", Value)
}
