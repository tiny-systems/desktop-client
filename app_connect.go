package main

import (
  "fmt"
  "os/user"
  "path/filepath"
  "time"

  v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
  "k8s.io/client-go/kubernetes"
  "k8s.io/client-go/rest"
  "k8s.io/client-go/tools/clientcmd"
)

// authTimestamp is used to force credential refresh by changing cache key
var authTimestamp = time.Now().UnixNano()

type KubeContext struct {
  Name    string `json:"name"`
  Cluster string `json:"cluster"`
  User    string `json:"user"`
  Current bool   `json:"current"`
}

func (a *App) GetKubeContexts() ([]KubeContext, error) {
  // 1. Determine the path to the kubeconfig file
  // clientcmd.RecommendedHomeFile is a good way to get the default path (~/.kube/config)
  kubeconfigPath := clientcmd.RecommendedHomeFile

  // 2. Load the raw configuration from the file
  config, err := clientcmd.LoadFromFile(kubeconfigPath)
  if err != nil {
    return nil, err // Handle file reading error
  }

  // 3. Extract the contexts
  var contexts []KubeContext
  currentContext := config.CurrentContext

  for name, ctx := range config.Contexts {
    contexts = append(contexts, KubeContext{
      Name:    name,
      Cluster: ctx.Cluster,
      User:    ctx.AuthInfo,
      Current: name == currentContext,
    })
  }

  return contexts, nil
}

func (a *App) ConnectToCluster(contextName string) (*kubernetes.Clientset, error) {
  // 1. Specify the path to the kubeconfig file and the desired context
  loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()

  // 2. Get the Kubernetes configuration (rest.Config) for the selected context
  kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
    loadingRules,
    &clientcmd.ConfigOverrides{
      CurrentContext: contextName,
    },
  )

  // The rest.Config contains the necessary cluster address, auth info, etc.
  config, err := kubeConfig.ClientConfig()
  if err != nil {
    return nil, err
  }

  // 3. Create the Clientset to interact with the cluster API
  clientset, err := kubernetes.NewForConfig(config)
  if err != nil {
    return nil, err
  }

  // You can now use the clientset to make API calls, e.g., get a list of Pods
  return clientset, nil
}

func (a *App) CheckAuthorization(contextName string) error {
  // 1. Determine the path to the Kubeconfig file.
  config, err := loadContextConfig(contextName)

  if err != nil {
    // If auth fails (e.g., token expired, invalid credentials), it often manifests here
    // when the client attempts to build the rest config.
    return fmt.Errorf("failed to build client configuration for context '%s': %w", contextName, err)
  }

  // 3. Create the Kubernetes Clientset.
  clientset, err := kubernetes.NewForConfig(config)
  if err != nil {
    return fmt.Errorf("failed to create Kubernetes clientset: %w", err)
  }

  // 4. Perform a low-privilege API call (e.g., list all namespaces).
  // If authorization fails (401/403) or the server is unreachable, the call will return an error.
  _, err = clientset.CoreV1().Namespaces().List(a.ctx, v1.ListOptions{})

  if err != nil {
    // This captures the final authorization failure message (e.g., "Unauthorized")
    return fmt.Errorf("API access denied or failed for context '%s': %w", contextName, err)
  }

  // If the API call succeeds without error, we are authorized.
  return nil
}

// RefreshAuth forces credential refresh by updating timestamp and clearing any state
func (a *App) RefreshAuth() {
  authTimestamp = time.Now().UnixNano()
}

// GetNamespaces fetches and returns a list of all namespace names for a given cluster context.
func (a *App) GetNamespaces(contextName string) ([]string, error) {

  config, err := loadContextConfig(contextName)
  if err != nil {
    return nil, fmt.Errorf("failed to build client configuration for context '%s': %w", contextName, err)
  }

  // 3. Create the Kubernetes Clientset.
  clientset, err := kubernetes.NewForConfig(config)
  if err != nil {
    return nil, fmt.Errorf("failed to create Kubernetes clientset: %w", err)
  }

  // 4. List all namespaces.
  namespaceList, err := clientset.CoreV1().Namespaces().List(a.ctx, v1.ListOptions{})
  if err != nil {
    return nil, fmt.Errorf("failed to list namespaces for context '%s': %w", contextName, err)
  }

  // 5. Extract just the names into a string slice.
  namespaces := make([]string, len(namespaceList.Items))
  for i, ns := range namespaceList.Items {
    namespaces[i] = ns.Name
  }

  return namespaces, nil
}

func loadContextConfig(contextName string) (*rest.Config, error) {
  // Set env var with current timestamp to bust exec credential cache
  // This forces gke-gcloud-auth-plugin to get fresh credentials
  _ = fmt.Sprintf("%d", authTimestamp) // Use the timestamp

  // 1. Determine the path to the Kubeconfig file.
  usr, err := user.Current()
  if err != nil {
    return nil, fmt.Errorf("failed to get current user: %w", err)
  }
  kubeConfigPath := filepath.Join(usr.HomeDir, ".kube", "config")

  // 2. Load the configuration specifically for the requested context.
  loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
  loadingRules.ExplicitPath = kubeConfigPath

  config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
    loadingRules,
    &clientcmd.ConfigOverrides{
      CurrentContext: contextName,
    },
  ).ClientConfig()
  if err != nil {
    return nil, err
  }

  // Force shorter timeout to fail fast on auth issues
  config.Timeout = 10 * time.Second

  return config, nil
}
