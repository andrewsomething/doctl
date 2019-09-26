package integration

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

const packagePath string = "github.com/digitalocean/doctl/cmd/doctl"

var (
	suite           spec.Suite
	builtBinaryPath string
)

func TestAll(t *testing.T) {
	suite.Run(t)
}

func TestMain(m *testing.M) {
	specOptions := []spec.Option{
		spec.Report(report.Terminal{}),
		spec.Random(),
		spec.Parallel(),
	}

	suite = spec.New("acceptance", specOptions...)
	suite("account/get", testAccountGet)
	suite("account/ratelimit", testAccountRateLimit)
	suite("auth/init", testAuthInit)

	tmpDir, err := ioutil.TempDir("", "acceptance-doctl")
	if err != nil {
		panic("failed to create temp dir")
	}

	builtBinaryPath = filepath.Join(tmpDir, path.Base(packagePath))
	if runtime.GOOS == "windows" {
		builtBinaryPath += ".exe"
	}

	cmd := exec.Command("go", "build", "-o", builtBinaryPath, packagePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		panic(fmt.Sprintf("failed to build doctl: %s", output))
	}

	code := m.Run()

	err = os.RemoveAll(tmpDir)
	if err != nil {
		panic("failed to cleanup the doctl acceptance artifacts")
	}

	os.Exit(code)
}
