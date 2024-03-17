package main

import (
	"github.com/spf13/pflag"
	"k8s.io/apiserver/pkg/server/options"
	"k8s.io/component-base/cli/globalflag"
)

type Options struct {
	// SecureServing is the options for the secure serving server.
	SecureServingOptions options.SecureServingOptions
}

const (
	vsValidatorName = "vs-validator"
)

func NewDefaultOptions() *Options {
	o := &Options{}
	o.SecureServingOptions = *options.NewSecureServingOptions()

	o.SecureServingOptions.BindPort = 8080
	o.SecureServingOptions.ServerCert.PairName = vsValidatorName
	return o
}

func (o *Options) AddFlagSet(fs *pflag.FlagSet )  {
	o.SecureServingOptions.AddFlags(fs)
}

func main() {
	serverOptions := NewDefaultOptions()

	fs := pflag.NewFlagSet(vsValidatorName, pflag.ExitOnError)
	globalflag.AddGlobalFlags(fs, vsValidatorName)
	serverOptions.AddFlagSet(fs)
	
}
