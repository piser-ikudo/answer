/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package schema

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReplaceSearchContent(t *testing.T) {
	content := "user:aaa [tag] ssssfdfdf-as#fsadf"
	replacedContent, patterns := ReplaceSearchContent(content)
	ret := strings.Join(append(patterns, replacedContent), " ")

	assert.Equal(t, "user:aaa [tag] ssssfdfdf as fsadf", ret)

	content = "user:aaa-sss [tag1] ssssfdfdf-as#fsadf [tag2] score:3"
	replacedContent, patterns = ReplaceSearchContent(content)
	ret = strings.Join(append(patterns, replacedContent), " ")

	assert.Equal(t, "user:aaa-sss score:3 [tag1] [tag2] ssssfdfdf as fsadf", ret)
}

// TestSearchDTO_Check covers the rule that allows searching by tag groups only,
// while still refusing a request without any search condition.
func TestSearchDTO_Check(t *testing.T) {
	cases := []struct {
		name      string
		query     string
		tagGroups []string
		wantError bool
	}{
		{
			name:      "keyword only is allowed",
			query:     "golang",
			wantError: false,
		},
		{
			name:      "tag groups with an empty keyword are allowed",
			query:     "",
			tagGroups: []string{"Go", "Database"},
			wantError: false,
		},
		{
			name:      "tag groups with a blank keyword are allowed",
			query:     "   ",
			tagGroups: []string{"Go"},
			wantError: false,
		},
		{
			name:      "neither keyword nor tag group is refused",
			query:     "",
			wantError: true,
		},
		{
			name:      "only blank keyword is refused",
			query:     "   ",
			wantError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dto := &SearchDTO{Query: tc.query, TagGroups: tc.tagGroups}
			errFields, err := dto.Check()
			assert.NoError(t, err)
			if tc.wantError {
				assert.NotEmpty(t, errFields)
				assert.Equal(t, "q", errFields[0].ErrorField)
				return
			}
			assert.Empty(t, errFields)
		})
	}
}
