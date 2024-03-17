package main

import (
	"k8s.io/apiserver/pkg/server/options"
)

type Options struct {
	// SecureServing is the options for the secure serving server.
	SecureServingOptions options.SecureServingOptions
}
