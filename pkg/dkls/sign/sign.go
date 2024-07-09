//
// Copyright Coinbase, Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0
//

// Package sign implements the 2-2 threshold signature protocol of [DKLs18](https://eprint.iacr.org/2018/499.pdf).
// The signing protocol is defined in "Protocol 4" page 9, of the paper. The Zero Knowledge Proof ideal functionalities are
// realized using schnorr proofs.
package sign

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/subtle"
	"fmt"
	"hash"
	"math/big"
	"tecdsa/pkg/dkls/dkg"

	"github.com/gtank/merlin"
	"github.com/pkg/errors"
	"golang.org/x/crypto/sha3"

	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/coinbase/kryptology/pkg/ot/base/simplest"
	"github.com/coinbase/kryptology/pkg/ot/extension/kos"
	"github.com/coinbase/kryptology/pkg/zkp/schnorr"
)

const multiplicationCount = 2

// This implements the Multiplication protocol of DKLs, protocol 5. https://eprint.iacr.org/2018/499.pdf
// two parties---the "sender" and "receiver", let's say---each input a scalar modulo q.
// the functionality multiplies their two scalars modulo q, and then randomly additively shares the product mod q.
// it then returns the two respective additive shares to the two parties.

// MultiplySender is the party that plays the role of Sender in the multiplication protocol (protocol 5 of the paper).
type MultiplySender struct {
	cOtSender           *kos.Sender   // underlying cOT sender struct, used by mult.
	outputAdditiveShare curves.Scalar // ultimate output share of mult.
	gadget              [kos.L]curves.Scalar
	curve               *curves.Curve
	transcript          *merlin.Transcript
	uniqueSessionId     [simplest.DigestSize]byte
}

// MultiplyReceiver is the party that plays the role of Sender in the multiplication protocol (protocol 5 of the paper).
type MultiplyReceiver struct {
	cOtReceiver         *kos.Receiver               // underlying cOT receiver struct, used by mult.
	outputAdditiveShare curves.Scalar               // ultimate output share of mult.
	omega               [kos.COtBlockSizeBytes]byte // this is used as an intermediate result during the course of mult.
	gadget              [kos.L]curves.Scalar
	curve               *curves.Curve
	transcript          *merlin.Transcript
	uniqueSessionId     [simplest.DigestSize]byte
}

func generateGadgetVector(curve *curves.Curve) ([kos.L]curves.Scalar, error) {
	var err error
	gadget := [kos.L]curves.Scalar{}
	for i := 0; i < kos.Kappa; i++ {
		gadget[i], err = curve.Scalar.SetBigInt(new(big.Int).Lsh(big.NewInt(1), uint(i)))
		if err != nil {
			return gadget, errors.Wrap(err, "creating gadget scalar from big int")
		}
	}
	shake := sha3.NewCShake256(nil, []byte("Coinbase DKLs gadget vector"))
	for i := kos.Kappa; i < kos.L; i++ {
		var err error
		bytes := [simplest.DigestSize]byte{}
		if _, err = shake.Read(bytes[:]); err != nil {
			return gadget, err
		}
		gadget[i], err = curve.Scalar.SetBytes(bytes[:])
		if err != nil {
			return gadget, errors.Wrap(err, "creating gadget scalar from bytes")
		}
	}
	return gadget, nil
}

// NewMultiplySender generates a `MultiplySender` instance, ready to take part in multiplication as the "sender".
// You must supply it the _output_ of a seed OT, from the receiver's point of view, as well as params and a unique ID.
// That is, the mult sender must run the base OT as the receiver; note the (apparent) reversal of roles.
func NewMultiplySender(seedOtResults *simplest.ReceiverOutput, curve *curves.Curve, uniqueSessionId [simplest.DigestSize]byte) (*MultiplySender, error) {
	sender := kos.NewCOtSender(seedOtResults, curve)
	gadget, err := generateGadgetVector(curve)
	if err != nil {
		return nil, errors.Wrap(err, "error generating gadget vector in new multiply sender")
	}

	transcript := merlin.NewTranscript("Coinbase_DKLs_Multiply")
	transcript.AppendMessage([]byte("session_id"), uniqueSessionId[:])
	return &MultiplySender{
		cOtSender:       sender,
		curve:           curve,
		transcript:      transcript,
		uniqueSessionId: uniqueSessionId,
		gadget:          gadget,
	}, nil
}

