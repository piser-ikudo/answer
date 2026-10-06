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

package repo_test

import (
	"context"
	"fmt"
	"log"
	"sync"
	"testing"

	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/repo/tag"
	"github.com/apache/answer/internal/repo/tag_common"
	"github.com/apache/answer/internal/repo/unique"
	"github.com/apache/answer/pkg/converter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	tagOnce     sync.Once
	testTagList = []*entity.Tag{
		{
			SlugName:     "go",
			DisplayName:  "Golang",
			OriginalText: "golang",
			ParsedText:   "<p>golang</p>",
			Status:       entity.TagStatusAvailable,
		},
		{
			SlugName:     "js",
			DisplayName:  "javascript",
			OriginalText: "javascript",
			ParsedText:   "<p>javascript</p>",
			Status:       entity.TagStatusAvailable,
		},
		{
			SlugName:     "go2",
			DisplayName:  "Golang2",
			OriginalText: "golang2",
			ParsedText:   "<p>golang2</p>",
			Status:       entity.TagStatusAvailable,
		},
	}
)

func addTagList() {
	uniqueIDRepo := unique.NewUniqueIDRepo(testDataSource)
	tagCommonRepo := tag_common.NewTagCommonRepo(testDataSource, uniqueIDRepo)
	err := tagCommonRepo.AddTagList(context.TODO(), testTagList)
	if err != nil {
		log.Fatalf("%+v", err)
	}
}

func Test_tagRepo_GetTagByID(t *testing.T) {
	tagOnce.Do(addTagList)
	tagCommonRepo := tag_common.NewTagCommonRepo(testDataSource, unique.NewUniqueIDRepo(testDataSource))

	gotTag, exist, err := tagCommonRepo.GetTagByID(context.TODO(), testTagList[0].ID, true)
	require.NoError(t, err)
	assert.True(t, exist)
	assert.Equal(t, testTagList[0].SlugName, gotTag.SlugName)
}

func Test_tagRepo_GetTagBySlugName(t *testing.T) {
	tagOnce.Do(addTagList)
	tagCommonRepo := tag_common.NewTagCommonRepo(testDataSource, unique.NewUniqueIDRepo(testDataSource))

	gotTag, exist, err := tagCommonRepo.GetTagBySlugName(context.TODO(), testTagList[0].SlugName)
	require.NoError(t, err)
	assert.True(t, exist)
	assert.Equal(t, testTagList[0].SlugName, gotTag.SlugName)
}

func Test_tagRepo_GetTagList(t *testing.T) {
	tagOnce.Do(addTagList)
	tagRepo := tag.NewTagRepo(testDataSource, unique.NewUniqueIDRepo(testDataSource))

	gotTags, err := tagRepo.GetTagList(context.TODO(), &entity.Tag{ID: testTagList[0].ID})
	require.NoError(t, err)
	assert.Equal(t, testTagList[0].SlugName, gotTags[0].SlugName)
}

func Test_tagRepo_GetTagListByIDs(t *testing.T) {
	tagOnce.Do(addTagList)
	tagCommonRepo := tag_common.NewTagCommonRepo(testDataSource, unique.NewUniqueIDRepo(testDataSource))

	gotTags, err := tagCommonRepo.GetTagListByIDs(context.TODO(), []string{testTagList[0].ID})
	require.NoError(t, err)
	assert.Equal(t, testTagList[0].SlugName, gotTags[0].SlugName)
}

func Test_tagRepo_GetTagListByName(t *testing.T) {
	tagOnce.Do(addTagList)
	tagCommonRepo := tag_common.NewTagCommonRepo(testDataSource, unique.NewUniqueIDRepo(testDataSource))

	gotTags, err := tagCommonRepo.GetTagListByName(context.TODO(), testTagList[0].SlugName, false, false)
	require.NoError(t, err)
	assert.Equal(t, testTagList[0].SlugName, gotTags[0].SlugName)
}

func Test_tagRepo_GetTagListByNames(t *testing.T) {
	tagOnce.Do(addTagList)
	tagCommonRepo := tag_common.NewTagCommonRepo(testDataSource, unique.NewUniqueIDRepo(testDataSource))

	gotTags, err := tagCommonRepo.GetTagListByNames(context.TODO(), []string{testTagList[0].SlugName})
	require.NoError(t, err)
	assert.Equal(t, testTagList[0].SlugName, gotTags[0].SlugName)
}

func Test_tagRepo_GetTagPage(t *testing.T) {
	tagOnce.Do(addTagList)
	tagCommonRepo := tag_common.NewTagCommonRepo(testDataSource, unique.NewUniqueIDRepo(testDataSource))

	gotTags, _, err := tagCommonRepo.GetTagPage(context.TODO(), 1, 1, &entity.Tag{SlugName: testTagList[0].SlugName}, nil, "")
	require.NoError(t, err)
	assert.Equal(t, testTagList[0].SlugName, gotTags[0].SlugName)
}

