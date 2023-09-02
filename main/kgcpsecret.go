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

package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"google.golang.org/api/iterator"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"

	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type kvMap map[string]string

// TypeMeta defines the resource type
type TypeMeta struct {
	APIVersion string `json:"apiVersion" yaml:"apiVersion"`
	Kind       string `json:"kind" yaml:"kind"`
}

// ObjectMeta contains Kubernetes resource metadata such as the name
type ObjectMeta struct {
	Name        string `json:"name" yaml:"name"`
	Namespace   string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Labels      kvMap  `json:"labels,omitempty" yaml:"labels,omitempty"`
	Annotations kvMap  `json:"annotations,omitempty" yaml:"annotations,omitempty"`
}

// GCPObjectMeta contains the meta data for KGCPSecret
type GCPObjectMeta struct {
	Environment string `json:"environment" yaml:"environment"`
	Tag         string `json:"tag" yaml:"tag"`
	Dc          string `json:"dc" yaml:"dc"`
	Stage       string `json:"stage" yaml:"stage"`
	Name        string `json:"name" yaml:"name"`
	Namespace   string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Labels      kvMap  `json:"labels,omitempty" yaml:"labels,omitempty"`
	Annotations kvMap  `json:"annotations,omitempty" yaml:"annotations,omitempty"`
}

// KGCPSecret is data used to generate a secret
type KGCPSecret struct {
	TypeMeta              `json:",inline" yaml:",inline"`
	GCPObjectMeta         `json:"metadata" yaml:"metadata"`
	GCPProjectID          string   `json:"gcpProjectID,omitempty" yaml:"gcpProjectID,omitempty"`
	DisableNameSuffixHash bool     `json:"disableNameSuffixHash,omitempty" yaml:"disableNameSuffixHash,omitempty"`
	Type                  string   `json:"type,omitempty" yaml:"type,omitempty"`
	Behavior              string   `json:"behavior,omitempty" yaml:"behavior,omitempty"`
	Keys                  []string `json:"keys,omitempty" yaml:"keys,omitempty"`
	DataType              string   `json:"dataType,omitempty" yaml:"dataType,omitempty"`
}

// K8SSecret is a Kubernetes Secret
type K8SSecret struct {
	TypeMeta   `json:",inline" yaml:",inline"`
	ObjectMeta `json:"metadata" yaml:"metadata"`
	Data       kvMap  `json:"data" yaml:"data"`
	Type       string `json:"type,omitempty" yaml:"type,omitempty"`
}

func main() {
	if len(os.Args) != 2 {
		_, _ = fmt.Fprintln(os.Stderr, "usage: KGCPSecret FILE")
		os.Exit(1)
	}

	output, err := processEncryptedGCPSecret(os.Args[1])
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(2)
	}
	fmt.Print(output)
}

func processEncryptedGCPSecret(fn string) (string, error) {
	input, err := readInput(fn)
	if err != nil {
		return "", err
	}
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return "", err
	}
	defer client.Close()

	secret, err := GetSecrets(ctx, client, &input, listGCPSecrets, getGCPSecretValue)
	if err != nil {
		return "", err
	}
	output, err := yaml.Marshal(secret)
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func readInput(fn string) (KGCPSecret, error) {
	content, err := os.ReadFile(fn)
	if err != nil {
		return KGCPSecret{}, err
	}

	input := KGCPSecret{
		TypeMeta: TypeMeta{},
		GCPObjectMeta: GCPObjectMeta{
			Annotations: make(kvMap),
		},
	}
	err = yaml.Unmarshal(content, &input)
	if err != nil {
		return KGCPSecret{}, err
	}

	if input.Name == "" {
		return KGCPSecret{}, errors.New("input must contain metadata.name value")
	}

	return input, nil
}

// GetSecrets gets the secret data out of the chosen secrets manager and creates the Kubernetes secret
// It is the entry point for testing
// All interactions with GCP are encapsulated in two functions:
// listGCPSecrets: get a list of all defined secrets in Google Secret Manager
// getGCPSecretValue: get the value for a specific secret in Google Secret Manager
func GetSecrets(ctx context.Context, client *secretmanager.Client,
	plugin *KGCPSecret, listGCPSecrets secretsGetter, getGCPSecretValue secretValueGetter) (K8SSecret, error) {
	data, err := createGCPSecretValuesGetter(plugin, listGCPSecrets)(ctx, client, plugin, getGCPSecretValue)

	if err != nil {
		return K8SSecret{}, err
	}

	annotations := make(kvMap)
	for k, v := range plugin.Annotations {
		annotations[k] = v
	}
	if !plugin.DisableNameSuffixHash {
		annotations["kustomize.config.k8s.io/needs-hash"] = "true"
	}
	if plugin.Behavior != "" {
		annotations["kustomize.config.k8s.io/behavior"] = plugin.Behavior
	}

	secret := K8SSecret{
		TypeMeta: TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: ObjectMeta{
			Name:        plugin.Name,
			Namespace:   plugin.Namespace,
			Labels:      plugin.Labels,
			Annotations: annotations,
		},
		Data: data,
		Type: plugin.Type,
	}
	return secret, nil
}