// NewMultiplyReceiver generates a `MultiplyReceiver` instance, ready to take part in multiplication as the "receiver".
// You must supply it the _output_ of a seed OT, from the sender's point of view, as well as params and a unique ID.
// That is, the mult sender must run the base OT as the sender; note the (apparent) reversal of roles.
func NewMultiplyReceiver(seedOtResults *simplest.SenderOutput, curve *curves.Curve, uniqueSessionId [simplest.DigestSize]byte) (*MultiplyReceiver, error) {
	receiver := kos.NewCOtReceiver(seedOtResults, curve)
	gadget, err := generateGadgetVector(curve)
	if err != nil {
		return nil, errors.Wrap(err, "error generating gadget vector in new multiply receiver")
	}
	transcript := merlin.NewTranscript("Coinbase_DKLs_Multiply")
	transcript.AppendMessage([]byte("session_id"), uniqueSessionId[:])
	return &MultiplyReceiver{
		cOtReceiver:     receiver,
		curve:           curve,
		transcript:      transcript,
		uniqueSessionId: uniqueSessionId,
		gadget:          gadget,
	}, nil
}

// MultiplyRound2Output is the output of the second round of the multiplication protocol.
type MultiplyRound2Output struct {
	COTRound2Output *kos.Round2Output
	R               [kos.L]curves.Scalar
	U               curves.Scalar
}

func ReverseScalarBytes(inBytes []byte) []byte {
	outBytes := make([]byte, len(inBytes))

	for i, j := 0, len(inBytes)-1; j >= 0; i, j = i+1, j-1 {
		outBytes[i] = inBytes[j]
	}

	return outBytes
}

// Algorithm 5. in DKLs. "Encodes" Bob's secret input scalars `beta` in the right way, using the opts.
// The idea is that if Bob were to just put beta's as the choice vector, then Alice could learn a few of Bob's bits.
// using selective failure attacks. so you subtract random components of a public random vector. see paper for details.
func (receiver *MultiplyReceiver) encode(beta curves.Scalar) ([kos.COtBlockSizeBytes]byte, error) {
	// passing beta by value, so that we can mutate it locally. check that this does what i want.
	encoding := [kos.COtBlockSizeBytes]byte{}
	bytesOfBetaMinusDotProduct := beta.Bytes()
	if _, err := rand.Read(encoding[kos.KappaBytes:]); err != nil {
		return encoding, errors.Wrap(err, "sampling `gamma` random bytes in multiply receiver encode")
	}
	for j := kos.Kappa; j < kos.L; j++ {
		jthBitOfGamma := simplest.ExtractBitFromByteVector(encoding[:], j)
		// constant-time computation of the dot product beta - < gR, gamma >.
		// we can only `ConstantTimeCopy` byte slices (as opposed to big ints). so keep them as bytes.
		option0, err := receiver.curve.Scalar.SetBytes(bytesOfBetaMinusDotProduct[:])
		if err != nil {
			return encoding, errors.Wrap(err, "setting masking bits scalar from bytes")
		}
		option0Bytes := option0.Bytes()
		option1 := option0.Sub(receiver.gadget[j])
		option1Bytes := option1.Bytes()
		bytesOfBetaMinusDotProduct = option0Bytes
		subtle.ConstantTimeCopy(int(jthBitOfGamma), bytesOfBetaMinusDotProduct[:], option1Bytes)
	}
	copy(encoding[0:kos.KappaBytes], ReverseScalarBytes(bytesOfBetaMinusDotProduct[:]))
	return encoding, nil
}

// Round1Initialize Protocol 5., Multiplication, 3). Bob (receiver) encodes beta and initiates the cOT extension
func (receiver *MultiplyReceiver) Round1Initialize(beta curves.Scalar) (*kos.Round1Output, error) {
	var err error
	if receiver.omega, err = receiver.encode(beta); err != nil {
		return nil, errors.Wrap(err, "encoding input beta in receiver round 1 initialize")
	}
	cOtRound1Output, err := receiver.cOtReceiver.Round1Initialize(receiver.uniqueSessionId, receiver.omega)
	if err != nil {
		return nil, errors.Wrap(err, "error in cOT round 1 initialize within multiply round 1 initialize")
	}
	// write the output of the first round to the transcript
	for i := 0; i < kos.Kappa; i++ {
		label := []byte(fmt.Sprintf("row %d of U", i))
		receiver.transcript.AppendMessage(label, cOtRound1Output.U[i][:])
	}
	receiver.transcript.AppendMessage([]byte("wPrime"), cOtRound1Output.WPrime[:])
	receiver.transcript.AppendMessage([]byte("vPrime"), cOtRound1Output.VPrime[:])
	return cOtRound1Output, nil
}

