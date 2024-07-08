package dkg

import (
	"crypto/rand"

	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/coinbase/kryptology/pkg/ot/base/simplest"
	"github.com/coinbase/kryptology/pkg/ot/extension/kos"
	"github.com/coinbase/kryptology/pkg/zkp/schnorr"
	"github.com/gtank/merlin"
	"github.com/pkg/errors"
)

const (
	// keyCount is the number of encryption keys created. Since this is a 1-out-of-2 OT, the key count is set to 2.
	keyCount = 2

	// DigestSize is the length of hash. Similarly, when it comes to encrypting and decryption, it is the size of the
	// plaintext and ciphertext.
	DigestSize = 32
)

// AliceOutput is the result of running DKG for Alice. It contains both the public and secret values that are needed
// for signing.
type AliceOutput struct {
	// PublicKey is the joint public key of Alice and Bob.
	// This value is public.
	PublicKey curves.Point

	// SecretKeyShare is Alice's secret key for the joint public key.
	// This output must be kept secret. If it is lost, the users will lose access and cannot create signatures.
	SecretKeyShare curves.Scalar

	// SeedOtResult are the outputs that the receiver will obtain as a result of running the "random" OT protocol.
	// This output must be kept secret. Although, if it is lost the users can run another OT protocol and obtain
	// new values to replace it.
	SeedOtResult *simplest.ReceiverOutput
}

// BobOutput is the result of running DKG for Bob. It contains both the public and secret values that are needed
// for signing.
type BobOutput struct {
	// PublicKey is the joint public key of Alice and Bob.
	// This value is public.
	PublicKey curves.Point

	// SecretKeyShare is Bob's secret key for the joint public key.
	// This output must be kept secret. If it is lost, the users will lose access and cannot create signatures.
	SecretKeyShare curves.Scalar

	// SeedOtResult are the outputs that the sender will obtain as a result of running the "random" OT protocol.
	// This output must be kept secret. Although, if it is lost the users can run another OT protocol and obtain
	// new values to replace it.
	SeedOtResult *simplest.SenderOutput
}
type Alice struct {
	prover         *schnorr.Prover
	proof          *schnorr.Proof
	receiver       *simplest.Receiver
	secretKeyShare curves.Scalar
	publicKey      curves.Point
	curve          *curves.Curve // Add this line
	transcript     *merlin.Transcript
}

type Bob struct {
	prover          *schnorr.Prover
	sender          *simplest.Sender
	secretKeyShare  curves.Scalar
	publicKey       curves.Point
	aliceCommitment schnorr.Commitment
	aliceSalt       [simplest.DigestSize]byte
	curve           *curves.Curve
	transcript      *merlin.Transcript
}

type Round2Output struct {
	// Seed is the random value used to derive the joint unique session id.
	Seed [simplest.DigestSize]byte

	// Commitment is the commitment to the ZKP to Alice's secret key share.
	Commitment schnorr.Commitment
}

type Proof struct {
	C         curves.Scalar
	S         curves.Scalar
	Statement curves.Point
}

type (
	// OneTimePadDecryptionKey is the type of Rho^w, Rho^0, and RHo^1 in the paper.
	OneTimePadDecryptionKey = [DigestSize]byte

	// OneTimePadEncryptionKeys is the type of Rho^0, and RHo^1 in the paper.
	OneTimePadEncryptionKeys = [keyCount][DigestSize]byte

	// OtChallenge is the type of xi in the paper.
	OtChallenge = [DigestSize]byte

	// OtChallengeResponse is the type of Rho' in the paper.
	OtChallengeResponse = [DigestSize]byte

	// ChallengeOpening is the type of hashed Rho^0 and Rho^1
	ChallengeOpening = [keyCount][DigestSize]byte

	// ReceiversMaskedChoices corresponds to the "A" value in the paper in compressed format.
	ReceiversMaskedChoices = []byte
)

