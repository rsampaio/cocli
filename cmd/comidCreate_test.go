// Copyright 2021-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/veraison/corim/comid"
	"github.com/veraison/corim/profiles/tdx"
)

func Test_ComidCreateCmd_unknown_argument(t *testing.T) {
	cmd := NewComidCreateCmd()

	args := []string{"--unknown-argument=val"}
	cmd.SetArgs(args)

	err := cmd.Execute()
	assert.EqualError(t, err, "unknown flag: --unknown-argument")
}

func Test_ComidCreateCmd_no_templates(t *testing.T) {
	cmd := NewComidCreateCmd()

	// no args

	err := cmd.Execute()
	assert.EqualError(t, err, "no templates supplied")
}

func Test_ComidCreateCmd_no_files_found(t *testing.T) {
	cmd := NewComidCreateCmd()

	args := []string{
		"--template=unknown",
		"--template-dir=unsure",
	}
	cmd.SetArgs(args)

	err := cmd.Execute()
	assert.EqualError(t, err, "no files found")
}

func Test_ComidCreateCmd_template_with_invalid_json(t *testing.T) {
	var err error

	cmd := NewComidCreateCmd()

	fs = afero.NewMemMapFs()
	err = afero.WriteFile(fs, "invalid.json", []byte("..."), 0644)
	require.NoError(t, err)

	args := []string{
		"--template=invalid.json",
	}
	cmd.SetArgs(args)

	err = cmd.Execute()
	assert.EqualError(t, err, "1/1 creations(s) failed")
}

func Test_ComidCreateCmd_template_with_invalid_comid(t *testing.T) {
	var err error

	cmd := NewComidCreateCmd()

	fs = afero.NewMemMapFs()
	err = afero.WriteFile(fs, "bad-comid.json", []byte("{}"), 0644)
	require.NoError(t, err)

	args := []string{
		"--template=bad-comid.json",
	}
	cmd.SetArgs(args)

	err = cmd.Execute()
	assert.EqualError(t, err, "1/1 creations(s) failed")
}

func Test_ComidCreateCmd_template_from_file_to_default_dir(t *testing.T) {
	var err error

	cmd := NewComidCreateCmd()

	fs = afero.NewMemMapFs()
	err = afero.WriteFile(fs, "ok.json", testComidTemplate, 0644)
	require.NoError(t, err)

	args := []string{
		"--template=ok.json",
	}
	cmd.SetArgs(args)

	err = cmd.Execute()
	assert.NoError(t, err)

	expectedFileName := "ok.cbor"

	_, err = fs.Stat(expectedFileName)
	assert.NoError(t, err)
}

func Test_ComidCreateCmd_template_from_dir_to_custom_dir(t *testing.T) {
	var err error

	cmd := NewComidCreateCmd()

	fs = afero.NewMemMapFs()
	err = afero.WriteFile(fs, "testdir/ok.json", testComidTemplate, 0644)
	require.NoError(t, err)

	args := []string{
		"--template-dir=testdir",
		"--output-dir=testdir",
	}
	cmd.SetArgs(args)

	err = cmd.Execute()
	assert.NoError(t, err)

	expectedFileName := "testdir/ok.cbor"

	_, err = fs.Stat(expectedFileName)
	assert.NoError(t, err)
}

func Test_ComidCreateCmd_WithProfile(t *testing.T) {
	var err error
	profile := "--profile=" + testProfile
	cmd := NewComidCreateCmd()
	fs = afero.NewMemMapFs()
	err = afero.WriteFile(fs, "ok.json", []byte(tdx.TDXSeamRefValJSONTemplate), 0644)
	require.NoError(t, err)

	args := []string{
		"--template=ok.json",
		profile,
	}
	cmd.SetArgs(args)

	err = cmd.Execute()
	assert.NoError(t, err)

	expectedFileName := "ok.cbor"

	_, err = fs.Stat(expectedFileName)
	assert.NoError(t, err)

}

func Test_ComidCreateCmd_InvalidProfile(t *testing.T) {
	var err error
	profile := "--profile=" + testInvalidProfile
	cmd := NewComidCreateCmd()
	fs = afero.NewMemMapFs()
	err = afero.WriteFile(fs, "ok.json", []byte(tdx.TDXSeamRefValJSONTemplate), 0644)
	require.NoError(t, err)

	args := []string{
		"--template=ok.json",
		profile,
	}
	cmd.SetArgs(args)

	err = cmd.Execute()
	assert.EqualError(t, err, "1/1 creations(s) failed")
}

// Test_ComidCreateCmd_template_with_dependency_triples checks that:
// 1) A CoMID template containing draft-9 dependency-triples (domain-id + trustees) can be
//    loaded from data/comid/templates and encoded to CBOR without error.
// 2) The generated CBOR round-trips and contains the expected dependency-triples.
func Test_ComidCreateCmd_template_with_dependency_triples(t *testing.T) {
	templatePath := filepath.Join("..", "data", "comid", "templates", "comid-with-dependency-triples.json")
	if _, err := os.Stat(templatePath); err != nil {
		t.Skipf("template not found: %s", templatePath)
	}

	cmd := NewComidCreateCmd()
	fs = afero.NewOsFs()
	outDir := t.TempDir()
	cborPath := filepath.Join(outDir, "comid-with-dependency-triples.cbor")

	cmd.SetArgs([]string{"--template=" + templatePath, "--output-dir=" + outDir})
	err := cmd.Execute()
	require.NoError(t, err)

	_, err = os.Stat(cborPath)
	require.NoError(t, err, "output CBOR file should exist")

	// Round-trip: decode CBOR and assert dependency-triples are present and valid.
	cborData, err := os.ReadFile(cborPath)
	require.NoError(t, err)
	var c comid.Comid
	err = c.FromCBOR(cborData)
	require.NoError(t, err)
	require.NotNil(t, c.Triples, "triples should be set")
	require.NotNil(t, c.Triples.DomainDependencies, "dependency-triples should be set")
	require.False(t, c.Triples.DomainDependencies.IsEmpty(), "dependency-triples should not be empty")
	dd := *c.Triples.DomainDependencies
	require.Len(t, dd, 1, "template has one dependency-triple")
	assert.GreaterOrEqual(t, len(dd[0].Trustees), 1, "triple should have at least one trustee")
	err = dd[0].Valid()
	assert.NoError(t, err)
}

// Test_ComidCreateCmd_template_with_invalid_dependency_triples checks that creation fails
// when the template has invalid dependency-triples (e.g. empty trustees).
func Test_ComidCreateCmd_template_with_invalid_dependency_triples(t *testing.T) {
	invalidTemplate := `{
  "tag-identity": {"id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890"},
  "triples": {
    "reference-values": [
      {
        "environment": {"class": {"id": {"type": "uuid", "value": "DD6661F0-0928-4401-966B-589EA74E3272"}}},
        "measurements": [{"value": {"digests": ["sha-256:RKozavTLFKh5Qy5T3WVxx/qbzK+3X0iCWSYtbqOk2Rs="]}}]
      }
    ],
    "dependency-triples": [
      {
        "domain-id": {"class": {"id": {"type": "uuid", "value": "DD6661F0-0928-4401-966B-589EA74E3272"}}},
        "trustees": []
      }
    ]
  }
}`
	cmd := NewComidCreateCmd()
	fs = afero.NewMemMapFs()
	require.NoError(t, afero.WriteFile(fs, "bad.json", []byte(invalidTemplate), 0644))

	cmd.SetArgs([]string{"--template=bad.json"})
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed")
}
