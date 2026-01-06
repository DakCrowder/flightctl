package fileio

// FileOption is a functional option for file operations.
type FileOption func(*fileOptions)

type fileOptions struct {
	uid int
	gid int
}

// WithUid sets the uid for the file.
func WithUid(uid int) FileOption {
	return func(o *fileOptions) {
		o.uid = uid
	}
}

// WithGid sets the gid for the file.
func WithGid(gid int) FileOption {
	return func(o *fileOptions) {
		o.gid = gid
	}
}

// symlinkBehavior defines how symlinks are handled during directory copy operations.
type symlinkBehavior int

const (
	symlinkSkip symlinkBehavior = iota
	symlinkError
	symlinkPreserve
	symlinkFollow
	symlinkFollowWithinRoot
	symlinkPreserveWithinRoot
)

type copyDirOptions struct {
	symlinkBehavior symlinkBehavior
	rootDir         string
}

// CopyDirOption is a functional option for CopyDir.
type CopyDirOption func(*copyDirOptions)

// WithSkipSymlink skips symlinks during directory copy.
func WithSkipSymlink() CopyDirOption {
	return func(opts *copyDirOptions) {
		opts.symlinkBehavior = symlinkSkip
	}
}

// WithErrorOnSymlink returns an error if a symlink is encountered during directory copy.
func WithErrorOnSymlink() CopyDirOption {
	return func(opts *copyDirOptions) {
		opts.symlinkBehavior = symlinkError
	}
}

// WithPreserveSymlink preserves symlinks as-is during directory copy.
func WithPreserveSymlink() CopyDirOption {
	return func(opts *copyDirOptions) {
		opts.symlinkBehavior = symlinkPreserve
	}
}

// WithFollowSymlink follows symlinks during directory copy with validation.
func WithFollowSymlink() CopyDirOption {
	return func(opts *copyDirOptions) {
		opts.symlinkBehavior = symlinkFollow
	}
}

// WithFollowSymlinkWithinRoot follows symlinks only if they resolve within the source root directory.
func WithFollowSymlinkWithinRoot() CopyDirOption {
	return func(opts *copyDirOptions) {
		opts.symlinkBehavior = symlinkFollowWithinRoot
	}
}

// WithPreserveSymlinkWithinRoot preserves symlinks only if they resolve within the source root directory.
func WithPreserveSymlinkWithinRoot() CopyDirOption {
	return func(opts *copyDirOptions) {
		opts.symlinkBehavior = symlinkPreserveWithinRoot
	}
}
