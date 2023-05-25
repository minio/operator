package portforward

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
)

var hostTarget = map[string]string{}
var hostTargetMutex = sync.Mutex{}

// NewSvcKubectlPortForwarder creates a new PortForwarder using kubectl tooling
func NewSvcKubectlPortForwarder(
	ctx context.Context,
	reqHost *string,
) error {
	//cfg, err := config.GetConfig()
	//if err != nil {
	//	return nil, err
	//}
	hostTargetMutex.Lock()
	defer hostTargetMutex.Unlock()
	hosts := strings.SplitN(*reqHost, ".", -1)
	if len(hosts) < 3 {
		return fmt.Errorf("don't need forward")
	}
	hosts = hosts[0:3]
	if v, ok := hostTarget[*reqHost]; ok {
		*reqHost = v
		return nil
	}
	cfg, err := clientcmd.BuildConfigFromFlags("", "/Users/guozhili/Library/Application Support/Lens/kubeconfigs/0bd256ad-69a5-4ef5-ab00-7e074498b607")
	if err != nil {
		return err
	}
	clientSet, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return err
	}
	// find endpoint
	endPoint, err := clientSet.CoreV1().Endpoints(hosts[1]).Get(ctx, hosts[0], metav1.GetOptions{})
	if err != nil {
		return err
	}

	var namespace, podName string
	var ports []string

	// we test one
	for _, s := range endPoint.Subsets {
		for _, addr := range s.Addresses {
			if addr.TargetRef.Kind == "Pod" {
				namespace = addr.TargetRef.Namespace
				podName = addr.TargetRef.Name
			}
		}
		for _, port := range s.Ports {
			ports = append(ports, fmt.Sprintf("%d:%d", 30000+rand.Int31n(10000), port.Port))
		}
	}

	if len(ports) == 0 {
		return fmt.Errorf("no port to forword")
	}

	req := clientSet.RESTClient().Post().
		Resource("pods").
		Namespace(namespace).
		Name(podName).
		SubResource("portforward")

	u := url.URL{
		Scheme:   req.URL().Scheme,
		Host:     req.URL().Host,
		Path:     "/api/v1" + req.URL().Path,
		RawQuery: "timeout=32s",
	}

	transport, upgrader, err := spdy.RoundTripperFor(cfg)
	if err != nil {
		return err
	}
	readyChan := make(chan struct{})
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, &u)
	f, err := portforward.New(dialer, ports, ctx.Done(), readyChan, os.Stdout, os.Stdout)
	if err != nil {
		close(readyChan)
		return err
	}
	go func() {
		err := f.ForwardPorts()
		if err != nil {
			fmt.Println(err.Error())
		}
	}()
	<-readyChan
	ps, err := f.GetPorts()
	if err != nil {
		return err
	}
	for _, p := range ps {
		addr := fmt.Sprintf("localhost:%d", p.Local)
		hostTarget[*reqHost] = addr
		*reqHost = addr
	}
	return nil
}

func Proxy(req *http.Request) (*url.URL, error) {
	err := NewSvcKubectlPortForwarder(context.Background(), &req.URL.Host)
	if err != nil {
		return http.ProxyFromEnvironment(req)
	}
	req.Host = req.URL.Host
	return req.URL, nil
}
