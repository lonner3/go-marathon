/*
Copyright 2014 Rohith All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package marathon

import (
	"errors"
	"fmt"
)

var (
	ErrApplicationExists = errors.New("The application already exists in marathon, you must update")
	/* no container has been specified yet */
	ErrNoApplicationContainer = errors.New("You have not specified a docker container yet")
)

type Applications struct {
	Apps []Application `json:"apps"`
}

type ApplicationWrap struct {
	Application Application `json:"app"`
}

type Application struct {
	ID              string            `json:"id",omitempty`
	Cmd             string            `json:"cmd,omitempty"`
	Args            []string          `json:"args,omitempty"`
	Constraints     [][]string        `json:"constraints,omitempty"`
	Container       *Container        `json:"container,omitempty"`
	CPUs            float32           `json:"cpus,omitempty"`
	Disk            float32           `json:"disk,omitempty"`
	Env             map[string]string `json:"env,omitempty"`
	Executor        string            `json:"executor,omitempty"`
	HealthChecks    []*HealthCheck    `json:"healthChecks,omitempty"`
	Instances       int               `json:"instances,omitemptys"`
	Mem             float32           `json:"mem,omitempty"`
	Tasks           []*Task           `json:"tasks,omitempty"`
	Ports           []int             `json:"ports,omitempty"`
	RequirePorts    bool              `json:"requirePorts,omitempty"`
	BackoffFactor   float32           `json:"backoffFactor,omitempty"`
	Dependencies    []string          `json:"dependencies,omitempty"`
	TasksRunning    int               `json:"tasksRunning,omitempty"`
	TasksStaged     int               `json:"tasksStaged,omitempty"`
	User            string            `json:"user,omitempty"`
	UpgradeStrategy *UpgradeStrategy  `json:"upgradeStrategy,omitempty"`
	Uris            []string          `json:"uris,omitempty"`
	Version         string            `json:"version,omitempty"`
}

type ApplicationVersions struct {
	Versions []string `json:"versions"`
}

type ApplicationVersion struct {
	Version string `json:"version"`
}

func NewDockerApplication() *Application {
	application := new(Application)
	application.Container = NewDockerContainer()
	return application
}

// The name of the application i.e. the identifier for this application
func (application *Application) Name(id string) *Application {
	application.ID = id
	return application
}

// The amount of CPU shares per instance which is assigned to the application
//		cpu:	the CPU shared (check Docker docs) per instance
func (application *Application) CPU(cpu float32) *Application {
	application.CPUs = cpu
	return application
}

// The amount of disk space the application is assigned, which for docker
// application I don't believe is relevant
//		disk:	the disk space in MB
func (application *Application) Storage(disk float32) *Application {
	application.Disk = disk
	return application
}

// Adds a dependency for this application. Note, if you want to wait for a application
// dependency to actually be UP, i.e. not just deployed, you need a health check on the
// dependant app.
//		name:	the application id which this application depends on
func (application *Application) DependsOn(name string) *Application {
	if application.Dependencies == nil {
		application.Dependencies = make([]string, 0)
	}
	application.Dependencies = append(application.Dependencies, name)
	return application

}

// The amount of memory the application can consume per instance
//		memory:	the amount of MB to assign
func (application *Application) Memory(memory float32) *Application {
	application.Mem = memory
	return application
}

// Set the number of instances of the application to run
//		count:	the number of instances to run
func (application *Application) Count(count int) *Application {
	application.Instances = count
	return application
}

// Add a argument to the applications
//		argument:	the argument you are adding
func (application *Application) Arg(argument string) *Application {
	if application.Args == nil {
		application.Args = make([]string, 0)
	}
	application.Args = append(application.Args, argument)
	return application
}

// Add a environment variable to the application
//		name:	the name of the variable
//		value:	go figure, the value associated to the above
func (application *Application) AddEnv(name, value string) *Application {
	if application.Env == nil {
		application.Env = make(map[string]string, 0)
	}
	application.Env[name] = value
	return application
}

