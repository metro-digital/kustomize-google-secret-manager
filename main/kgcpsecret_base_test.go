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

var base_secret_values = map[string]string{
	"secret1": "secret1-42",
	"secret2": "secret2-42",
	"secret3": "val-secret3",
}

func getBaseTestValue(ctx context.Context, client *secretmanager.Client, plugin *KGCPSecret, key string) (string, error) {

	for k, v := range base_secret_values {
		if key == k {
			return v, nil
		}
	}

	return "", errors.New("no value found for key")
}

func getBaseTestKeys(project_id string) ([]string, error) {
	keys := []string{}
	for k := range base_secret_values {
		keys = append(keys, k)
	}
	return keys, nil
}

var _ = Describe("when creating a Kubernetes secret from an KGCPSecret with minimal data", func() {
	encryptedSecret := KGCPSecret{
		TypeMeta: TypeMeta{
			APIVersion: "metro.digital/v1",
			Kind:       "KGCPSecret",
		},
		GCPObjectMeta: GCPObjectMeta{
			Name:        "my-secret",
			Labels:      map[string]string{},
			Annotations: map[string]string{},
		},
		GCPProjectID:          "cf-2tier-uhd-test-d7",
		DisableNameSuffixHash: true,
		Keys: []string{
			"secret1",
			"secret2",
		},
	}

	expected := K8SSecret{
		TypeMeta: TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: ObjectMeta{
			Name:        "my-secret",
			Namespace:   "",
			Labels:      map[string]string{},
			Annotations: map[string]string{},
		},
		Data: map[string]string{
			"secret1": "secret1-42",
			"secret2": "secret2-42",
		},
		Type: "",
	}

	It("should create a correct K8S secret", func() {
		actual, err := GetSecrets(ctx, nil, &encryptedSecret, getBaseTestKeys, getBaseTestValue)

		Expect(err).ToNot(HaveOccurred())
		Expect(actual).To(Equal(expected))
	})
})

var _ = Describe("when creating a Kubernetes secret from an KGCPSecret with non existing secret key", func() {
	encryptedSecret := KGCPSecret{
		TypeMeta: TypeMeta{
			APIVersion: "metro.digital/v1",
			Kind:       "KGCPSecret",
		},
		GCPObjectMeta: GCPObjectMeta{
			Name:        "my-secret",
			Labels:      map[string]string{},
			Annotations: map[string]string{},
		},
		GCPProjectID:          "cf-2tier-uhd-test-d7",
		DisableNameSuffixHash: true,
		Keys: []string{
			"do-not-exist",
		},
	}

	It("should create an error", func() {
		_, err := GetSecrets(ctx, nil, &encryptedSecret, getBaseTestKeys, getBaseTestValue)

		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("when creating a Kubernetes secret from an KGCPSecret with full data", func() {
	encryptedSecret := KGCPSecret{
		TypeMeta: TypeMeta{
			APIVersion: "metro.digital/v1",
			Kind:       "KGCPSecret",
		},
		GCPObjectMeta: GCPObjectMeta{
			Name:      "my-secret",
			Namespace: "my-namespace",
			Labels: map[string]string{
				"label1": "lvalue1",
				"label2": "lvalue2",
			},
			Annotations: map[string]string{
				"annotation1": "avalue1",
				"annotation2": "avalue2",
			},
		},
		GCPProjectID:          "cf-2tier-uhd-test-d7",
		DisableNameSuffixHash: false,
		Type:                  "opaque",
		Behavior:              "replace",
		Keys: []string{
			"secret1",
			"secret2",
			"secret3",
		},
	}

	expected := K8SSecret{
		TypeMeta: TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: ObjectMeta{
			Name:      "my-secret",
			Namespace: "my-namespace",
			Labels: map[string]string{
				"label1": "lvalue1",
				"label2": "lvalue2",
			},
			Annotations: map[string]string{
				"annotation1":                        "avalue1",
				"annotation2":                        "avalue2",
				"kustomize.config.k8s.io/needs-hash": "true",
				"kustomize.config.k8s.io/behavior":   "replace",
			},
		},
		Data: map[string]string{
			"secret1": "secret1-42",
			"secret2": "secret2-42",
			"secret3": "val-secret3",
		},
		Type: "opaque",
	}

	It("should create a correct K8S secret", func() {
		actual, err := GetSecrets(ctx, nil, &encryptedSecret, getBaseTestKeys, getBaseTestValue)

		Expect(err).ToNot(HaveOccurred())
		Expect(actual).To(Equal(expected))
	})
})
