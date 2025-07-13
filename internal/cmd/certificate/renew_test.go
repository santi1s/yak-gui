package certificate

import (
	"errors"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/stretchr/testify/assert"
)

func TestBuildRoute53Record(t *testing.T) {
	hclFile := hclwrite.NewEmptyFile()
	r := buildRoute53Record("my_record", "my.zone.id", "record.name", "my_record_value")
	hclFile.Body().AppendBlock(r)
	assert.Equal(t, string(hclFile.Bytes()), `resource "aws_route53_record" "my_record" {
  zone_id = my.zone.id
  name    = "record.name"
  records = ["my_record_value"]
  type    = "CNAME"
  ttl     = 120
}
`, "should build the correct terraform resource for route53")
}

func TestBuildCloudflareRecord(t *testing.T) {
	hclFile := hclwrite.NewEmptyFile()
	r := buildCloudflareRecord("my_record", "my.zone.id", "record_name", "my_record_value")
	hclFile.Body().AppendBlock(r)
	assert.Equal(t, string(hclFile.Bytes()), `resource "cloudflare_record" "my_record" {
  zone_id = my.zone.id
  name    = "record_name"
  value   = "my_record_value"
  type    = "CNAME"
  ttl     = 120
}
`, "should build the correct terraform resource for cloudflare")
}

func TestSearchAndPatchRecord(t *testing.T) {
	terraformFile := `
# foobar

resource "aws_route53_record" "gandi-dcv-wildcard-doctolib_tech" {
  zone_id = module.route53_external_doctolib_tech.zone_id
  name    = "_464A5C5C943F855929A61D03DC790FB4.doctolib.tech"
  records = ["c5af681f59e00a45eadf4fb6c3b52c97.223720d98a6c295acbc6170754a0a9ef.9e3aaf2e75599987dc69.sectigo.com"]
  type    = "CNAME"
  ttl     = 120
}

# foobar`
	file, diags := hclwrite.ParseConfig([]byte(terraformFile), "test.tf", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		err := errors.New("an error occurred")
		if err != nil {
			t.Fatal(err)
		}
	}

	newResouce := buildRoute53Record(
		"gandi-dcv-wildcard-doctolib_tech",
		"module.route53_external_doctolib_tech.zone_id",
		"_FOOBAR.doctolib.tech",
		"foobar.sectigo.com",
	)

	patched := searchAndPatchRecord(file, "gandi-dcv-wildcard-doctolib_tech", *newResouce)
	assert.Equal(t, true, patched)

	assert.Equal(t, string(file.Bytes()), `
# foobar

resource "aws_route53_record" "gandi-dcv-wildcard-doctolib_tech" {
  zone_id = module.route53_external_doctolib_tech.zone_id
  name    = "_FOOBAR.doctolib.tech"
  records = ["foobar.sectigo.com"]
  type    = "CNAME"
  ttl     = 120
}

# foobar`, "should build the correct terraform resource")
}

func TestGenerateTFResourceName(t *testing.T) {
	tests := map[string]struct {
		dcvRecord            *DCVRecord
		expectedDomain       string
		expectedResourceName string
	}{
		"test1": {
			dcvRecord: &DCVRecord{
				Name: "_AE72C66BC33BD769FA49208A140A40A7.doctolib.net.",
			},
			expectedDomain:       "doctolib.net",
			expectedResourceName: "gandi-dcv-doctolib_net",
		},
		"test2": {
			dcvRecord: &DCVRecord{
				Name: "_36F9BCD4C0D575944D4E8777385A8631.portal-staging.doctolib.it.",
			},
			expectedDomain:       "portal-staging.doctolib.it",
			expectedResourceName: "gandi-dcv-portal-staging_doctolib_it",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expectedResourceName, test.dcvRecord.generateTFResourceName())
			assert.Equal(t, test.expectedDomain, test.dcvRecord.domain())
		})
	}
}

func TestSanitizeValue(t *testing.T) {
	tests := map[string]struct {
		value    string
		expected string
	}{
		"test1": {
			value:    "c5af681f59e00a45eadf4fb6c3b52c97.223720d98a6c295acbc6170754a0a9ef.9e3aaf2e75599987dc69.sectigo.com.",
			expected: "c5af681f59e00a45eadf4fb6c3b52c97.223720d98a6c295acbc6170754a0a9ef.9e3aaf2e75599987dc69.sectigo.com",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expected, sanitizeValue(test.value))
		})
	}
}

