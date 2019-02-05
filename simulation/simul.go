package main

import (
	// Service needs to be imported here to be instantiated.
	_ "github.com/SabrinaKall/cothority_lottery/service"
	"go.dedis.ch/onet/v3/simul"
)

func main() {
	simul.Start()
}
