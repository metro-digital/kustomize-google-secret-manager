# kustomize-google-secret-manager

A Kustomize Plugin to create Kubernetes Secrets populated with values from Google Secret Manager.

Each Kubernetes secret object is represented by one object of kind `KGCPSecret`.
The `metadata.name` and `metadata.namespace` of the object will be the name and namespace of
the Kubernetes secret, with a possible suffix hash. The key names are each represented by
a secret in a Secrets Manager, see below for naming.

## Notes

* Kustomize updates all references to a secret's name in all other Kubernetes objects, even when a suffix hash is used.
* You can disable the suffix hash by setting `disableNameSuffixHash: true`, see [examples](example).
* You can set the Kubernetes secret `type` for TLS secrets and the like, see [examples](example).
* You can set the Kustomize `behavior:` to `replace`, `merge`, or `create` (default is `create`.)
* You can set the Secret data output as envvar by configuring the `dataType:` to `envvar` (default is `null` or file), see [examples](example). Make sure that your secrets stored in Google Secret Manager use envvar formatted (`KEY=VALUE`) to use this feature.

## Naming Secrets Manager Secrets

The Google Secret Manager doesn't allow for `.` and `/`, so all occurences will be replaces by `_`.

This plugin does some lookup in Secret Manager to find the right value for your secret key. It takes the key from the `KGCPSecret`, e.g. `password` and does lookups to the Secret Manager with different combinations of prefixes and postfixes.

Possible prefixes:

* `namespace`
* `name`

Possible postfixes:

* `environment` (old `stage` is still supported, but will be overwritten by this one if both exists)
* `tag` (old `dc` is still supported, but will be overwritten by this one if both exists)

So the most specific entry for key `password` in Secret Manager is `<namespace>_<name>_password_<environment>_<tag>` e.g. `bdm-ns_db-secrets_password_prod_be-gcw1`.
And the most generic one is `password`.

## Authentication to Google Secrets Manager

The plugin uses Go libraries provided by Google Cloud Platform that automatically tries various forms of authentication.

* Run [`gcloud auth application-default login`](https://cloud.google.com/sdk/gcloud/reference/auth/application-default/), follow the instructions, done, OR
* Set the `GOOGLE_APPLICATION_CREDENTIALS` environment variable to the path of a GCP Service or User Account credentials file.
* For additional options and more information, see the [library docs](https://pkg.go.dev/cloud.google.com/go@v0.53.0?tab=doc).

## Running Kustomize with the plugin

* Kustomize expects the plugin to be installed here: `$XDG_CONFIG_HOME/kustomize/plugin/metro.digital/v1/kgcpsecret/KGCPSecret`.
* On most Unix systems, `$XDG_CONFIG_HOME` is `~/.config`.
* Build and run the plugin without Docker like this:

```shell
git clone git@github.com/metro-digital/kustomize-google-secret-manager.git
cd kustomize-google-secret-manager
make build
```

## Supported architectures

* LINUX AMD64 (tested)
* MAC AMD64 (only cross compiled)
* Windows 386 (only cross compiled)

## Copyright and License

This software is Copyright by METRO Digital GmbH, 2021. Licensed under Apache Version 2.0.

This implementation is inspired by:

* kustomize-sopssecretgenerator (https://github.com/goabout/kustomize-sopssecretgenerator), copyright 2019-2020 Go About B.V. and contributors, licensed under the Apache License, Version 2.0.
* ksecrets (https://github.com/ForgeCloud/ksecrets), copyright 2020 ForgeRock, licensed under MIT License
