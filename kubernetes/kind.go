package kind

import (
	"github.com/pkg/errors"
	"os"
	"strings"
	"log"
	"os/exec"

	"github.com/magefile/mage/sh"
)

func buildKind() error {
	if _, err := exec.LookPath("kind"); err != nil {
		log.Println("Building kubesig/kind")
		if err := sh.Run("go", "get", "sigs.k8s.io/kind@v0.4.0"); err != nil {
			return errors.Wrapf(err, "unable to build kubesig/kind")
		}
	}
	return nil
}

func CreateKindCluster() error{
	clusterName := os.Getenv("CLUSTER_NAME")
	if clusterName == "" {
		clusterName = "test-cluster"
	}
	if err := buildKind(); err != nil {
		return err
	}

	list, err := sh.Output("kind", "get", "clusters")
	if err != nil {
		return errors.Wrapf(err, "unable to list kubesig/kind clusters")
	}
	clusters := strings.Split(list,"\n")
	for _, c := range clusters {
		if c == clusterName {
			return nil
		}
	}

	return errors.Wrapf(sh.RunV("kind", "create", "cluster", "--name", clusterName), "unable to create kubesig/kind cluster")
}

func DestroyKindCluster() error{
	clusterName := os.Getenv("CLUSTER_NAME")
	if clusterName == "" {
		clusterName = "test-cluster"
	}
	if err := buildKind(); err != nil {
		return err
	}

	list, err := sh.Output("kind", "get", "clusters")
	if err != nil {
		return errors.Wrapf(err, "unable to list kubesig/kind clusters")
	}
	clusters := strings.Split(list,"\n")
	for _, c := range clusters {
		if c == clusterName {
			return errors.Wrapf(sh.RunV("kind", "delete", "cluster", "--name", clusterName), "unable to delete kind cluster")
		}
	}
	return nil
}