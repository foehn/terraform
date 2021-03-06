package google

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"google.golang.org/api/compute/v1"
)

func TestAccComputeBackendService_basic(t *testing.T) {
	var svc compute.BackendService

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeBackendServiceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccComputeBackendService_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeBackendServiceExists(
						"google_compute_backend_service.foobar", &svc),
				),
			},
			resource.TestStep{
				Config: testAccComputeBackendService_basicModified,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeBackendServiceExists(
						"google_compute_backend_service.foobar", &svc),
				),
			},
		},
	})
}

func TestAccComputeBackendService_withBackend(t *testing.T) {
	var svc compute.BackendService

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckComputeBackendServiceDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccComputeBackendService_withBackend,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeBackendServiceExists(
						"google_compute_backend_service.lipsum", &svc),
				),
			},
		},
	})

	if svc.TimeoutSec != 10 {
		t.Errorf("Expected TimeoutSec == 10, got %d", svc.TimeoutSec)
	}
	if svc.Protocol != "HTTP" {
		t.Errorf("Expected Protocol to be HTTP, got %q", svc.Protocol)
	}
	if len(svc.Backends) != 1 {
		t.Errorf("Expected 1 backend, got %d", len(svc.Backends))
	}
}

func testAccCheckComputeBackendServiceDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "google_compute_backend_service" {
			continue
		}

		_, err := config.clientCompute.BackendServices.Get(
			config.Project, rs.Primary.ID).Do()
		if err == nil {
			return fmt.Errorf("Backend service still exists")
		}
	}

	return nil
}

func testAccCheckComputeBackendServiceExists(n string, svc *compute.BackendService) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.clientCompute.BackendServices.Get(
			config.Project, rs.Primary.ID).Do()
		if err != nil {
			return err
		}

		if found.Name != rs.Primary.ID {
			return fmt.Errorf("Backend service not found")
		}

		*svc = *found

		return nil
	}
}

const testAccComputeBackendService_basic = `
resource "google_compute_backend_service" "foobar" {
    name = "blablah"
    health_checks = ["${google_compute_http_health_check.zero.self_link}"]
}

resource "google_compute_http_health_check" "zero" {
    name = "tf-test-zero"
    request_path = "/"
    check_interval_sec = 1
    timeout_sec = 1
}
`

const testAccComputeBackendService_basicModified = `
resource "google_compute_backend_service" "foobar" {
    name = "blablah"
    health_checks = ["${google_compute_http_health_check.one.self_link}"]
}

resource "google_compute_http_health_check" "zero" {
    name = "tf-test-zero"
    request_path = "/"
    check_interval_sec = 1
    timeout_sec = 1
}

resource "google_compute_http_health_check" "one" {
    name = "tf-test-one"
    request_path = "/one"
    check_interval_sec = 30
    timeout_sec = 30
}
`

const testAccComputeBackendService_withBackend = `
resource "google_compute_backend_service" "lipsum" {
    name = "hello-world-bs"
    description = "Hello World 1234"
    port_name = "http"
    protocol = "HTTP"
    timeout_sec = 10

    backend {
        group = "${google_compute_instance_group_manager.foobar.instance_group}"
    }

    health_checks = ["${google_compute_http_health_check.default.self_link}"]
}

resource "google_compute_instance_group_manager" "foobar" {
    name = "terraform-test"
    instance_template = "${google_compute_instance_template.foobar.self_link}"
    base_instance_name = "foobar"
    zone = "us-central1-f"
    target_size = 1
}

resource "google_compute_instance_template" "foobar" {
    name = "terraform-test"
    machine_type = "n1-standard-1"

    network_interface {
        network = "default"
    }

    disk {
        source_image = "debian-7-wheezy-v20140814"
        auto_delete = true
        boot = true
    }
}

resource "google_compute_http_health_check" "default" {
    name = "test2"
    request_path = "/"
    check_interval_sec = 1
    timeout_sec = 1
}
`
