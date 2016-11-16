package dockerclient

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	rauth "github.com/Dataman-Cloud/crane/src/plugins/registryauth"

	"github.com/docker/engine-api/types/swarm"
	"github.com/stretchr/testify/assert"
)

func TestToCraneServiceSpec(t *testing.T) {
	body := `
	{
	        "Name": "none",
	        "Id": "1836d62be355e36050913f118835bd1fd6be10638e799ccaf5ea76bc6820ced2",
	        "Scope": "local",
	        "Driver": "null",
	        "EnableIPv6": false,
	        "IPAM": {
                        "Driver": "default",
	                "Options": null,
	                "Config": []
                 },
                "Internal": false,
                "Containers": {},
                "Options": {},
	        "Labels": {}
	 }`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(body))
	}))

	defer server.Close()

	httpClient, err := NewHttpClient()
	assert.Nil(t, err)

	client := &CraneDockerClient{
		sharedHttpClient: httpClient,
	}

	craneServiceSpe := client.ToCraneServiceSpec(swarm.ServiceSpec{})
	assert.NotNil(t, craneServiceSpe)
}

func TestParseEndpoint(t *testing.T) {
	os.Setenv("CRANE_ADDR", "foobar")
	os.Setenv("CRANE_SWARM_MANAGER_IP", "foobar")
	os.Setenv("CRANE_DOCKER_CERT_PATH", "foobar")
	os.Setenv("CRANE_DB_DRIVER", "foobar")
	os.Setenv("CRANE_DB_DSN", "foobar")
	os.Setenv("CRANE_FEATURE_FLAGS", "foobar")
	os.Setenv("CRANE_REGISTRY_PRIVATE_KEY_PATH", "foobar")
	os.Setenv("CRANE_REGISTRY_ADDR", "foobar")
	os.Setenv("CRANE_ACCOUNT_AUTHENTICATOR", "foobar")
	defer os.Setenv("CRANE_ADDR", "")
	defer os.Setenv("CRANE_SWARM_MANAGER_IP", "")
	defer os.Setenv("CRANE_DOCKER_CERT_PATH", "")
	defer os.Setenv("CRANE_DB_DRIVER", "")
	defer os.Setenv("CRANE_DB_DSN", "")
	defer os.Setenv("CRANE_FEATURE_FLAGS", "")
	defer os.Setenv("CRANE_REGISTRY_PRIVATE_KEY_PATH", "")
	defer os.Setenv("CRANE_REGISTRY_ADDR", "")
	defer os.Setenv("CRANE_ACCOUNT_AUTHENTICATOR", "")

	_, err := parseEndpoint("localhost:2375")
	assert.Nil(t, err)

	_, err = parseEndpoint("tcp://localhost:2375")
	assert.Nil(t, err)

	_, err = parseEndpoint("tcp://localhost")
	assert.Nil(t, err)

	_, err = parseEndpoint("://localhost:2375")
	assert.NotNil(t, err)
}

func TestgetAdvertiseAddrByEndpoint(t *testing.T) {
	var host string
	var err error
	host, err = getAdvertiseAddrByEndpoint("localhost:2375")
	assert.Nil(t, err)
	assert.Equal(t, host, "localhost")
	host, err = getAdvertiseAddrByEndpoint("tcp://localhost:2375")
	assert.Nil(t, err)
	assert.Equal(t, host, "localhost")
	host, err = getAdvertiseAddrByEndpoint("localhost")
	assert.Nil(t, err)
	assert.Equal(t, host, "localhost")
	host, err = getAdvertiseAddrByEndpoint("http://localhost")
	assert.Nil(t, err)
	assert.Equal(t, host, "localhost")
	host, err = getAdvertiseAddrByEndpoint("://localhost")
	assert.NotNil(t, err)
}

func TestGetServicesNamespace(t *testing.T) {
	namespace := GetServicesNamespace(swarm.ServiceSpec{})
	assert.Equal(t, namespace, "")

	spec := swarm.ServiceSpec{}
	spec.Annotations.Labels = map[string]string{
		LabelNamespace: "value",
	}
	namespace = GetServicesNamespace(spec)
	assert.Equal(t, namespace, "value")
}

func TestEncodeRegistryAuth(t *testing.T) {
	authInfo := rauth.RegistryAuth{
		Username: "Username",
		Password: "Password",
	}

	_, err := EncodeRegistryAuth(&authInfo)
	assert.Nil(t, err)
}
