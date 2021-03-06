package slss

import (
	"bytes"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Proxy types
const (
	ProxyProtoHTTP  = "http"
	ProxyProtoHTTPS = "https"
	ProxyProtoTCP   = "tcp"

	ngrokBinPath = "./bin/ngrok"
)

// StartNgrokProxy starts the ngrok proxy
func StartNgrokProxy(config *ngrokConfig, protoType, port string) (string, error) {
	if err := authNgrok(config.AuthToken); err != nil {
		return "", errors.WithStack(err)
	}

	return start(protoType, port)
}

func authNgrok(authToken string) error {
	cmd := exec.Command(ngrokBinPath, "authtoken", authToken)
	return cmd.Run()
}

func start(proxyType, port string) (string, error) {
	var responseMessage bytes.Buffer

	cmd := exec.Command(ngrokBinPath, proxyType, port, "-log=stdout", "--log-level=debug", "--region=ap")
	cmd.Stdout = &responseMessage

	if err := cmd.Start(); err != nil {
		return "", errors.WithStack(err)
	}

	go cmd.Wait()

	var proxyTypePrefix = proxyType + "://"
	for range time.Tick(time.Second) {
		output := responseMessage.String()
		if !strings.Contains(output, proxyTypePrefix) {
			continue
		}

		i := strings.LastIndex(output, proxyTypePrefix)

		return output[i+len(proxyTypePrefix) : i+strings.Index(output[i:], " ")], nil
	}

	return "", errors.New("unreachable")
}
