// Copyright 2019 Google Cloud Platform Proxy Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package integration

import (
	"fmt"
	"strings"
	"testing"

	"cloudesf.googlesource.com/gcpproxy/src/go/util"
	"cloudesf.googlesource.com/gcpproxy/tests/endpoints/echo/client"
	"cloudesf.googlesource.com/gcpproxy/tests/env"
	"cloudesf.googlesource.com/gcpproxy/tests/utils"

	comp "cloudesf.googlesource.com/gcpproxy/tests/env/components"
)

func TestServiceControlCache(t *testing.T) {
	serviceName := "test-echo"
	configId := "test-config-id"

	args := []string{"--service=" + serviceName, "--version=" + configId,
		"--backend_protocol=http1", "--rollout_strategy=fixed", "--suppress_envoy_headers"}

	s := env.NewTestEnv(comp.TestServiceControlCache, "echo", []string{"google_jwt"})
	if err := s.Setup(args); err != nil {
		t.Fatalf("fail to setup test env, %v", err)
	}
	defer s.TearDown()

	url := fmt.Sprintf("http://localhost:%v%v%v", s.Ports().ListenerPort, "/echo", "?key=api-key")
	message := "hello"
	num := 10
	wantResp := `{"message":"hello"}`
	for i := 0; i < num; i++ {
		resp, err := client.DoPost(url, message)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(resp), wantResp) {
			t.Errorf("expected: %s, got: %s", wantResp, string(resp))
		}
	}

	wantScRequests := []interface{}{
		&utils.ExpectedCheck{
			Version:         utils.APIProxyVersion,
			ServiceName:     "echo-api.endpoints.cloudesf-testing.cloud.goog",
			ServiceConfigID: "test-config-id",
			ConsumerID:      "api_key:api-key",
			OperationName:   "1.echo_api_endpoints_cloudesf_testing_cloud_goog.Echo",
			CallerIp:        "127.0.0.1",
		},
		&utils.ExpectedReport{
			Aggregate:         int64(num),
			Version:           utils.APIProxyVersion,
			ServiceName:       "echo-api.endpoints.cloudesf-testing.cloud.goog",
			ServiceConfigID:   "test-config-id",
			URL:               "/echo?key=api-key",
			ApiKey:            "api-key",
			ApiMethod:         "1.echo_api_endpoints_cloudesf_testing_cloud_goog.Echo",
			ProducerProjectID: "producer-project",
			ConsumerProjectID: "123456",
			FrontendProtocol:  "http",
			HttpMethod:        "POST",
			LogMessage:        "1.echo_api_endpoints_cloudesf_testing_cloud_goog.Echo is called",
			StatusCode:        "0",
			RequestSize:       238,
			ResponseSize:      126,
			RequestBytes:      238,
			ResponseBytes:     126,
			ResponseCode:      200,
			Platform:          util.GCE,
			Location:          "test-zone",
		},
	}

	scRequests, err := s.ServiceControlServer.GetRequests(len(wantScRequests))
	if err != nil {
		t.Fatalf("GetRequests returns error: %v", err)
	}

	utils.CheckScRequest(t, scRequests, wantScRequests, "TestServiceControlCache")
}