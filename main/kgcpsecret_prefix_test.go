//
// Copyright 2021 METRO Digital GmbH
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
//go:build unitTests
// +build unitTests

package main_test

import (
	"context"
	"errors"
	"encoding/base64"

	. "github.com/metro-digital/kustomize-google-secret-manager/main"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
)

var prefix_secret_values = map[string]string{
	"VALUE1":                        "main-VALUE1-value",
	"my-secret_VALUE1":              "my-secret-VALUE1-value",
	"VALUE2":                        "main-VALUE1-value",
	"my-namespace_VALUE2":           "my-namespace-VALUE2-value",
	"VALUE3":                        "main-VALUE3-value",
	"my-namespace_my-secret_VALUE3": "my-namespace-and-secret-VALUE3-value",
	"my-namespace_VALUE3":           "my-namespace-VALUE3-value",
	"my-secret_VALUE3":              "my-secret-VALUE3-value",
	"dockercfg1_data":               "my-dockercfg1-data-value",
	"dockercfg2_data":               "my-dockercfg2-data-value",
}

func getPrefixTestValue(ctx context.Context, client *secretmanager.Client, plugin *KGCPSecret, key string) (string, error) {
	for k, v := range prefix_secret_values {
		if key == k {
			return v, nil
		}
	}

	return "", errors.New("no value found for key")
}

func getPrefixTestKeys(project_id string) ([]string, error) {
	keys := []string{}
	for k := range prefix_secret_values {
		keys = append(keys, k)
	}
	return keys, nil
}

var _ = Describe("when creating a Kubernetes secret", func() {

	Describe("with a secret manager not containing a secret prefixed with secret name", func() {
		name := "your-secret"
		key := "VALUE1"
		value := base64.StdEncoding.EncodeToString([]byte("main-VALUE1-value"))
		encryptedSecret := createEncryptedGCPSecret(name, key)
		expected := createExpectedK8SSecret(name, key, value)

		It("should use the secret without prefix", func() {
			actual, err := GetSecrets(ctx, nil, &encryptedSecret, getPrefixTestKeys, getPrefixTestValue)

			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(Equal(expected))
		})
	})

	Describe("with a secret manager containing a secret prefixed with secret name", func() {
		name := "my-secret"
		key := "VALUE1"
		value := base64.StdEncoding.EncodeToString([]byte("my-secret-VALUE1-value"))
		encryptedSecret := createEncryptedGCPSecret(name, key)
		expected := createExpectedK8SSecret(name, key, value)

		It("should use the secret with secret name prefix", func() {
			actual, err := GetSecrets(ctx, nil, &encryptedSecret, getPrefixTestKeys, getPrefixTestValue)

			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(Equal(expected))
		})
	})

	Describe("with a secret manager containing a secret prefixed with namespace name", func() {
		name := "my-secret"
		key := "VALUE2"
		value := base64.StdEncoding.EncodeToString([]byte("my-namespace-VALUE2-value"))
		encryptedSecret := createEncryptedGCPSecret(name, key)
		encryptedSecret.Namespace = "my-namespace"
		expected := createExpectedK8SSecret(name, key, value)
		expected.Namespace = "my-namespace"

		It("should use the secret with namespace prefix", func() {
			actual, err := GetSecrets(ctx, nil, &encryptedSecret, getPrefixTestKeys, getPrefixTestValue)

			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(Equal(expected))
		})
	})

	Describe("with a secret manager containing a secret prefixed with namespace and secret name", func() {
		name := "my-secret"
		key := "VALUE3"
		value := base64.StdEncoding.EncodeToString([]byte("my-namespace-and-secret-VALUE3-value"))
		encryptedSecret := createEncryptedGCPSecret(name, key)
		encryptedSecret.Namespace = "my-namespace"
		expected := createExpectedK8SSecret(name, key, value)
		expected.Namespace = "my-namespace"

		It("should use the most specific secret", func() {
			actual, err := GetSecrets(ctx, nil, &encryptedSecret, getPrefixTestKeys, getPrefixTestValue)

			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(Equal(expected))
		})
	})
})

var _ = Describe("when creating a Kubernetes secret of type 'dockercfg'", func() {

	It("should use the correspondig data value for every secret", func() {
		name := "dockercfg1"
		key := "data"
		value := base64.StdEncoding.EncodeToString([]byte("my-dockercfg1-data-value"))
		encryptedSecret := createEncryptedGCPSecret(name, key)
		expected := createExpectedK8SSecret(name, key, value)

		actual, err := GetSecrets(ctx, nil, &encryptedSecret, getPrefixTestKeys, getPrefixTestValue)

		Expect(err).ToNot(HaveOccurred())
		Expect(actual).To(Equal(expected))

		name = "dockercfg2"
		key = "data"
		value = base64.StdEncoding.EncodeToString([]byte("my-dockercfg2-data-value"))
		encryptedSecret = createEncryptedGCPSecret(name, key)
		expected = createExpectedK8SSecret(name, key, value)

		actual, err = GetSecrets(ctx, nil, &encryptedSecret, getPrefixTestKeys, getPrefixTestValue)

		Expect(err).ToNot(HaveOccurred())
		Expect(actual).To(Equal(expected))
	})
})
