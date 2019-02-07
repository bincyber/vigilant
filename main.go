package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	api "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type WebHookRequest struct {
	Namespace api.Namespace `json:"object"`
}

type WebHookResponse struct {
	Labels      map[string]string            `json:"labels"`
	Attachments []networkingv1.NetworkPolicy `json:"attachments"`
}

func healthEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}

func syncEndpoint(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var webHookRequest WebHookRequest

		if err := json.NewDecoder(r.Body).Decode(&webHookRequest); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		namespace := webHookRequest.Namespace.ObjectMeta.Name

		if namespace == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		protocolTCP := api.ProtocolTCP
		protocolUDP := api.ProtocolUDP

		webHookResponse := WebHookResponse{
			Labels: map[string]string{
				"name": namespace,
			},
			Attachments: []networkingv1.NetworkPolicy{networkingv1.NetworkPolicy{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "networking.k8s.io/v1",
					Kind:       "NetworkPolicy",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "default-deny-all",
					Namespace: namespace,
				},
				Spec: networkingv1.NetworkPolicySpec{
					PolicyTypes: []networkingv1.PolicyType{
						networkingv1.PolicyTypeIngress,
						networkingv1.PolicyTypeEgress,
					},
					Egress: []networkingv1.NetworkPolicyEgressRule{
						networkingv1.NetworkPolicyEgressRule{
							Ports: []networkingv1.NetworkPolicyPort{
								{
									Protocol: &protocolTCP,
									Port:     &intstr.IntOrString{Type: intstr.Int, IntVal: 53},
								},
								{
									Protocol: &protocolUDP,
									Port:     &intstr.IntOrString{Type: intstr.Int, IntVal: 53},
								},
							},
						},
					},
				},
			},
			},
		}

		log.Printf("Added name label to namespace: %s\n", namespace)
		log.Printf("Added NetworkPolicy default-deny-all to namespace: %s\n", namespace)

		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(webHookResponse)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func main() {
	address := "0.0.0.0:8000"

	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/health", http.HandlerFunc(healthEndpoint))
	http.Handle("/sync", http.HandlerFunc(syncEndpoint))

	log.Printf("Starting vigilant on %s", address)

	stop := make(chan os.Signal)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	server := &http.Server{
		Addr: address,
	}

	e := make(chan error)

	go func() {
		e <- server.ListenAndServe()
	}()

	select {
	case err := <-e:
		log.Fatalf("%v\n", err)
	case <-stop:
	}

	log.Printf("Received signal, gracefully shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Failed to shutdown: %v\n", err)
	}
}
