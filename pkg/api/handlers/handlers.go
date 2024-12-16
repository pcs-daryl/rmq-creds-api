package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/pcs-aa-aas/commons/pkg/api/server"
	"github.com/rabbitmq/messaging-topology-operator/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubectl/pkg/scheme"
)

type HandlerGroup struct{}

func (h HandlerGroup) GroupPath() string {
	return "rmq/v1"
}

func (h HandlerGroup) HandlerManifests() []server.APIHandlerManifest {
	return []server.APIHandlerManifest{
		{
			Path:        "permissions",
			HTTPMethod:  http.MethodGet,
			HandlerFunc: h.getPermissions,
		},
		{
			Path:        "permissions",
			HTTPMethod:  http.MethodPost,
			HandlerFunc: h.addPermission,
		},
	}
}

func getK8sClient() client.Client {
	//TODO handle this kubeconfig part
	kubeconfigPath := "/mnt/c/Users/Daryl/.kube/config"
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		log.Fatalf("Error reading kubeconfig: %v", err)
	}

	err = v1beta1.AddToScheme(scheme.Scheme)
	if err != nil {
		log.Fatalf("Error Initialising scheme: %v", err)
	}

	// Create the controller-runtime client
	k8sClient, err := client.New(cfg, client.Options{Scheme: scheme.Scheme})
	if err != nil {
		log.Fatalf("Error creating Kubernetes client: %v", err)
	}
	return k8sClient
}

func (k *HandlerGroup) getPermissions(s *server.APIServer, c *server.APICtx) (code int, obj interface{}) {
	k8sClient := getK8sClient()

	rmq_permissions, err := GetPermissionsFromCluster(c, k8sClient, "default")

	if err != nil{
		return http.StatusBadRequest, err
	}

	return http.StatusOK, rmq_permissions
}

func GetPermissionsFromCluster(ctx context.Context, k8sClient client.Client, namespace string)([]v1beta1.Permission, error){
	permissions := &v1beta1.PermissionList{}
	listOptions := &client.ListOptions{
        Namespace: namespace,
    }
	err := k8sClient.List(ctx, permissions, listOptions)
	return permissions.Items, err
}

func (k *HandlerGroup) addPermission(s *server.APIServer, c *server.APICtx) (code int, obj interface{}) {
	// if err := c.BindJSON(&newAlbum); err != nil {
	// 	return http.StatusBadRequest, "add album failed"
	// }
	clientset := s.SuperClientset

	rmq_permissions := v1beta1.Permission{
		TypeMeta: metaV1.TypeMeta{
			APIVersion: "rabbitmq.com/v1beta1",
			Kind:       "Permission",
		},
		ObjectMeta: metaV1.ObjectMeta{
			Name:      "test-generated-permission",
			Namespace: "default",
		},
		Spec: v1beta1.PermissionSpec{
			Vhost: "test2",
			User:  "user1",
			Permissions: v1beta1.VhostPermissions{
				Write:     ".*",
				Configure: ".*",
				Read:      ".*",
			},
			RabbitmqClusterReference: v1beta1.RabbitmqClusterReference{
				Name: "rabbitmqcluster-sample",
			},
		},
	}

	//TODO error handling
	body, _ := json.Marshal(rmq_permissions)

	clientset.AdmissionregistrationV1().RESTClient().
		Post().
		AbsPath("/apis/rabbitmq.com/v1beta1").
		Namespace("default").
		Resource("permissions").
		Body(body).
		Do(c.Context).
		Into(&rmq_permissions)
	return http.StatusOK, rmq_permissions.Spec
}