// More of a helper method, used to check if a application has healtchecks
func (application *Application) HasHealthChecks() bool {
	if application.HealthChecks != nil && len(application.HealthChecks) > 0 {
		return true
	}
	return false
}

// Add a HTTP check to an application
//		port: 		the port the check should be checking
// 		interval:	the interval in seconds the check should be performed
func (application *Application) CheckHTTP(uri string, port, interval int) (*Application, error) {
	if application.HealthChecks == nil {
		application.HealthChecks = make([]*HealthCheck, 0)
	}
	if application.Container == nil || application.Container.Docker == nil {
		return nil, ErrNoApplicationContainer
	}
	/* step: get the port index */
	if port_index, err := application.Container.Docker.ServicePortIndex(port); err != nil {
		return nil, err
	} else {
		health := NewDefaultHealthCheck()
		health.Path = uri
		health.IntervalSeconds = interval
		health.PortIndex = port_index
		/* step: add to the checks */
		application.HealthChecks = append(application.HealthChecks, health)
		return application, nil
	}
}

// Add a TCP check to a application; note the port mapping must already exist, or an
// error will thrown
//		port: 		the port the check should, err, check
// 		interval:	the interval in seconds the check should be performed
func (application *Application) CheckTCP(port, interval int) (*Application, error) {
	if application.HealthChecks == nil {
		application.HealthChecks = make([]*HealthCheck, 0)
	}
	if application.Container == nil || application.Container.Docker == nil {
		return nil, ErrNoApplicationContainer
	}
	/* step: get the port index */
	if port_index, err := application.Container.Docker.ServicePortIndex(port); err != nil {
		return nil, err
	} else {
		health := NewDefaultHealthCheck()
		health.Protocol = "TCP"
		health.IntervalSeconds = interval
		health.PortIndex = port_index
		/* step: add to the checks */
		application.HealthChecks = append(application.HealthChecks, health)
		return application, nil
	}
}

// Retrieve an array of all the applications which are running in marathon
func (client *Client) Applications() (*Applications, error) {
	applications := new(Applications)
	if err := client.apiGet(MARATHON_API_APPS, nil, applications); err != nil {
		return nil, err
	} else {
		return applications, nil
	}
}

// Retrieve an array of the application names currently running in marathon
func (client *Client) ListApplications() ([]string, error) {
	if applications, err := client.Applications(); err != nil {
		return nil, err
	} else {
		list := make([]string, 0)
		for _, application := range applications.Apps {
			list = append(list, application.ID)
		}
		return list, nil
	}
}

// Checks to see if the application version exists in Marathon
// 		name: 		the id used to identify the application
//		version: 	the version (normally a timestamp) your looking for
func (client *Client) HasApplicationVersion(name, version string) (bool, error) {
	if versions, err := client.ApplicationVersions(name); err != nil {
		return false, err
	} else {
		if contains(versions.Versions, version) {
			return true, nil
		}
		return false, nil
	}
}

// A list of versions which has been deployed with marathon for a specific application
//		name:		the id used to identify the application
func (client *Client) ApplicationVersions(name string) (*ApplicationVersions, error) {
	uri := fmt.Sprintf("%s%s/versions", MARATHON_API_APPS, name)
	versions := new(ApplicationVersions)
	if err := client.apiGet(uri, nil, versions); err != nil {
		return nil, err
	}
	return versions, nil
}

// Change / Revert the version of the application
// 		name: 		the id used to identify the application
//		version: 	the version (normally a timestamp) you wish to change to
func (client *Client) SetApplicationVersion(name string, version *ApplicationVersion) (*DeploymentID, error) {
	client.debug("Changing the application: %s to version: %s", name, version)
	uri := fmt.Sprintf("%s%s", MARATHON_API_APPS, name)
	deploymentId := new(DeploymentID)
	if err := client.apiPut(uri, version, deploymentId); err != nil {
		client.debug("Failed to change the application to version: %s, error: %s", version.Version, err)
		return nil, err
	}
	return deploymentId, nil
}