func Test_tagRepo_GetTagGroups(t *testing.T) {
	tagCommonRepo := tag_common.NewTagCommonRepo(testDataSource, unique.NewUniqueIDRepo(testDataSource))
	ctx := context.TODO()

	// three tags in group "Golang", one in group "Database", one without a group
	groupedTags := []*entity.Tag{
		{SlugName: "gp-go-1", DisplayName: "Go 1", Status: entity.TagStatusAvailable, TagG: "Golang"},
		{SlugName: "gp-go-2", DisplayName: "Go 2", Status: entity.TagStatusAvailable, TagG: "Golang"},
		{SlugName: "gp-go-3", DisplayName: "Go 3", Status: entity.TagStatusAvailable, TagG: "Golang"},
		{SlugName: "gp-db-1", DisplayName: "DB 1", Status: entity.TagStatusAvailable, TagG: "Database"},
	}
	// the empty group is only used to check that ungrouped tags are still listed
	ungroupedTag := &entity.Tag{
		SlugName: "gp-none-1", DisplayName: "None 1", Status: entity.TagStatusAvailable, TagG: "",
	}
	require.NoError(t, tagCommonRepo.AddTagList(ctx, groupedTags))
	require.NoError(t, tagCommonRepo.AddTagList(ctx, []*entity.Tag{ungroupedTag}))

	// only the available relation above may be counted, the deleted one and the
	// tags without any relation must be reported as 0
	_, err := testDataSource.DB.Context(ctx).Insert(&entity.TagRel{
		ObjectID: "900001", TagID: groupedTags[0].ID, Status: entity.TagRelStatusAvailable,
	})
	require.NoError(t, err)
	// a deleted relation must not be counted
	_, err = testDataSource.DB.Context(ctx).Insert(&entity.TagRel{
		ObjectID: "900002", TagID: groupedTags[1].ID, Status: entity.TagRelStatusDeleted,
	})
	require.NoError(t, err)

	groups, err := tagCommonRepo.GetTagGroups(ctx)
	require.NoError(t, err)

	// The test database is shared with the other tests, so only the groups
	// created here are asserted and the whole result is used for ordering.
	counts := make(map[string]int64, len(groups))
	foundGroups := make(map[string]bool, len(groups))
	for _, group := range groups {
		counts[group.TagGroup] = group.TagCount
		foundGroups[group.TagGroup] = true
	}
	require.True(t, foundGroups["Golang"])
	require.True(t, foundGroups["Database"])
	require.True(t, foundGroups[""])
	// only the available relation counts, the deleted one and the tags without
	// any relation must be reported as 0
	assert.Equal(t, int64(1), counts["Golang"])
	assert.Equal(t, int64(0), counts["Database"])
	// ordered by tag count descending
	for i := 1; i < len(groups); i++ {
		assert.GreaterOrEqual(t, groups[i-1].TagCount, groups[i].TagCount)
	}
}

func Test_tagRepo_GetTagPage_FilterByTagGroup(t *testing.T) {
	tagCommonRepo := tag_common.NewTagCommonRepo(testDataSource, unique.NewUniqueIDRepo(testDataSource))
	ctx := context.TODO()

	filterTags := []*entity.Tag{
		{SlugName: "flt-a", DisplayName: "Filter A", Status: entity.TagStatusAvailable, TagG: "GroupA"},
		{SlugName: "flt-b", DisplayName: "Filter B", Status: entity.TagStatusAvailable, TagG: "GroupB"},
		{SlugName: "flt-c", DisplayName: "Filter C", Status: entity.TagStatusAvailable, TagG: "GroupC"},
	}
	require.NoError(t, tagCommonRepo.AddTagList(ctx, filterTags))

	gotTags, total, err := tagCommonRepo.GetTagPage(ctx, 1, 20, &entity.Tag{}, []string{"GroupA"}, "name")
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, gotTags, 1)
	assert.Equal(t, "flt-a", gotTags[0].SlugName)
	assert.Equal(t, "GroupA", gotTags[0].TagG)

	// several groups are a union: every tag of any selected group is returned
	gotTags, total, err = tagCommonRepo.GetTagPage(ctx, 1, 20, &entity.Tag{}, []string{"GroupA", "GroupC"}, "name")
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	gotSlugs := make([]string, 0, len(gotTags))
	for _, item := range gotTags {
		gotSlugs = append(gotSlugs, item.SlugName)
	}
	assert.ElementsMatch(t, []string{"flt-a", "flt-c"}, gotSlugs)
}

