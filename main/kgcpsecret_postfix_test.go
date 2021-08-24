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
// +build unitTests

package main_test

import (
	"context"
	"errors"

	. "github.com/metro-digital/kustomize-google-secret-manager/main"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
)

var post_fix_secret_values = map[string]string{
	"KUBERNETES_URL_pp":          "https://kubernetes-pp.metro.digital",
	"KUBERNETES_URL_prod":        "https://kubernetes-prod.metro.digital",
	"CDN_URL":                    "https://europe.cdn.net",
	"CDN_URL_cn-tcs1":            "https://asia.cdn.net",
	"CDN_URL_ru-tcm1":            "https://russia.cdn.net",
	"CASSANDRA_URL_pp":           "cassandra-pp.be-gcw1.metro.digital",
	"CASSANDRA_URL_pp_cn-tcs1":   "cassandra-pp.cn-tcs1.metro.digital",
	"CASSANDRA_URL_pp_ru-tcm1":   "cassandra-pp.ru-tcm1.metro.digital",
	"CASSANDRA_URL_prod":         "cassandra-prod.be-gcw1.metro.digital",
	"CASSANDRA_URL_prod_cn-tcs1": "cassandra-prod.cn-tcs1.metro.digital",
	"CASSANDRA_URL_prod_ru-tcm1": "cassandra-prod.ru-tcm1.metro.digital",
}

func getPostfixTestValue(ctx context.Context, client *secretmanager.Client, plugin *KGCPSecret, key string) (string, error) {
	for k, v := range post_fix_secret_values {
		if key == k {
			return v, nil
		}
	}

	return "", errors.New("no value found for key")
}

func getPostfixTestKeys(project_id string) ([]string, error) {
	keys := []string{}
	for k := range post_fix_secret_values {
		keys = append(keys, k)
	}
	return keys, nil
}

var _ = Describe("when creating a Kubernetes secret with different values for stages", func() {

	It("should use the correspondig stage data value for every secret", func() {
		name := "my-secret"
		key := "KUBERNETES_URL"
		encryptedSecret := createEncryptedGCPSecret(name, key)

		encryptedSecret.Stage = "pp"
		encryptedSecret.Dc = "be-gcw1"
		value := "https://kubernetes-pp.metro.digital"
		expected := createExpectedK8SSecret(name, key, value)
		actual, err := GetSecrets(ctx, nil, &encryptedSecret, getPostfixTestKeys, getPostfixTestValue)
		Expect(err).ToNot(HaveOccurred())
		Expect(actual).To(Equal(expected))

		encryptedSecret.Stage = "prod"
		value = "https://kubernetes-prod.metro.digital"
		expected = createExpectedK8SSecret(name, key, value)
		actual, err = GetSecrets(ctx, nil, &encryptedSecret, getPostfixTestKeys, getPostfixTestValue)
		Expect(err).ToNot(HaveOccurred())
		Expect(actual).To(Equal(expected))
	})
})

var _ = Describe("when creating a Kubernetes secret with different values for data-centers", func() {

	It("should use the correspondig stage data value for every secret", func() {
		name := "my-secret"
		key := "CDN_URL"
		encryptedSecret := createEncryptedGCPSecret(name, key)

		encryptedSecret.Stage = "prod"
		encryptedSecret.Dc = "be-gcw1"
		value := "https://europe.cdn.net"
		expected := createExpectedK8SSecret(name, key, value)
		actual, err := GetSecrets(ctx, nil, &encryptedSecret, getPostfixTestKeys, getPostfixTestValue)
		Expect(err).ToNot(HaveOccurred())
		Expect(actual).To(Equal(expected))

		encryptedSecret.Stage = "prod"
		encryptedSecret.Dc = "nl-gcw4"
		value = "https://europe.cdn.net"
		expected = createExpectedK8SSecret(name, key, value)
		actual, err = GetSecrets(ctx, nil, &encryptedSecret, getPostfixTestKeys, getPostfixTestValue)
		Expect(err).ToNot(HaveOccurred())
		Expect(actual).To(Equal(expected))

		encryptedSecret.Stage = "prod"
		encryptedSecret.Dc = "cn-tcs1"
		value = "https://asia.cdn.net"
		expected = createExpectedK8SSecret(name, key, value)
		actual, err = GetSecrets(ctx, nil, &encryptedSecret, getPostfixTestKeys, getPostfixTestValue)
		Expect(err).ToNot(HaveOccurred())
		Expect(actual).To(Equal(expected))

		encryptedSecret.Stage = "prod"
		encryptedSecret.Dc = "ru-tcm1"
		value = "https://russia.cdn.net"
		expected = createExpectedK8SSecret(name, key, value)
		actual, err = GetSecrets(ctx, nil, &encryptedSecret, getPostfixTestKeys, getPostfixTestValue)
		Expect(err).ToNot(HaveOccurred())
		Expect(actual).To(Equal(expected))

	})
})