// Round2Multiply Protocol 5., steps 3) 5), 7). Alice _responds_ to Bob's initial cOT message, using alpha as input.
// Doesn't actually send the message yet, only stashes it and moves onto the next steps of the multiplication protocol
// specifically, Alice can then do step 5) (compute the outputs of the multiplication protocol), also stashes this.
// Finishes by taking care of 7), after that, Alice is totally done with multiplication and has stashed the outputs.
func (sender *MultiplySender) Round2Multiply(alpha curves.Scalar, round1Output *kos.Round1Output) (*MultiplyRound2Output, error) {
	var err error
	alphaHat := sender.curve.Scalar.Random(rand.Reader)
	input := [kos.L][2]curves.Scalar{} // sender's input, namely integer "sums" in case w_j == 1.
	for j := 0; j < kos.L; j++ {
		input[j][0] = alpha
		input[j][1] = alphaHat
	}
	round2Output := &MultiplyRound2Output{}
	round2Output.COTRound2Output, err = sender.cOtSender.Round2Transfer(sender.uniqueSessionId, input, round1Output)
	if err != nil {
		return nil, errors.Wrap(err, "error in cOT within round 2 multiply")
	}
	// write the output of the first round to the transcript
	for i := 0; i < kos.Kappa; i++ {
		label := []byte(fmt.Sprintf("row %d of U", i))
		sender.transcript.AppendMessage(label, round1Output.U[i][:])
	}
	sender.transcript.AppendMessage([]byte("wPrime"), round1Output.WPrime[:])
	sender.transcript.AppendMessage([]byte("vPrime"), round1Output.VPrime[:])
	// write our own output of the second round to the transcript
	chiWidth := 2
	for i := 0; i < kos.Kappa; i++ {
		for k := 0; k < chiWidth; k++ {
			label := []byte(fmt.Sprintf("row %d of Tau", i))
			sender.transcript.AppendMessage(label, round2Output.COTRound2Output.Tau[i][k].Bytes())
		}
	}
	chi := make([]curves.Scalar, chiWidth)
	for k := 0; k < 2; k++ {
		label := []byte(fmt.Sprintf("draw challenge chi %d", k))
		randomBytes := sender.transcript.ExtractBytes(label, kos.KappaBytes)
		chi[k], err = sender.curve.Scalar.SetBytes(randomBytes)
		if err != nil {
			return nil, errors.Wrap(err, "setting chi scalar from bytes")
		}
	}
	sender.outputAdditiveShare = sender.curve.Scalar.Zero()
	for j := 0; j < kos.L; j++ {
		round2Output.R[j] = sender.curve.Scalar.Zero()
		for k := 0; k < chiWidth; k++ {
			round2Output.R[j] = round2Output.R[j].Add(chi[k].Mul(sender.cOtSender.OutputAdditiveShares[j][k]))
		}
		sender.outputAdditiveShare = sender.outputAdditiveShare.Add(sender.gadget[j].Mul(sender.cOtSender.OutputAdditiveShares[j][0]))
	}
	round2Output.U = chi[0].Mul(alpha).Add(chi[1].Mul(alphaHat))
	return round2Output, nil
}

