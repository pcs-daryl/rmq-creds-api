package handlers

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"aaaas/rmq-permissions-api/pkg/api/model"

	"github.com/pcs-aa-aas/commons/pkg/api/server"
	"github.com/rabbitmq/messaging-topology-operator/api/v1beta1"
	"golang.org/x/exp/rand"
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
	k8sClient := getK8sClient()

	var permission model.Permission
	if err := c.ShouldBindJSON(&permission); err != nil {
		return http.StatusBadRequest, err
	}

	err := AddPermissionToCluster(c, k8sClient, "default", permission)
	if err != nil{
		return http.StatusBadRequest, err
	}

	return http.StatusOK, map[string]interface{}{
		"message":   "success",
	}
}


func AddPermissionToCluster(ctx context.Context, k8sClient client.Client, namespace string, permission model.Permission ) error{
	permissionName := permission.User + "-" + permission.Vhost + "-" +  generateRandomString()
	rmq_permissions := v1beta1.Permission{
		TypeMeta: metaV1.TypeMeta{
			APIVersion: "rabbitmq.com/v1beta1",
			Kind:       "Permission",
		},
		ObjectMeta: metaV1.ObjectMeta{
			Name:      permissionName,
			Namespace: namespace,
		},
		Spec: v1beta1.PermissionSpec{
			Vhost: permission.Vhost,
			User:  permission.User,
			Permissions: v1beta1.VhostPermissions{
				Write:     permission.Access.Write,
				Configure: permission.Access.Configure,
				Read:      permission.Access.Read,
			},
			RabbitmqClusterReference: v1beta1.RabbitmqClusterReference{
				//TODO see how to better handle this
				Name: "rabbitmqcluster-sample",
			},
		},
	}
	return k8sClient.Create(ctx, &rmq_permissions)
}

func generateRandomString() string {
	randomNumber := rand.Intn(999999) + 1
	// Convert the number to a string
	return strconv.Itoa(randomNumber)
}