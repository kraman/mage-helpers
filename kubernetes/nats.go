package kubernetes

import (
	"os"
	"fmt"
	"io/ioutil"
	"time"
	"log"
	"strings"
	
	"github.com/pkg/errors"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const natsClusterDef = `
apiVersion: "nats.io/v1alpha2"
kind: "NatsCluster"
metadata:
  name: "nats-cluster"
spec:
  size: 1
`

const natsStreamingClusterDef = `
apiVersion: "streaming.nats.io/v1alpha1"
kind: "NatsStreamingCluster"
metadata:
  name: "stan-cluster"
spec:
  size: 1
  natsSvc: "nats-cluster"
`

func checkKubeResource(t, name string) (bool, error) {
	o, err := sh.Output("kubectl", "get", t, name)
	if err != nil {
		return false, nil
	}
	return strings.Contains(o, "1/1"), nil
}

func waitForKubeResource(t, name string) (error) {
	tick := time.Tick(time.Second*1)
	timeout := time.After(time.Second*30)
	for {
		select {
		case <- tick:
			o, _ := sh.Output("kubectl", "get", t, name)
			if strings.Contains(o, "1/1") {
				return nil
			}
		case <-timeout:
			return errors.Errorf("timeout waiting for %s/%s", t,name)
		}		
	}
}

func StartNATS() error {
	mg.Deps(LoadKindConfig)

	log.Println(os.Getenv("KUBECONFIG"))

	ok, err := checkKubeResource("deployment", "nats-operator")
	if err != nil {
		return err
	}
	if ok {
		return nil
	}

	log.Println("deploying operator")
	if err := sh.Run("kubectl", "apply", "-f", "https://github.com/nats-io/nats-operator/releases/download/v0.5.0/00-prereqs.yaml"); err != nil {
		return err
	}
	if err := sh.Run("kubectl", "apply", "-f", "https://github.com/nats-io/nats-operator/releases/download/v0.5.0/10-deployment.yaml"); err != nil {
		return err
	}

	if err := waitForKubeResource("deployment", "nats-operator"); err != nil {
		return err
	}

	ok, err = checkKubeResource("pod", "nats-cluster-1")
	if err != nil {
		return err
	}
	if ok {
		return nil
	}

	f, err := ioutil.TempFile("", "")
	if err != nil {
		return errors.Wrapf(err, "unable to create temp file")
	}
	defer os.Remove(f.Name())
	fmt.Fprintln(f, natsClusterDef)
	f.Close()
	if err := sh.Run("kubectl", "apply", "-f", f.Name()); err != nil {
		return err
	}
	
	return waitForKubeResource("pod", "nats-cluster-1")
}

func StartNATSStreaming() error {
	mg.Deps(StartNATS)
	ok, err := checkKubeResource("deployment", "nats-streaming-operator")
	if err != nil {
		return err
	}
	if ok {
		return nil
	}

	log.Println("deploying operator")
	if err := sh.Run("kubectl", "apply", "-f", "https://raw.githubusercontent.com/nats-io/nats-streaming-operator/master/deploy/default-rbac.yaml"); err != nil {
		return err
	}
	if err := sh.Run("kubectl", "apply", "-f", "https://raw.githubusercontent.com/nats-io/nats-streaming-operator/master/deploy/deployment.yaml"); err != nil {
		return err
	}

	if err := waitForKubeResource("deployment", "nats-streaming-operator"); err != nil {
		return err
	}

	f, err := ioutil.TempFile("", "")
	if err != nil {
		return errors.Wrapf(err, "unable to create temp file")
	}
	defer os.Remove(f.Name())
	fmt.Fprintln(f, natsStreamingClusterDef)
	f.Close()
	if err := sh.Run("kubectl", "apply", "-f", f.Name()); err != nil {
		return err
	}
	
	return waitForKubeResource("pod", "stan-cluster-1")
}