// Round3Multiply Protocol 5., Multiplication, 3) and 6). Bob finalizes the cOT extension.
// using that and Alice's multiplication message, Bob completes the multiplication protocol, including checks.
// At the end, Bob's values tB_j are populated.
func (receiver *MultiplyReceiver) Round3Multiply(round2Output *MultiplyRound2Output) error {
	chiWidth := 2
	// write the output of the second round to the transcript
	for i := 0; i < kos.Kappa; i++ {
		for k := 0; k < chiWidth; k++ {
			label := []byte(fmt.Sprintf("row %d of Tau", i))
			receiver.transcript.AppendMessage(label, round2Output.COTRound2Output.Tau[i][k].Bytes())
		}
	}
	if err := receiver.cOtReceiver.Round3Transfer(round2Output.COTRound2Output); err != nil {
		return errors.Wrap(err, "error within cOT round 3 transfer within round 3 multiply")
	}
	var err error
	chi := make([]curves.Scalar, chiWidth)
	for k := 0; k < chiWidth; k++ {
		label := []byte(fmt.Sprintf("draw challenge chi %d", k))
		randomBytes := receiver.transcript.ExtractBytes(label, kos.KappaBytes)
		chi[k], err = receiver.curve.Scalar.SetBytes(randomBytes)
		if err != nil {
			return errors.Wrap(err, "setting chi scalar from bytes")
		}
	}

	receiver.outputAdditiveShare = receiver.curve.Scalar.Zero()
	for j := 0; j < kos.L; j++ {
		// compute the LHS of bob's step 6) for j. note that we're "adding r_j" to both sides"; so this LHS includes r_j.
		// the reason to do this is so that the constant-time (i.e., independent of w_j) calculation of w_j * u can proceed more cleanly.
		leftHandSideOfCheck := round2Output.R[j]
		for k := 0; k < chiWidth; k++ {
			leftHandSideOfCheck = leftHandSideOfCheck.Add(chi[k].Mul(receiver.cOtReceiver.OutputAdditiveShares[j][k]))
		}
		rightHandSideOfCheck := [simplest.DigestSize]byte{}
		jthBitOfOmega := simplest.ExtractBitFromByteVector(receiver.omega[:], j)
		subtle.ConstantTimeCopy(int(jthBitOfOmega), rightHandSideOfCheck[:], round2Output.U.Bytes())
		if subtle.ConstantTimeCompare(rightHandSideOfCheck[:], leftHandSideOfCheck.Bytes()) != 1 {
			return fmt.Errorf("alice's values R and U failed to check in round 3 multiply")
		}
		receiver.outputAdditiveShare = receiver.outputAdditiveShare.Add(receiver.gadget[j].Mul(receiver.cOtReceiver.OutputAdditiveShares[j][0]))
	}
	return nil
}

// Alice struct encoding Alice's state during one execution of the overall signing algorithm.
// At the end of the joint computation, Alice will not possess the signature.
type Alice struct {
	hash           hash.Hash // which hash function should we use to compute message (i.e, teh digest)
	seedOtResults  *simplest.ReceiverOutput
	secretKeyShare curves.Scalar // the witness
	publicKey      curves.Point
	curve          *curves.Curve
	transcript     *merlin.Transcript
}

// Bob struct encoding Bob's state during one execution of the overall signing algorithm.
// At the end of the joint computation, Bob will obtain the signature.
type Bob struct {
	// Signature is the resulting digital signature and is the output of this protocol.
	Signature *curves.EcdsaSignature

	hash           hash.Hash // which hash function should we use to compute message
	seedOtResults  *simplest.SenderOutput
	secretKeyShare curves.Scalar
	publicKey      curves.Point
	transcript     *merlin.Transcript
	// multiplyReceivers are 2 receivers that are used to perform the two multiplications needed:
	// 1. (phi + 1/kA) * (1/kB)
	// 2. skA/KA * skB/kB
	multiplyReceivers [multiplicationCount]*MultiplyReceiver
	kB                curves.Scalar
	dB                curves.Point
	curve             *curves.Curve
}

// NewAlice creates a party that can participate in protocol runs of DKLs sign, in the role of Alice.
func NewAlice(curve *curves.Curve, hash hash.Hash, dkgOutput *dkg.AliceOutput) *Alice {
	return &Alice{
		hash:           hash,
		seedOtResults:  dkgOutput.SeedOtResult,
		curve:          curve,
		secretKeyShare: dkgOutput.SecretKeyShare,
		publicKey:      dkgOutput.PublicKey,
		transcript:     merlin.NewTranscript("Coinbase_DKLs_Sign"),
	}
}

