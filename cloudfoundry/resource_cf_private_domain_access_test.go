package cloudfoundry

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/cfapi"
)

const privateDomainAccessResourceCreate = `
resource "cloudfoundry_org" "org1" {
  name = "org1"
}

resource "cloudfoundry_org" "org2" {
  name = "org2"
}

resource "cloudfoundry_org" "org3" {
  name = "org3"
}

resource "cloudfoundry_domain" "private" {
    sub_domain = "private"
    domain     = "%s"
    org        = "${cloudfoundry_org.org1.id}"
}

resource "cloudfoundry_private_domain_access" "access-to-org" {
    domain     = "${cloudfoundry_domain.private.id}"
    org        = "${cloudfoundry_org.org2.id}"
}
`

const privateDomainAccessResourceUpdate = `
resource "cloudfoundry_org" "org1" {
  name = "org1"
}

resource "cloudfoundry_org" "org2" {
  name = "org2"
}

resource "cloudfoundry_org" "org3" {
  name = "org3"
}

resource "cloudfoundry_domain" "private" {
    sub_domain = "private"
    domain     = "%s"
    org        = "${cloudfoundry_org.org1.id}"
}

resource "cloudfoundry_private_domain_access" "access-to-org" {
    domain     = "${cloudfoundry_domain.private.id}"
    org        = "${cloudfoundry_org.org3.id}"
}
`

const privateDomainAccessResourceDelete = `
resource "cloudfoundry_org" "org1" {
  name = "org1"
}

resource "cloudfoundry_org" "org2" {
  name = "org2"
}

resource "cloudfoundry_org" "org3" {
  name = "org3"
}

resource "cloudfoundry_domain" "private" {
    sub_domain = "private"
    domain     = "%s"
    org        = "${cloudfoundry_org.org1.id}"
}
`

func TestAccPrivateDomainAccess_normal(t *testing.T) {
	ref := "cloudfoundry_private_domain_access.access-to-org"

	resource.Test(t,
		resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				resource.TestStep{
					Config: fmt.Sprintf(privateDomainAccessResourceCreate, defaultAppDomain()),
					Check: resource.ComposeTestCheckFunc(
						checkPrivateDomainShare(ref, "cloudfoundry_domain.private", "cloudfoundry_org.org2", true),
						checkPrivateDomainShare(ref, "cloudfoundry_domain.private", "cloudfoundry_org.org3", false),
					),
				},
				resource.TestStep{
					Config: fmt.Sprintf(privateDomainAccessResourceUpdate, defaultAppDomain()),
					Check: resource.ComposeTestCheckFunc(
						checkPrivateDomainShare(ref, "cloudfoundry_domain.private", "cloudfoundry_org.org2", false),
						checkPrivateDomainShare(ref, "cloudfoundry_domain.private", "cloudfoundry_org.org3", true),
					),
				},
				resource.TestStep{
					Config: fmt.Sprintf(privateDomainAccessResourceDelete, defaultAppDomain()),
					Check: resource.ComposeTestCheckFunc(
						checkPrivateDomainShare("", "cloudfoundry_domain.private", "cloudfoundry_org.org2", false),
						checkPrivateDomainShare("", "cloudfoundry_domain.private", "cloudfoundry_org.org3", false),
					),
				},
			},
		})
}

func checkPrivateDomainShare(resource, domain, org string, exists bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		session := testAccProvider.Meta().(*cfapi.Session)

		drs, ok := s.RootModule().Resources[domain]
		if !ok {
			return fmt.Errorf("domain '%s' not found in terraform state", domain)
		}

		ors, ok := s.RootModule().Resources[org]
		if !ok {
			return fmt.Errorf("org '%s' not found in terraform state", org)
		}

		orgID := ors.Primary.ID
		domainID := drs.Primary.ID

		dm := session.DomainManager()
		found, err := dm.HasPrivateDomainAccess(orgID, domainID)
		if err != nil {
			return err
		}

		if !found && exists {
			return fmt.Errorf("unable to find private domain access '%s(%s)' to org '%s(%s)'", domain, domainID, org, orgID)
		}

		if found && !exists {
			return fmt.Errorf("private domain access '%s(%s)' to org '%s(%s)' not deleted as it ought to be", domain, domainID, org, orgID)
		}

		if len(resource) > 0 {
			rs, ok := s.RootModule().Resources[resource]
			if !ok {
				return fmt.Errorf("private_domain_access '%s' not found in terraform state", resource)
			}
			session.Log.DebugMessage("terraform state for resource '%s': %# v", resource, rs)

			id := rs.Primary.ID

			if exists && id != fmt.Sprintf("%s/%s", orgID, domainID) {
				return fmt.Errorf("unexpected private_domain_access resource identifier '%s' mismatch with '%s/%s'", id, orgID, domainID)
			}
		}

		return nil
	}
}

// Local Variables:
// ispell-local-dictionary: "american"
// End:
