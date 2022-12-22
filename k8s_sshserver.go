package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"

	"golang.org/x/crypto/ssh"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

func main() {
	// Create a context to cancel the operation in case of error
	ctx := context.Background()

	// Get the authentication settings for the Kubernetes cluster
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err)
	}

	// Create a new Kubernetes client using the authentication settings
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	// Define the namespace and name of the POD you want to connect to
	namespace := "default"
	podName := "my-pod"

	// Get the POD from Kubernetes
	pod, err := client.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		log.Fatal(err)
	}

	// Get the IP address of the POD
	podIP := pod.Status.PodIP

	// Define the port for the SSH server
	port := 22

	// Create a new SSH server
	server := &ssh.Server{
		Addr: fmt.Sprintf("%s:%d", podIP, port),
		Handler: func(s ssh.Session) {
			// Create a new remote command executor
			executor := remotecommand.New(client, client.CoreV1().RESTClient(), "POST", remotecommand.StreamOptions{
				Stdin:  s,
				Stdout: s,
				Stderr: s,
			})

			// Execute the command on the POD
			err := executor.Execute(ctx, "bash", &remotecommand.Options{
				Namespace: namespace,
				PodName:   podName,
				Container: pod.Spec.Containers[0].Name,
				Stdin:     true,
				Stdout: true,
				Stderr: true,
			})
			if err != nil {
				log.Printf("Error executing command: %v", err)
			}
		},
	}

	// Start the SSH server
	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()
	log.Printf("SSH server started on port %d", port)
	if err := server.Serve(listener); err != nil {
		log.Fatal(err)
	}
}

