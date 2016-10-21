package externalip_test

import (
	"fmt"
	"net/http"
	"time"

	externalip "github.com/Mirantis/k8s-externalipcontroller/pkg"
	testutils "github.com/Mirantis/k8s-externalipcontroller/test/utils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/1.5/pkg/api/v1"
	"k8s.io/client-go/1.5/pkg/util/intstr"
)

var _ = Describe("Basic", func() {

	It("Service should be reachable using assigned external ips", func() {
		clientset, err := testutils.KubeClient()
		namespaceObj := &v1.Namespace{
			ObjectMeta: v1.ObjectMeta{
				GenerateName: "e2e-tests-ipcontroller-",
				Namespace:    "",
			},
			Status: v1.NamespaceStatus{},
		}
		var ns *v1.Namespace
		Eventually(func() error {
			var err error
			ns, err = clientset.Namespaces().Create(namespaceObj)
			if err != nil {
				return err
			}
			return nil
		}, 5*time.Second, 1*time.Second).Should(BeNil())

		Expect(err).NotTo(HaveOccurred())

		By("deploying externalipcontroller pod")
		externalipcontroller := newPod(
			"externalipcontroller", "externalipcontroller", "mirantis/k8s-externalipcontroller",
			[]string{"sh", "-c", "./ipcontroller", "--alsologtostderr=true", "-v=4", "-iface=eth0", "-mask=24"}, nil, true, true)
		pod, err := clientset.Pods(ns.Name).Create(externalipcontroller)
		Expect(err).Should(BeNil())
		testutils.WaitForReady(clientset, pod)

		By("deploying nginx pod application and service with extnernal ips")
		nginxLabels := map[string]string{"app": "nginx"}
		nginx := newPod(
			"nginx", "nginx", "gcr.io/google_containers/nginx-slim:0.7", nil, nginxLabels, false, false)
		pod, err = clientset.Pods(ns.Name).Create(nginx)
		Expect(err).Should(BeNil())
		testutils.WaitForReady(clientset, pod)

		servicePorts := []v1.ServicePort{{Protocol: v1.ProtocolTCP, Port: 2288, TargetPort: intstr.FromInt(80)}}
		svc := newService("nginx-service", nginxLabels, servicePorts, []string{"10.108.10.3"})
		svc, err = clientset.Services(ns.Name).Create(svc)
		Expect(err).Should(BeNil())

		By("assigning ip from external ip pool to a node where test is running")
		Expect(externalip.EnsureIPAssigned("eth0", "10.108.10.4/24")).Should(BeNil())

		By("veryfiying that service is reachable using external ip")
		Eventually(func() error {
			resp, err := http.Get("http://10.108.10.3:2288/")
			if err != nil {
				return err
			}
			if resp.StatusCode > 200 {
				return fmt.Errorf("Unexpected error from nginx service: %s", resp.Status)
			}
			return nil
		}, 30*time.Second, 1*time.Second).Should(BeNil())
	})
})

func newPod(podName, containerName, imageName string, cmd []string, labels map[string]string, hostNetwork bool, privileged bool) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Name:   podName,
			Labels: labels,
		},
		Spec: v1.PodSpec{
			HostNetwork: hostNetwork,
			Containers: []v1.Container{
				{
					Name:            containerName,
					Image:           imageName,
					Args:            cmd,
					SecurityContext: &v1.SecurityContext{Privileged: &privileged},
					ImagePullPolicy: v1.PullIfNotPresent,
				},
			},
		},
	}
}

func newService(serviceName string, labels map[string]string, ports []v1.ServicePort, externalIPs []string) *v1.Service {
	return &v1.Service{
		ObjectMeta: v1.ObjectMeta{
			Name: serviceName,
		},
		Spec: v1.ServiceSpec{
			Selector:    labels,
			Type:        v1.ServiceTypeNodePort,
			Ports:       ports,
			ExternalIPs: externalIPs,
		},
	}
}