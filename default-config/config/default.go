package config

import (
	"context"
	"fmt"
)

const passwordLength = 20

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

type globalConfigWriter interface {
	applyDefaultGlobalConfig(ctx context.Context, defaultGlobalConfig map[string]string) error
}

type doguConfigWriter interface {
	applyDefaultDoguConfig(ctx context.Context, defaultDoguConfig map[string]map[string]string, sensitiveDefaultDoguConfig map[string]map[string]string) error
}

type passwordGenerator interface {
	generatePassword(length int) string
}

type DefaultConfigApplier struct {
	globalConfigWriter globalConfigWriter
	doguConfigWriter   doguConfigWriter
	passwordGenerator  passwordGenerator
}

func NewDefaultConfigApplier(globalConfigRepo globalConfigRepo, doguConfigRepo doguConfigRepo, sensitiveDoguConfigRepo doguConfigRepo) *DefaultConfigApplier {
	gcw := &cesGlobalConfigWriter{
		globalConfigRepo: globalConfigRepo,
	}
	dcw := &cesDoguConfigWriter{
		doguConfigRepo:          doguConfigRepo,
		sensitiveDoguConfigRepo: sensitiveDoguConfigRepo,
	}

	return &DefaultConfigApplier{
		globalConfigWriter: gcw,
		doguConfigWriter:   dcw,
		passwordGenerator:  &adminPasswordGenerator{},
	}
}

func (dca *DefaultConfigApplier) ApplyDefaultConfig(ctx context.Context) error {

	if err := dca.globalConfigWriter.applyDefaultGlobalConfig(ctx, globalDefaults); err != nil {
		return fmt.Errorf("failed to apply default global config: %w", err)
	}

	sensitiveDoguDefaults := map[string]map[string]string{
		"ldap": {
			"admin_password": dca.passwordGenerator.generatePassword(passwordLength),
		},
	}

	if err := dca.doguConfigWriter.applyDefaultDoguConfig(ctx, doguDefaults, sensitiveDoguDefaults); err != nil {
		return fmt.Errorf("failed to apply default dogu config: %w", err)
	}

	return nil
}
