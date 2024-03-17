package main

import (
	"net/http"
	"os"
	"time"

	"github.com/spf13/pflag"
	"k8s.io/apiserver/pkg/server"
	"k8s.io/apiserver/pkg/server/options"
	"k8s.io/component-base/cli/globalflag"
)

type Options struct {
	// SecureServing is the options for the secure serving server.
	SecureServingOptions options.SecureServingOptions
}

type Config struct {
	// SecureServing is the config for the secure serving server.
	SecureServingInfo *server.SecureServingInfo
}

const (
	vsValidatorName = "vs-validator"
)

func NewDefaultOptions() *Options {
	o := &Options{
		SecureServingOptions: *options.NewSecureServingOptions(),
	}

	o.SecureServingOptions.BindPort = 8080
	o.SecureServingOptions.ServerCert.PairName = vsValidatorName
	return o
}

func (o *Options) ServerConfig() *Config {
	err := o.SecureServingOptions.MaybeDefaultWithSelfSignedCerts("0.0.0.0", nil, nil)
	if err != nil {
		panic(err)
	}
	config := &Config{}
	o.SecureServingOptions.ApplyTo(&config.SecureServingInfo)
	return config
}

func (o *Options) AddFlagSet(fs *pflag.FlagSet) {
	o.SecureServingOptions.AddFlags(fs)
}

func main() {
	serverOptions := NewDefaultOptions()

	fs := pflag.NewFlagSet(vsValidatorName, pflag.ExitOnError)
	globalflag.AddGlobalFlags(fs, vsValidatorName)
	serverOptions.AddFlagSet(fs)

	if err := fs.Parse(os.Args); err != nil {
		panic(err)
	}

	c := serverOptions.ServerConfig()

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.Handle("/", http.HandlerFunc(ValidatorHandler))
	stopCh := server.SetupSignalHandler()

	c.SecureServingInfo.Serve(mux, 30*time.Second, stopCh)

}

func ValidatorHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement the validator logic
}
