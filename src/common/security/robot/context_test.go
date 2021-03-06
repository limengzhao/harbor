// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package robot

import (
	"fmt"
	"os"
	"testing"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/utils/test"
	"github.com/goharbor/harbor/src/core/promgr"
	"github.com/goharbor/harbor/src/core/promgr/pmsdriver/local"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	"github.com/goharbor/harbor/src/pkg/robot/model"
	"github.com/stretchr/testify/assert"
)

var (
	private = &models.Project{
		Name:    "testrobot",
		OwnerID: 1,
	}
	pm = promgr.NewDefaultProjectManager(local.NewDriver(), true)
)

func TestMain(m *testing.M) {
	test.InitDatabaseFromEnv()

	// add project
	id, err := dao.AddProject(*private)
	if err != nil {
		log.Fatalf("failed to add project: %v", err)
	}
	private.ProjectID = id
	defer dao.DeleteProject(id)

	os.Exit(m.Run())
}

func TestIsAuthenticated(t *testing.T) {
	// unauthenticated
	ctx := NewSecurityContext(nil, nil, nil)
	assert.False(t, ctx.IsAuthenticated())

	// authenticated
	ctx = NewSecurityContext(&model.Robot{
		Name:     "test",
		Disabled: false,
	}, nil, nil)
	assert.True(t, ctx.IsAuthenticated())
}

func TestGetUsername(t *testing.T) {
	// unauthenticated
	ctx := NewSecurityContext(nil, nil, nil)
	assert.Equal(t, "", ctx.GetUsername())

	// authenticated
	ctx = NewSecurityContext(&model.Robot{
		Name:     "test",
		Disabled: false,
	}, nil, nil)
	assert.Equal(t, "test", ctx.GetUsername())
}

func TestIsSysAdmin(t *testing.T) {
	// unauthenticated
	ctx := NewSecurityContext(nil, nil, nil)
	assert.False(t, ctx.IsSysAdmin())

	// authenticated, non admin
	ctx = NewSecurityContext(&model.Robot{
		Name:     "test",
		Disabled: false,
	}, nil, nil)
	assert.False(t, ctx.IsSysAdmin())
}

func TestIsSolutionUser(t *testing.T) {
	ctx := NewSecurityContext(nil, nil, nil)
	assert.False(t, ctx.IsSolutionUser())
}

func TestHasPullPerm(t *testing.T) {
	policies := []*types.Policy{
		{
			Resource: rbac.Resource(fmt.Sprintf("/project/%d/repository", private.ProjectID)),
			Action:   rbac.ActionPull,
		},
	}
	robot := &model.Robot{
		Name:        "test_robot_1",
		Description: "desc",
	}

	ctx := NewSecurityContext(robot, pm, policies)
	resource := rbac.NewProjectNamespace(private.ProjectID).Resource(rbac.ResourceRepository)
	assert.True(t, ctx.Can(rbac.ActionPull, resource))
}

func TestHasPushPerm(t *testing.T) {
	policies := []*types.Policy{
		{
			Resource: rbac.Resource(fmt.Sprintf("/project/%d/repository", private.ProjectID)),
			Action:   rbac.ActionPush,
		},
	}
	robot := &model.Robot{
		Name:        "test_robot_2",
		Description: "desc",
	}

	ctx := NewSecurityContext(robot, pm, policies)
	resource := rbac.NewProjectNamespace(private.ProjectID).Resource(rbac.ResourceRepository)
	assert.True(t, ctx.Can(rbac.ActionPush, resource))
}

func TestHasPushPullPerm(t *testing.T) {
	policies := []*types.Policy{
		{
			Resource: rbac.Resource(fmt.Sprintf("/project/%d/repository", private.ProjectID)),
			Action:   rbac.ActionPush,
		},
		{
			Resource: rbac.Resource(fmt.Sprintf("/project/%d/repository", private.ProjectID)),
			Action:   rbac.ActionPull,
		},
	}
	robot := &model.Robot{
		Name:        "test_robot_3",
		Description: "desc",
	}

	ctx := NewSecurityContext(robot, pm, policies)
	resource := rbac.NewProjectNamespace(private.ProjectID).Resource(rbac.ResourceRepository)
	assert.True(t, ctx.Can(rbac.ActionPush, resource) && ctx.Can(rbac.ActionPull, resource))
}
