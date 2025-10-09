package kubernetes

import (
	"fmt"
	"log"
)

func main() {
	// Example usage of the Kubernetes client
	client, err := NewKubeClient()
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	// Test connection
	if err := client.TestConnection(); err != nil {
		log.Fatalf("Failed to connect to Kubernetes: %v", err)
	}

	fmt.Println("Successfully connected to Kubernetes cluster!")

	// Get namespaces
	namespaces, err := client.GetNamespaces()
	if err != nil {
		log.Printf("Failed to get namespaces: %v", err)
	} else {
		fmt.Printf("Namespaces: %v\n", namespaces)
	}

	// Get API resources
	resources, err := client.GetAPIResources()
	if err != nil {
		log.Printf("Failed to get API resources: %v", err)
	} else {
		fmt.Printf("API Resources: %v\n", resources)
	}
}