func Test_tagRepo_GetTagIDsByGroups(t *testing.T) {
	tagCommonRepo := tag_common.NewTagCommonRepo(testDataSource, unique.NewUniqueIDRepo(testDataSource))
	ctx := context.TODO()

	groupTags := []*entity.Tag{
		{SlugName: "gid-a1", DisplayName: "A1", Status: entity.TagStatusAvailable, TagG: "IdGroupA"},
		{SlugName: "gid-a2", DisplayName: "A2", Status: entity.TagStatusAvailable, TagG: "IdGroupA"},
		{SlugName: "gid-b1", DisplayName: "B1", Status: entity.TagStatusAvailable, TagG: "IdGroupB"},
		{SlugName: "gid-c1", DisplayName: "C1", Status: entity.TagStatusDeleted, TagG: "IdGroupA"},
	}
	require.NoError(t, tagCommonRepo.AddTagList(ctx, groupTags))

	gotIDs, err := tagCommonRepo.GetTagIDsByGroups(ctx, []string{"IdGroupA", "IdGroupB"})
	require.NoError(t, err)

	idSet := make(map[string]bool, len(gotIDs))
	for _, id := range gotIDs {
		idSet[id] = true
	}
	// both tags of group A and the single tag of group B are returned (union)
	assert.True(t, idSet[groupTags[0].ID])
	assert.True(t, idSet[groupTags[1].ID])
	assert.True(t, idSet[groupTags[2].ID])
	// the deleted tag of group A and any tag of another group are not returned
	assert.False(t, idSet[groupTags[3].ID])
	assert.Len(t, gotIDs, 3)

	// an unknown group yields nothing instead of every tag
	gotIDs, err = tagCommonRepo.GetTagIDsByGroups(ctx, []string{"NoSuchGroup"})
	require.NoError(t, err)
	assert.Empty(t, gotIDs)

	// no group selected means no filtering
	gotIDs, err = tagCommonRepo.GetTagIDsByGroups(ctx, nil)
	require.NoError(t, err)
	assert.Empty(t, gotIDs)
}

func Test_tagRepo_RemoveTag(t *testing.T) {
	tagOnce.Do(addTagList)
	uniqueIDRepo := unique.NewUniqueIDRepo(testDataSource)
	tagRepo := tag.NewTagRepo(testDataSource, uniqueIDRepo)
	err := tagRepo.RemoveTag(context.TODO(), testTagList[1].ID)
	require.NoError(t, err)

	tagCommonRepo := tag_common.NewTagCommonRepo(testDataSource, unique.NewUniqueIDRepo(testDataSource))

	_, exist, err := tagCommonRepo.GetTagBySlugName(context.TODO(), testTagList[1].SlugName)
	require.NoError(t, err)
	assert.False(t, exist)
}

func Test_tagRepo_UpdateTag(t *testing.T) {
	uniqueIDRepo := unique.NewUniqueIDRepo(testDataSource)
	tagRepo := tag.NewTagRepo(testDataSource, uniqueIDRepo)

	testTagList[0].DisplayName = "golang"
	err := tagRepo.UpdateTag(context.TODO(), testTagList[0])
	require.NoError(t, err)

	tagCommonRepo := tag_common.NewTagCommonRepo(testDataSource, unique.NewUniqueIDRepo(testDataSource))

	gotTag, exist, err := tagCommonRepo.GetTagByID(context.TODO(), testTagList[0].ID, true)
	require.NoError(t, err)
	assert.True(t, exist)
	assert.Equal(t, testTagList[0].DisplayName, gotTag.DisplayName)
}

func Test_tagRepo_UpdateTagQuestionCount(t *testing.T) {
	tagCommonRepo := tag_common.NewTagCommonRepo(testDataSource, unique.NewUniqueIDRepo(testDataSource))

	testTagList[0].DisplayName = "golang"
	err := tagCommonRepo.UpdateTagQuestionCount(context.TODO(), testTagList[0].ID, 100)
	require.NoError(t, err)

	gotTag, exist, err := tagCommonRepo.GetTagByID(context.TODO(), testTagList[0].ID, true)
	require.NoError(t, err)
	assert.True(t, exist)
	assert.Equal(t, 100, gotTag.QuestionCount)
}

func Test_tagRepo_UpdateTagSynonym(t *testing.T) {
	uniqueIDRepo := unique.NewUniqueIDRepo(testDataSource)
	tagRepo := tag.NewTagRepo(testDataSource, uniqueIDRepo)

	testTagList[0].DisplayName = "golang"
	err := tagRepo.UpdateTag(context.TODO(), testTagList[0])
	require.NoError(t, err)

	err = tagRepo.UpdateTagSynonym(context.TODO(), []string{testTagList[2].SlugName},
		converter.StringToInt64(testTagList[0].ID), testTagList[0].SlugName)
	require.NoError(t, err)

	tagCommonRepo := tag_common.NewTagCommonRepo(testDataSource, unique.NewUniqueIDRepo(testDataSource))

	gotTag, exist, err := tagCommonRepo.GetTagByID(context.TODO(), testTagList[2].ID, true)
	require.NoError(t, err)
	assert.True(t, exist)
	assert.Equal(t, testTagList[0].ID, fmt.Sprintf("%d", gotTag.MainTagID))
}