type secretValueGetter func(ctx context.Context, client *secretmanager.Client, plugin *KGCPSecret, key string) (string, error)
type secretValuesGetter func(ctx context.Context, client *secretmanager.Client, plugin *KGCPSecret, f secretValueGetter) (kvMap, error)
type secretsGetter func(string) ([]string, error)

func createGCPSecretValuesGetter(plugin *KGCPSecret, listGCPSecrets secretsGetter) secretValuesGetter {
	allSecretKeys, _ := listGCPSecrets(plugin.GCPProjectID)

	return func(ctx context.Context, client *secretmanager.Client,
		plugin *KGCPSecret, getSecretValues secretValueGetter) (secrets kvMap, err error) {
		secrets = make(map[string]string)
		for _, key := range plugin.Keys {
			value, err := getBestFittingSecretValue(ctx, client, plugin, allSecretKeys, key, getSecretValues)
			if err != nil {
				return nil, err
			}
			if plugin.DataType == "envvar" {
				envvar, err := godotenv.Unmarshal(value)
				if err != nil {
					return nil, fmt.Errorf("error unmarshalling secret %q: %w", key, err)
				}
				for k, v := range envvar {
					secrets[k] = base64.StdEncoding.EncodeToString([]byte(v))
				}
			} else {
				secrets[key] = base64.StdEncoding.EncodeToString([]byte(value))
			}
		}

		return
	}
}

func getBestFittingSecretValue(ctx context.Context, client *secretmanager.Client,
	plugin *KGCPSecret, allKeys []string, key string, getSecretValue secretValueGetter) (string, error) {
	var err = errors.New(fmt.Sprintf("key '%s' was not found", key))
	value := ""
	environment := plugin.Stage
	if plugin.Environment != "" {
		environment = plugin.Environment
	}
	tag := plugin.Dc
	if plugin.Tag != "" {
		tag = plugin.Tag
	}
	prefixes := []string{
		plugin.Namespace + "_" + plugin.Name + "_",
		plugin.Name + "_",
		plugin.Namespace + "_",
		"",
	}
	postfixes := []string{
		"_" + environment + "_" + tag,
		"_" + environment,
		"_" + tag,
		"",
	}
	for _, prefix := range prefixes {
		for _, postfix := range postfixes {
			lookupKey := prefix + key + postfix
			for _, k := range allKeys {
				if k == lookupKey {
					value, err = getSecretValue(ctx, client, plugin, lookupKey)
					if err == nil && value != "" {
						return value, nil
					}
				}
			}
		}
	}
	return "", fmt.Errorf("error getting '%s' secret in Google project '%s'. %s", key, plugin.GCPProjectID, err)
}

func listGCPSecrets(projectID string) ([]string, error) {
	secrets := []string{}

	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return []string{}, fmt.Errorf("failed to create secretmanager client: %v", err)
	}
	defer client.Close()

	req := &secretmanagerpb.ListSecretsRequest{
		Parent: "projects/" + projectID,
	}

	it := client.ListSecrets(ctx, req)
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return []string{}, fmt.Errorf("failed to list secret versions: %v", err)
		}

		name := strings.Split(resp.Name, "/")[3]
		secrets = append(secrets, name)
	}

	return secrets, nil
}

func getGCPSecretValue(ctx context.Context, client *secretmanager.Client, plugin *KGCPSecret, key string) (string, error) {
	sanitizedKeyName := sanitizeKeyName(key)
	name := fmt.Sprintf("projects/%s/secrets/%s/versions/latest", plugin.GCPProjectID, sanitizedKeyName)
	request := &secretmanagerpb.AccessSecretVersionRequest{Name: name}
	secret, err := client.AccessSecretVersion(ctx, request)
	if err != nil {
		return "", errors.Wrapf(err, "trouble retrieving secret: %s", name)
	}

	value := string(secret.GetPayload().GetData())

	return value, nil
}

func sanitizeKeyName(name string) string {
	return strings.ReplaceAll(strings.ReplaceAll(name, ".", "_"), "/", "_")
}
