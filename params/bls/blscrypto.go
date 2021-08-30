package blscrypto

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"reflect"

	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	serializedPublicKeyT = reflect.TypeOf(SerializedPublicKey{})
	serializedSignatureT = reflect.TypeOf(SerializedSignature{})
)

type SerializedPublicKey [PUBLICKEYBYTES]byte

// EpochEntropyFromHash truncates the given hash to the length of epoch SNARK entropy.
func EpochEntropyFromHash(hash common.Hash) EpochEntropy {
	var entropy EpochEntropy
	copy(entropy[:], hash[:EPOCHENTROPYBYTES])
	return entropy
}

// MarshalText returns the hex representation of pk.
func (pk SerializedPublicKey) MarshalText() ([]byte, error) {
	return hexutil.Bytes(pk[:]).MarshalText()
}

// UnmarshalText parses a BLS public key in hex syntax.
func (pk *SerializedPublicKey) UnmarshalText(input []byte) error {
	return hexutil.UnmarshalFixedText("SerializedPublicKey", input, pk[:])
}

// UnmarshalJSON parses a BLS public key in hex syntax.
func (pk *SerializedPublicKey) UnmarshalJSON(input []byte) error {
	return hexutil.UnmarshalFixedJSON(serializedPublicKeyT, input, pk[:])
}

type SerializedSignature [SIGNATUREBYTES]byte

// MarshalText returns the hex representation of sig.
func (sig SerializedSignature) MarshalText() ([]byte, error) {
	return hexutil.Bytes(sig[:]).MarshalText()
}

// UnmarshalText parses a BLS signature in hex syntax.
func (sig *SerializedSignature) UnmarshalText(input []byte) error {
	return hexutil.UnmarshalFixedText("SerializedSignature", input, sig[:])
}

// UnmarshalJSON parses a BLS signature in hex syntax.
func (sig *SerializedSignature) UnmarshalJSON(input []byte) error {
	return hexutil.UnmarshalFixedJSON(serializedSignatureT, input, sig[:])
}

func ECDSAToBLS(privateKeyECDSA *ecdsa.PrivateKey) ([]byte, error) {
	for i := 0; i < 256; i++ {
		modulus := big.NewInt(0)
		modulus, ok := modulus.SetString(MODULUS377, 10)
		if !ok {
			return nil, errors.New("can't parse modulus")
		}
		privateKeyECDSABytes := crypto.FromECDSA(privateKeyECDSA)

		keyBytes := []byte("ecdsatobls")
		keyBytes = append(keyBytes, uint8(i))
		keyBytes = append(keyBytes, privateKeyECDSABytes...)

		privateKeyBLSBytes := crypto.Keccak256(keyBytes)
		privateKeyBLSBytes[0] &= MODULUSMASK
		privateKeyBLSBig := big.NewInt(0)
		privateKeyBLSBig.SetBytes(privateKeyBLSBytes)
		if privateKeyBLSBig.Cmp(modulus) >= 0 {
			continue
		}

		privateKeyBytes := privateKeyBLSBig.Bytes()
		for len(privateKeyBytes) < len(privateKeyBLSBytes) {
			privateKeyBytes = append([]byte{0x00}, privateKeyBytes...)
		}
		if !bytes.Equal(privateKeyBLSBytes, privateKeyBytes) {
			return nil, fmt.Errorf("private key bytes should have been the same: %s, %s", hex.EncodeToString(privateKeyBLSBytes), hex.EncodeToString(privateKeyBytes))
		}
		// reverse order, as the BLS library expects little endian
		for i := len(privateKeyBytes)/2 - 1; i >= 0; i-- {
			opp := len(privateKeyBytes) - 1 - i
			privateKeyBytes[i], privateKeyBytes[opp] = privateKeyBytes[opp], privateKeyBytes[i]
		}

		privateKeyBLS, err := DeserializePrivateKey(privateKeyBytes)
		if err != nil {
			return nil, err
		}
		defer privateKeyBLS.Destroy()
		privateKeyBLSBytesFromLib, err := privateKeyBLS.Serialize()
		if err != nil {
			return nil, err
		}
		if !bytes.Equal(privateKeyBytes, privateKeyBLSBytesFromLib) {
			return nil, errors.New("private key bytes from library should have been the same")
		}

		return privateKeyBLSBytesFromLib, nil
	}

	return nil, errors.New("couldn't derive a BLS key from an ECDSA key")
}

