package portforward

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

var NoNeedDebugError = fmt.Errorf("NoNeedDebugError")

type DebugConfig struct {
	rwLocker    sync.RWMutex
	Kubeconfig  string
	Development bool
	// use for runtime
	hostTarget      map[string]string
	hostTargetMutex map[string]*sync.Mutex
}

var GlobalDebugConfig = DebugConfig{}

func InitGlobalDebugConfig(conf *DebugConfig) {
	GlobalDebugConfig.Development = conf.Development // only set here
	GlobalDebugConfig.Kubeconfig = conf.Kubeconfig   // only set here
	GlobalDebugConfig.hostTarget = map[string]string{}
	GlobalDebugConfig.hostTargetMutex = map[string]*sync.Mutex{}
}

// PortForwarder creates a new PortForwarder using kubectl tooling
func PortForwarder(
	ctx context.Context,
	reqHost *string,
) error {
	if !GlobalDebugConfig.Development || GlobalDebugConfig.Kubeconfig == "" {
		return NoNeedDebugError
	}
	GlobalDebugConfig.rwLocker.RLock()
	if v, ok := GlobalDebugConfig.hostTarget[*reqHost]; ok {
		*reqHost = v
		GlobalDebugConfig.rwLocker.RUnlock()
		return nil
	}
	GlobalDebugConfig.rwLocker.RUnlock()

	GlobalDebugConfig.rwLocker.Lock()
	mu, ok := GlobalDebugConfig.hostTargetMutex[*reqHost]
	if !ok {
		mu = &sync.Mutex{}
		GlobalDebugConfig.hostTargetMutex[*reqHost] = mu
	}
	GlobalDebugConfig.rwLocker.Unlock()

	mu.Lock()
	defer mu.Unlock()
	host, err := portForwarder(ctx, reqHost)
	if err != nil {
		return err
	}

	GlobalDebugConfig.rwLocker.Lock()
	GlobalDebugConfig.hostTarget[*reqHost] = host
	*reqHost = host
	GlobalDebugConfig.rwLocker.Unlock()
	return nil
}

func portForwarder(
	ctx context.Context,
	reqHost *string,
) (string, error) {
	hosts := strings.SplitN(*reqHost, ".", -1)
	if len(hosts) < 3 {
		return "", fmt.Errorf("don't need forward")
	}
	switch hosts[2] {
	case "svc":
		// may be svc direct
		hosts = hosts[0:3]
		cfg, err := clientcmd.BuildConfigFromFlags("", GlobalDebugConfig.Kubeconfig)
		if err != nil {
			return "", err
		}
		clientSet, err := kubernetes.NewForConfig(cfg)
		if err != nil {
			return "", err
		}

		// find endpoint
		endPoint, err := clientSet.CoreV1().Endpoints(hosts[1]).Get(ctx, hosts[0], metav1.GetOptions{})
		if err != nil {
			return "", err
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
			return "", fmt.Errorf("no port to forword")
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
			return "", err
		}
		readyChan := make(chan struct{})
		dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, &u)
		f, err := portforward.New(dialer, ports, ctx.Done(), readyChan, os.Stdout, os.Stdout)
		if err != nil {
			close(readyChan)
			return "", err
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
			return "", err
		}
		for _, p := range ps {
			return fmt.Sprintf("localhost:%d", p.Local), nil
		}
	case "pod":
	default:
		// clusterIP
	}
	return "", fmt.Errorf("empty ports")
}

func Proxy(req *http.Request) (*url.URL, error) {
	err := PortForwarder(context.Background(), &req.URL.Host)
	if err != nil {
		return http.ProxyFromEnvironment(req)
	}
	req.Host = req.URL.Host
	return req.URL, nil
}
