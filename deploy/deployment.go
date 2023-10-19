package deploy

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
)

const DEPLOYMENTS_DIR string = "$HOME/.soil"

type Deployment interface {
	Make() string
	Test()
	Remove(bool)
	makeWorkdir(any) string
	saveStatus()
	DName() string
	Brief() string
	StatusFile() string
	Extra() map[string]interface{}
}

type CommonDeployment struct {
	// Deployment
	Name string `json:"deployment_name"`
}

func (d CommonDeployment) Brief() string {
	return fmt.Sprintf("%s", d.Name)
}

func (d CommonDeployment) DName() string {
	return d.Name
}

func (d CommonDeployment) Workdir() string {
	return workdir(d.Name)
}

func deploymentsDir() (path string) {
	path, _ = filepath.Abs(os.ExpandEnv(DEPLOYMENTS_DIR))
	return
}
func workdir(name string) (path string) {
	path, _ = filepath.Abs(os.ExpandEnv(
		fmt.Sprintf("%s/%s", DEPLOYMENTS_DIR, name)))
	return
}

func DeploymentWorkdir(p Deployment) string {
	return workdir(p.DName())
}

type Status struct {
	Deployment     Deployment             `json:"deployment"`
	DeploymentType string                 `json:"type"`
	Name           string                 `json:"name"`
	Extra          map[string]interface{} `json:"extra""`
	// Extra parameters type dependent
	//
}

func statusFile(name string) string {
	return fmt.Sprintf("%s/status", workdir(name))
}
func (d CommonDeployment) StatusFile() string {
	return fmt.Sprintf("%s/status", d.Workdir())
}

func (d CommonDeployment) makeWorkdir(p any) (path string) {
	path = d.Workdir()
	log.Printf("Checking '%s' exists...", path)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// create path
		os.MkdirAll(path, 0755)
	} else {
		log.Printf("Working directory already exists")
	}
	return
}

func storeStatus(s Status) {
	data, _ := json.Marshal(s)
	log.Printf("Status: %s\n", string(data))
	path := s.Deployment.StatusFile()
	log.Printf("Saving status to: %s", path)
	f, _ := os.Create(path)
	f.Write(data)
	defer f.Close()
}

func (d ScalabilityDeployment) saveStatus() {
	log.Printf("Saving status for: %s", reflect.TypeOf(d))
	saveStatus(&d)
}
func (d CommonDeployment) Test() {
	log.Panicf("Test is not implemented for deployment: %s", reflect.TypeOf(d).Elem())
}
func saveStatus(d Deployment) {
	log.Printf("Saving deployment status for %s of type %s\n",
		d.DName(), reflect.TypeOf(d).Elem())
	s := Status{
		Deployment:     d,
		Name:           d.DName(),
		DeploymentType: fmt.Sprintf("%s", reflect.TypeOf(d).Elem()),
		Extra:          d.Extra(),
	}
	storeStatus(s)
}

func MakeDeployment(name string, kind Kind) Deployment {
	log.Printf("Creating deployment: %s", name)
	replicas := 3
	if kind.Name == "k3d" {
		replicas = 1
	}
	return ScalabilityDeployment{
		CommonDeployment: CommonDeployment{Name: name},
		Repo:             kind.TerraformRepoRef,
		TerraformWorkDir: kind.TerraformWorkDir,
		TerraformVarFile: kind.TerraformVarFile,
		RancherReplicas:  replicas,
		Kind:             kind.Name,
	}
}

func (c *Status) UnmarshalJSON(data []byte) error {
	m, err := UnmarshalStatus(data, map[string]reflect.Type{
		"deploy.ScalabilityDeployment": reflect.TypeOf(ScalabilityDeployment{}),
		"deploy.CommonDeployment":      reflect.TypeOf(CommonDeployment{}),
	})
	if err != nil {
		return err
	}
	c.Deployment = m["deployment"].(Deployment)
	c.DeploymentType = m["type"].(string)
	c.Name = m["name"].(string)
	c.Extra = m["extra"].(map[string]interface{})

	return nil
}

func UnmarshalStatus(data []byte, customTypes map[string]reflect.Type) (map[string]interface{}, error) {
	m := map[string]interface{}{}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}

	typeName := m["type"].(string)

	var value Deployment
	if ty, found := customTypes[typeName]; found {
		value = reflect.New(ty).Interface().(Deployment)
	}

	valueBytes, err := json.Marshal(m["deployment"])
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(valueBytes, &value); err != nil {
		return nil, err
	}

	m["deployment"] = value
	return m, nil
}

var _ json.Unmarshaler = (*Status)(nil)

func LookupDeployments() []Deployment {
	items, err := os.ReadDir(deploymentsDir())
	if err != nil {
		log.Fatal(err)
	}
	deployments := []Deployment{}
	for _, f := range items {
		if f.IsDir() {
			d, err := LookupDeployment(f.Name())
			if err == nil {
				deployments = append(deployments, d)
			} else {
				log.Printf("Deployment '%s' is not created yet", d.DName())
			}
		}
	}
	return deployments
}

func LookupDeployment(name string) (Deployment, error) {
	workdir := workdir(name)
	log.Printf("Looking up for deployment '%s' in directory '%s'", name, workdir)
	data, err := os.ReadFile(statusFile(name))
	if err != nil {
		log.Output(2, fmt.Sprintln(err))
		return nil, err
	}
	var s Status
	err = json.Unmarshal(data, &s)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	p := s.Deployment
	return p, nil
}
