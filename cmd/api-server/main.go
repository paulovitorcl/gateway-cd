package main

import (
	"flag"
	"log"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1"

	gatewaycdv1alpha1 "gateway-cd/pkg/api/v1alpha1"
	"gateway-cd/pkg/api"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(gatewaycdv1alpha1.AddToScheme(scheme))
	utilruntime.Must(gatewayapi.AddToScheme(scheme))
}

func main() {
	var addr string

	flag.StringVar(&addr, "addr", ":8080", "The address to bind the API server to")
	flag.Parse()

	// Set up Kubernetes client
	config := ctrl.GetConfigOrDie()

	client, err := client.New(config, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		log.Fatal("Failed to create Kubernetes client:", err)
	}

	// Create API server
	server := api.NewServer(client)

	log.Printf("Starting API server on %s", addr)
	if err := server.Run(addr); err != nil {
		log.Fatal("Failed to start API server:", err)
	}
}