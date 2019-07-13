package protobuf

import (
	"github.com/pkg/errors"
	"strings"
	"io/ioutil"
	"fmt"
	"runtime"
	"path/filepath"
	"os"
	
	"github.com/magefile/mage/target"
	"github.com/mholt/archiver"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/spf13/viper"

	"github.com/kraman/mage-helpers/net"
	_ "github.com/kraman/mage-helpers/config"
)

var (
	protocDirPath = filepath.Join(mg.CacheDir(), "protoc")
	protocBinDirPath = filepath.Join(mg.CacheDir(), "protoc", "bin")
	protocPath = filepath.Join(protocBinDirPath, "protoc")
)

func derefModule(module string) (string, error) {
	if err := sh.Run("go", "get", module); err != nil {
		return "", errors.Wrapf(err, "unable to download %s", module)
	}
	return sh.Output("go", "list", "-f", "{{ .Dir }}", "-m", module)
}

func compileProtocPlugin(module, srcPath, pluginName string) (error) {
	destPath := filepath.Join(protocBinDirPath, pluginName)
	if _, err := os.Stat(destPath); err == nil {
		return nil
	}

	modDir, err := derefModule(module)
	if err != nil {
		return err
	}
	if err := sh.Run("go", "build", "-o", destPath, filepath.Join(modDir,srcPath)); err != nil {
		return errors.Wrapf(err, "unable to build %s", pluginName)
	}
	return nil
}

func compileProtocPlugins() error {
	mg.Deps(getProtoc)

	if err := compileProtocPlugin("github.com/gogo/protobuf", "protoc-gen-gogoslick", "protoc-gen-gogoslick"); err != nil {
		return err
	}
	if err := compileProtocPlugin("github.com/grpc-ecosystem/grpc-gateway", "protoc-gen-grpc-gateway", "protoc-gen-grpc-gateway"); err != nil {
		return err
	}
	if err := compileProtocPlugin("github.com/grpc-ecosystem/grpc-gateway", "protoc-gen-swagger", "protoc-gen-swagger"); err != nil {
		return err
	}
	return nil
}

func getProtoc() (err error) {
	if _, err = os.Stat(protocPath); err == nil {
		return nil
	}

	var goos string
	switch runtime.GOOS {
	case "darwin":
		goos = "osx"
	default:
		goos = runtime.GOOS
	}

	tempDir, err := ioutil.TempDir("","")
	if err != nil {
		return errors.Wrapf(err, "unable to create temp dir")
	}
	defer os.RemoveAll(tempDir)

	protocZip := filepath.Join(tempDir, "protoc.zip")
	protocVersion := viper.GetString("protoc_version")
	if err = net.Download(fmt.Sprintf("https://github.com/protocolbuffers/protobuf/releases/download/v%s/protoc-%s-%s-%s.zip", protocVersion, protocVersion, goos, "x86_64"), protocZip); err != nil {
		return err
	}

	return archiver.Unarchive(protocZip, protocDirPath)
}

func Compile() error {
	mg.Deps(compileProtocPlugins)

	grpcGatewayModule, err := derefModule("github.com/grpc-ecosystem/grpc-gateway")
	if err != nil {
		return err
	}
	gogoProtoModule, err := derefModule("github.com/gogo/protobuf")
	if err != nil {
		return err
	}

	return filepath.Walk(".", func(srcFile string, info os.FileInfo, err error) error {
		if strings.HasPrefix(srcFile, "vendor") {
			return nil
		}
		if filepath.Ext(srcFile) == ".proto" && !info.IsDir() {
			srcDir := filepath.Dir(srcFile)

			pbFile, _ := target.Path(strings.Replace(srcFile, ".proto", ".pb.go", 1), srcFile)
			gwFile, _ := target.Path(strings.Replace(srcFile, ".proto", ".pb.gw.go", 1), srcFile)
			swaggerFile, _ := target.Path(strings.Replace(srcFile, ".proto", ".swagger.json", 1), srcFile)

			if pbFile || gwFile || swaggerFile {
				return sh.RunWith(
					map[string]string{
						"PATH": os.Getenv("PATH") + ":" + protocBinDirPath,
						"DEBUG": "1",
					}, 
					protocPath, 
					"-I", filepath.Join(grpcGatewayModule, "third_party", "googleapis"),
					"-I", filepath.Join(protocDirPath, "include"),
					"-I", gogoProtoModule,
					"-I", srcDir,
					"--gogoslick_out=plugins=grpc:" + srcDir,
					"--grpc-gateway_out=logtostderr=true:" + srcDir,
					"--swagger_out=logtostderr=true:" + srcDir,
					srcFile,
				)
			}
		}
		return nil
	})
}