func PrivateToPublic(privateKeyBytes []byte) (SerializedPublicKey, error) {
	privateKey, err := DeserializePrivateKey(privateKeyBytes)
	if err != nil {
		return SerializedPublicKey{}, err
	}
	defer privateKey.Destroy()

	publicKey, err := privateKey.ToPublic()
	if err != nil {
		return SerializedPublicKey{}, err
	}
	defer publicKey.Destroy()

	pubKeyBytes, err := publicKey.Serialize()
	if err != nil {
		return SerializedPublicKey{}, err
	}

	pubKeyBytesFixed := SerializedPublicKey{}
	copy(pubKeyBytesFixed[:], pubKeyBytes)

	return pubKeyBytesFixed, nil
}

func VerifyAggregatedSignature(publicKeys []SerializedPublicKey, message []byte, extraData []byte, signature []byte, shouldUseCompositeHasher, cip22 bool) error {
	publicKeyObjs := []*PublicKey{}
	for _, publicKey := range publicKeys {
		publicKeyObj, err := DeserializePublicKeyCached(publicKey[:])
		if err != nil {
			return err
		}
		defer publicKeyObj.Destroy()
		publicKeyObjs = append(publicKeyObjs, publicKeyObj)
	}
	apk, err := AggregatePublicKeys(publicKeyObjs)
	if err != nil {
		return err
	}
	defer apk.Destroy()

	signatureObj, err := DeserializeSignature(signature)
	if err != nil {
		return err
	}
	defer signatureObj.Destroy()

	err = apk.VerifySignature(message, extraData, signatureObj, shouldUseCompositeHasher, cip22)
	return err
}

func AggregateSignatures(signatures [][]byte) ([]byte, error) {
	signatureObjs := []*Signature{}
	for _, signature := range signatures {
		signatureObj, err := DeserializeSignature(signature)
		if err != nil {
			return nil, err
		}
		defer signatureObj.Destroy()
		signatureObjs = append(signatureObjs, signatureObj)
	}

	asig, err := AggregateSignatures2(signatureObjs)
	if err != nil {
		return nil, err
	}
	defer asig.Destroy()

	asigBytes, err := asig.Serialize()
	if err != nil {
		return nil, err
	}

	return asigBytes, nil
}

func VerifySignature(publicKey SerializedPublicKey, message []byte, extraData []byte, signature []byte, shouldUseCompositeHasher, cip22 bool) error {
	publicKeyObj, err := DeserializePublicKeyCached(publicKey[:])
	if err != nil {
		return err
	}
	defer publicKeyObj.Destroy()

	signatureObj, err := DeserializeSignature(signature)
	if err != nil {
		return err
	}
	defer signatureObj.Destroy()

	err = publicKeyObj.VerifySignature(message, extraData, signatureObj, shouldUseCompositeHasher, cip22)
	return err
}

func EncodeEpochSnarkData(newValSet []SerializedPublicKey, maximumNonSigners uint32, epochIndex uint16) ([]byte, []byte, error) {
	pubKeys := []*PublicKey{}
	for _, pubKey := range newValSet {
		publicKeyObj, err := DeserializePublicKeyCached(pubKey[:])
		if err != nil {
			return nil, nil, err
		}
		defer publicKeyObj.Destroy()

		pubKeys = append(pubKeys, publicKeyObj)
	}

	message, err := EncodeEpochToBytes(epochIndex, maximumNonSigners, pubKeys)
	return message, nil, err
}

func EncodeEpochSnarkDataCIP22(newValSet []SerializedPublicKey, maximumNonSigners, maxValidators uint32, epochIndex uint16, round uint8, blockHash, parentHash EpochEntropy) ([]byte, []byte, error) {
	pubKeys := []*PublicKey{}
	for _, pubKey := range newValSet {
		publicKeyObj, err := DeserializePublicKeyCached(pubKey[:])
		if err != nil {
			return nil, nil, err
		}
		defer publicKeyObj.Destroy()

		pubKeys = append(pubKeys, publicKeyObj)
	}

	return EncodeEpochToBytesCIP22(epochIndex, round, blockHash, parentHash, maximumNonSigners, maxValidators, pubKeys)
}

func SerializedSignatureFromBytes(serializedSignature []byte) (SerializedSignature, error) {
	if len(serializedSignature) != SIGNATUREBYTES {
		return SerializedSignature{}, fmt.Errorf("wrong length for serialized signature: expected %d, got %d", SIGNATUREBYTES, len(serializedSignature))
	}
	signatureBytesFixed := SerializedSignature{}
	copy(signatureBytesFixed[:], serializedSignature)
	return signatureBytesFixed, nil
}

func UncompressKey(serialized SerializedPublicKey) ([]byte, error) {
	publicKey, err := DeserializePublicKeyCached(serialized[:])
	if err != nil {
		return nil, err
	}
	uncompressedBytes, err := publicKey.SerializeUncompressed()
	if err != nil {
		return nil, err
	}
	return uncompressedBytes, nil
}
