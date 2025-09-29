package config

import (
	cryptoRand "crypto/rand"
	"math/big"
)

type adminPasswordGenerator struct{}

func (a *adminPasswordGenerator) generatePassword(desiredLength int) string {
	/// Password policy:
	// - must contain at least one uppercase, one lowercase, one digit, one special character

	lower := "abcdefghijklmnopqrstuvwxyz"
	upper := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digits := "0123456789"
	special := "!@#$%^&*()-_=+[]{}<>?,.:;"

	all := lower + upper + digits + special

	// Helper to draw a random byte from a charset using crypto/rand without modulo bias.
	pick := func(charset string) byte {
		maxIndex := big.NewInt(int64(len(charset)))
		for {
			n, err := cryptoRand.Int(cryptoRand.Reader, maxIndex)
			if err != nil {
				// In case of unexpected RNG error, fall back to 'a' to avoid panic.
				return charset[0]
			}
			idx := n.Int64()
			if idx >= 0 && idx < int64(len(charset)) {
				return charset[idx]
			}
		}
	}

	// Ensure at least one character from each required class.
	pwd := make([]byte, 0, desiredLength)
	pwd = append(pwd, pick(lower))
	pwd = append(pwd, pick(upper))
	pwd = append(pwd, pick(digits))
	pwd = append(pwd, pick(special))

	// Fill the rest with random characters from the full set.
	for len(pwd) < desiredLength {
		pwd = append(pwd, pick(all))
	}

	// Secure Fisher–Yates shuffle using crypto/rand.
	for i := len(pwd) - 1; i > 0; i-- {
		jBig, err := cryptoRand.Int(cryptoRand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			continue
		}
		j := int(jBig.Int64())
		pwd[i], pwd[j] = pwd[j], pwd[i]
	}

	return string(pwd)
}