// NewBob creates a party that can participate in protocol runs of DKLs sign, in the role of Bob.
// This party receives the signature at the end.
func NewBob(curve *curves.Curve, hash hash.Hash, dkgOutput *dkg.BobOutput) *Bob {
	return &Bob{
		hash:           hash,
		seedOtResults:  dkgOutput.SeedOtResult,
		curve:          curve,
		secretKeyShare: dkgOutput.SecretKeyShare,
		publicKey:      dkgOutput.PublicKey,
		transcript:     merlin.NewTranscript("Coinbase_DKLs_Sign"),
	}
}

// SignRound2Output is the output of the 3rd round of the protocol.
type SignRound2Output struct {
	// KosRound1Outputs is the output of the first round of OT Extension, stored for future rounds.
	KosRound1Outputs [multiplicationCount]*kos.Round1Output

	// DB is D_{B} = k_{B} . G from the paper.
	DB curves.Point

	// Seed is the random value used to derive the joint unique session id.
	Seed [simplest.DigestSize]byte
}

// SignRound3Output is the output of the 3rd round of the protocol.
type SignRound3Output struct {
	// MultiplyRound2Outputs is the output of the second round of multiply sub-protocol. Stored to use in future rounds.
	MultiplyRound2Outputs [multiplicationCount]*MultiplyRound2Output

	// RSchnorrProof is ZKP for the value R = k_{A} . D_{B} from the paper.
	RSchnorrProof *schnorr.Proof

	// RPrime is R' = k'_{A} . D_{B} from the paper.
	RPrime curves.Point

	// EtaPhi is the Eta_{Phi} from the paper.
	EtaPhi curves.Scalar

	// EtaSig is the Eta_{Sig} from the paper.
	EtaSig curves.Scalar
}

// Round1GenerateRandomSeed first step of the generation of the shared random salt `idExt`
// in this round, Alice flips 32 random bytes and sends them to Bob.
// Note that this is not _explicitly_ given as part of the protocol in https://eprint.iacr.org/2018/499.pdf, Protocol 1).
// Rather, it is part of our generation of `idExt`, the shared random salt which both parties must use in cOT.
// This value introduced in Protocol 9), very top of page 16. it is not indicated how it should be derived.
// We do it by having each party sample 32 bytes, then by appending _both_ as salts. Secure if either party is honest
func (alice *Alice) Round1GenerateRandomSeed() ([simplest.DigestSize]byte, error) {
	aliceSeed := [simplest.DigestSize]byte{}
	if _, err := rand.Read(aliceSeed[:]); err != nil {
		return [simplest.DigestSize]byte{}, errors.Wrap(err, "generating random bytes in alice round 1 generate")
	}
	alice.transcript.AppendMessage([]byte("session_id_alice"), aliceSeed[:])
	return aliceSeed, nil
}

// Round2Initialize Bob's initial message, which kicks off the signature process. Protocol 1, Bob's steps 1) - 3).
// Bob's work here entails beginning the Diffie–Hellman-like construction of the instance key / nonce,
// as well as preparing the inputs which he will feed into the multiplication protocol,
// and moreover actually initiating the (first respective messages of) the multiplication protocol using these inputs.
// This latter step in turn amounts to sending the initial message in a new cOT extension.
// All the resulting data gets packaged and sent to Alice.
func (bob *Bob) Round2Initialize(aliceSeed [simplest.DigestSize]byte) (*SignRound2Output, error) {
	bobSeed := [simplest.DigestSize]byte{}
	if _, err := rand.Read(bobSeed[:]); err != nil {
		return nil, errors.Wrap(err, "flipping random coins in bob round 2 initialize")
	}
	bob.transcript.AppendMessage([]byte("session_id_alice"), aliceSeed[:])
	bob.transcript.AppendMessage([]byte("session_id_bob"), bobSeed[:])

	var err error
	uniqueSessionId := [simplest.DigestSize]byte{} // will use and _re-use_ this throughout, for sub-session IDs
	copy(uniqueSessionId[:], bob.transcript.ExtractBytes([]byte("multiply receiver id 0"), simplest.DigestSize))
	bob.multiplyReceivers[0], err = NewMultiplyReceiver(bob.seedOtResults, bob.curve, uniqueSessionId)
	if err != nil {
		return nil, errors.Wrap(err, "error creating multiply receiver 0 in Bob sign round 3")
	}
	copy(uniqueSessionId[:], bob.transcript.ExtractBytes([]byte("multiply receiver id 1"), simplest.DigestSize))
	bob.multiplyReceivers[1], err = NewMultiplyReceiver(bob.seedOtResults, bob.curve, uniqueSessionId)
	if err != nil {
		return nil, errors.Wrap(err, "error creating multiply receiver 1 in Bob sign round 3")
	}
	round2Output := &SignRound2Output{
		Seed: bobSeed,
	}
	bob.kB = bob.curve.Scalar.Random(rand.Reader)
	bob.dB = bob.curve.ScalarBaseMult(bob.kB)
	round2Output.DB = bob.dB
	kBInv := bob.curve.Scalar.One().Div(bob.kB)

	round2Output.KosRound1Outputs[0], err = bob.multiplyReceivers[0].Round1Initialize(kBInv)
	if err != nil {
		return nil, errors.Wrap(err, "error in multiply round 1 initialize 0 within Bob sign round 3 initialize")
	}
	round2Output.KosRound1Outputs[1], err = bob.multiplyReceivers[1].Round1Initialize(bob.secretKeyShare.Mul(kBInv))
	if err != nil {
		return nil, errors.Wrap(err, "error in multiply round 1 initialize 1 within Bob sign round 3 initialize")
	}
	return round2Output, nil
}

