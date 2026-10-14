package jj

func BookmarkDeleteCommand(bookmark Bookmark) (CommandArgs, bool) {
	if !bookmark.IsDeletable() {
		return nil, false
	}
	return BookmarkDelete(bookmark.Name), true
}

func BookmarkForgetCommand(bookmark Bookmark) (CommandArgs, bool) {
	if bookmark.Name == "" {
		return nil, false
	}
	return BookmarkForget(bookmark.Name), true
}

func BookmarkTrackLocalCommand(bookmark Bookmark, defaultRemote string) (CommandArgs, bool) {
	return RefTrackLocalCommand("bookmark", bookmark, defaultRemote)
}

func BookmarkTrackRemoteCommand(bookmark Bookmark, remote BookmarkRemote) (CommandArgs, bool) {
	return RefTrackRemoteCommand("bookmark", bookmark, remote)
}

func BookmarkUntrackRemoteCommand(bookmark Bookmark, remote BookmarkRemote) (CommandArgs, bool) {
	return RefUntrackRemoteCommand("bookmark", bookmark, remote)
}

func RefTrackLocalCommand(subcommand string, ref Bookmark, defaultRemote string) (CommandArgs, bool) {
	if ref.Local == nil || !ref.Local.Present || defaultRemote == "" {
		return nil, false
	}
	return RefTrack(subcommand, ref.Name, defaultRemote), true
}

func RefTrackRemoteCommand(subcommand string, ref Bookmark, remote BookmarkRemote) (CommandArgs, bool) {
	if remote.Remote == "" || remote.Tracked {
		return nil, false
	}
	return RefTrack(subcommand, ref.Name, remote.Remote), true
}

func RefUntrackRemoteCommand(subcommand string, ref Bookmark, remote BookmarkRemote) (CommandArgs, bool) {
	if remote.Remote == "" || !remote.Tracked {
		return nil, false
	}
	return RefUntrack(subcommand, ref.Name, remote.Remote), true
}
