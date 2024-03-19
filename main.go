package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/spf13/pflag"
	networkingv1alpha3 "istio.io/api/networking/v1alpha3"
	istiov1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	admv1beta1 "k8s.io/api/admission/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apiserver/pkg/endpoints/handlers/responsewriters"
	"k8s.io/apiserver/pkg/server"
	"k8s.io/apiserver/pkg/server/options"
	"k8s.io/component-base/cli/globalflag"
)

type Options struct {
	SecureServingOptions options.SecureServingOptions
}

type Config struct {
	SecureServingInfo *server.SecureServingInfo
}

const (
	vsValdCon = "vs-vald-con"
)

var (
	scheme = runtime.NewScheme()
	codecs = serializer.NewCodecFactory(scheme)
)

func (o *Options) AddFlagSet(fs *pflag.FlagSet) {
	o.SecureServingOptions.AddFlags(fs)
}

func (o *Options) ServerConfig() *Config {
	if err := o.SecureServingOptions.MaybeDefaultWithSelfSignedCerts("0.0.0.0", nil, nil); err != nil {
		panic(err)
	}

	c := Config{}
	o.SecureServingOptions.ApplyTo(&c.SecureServingInfo)
	return &c
}

func NewDefaultOptions() *Options {
	o := &Options{
		SecureServingOptions: *options.NewSecureServingOptions(),
	}
	o.SecureServingOptions.BindPort = 8443
	o.SecureServingOptions.ServerCert.PairName = vsValdCon
	return o
}

func main() {
	options := NewDefaultOptions()

	fs := pflag.NewFlagSet(vsValdCon, pflag.ExitOnError)
	globalflag.AddGlobalFlags(fs, vsValdCon)

	options.AddFlagSet(fs)

	if err := fs.Parse(os.Args); err != nil {
		panic(err)
	}

	c := options.ServerConfig()

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.Handle("/validate", http.HandlerFunc(VirtualServiceValidator))

	stopCh := server.SetupSignalHandler()
	readyCh, stoppedCh, err := c.SecureServingInfo.Serve(mux, 30*time.Second, stopCh)
	if err != nil {
		panic(err)
	} else {
		<-readyCh
		<-stoppedCh
	}
}

func VirtualServiceValidator(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responsewriters.InternalError(w, r, err)
		fmt.Printf("Error %s, reading the body", err.Error())
	}
	defer r.Body.Close()

	gvk := admv1beta1.SchemeGroupVersion.WithKind("AdmissionReview")
	var admissionReview admv1beta1.AdmissionReview
	if _, _, err := codecs.UniversalDeserializer().Decode(body, &gvk, &admissionReview); err != nil {
		responsewriters.InternalError(w, r, err)
		fmt.Printf("Error %s, decoding the body", err.Error())
	}

	virtualService := istiov1alpha3.SchemeGroupVersion.WithKind("VirtualService")
	var vs istiov1alpha3.VirtualService
	if _, _, err := codecs.UniversalDeserializer().Decode(admissionReview.Request.Object.Raw, &virtualService, &vs); err != nil {
		responsewriters.InternalError(w, r, err)
		fmt.Printf("Error %s, decoding the VirtualService", err.Error())
	}
	vsHttp := vs.Spec.GetHttp()
	response := admv1beta1.AdmissionResponse{}
	allow := validateRoutes(vsHttp)

	if !allow {
		response = admv1beta1.AdmissionResponse{
			UID:     admissionReview.Request.UID,
			Allowed: allow,
			Result: &v1.Status{
				Message: "The provided api prefix already exists",
			},
		}
	} else {
		response = admv1beta1.AdmissionResponse{
			UID:     admissionReview.Request.UID,
			Allowed: allow,
		}
	}

	admissionReview.Response = &response
	res, err := json.Marshal(admissionReview)
	if err != nil {
		fmt.Printf("error %s, while converting response to byte slice", err.Error())
	}

	_, err = w.Write(res)
	if err != nil {
		fmt.Printf("error %s, writing respnse to responsewriter", err.Error())
	}

}

func validateRoutes(rules []*networkingv1alpha3.HTTPRoute) bool {
	fmt.Println("Validating VirtualService...")
	fmt.Println("VirtualService: ", rules)
	return true
}
