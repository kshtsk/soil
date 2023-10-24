package deploy

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"soil/util"
	"strconv"
	"strings"
)

const RANCHER_VERSION = "2.7.6"
const RANCHER_CHART = "https://releases.rancher.com/server-charts/latest/rancher-2.7.6.tgz"
const RANCHER_IMAGE_TAG = "v2.7.6"
const CERT_MANAGER_CHART = "https://charts.jetstack.io/charts/cert-manager-v1.8.0.tgz"
const GRAFANA_CHART = "https://github.com/grafana/helm-charts/releases/download/grafana-6.56.5/grafana-6.56.5.tgz"

var ADMIN_PASSWORD string = "adminadminadmin"
var TerraformWorkDir string
var TerraformVarFile string
var TerraformRepoRef string
var SshIdentityFile string

type ScalabilityDeployment struct {
	CommonDeployment
	Repo             string `json:"repo_url"`
	Branch           string `json:"branch_name"`
	TerraformWorkDir string `json:"terraform_work_dir"`
	TerraformVarFile string `json:"terraform_var_file"`
	RancherReplicas  int    `json:"rancher_replicas"`
	Kind             string `json:"kind"`
}

func (d ScalabilityDeployment) saveStatus() {
	log.Printf("Saving status for: %s", reflect.TypeOf(d))
	saveStatus(&d)
}

func (d ScalabilityDeployment) CheckRequirements() (result bool) {
	/*
		ScalabilityTests needs:
		- terraform
		- kubectl
		- helm
		- git
	*/
	result = true
	gitVersion, err := util.ShellQuietOutput("git --version")
	if err != nil {
		log.Printf("Error: no git found, please install git")
		result = false
	} else {
		log.Printf("Found git version: %s", util.SplitLast(strings.TrimSpace(gitVersion), " "))
	}
	kubectlVersion, err := util.ShellQuietUnmarshalJson("kubectl version --client=true -o=json 2>/dev/null")
	if err != nil {
		log.Printf("Error: no kubectl found, please install kubectl")
		result = false
	} else {
		ver, _ := kubectlVersion["kustomizeVersion"]
		log.Printf("Found kubectl version: %s", ver)
	}

	terraformVersion, err := util.ShellQuietUnmarshalJson("terraform version -json")
	if err != nil {
		log.Printf("Error: no terraform found, please install terraform from " +
			"https://releases.hashicorp.com/terraform/")
		result = false
	} else {
		ver, _ := terraformVersion["terraform_version"]
		log.Printf("Found terraform version: %s", ver)

	}
	helmVersion, err := util.ShellQuietOutput("helm version --template='{{.Version}}'")
	if err != nil {
		log.Printf("Error: no helm found, try:\n" +
			"    curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash\n" +
			"or:\n" +
			"    curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | " +
			"HELM_INSTALL_DIR=$HOME/bin USE_SUDO=false bash")
		result = false
	} else {
		log.Printf("Found helm version: %s", helmVersion)
	}
	if d.Kind == "aws" {
		awsVersion, err := util.ShellQuietOutput("aws --version")
		if err != nil {
			log.Printf("Error: no aws cli found, please install aws, for details refer: " +
				"https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html")
			result = false
		} else {
			ver := strings.Split(strings.TrimSpace(awsVersion), " ")[0]
			log.Printf("Found aws cli version: %s", ver)
		}
	}
	return result
}

func (d ScalabilityDeployment) Brief() string {
	k := d.Kind
	if k == "" {
		k = "undefined"
	}
	return fmt.Sprintf("%s (%s)", d.Name, k)
}

func (d ScalabilityDeployment) String() string {
	banner := fmt.Sprintf("'%s' (%s):\n", d.Name, d.Kind) +
		fmt.Sprintf("  scalability-tests:\n") +
		fmt.Sprintf("    repo: %s\n", d.Repo) +
		fmt.Sprintf("    dir: %s\n", d.TerraformWorkDir) +
		fmt.Sprintf("    rancher replicas: %v\n", d.RancherReplicas)
	details := d.TextAccessDetails()
	if details == "" {
		return banner
	}
	return banner + "    " + strings.Replace(details, "\n", "\n    ", -1)
}

func (d ScalabilityDeployment) Extra() map[string]interface{} {
	extra := map[string]interface{}{
		"repo":     d.Repo,
		"branch":   d.Branch,
		"workdir":  d.TerraformWorkDir,
		"varfile":  d.TerraformVarFile,
		"replicas": d.RancherReplicas,
	}
	return extra
}

