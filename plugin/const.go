// Copyright  2021 Splunk Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gitlabtoken

import "github.com/xanzy/go-gitlab"

type PAT = gitlab.ProjectAccessToken

const (
	pathPatternConfig = "config"
	pathPatternToken  = "token"
	pathPatternRoles  = "roles"

	// accessLevelGuest      = 10
	// accessLevelReporter   = 20
	// accessLevelDeveloper  = 30
	// accessLevelMaintainer = 40
)
