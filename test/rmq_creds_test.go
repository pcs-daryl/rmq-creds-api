package main_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rabbitmq/messaging-topology-operator/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"aaaas/rmq-permissions-api/pkg/api/handlers"
	"aaaas/rmq-permissions-api/pkg/api/model"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("RMQ Creds API", func() {
	ctx := context.Background()

	BeforeEach(func() {
		By("Creating some test users")
		Expect(createUser(ctx, "daryl")).To(Succeed())
		Expect(createUser(ctx, "jerry")).To(Succeed())

		By("Creating some test permissions")
		Expect(createPermission(ctx, "test-daryl-permission", "daryl", "test")).To(Succeed())
		Expect(createPermission(ctx, "test-jerry-permission", "jerry", "test")).To(Succeed())
	})

	AfterEach(func() {
		By("Deleting tests env")
		Expect(deleteAllPermissions(ctx)).To(Succeed())
		Expect(deleteAllUsers(ctx)).To(Succeed())
	})

	Context("when verifying the startup environment", func() {
		It("should check that initial resources are created", func() {
			user, err := getUser(ctx, "daryl")
			Expect(err).NotTo(HaveOccurred())
			Expect(user.ObjectMeta.Name).To(BeEquivalentTo("daryl"))
		})
	})

	Context("Checking handler functions", func() {
		It("Should check that we can list permissions", func() {
			permissionList, err := handlers.GetPermissionsFromCluster(ctx, k8sClient, namespace)

			Expect(err).NotTo(HaveOccurred())
			Expect(permissionList).To(HaveLen(2))
		})

		It("Should add permissions", func() {
			permission := model.Permission{
				User:  "daryl",
				Vhost: "newhost",
				Access: model.Access{
					Read:      "*",
					Write:     "*",
					Configure: "*",
				},
			}
			err := handlers.AddPermissionToCluster(ctx, k8sClient, namespace, permission)
			Expect(err).NotTo(HaveOccurred())

			permissionList, err := handlers.GetPermissionsFromCluster(ctx, k8sClient, namespace)
			Expect(err).NotTo(HaveOccurred())
			Expect(permissionList).To(HaveLen(3))
		})

		It("Should upsert permissions", func() {
			permission := model.Permission{
				User:  "daryl",
				Vhost: "test",
				Access: model.Access{
					Read:      ".*",
					Write:     "*.",
					Configure: "/*",
				},
			}
			err := handlers.UpsertPermission(ctx, k8sClient, namespace, permission)
			Expect(err).NotTo(HaveOccurred())

			// should not add new entry since it exists
			permissionList, err := handlers.GetPermissionsFromCluster(ctx, k8sClient, namespace)
			Expect(err).NotTo(HaveOccurred())
			Expect(permissionList).To(HaveLen(2))

			updatedPermission, err := handlers.GetPermissionFromCluster(ctx, k8sClient, namespace, permission)
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedPermission.Spec.Permissions.Read).To(BeEquivalentTo(".*"))
			Expect(updatedPermission.Spec.Permissions.Write).To(BeEquivalentTo("*."))
			Expect(updatedPermission.Spec.Permissions.Configure).To(BeEquivalentTo("/*"))

			permission = model.Permission{
				User:  "daryl",
				Vhost: "test2",
				Access: model.Access{
					Read:      "*",
					Write:     "*",
					Configure: "*",
				},
			}
			err = handlers.UpsertPermission(ctx, k8sClient, namespace, permission)
			Expect(err).NotTo(HaveOccurred())

			// should add new entry
			permissionList, err = handlers.GetPermissionsFromCluster(ctx, k8sClient, namespace)
			Expect(err).NotTo(HaveOccurred())
			Expect(permissionList).To(HaveLen(3))
		})

		It("Should handle both insert and update given the same user and vhost", func() {
			permission := model.Permission{
				User:  "daryl",
				Vhost: "test",
				Access: model.Access{
					Read:      ".*",
					Write:     "*.",
					Configure: "/*",
				},
			}
			err := handlers.UpsertPermission(ctx, k8sClient, namespace, permission)
			Expect(err).NotTo(HaveOccurred())

			permission = model.Permission{
				User:  "daryl",
				Vhost: "test",
				Access: model.Access{
					Read:      "*",
					Write:     "*",
					Configure: "*",
				},
			}
			err = handlers.UpsertPermission(ctx, k8sClient, namespace, permission)
			Expect(err).NotTo(HaveOccurred())

			updatedPermission, err := handlers.GetPermissionFromCluster(ctx, k8sClient, namespace, permission)
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedPermission.Spec.Permissions.Read).To(BeEquivalentTo("*"))
			Expect(updatedPermission.Spec.Permissions.Write).To(BeEquivalentTo("*"))
			Expect(updatedPermission.Spec.Permissions.Configure).To(BeEquivalentTo("*"))
		})

		It("Should delete permissions", func() {
			err := handlers.DeletePermissionFromCluster(ctx, k8sClient, namespace, "daryl", "test")
			Expect(err).NotTo(HaveOccurred())

			permissionList, err := handlers.GetPermissionsFromCluster(ctx, k8sClient, namespace)
			Expect(err).NotTo(HaveOccurred())
			Expect(permissionList).To(HaveLen(1))
		})
	})
})

func createUser(ctx context.Context, name string) error {
	user := &v1beta1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Status: v1beta1.UserStatus{
			Username: name,
		},
	}
	return k8sClient.Create(ctx, user)
}

func createPermission(ctx context.Context, name string, user string, vhost string) error {
	permission := &v1beta1.Permission{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1beta1.PermissionSpec{
			User:  user,
			Vhost: vhost,
			Permissions: v1beta1.VhostPermissions{
				Configure: "*",
				Write:     "*",
				Read:      "*",
			},
		},
	}
	return k8sClient.Create(ctx, permission)
}

func getUser(ctx context.Context, name string) (*v1beta1.User, error) {
	user := &v1beta1.User{}
	typeNamespacedName := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
	err := k8sClient.Get(ctx, typeNamespacedName, user)
	return user, err
}

func deleteAllUsers(ctx context.Context) error {
	// Define a list to hold all user
	userList := &v1beta1.UserList{}
	// List all sequences in the given namespace
	if err := k8sClient.List(ctx, userList, client.InNamespace(namespace)); err != nil {
		return fmt.Errorf("failed to list user in namespace %s: %w", namespace, err)
	}

	// Iterate through each user and delete it
	for _, user := range userList.Items {
		if err := k8sClient.Delete(ctx, &user); err != nil {
			return fmt.Errorf("failed to delete user %s in namespace %s: %w", user.Name, namespace, err)
		}
	}
	return nil
}

func deleteAllPermissions(ctx context.Context) error {
	// Define a list to hold all user
	permissionList := &v1beta1.PermissionList{}
	// List all sequences in the given namespace
	if err := k8sClient.List(ctx, permissionList, client.InNamespace(namespace)); err != nil {
		return fmt.Errorf("failed to list permission in namespace %s: %w", namespace, err)
	}

	// Iterate through each user and delete it
	for _, permission := range permissionList.Items {
		if err := k8sClient.Delete(ctx, &permission); err != nil {
			return fmt.Errorf("failed to delete permission %s in namespace %s: %w", permission.Name, namespace, err)
		}
	}
	return nil
}