// Round3Sign Alice's first message. Alice is the _responder_; she is responding to Bob's initial message.
// This is Protocol 1 (p. 6), and contains Alice's steps 3) -- 8). these can all be combined into one message.
// Alice's job here is to finish computing the shared instance key / nonce, as well as multiplication input values;
// then to invoke the multiplication on these two input values (stashing the outputs in her running result struct),
// then to use the _output_ of the multiplication (which she already possesses as of the end of her computation),
// and use that to compute some final values which will help Bob compute the final signature.
func (alice *Alice) Round3Sign(message []byte, round2Output *SignRound2Output) (*SignRound3Output, error) {
	alice.transcript.AppendMessage([]byte("session_id_bob"), round2Output.Seed[:])

	multiplySenders := [multiplicationCount]*MultiplySender{}
	var err error
	uniqueSessionId := [simplest.DigestSize]byte{} // will use and _re-use_ this throughout, for sub-session IDs
	copy(uniqueSessionId[:], alice.transcript.ExtractBytes([]byte("multiply receiver id 0"), simplest.DigestSize))
	if multiplySenders[0], err = NewMultiplySender(alice.seedOtResults, alice.curve, uniqueSessionId); err != nil {
		return nil, errors.Wrap(err, "creating multiply sender 0 in Alice round 4 sign")
	}
	copy(uniqueSessionId[:], alice.transcript.ExtractBytes([]byte("multiply receiver id 1"), simplest.DigestSize))
	if multiplySenders[1], err = NewMultiplySender(alice.seedOtResults, alice.curve, uniqueSessionId); err != nil {
		return nil, errors.Wrap(err, "creating multiply sender 1 in Alice round 4 sign")
	}
	round3Output := &SignRound3Output{}
	kPrimeA := alice.curve.Scalar.Random(rand.Reader)
	round3Output.RPrime = round2Output.DB.Mul(kPrimeA)
	hashRPrimeBytes := sha3.Sum256(round3Output.RPrime.ToAffineCompressed())
	hashRPrime, err := alice.curve.Scalar.SetBytes(hashRPrimeBytes[:])
	if err != nil {
		return nil, errors.Wrap(err, "setting hashRPrime scalar from bytes")
	}
	kA := hashRPrime.Add(kPrimeA)
	copy(uniqueSessionId[:], alice.transcript.ExtractBytes([]byte("schnorr proof for R"), simplest.DigestSize))
	rSchnorrProver := schnorr.NewProver(alice.curve, round2Output.DB, uniqueSessionId[:])
	round3Output.RSchnorrProof, err = rSchnorrProver.Prove(kA)
	if err != nil {
		return nil, errors.Wrap(err, "generating schnorr proof for R = kA * DB in alice round 4 sign")
	}
	// reassign / stash the below value here just for notational clarity.
	// this is _the_ key public point R in the ECDSA signature. we'll use its coordinate X in various places.
	r := round3Output.RSchnorrProof.Statement
	phi := alice.curve.Scalar.Random(rand.Reader)
	kAInv := alice.curve.Scalar.One().Div(kA)

	if round3Output.MultiplyRound2Outputs[0], err = multiplySenders[0].Round2Multiply(phi.Add(kAInv), round2Output.KosRound1Outputs[0]); err != nil {
		return nil, errors.Wrap(err, "error in round 2 multiply 0 within alice round 4 sign")
	}
	if round3Output.MultiplyRound2Outputs[1], err = multiplySenders[1].Round2Multiply(alice.secretKeyShare.Mul(kAInv), round2Output.KosRound1Outputs[1]); err != nil {
		return nil, errors.Wrap(err, "error in round 2 multiply 1 within alice round 4 sign")
	}

	one := alice.curve.Scalar.One()
	gamma1 := alice.curve.ScalarBaseMult(kA.Mul(phi).Add(one))
	other := r.Mul(multiplySenders[0].outputAdditiveShare.Neg())
	gamma1 = gamma1.Add(other)
	hashGamma1Bytes := sha3.Sum256(gamma1.ToAffineCompressed())
	hashGamma1, err := alice.curve.Scalar.SetBytes(hashGamma1Bytes[:])
	if err != nil {
		return nil, errors.Wrap(err, "setting hashGamma1 scalar from bytes")
	}
	round3Output.EtaPhi = hashGamma1.Add(phi)
	if _, err = alice.hash.Write(message); err != nil {
		return nil, errors.Wrap(err, "writing message to hash in alice round 4 sign")
	}
	digest := alice.hash.Sum(nil)
	hOfMAsInteger, err := alice.curve.Scalar.SetBytes(digest)
	if err != nil {
		return nil, errors.Wrap(err, "setting hOfMAsInteger scalar from bytes")
	}
	affineCompressedForm := r.ToAffineCompressed()
	if len(affineCompressedForm) != 33 {
		return nil, errors.New("the compressed form must be exactly 33 bytes")
	}
	// Discard the leading byte and parse the rest as the X coordinate.
	rX, err := alice.curve.Scalar.SetBytes(affineCompressedForm[1:])
	if err != nil {
		return nil, errors.Wrap(err, "setting rX scalar from bytes")
	}

	sigA := hOfMAsInteger.Mul(multiplySenders[0].outputAdditiveShare).Add(rX.Mul(multiplySenders[1].outputAdditiveShare))
	gamma2 := alice.publicKey.Mul(multiplySenders[0].outputAdditiveShare)
	other = alice.curve.ScalarBaseMult(multiplySenders[1].outputAdditiveShare.Neg())
	gamma2 = gamma2.Add(other)
	hashGamma2Bytes := sha3.Sum256(gamma2.ToAffineCompressed())
	hashGamma2, err := alice.curve.Scalar.SetBytes(hashGamma2Bytes[:])
	if err != nil {
		return nil, errors.Wrap(err, "setting hashGamma2 scalar from bytes")
	}
	round3Output.EtaSig = hashGamma2.Add(sigA)
	return round3Output, nil
}