// Retrieve the application configuration from marathon
// 		name: 		the id used to identify the application
func (client *Client) Application(name string) (*Application, error) {
	application := new(ApplicationWrap)
	if err := client.apiGet(fmt.Sprintf("%s%s", MARATHON_API_APPS, name), nil, application); err != nil {
		return nil, err
	} else {
		return &application.Application, nil
	}
}

// Validates that the application, or more appropriately it's tasks have passed all the health checks.
// If no health checks exist, we simply return true
// 		name: 		the id used to identify the application
func (client *Client) ApplicationOK(name string) (bool, error) {
	/* step: check the application even exists */
	if found, err := client.HasApplication(name); err != nil {
		return false, err
	} else if !found {
		return false, ErrDoesNotExist
	}
	/* step: get the application */
	if application, err := client.Application(name); err != nil {
		return false, err
	} else {
		/* step: if the application has not health checks, just return true */
		if application.HealthChecks == nil || len(application.HealthChecks) <= 0 {
			return true, nil
		}
		/* step: does the application have any tasks */
		if application.Tasks == nil || len(application.Tasks) <= 0 {
			return true, nil
		}

		/* step: iterate the application checks and look for false */
		for _, task := range application.Tasks {
			if task.HealthCheckResult != nil {
				for _, check := range task.HealthCheckResult {
					if !check.Alive {
						return false, nil
					}
				}
			}
		}
		return true, nil
	}
}

// Creates a new application in Marathon
// 		application: 		the structure holding the application configuration
func (client *Client) CreateApplication(application *Application) (*DeploymentID, error) {
	deployID := new(DeploymentID)
	client.debug("Creating an application: %s", application)
	if err := client.apiPost(MARATHON_API_APPS, &application, deployID); err != nil {
		return nil, err
	} else {
		return deployID, nil
	}
}

// Checks to see if the application exists in marathon
// 		name: 		the id used to identify the application
func (client *Client) HasApplication(name string) (bool, error) {
	client.debug("Checking if application: %s exists in marathon", name)
	if name == "" {
		return false, ErrInvalidArgument
	} else {
		if applications, err := client.ListApplications(); err != nil {
			return false, err
		} else {
			for _, id := range applications {
				if name == id {
					client.debug("The application: %s presently exist in maration", name)
					return true, nil
				}
			}
		}
		return false, nil
	}
}

// Deletes an application from marathon
// 		name: 		the id used to identify the application
func (client *Client) DeleteApplication(name string) (*DeploymentID, error) {
	/* step: check of the application already exists */
	client.debug("Deleting the application: %s", name)
	deployID := new(DeploymentID)
	if err := client.apiDelete(fmt.Sprintf("%s%s", MARATHON_API_APPS, name), nil, deployID); err != nil {
		return nil, err
	} else {
		return deployID, nil
	}
}

// Performs a rolling restart of marathon application
// 		name: 		the id used to identify the application
func (client *Client) RestartApplication(name string, force bool) (*DeploymentID, error) {
	client.debug("Restarting the application: %s, force: %s", name, force)
	deployment := new(DeploymentID)
	var options struct {
		Force bool `json:"force"`
		}
	options.Force = force
	if err := client.apiGet(fmt.Sprintf("%s%s/restart", MARATHON_API_APPS, name), &options, deployment); err != nil {
		return nil, err
	}
	return deployment, nil
}

// Change the number of instance an application is running
// 		name: 		the id used to identify the application
// 		instances:	the number of instances you wish to change to
func (client *Client) ScaleApplicationInstances(name string, instances int) (*DeploymentID, error) {
	client.debug("ScaleApplication: application: %s, instance: %d", name, instances)
	changes := new(Application)
	changes.ID = name
	changes.Instances = instances
	uri := fmt.Sprintf("%s%s", MARATHON_API_APPS, name)
	deployID := new(DeploymentID)
	if err := client.apiPut(uri, &changes, deployID); err != nil {
		return nil, err
	} else {
		return deployID, nil
	}
}