func NewAlice(curve *curves.Curve) *Alice {
	return &Alice{
		curve:      curve,
		transcript: merlin.NewTranscript("Coinbase_DKLs_DKG"),
	}
}

func NewBob(curve *curves.Curve) *Bob {
	return &Bob{
		curve:      curve,
		transcript: merlin.NewTranscript("Coinbase_DKLs_DKG"),
	}
}

func (bob *Bob) Round1GenerateRandomSeed() ([simplest.DigestSize]byte, error) {
	bobSeed := [simplest.DigestSize]byte{}
	if _, err := rand.Read(bobSeed[:]); err != nil {
		return [simplest.DigestSize]byte{}, errors.Wrap(err, "generating random bytes in bob DKG round 1 generate")
	}
	bob.transcript.AppendMessage([]byte("session_id_bob"), bobSeed[:])
	return bobSeed, nil
}

func (alice *Alice) Round2CommitToProof(bobSeed [simplest.DigestSize]byte) (*Round2Output, error) {
	aliceSeed := [simplest.DigestSize]byte{}
	if _, err := rand.Read(aliceSeed[:]); err != nil {
		return nil, errors.Wrap(err, "generating random bytes in alice DKG round 2 generate")
	}
	alice.transcript.AppendMessage([]byte("session_id_bob"), bobSeed[:])
	alice.transcript.AppendMessage([]byte("session_id_alice"), aliceSeed[:])

	uniqueSessionId := [simplest.DigestSize]byte{}
	copy(uniqueSessionId[:], alice.transcript.ExtractBytes([]byte("salt for simplest OT"), simplest.DigestSize))
	alice.receiver, _ = simplest.NewReceiver(alice.curve, kos.Kappa, uniqueSessionId)

	alice.secretKeyShare = alice.curve.Scalar.Random(rand.Reader)
	copy(uniqueSessionId[:], alice.transcript.ExtractBytes([]byte("salt for alice schnorr"), simplest.DigestSize))
	alice.prover = schnorr.NewProver(alice.curve, nil, uniqueSessionId[:])
	var commitment schnorr.Commitment
	alice.proof, commitment, _ = alice.prover.ProveCommit(alice.secretKeyShare)

	return &Round2Output{
		Commitment: commitment,
		Seed:       aliceSeed,
	}, nil
}

// Implement other rounds similarly

// Round3SchnorrProve receives Bob's Commitment and returns schnorr statment + proof.
// Steps 1 and 3 of protocol 2 on page 7.
func (bob *Bob) Round3SchnorrProve(round2Output *Round2Output) (*schnorr.Proof, error) {
	bob.transcript.AppendMessage([]byte("session_id_alice"), round2Output.Seed[:])

	bob.aliceCommitment = round2Output.Commitment // store it, so that we can check when alice decommits

	var err error
	uniqueSessionId := [simplest.DigestSize]byte{} // note: will use and re-use this below for sub-session IDs.
	copy(uniqueSessionId[:], bob.transcript.ExtractBytes([]byte("salt for simplest OT"), simplest.DigestSize))
	bob.sender, err = simplest.NewSender(bob.curve, kos.Kappa, uniqueSessionId)
	if err != nil {
		return nil, errors.Wrap(err, "bob constructing new OT sender in DKG round 2")
	}
	// extract alice's salt in the right order; we won't use this until she reveals her proof and we verify it below
	copy(bob.aliceSalt[:], bob.transcript.ExtractBytes([]byte("salt for alice schnorr"), simplest.DigestSize))
	bob.secretKeyShare = bob.curve.Scalar.Random(rand.Reader)
	copy(uniqueSessionId[:], bob.transcript.ExtractBytes([]byte("salt for bob schnorr"), simplest.DigestSize))
	bob.prover = schnorr.NewProver(bob.curve, nil, uniqueSessionId[:])
	proof, err := bob.prover.Prove(bob.secretKeyShare)
	if err != nil {
		return nil, errors.Wrap(err, "bob schnorr proving in DKG round 2")
	}
	return proof, err
}

