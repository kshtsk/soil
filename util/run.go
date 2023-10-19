package util

import (
	"bufio"
	"bytes"
	"encoding/json"
	_ "errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
)

func run(args []string) (output string) {
	cmd := exec.Command(args[0], args[1:]...)
	data, _ := cmd.CombinedOutput()
	output = string(data)
	log.Printf("%s", output)
	return
}

func runScanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, data[0 : i+1], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}

func runCombined(args []string) (output string) {
	cmd := exec.Command(args[0], args[1:]...)
	out, _ := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout
	done := make(chan struct{})
	scanner := bufio.NewScanner(out)
	scanner.Split(runScanLines)
	output = ""
	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			log.Printf("%s", line)
			output = fmt.Sprintf("%s%s", output, line)
		}
		done <- struct{}{}
	}()
	cmd.Start()

	<-done

	cmd.Wait()
	return
}

type Capture struct {
	Stdout bool `default:"true"`
	Stderr bool `default:"false"`
}

type Logging struct {
	Stdout bool `default:"true"`
	Stderr bool `default:"true"`
}

type Ctx struct {
	Capture Capture
	Logging Logging
}

func NewRun() *Ctx {
	c := new(Ctx)
	c.Logging.Stdout = true
	c.Logging.Stderr = true
	c.Capture.Stdout = true
	c.Capture.Stderr = false
	return c
}

func (c Ctx) execLogging(args ...string) (string, error) {
	cmd := exec.Command(args[0], args[1:]...)
	o, _ := cmd.StdoutPipe()
	e, _ := cmd.StderrPipe()

	stdoutScanner := bufio.NewScanner(o)
	stderrScanner := bufio.NewScanner(e)

	stdoutScanner.Split(runScanLines)
	stderrScanner.Split(runScanLines)
	var wg sync.WaitGroup
	wg.Add(2)
	output := ""

	go func() {
		for stdoutScanner.Scan() {
			line := stdoutScanner.Text()
			if c.Logging.Stdout {
				log.Printf(">> %s", line)
			}
			if c.Capture.Stdout {
				output += line
			}
		}
		wg.Done()
	}()

	go func() {
		for stderrScanner.Scan() {
			line := stderrScanner.Text()
			if c.Logging.Stderr {
				log.Printf("EE %s", line)
			}
			if c.Capture.Stderr {
				output += line
			}
		}
		wg.Done()
	}()

	cmd.Start()

	wg.Wait()

	err := cmd.Wait()
	if err != nil {
		log.Printf("*** Command returns: %v", err)
	}
	return output, err
}

func Exec(args ...string) (string, error) {
	cmdStr := strings.Join(args, " ")
	log.Printf("*** Running command: %s", cmdStr)
	run := NewRun()
	run.Capture.Stdout = false
	return run.execLogging("bash", "-c", cmdStr)
}

func ExecTty(args ...string) error {
	cmdStr := strings.Join(args, " ")
	log.Printf("*** Running command: %s", cmdStr)
	cmd := exec.Command("bash", "-c", cmdStr)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Printf("*** Command returns: %v", err)
	}
	return err
}

func ExecOutput(args ...string) (string, error) {
	cmdStr := strings.Join(args, " ")
	log.Printf("Running command: %s", cmdStr)
	run := NewRun()
	return run.execLogging("bash", "-c", cmdStr)
}

func ExecUnmarshalJson(args ...string) (map[string]any, error) {
	cmdStr := strings.Join(args, " ")
	log.Printf("Running command: %s", cmdStr)
	run := NewRun()
	output, err := run.execLogging("bash", "-c", cmdStr)
	var result map[string]any
	json.Unmarshal([]byte(output), &result)
	return result, err
}

func ExecQuietUnmarshalJson(args ...string) (map[string]any, error) {
	cmdStr := strings.Join(args, " ")
	log.Printf("Running command: %s", cmdStr)
	run := NewRun()
	run.Logging.Stdout = false
	output, err := run.execLogging("bash", "-c", cmdStr)
	var result map[string]any
	json.Unmarshal([]byte(output), &result)
	return result, err
}

func Shell(cmd string) (string, error) {
	run := NewRun()
	run.Capture.Stdout = false
	log.Printf("Running command: %s", cmd)
	return run.execLogging("bash", "-c", cmd)
}

func ShellUnmarshalJson(cmd string) (map[string]any, error) {
	log.Printf("Running command: %s", cmd)
	run := NewRun()
	output, err := run.execLogging("bash", "-c", cmd)
	var result map[string]any
	json.Unmarshal([]byte(output), &result)
	return result, err
}

func ShellQuietUnmarshalJson(cmd string) (map[string]any, error) {
	run := NewRun()
	run.Logging.Stderr = false
	run.Logging.Stdout = false
	output, err := run.execLogging("bash", "-c", cmd)
	var result map[string]any
	json.Unmarshal([]byte(output), &result)
	return result, err
}

func ShellCombined(cmd string) (string, error) {
	run := NewRun()
	run.Capture.Stderr = true
	log.Printf("Running command: %s", cmd)
	return run.execLogging("bash", "-c", cmd)
}

func ShellOutput(cmd string) (string, error) {
	run := NewRun()
	return run.execLogging("bash", "-c", cmd)
}

func ShellQuietOutput(cmd string) (string, error) {
	run := NewRun()
	run.Logging.Stdout = false
	return run.execLogging("bash", "-c", cmd)
}

