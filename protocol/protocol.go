package protocol

/*
The `NewProtocol` method is used to define the protocol and to register
the handlers that will be called if a certain type of message is received.
The handlers will be treated according to their signature.

The protocol-file defines the actions that the protocol needs to do in each
step. The root-node will call the `Start`-method of the protocol. Each
node will only use the `Handle`-methods, and not call `Start` again.
*/

import (
	"errors"
	"math/rand"

	"go.dedis.ch/onet/v3"
	"go.dedis.ch/onet/v3/log"
	"go.dedis.ch/onet/v3/network"
)

func init() {
	network.RegisterMessage(Announce{})
	network.RegisterMessage(Reply{})
	onet.GlobalProtocolRegister(Name, NewProtocol)
}

// LotteryTicket holds a "lottery ticket"
//
// Has a number and the id of the original owner node
//
type LotteryTicket struct {
	Number int
	//OwnerID onet.TreeNodeID
	OwnerID int
}

// LotteryProtocol holds the state of a given protocol.
//
// For this example, it defines a channel that will receive the biggest lottery number among
// children. Only the root-node will write to the channel.
type LotteryProtocol struct {
	*onet.TreeNodeInstance
	Ticket chan LotteryTicket
}

// Check that *TemplateProtocol implements onet.ProtocolInstance
var _ onet.ProtocolInstance = (*LotteryProtocol)(nil)

// NewProtocol initialises the structure for use in one round
func NewProtocol(n *onet.TreeNodeInstance) (onet.ProtocolInstance, error) {
	t := &LotteryProtocol{
		TreeNodeInstance: n,
		Ticket:           make(chan LotteryTicket),
	}
	for _, handler := range []interface{}{t.HandleAnnounce, t.HandleReply} {
		if err := t.RegisterHandler(handler); err != nil {
			return nil, errors.New("couldn't register handler: " + err.Error())
		}
	}
	return t, nil
}

// Start sends the Announce-message to all children
func (p *LotteryProtocol) Start() error {
	log.Lvl3("Starting LotteryProtocol")
	return p.HandleAnnounce(StructAnnounce{p.TreeNode(),
		Announce{"cothority rulez!"}})
}

// HandleAnnounce is the first message and is used to send an ID that
// is stored in all nodes.
func (p *LotteryProtocol) HandleAnnounce(msg StructAnnounce) error {
	log.Lvl3("Parent announces:", msg.Message)
	if !p.IsLeaf() {
		// If we have children, send the same message to all of them
		p.SendToChildren(&msg.Announce)
	} else {
		// If we're the leaf, start to reply
		p.HandleReply(nil)
	}
	return nil
}

// HandleReply is the message going up the tree and holding a comparator to pick the biggest lottery number.
func (p *LotteryProtocol) HandleReply(reply []StructReply) error {
	defer p.Done()

	ownNumber := rand.Intn(100)
	//ownID := p.TreeNode().ID
	ownID := p.TreeNode().RosterIndex

	log.Lvl1(ownID, " draws lottery number ", ownNumber)

	for _, c := range reply {
		if c.LotteryNumber > ownNumber {
			ownNumber = c.LotteryNumber
			ownID = c.OwnerID
		}
	}
	log.Lvl1(p.ServerIdentity().Address, "is done with lottery number", ownNumber)
	if !p.IsRoot() {
		log.Lvl3("Sending to parent")
		return p.SendTo(p.Parent(), &Reply{ownNumber, ownID})
	}
	log.Lvl3("Root-node is done - biggest lottery number found:", ownNumber)
	p.Ticket <- LotteryTicket{ownNumber, ownID}
	return nil
}
