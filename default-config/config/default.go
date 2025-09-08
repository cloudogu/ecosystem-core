package config

import (
	"context"
	cryptoRand "crypto/rand"
	"fmt"
	"math/big"
)

var globalDefaults = map[string]string{
	"domain":              "ces.local",
	"admin_group":         "cesAdmin",
	"mail_address":        "",
	"certificateType":     "selfsigned",
	"default_dogu":        "cas",
	"k8s/use_internal_ip": "false",
	"k8s/internal_ip":     "",

	"password-policy/must_contain_capital_letter":    "true",
	"password-policy/must_contain_lower_case_letter": "true",
	"password-policy/must_contain_digit":             "true",
	"password-policy/must_contain_special_character": "true",
	"password-policy/min_length":                     "14",
}

var doguDefaults = map[string]map[string]string{
	"postfix": {
		"relayhost": "n/a",
	},
	"ldap": {
		"admin_username": "admin",
		"admin_mail":     "admin@ces.invalid",
		"admin_member":   "true",
	},
	"cas": {
		"ldap/ds_type":            "embedded",
		"ldap/host":               "ldap",
		"ldap/port":               "389",
		"ldap/search_filter":      "(objectClass=person)",
		"ldap/attribute_id":       "uid",
		"ldap/attribute_mail":     "mail",
		"ldap/attribute_fullname": "cn",
		"ldap/attribute_group":    "memberOf",
	},
}

func ApplyDefaultConfig(ctx context.Context, globalConfigRepo globalConfigRepo, doguConfigRepo doguConfigRepo, sensitiveDoguConfigRepo doguConfigRepo) error {

	gcw := globalConfigWriter{
		globalConfigRepo: globalConfigRepo,
	}
	dcw := doguConfigWriter{
		doguConfigRepo:          doguConfigRepo,
		sensitiveDoguConfigRepo: sensitiveDoguConfigRepo,
	}

	if err := gcw.applyDefaultGlobalConfig(ctx, globalDefaults); err != nil {
		return fmt.Errorf("failed to apply default global config: %w", err)
	}

	sensitiveDoguDefaults := map[string]map[string]string{
		"ldap": {
			"admin_password": generateAdminPassword(),
		},
	}

	if err := dcw.applyDefaultDoguConfig(ctx, doguDefaults, sensitiveDoguDefaults); err != nil {
		return fmt.Errorf("failed to apply default dogu config: %w", err)
	}

	return nil
}

func generateAdminPassword() string {
	// Password policy:
	// - must contain at least one uppercase, one lowercase, one digit, one special character
	const desiredLength = 20

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
