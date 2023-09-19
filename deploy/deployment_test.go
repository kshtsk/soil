package deploy

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func TestDeploymentMake(t *testing.T) {
	d := MakeDeployment("test_deployment", KindMap["k3d"])
	fmt.Printf("Deployment Name: %s\n", d.DName())
	p := &d
	fmt.Printf("Deployment Type: %s\n", reflect.TypeOf(d))
	fmt.Printf("Pointer Type: %s\n", reflect.TypeOf(p))
	fmt.Printf("Pointer Type: %s\n", reflect.TypeOf(*p))
}

type X struct {
	Name string `json:"name"`
	Age  string `json:"age"`
}

func TestEncodingJson(t *testing.T) {
	x := X{Name: "John", Age: "super star"}
	fmt.Printf("X=%s\n", x)

	j, err := json.Marshal(x)
	if err != nil {
		fmt.Printf("encoding error: %s\n", err)
	}
	fmt.Printf("Encoded data: %s\n", j)
	var dx any
	err = json.Unmarshal(j, &dx)
	if err != nil {
		fmt.Printf("decoding error: %s\n", err)
	}
	fmt.Printf("Unmarshalled: %s, type %s\n", dx, reflect.TypeOf(dx))
}
