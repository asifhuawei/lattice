package app_runner

import (
	"fmt"

	"github.com/cloudfoundry-incubator/runtime-schema/bbs"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
)

const repUrlRelativeToExecutor string = "http://127.0.0.1:20515"

type diegoAppRunner struct {
	bbs bbs.NsyncBBS
}

func NewDiegoAppRunner(bbs bbs.NsyncBBS) *diegoAppRunner {
	return &diegoAppRunner{bbs}
}

func (appRunner *diegoAppRunner) StartDockerApp(name string, startCommand string, dockerImagePath string) error {
	err := appRunner.bbs.DesireLRP(models.DesiredLRP{
		Domain:      "diego-edge",
		ProcessGuid: name,
		Instances:   1,
		Stack:       "lucid64",
		RootFSPath:  dockerImagePath,
		Routes:      []string{fmt.Sprintf("%s.192.168.11.11.xip.io", name)},
		MemoryMB:    128,
		DiskMB:      1024,
		Ports: []models.PortMapping{
			{ContainerPort: 8080},
		},
		Log: models.LogConfig{
			Guid:       name,
			SourceName: "APP",
		},
		Actions: []models.ExecutorAction{
			models.Parallel(
				models.ExecutorAction{
					models.RunAction{
						Path: startCommand,
					},
				},
				models.ExecutorAction{
					models.MonitorAction{
						Action: models.ExecutorAction{
							models.RunAction{
								Path: "echo",
								Args: []string{"I'm a healthy little spy"},
							},
						},
						HealthyThreshold:   1,
						UnhealthyThreshold: 1,
						HealthyHook: models.HealthRequest{
							Method: "PUT",
							URL: fmt.Sprintf(
								"%s/lrp_running/%s/PLACEHOLDER_INSTANCE_INDEX/PLACEHOLDER_INSTANCE_GUID",
								repUrlRelativeToExecutor,
								name,
							),
						},
					},
				},
			),
		},
	})

	return err
}
