package kubernetes

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// ResourceInfo holds detailed information about a Kubernetes resource
type ResourceInfo struct {
	Name      string
	Ready     string
	Status    string
	Restarts  string
	Age       string
	IP        string
	Node      string
	Namespace string
	Type      string
}

type KubeClient struct {
	clientset *kubernetes.Clientset
	config    *rest.Config
}

// NewKubeClient creates a new Kubernetes client
func NewKubeClient() (*KubeClient, error) {
	// Try to get kubeconfig from default location
	var kubeconfig string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	// Load kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %v", err)
	}

	// Create clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %v", err)
	}

	return &KubeClient{
		clientset: clientset,
		config:    config,
	}, nil
}

// GetNamespaces returns a list of all namespaces
func (k *KubeClient) GetNamespaces() ([]string, error) {
	namespaces, err := k.clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %v", err)
	}

	var nsList []string
	for _, ns := range namespaces.Items {
		nsList = append(nsList, ns.Name)
	}

	return nsList, nil
}

// GetAPIResources returns a list of available API resources
func (k *KubeClient) GetAPIResources() ([]string, error) {
	// Get server resources
	resourceList, err := k.clientset.Discovery().ServerPreferredResources()
	if err != nil {
		return nil, fmt.Errorf("failed to get server resources: %v", err)
	}

	var resources []string
	for _, group := range resourceList {
		for _, resource := range group.APIResources {
			// Skip subresources and non-namespaced resources for simplicity
			// if !resource.Namespaced || len(resource.Verbs) == 0 {
			// 	continue
			// }
			resources = append(resources, resource.Name)
		}
	}

	return resources, nil
}

// GetPodsDetailed returns detailed pod information
func (k *KubeClient) GetPodsDetailed(namespace string) ([]ResourceInfo, error) {
	pods, err := k.clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods in namespace %s: %v", namespace, err)
	}

	var podList []ResourceInfo
	for _, pod := range pods.Items {
		// Calculate ready status
		ready := "0/0"
		if len(pod.Spec.Containers) > 0 {
			readyContainers := 0
			for _, containerStatus := range pod.Status.ContainerStatuses {
				if containerStatus.Ready {
					readyContainers++
				}
			}
			ready = fmt.Sprintf("%d/%d", readyContainers, len(pod.Spec.Containers))
		}

		// Get pod status
		status := string(pod.Status.Phase)
		if pod.Status.Phase == "" {
			status = "Unknown"
		}

		// Calculate restarts
		restarts := int32(0)
		for _, containerStatus := range pod.Status.ContainerStatuses {
			restarts += containerStatus.RestartCount
		}

		// Calculate age in hours
		age := "Unknown"
		if !pod.CreationTimestamp.IsZero() {
			hours := int(time.Since(pod.CreationTimestamp.Time).Hours())
			age = fmt.Sprintf("%dh", hours)
		}

		// Get IP
		ip := pod.Status.PodIP
		if ip == "" {
			ip = "<none>"
		}

		// Get node
		node := pod.Spec.NodeName
		if node == "" {
			node = "<none>"
		}

		podList = append(podList, ResourceInfo{
			Name:      pod.Name,
			Ready:     ready,
			Status:    status,
			Restarts:  fmt.Sprintf("%d", restarts),
			Age:       age,
			IP:        ip,
			Node:      node,
			Namespace: pod.Namespace,
			Type:      "Pod",
		})
	}

	return podList, nil
}

// GetServicesDetailed returns detailed service information
func (k *KubeClient) GetServicesDetailed(namespace string) ([]ResourceInfo, error) {
	services, err := k.clientset.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list services in namespace %s: %v", namespace, err)
	}

	var serviceList []ResourceInfo
	for _, service := range services.Items {
		// Get service type
		serviceType := string(service.Spec.Type)
		if serviceType == "" {
			serviceType = "ClusterIP"
		}

		// Get cluster IP
		clusterIP := service.Spec.ClusterIP
		if clusterIP == "" {
			clusterIP = "<none>"
		}

		// Get external IP
		externalIP := "<none>"
		if len(service.Spec.ExternalIPs) > 0 {
			externalIP = service.Spec.ExternalIPs[0]
		} else if service.Spec.Type == "LoadBalancer" && len(service.Status.LoadBalancer.Ingress) > 0 {
			externalIP = service.Status.LoadBalancer.Ingress[0].IP
		}

		// Get ports
		ports := "<none>"
		if len(service.Spec.Ports) > 0 {
			port := service.Spec.Ports[0]
			ports = fmt.Sprintf("%d/%s", port.Port, port.Protocol)
		}

		// Calculate age in hours
		age := "Unknown"
		if !service.CreationTimestamp.IsZero() {
			hours := int(time.Since(service.CreationTimestamp.Time).Hours())
			age = fmt.Sprintf("%dh", hours)
		}

		serviceList = append(serviceList, ResourceInfo{
			Name:      service.Name,
			Ready:     serviceType,
			Status:    "Active",
			Restarts:  ports,
			Age:       age,
			IP:        clusterIP,
			Node:      externalIP,
			Namespace: service.Namespace,
			Type:      "Service",
		})
	}

	return serviceList, nil
}

