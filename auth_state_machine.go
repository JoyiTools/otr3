package otr3

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"strconv"
)

// ignoreMessage should never be called with a too small message buffer, it is assumed the caller will have checked this before calling it
func (c *akeContext) ignoreMessage(msg []byte) bool {
	_, protocolVersion, _ := extractShort(msg)
	unexpectedV2Msg := protocolVersion == 2 && !c.has(allowV2)
	unexpectedV3Msg := protocolVersion == 3 && !c.has(allowV3)

	return unexpectedV2Msg || unexpectedV3Msg
}

const minimumMessageLength = 3 // length of protocol version (SHORT) and message type (BYTE)

func (c *akeContext) receiveMessage(msg []byte) (toSend []byte, err error) {
	if len(msg) < minimumMessageLength {
		return nil, errInvalidOTRMessage
	}

	if c.ignoreMessage(msg) {
		return
	}

	switch msg[2] {
	case msgTypeDHCommit:
		c.authState, toSend, err = c.authState.receiveDHCommitMessage(c, msg)
	case msgTypeDHKey:
		c.authState, toSend, err = c.authState.receiveDHKeyMessage(c, msg)
	case msgTypeRevealSig:
		c.authState, toSend, err = c.authState.receiveRevealSigMessage(c, msg)
		if err == nil {
			c.msgState = encrypted
		}
	case msgTypeSig:
		c.authState, toSend, err = c.authState.receiveSigMessage(c, msg)
		if err == nil {
			c.msgState = encrypted
		}
	default:
		err = fmt.Errorf("otr: unknown message type 0x%X", msg[2])
	}

	return
}

func (c *akeContext) receiveQueryMessage(msg []byte) (toSend []byte, err error) {
	c.authState, toSend, err = c.authState.receiveQueryMessage(c, msg)

	if err == nil {
		c.ourKeyID = 0
		c.ourCurrentDHKeys = dhKeyPair{}
	}

	return
}

type authStateBase struct{}
type authStateNone struct{ authStateBase }
type authStateAwaitingDHKey struct{ authStateBase }
type authStateAwaitingRevealSig struct{ authStateBase }
type authStateAwaitingSig struct{ authStateBase }
type authStateV1Setup struct{ authStateBase }

type authState interface {
	receiveQueryMessage(*akeContext, []byte) (authState, []byte, error)
	receiveDHCommitMessage(*akeContext, []byte) (authState, []byte, error)
	receiveDHKeyMessage(*akeContext, []byte) (authState, []byte, error)
	receiveRevealSigMessage(*akeContext, []byte) (authState, []byte, error)
	receiveSigMessage(*akeContext, []byte) (authState, []byte, error)
}

func (authStateBase) receiveQueryMessage(c *akeContext, msg []byte) (authState, []byte, error) {
	return authStateNone{}.receiveQueryMessage(c, msg)
}

func (authStateBase) receiveDHCommitMessage(c *akeContext, msg []byte) (authState, []byte, error) {
	return authStateNone{}.receiveDHCommitMessage(c, msg)
}

func (s authStateNone) receiveQueryMessage(c *akeContext, msg []byte) (authState, []byte, error) {
	v, ok := s.acceptOTRRequest(c.policies, msg)
	if !ok {
		return nil, nil, errInvalidVersion
	}

	//TODO set the version for every existing otrContext
	c.version = v
	c.senderInstanceTag = generateInstanceTag()

	out, err := c.dhCommitMessage()
	if err != nil {
		return s, nil, err
	}

	return authStateAwaitingDHKey{}, out, nil
}

func (authStateNone) parseOTRQueryMessage(msg []byte) []int {
	ret := []int{}

	if bytes.HasPrefix(msg, queryMarker) && len(msg) > len(queryMarker) {
		versions := msg[len(queryMarker):]

		if versions[0] == '?' {
			ret = append(ret, 1)
			versions = versions[1:]
		}

		if len(versions) > 0 && versions[0] == 'v' {
			for _, c := range versions {
				if v, err := strconv.Atoi(string(c)); err == nil {
					ret = append(ret, v)
				}
			}
		}
	}

	return ret
}

func (s authStateNone) acceptOTRRequest(p policies, msg []byte) (otrVersion, bool) {
	versions := s.parseOTRQueryMessage(msg)

	for _, v := range versions {
		switch {
		case v == 3 && p.has(allowV3):
			return otrV3{}, true
		case v == 2 && p.has(allowV2):
			return otrV2{}, true
		}
	}

	return nil, false
}

func (s authStateNone) receiveDHCommitMessage(c *akeContext, msg []byte) (authState, []byte, error) {
	if err := generateCommitMsgInstanceTags(c, msg); err != nil {
		return s, nil, err
	}

	ret, err := c.dhKeyMessage()
	if err != nil {
		return s, nil, err
	}

	if err = c.processDHCommit(msg); err != nil {
		return s, nil, err
	}

	c.ourKeyID = 1

	return authStateAwaitingRevealSig{}, ret, nil
}

func generateCommitMsgInstanceTags(ake *akeContext, msg []byte) error {
	if ake.version.needInstanceTag() {
		if len(msg) < lenMsgHeader+4 {
			return errInvalidOTRMessage
		}

		_, receiverInstanceTag, _ := extractWord(msg[lenMsgHeader:])
		ake.senderInstanceTag = generateInstanceTag()
		ake.receiverInstanceTag = receiverInstanceTag
	}
	return nil
}

func generateInstanceTag() uint32 {
	//TODO generate this
	return 0x00000100 + 0x01
}