// Round4VerifyAndReveal step 4 of protocol 2 on page 7.
func (alice *Alice) Round4VerifyAndReveal(proof *schnorr.Proof) (*schnorr.Proof, error) {
	var err error
	uniqueSessionId := [simplest.DigestSize]byte{}
	copy(uniqueSessionId[:], alice.transcript.ExtractBytes([]byte("salt for bob schnorr"), simplest.DigestSize))
	if err = schnorr.Verify(proof, alice.curve, nil, uniqueSessionId[:]); err != nil {
		return nil, errors.Wrap(err, "alice's verification of Bob's schnorr proof failed in DKG round 3")
	}
	alice.publicKey = proof.Statement.Mul(alice.secretKeyShare)
	return alice.proof, nil
}

// Round5DecommitmentAndStartOt step 5 of protocol 2 on page 7.
func (bob *Bob) Round5DecommitmentAndStartOt(proof *schnorr.Proof) (*schnorr.Proof, error) {
	var err error
	if err = schnorr.DecommitVerify(proof, bob.aliceCommitment, bob.curve, nil, bob.aliceSalt[:]); err != nil {
		return nil, errors.Wrap(err, "decommit + verify failed in bob's DKG round 4")
	}
	bob.publicKey = proof.Statement.Mul(bob.secretKeyShare)
	seedOTRound1Output, err := bob.sender.Round1ComputeAndZkpToPublicKey()
	if err != nil {
		return nil, errors.Wrap(err, "bob computing round 1 of seed  OT within DKG round 4")
	}
	return seedOTRound1Output, nil
}

// Round6DkgRound2Ot is a thin wrapper around the 2nd round of seed OT protocol.
func (alice *Alice) Round6DkgRound2Ot(proof *schnorr.Proof) ([]simplest.ReceiversMaskedChoices, error) {
	return alice.receiver.Round2VerifySchnorrAndPadTransfer(proof)
}

// Round7DkgRound3Ot is a thin wrapper around the 3rd round of seed OT protocol.
func (bob *Bob) Round7DkgRound3Ot(compressedReceiversMaskedChoice []simplest.ReceiversMaskedChoices) ([]simplest.OtChallenge, error) {
	return bob.sender.Round3PadTransfer(compressedReceiversMaskedChoice)
}

// Round8DkgRound4Ot is a thin wrapper around the 4th round of seed OT protocol.
func (alice *Alice) Round8DkgRound4Ot(challenge []simplest.OtChallenge) ([]simplest.OtChallengeResponse, error) {
	return alice.receiver.Round4RespondToChallenge(challenge)
}

// Round9DkgRound5Ot is a thin wrapper around the 5th round of seed OT protocol.
func (bob *Bob) Round9DkgRound5Ot(challengeResponses []simplest.OtChallengeResponse) ([]simplest.ChallengeOpening, error) {
	return bob.sender.Round5Verify(challengeResponses)
}

// Round10DkgRound6Ot is a thin wrapper around the 6th round of seed OT protocol.
func (alice *Alice) Round10DkgRound6Ot(challengeOpenings []simplest.ChallengeOpening) error {
	return alice.receiver.Round6Verify(challengeOpenings)
}

// Output returns the output of the DKG operation. Must be called after step 9. Calling it before that step
// has undefined behaviour.
func (alice *Alice) Output() *AliceOutput {
	return &AliceOutput{
		PublicKey:      alice.publicKey,
		SecretKeyShare: alice.secretKeyShare,
		SeedOtResult:   alice.receiver.Output,
	}
}

// Output returns the output of the DKG operation. Must be called after step 9. Calling it before that step
// has undefined behaviour.
func (bob *Bob) Output() *BobOutput {
	return &BobOutput{
		PublicKey:      bob.publicKey,
		SecretKeyShare: bob.secretKeyShare,
		SeedOtResult:   bob.sender.Output,
	}
}
