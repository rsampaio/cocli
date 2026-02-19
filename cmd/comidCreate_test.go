// Copyright 2021-2026 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

const comidWithDependencyTriplesTemplate = `{
  "tag-identity": {"id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890"},
  "triples": {
    "reference-values": [
      {
        "environment": {
          "class": {
            "id": {"type": "uuid", "value": "DD6661F0-0928-4401-966B-589EA74E3272"},
            "model": "FMC",
            "layer": 0,
            "index": 0
          }
        },
        "measurements": [{"value": {"digests": ["sha-256:RKozavTLFKh5Qy5T3WVxx/qbzK+3X0iCWSYtbqOk2Rs="]}}]
      }
    ],
    "dependency-triples": [
      {
        "subject": {"type": "uuid", "value": "DD6661F0-0928-4401-966B-589EA74E3272"},
        "dependents": [
          {"type": "uuid", "value": "FFDA7CF3-2333-4A91-99A8-068626203ACA"},
          {"type": "text", "value": "downstream-domain"}
        ]
      }
    ]
  }
}`

func Test_ComidCreateCmd_template_with_dependency_triples(t *testing.T) {
	cmd := NewComidCreateCmd()
	fs = afero.NewMemMapFs()
	err := afero.WriteFile(fs, "with-dep-triples.json", []byte(comidWithDependencyTriplesTemplate), 0644)
	require.NoError(t, err)

	cmd.SetArgs([]string{"--template=with-dep-triples.json"})
	err = cmd.Execute()
	assert.NoError(t, err)

	_, err = fs.Stat("with-dep-triples.cbor")
	assert.NoError(t, err)
}
