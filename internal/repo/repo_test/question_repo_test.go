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
	"testing"

	"github.com/apache/answer/internal/entity"
	"github.com/apache/answer/internal/repo/question"
	"github.com/apache/answer/internal/repo/revision"
	"github.com/apache/answer/internal/repo/site_info"
	"github.com/apache/answer/internal/repo/tag"
	"github.com/apache/answer/internal/repo/tag_common"
	"github.com/apache/answer/internal/repo/unique"
	"github.com/apache/answer/internal/repo/user"
	"github.com/apache/answer/internal/schema"
	"github.com/apache/answer/internal/service/activityqueue"
	questioncommon "github.com/apache/answer/internal/service/question_common"
	revisionservice "github.com/apache/answer/internal/service/revision_common"
	"github.com/apache/answer/internal/service/siteinfo_common"
	tagcommonservice "github.com/apache/answer/internal/service/tag_common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newQuestionRepo() questioncommon.QuestionRepo {
	return question.NewQuestionRepo(testDataSource, unique.NewUniqueIDRepo(testDataSource))
}

// newTagCommonService wires the real tag service used by the question write path.
func newTagCommonService() *tagcommonservice.TagCommonService {
	uniqueIDRepo := unique.NewUniqueIDRepo(testDataSource)
	return tagcommonservice.NewTagCommonService(
		tag_common.NewTagCommonRepo(testDataSource, uniqueIDRepo),
		tag.NewTagRelRepo(testDataSource, uniqueIDRepo),
		tag.NewTagRepo(testDataSource, uniqueIDRepo),
		revisionservice.NewRevisionService(
			revision.NewRevisionRepo(testDataSource, uniqueIDRepo),
			user.NewUserRepo(testDataSource),
		),
		siteinfo_common.NewSiteInfoCommonService(site_info.NewSiteInfo(testDataSource)),
		activityqueue.NewService(),
	)
}

// newQuestionWithTag inserts a question tagged with the given tag, mimicking the
// question + tag_rel rows the real write path creates.
func newQuestionWithTag(t *testing.T, ctx context.Context, title string, tagIDs []string) *entity.Question {
	t.Helper()
	questionRepo := newQuestionRepo()
	q := &entity.Question{
		Title:        title,
		OriginalText: title,
		ParsedText:   title,
		Status:       entity.QuestionStatusAvailable,
		Pin:          entity.QuestionUnPin,
		Show:         entity.QuestionShow,
		UserID:       "1",
		RevisionID:   "0",
	}
	require.NoError(t, questionRepo.AddQuestion(ctx, q))
	for _, tagID := range tagIDs {
		_, err := testDataSource.DB.Context(ctx).Insert(&entity.TagRel{
			ObjectID: q.ID, TagID: tagID, Status: entity.TagRelStatusAvailable,
		})
		require.NoError(t, err)
	}
	return q
}

