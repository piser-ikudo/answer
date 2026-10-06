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

package migrations

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/apache/answer/internal/base/data"
	"github.com/apache/answer/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

// legacyTag mirrors the tag table as it looked before the tag group column was
// introduced, so the upgrade path of addTagGroup can be tested. Every column of
// entity.Tag is declared except TagG.
type legacyTag struct {
	ID              string `xorm:"not null pk comment('tag_id') BIGINT(20) id"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	MainTagID       int64  `xorm:"not null default 0 BIGINT(20) main_tag_id"`
	MainTagSlugName string `xorm:"not null default '' VARCHAR(35) main_tag_slug_name"`
	SlugName        string `xorm:"not null default '' unique VARCHAR(35) slug_name"`
	DisplayName     string `xorm:"not null default '' VARCHAR(35) display_name"`
	OriginalText    string `xorm:"not null MEDIUMTEXT original_text"`
	ParsedText      string `xorm:"not null MEDIUMTEXT parsed_text"`
	FollowCount     int    `xorm:"not null default 0 INT(11) follow_count"`
	QuestionCount   int    `xorm:"not null default 0 INT(11) question_count"`
	Status          int    `xorm:"not null default 1 INT(11) status"`
	Recommend       bool   `xorm:"not null default false BOOL recommend"`
	Reserved        bool   `xorm:"not null default false BOOL reserved"`
	RevisionID      string `xorm:"not null default 0 BIGINT(20) revision_id"`
	UserID          string `xorm:"not null default 0 BIGINT(20) user_id"`
}

func (legacyTag) TableName() string {
	return "tag"
}

func newTestEngine(t *testing.T) *xorm.Engine {
	t.Helper()
	dbFile := filepath.Join(t.TempDir(), "migration-test.db")
	engine, err := data.NewDB(false, &data.Database{Driver: "sqlite3", Connection: dbFile})
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = engine.Close()
		_ = os.Remove(dbFile)
	})
	return engine
}

func Test_addTagGroup_AddsColumnAndBackfillsLegacyRows(t *testing.T) {
	ctx := context.TODO()
	engine := newTestEngine(t)

	// 1. a database created before the tag group feature
	require.NoError(t, engine.Context(ctx).Sync(new(legacyTag)))
	_, err := engine.Context(ctx).Insert(&legacyTag{ID: "1", SlugName: "legacy", DisplayName: "Legacy", Status: 1})
	require.NoError(t, err)

	// Before the upgrade the tag group column does not exist, so reading the
	// table through entity.Tag (which requires that column) must fail.
	found, err := engine.Context(ctx).Get(&entity.Tag{ID: "1"})
	require.Error(t, err, "reading the legacy table through entity.Tag must fail before the migration")
	require.False(t, found)

	// 2. run the upgrade
	require.NoError(t, addTagGroup(ctx, engine))

	// 3. the column now exists and the legacy row has been backfilled
	legacy := &entity.Tag{ID: "1"}
	found, err = engine.Context(ctx).Get(legacy)
	require.NoError(t, err)
	require.True(t, found)
	assert.Equal(t, "legacy", legacy.SlugName)
	assert.Equal(t, entity.DefaultTagGroup, legacy.TagG, "the legacy row must be backfilled with the default group")

	// 4. running it again must be safe
	require.NoError(t, addTagGroup(ctx, engine))
}

func Test_addTagGroup_SkipsWhenTagTableIsMissing(t *testing.T) {
	ctx := context.TODO()
	engine := newTestEngine(t)

	// a database that has not been installed yet must not be broken
	require.NoError(t, addTagGroup(ctx, engine))

	exist, err := engine.Context(ctx).IsTableExist(new(legacyTag))
	require.NoError(t, err)
	assert.False(t, exist, "the migration must not create the tag table")
}
