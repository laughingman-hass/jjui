package bookmarks

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/idursun/jjui/internal/jj"
	"github.com/idursun/jjui/internal/ui/actions"
	"github.com/idursun/jjui/internal/ui/bindings"
)

// refKind holds everything the menu needs to know about one kind of jj ref.
type refKind struct {
	subcommand string
	title      string
	scope      bindings.ScopeName
	themeKey   string
	loaders    []func(*Model) tea.Msg
}

var bookmarkKind = refKind{
	subcommand: "bookmark",
	title:      "Bookmark Operations",
	scope:      actions.ScopeBookmarks,
	themeKey:   "bookmarks",
	loaders:    []func(*Model) tea.Msg{(*Model).loadAll, (*Model).loadMovables},
}

func (m *Model) loadMovables() tea.Msg {
	// A hidden revision has no visible descendants, so look for bookmarks
	// around the visible revision of its change instead.
	output, _ := m.context.RunCommandImmediate(jj.BookmarkListMovable(m.current.ChangeId))
	var bookmarkItems []item
	bookmarks := jj.ParseBookmarkListOutput(string(output))
	for _, b := range bookmarks {
		if b.Local == nil || !b.Local.Present {
			continue
		}
		localCommitID := b.Local.CommitId
		if !b.Conflict && localCommitID == m.current.CommitId {
			continue
		}

		name := fmt.Sprintf("move '%s' to %s", b.Name, m.current.GetChangeId())
		if b.Conflict {
			name = fmt.Sprintf("move conflicted '%s' to %s", b.Name, m.current.GetChangeId())
		}
		var extraFlags []string
		if b.Backwards {
			name = fmt.Sprintf("move '%s' backwards to %s", b.Name, m.current.GetChangeId())
			extraFlags = append(extraFlags, "--allow-backwards")
		}
		elem := item{
			name:     name,
			priority: moveCommand,
			args:     jj.BookmarkMove(m.current.GetChangeId(), b.Name, extraFlags...),
			dist:     m.distance(localCommitID),
		}
		if b.Name == "main" || b.Name == "master" {
			elem.key = "m"
		}
		bookmarkItems = append(bookmarkItems, elem)
	}
	return updateItemsMsg{items: bookmarkItems}
}

func (m *Model) loadAll() tea.Msg {
	output, err := m.context.RunCommandImmediate(jj.BookmarkListAll())
	if err != nil {
		return nil
	}

	bookmarks := jj.ParseBookmarkListOutput(string(output))
	items := make([]item, 0)
	defaultRemote := m.defaultTrackRemote()
	for _, b := range bookmarks {
		localCommitID := ""
		if b.Local != nil {
			localCommitID = b.Local.CommitId
		}
		distance := m.distance(localCommitID)
		if command, ok := jj.BookmarkDeleteCommand(b); ok {
			items = append(items, item{
				name:     fmt.Sprintf("delete '%s'", b.Name),
				priority: deleteCommand,
				dist:     distance,
				args:     command,
			})
		}

		if command, ok := jj.BookmarkForgetCommand(b); ok {
			items = append(items, item{
				name:     fmt.Sprintf("forget '%s'", b.Name),
				priority: forgetCommand,
				dist:     distance,
				args:     command,
			})
		}

		if b.IsTrackable() {
			if command, ok := jj.RefTrackLocalCommand(m.kind.subcommand, b, defaultRemote); ok {
				items = append(items, item{
					name:     fmt.Sprintf("track '%s'", b.Name),
					priority: trackCommand,
					dist:     distance,
					args:     command,
				})
			}
		}

		for _, remote := range b.Remotes {
			nameWithRemote := fmt.Sprintf("%s@%s", b.Name, remote.Remote)
			if command, ok := jj.RefUntrackRemoteCommand(m.kind.subcommand, b, remote); ok {
				items = append(items, item{
					name:     fmt.Sprintf("untrack '%s'", nameWithRemote),
					priority: untrackCommand,
					dist:     distance,
					args:     command,
				})
				continue
			}
			if command, ok := jj.RefTrackRemoteCommand(m.kind.subcommand, b, remote); ok {
				items = append(items, item{
					name:     fmt.Sprintf("track '%s'", nameWithRemote),
					priority: trackCommand,
					dist:     distance,
					args:     command,
				})
			}
		}
	}
	return updateItemsMsg{items: items}
}