func (d ScalabilityDeployment) getRepoName() string {
	return util.SplitLast(d.Repo, "/")
}

func (d ScalabilityDeployment) getRepoLocalPath() (path string) {
	name := d.getRepoName()
	if name != "" {
		path = d.Workdir() + "/" + d.getRepoName()
	}
	return
}

func (d ScalabilityDeployment) getRepo() {
	util.CloneGitRepo(d.Repo, d.getRepoLocalPath())
}

/**
 * Make: create workdir, get repo, and build environment
 *
 * Returns deployment workdir path.
 */
func (d ScalabilityDeployment) Make() (path string) {
	path = d.makeWorkdir(&d)
	//saveStatus(&d)
	d.saveStatus()

	d.getRepo()
	// run terraform
	d.Run()
	// run install
	return
}

func (d ScalabilityDeployment) DName() string {
	return d.CommonDeployment.Name
}

func (d ScalabilityDeployment) StatusFile() string {
	return d.CommonDeployment.StatusFile()
}

func (d ScalabilityDeployment) Remove(force bool) {
	log.Printf("Removing deployment %s", d.DName())
	tfWorkdir := d.getRepoLocalPath() + "/" + d.TerraformWorkDir
	tfState := d.getTerraformStatePath()
	util.Exec("terraform", "-chdir="+tfWorkdir, "destroy", "-auto-approve", "-state="+tfState)
	if force {
		os.RemoveAll(d.Workdir())
	}
}

func (d ScalabilityDeployment) TerraformVarFilePath() (path string) {
	path = ""
	if d.TerraformVarFile != "" {
		if d.TerraformVarFile[0] != '/' {
			path = d.TerraformVarFile
		} else {
			path = d.getRepoLocalPath() + "/" + d.TerraformVarFile
		}
	}
	return
}

// {scalability-tests}/terraform/main/ssh
func (d ScalabilityDeployment) getTerraformWorkDir() string {
	return d.getRepoLocalPath() + "/" + d.TerraformWorkDir
}

func (d ScalabilityDeployment) getTerraformStatePath() string {
	return d.Workdir() + "/terraform.state"
}

func (d ScalabilityDeployment) getClusters() (map[string]any, error) {
	path := d.getTerraformStatePath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Fatalf("Terraform status file does not exist: %s", path)
		return nil, err
	}

	status, err := util.ExecQuietUnmarshalJson("terraform",
		"-chdir="+d.getTerraformWorkDir(),
		"output", "-json",
		"-state="+d.getTerraformStatePath())

	if err != nil {
		return nil, err
	}
	clusters := status["clusters"].(map[string]any)["value"].(map[string]any)
	return clusters, nil
}

func (d ScalabilityDeployment) getChartsDir() string {
	return d.getRepoLocalPath() + "/charts"
}

