package promotions

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	api "github.com/akuity/kargo/api/v1alpha1"
)

func TestApplyKustomize(t *testing.T) {
	testCases := []struct {
		name                string
		newState            api.EnvironmentState
		update              api.KustomizePromotionMechanism
		kustomizeSetImageFn func(dir, repo, tag string) error
		assertions          func(changeSummary []string, err error)
	}{
		{
			name: "error setting image",
			newState: api.EnvironmentState{
				Images: []api.Image{
					{
						RepoURL: "fake-url",
						Tag:     "fake-tag",
					},
				},
			},
			update: api.KustomizePromotionMechanism{
				Images: []api.KustomizeImageUpdate{
					{
						Image: "fake-url",
						Path:  "/fake/path",
					},
				},
			},
			kustomizeSetImageFn: func(string, string, string) error {
				return errors.New("something went wrong")
			},
			assertions: func(changeSummary []string, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "error updating image")
				require.Contains(t, err.Error(), "something went wrong")
				require.Nil(t, changeSummary)
			},
		},

		{
			name: "success",
			newState: api.EnvironmentState{
				Images: []api.Image{
					{
						RepoURL: "fake-url",
						Tag:     "fake-tag",
					},
				},
			},
			update: api.KustomizePromotionMechanism{
				Images: []api.KustomizeImageUpdate{
					{
						Image: "fake-url",
						Path:  "fake/path",
					},
				},
			},
			kustomizeSetImageFn: func(string, string, string) error {
				return nil
			},
			assertions: func(changeSummary []string, err error) {
				require.NoError(t, err)
				require.Len(t, changeSummary, 1)
				require.Equal(
					t,
					"updated fake/path/kustomization.yaml to use image "+
						"fake-url:fake-tag",
					changeSummary[0],
				)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			r := reconciler{
				kustomizeSetImageFn: testCase.kustomizeSetImageFn,
			}
			testCase.assertions(
				r.applyKustomize(
					testCase.newState,
					testCase.update,
					"",
				),
			)
		})
	}
}
