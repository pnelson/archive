package deb

// A Package represents a Debian package.
type Package struct {
	Path         string            // The path to the .deb file.
	Name         string            // The base filename.
	Size         int64             // The size of the package in bytes.
	Files        []string          // The list of installable files.
	Fields       map[string]string // The control file as key/value pairs.
	ControlData  string            // The ./control file data.
	PreInstData  string            // The ./preinst file data.
	PreRmData    string            // The ./prerm file data.
	PostInstData string            // The ./postinst file data.
	PostRmData   string            // The ./postrm file data.
}