func (d ScalabilityDeployment) Run() {
	//configName := util.SplitLast(d.TerraformWorkDir, "/")
	//tfWorkdir := d.getRepoLocalPath() + "/" + d.TerraformWorkDir
	// {scalability-tests}/terraform/examples/ssh.tfvars.json
	varFilePath := d.TerraformVarFilePath()

	initCmd := fmt.Sprintf(`terraform -chdir=%s init -upgrade`, d.getTerraformWorkDir())
	util.Shell(initCmd)
	applyCmd := []string{
		"terraform",
		"-chdir=" + d.getTerraformWorkDir(),
		"apply", "-auto-approve",
		"-state=" + d.getTerraformStatePath(),
	}
	if varFilePath != "" {
		applyCmd = append(applyCmd, "-var-file="+varFilePath)
	}
	_, err := util.Exec(applyCmd...)
	if err != nil {
		log.Panicf("%v", err)
	}
	status, _ := util.ExecUnmarshalJson("terraform",
		"-chdir="+d.getTerraformWorkDir(),
		"output", "-json",
		"-state="+d.getTerraformStatePath())
	log.Printf("Terraform Status: %#v", status)
	clusters := status["clusters"].(map[string]any)["value"].(map[string]any)
	log.Printf("Terraform Clusters: %v", clusters)
	localCharts := d.getRepoLocalPath() + "/charts"

	tester := clusters["tester"].(map[string]any)
	upstream := clusters["upstream"].(map[string]any)
	testerLocalName := tester["local_name"].(string)
	testerPrivateName := tester["private_name"].(string)
	upstreamLocalName := upstream["local_name"].(string)
	upstreamPrivateName := upstream["private_name"].(string)

	log.Printf("Terraform Tester: %#v", tester)
	log.Printf("*** Installing helm charts to tester cluster '%s' from %s", testerLocalName, localCharts)
	util.HelmInstall("mimir", localCharts+"/mimir", tester, "tester", nil)
	util.HelmInstall("k6-files", localCharts+"/k6-files", tester, "tester", nil)
	util.HelmInstall("grafana-dashboards", localCharts+"/grafana-dashboards", tester, "tester", nil)
	grafanaJson := map[string]interface{}{
		"datasources": map[string]interface{}{
			"datasources.yaml": map[string]interface{}{
				"apiVersion": 1,
				"datasources": []interface{}{
					map[string]interface{}{
						"name":      "mimir",
						"type":      "prometheus",
						"url":       "http://mimir.tester:9009/mimir/prometheus",
						"access":    "proxy",
						"isDefault": true,
					},
				},
			},
		},
		"dashboardProviders": map[string]interface{}{
			"dashboardproviders.yaml": map[string]interface{}{
				"apiVersion": 1,
				"providers": []interface{}{
					map[string]interface{}{
						"name":            "default",
						"folder":          "",
						"type":            "file",
						"disableDeletion": false,
						"editable":        true,
						"options": map[string]interface{}{
							"path": "/var/lib/grafana/dashboards/default",
						},
					},
				},
			},
		},
		"dashboardsConfigMaps": map[string]interface{}{
			"default": "grafana-dashboards"},
		"ingress": map[string]interface{}{
			"enabled": true,
			"path":    "/grafana",
			"hosts":   []interface{}{testerLocalName},
		},
		"env": map[string]interface{}{
			"GF_SERVER_ROOT_URL":            "http://" + testerLocalName + "/grafana",
			"GF_SERVER_SERVE_FROM_SUB_PATH": "true",
		},
		"adminPassword": ADMIN_PASSWORD,
	}
	util.HelmInstall("grafana", GRAFANA_CHART, tester, "tester", grafanaJson)
	log.Printf("*** Upstream Cluster")
	certmanagerJson := map[string]interface{}{"installCRDs": true}
	util.HelmInstall("cert-manager", CERT_MANAGER_CHART, upstream, "cert-manager", certmanagerJson)
	const RANCHER_BOOTSTRAP_PASSWORD = "admin"
	rancherPrivateUrl := "https://" + upstreamPrivateName
	rancherJson := map[string]interface{}{
		"bootstrapPassword": RANCHER_BOOTSTRAP_PASSWORD,
		"hostname":          upstreamPrivateName,
		"replicas":          d.RancherReplicas,
		"rancherImageTag":   RANCHER_IMAGE_TAG,
		"extraEnv": []interface{}{
			map[string]interface{}{
				"name":  "CATTLE_SERVER_URL",
				"value": rancherPrivateUrl},
			map[string]interface{}{
				"name":  "CATTLE_PROMETHEUS_METRICS",
				"value": "true"},
			map[string]interface{}{
				"name":  "CATTLE_DEV_MODE",
				"value": "true"},
		},
		"livenessProbe": map[string]interface{}{
			"initialDelaySeconds": 30,
			"periodSeconds":       3600,
		},
	}
	util.HelmInstall("rancher", RANCHER_CHART, upstream, "cattle-system", rancherJson)
	rancherIngressJson := map[string]interface{}{"san": upstreamLocalName}
	util.HelmInstall("rancher-ingress", localCharts+"/rancher-ingress", upstream, "default", rancherIngressJson)
	restrictions := map[string]any{}
	if d.Kind == "k3d" {
		restrictions = map[string]any{
			"nodeSelector": map[string]any{
				"monitoring": "true",
			},
			"tolerations": []any{
				map[string]any{
					"key":      "monitoring",
					"operator": "Exists",
					"effect":   "NoSchedule",
				},
			},
		}
	}

	d.installRancherMonitoring(upstream, restrictions, "http://"+testerPrivateName+"/mimir/api/v1/push")

	util.HelmInstall("cgroups-exporter", localCharts+"/cgroups-exporter", upstream, "cattle-monitoring-system", nil)

	upstreamKubeConfig := upstream["kubeconfig"].(string)
	upstreamContext := upstream["context"].(string)
	util.Exec("kubectl", "wait", "deployment/rancher", "--namespace", "cattle-system",
		"--for", "condition=Available=true", "--timeout=1h",
		"--kubeconfig="+upstreamKubeConfig,
		"--context="+upstreamContext,
	)

	// *** Step 3: Import downstream clusters
	rancherLocalUrl := clusterLocalUrl(upstream)
	importedClusters := map[string]any{}
	for k, v := range clusters {
		if strings.HasPrefix(k, "downstream") {
			importedClusters[k] = v
		}
	}
	importedClusterNames := []string{}
	for k, _ := range importedClusters {
		importedClusterNames = append(importedClusterNames, k)
	}

	k6Env := map[string]string{
		"BASE_URL":               rancherPrivateUrl,
		"BOOTSTRAP_PASSWORD":     RANCHER_BOOTSTRAP_PASSWORD,
		"PASSWORD":               ADMIN_PASSWORD,
		"IMPORTED_CLUSTER_NAMES": strings.Join(importedClusterNames, ","),
	}
	util.K6Run(tester, k6Env, nil, "k6/rancher_setup.js", false, true)
	j := map[string]any{}
	for name, cluster := range importedClusters {
		j, _ = util.ExecUnmarshalJson("kubectl",
			"--kubeconfig="+upstreamKubeConfig,
			"--context="+upstreamContext,
			"get", "-n", "fleet-default", "cluster", name,
			"-o", "json",
		)
		log.Printf("Fleet %#v", j)
		clusterId := j["status"].(map[string]any)["clusterName"].(string)
		j, _ = util.ExecUnmarshalJson("kubectl", "get", "-n", clusterId,
			"clusterregistrationtoken.management.cattle.io", "default-token", "-o", "json",
			"--kubeconfig="+upstreamKubeConfig,
			"--context="+upstreamContext,
		)
		token := j["status"].(map[string]any)["token"].(string)
		clusterYaml := token + "_" + clusterId + ".yaml"
		clusterYamlPath := d.Workdir() + "/" + clusterYaml
		clusterYamlUrl := rancherLocalUrl + "/v3/import/" + clusterYaml
		util.Exec("curl", "--insecure", "-fL", clusterYamlUrl, "-o", clusterYamlPath)
		util.Exec("cat", clusterYamlPath)
		util.KubeCtl(cluster.(map[string]any), "apply", "-f", clusterYamlPath)
	}

	util.Exec("kubectl", "wait", "clusters.management.cattle.io", "--all",
		"--for", "condition=ready=true", "--timeout=1h",
		"--kubeconfig="+upstreamKubeConfig,
		"--context="+upstreamContext,
	)

	if len(importedClusters) > 0 {
		util.Exec("kubectl", "wait", "cluster.fleet.cattle.io", "--all", "--namespace", "fleet-default",
			"--for", "condition=ready=true", "--timeout=1h",
			"--kubeconfig="+upstreamKubeConfig,
			"--context="+upstreamContext,
		)
	}

	for _, cluster := range importedClusters {
		d.installRancherMonitoring(cluster.(map[string]any), map[string]any{}, "")
	}
}