func HelmJson(values map[string]any) string {
	var keyvalues []string
	for k, v := range values {
		j, _ := json.Marshal(v)

		keyvalues = append(keyvalues, fmt.Sprintf("%s=%s", k, j))
	}
	return strings.Join(keyvalues, ",")
}
func HelmInstall(name string, chart string, cluster map[string]any, namespace string, values map[string]any) {
	//log.Printf("json %#v", values)
	args := []string{
		"helm",
		"--kubeconfig=" + cluster["kubeconfig"].(string),
		"--kube-context=" + cluster["context"].(string),
		"upgrade", "--install",
		"--namespace=" + namespace,
		name, chart,
		"--create-namespace",
	}
	if values != nil {
		args = append(args, "--set-json='"+HelmJson(values)+"'")
	}
	Exec(args...)
}

/**
 * Runs a k6 script from inside the specified k8s cluster via kubectl run.
 * Specify envs, tags, and a test file from the k6/ dir (which will be transferred via ConfigMaps).
 * Set record = true for result metrics to be sent to Mimir, which is expected to be running in the same cluster.
 *
 * If the test script exercises the Kubernetes API, specify a KUBECONFIG env, the corresponding file will be transferred
 * to the cluster via a Secret.
 */
func KubeCtl(cluster map[string]any, args ...string) {
	a := []string{
		"kubectl",
	}
	a = append(a, args...)
	a = append(a,
		"--kubeconfig="+cluster["kubeconfig"].(string),
		"--context="+cluster["context"].(string),
	)
	Exec(a...)
}
func KubeCtlTty(cluster map[string]any, args ...string) {
	a := []string{
		"kubectl",
	}
	a = append(a, args...)
	a = append(a,
		"--kubeconfig="+cluster["kubeconfig"].(string),
		"--context="+cluster["context"].(string),
	)
	ExecTty(a...)
}

const MIMIR_URL = "http://mimir.tester:9009/mimir"
const K6_IMAGE = "grafana/k6:0.46.0"

func K6Run(cluster map[string]any, envs map[string]string, tags map[string]string, test string, record bool, tty bool) {

	kubeconfig, ok := envs["KUBECONFIG"]
	if ok {
		KubeCtl(cluster, "--namespace=tester", "delete", "secret", "kube", "--ignore-not-found")
		KubeCtl(cluster, "--namespace=tester", "create", "secret", "generic", "kube",
			"--from-file=config="+kubeconfig)
		envs["KUBECONFIG"] = "/kube/config"
	}
	log.Printf("k6 env=%#v", envs)
	cmdArgs := []string{"k6"}
	args := []string{"run"}
	for k, v := range envs {
		args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
	}
	if tags != nil {
		for k, v := range tags {
			args = append(args, "--tag", fmt.Sprintf("%s=%s", k, v))
		}
	}
	args = append(args, test)

	containerArgs := []string{}
	containerArgs = append(containerArgs, args...)
	if record {
		containerArgs = append(containerArgs, "-o", "experimental-prometheus-rw")
	}
	log.Printf("Container args: %#v", containerArgs)
	log.Printf("***Running equivalent of:\n %s\n", strings.Join(append(cmdArgs, args...), " "))
	volumeMounts := []any{
		map[string]any{"mountPath": "/k6", "name": "k6-test-files"},
		map[string]any{"mountPath": "/k6/lib", "name": "k6-lib-files"},
	}
	volumes := []any{
		map[string]any{"name": "k6-test-files", "configMap": map[string]any{"name": "k6-test-files"}},
		map[string]any{"name": "k6-lib-files", "configMap": map[string]any{"name": "k6-lib-files"}},
	}
	if kubeconfig != "" {
		volumeMounts = append(volumeMounts, map[string]any{"mountPath": "/kube", "name": "kube"})
		volumes = append(volumes, map[string]any{"name": "kube", "secret": map[string]any{"secretName": "kube"}})
	}

	overrides := map[string]any{
		"apiVersion": "v1",
		"spec": map[string]any{
			"containers": []any{
				map[string]any{
					"name":  "k6",
					"image": K6_IMAGE,
					"stdin": true,
					"tty":   tty,
					// "run" , envArgs, tagArgs, test, outputArgs
					"args":       containerArgs,
					"workingDir": "/",
					"env": []any{
						map[string]any{"name": "K6_PROMETHEUS_RW_SERVER_URL", "value": MIMIR_URL + "/api/v1/push"},
						map[string]any{"name": "K6_PROMETHEUS_RW_TREND_AS_NATIVE_HISTOGRAM", "value": "true"},
						map[string]any{"name": "K6_PROMETHEUS_RW_STALE_MARKERS", "value": "true"},
					},
					"volumeMounts": volumeMounts,
				},
			},
			"volumes": volumes,
		},
	}
	overridesJson, _ := json.Marshal(overrides)
	KubeCtlTty(cluster, "run", "k6", "--image", K6_IMAGE, "--namespace=tester",
		"--rm",
		/*
			EE Unable to use a TTY - input is not a terminal or the right kind of file
			EE If you don't see a command prompt, try pressing enter.
			EE warning: couldn't attach to pod/k6, falling back to streaming logs:
		*/
		"-i",
		fmt.Sprintf("--tty=%v", tty),
		"--restart=Never",
		fmt.Sprintf("--overrides='%s'", string(overridesJson)))
}
