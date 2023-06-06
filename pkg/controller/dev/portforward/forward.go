// Copyright (C) 2023, MinIO, Inc.
//
// This code is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License, version 3,
// as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License, version 3,
// along with this program.  If not, see <http://www.gnu.org/licenses/>

package portforward

import (
	"context"
	"fmt"
	"k8s.io/utils/env"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

// ErrorNoNeedDebug don't need debug
var ErrorNoNeedDebug = fmt.Errorf("ErrorNoNeedDebug")

// DebugConfig define debug config
type DebugConfig struct {
	rwLocker    sync.RWMutex
	Kubeconfig  string
	Development bool
	// use for runtime
	hostTarget      map[string]string
	hostTargetMutex map[string]*sync.Mutex
	clientSet       *kubernetes.Clientset
	cfg             *rest.Config
}

// GlobalDebugConfig is for global debuginfo , set it once.
var GlobalDebugConfig = DebugConfig{}

func init() {
	Kubeconfig := env.GetString("KUBECONFIG", "")
	Development := env.GetString("DEVELOPMENT", "false") == "true"
	GlobalDebugConfig.Development = Development // only set here
	GlobalDebugConfig.Kubeconfig = Kubeconfig   // only set here
	GlobalDebugConfig.hostTarget = map[string]string{}
	GlobalDebugConfig.hostTargetMutex = map[string]*sync.Mutex{}
	if Development {
		cfg, err := clientcmd.BuildConfigFromFlags("", GlobalDebugConfig.Kubeconfig)
		if err != nil {
			panic(err)
		}
		GlobalDebugConfig.cfg = cfg
		clientSet, err := kubernetes.NewForConfig(cfg)
		if err != nil {
			panic(err)
		}
		GlobalDebugConfig.clientSet = clientSet
	}
}

// PortForward creates a new PortForwarder using kubectl tooling
func PortForward(
	ctx context.Context,
	httpReq *http.Request,
) error {
	if !GlobalDebugConfig.Development || GlobalDebugConfig.Kubeconfig == "" {
		return ErrorNoNeedDebug
	}
	GlobalDebugConfig.rwLocker.RLock()
	if v, ok := GlobalDebugConfig.hostTarget[httpReq.URL.Host]; ok {
		httpReq.URL.Host = v
		GlobalDebugConfig.rwLocker.RUnlock()
		return nil
	}
	GlobalDebugConfig.rwLocker.RUnlock()

	GlobalDebugConfig.rwLocker.Lock()
	mu, ok := GlobalDebugConfig.hostTargetMutex[httpReq.URL.Host]
	if !ok {
		mu = &sync.Mutex{}
		GlobalDebugConfig.hostTargetMutex[httpReq.URL.Host] = mu
	}
	GlobalDebugConfig.rwLocker.Unlock()

	mu.Lock()
	defer mu.Unlock()
	host, err := portForward(ctx, httpReq)
	if err != nil {
		return err
	}

	GlobalDebugConfig.rwLocker.Lock()
	GlobalDebugConfig.hostTarget[httpReq.URL.Host] = host
	httpReq.URL.Host = host
	GlobalDebugConfig.rwLocker.Unlock()
	return nil
}

var podIPv4Regex = regexp.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`)

func portForward(
	ctx context.Context,
	httpReq *http.Request,
) (string, error) {
	if podIPv4Regex.MatchString(httpReq.URL.Host) {
		return "", fmt.Errorf("IPV4 %s, Don't need forward", httpReq.URL.Host)
	}
	hosts := strings.SplitN(httpReq.URL.Host, ".", -1)
	// svcName.namespace => svcName.namespace.svc
	if len(hosts) == 2 {
		hosts = append(hosts, "svc")
	}
	if len(hosts) < 3 {
		return "", fmt.Errorf("don't need forward")
	}
	switch hosts[2] {
	case "svc":
		// svcName.namespace.svc
		hosts = hosts[0:3]
		namespace := hosts[1]
		svcName := hosts[0]

		portInReqURL, err := getPortInReqURL(httpReq)
		if err != nil {
			return "", err
		}

		portInService := int32(0)
		svc, err := GlobalDebugConfig.clientSet.CoreV1().Services(namespace).Get(ctx, svcName, metav1.GetOptions{})
		if err != nil {
			return "", err
		}
		for _, p := range svc.Spec.Ports {
			if p.Port == portInReqURL {
				portInService = int32(p.TargetPort.IntValue())
				break
			}
		}
		if portInService == 0 {
			return "", fmt.Errorf("can't find any port in svc(%s)", svcName)
		}
		// find endpoint
		endPoint, err := GlobalDebugConfig.clientSet.CoreV1().Endpoints(namespace).Get(ctx, svcName, metav1.GetOptions{})
		if err != nil {
			return "", err
		}

		var podName string
		var ports []string
		// we test one
		for _, s := range endPoint.Subsets {
			for _, addr := range s.Addresses {
				if addr.TargetRef.Kind == "Pod" {
					if podName == "" {
						// we just use first
						podName = addr.TargetRef.Name
					}
				}
			}
			for _, port := range s.Ports {
				if port.Port == portInService {
					ports = append(ports, fmt.Sprintf("%d:%d", 30000+rand.Int31n(10000), port.Port))
				}
			}
		}
		return portForwardWithArgs(ctx, namespace, podName, ports, httpReq.URL.Host)
	default:
		// headless
		// podName.svcName.namespace:podPort
		hosts = hosts[0:3]
		namespace := hosts[2]
		podName := hosts[0]

		portInReqURL, err := getPortInReqURL(httpReq)
		if err != nil {
			return "", err
		}

		return portForwardWithArgs(ctx, namespace, podName, []string{strconv.Itoa(int(portInReqURL))}, httpReq.URL.Host)
	}
}

// Proxy proxy req.Host if need
func Proxy(req *http.Request) (*url.URL, error) {
	err := PortForward(context.Background(), req)
	if err != nil {
		return http.ProxyFromEnvironment(req)
	}
	req.Host = req.URL.Host
	return req.URL, nil
}

func getPortInReqURL(httpReq *http.Request) (int32, error) {
	portInReqURL := int32(80)
	if strings.Contains(httpReq.URL.Host, ":") {
		hostsIPAndPort := strings.SplitN(httpReq.URL.Host, ":", 2)
		port, err := strconv.ParseInt(hostsIPAndPort[1], 10, 64)
		if err != nil {
			return 0, fmt.Errorf("%s parse port error %w", httpReq.URL.Host, err)
		}
		portInReqURL = int32(port)
	} else if httpReq.URL.Scheme == "https" {
		portInReqURL = 443
	}
	return portInReqURL, nil
}

func portForwardWithArgs(ctx context.Context, namespace string, podName string, ports []string, host string) (string, error) {
	req := GlobalDebugConfig.clientSet.RESTClient().Post().
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

	transport, upgrader, err := spdy.RoundTripperFor(GlobalDebugConfig.cfg)
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
			fmt.Printf("forwardPorts forward %s error: %s\n", host, err)
		} else {
			fmt.Printf("forwardPorts forward %s success\n", host)
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
	return "", fmt.Errorf("forwardPorts forward %s with no ports", host)
}
