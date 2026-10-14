package bookmark

import (
	"github.com/idursun/jjui/internal/jj"
	"github.com/idursun/jjui/internal/ui/actions"
	"github.com/idursun/jjui/internal/ui/context"
)

var tagSetKind = refSetKind{
	scope:       actions.ScopeSetTag,
	name:        "set tag",
	set:         jj.TagSet,
	suggestions: tagSuggestions,
}

func tagSuggestions(runner context.CommandRunner, _ string) []string {
	output, err := runner.RunCommandImmediate(jj.TagListAll())
	if err != nil {
		return nil
	}
	var suggestions []string
	for _, t := range jj.ParseBookmarkListOutput(string(output)) {
		if t.Name != "" {
			suggestions = append(suggestions, t.Name)
		}
	}
	return suggestions
}

// NewSetTagOperation sets or moves a tag on a revision; it reuses the set-bookmark input.
func NewSetTagOperation(context *context.MainContext, changeId string, initialValue string) *SetBookmarkOperation {
	return newSetRefOperation(tagSetKind, context, changeId, initialValue)
}
