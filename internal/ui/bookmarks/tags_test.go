package bookmarks

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/idursun/jjui/internal/jj"
	"github.com/idursun/jjui/internal/ui/intents"
	"github.com/idursun/jjui/internal/ui/layout"
	"github.com/idursun/jjui/internal/ui/render"
	"github.com/idursun/jjui/test"
	"github.com/stretchr/testify/assert"
)

// v1.0 is on the target and tracked on origin, v2.0 is local elsewhere,
// v3.0 is conflicted on the target, v4.0 exists only on upstream and
// v5.0 is local with no remotes.
const tagListOutput = `
v1.0;.;true;false;false;false;commit123
v1.0;origin;true;true;false;false;commit123
v2.0;.;true;false;false;false;nearby
v3.0;.;true;false;true;false;commit123
v4.0;upstream;true;false;false;false;far
v5.0;.;true;false;false;false;outside
`

func newTagMenu(t *testing.T, commandRunner *test.CommandRunner) *Model {
	t.Helper()
	commit := &jj.Commit{ChangeId: "abc123", CommitId: "commit123"}
	op := NewTagModel(test.NewTestContext(commandRunner), commit, []string{"commit123", "nearby", "far"})
	test.SimulateModel(op, op.Init())
	_ = test.RenderImmediate(op, 100, 40)
	return op
}

func itemsFor(op *Model, kind intents.BookmarksFilterKind) []item {
	op.filtered(string(kind))
	return op.visibleItems()
}

func itemNames(items []item) []string {
	var names []string
	for _, i := range items {
		names = append(names, i.name)
	}
	return names
}

func Test_TagMenu_Choosers(t *testing.T) {
	commandRunner := test.NewTestCommandRunner(t)
	commandRunner.Expect(jj.GitRemoteList()).SetOutput([]byte("origin\nupstream"))
	commandRunner.Expect(jj.TagListAll()).SetOutput([]byte(tagListOutput))
	defer commandRunner.Verify()

	op := newTagMenu(t, commandRunner)

	move := itemsFor(op, intents.BookmarksFilterMove)
	assert.Equal(t, []string{"move conflicted 'v3.0' to abc123", "move 'v2.0' to abc123", "move 'v5.0' to abc123"}, itemNames(move))
	for _, i := range move {
		assert.Empty(t, i.key, "tag move items have no shortcut")
	}
	assert.Equal(t, []string(jj.TagSet("abc123", "v3.0")), move[0].args)

	delete := itemsFor(op, intents.BookmarksFilterDelete)
	assert.Equal(t, []string{"delete 'v1.0'", "delete 'v3.0'", "delete 'v2.0'", "delete 'v5.0'"}, itemNames(delete))
	assert.Equal(t, []string(jj.TagDelete("v1.0")), delete[0].args)

	// Local mode lists track-to-default-remote and every untrack item.
	assert.Equal(t, []string{"track 'v3.0'", "track 'v2.0'", "track 'v5.0'"}, itemNames(itemsFor(op, intents.BookmarksFilterTrack)))
	untrack := itemsFor(op, intents.BookmarksFilterUntrack)
	assert.Equal(t, []string{"untrack 'v1.0@origin'"}, itemNames(untrack))
	assert.Equal(t, []string(jj.RefUntrack("tag", "v1.0", "origin")), untrack[0].args)

	// Cycling to a remote narrows track items to that remote.
	op.cycleRemotes(2)
	track := itemsFor(op, intents.BookmarksFilterTrack)
	assert.Equal(t, []string{"track 'v4.0@upstream'"}, itemNames(track))
	assert.Equal(t, []string(jj.RefTrack("tag", "v4.0", "upstream")), track[0].args)

	for _, i := range op.allItems {
		assert.False(t, strings.HasPrefix(i.name, "forget"), "tags have no forget: %s", i.name)
	}
}

func Test_TagMenu_DoublePress_RunsNearestItem(t *testing.T) {
	tests := []struct {
		kind intents.BookmarksFilterKind
		want jj.CommandArgs
	}{
		{intents.BookmarksFilterMove, jj.TagSet("abc123", "v3.0")},
		{intents.BookmarksFilterDelete, jj.TagDelete("v1.0")},
	}
	for _, tt := range tests {
		t.Run(string(tt.kind), func(t *testing.T) {
			commandRunner := test.NewTestCommandRunner(t)
			commandRunner.Expect(jj.GitRemoteList()).SetOutput([]byte("origin"))
			commandRunner.Expect(jj.TagListAll()).SetOutput([]byte(tagListOutput))
			commandRunner.Expect(tt.want)
			defer commandRunner.Verify()

			op := newTagMenu(t, commandRunner)
			test.SimulateModel(op, func() tea.Msg { return intents.BookmarksFilter{Kind: tt.kind} })
			test.SimulateModel(op, func() tea.Msg { return intents.BookmarksFilter{Kind: tt.kind} })
		})
	}
}

func Test_TagMenu_UsesTagsScopeAndTheme(t *testing.T) {
	commandRunner := test.NewTestCommandRunner(t)
	commandRunner.Expect(jj.GitRemoteList()).SetOutput([]byte(""))

	op := NewTagModel(test.NewTestContext(commandRunner), &jj.Commit{ChangeId: "abc123", CommitId: "commit123"}, nil)
	assert.Equal(t, "tags", string(op.Scopes()[0].Name))
	assert.Equal(t, "tags", op.kind.themeKey)

	test.SimulateModel(op, func() tea.Msg { return intents.BookmarksOpenFilter{} })
	assert.Equal(t, "tags.filter", string(op.Scopes()[0].Name))
}

func Test_TagMenu_FilterEditing_AcceptsPasteMsg(t *testing.T) {
	commandRunner := test.NewTestCommandRunner(t)
	commandRunner.Expect(jj.GitRemoteList()).SetOutput([]byte(""))
	defer commandRunner.Verify()

	op := NewTagModel(test.NewTestContext(commandRunner), &jj.Commit{ChangeId: "abc123", CommitId: "commit123"}, nil)
	test.SimulateModel(op, func() tea.Msg { return intents.BookmarksOpenFilter{} })
	test.SimulateModel(op, func() tea.Msg { return tea.PasteMsg{Content: "track v1"} })

	assert.Equal(t, "track v1", op.filterInput.Value())
}

func TestTags_ZIndex_RendersAboveMainContent(t *testing.T) {
	commandRunner := test.NewTestCommandRunner(t)
	commandRunner.Expect(jj.GitRemoteList()).SetOutput([]byte("origin"))
	commandRunner.Expect(jj.TagListAll()).SetOutput([]byte(""))

	op := NewTagModel(test.NewTestContext(commandRunner), &jj.Commit{ChangeId: "abc123", CommitId: "commit123"}, []string{"commit123"})
	test.SimulateModel(op, op.Init())

	dl := render.NewDisplayContext()
	box := layout.Box{R: layout.Rect(0, 0, 100, 40)}
	dl.AddDraw(box.R, strings.Repeat("x", box.R.Dx()*box.R.Dy()), render.ZBase)
	op.ViewRect(dl, box)

	rendered := dl.RenderToString(box.R.Dx(), box.R.Dy())
	assert.Contains(t, rendered, "Tag Operations")
	assert.Contains(t, rendered, "Remotes:")
}
