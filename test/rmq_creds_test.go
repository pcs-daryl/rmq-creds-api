package main_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rabbitmq/messaging-topology-operator/api/v1beta1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("RMQ Creds API", func(){
	ctx := context.Background()

	BeforeEach(func() {
		By("Creating some test users")
		Expect(createUser(ctx, "daryl")).To(Succeed())
	})

	Context("when verifying the startup environment", func(){
		It("should check that initial resources are created", func(){
			user , err := getUser(ctx, "daryl")
			Expect(err).NotTo(HaveOccurred())
			Expect(user.ObjectMeta.Name).To(BeEquivalentTo("daryl"))
		})
	})
})

func createUser(ctx context.Context, name string) error {
	user := &v1beta1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Namespace: namespace,
		},
		Status: v1beta1.UserStatus{
			Username: name,
		},
	}
	return k8sClient.Create(ctx, user)
}

func getUser(ctx context.Context, name string)(*v1beta1.User, error) {
	user := &v1beta1.User{}
	typeNamespacedName := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
	err := k8sClient.Get(ctx, typeNamespacedName, user)
	return user, err
}