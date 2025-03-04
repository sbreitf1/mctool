package main

import (
	"fmt"

	"github.com/sbreitf1/mctool/pkg/mclib/nbt"
)

func main() {
	nbtFile, err := nbt.ReadFromFile(`C:\Users\simon\AppData\Roaming\.minecraft\saves\world\level.dat`)
	fmt.Println(nbtFile, err)
}