var _ = Describe("when creating a Kubernetes secret with different values for stages and data-centers", func() {

	It("should use the most specific data value for every secret", func() {
		name := "my-secret"
		key := "CASSANDRA_URL"
		encryptedSecret := createEncryptedGCPSecret(name, key)

		encryptedSecret.Stage = "pp"
		encryptedSecret.Dc = "be-gcw1"
		value := "cassandra-pp.be-gcw1.metro.digital"
		expected := createExpectedK8SSecret(name, key, value)
		actual, err := GetSecrets(ctx, nil, &encryptedSecret, getPostfixTestKeys, getPostfixTestValue)
		Expect(err).ToNot(HaveOccurred())
		Expect(actual).To(Equal(expected))

		encryptedSecret.Stage = "prod"
		encryptedSecret.Dc = "be-gcw1"
		value = "cassandra-prod.be-gcw1.metro.digital"
		expected = createExpectedK8SSecret(name, key, value)
		actual, err = GetSecrets(ctx, nil, &encryptedSecret, getPostfixTestKeys, getPostfixTestValue)
		Expect(err).ToNot(HaveOccurred())
		Expect(actual).To(Equal(expected))

		encryptedSecret.Stage = "pp"
		encryptedSecret.Dc = "nl-gcw4"
		value = "cassandra-pp.be-gcw1.metro.digital"
		expected = createExpectedK8SSecret(name, key, value)
		actual, err = GetSecrets(ctx, nil, &encryptedSecret, getPostfixTestKeys, getPostfixTestValue)
		Expect(err).ToNot(HaveOccurred())
		Expect(actual).To(Equal(expected))

		encryptedSecret.Stage = "prod"
		encryptedSecret.Dc = "nl-gcw4"
		value = "cassandra-prod.be-gcw1.metro.digital"
		expected = createExpectedK8SSecret(name, key, value)
		actual, err = GetSecrets(ctx, nil, &encryptedSecret, getPostfixTestKeys, getPostfixTestValue)
		Expect(err).ToNot(HaveOccurred())
		Expect(actual).To(Equal(expected))

		encryptedSecret.Stage = "pp"
		encryptedSecret.Dc = "cn-tcs1"
		value = "cassandra-pp.cn-tcs1.metro.digital"
		expected = createExpectedK8SSecret(name, key, value)
		actual, err = GetSecrets(ctx, nil, &encryptedSecret, getPostfixTestKeys, getPostfixTestValue)
		Expect(err).ToNot(HaveOccurred())
		Expect(actual).To(Equal(expected))

		encryptedSecret.Stage = "prod"
		encryptedSecret.Dc = "cn-tcs1"
		value = "cassandra-prod.cn-tcs1.metro.digital"
		expected = createExpectedK8SSecret(name, key, value)
		actual, err = GetSecrets(ctx, nil, &encryptedSecret, getPostfixTestKeys, getPostfixTestValue)
		Expect(err).ToNot(HaveOccurred())
		Expect(actual).To(Equal(expected))

		encryptedSecret.Stage = "pp"
		encryptedSecret.Dc = "ru-tcm1"
		value = "cassandra-pp.ru-tcm1.metro.digital"
		expected = createExpectedK8SSecret(name, key, value)
		actual, err = GetSecrets(ctx, nil, &encryptedSecret, getPostfixTestKeys, getPostfixTestValue)
		Expect(err).ToNot(HaveOccurred())
		Expect(actual).To(Equal(expected))

		encryptedSecret.Stage = "prod"
		encryptedSecret.Dc = "ru-tcm1"
		value = "cassandra-prod.ru-tcm1.metro.digital"
		expected = createExpectedK8SSecret(name, key, value)
		actual, err = GetSecrets(ctx, nil, &encryptedSecret, getPostfixTestKeys, getPostfixTestValue)
		Expect(err).ToNot(HaveOccurred())
		Expect(actual).To(Equal(expected))

	})
})
