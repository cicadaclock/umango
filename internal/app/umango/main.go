package main

import (
	"fmt"

	"github.com/cicadaclock/umango/internal/pkg/calculator"
	"github.com/cicadaclock/umango/internal/pkg/structs"
)

func main() {
	uma_1 := structs.Uma{
		Id:   1001,
		Name: "Special Week",
	}
	uma_2 := structs.Uma{
		Id:   1002,
		Name: "Silence Suzuka",
	}
	fmt.Println(calculator.CalculateDuoAffinity(uma_1, uma_2))
}
