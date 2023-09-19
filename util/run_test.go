package util

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHelloWorld(t *testing.T) {
	message := "Hello World"
	output, _ := ExecOutput("echo", "-n", message)
	assert.Equal(t, message, output)
}

func TestEmptyLine(t *testing.T) {
	output, _ := ExecOutput("echo", "-n", "")
	assert.Equal(t, "", output)
}

func TestEmptyLineN(t *testing.T) {
	output, _ := ExecOutput("echo", "")
	assert.Equal(t, "\n", output)
}

func TestMultiLineScript(t *testing.T) {
	script := `echo Err1 1>&2
               sleep 0.01
               echo Out2
               sleep 0.01
               echo Out3
               sleep 0.01
               echo Err4 1>&2
               sleep 0.01
               echo Out5`
	output, _ := ShellCombined(script)
	assert.NotEmpty(t, output)
	expected := "Err1\nOut2\nOut3\nErr4\nOut5\n"
	assert.Equal(t, expected, output)
}

func TestDefer(t *testing.T) {
	for i := 0; i < 3; i++ {
		defer Shell(fmt.Sprintf("echo %d", i))
	}
	Shell("echo hello")
}

func TestHelmJson(t *testing.T) {
	values := map[string]interface{}{
		"param1": map[string]interface{}{
			"String":  "name",
			"Integer": 1,
			"Boolean": true,
		},
		"param2": map[string]interface{}{
			"Strings": []interface{}{
				"alpha", "omega",
			},
			"Integers": []interface{}{
				0, 1, 2,
			},
			"Booleans": []interface{}{
				true, false,
			},
		},
	}
	result := HelmJson(values)
	expected := `param1={"Boolean":true,"Integer":1,"String":"name"},` +
		`param2={"Booleans":[true,false],"Integers":[0,1,2],"Strings":["alpha","omega"]}`
	assert.Equal(t, expected, result)
}
func TestHelmJsonMixed(t *testing.T) {
	values := map[string]interface{}{
		"param1": "String",
		"param2": map[string]interface{}{
			"Mixed": []interface{}{
				"Zero", 0, false,
			},
		},
	}
	result := HelmJson(values)
	expected := `param1="String",param2={"Mixed":["Zero",0,false]}`
	assert.Equal(t, expected, result)
}

func TestHelmJsonEmpty(t *testing.T) {
	values := map[string]interface{}{}
	result := HelmJson(values)
	expected := ``
	assert.Equal(t, expected, result)
}

func TestHelmInstall(t *testing.T) {
	ADMIN_PASSWORD := "blablabla"
	localTesterName := "local_name"
	graphanaJson := map[string]interface{}{
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
			"hosts":   []interface{}{localTesterName},
		},
		"env": map[string]interface{}{
			"GF_SERVER_ROOT_URL":            `http://${localTesterName}/grafana`,
			"GF_SERVER_SERVE_FROM_SUB_PATH": "true",
		},
		"adminPassword": ADMIN_PASSWORD,
	}
	jsonStr := `{
        "fruits" : {
            "a": "apple",
            "b": "banana"
        },
        "colors" : {
            "r": "red",
            "g": "green",
			"b": [0, 0, 255]
        }
    }`

	var x map[string]interface{}

	json.Unmarshal([]byte(jsonStr), &x)
	fmt.Printf("%#v", x)
	/*
		graphanaCluster := map[string]any{
			"context":    "CONTEXT",
			"kubeconfig": "KUBECONFIG",
		}*/
	fmt.Printf("graphana json: %s", HelmJson(graphanaJson))
	//HelmInstall("Name", "Chart", graphanaCluster, "NameSpace", graphanaJson)
}