func TestBuildDCVRecords(t *testing.T) {
	terraformFile := `# foobar

resource "aws_route53_record" "not-to-be-modified" {
  zone_id = module.route53_external_doctolib_tech.zone_id
  name    = "www.doctolib.tech"
  records = ["127.0.01"]
  type    = "A"
  ttl     = 120
}

resource "aws_route53_record" "gandi-dcv-wildcard-doctolib_tech" {
  zone_id = module.route53_external_doctolib_tech.zone_id
  name    = "_464A5C5C943F855929A61D03DC790FB4.doctolib.tech"
  records  = ["42.sectigo.com"]
  type    = "CNAME"
  ttl     = 120
}

resource "aws_route53_record" "gandi-dcv-to-be-modified_doctolib_tech" {
  zone_id = module.route53_external_doctolib_tech.zone_id
  name    = "_464A5C5C943F855929A61D03DC790FB4.doctolib.tech"
  records = ["42.sectigo.com"]
  type    = "CNAME"
  ttl     = 120
}

# foobar 2`

	file, diags := hclwrite.ParseConfig([]byte(terraformFile), "test.tf", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		err := errors.New("an error occurred")
		if err != nil {
			t.Fatal(err)
		}
	}

	dcvRecords := []DCVRecord{
		{
			Name:  "_FOOBAR.foobar.doctolib.tech.",
			Value: "foobar.sectigo.com.",
		},
		{
			Name:  "_BARFOO.barfoo.doctolib.tech",
			Value: "barfoo.sectigo.com.",
		},
		{
			Name:  "_BARFOO.to-be-modified.doctolib.tech",
			Value: "to-be-modified.sectigo.com.",
		},
	}

	config := &CertConfig{
		Route53: &Route53Config{
			Zone: "module.route53_external_doctolib_tech.zone_id",
		},
	}

	err := buildDCVRecords(config, dcvRecords, file)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, `# foobar

resource "aws_route53_record" "not-to-be-modified" {
  zone_id = module.route53_external_doctolib_tech.zone_id
  name    = "www.doctolib.tech"
  records = ["127.0.01"]
  type    = "A"
  ttl     = 120
}

resource "aws_route53_record" "gandi-dcv-wildcard-doctolib_tech" {
  zone_id = module.route53_external_doctolib_tech.zone_id
  name    = "_464A5C5C943F855929A61D03DC790FB4.doctolib.tech"
  records = ["42.sectigo.com"]
  type    = "CNAME"
  ttl     = 120
}

resource "aws_route53_record" "gandi-dcv-to-be-modified_doctolib_tech" {
  zone_id = module.route53_external_doctolib_tech.zone_id
  name    = "_BARFOO.to-be-modified.doctolib.tech"
  records = ["to-be-modified.sectigo.com"]
  type    = "CNAME"
  ttl     = 120
}

# foobar 2
resource "aws_route53_record" "gandi-dcv-foobar_doctolib_tech" {
  zone_id = module.route53_external_doctolib_tech.zone_id
  name    = "_FOOBAR.foobar.doctolib.tech"
  records = ["foobar.sectigo.com"]
  type    = "CNAME"
  ttl     = 120
}

resource "aws_route53_record" "gandi-dcv-barfoo_doctolib_tech" {
  zone_id = module.route53_external_doctolib_tech.zone_id
  name    = "_BARFOO.barfoo.doctolib.tech"
  records = ["barfoo.sectigo.com"]
  type    = "CNAME"
  ttl     = 120
}
`, string(file.Bytes()), "should build the correct terraform resource")
}

func TestRemoveDomain(t *testing.T) {
	tests := map[string]struct {
		domain   string
		expected string
	}{
		"4 levels": {
			domain:   "FOOBAR.portal.doctolib.net",
			expected: "FOOBAR.portal",
		},
		"3 levels": {
			domain:   "FOOBAR.doctolib.net",
			expected: "FOOBAR",
		},
		"4 levels dotted": {
			domain:   "FOOBAR.portal.doctolib.net.",
			expected: "FOOBAR.portal",
		},
		"3 levels dotted": {
			domain:   "FOOBAR.doctolib.net.",
			expected: "FOOBAR",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expected, removeDomain(test.domain))
		})
	}
}