// GetDeploymentsDetailed returns detailed deployment information
func (k *KubeClient) GetDeploymentsDetailed(namespace string) ([]ResourceInfo, error) {
	deployments, err := k.clientset.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments in namespace %s: %v", namespace, err)
	}

	var deploymentList []ResourceInfo
	for _, deployment := range deployments.Items {
		// Get ready replicas
		ready := "0/0"
		if deployment.Spec.Replicas != nil {
			ready = fmt.Sprintf("%d/%d", deployment.Status.ReadyReplicas, *deployment.Spec.Replicas)
		}

		// Get status
		status := "Unknown"
		if deployment.Status.ReadyReplicas == *deployment.Spec.Replicas {
			status = "Available"
		} else if deployment.Status.ReadyReplicas > 0 {
			status = "Progressing"
		}

		// Calculate age in hours
		age := "Unknown"
		if !deployment.CreationTimestamp.IsZero() {
			hours := int(time.Since(deployment.CreationTimestamp.Time).Hours())
			age = fmt.Sprintf("%dh", hours)
		}

		deploymentList = append(deploymentList, ResourceInfo{
			Name:      deployment.Name,
			Ready:     ready,
			Status:    status,
			Restarts:  fmt.Sprintf("%d", deployment.Status.UpdatedReplicas),
			Age:       age,
			IP:        "<none>",
			Node:      "<none>",
			Namespace: deployment.Namespace,
			Type:      "Deployment",
		})
	}

	return deploymentList, nil
}

// GetConfigMaps returns configmaps in a specific namespace
func (k *KubeClient) GetConfigMaps(namespace string) ([]string, error) {
	configmaps, err := k.clientset.CoreV1().ConfigMaps(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list configmaps in namespace %s: %v", namespace, err)
	}

	var cmList []string
	for _, cm := range configmaps.Items {
		cmList = append(cmList, cm.Name)
	}

	return cmList, nil
}

// GetSecrets returns secrets in a specific namespace
func (k *KubeClient) GetSecrets(namespace string) ([]string, error) {
	secrets, err := k.clientset.CoreV1().Secrets(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets in namespace %s: %v", namespace, err)
	}

	var secretList []string
	for _, secret := range secrets.Items {
		secretList = append(secretList, secret.Name)
	}

	return secretList, nil
}

// GetResourceList returns a list of resources for a specific type and namespace
func (k *KubeClient) GetResourceList(resourceType, namespace string) ([]ResourceInfo, error) {
	switch resourceType {
	case "pods":
		return k.GetPodsDetailed(namespace)
	case "services":
		return k.GetServicesDetailed(namespace)
	case "deployments":
		return k.GetDeploymentsDetailed(namespace)
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}
}

// GetResourceListDetailed returns detailed resource information for a specific type
func (k *KubeClient) GetResourceListDetailed(resourceType, namespace string) ([]ResourceInfo, error) {
	switch resourceType {
	case "pods":
		return k.GetPodsDetailed(namespace)
	case "services":
		return k.GetServicesDetailed(namespace)
	case "deployments":
		return k.GetDeploymentsDetailed(namespace)
	default:
		// For unsupported types, fall back to simple name list
		names, err := k.GetResourceList(resourceType, namespace)
		if err != nil {
			return nil, err
		}
		var resources []ResourceInfo
		for _, name := range names {
			resources = append(resources, ResourceInfo{
				Name:      name.Name,
				Ready:     "<none>",
				Status:    "<none>",
				Restarts:  "<none>",
				Age:       "<none>",
				IP:        "<none>",
				Node:      "<none>",
				Namespace: namespace,
				Type:      resourceType,
			})
		}
		return resources, nil
	}
}

// TestConnection tests the connection to the Kubernetes cluster
func (k *KubeClient) TestConnection() error {
	_, err := k.clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{Limit: 1})
	if err != nil {
		return fmt.Errorf("failed to connect to Kubernetes cluster: %v", err)
	}
	return nil
}

// GetPodLogs retrieves logs from a specific pod
func (k *KubeClient) GetPodLogs(namespace, podName string, tailLines int64) (string, error) {
	podLogOpts := &corev1.PodLogOptions{
		TailLines: &tailLines,
	}

	req := k.clientset.CoreV1().Pods(namespace).GetLogs(podName, podLogOpts)
	podLogs, err := req.Stream(context.TODO())
	if err != nil {
		return "", fmt.Errorf("failed to get logs for pod %s in namespace %s: %v", podName, namespace, err)
	}
	defer podLogs.Close()

	buf := make([]byte, 0, 1024*1024) // 1MB buffer
	tmp := make([]byte, 1024)
	for {
		n, err := podLogs.Read(tmp)
		if n > 0 {
			buf = append(buf, tmp[:n]...)
		}
		if err != nil {
			break
		}
	}

	return string(buf), nil
}
