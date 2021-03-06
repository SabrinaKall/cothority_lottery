package protocol_test

/*
The test-file should at the very least run the protocol for a varying number
of nodes. It is even better practice to test the different methods of the
protocol, as in Test Driven Development.
*/

import (
	"testing"
	"time"

	"github.com/SabrinaKall/cothority_lottery/protocol"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3/suites"
	"go.dedis.ch/onet/v3"
	"go.dedis.ch/onet/v3/log"
	"go.dedis.ch/onet/v3/network"
)

var tSuite = suites.MustFind("Ed25519")

func TestMain(m *testing.M) {
	log.MainTest(m)
}

// Tests a 2, 5 and 13-node system. It is good practice to test different
// sizes of trees to make sure your protocol is stable.
func TestNode(t *testing.T) {

	log.SetDebugVisible(1)
	nodes := []int{13}
	for _, nbrNodes := range nodes {
		local := onet.NewLocalTest(tSuite)
		_, _, tree := local.GenTree(nbrNodes, true)
		log.Lvl3(tree.Dump())

		pi, err := local.StartProtocol("Lottery", tree)
		require.Nil(t, err)
		protocol := pi.(*protocol.LotteryProtocol)
		timeout := network.WaitRetry * time.Duration(network.MaxRetryConnect*nbrNodes*2) * time.Millisecond
		select {
		case highestTicket := <-protocol.Ticket:
			log.Lvl2("Instance 1 is done")
			t.Log("Number of nodes: ", nbrNodes)
			t.Log("Highest found number: ", highestTicket.Number)
			t.Log("Highest number owner ID: ", highestTicket.OwnerID)

		case <-time.After(timeout):
			t.Fatal("Didn't finish in time")
		}
		local.CloseAll()
	}
}