// Test_questionRepo_GetQuestionPage_FilterByTagGroup covers the query behind
// "select tag groups and get every post tagged with a tag of those groups".
func Test_questionRepo_GetQuestionPage_FilterByTagGroup(t *testing.T) {
	ctx := context.TODO()
	tagCommonRepo := tag_common.NewTagCommonRepo(testDataSource, unique.NewUniqueIDRepo(testDataSource))
	questionRepo := newQuestionRepo()

	// two groups, one tag each. The slug prefix must not collide with the slugs
	// used by the other tests, because the tag page search matches by slug prefix.
	groupTags := []*entity.Tag{
		{SlugName: "grp-alpha", DisplayName: "Grp Alpha", Status: entity.TagStatusAvailable, TagG: "QGroupGo"},
		{SlugName: "grp-beta", DisplayName: "Grp Beta", Status: entity.TagStatusAvailable, TagG: "QGroupDB"},
		{SlugName: "grp-gamma", DisplayName: "Grp Gamma", Status: entity.TagStatusAvailable, TagG: "QGroupOther"},
	}
	require.NoError(t, tagCommonRepo.AddTagList(ctx, groupTags))

	goQuestion := newQuestionWithTag(t, ctx, "question in go group", []string{groupTags[0].ID})
	dbQuestion := newQuestionWithTag(t, ctx, "question in db group", []string{groupTags[1].ID})
	otherQuestion := newQuestionWithTag(t, ctx, "question in other group", []string{groupTags[2].ID})

	// resolve groups to tag ids, exactly like the question service does
	goAndDBTagIDs, err := tagCommonRepo.GetTagIDsByGroups(ctx, []string{"QGroupGo", "QGroupDB"})
	require.NoError(t, err)
	require.Len(t, goAndDBTagIDs, 2)

	// OR semantics: both selected groups are returned in one page
	questions, total, err := questionRepo.GetQuestionPage(ctx, 1, 20, goAndDBTagIDs, "", "newest", 0, false, false)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	returnedIDs := make(map[string]bool, len(questions))
	for _, q := range questions {
		returnedIDs[q.ID] = true
	}
	assert.True(t, returnedIDs[goQuestion.ID], "the go group question must be returned")
	assert.True(t, returnedIDs[dbQuestion.ID], "the db group question must be returned")
	assert.False(t, returnedIDs[otherQuestion.ID], "a question of an unselected group must not be returned")

	// a single group narrows the result down
	onlyDBTagIDs, err := tagCommonRepo.GetTagIDsByGroups(ctx, []string{"QGroupDB"})
	require.NoError(t, err)
	questions, total, err = questionRepo.GetQuestionPage(ctx, 1, 20, onlyDBTagIDs, "", "newest", 0, false, false)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, questions, 1)
	assert.Equal(t, dbQuestion.ID, questions[0].ID)

	// an unknown group resolves to no tag, so no question is returned
	noTagIDs, err := tagCommonRepo.GetTagIDsByGroups(ctx, []string{"QGroupMissing"})
	require.NoError(t, err)
	assert.Empty(t, noTagIDs)
}

// Test_ObjectChangeTag_TagGroupOfNewAndExistingTags covers the rule that a tag
// already stored keeps its group when it is picked while posting a question,
// while a tag created on that screen falls into the default group.
func Test_ObjectChangeTag_TagGroupOfNewAndExistingTags(t *testing.T) {
	ctx := context.TODO()
	uniqueIDRepo := unique.NewUniqueIDRepo(testDataSource)
	tagCommonRepo := tag_common.NewTagCommonRepo(testDataSource, uniqueIDRepo)
	tagCommonService := newTagCommonService()

	// a tag that already exists with an explicit group
	existing := []*entity.Tag{{
		SlugName: "keep-group-tag", DisplayName: "Keep Group", Status: entity.TagStatusAvailable, TagG: "CustomGroup",
	}}
	require.NoError(t, tagCommonRepo.AddTagList(ctx, existing))

	question := newQuestionWithTag(t, ctx, "question grouping tags", nil)

	// post the question with one existing tag and one brand new tag
	_, err := tagCommonService.ObjectChangeTag(ctx, &schema.TagChange{
		ObjectID: question.ID,
		UserID:   "1",
		Tags: []*schema.TagItem{
			{SlugName: "keep-group-tag", DisplayName: "Keep Group"},
			{SlugName: "brand-new-tag", DisplayName: "Brand New"},
		},
	}, 0)
	require.NoError(t, err)

	// the existing tag must keep the group it was created with
	stored, exist, err := tagCommonRepo.GetTagBySlugName(ctx, "keep-group-tag")
	require.NoError(t, err)
	require.True(t, exist)
	assert.Equal(t, "CustomGroup", stored.TagG, "an existing tag must keep its own group")

	// the tag created while posting goes to the default group
	created, exist, err := tagCommonRepo.GetTagBySlugName(ctx, "brand-new-tag")
	require.NoError(t, err)
	require.True(t, exist)
	assert.Equal(t, entity.DefaultTagGroup, created.TagG, "a newly created tag must use the default group")
	assert.Equal(t, "default", created.TagG)

	// and both are linked to the question
	relList := make([]*entity.TagRel, 0)
	require.NoError(t, testDataSource.DB.Context(ctx).
		Where("object_id = ?", question.ID).Find(&relList))
	assert.Len(t, relList, 2)
}
