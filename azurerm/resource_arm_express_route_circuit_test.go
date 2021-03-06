package azurerm

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2017-09-01/network"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func TestAccAzureRMExpressRouteCircuit(t *testing.T) {
	// NOTE: this is a combined test rather than separate split out tests due to
	// Azure only being happy about provisioning a couple at a time
	testCases := map[string]map[string]func(t *testing.T){
		"basic": {
			"metered":   testAccAzureRMExpressRouteCircuit_basicMetered,
			"unlimited": testAccAzureRMExpressRouteCircuit_basicUnlimited,
			"update":    testAccAzureRMExpressRouteCircuit_update,
		},
		"PrivatePeering": {
			"azurePrivatePeering":  testAccAzureRMExpressRouteCircuitPeering_azurePrivatePeering,
			"importPrivatePeering": testAccAzureRMExpressRouteCircuitPeering_importAzurePrivatePeering,
		},
		"PublicPeering": {
			"azurePublicPeering":  testAccAzureRMExpressRouteCircuitPeering_azurePublicPeering,
			"importPublicPeering": testAccAzureRMExpressRouteCircuitPeering_importAzurePublicPeering,
		},
		"MicrosoftPeering": {
			"microsoftPeering":       testAccAzureRMExpressRouteCircuitPeering_microsoftPeering,
			"importMicrosoftPeering": testAccAzureRMExpressRouteCircuitPeering_importMicrosoftPeering,
		},
		"authorization": {
			"basic":    testAccAzureRMExpressRouteCircuitAuthorization_basic,
			"multiple": testAccAzureRMExpressRouteCircuitAuthorization_multiple,
			"import":   testAccAzureRMExpressRouteCircuitAuthorization_importBasic,
		},
		"authorizationImport": {
			"basic": testAccAzureRMExpressRouteCircuitAuthorization_importBasic,
		},
		"circuitImport": {
			"metered":   testAccAzureRMExpressRouteCircuit_importMetered,
			"unlimited": testAccAzureRMExpressRouteCircuit_importUnlimited,
		},
	}

	for group, m := range testCases {
		m := m
		t.Run(group, func(t *testing.T) {
			for name, tc := range m {
				tc := tc
				t.Run(name, func(t *testing.T) {
					tc(t)
				})
			}
		})
	}
}

func testAccAzureRMExpressRouteCircuit_basicMetered(t *testing.T) {
	var erc network.ExpressRouteCircuit
	ri := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMExpressRouteCircuitDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMExpressRouteCircuit_basicMeteredConfig(ri, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMExpressRouteCircuitExists("azurerm_express_route_circuit.test", &erc),
				),
			},
		},
	})
}

func testAccAzureRMExpressRouteCircuit_basicUnlimited(t *testing.T) {
	var erc network.ExpressRouteCircuit
	ri := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMExpressRouteCircuitDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMExpressRouteCircuit_basicUnlimitedConfig(ri, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMExpressRouteCircuitExists("azurerm_express_route_circuit.test", &erc),
				),
			},
		},
	})
}

func testAccAzureRMExpressRouteCircuit_update(t *testing.T) {
	resourceName := "azurerm_express_route_circuit.test"
	var erc network.ExpressRouteCircuit
	ri := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMExpressRouteCircuitDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAzureRMExpressRouteCircuit_basicMeteredConfig(ri, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMExpressRouteCircuitExists(resourceName, &erc),
					resource.TestCheckResourceAttr(resourceName, "sku.0.family", "MeteredData"),
				),
			},
			{
				Config: testAccAzureRMExpressRouteCircuit_basicUnlimitedConfig(ri, testLocation()),
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMExpressRouteCircuitExists(resourceName, &erc),
					resource.TestCheckResourceAttr(resourceName, "sku.0.family", "UnlimitedData"),
				),
			},
		},
	})
}

func testCheckAzureRMExpressRouteCircuitExists(name string, erc *network.ExpressRouteCircuit) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		expressRouteCircuitName := rs.Primary.Attributes["name"]
		resourceGroup, hasResourceGroup := rs.Primary.Attributes["resource_group_name"]
		if !hasResourceGroup {
			return fmt.Errorf("Bad: no resource group found in state for Express Route Circuit: %s", expressRouteCircuitName)
		}

		client := testAccProvider.Meta().(*ArmClient).expressRouteCircuitClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext

		resp, err := client.Get(ctx, resourceGroup, expressRouteCircuitName)
		if err != nil {
			if utils.ResponseWasNotFound(resp.Response) {
				return fmt.Errorf("Bad: Express Route Circuit %q (resource group: %q) does not exist", expressRouteCircuitName, resourceGroup)
			}

			return fmt.Errorf("Bad: Get on expressRouteCircuitClient: %+v", err)
		}

		*erc = resp

		return nil
	}
}

func testCheckAzureRMExpressRouteCircuitDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*ArmClient).expressRouteCircuitClient
	ctx := testAccProvider.Meta().(*ArmClient).StopContext

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azurerm_express_route_circuit" {
			continue
		}

		name := rs.Primary.Attributes["name"]
		resourceGroup := rs.Primary.Attributes["resource_group_name"]

		resp, err := client.Get(ctx, resourceGroup, name)

		if err != nil {
			return nil
		}

		if resp.StatusCode != http.StatusNotFound {
			return fmt.Errorf("Express Route Circuit still exists:\n%#v", resp.ExpressRouteCircuitPropertiesFormat)
		}
	}

	return nil
}

func testAccAzureRMExpressRouteCircuit_basicMeteredConfig(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurerm_resource_group" "test" {
  name     = "acctestrg-%d"
  location = "%s"
}

resource "azurerm_express_route_circuit" "test" {
  name                  = "acctest-erc-%d"
  location              = "${azurerm_resource_group.test.location}"
  resource_group_name   = "${azurerm_resource_group.test.name}"
  service_provider_name = "Equinix"
  peering_location      = "Silicon Valley"
  bandwidth_in_mbps     = 50

  sku {
    tier   = "Standard"
    family = "MeteredData"
  }

  allow_classic_operations = false

  tags {
    Environment = "production"
    Purpose     = "AcceptanceTests"
  }
}
`, rInt, location, rInt)
}
func testAccAzureRMExpressRouteCircuit_basicUnlimitedConfig(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurerm_resource_group" "test" {
  name     = "acctestrg-%d"
  location = "%s"
}

resource "azurerm_express_route_circuit" "test" {
  name                  = "acctest-erc-%d"
  location              = "${azurerm_resource_group.test.location}"
  resource_group_name   = "${azurerm_resource_group.test.name}"
  service_provider_name = "Equinix"
  peering_location      = "Silicon Valley"
  bandwidth_in_mbps     = 50

  sku {
    tier   = "Standard"
    family = "UnlimitedData"
  }

  allow_classic_operations = false

  tags {
    Environment = "production"
    Purpose     = "AcceptanceTests"
  }
}
`, rInt, location, rInt)
}
