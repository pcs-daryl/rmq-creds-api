package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/pcs-aa-aas/commons/pkg/api/server"
	"github.com/rabbitmq/messaging-topology-operator/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/apimachinery/pkg/api/errors"
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
			Path:        "pods",
			HTTPMethod:  http.MethodGet,
			HandlerFunc: h.getPods,
		},
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

func (k *HandlerGroup) getPods(s *server.APIServer, c *server.APICtx) (code int, obj interface{}) {
	clientset := s.SuperClientset

	pods, err := clientset.CoreV1().Pods("default").List(c.Context, metaV1.ListOptions{})

	if errors.IsNotFound(err) {
		fmt.Printf("Pods not found in default namespace\n")
	} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
		fmt.Printf("Error getting pod %v\n", statusError.ErrStatus.Message)
	} else if err != nil {
		panic(err.Error())
	} else {
		fmt.Printf("Found test pod in default namespace\n")
	}
	return http.StatusOK, pods
}

func getK8sClient() client.Client {
	//TODO handle this kubeconfig part
	kubeconfigPath := "/home/administrator/Documents/pipeline-api/conf/supervisorconf"
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		log.Fatalf("Error reading kubeconfig: %v", err)
	}

	// Create the controller-runtime client
	k8sClient, err := client.New(cfg, client.Options{Scheme: scheme.Scheme})
	if err != nil {
		log.Fatalf("Error creating Kubernetes client: %v", err)
	}
	return k8sClient
}

func GetPodsFromCluster(ctx context.Context, k8sClient client.Client ){

}

func (k *HandlerGroup) getPermissions(s *server.APIServer, c *server.APICtx) (code int, obj interface{}) {
	clientset := s.SuperClientset

	rmq_permissions := v1beta1.Permission{}
	err := clientset.AdmissionregistrationV1().RESTClient().
		Get().
		AbsPath("/apis/rabbitmq.com/v1beta1").
		Namespace("default").
		Resource("permissions").
		Name("rabbit-user1-test-permission").
		Do(c.Context).
		Into(&rmq_permissions)

	if errors.IsNotFound(err) {
		fmt.Printf("Permission not found in default namespace\n")
	} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
		fmt.Printf("Error getting Permission %v\n", statusError.ErrStatus.Message)
	} else if err != nil {
		panic(err.Error())
	} else {
		fmt.Printf("Found test Permission in default namespace\n")
	}

	return http.StatusOK, rmq_permissions.Spec
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
