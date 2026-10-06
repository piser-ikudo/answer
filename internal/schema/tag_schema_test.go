package schema

import (
	"testing"

	"github.com/apache/answer/internal/base/validator"
	"github.com/stretchr/testify/assert"
)

// TagItem is reused to reference existing tags, so a question must be creatable
// with tags that carry no tag group. Requiring the group here used to reject
// every question submission with "tag_group is a required field", which the UI
// then dropped silently because the ask form has no tag_group field.
func Test_QuestionAdd_TagsWithoutTagGroupAreAccepted(t *testing.T) {
	req := &QuestionAdd{
		Title:   "a valid title here",
		Content: "some content",
		Tags: []*TagItem{
			{SlugName: "go", DisplayName: "Go"},
		},
	}
	fields, err := validator.GetValidatorByLang("en_US").Check(req)
	assert.NoError(t, err)
	assert.Empty(t, fields, "posting a question with existing tags must not require a tag group")
}

// Creating a tag still requires a group, that rule lives on AddTagReq.
func Test_AddTagReq_RequiresTagGroup(t *testing.T) {
	req := &AddTagReq{
		SlugName:     "go",
		DisplayName:  "Go",
		OriginalText: "golang",
	}
	fields, err := validator.GetValidatorByLang("en_US").Check(req)
	assert.Error(t, err)
	assert.NotEmpty(t, fields, "creating a tag without a group must be rejected")
}
