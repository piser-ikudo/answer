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
	"fmt"

	"github.com/apache/answer/internal/entity"
	"github.com/segmentfault/pacman/log"
	"xorm.io/xorm"
)

// addTagGroup adds the tag group column to the tag table. Databases created by
// this version already have the column, the migration only upgrades databases
// that were created before the tag group feature existed.
func addTagGroup(ctx context.Context, x *xorm.Engine) error {
	tagTable := new(entity.Tag).TableName()

	exist, err := x.Context(ctx).IsTableExist(tagTable)
	if err != nil {
		return fmt.Errorf("check tag table exist failed: %w", err)
	}
	// The tag table does not exist yet, the install process will create it
	// together with the tag group column, so there is nothing to upgrade.
	if !exist {
		return nil
	}

	hasColumn, err := x.Dialect().IsColumnExist(x.DB(), ctx, tagTable, "TagG")
	if err != nil {
		return fmt.Errorf("check tag group column exist failed: %w", err)
	}
	if !hasColumn {
		// ADD COLUMN with a default is supported by MySQL, PostgreSQL and
		// SQLite, and it does not touch any other column of the table.
		_, err = x.Context(ctx).Exec(fmt.Sprintf(
			"ALTER TABLE %s ADD COLUMN TagG VARCHAR(20) NOT NULL DEFAULT '%s'",
			tagTable, entity.DefaultTagGroup))
		if err != nil {
			return fmt.Errorf("add tag group column failed: %w", err)
		}
		log.Info("added tag group column to the tag table")
	}

	// Normalize legacy rows that already got the column without a value.
	result, err := x.Context(ctx).Exec(fmt.Sprintf(
		"UPDATE %s SET TagG = '%s' WHERE TagG IS NULL OR TagG = ''",
		tagTable, entity.DefaultTagGroup))
	if err != nil {
		return fmt.Errorf("backfill tag group failed: %w", err)
	}
	if affected, err := result.RowsAffected(); err == nil && affected > 0 {
		log.Infof("backfilled %d tag(s) with the default tag group", affected)
	}
	return nil
}