func clusterLocalUrl(cluster map[string]any) string {
	localName := cluster["local_name"].(string)
	localPort := fmt.Sprintf("%v", cluster["local_https_port"])
	return "https://" + localName + ":" + localPort
}

func (d ScalabilityDeployment) installRancherMonitoring(cluster map[string]any, restrictions map[string]any, mimirUrl string) {
	const RANCHER_MONITORING_CHART = "https://github.com/rancher/charts/raw/release-v2.7/assets/rancher-monitoring/rancher-monitoring-102.0.0%2Bup40.1.2.tgz"
	const RANCHER_MONITORING_CRD_CHART = "https://github.com/rancher/charts/raw/release-v2.7/assets/rancher-monitoring-crd/rancher-monitoring-crd-102.0.0%2Bup40.1.2.tgz"

	rancherMonitoringCrd := map[string]any{
		"global": map[string]any{
			"cattle": map[string]any{
				"clusterId":             "local",
				"clusterName":           "local",
				"systemDefaultRegistry": "",
			},
		},
		"systemDefaultRegistry": "",
	}

	util.HelmInstall("rancher-monitoring-crd", RANCHER_MONITORING_CRD_CHART, cluster, "cattle-monitoring-system", rancherMonitoringCrd)
	remoteWrite := []any{}
	if mimirUrl != "" {
		remoteWrite = []any{
			map[string]any{
				"url": mimirUrl,
				"writeRelabelConfigs": []any{
					// drop all metrics except for the ones matching regex
					map[string]any{
						"sourceLabels": []any{"__name__"},
						"regex":        "(node_namespace_pod_container|node_cpu|node_load|node_memory|node_network_receive_bytes_total|container_network_receive_bytes_total|cgroups_).*",
						"action":       "keep",
					},
				},
			},
		}
	}
	rancherMonitoring := map[string]any{
		"alertmanager": map[string]any{"enabled": "false"},
		"grafana":      restrictions,
		"prometheus": map[string]any{
			"prometheusSpec": map[string]any{
				"evaluationInterval": "1m",
				"nodeSelector":       restrictions["nodeSelector"],
				"tolerations":        restrictions["tolerations"],
				"resources": map[string]any{
					"limits": map[string]any{
						"memory": "5000Mi",
					},
				},
				"retentionSize":  "50GiB",
				"scrapeInterval": "1m",
				// configure scraping from cgroups-exporter
				"additionalScrapeConfigs": []any{
					map[string]any{
						"job_name":     "node-cgroups-exporter",
						"honor_labels": false,
						"kubernetes_sd_configs": []any{
							map[string]any{
								"role": "node",
							},
						},
						"scheme": "http",
						"relabel_configs": []any{
							map[string]any{
								"action": "labelmap",
								"regex":  "__meta_kubernetes_node_label_(.+)",
							},
							map[string]any{
								"source_labels": []any{"__address__"},
								"action":        "replace",
								"target_label":  "__address__",
								"regex":         "([^:;]+):(\\d+)",
								"replacement":   "${1}:9753",
							},
							map[string]any{
								"source_labels": []any{"__meta_kubernetes_node_name"},
								"action":        "keep",
								"regex":         ".*",
							},
							map[string]any{
								"source_labels": []any{"__meta_kubernetes_node_name"},
								"action":        "replace",
								"target_label":  "node",
								"regex":         "(.*)",
								"replacement":   "${1}",
							},
						},
					},
				},
				// configure writing metrics to mimir
				"remoteWrite": remoteWrite,
			},
		},
		"prometheus-adapter": restrictions,
		"kube-state-metrics": restrictions,
		"prometheusOperator": restrictions,
		"global": map[string]any{
			"cattle": map[string]any{
				"clusterId":             "local",
				"clusterName":           "local",
				"systemDefaultRegistry": "",
			},
		},
		"systemDefaultRegistry": "",
	}

	util.HelmInstall("rancher-monitoring", RANCHER_MONITORING_CHART, cluster, "cattle-monitoring-system", rancherMonitoring)
}

