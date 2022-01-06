package main

import (
	"test_go/pack0"
	"test_go/pack1"
	"test_go/pack2"
)

func main() {
	pack0.PrintValue()
	pack1.ChangeValue(1)
	pack0.PrintValue()
	pack2.ChangeValue(2)
	pack0.PrintValue()
}
