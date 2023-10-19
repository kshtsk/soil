package deploy

import "reflect"

type Kind struct {
	Type             reflect.Type
	Name             string
	TerraformWorkDir string
	TerraformVarFile string
	TerraformRepoRef string
}

var KindMap = map[string]Kind{
	"k3d": {
		Name:             "k3d",
		Type:             reflect.TypeOf(ScalabilityDeployment{}),
		TerraformWorkDir: "terraform/main/k3d",
		TerraformVarFile: "",
		TerraformRepoRef: "",
	},
	"ssh": {
		Name:             "ssh",
		Type:             reflect.TypeOf(ScalabilityDeployment{}),
		TerraformWorkDir: "terraform/main/ssh",
		TerraformVarFile: "/terraform/examples/ssh.tfvars.json",
		TerraformRepoRef: "",
	},
	"aws": {
		Name:             "aws",
		Type:             reflect.TypeOf(ScalabilityDeployment{}),
		TerraformWorkDir: "terraform/main/aws",
		TerraformVarFile: "",
		TerraformRepoRef: "",
	},
}