func HelmInstall(name string, chart string, cluster map[string]any, namespace string, values map[string]any) {
	log.Printf("json %#v", values)
	util.Exec("helm",
		"--kubeconfig="+cluster["kubeconfig"].(string),
		"--kube-context="+cluster["context"].(string),
		"upgrade", "--install",
		"--namespace="+namespace,
		name, chart,
		"--create-namespace",
		"--set-json='"+util.HelmJson(values)+"'",
	)
}

func textNodeAccessCommands(cluster map[string]any) string {
	nodeAccessCommands := cluster["node_access_commands"].(map[string]any)
	text := ""
	for node, command := range nodeAccessCommands {
		text += fmt.Sprintf("    Node %s: %s\n", node, command)
	}
	return text

}
func (d ScalabilityDeployment) TextAccessDetails() string {
	clusters, _ := d.getClusters()
	tester := clusters["tester"].(map[string]any)
	upstream := clusters["upstream"].(map[string]any)

	testerLocalPort := fmt.Sprintf("%v", tester["local_http_port"])
	testerLocalName := tester["local_name"].(string)

	upstreamLocalPort := fmt.Sprintf("%v", upstream["local_https_port"])
	upstreamLocalName := upstream["local_name"].(string)
	upstreamKubeConfig := upstream["kubeconfig"].(string)
	downstreamClusters := map[string]any{}
	for k, v := range clusters {
		if strings.HasPrefix(k, "downstream") {
			downstreamClusters[k] = v
		}
	}
	text := ""
	rancherUrl := "https://" + upstreamLocalName + ":" + upstreamLocalPort + " (admin/" + ADMIN_PASSWORD + ")"
	text += "*** ACCESS DETAILS" +
		"\n*** UPSTREAM CLUSTER" +
		"\n    Rancher UI: " + rancherUrl +
		"\n    Kubernetes API:" +
		"\n      export KUBECONFIG=" + upstreamKubeConfig +
		"\n      kubectl config use-context " + upstream["context"].(string) +
		"\n" + textNodeAccessCommands(upstream)

	for name, cluster := range downstreamClusters {
		downstream := cluster.(map[string]any)
		text += "\n*** " + strings.ToUpper(name) + " CLUSTER" +
			"\n    Kubernetes API:" +
			"\n      export KUBECONFIG=" + downstream["kubeconfig"].(string) +
			"\n      kubectl config use-context " + downstream["context"].(string) +
			"\n" + textNodeAccessCommands(downstream)
	}
	grafanaUrl := "http://" + testerLocalName + ":" + testerLocalPort +
		"/grafana/d/a1508c35-b2e6-47f4-94ab-fec400d1c243/test-results?orgId=1&refresh=5s&from=now-30m&to=now" +
		" (admin/" + ADMIN_PASSWORD + ")"

	text += "\n*** TESTER CLUSTER" +
		"\n    Grafana UI: " + grafanaUrl +
		"\n" + textNodeAccessCommands(tester)
	return text
}
func (d ScalabilityDeployment) Test() {
	fmt.Printf("Running tests on the deployment: %s...\n", d.DName())
	clusters, _ := d.getClusters()
	tester := clusters["tester"].(map[string]any)
	upstream := clusters["upstream"].(map[string]any)

	upstreamPrivateName := upstream["private_name"].(string)
	// Refresh k6 files on the tester cluster
	util.HelmInstall("k6-files", d.getChartsDir()+"/k6-files", tester, "tester", nil)

	// Create config maps
	commit := util.GetRepoHead(d.getRepoLocalPath())
	log.Printf("Got git HEAD: %s", commit)
	downstreamClusters := map[string]any{}
	for k, v := range clusters {
		if strings.HasPrefix(k, "downstream") {
			downstreamClusters[k] = v
		}
	}
	const CONFIG_MAP_COUNT = 1000
	const SECRET_COUNT = 1000
	const ROLE_COUNT = 10
	const USER_COUNT = 5
	const PROJECT_COUNT = 20

	for name, cluster := range downstreamClusters {
		downstream := cluster.(map[string]any)
		privateName := downstream["private_name"].(string)
		vars := map[string]string{
			"BASE_URL":         "https://" + privateName + ":6443",
			"KUBECONFIG":       downstream["kubeconfig"].(string),
			"CONTEXT":          downstream["context"].(string),
			"CONFIG_MAP_COUNT": strconv.Itoa(CONFIG_MAP_COUNT),
			"SECRET_COUNT":     strconv.Itoa(SECRET_COUNT),
		}
		tags := map[string]string{
			"commit":     commit,
			"cluster":    name,
			"test":       "create_load.mjs",
			"ConfigMaps": strconv.Itoa(CONFIG_MAP_COUNT),
			"Secrets":    strconv.Itoa(SECRET_COUNT),
		}

		util.K6Run(tester, vars, tags, "k6/create_k8s_resources.js", true, true)
	}
	// create users and roles
	vars := map[string]string{
		"BASE_URL":   "https://" + upstreamPrivateName + ":443",
		"USERNAME":   "admin",
		"PASSWORD":   ADMIN_PASSWORD,
		"ROLE_COUNT": strconv.Itoa(ROLE_COUNT),
		"USER_COUNT": strconv.Itoa(USER_COUNT),
	}
	tags := map[string]string{
		"commit":  commit,
		"cluster": "upstream",
		"test":    "create_roles_users.mjs",
		"Roles":   strconv.Itoa(ROLE_COUNT),
		"Users":   strconv.Itoa(USER_COUNT),
	}
	util.K6Run(tester, vars, tags, "k6/create_roles_users.js", true, true)

	// create projects
	vars = map[string]string{
		"BASE_URL":      "https://" + upstreamPrivateName + ":443",
		"USERNAME":      "admin",
		"PASSWORD":      ADMIN_PASSWORD,
		"PROJECT_COUNT": strconv.Itoa(PROJECT_COUNT),
	}
	tags = map[string]string{
		"commit":   commit,
		"cluster":  "upstream",
		"test":     "create_projects.mjs",
		"Projects": strconv.Itoa(PROJECT_COUNT),
	}
	util.K6Run(tester, vars, tags, "k6/create_projects.js", true, true)
	log.Printf(d.TextAccessDetails())
}
