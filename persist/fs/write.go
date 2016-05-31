	"time"
	xtime "code.uber.internal/infra/memtsdb/x/time"
	defaultNewFileMode      = os.FileMode(0666)
	defaultNewDirectoryMode = os.ModeDir | os.FileMode(0755)
	start            time.Time
	window           time.Duration
	filePathPrefix   string
	newFileMode      os.FileMode
	newDirectoryMode os.FileMode

	infoBuffer   *proto.Buffer
type WriterOptions interface {
	// NewFileMode sets the new file mode.
	NewFileMode(value os.FileMode) WriterOptions

	// GetNewFileMode returns the new file mode.
	GetNewFileMode() os.FileMode

	// NewDirectoryMode sets the new directory mode.
	NewDirectoryMode(value os.FileMode) WriterOptions

	// GetNewDirectoryMode returns the new directory mode.
	GetNewDirectoryMode() os.FileMode
}

type writerOptions struct {
	newFileMode      os.FileMode
	newDirectoryMode os.FileMode
}

// NewWriterOptions creates a writer options.
func NewWriterOptions() WriterOptions {
	return &writerOptions{
		newFileMode:      defaultNewFileMode,
		newDirectoryMode: defaultNewDirectoryMode,
	}
}

func (o *writerOptions) NewFileMode(value os.FileMode) WriterOptions {
	opts := *o
	opts.newFileMode = value
	return &opts
}

func (o *writerOptions) GetNewFileMode() os.FileMode {
	return o.newFileMode
}

func (o *writerOptions) NewDirectoryMode(value os.FileMode) WriterOptions {
	opts := *o
	opts.newDirectoryMode = value
	return &opts
func (o *writerOptions) GetNewDirectoryMode() os.FileMode {
	return o.newDirectoryMode
}

// NewWriter returns a new writer for a filePathPrefix
func NewWriter(
	start time.Time,
	window time.Duration,
	filePathPrefix string,
	options WriterOptions,
) Writer {
	if options == nil {
		options = NewWriterOptions()
	}
	return &writer{
		start:            start,
		window:           window,
		filePathPrefix:   filePathPrefix,
		newFileMode:      options.GetNewFileMode(),
		newDirectoryMode: options.GetNewDirectoryMode(),
		infoBuffer:       proto.NewBuffer(nil),
		indexBuffer:      proto.NewBuffer(nil),
		varintBuffer:     proto.NewBuffer(nil),
		idxData:          make([]byte, idxLen),
	}
// Open initializes the internal state for writing to the given shard,
// specifically creating the shard directory if it doesn't exist, and
// opening / truncating files associated with that shard for writing.
func (w *writer) Open(shard uint32) error {
	shardDir := ShardDirPath(w.filePathPrefix, shard)
	if err := os.MkdirAll(shardDir, w.newDirectoryMode); err != nil {
		return err
	}
	nextVersion, err := nextVersion(shardDir)
		return err
	return openFiles(
		writeableFileOpener(w.newFileMode),
		map[string]**os.File{
			filepathFromVersion(shardDir, nextVersion, infoFileSuffix):  &w.infoFd,
			filepathFromVersion(shardDir, nextVersion, indexFileSuffix): &w.indexFd,
			filepathFromVersion(shardDir, nextVersion, dataFileSuffix):  &w.dataFd,
		},
	)
	info := &schema.IndexInfo{
		Start:   xtime.ToNanoseconds(w.start),
		Window:  int64(w.window),
		Entries: w.currIdx,
	}
	if err := w.infoBuffer.Marshal(info); err != nil {
	if _, err := w.infoFd.Write(w.infoBuffer.Bytes()); err != nil {