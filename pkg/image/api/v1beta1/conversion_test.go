package v1beta1_test

import (
	"reflect"
	"testing"

	kapi "github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/util"
	"github.com/fsouza/go-dockerclient"

	_ "github.com/openshift/origin/pkg/api/latest"
	newer "github.com/openshift/origin/pkg/image/api"
)

var Convert = kapi.Scheme.Convert

func TestRoundTripVersionedObject(t *testing.T) {
	d := &newer.DockerImage{
		Config: newer.DockerConfig{
			Env: []string{"A=1", "B=2"},
		},
	}
	i := &newer.Image{
		ObjectMeta: kapi.ObjectMeta{Name: "foo"},

		DockerImageMetadata:  *d,
		DockerImageReference: "foo/bar/baz",
	}

	data, err := kapi.Scheme.EncodeToVersion(i, "v1beta1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	obj, err := kapi.Scheme.Decode(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	image := obj.(*newer.Image)
	if image.DockerImageMetadataVersion != "1.0" {
		t.Errorf("did not default to correct metadata version: %#v", image)
	}
	image.DockerImageMetadataVersion = ""
	if !reflect.DeepEqual(i, image) {
		t.Errorf("unable to round trip object: %s", util.ObjectDiff(i, image))
	}
}

// This tests that JSON generated by an older version of v1beta1 still correctly parses and versions
func TestDecodeExistingAPIObjects(t *testing.T) {
	obj, err := kapi.Scheme.Decode([]byte(`{
		"kind":"Image",
		"apiVersion":"v1beta1",
		"metadata":{
			"name":"foo"
		},
		"dockerImageReference":"foo/bar/baz",
		"dockerImageMetadata":{
			"Id":"0001",
			"Config":{
				"Env":["A=1","B=2"]
				}
			}
		}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	image := obj.(*newer.Image)
	if image.Name != "foo" || image.DockerImageReference != "foo/bar/baz" || image.DockerImageMetadata.ID != "0001" || image.DockerImageMetadata.Config.Env[0] != "A=1" {
		t.Errorf("unexpected object: %#v", image)
	}
}

func TestDecodeDockerRegistryJSON(t *testing.T) {
	oldImage := docker.ImagePre012{
		ID: "something",
		Config: &docker.Config{
			Env: []string{"A=1", "B=2"},
		},
	}
	newImage := newer.DockerImage{}
	if err := kapi.Scheme.Convert(&oldImage, &newImage); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newImage.ID != "something" || newImage.Config.Env[0] != "A=1" {
		t.Errorf("unexpected object: %#v", newImage)
	}
}