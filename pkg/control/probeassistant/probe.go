package probeassistant

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/klog/v2"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func InjectPodProbe(pod *corev1.Pod, containerIdx *[]int) error {
	for _, cidx := range *containerIdx {
		if pod.Spec.Containers[cidx].LivenessProbe != nil {
			injectContainerProbe(pod.Spec.Containers[cidx].LivenessProbe, &pod.Spec.Containers[cidx], LIVENESS_CONFIGMAP_MOUNT_PATH_PREFIX)
		}
		if pod.Spec.Containers[cidx].ReadinessProbe != nil {
			injectContainerProbe(pod.Spec.Containers[cidx].LivenessProbe, &pod.Spec.Containers[cidx], READINESS_CONFIGMAP_MOUNT_PATH_PREFIX)
		}
	}
	return nil
}

func injectContainerProbe(probe *corev1.Probe, container *corev1.Container, mountPath string) error {
	var err error
	if probe.Exec != nil {
		probe.Exec, err = transferCommandActionContext(probe.Exec, mountPath)
	} else if probe.HTTPGet != nil {
		probe.Exec, err = transferHTTPActionContext(probe, container, mountPath)
		probe.HTTPGet = nil
	} else if probe.TCPSocket != nil {
		probe.Exec, err = transferTCPActionContext(probe, container, mountPath)
		probe.TCPSocket = nil
	}
	return err
}

func generateNewCommand(rawCmd string, mountPath string) string {
	raw := `
[ -e __MOUNT_PATH__/pre.sh ] && sh __MOUNT_PATH__/pre.sh;
___RAW_COMMAND__
export EXEC_RESULE=$?
[ -e __MOUNT_PATH__/after.sh ] && sh __MOUNT_PATH__/after.sh;
`
	raw = strings.Replace(raw, "__MOUNT_PATH__", mountPath, -1)
	return strings.Replace(raw, "___RAW_COMMAND__", rawCmd, 1)
}

func NewInjectionExecAction(rawCommand string) *corev1.ExecAction {
	return &corev1.ExecAction{
		Command: []string{"/bin/sh", "-c", "echo -e '" + rawCommand + "' | /bin/sh"},
	}
}

func transferCommandActionContext(cmd *corev1.ExecAction, mountPath string) (newCom *corev1.ExecAction, err error) {
	rawCommand := strings.Join(cmd.Command, " ")

	rawCommand = generateNewCommand(rawCommand, mountPath)
	rawCommand = strings.Replace(rawCommand, "'", "\\'", -1)

	return NewInjectionExecAction(rawCommand), err
}

func transferHTTPActionContext(probe *corev1.Probe, container *corev1.Container, mountPath string) (newCom *corev1.ExecAction, err error) {
	// prepare schema
	scheme := strings.ToLower(string(probe.HTTPGet.Scheme))
	host := probe.HTTPGet.Host
	if host == "" {
		host = "127.0.0.1"
	}
	port, err := extractPort(probe.HTTPGet.Port, *container)
	if err != nil {
		return nil, err
	}
	path := probe.HTTPGet.Path
	klog.V(4).Infof("HTTP-Probe Host: %v://%v, Port: %v, Path: %v", scheme, host, port, path)
	url := formatURL(scheme, host, port, path)

	// set curl options and set --max-time
	timeout := time.Duration(probe.TimeoutSeconds) * time.Second
	curlOptions := fmt.Sprintf("--max-time %d", timeout)

	for _, h := range probe.HTTPGet.HTTPHeaders {
		curlOptions = curlOptions + fmt.Sprintf("-H \"%s: %s\"", h.Name, h.Value)
	}

	rawCmd := fmt.Sprintf("curl %s %s", curlOptions, url)
	rawCmd = generateNewCommand(rawCmd, mountPath)

	return NewInjectionExecAction(rawCmd), nil
}

func transferTCPActionContext(probe *corev1.Probe, container *corev1.Container, mountPath string) (newCom *corev1.ExecAction, err error) {
	timeout := time.Duration(probe.TimeoutSeconds) * time.Second
	port, err := extractPort(probe.TCPSocket.Port, *container)
	if err != nil {
		return nil, err
	}
	host := probe.TCPSocket.Host
	if host == "" {
		host = "127.0.0.1"
	}
	klog.V(4).Infof("TCP-Probe Host: %v, Port: %v, Timeout: %v", host, port, timeout)
	var rawCmd string
	if host == "127.0.0.1" {
		rawCmd = fmt.Sprintf("netstat -tlnp |grep :%d", port)
	} else {
		rawCmd = fmt.Sprintf("nc -v %s %d", host, port)
	}

	rawCmd = generateNewCommand(rawCmd, mountPath)

	return NewInjectionExecAction(rawCmd), nil
}

// these codes are copy from: k8s.io/kubernetes/pkg/kubelet/prober/prober.go
// to help format probe's Host/Port/Schema.

func extractPort(param intstr.IntOrString, container corev1.Container) (int, error) {
	port := -1
	var err error
	switch param.Type {
	case intstr.Int:
		port = param.IntValue()
	case intstr.String:
		if port, err = findPortByName(container, param.StrVal); err != nil {
			// Last ditch effort - maybe it was an int stored as string?
			if port, err = strconv.Atoi(param.StrVal); err != nil {
				return port, err
			}
		}
	default:
		return port, fmt.Errorf("intOrString had no kind: %+v", param)
	}
	if port > 0 && port < 65536 {
		return port, nil
	}
	return port, fmt.Errorf("invalid port number: %v", port)
}

// findPortByName is a helper function to look up a port in a container by name.
func findPortByName(container corev1.Container, portName string) (int, error) {
	for _, port := range container.Ports {
		if port.Name == portName {
			return int(port.ContainerPort), nil
		}
	}
	return 0, fmt.Errorf("port %s not found", portName)
}

// formatURL formats a URL from args.  For testability.
func formatURL(scheme string, host string, port int, path string) *url.URL {
	u, err := url.Parse(path)
	// Something is busted with the path, but it's too late to reject it. Pass it along as is.
	if err != nil {
		u = &url.URL{
			Path: path,
		}
	}
	u.Scheme = scheme
	u.Host = net.JoinHostPort(host, strconv.Itoa(port))
	return u
}