// Round4Final this is Bob's last portion of the signature computation, and ultimately results in the complete signature
// corresponds to Protocol 1, Bob's steps 3) -- 10).
// Bob begins by _finishing_ the OT-based multiplication, using Alice's one and only message to him re: the mult.
// Bob then move's onto the remainder of Alice's message, which contains extraneous data used to finish the signature.
// Using this data, Bob completes the signature, which gets stored in `Bob.Sig`. Bob also verifies it.
func (bob *Bob) Round4Final(message []byte, round3Output *SignRound3Output) error {
	if err := bob.multiplyReceivers[0].Round3Multiply(round3Output.MultiplyRound2Outputs[0]); err != nil {
		return errors.Wrap(err, "error in round 3 multiply 0 within sign round 5")
	}
	if err := bob.multiplyReceivers[1].Round3Multiply(round3Output.MultiplyRound2Outputs[1]); err != nil {
		return errors.Wrap(err, "error in round 3 multiply 1 within sign round 5")
	}
	rPrimeHashedBytes := sha3.Sum256(round3Output.RPrime.ToAffineCompressed())
	rPrimeHashed, err := bob.curve.Scalar.SetBytes(rPrimeHashedBytes[:])
	if err != nil {
		return errors.Wrap(err, "setting rPrimeHashed scalar from bytes")
	}
	r := bob.dB.Mul(rPrimeHashed)
	r = r.Add(round3Output.RPrime)
	// To ensure that the correct public statement is used, we use the public statement that we have calculated
	// instead of the open Alice sent us.
	round3Output.RSchnorrProof.Statement = r
	uniqueSessionId := [simplest.DigestSize]byte{}
	copy(uniqueSessionId[:], bob.transcript.ExtractBytes([]byte("schnorr proof for R"), simplest.DigestSize))
	if err = schnorr.Verify(round3Output.RSchnorrProof, bob.curve, bob.dB, uniqueSessionId[:]); err != nil {
		return errors.Wrap(err, "bob's verification of alice's schnorr proof re: r failed")
	}
	zero := bob.curve.Scalar.Zero()
	affineCompressedForm := r.ToAffineCompressed()
	if len(affineCompressedForm) != 33 {
		return errors.New("the compressed form must be exactly 33 bytes")
	}
	rY := affineCompressedForm[0] & 0x1 // this is bit(0) of Y coordinate
	rX, err := bob.curve.Scalar.SetBytes(affineCompressedForm[1:])
	if err != nil {
		return errors.Wrap(err, "setting rX scalar from bytes")
	}
	bob.Signature = &curves.EcdsaSignature{
		R: rX.Add(zero).BigInt(), // slight trick here; add it to 0 just to mod it by q (now it's mod p!)
		V: int(rY),
	}
	gamma1 := r.Mul(bob.multiplyReceivers[0].outputAdditiveShare)
	gamma1HashedBytes := sha3.Sum256(gamma1.ToAffineCompressed())
	gamma1Hashed, err := bob.curve.Scalar.SetBytes(gamma1HashedBytes[:])
	if err != nil {
		return errors.Wrap(err, "setting gamma1Hashed scalar from bytes")
	}
	phi := round3Output.EtaPhi.Sub(gamma1Hashed)
	theta := bob.multiplyReceivers[0].outputAdditiveShare.Sub(phi.Div(bob.kB))
	if _, err = bob.hash.Write(message); err != nil {
		return errors.Wrap(err, "writing message to hash in Bob sign round 5 final")
	}
	digestBytes := bob.hash.Sum(nil)
	digest, err := bob.curve.Scalar.SetBytes(digestBytes)
	if err != nil {
		return errors.Wrap(err, "setting digest scalar from bytes")
	}
	capitalR, err := bob.curve.Scalar.SetBigInt(bob.Signature.R)
	if err != nil {
		return errors.Wrap(err, "setting capitalR scalar from big int")
	}
	sigB := digest.Mul(theta).Add(capitalR.Mul(bob.multiplyReceivers[1].outputAdditiveShare))
	gamma2 := bob.curve.ScalarBaseMult(bob.multiplyReceivers[1].outputAdditiveShare)
	other := bob.publicKey.Mul(theta.Neg())
	gamma2 = gamma2.Add(other)
	gamma2HashedBytes := sha3.Sum256(gamma2.ToAffineCompressed())
	gamma2Hashed, err := bob.curve.Scalar.SetBytes(gamma2HashedBytes[:])
	if err != nil {
		return errors.Wrap(err, "setting gamma2Hashed scalar from bytes")
	}
	scalarS := sigB.Add(round3Output.EtaSig.Sub(gamma2Hashed))
	bob.Signature.S = scalarS.BigInt()
	if bob.Signature.S.Bit(255) == 1 {
		bob.Signature.S = scalarS.Neg().BigInt()
		bob.Signature.V ^= 1
	}
	// now verify the signature
	unCompressedAffinePublicKey := bob.publicKey.ToAffineUncompressed()
	if len(unCompressedAffinePublicKey) != 65 {
		return errors.New("the uncompressed form must have exactly 65 bytes")
	}
	x := new(big.Int).SetBytes(unCompressedAffinePublicKey[1:33])
	y := new(big.Int).SetBytes(unCompressedAffinePublicKey[33:])
	ellipticCurve, err := bob.curve.ToEllipticCurve()
	if err != nil {
		return errors.Wrap(err, "invalid curve")
	}
	if !ecdsa.Verify(&ecdsa.PublicKey{Curve: ellipticCurve, X: x, Y: y}, digestBytes, bob.Signature.R, bob.Signature.S) {
		return fmt.Errorf("final signature failed to verify")
	}
	return nil
}