func (s authStateAwaitingRevealSig) receiveDHCommitMessage(c *akeContext, msg []byte) (authState, []byte, error) {
	//Forget the DH-commit received before we sent the DH-Key

	if err := c.processDHCommit(msg); err != nil {
		return s, nil, err
	}

	//TODO: this should not change my instanceTag, since this is supposed to be a retransmit
	// We can ignore errors from this function, since processDHCommit checks for the sameconditions
	generateCommitMsgInstanceTags(c, msg)

	return authStateAwaitingRevealSig{}, c.serializeDHKey(), nil
}

func (s authStateAwaitingDHKey) receiveDHCommitMessage(c *akeContext, msg []byte) (authState, []byte, error) {
	if len(msg) < c.version.headerLen() {
		return s, nil, errInvalidOTRMessage
	}

	newMsg, _, ok1 := extractData(msg[c.version.headerLen():])
	_, theirHashedGx, ok2 := extractData(newMsg)

	if !ok1 || !ok2 {
		return s, nil, errInvalidOTRMessage
	}

	gxMPI := appendMPI(nil, c.theirPublicValue)
	hashedGx := sha256.Sum256(gxMPI)
	if bytes.Compare(hashedGx[:], theirHashedGx) == 1 {
		//NOTE what about the sender and receiver instance tags?
		return authStateAwaitingRevealSig{}, c.serializeDHCommit(c.theirPublicValue), nil
	}

	return authStateNone{}.receiveDHCommitMessage(c, msg)
}

func (s authStateNone) receiveDHKeyMessage(c *akeContext, msg []byte) (authState, []byte, error) {
	return s, nil, nil
}

func (s authStateAwaitingRevealSig) receiveDHKeyMessage(c *akeContext, msg []byte) (authState, []byte, error) {
	return s, nil, nil
}

func (s authStateAwaitingDHKey) receiveDHKeyMessage(c *akeContext, msg []byte) (authState, []byte, error) {
	_, err := c.processDHKey(msg)
	if err != nil {
		return s, nil, err
	}

	if c.revealSigMsg, err = c.revealSigMessage(); err != nil {
		return s, nil, err
	}

	c.theirCurrentDHPubKey = c.theirPublicValue
	c.ourCurrentDHKeys.pub = c.ourPublicValue
	c.ourCurrentDHKeys.priv = c.secretExponent
	c.ourCounter++

	return authStateAwaitingSig{}, c.revealSigMsg, nil
}

func (s authStateAwaitingSig) receiveDHKeyMessage(c *akeContext, msg []byte) (authState, []byte, error) {
	isSame, err := c.processDHKey(msg)
	if err != nil {
		return s, nil, err
	}

	if isSame {
		return s, c.revealSigMsg, nil
	}

	return s, nil, nil
}

func (s authStateNone) receiveRevealSigMessage(c *akeContext, msg []byte) (authState, []byte, error) {
	return s, nil, nil
}

func (s authStateAwaitingRevealSig) receiveRevealSigMessage(c *akeContext, msg []byte) (authState, []byte, error) {
	if !c.has(allowV2) {
		return s, nil, nil
	}

	err := c.processRevealSig(msg)

	if err != nil {
		return nil, nil, err
	}

	ret, err := c.sigMessage()
	if err != nil {
		return s, nil, err
	}

	//TODO: check if theirKeyID (or the previous) mathches what we have stored for this
	c.ourKeyID = 0
	c.theirCurrentDHPubKey = c.theirPublicValue
	c.theirPreviousDHPubKey = nil

	c.ourCurrentDHKeys.priv = c.secretExponent
	c.ourCurrentDHKeys.pub = c.ourPublicValue
	c.ourCounter++

	return authStateNone{}, ret, nil
}

func (s authStateAwaitingDHKey) receiveRevealSigMessage(c *akeContext, msg []byte) (authState, []byte, error) {
	return s, nil, nil
}

func (s authStateAwaitingSig) receiveRevealSigMessage(c *akeContext, msg []byte) (authState, []byte, error) {
	return s, nil, nil
}

func (s authStateNone) receiveSigMessage(c *akeContext, msg []byte) (authState, []byte, error) {
	return s, nil, nil
}

func (s authStateAwaitingRevealSig) receiveSigMessage(c *akeContext, msg []byte) (authState, []byte, error) {
	return s, nil, nil
}

func (s authStateAwaitingDHKey) receiveSigMessage(c *akeContext, msg []byte) (authState, []byte, error) {
	return s, nil, nil
}

func (s authStateAwaitingSig) receiveSigMessage(c *akeContext, msg []byte) (authState, []byte, error) {
	if !c.has(allowV2) {
		return s, nil, nil
	}

	err := c.processSig(msg)

	if err != nil {
		return nil, nil, err
	}

	//gy was stored when we receive DH-Key
	c.theirCurrentDHPubKey = c.theirPublicValue
	c.theirPreviousDHPubKey = nil
	c.ourKeyID = 0

	return authStateNone{}, nil, nil
}

func (authStateNone) String() string              { return "AUTHSTATE_NONE" }
func (authStateAwaitingDHKey) String() string     { return "AUTHSTATE_AWAITING_DHKEY" }
func (authStateAwaitingRevealSig) String() string { return "AUTHSTATE_AWAITING_REVEALSIG" }
func (authStateAwaitingSig) String() string       { return "AUTHSTATE_AWAITING_SIG" }
func (authStateV1Setup) String() string           { return "AUTHSTATE_V1_SETUP" }

//TODO need to implements AUTHSTATE_V1_SETUP
