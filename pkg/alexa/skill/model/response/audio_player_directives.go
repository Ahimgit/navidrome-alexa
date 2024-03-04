package response

type AudioPlayerStopDirective struct {
	Directive
}

func NewAudioPlayerStopDirective() *AudioPlayerStopDirective {
	return &AudioPlayerStopDirective{
		Directive: Directive{
			Type: "AudioPlayer.Stop",
		},
	}
}

type AudioPlayerClearQueueDirective struct {
	Directive
	ClearBehavior string `json:"clearBehavior,omitempty"`
}

func NewAudioPlayerClearQueueDirective() *AudioPlayerClearQueueDirective {
	return &AudioPlayerClearQueueDirective{
		Directive: Directive{
			Type: "AudioPlayer.ClearQueue",
		},
	}
}

type AudioPlayerPlayDirective struct {
	Directive
	PlayBehavior string     `json:"playBehavior,omitempty"`
	AudioItem    *AudioItem `json:"audioItem,omitempty"`
}

type AudioItem struct {
	Stream   *Stream   `json:"stream,omitempty"`
	Metadata *Metadata `json:"metadata,omitempty"`
}

type Stream struct {
	ExpectedPreviousToken string `json:"expectedPreviousToken,omitempty"`
	Token                 string `json:"token,omitempty"`
	URL                   string `json:"url,omitempty"`
	OffsetInMilliseconds  int    `json:"offsetInMilliseconds,omitempty"`
}

type Metadata struct {
	Title           string `json:"title,omitempty"`
	Subtitle        string `json:"subtitle,omitempty"`
	Art             *Art   `json:"art,omitempty"`
	BackgroundImage *Art   `json:"backgroundImage,omitempty"`
}

type Art struct {
	ContentDescription string   `json:"contentDescription,omitempty"`
	Sources            []Source `json:"sources,omitempty"`
}

type Source struct {
	URL          string `json:"url,omitempty"`
	Size         string `json:"size,omitempty"`
	WidthPixels  int    `json:"widthPixels,omitempty"`
	HeightPixels int    `json:"heightPixels,omitempty"`
}

// builder
type AudioPlayerPlayDirectiveBuilder struct {
	directive *AudioPlayerPlayDirective
}

func NewAudioPlayerPlayDirectiveBuilder() *AudioPlayerPlayDirectiveBuilder {
	return &AudioPlayerPlayDirectiveBuilder{
		directive: &AudioPlayerPlayDirective{
			Directive: Directive{
				Type: "AudioPlayer.Play",
			},
		}}
}

func (b *AudioPlayerPlayDirectiveBuilder) WithPlayBehaviorEnqueue() *AudioPlayerPlayDirectiveBuilder {
	b.directive.PlayBehavior = "ENQUEUE"
	return b
}

func (b *AudioPlayerPlayDirectiveBuilder) WithPlayBehaviorReplaceEnqueued() *AudioPlayerPlayDirectiveBuilder {
	b.directive.PlayBehavior = "REPLACE_ENQUEUED"
	return b
}

func (b *AudioPlayerPlayDirectiveBuilder) WithPlayBehaviorReplaceAll() *AudioPlayerPlayDirectiveBuilder {
	b.directive.PlayBehavior = "REPLACE_ALL"
	return b
}

func (b *AudioPlayerPlayDirectiveBuilder) WithAudioItem(audioItem *AudioItem) *AudioPlayerPlayDirectiveBuilder {
	b.directive.AudioItem = audioItem
	return b
}

func (b *AudioPlayerPlayDirectiveBuilder) Build() *AudioPlayerPlayDirective {
	return b.directive
}

type AudioItemBuilder struct {
	audioItem *AudioItem
}

func NewAudioItemBuilder() *AudioItemBuilder {
	return &AudioItemBuilder{
		audioItem: &AudioItem{},
	}
}

func (b *AudioItemBuilder) WithStream(stream *Stream) *AudioItemBuilder {
	b.audioItem.Stream = stream
	return b
}

func (b *AudioItemBuilder) WithMetadata(metadata *Metadata) *AudioItemBuilder {
	b.audioItem.Metadata = metadata
	return b
}

func (b *AudioItemBuilder) Build() *AudioItem {
	return b.audioItem
}

type MetadataBuilder struct {
	metadata *Metadata
}

func NewMetadataBuilder() *MetadataBuilder {
	return &MetadataBuilder{
		metadata: &Metadata{},
	}
}

func (b *MetadataBuilder) WithTitle(title string) *MetadataBuilder {
	b.metadata.Title = title
	return b
}

func (b *MetadataBuilder) WithSubtitle(subtitle string) *MetadataBuilder {
	b.metadata.Subtitle = subtitle
	return b
}

func (b *MetadataBuilder) WithArt(art *Art) *MetadataBuilder {
	b.metadata.Art = art
	return b
}

func (b *MetadataBuilder) WithBackgroundImage(backgroundImage *Art) *MetadataBuilder {
	b.metadata.BackgroundImage = backgroundImage
	return b
}

func (b *MetadataBuilder) Build() *Metadata {
	return b.metadata
}

type StreamBuilder struct {
	stream *Stream
}

func NewStreamBuilder() *StreamBuilder {
	return &StreamBuilder{
		stream: &Stream{},
	}
}

func (b *StreamBuilder) WithExpectedPreviousToken(token string) *StreamBuilder {
	b.stream.ExpectedPreviousToken = token
	return b
}

func (b *StreamBuilder) WithToken(token string) *StreamBuilder {
	b.stream.Token = token
	return b
}

func (b *StreamBuilder) WithURL(url string) *StreamBuilder {
	b.stream.URL = url
	return b
}

func (b *StreamBuilder) WithOffsetInMilliseconds(offset int) *StreamBuilder {
	b.stream.OffsetInMilliseconds = offset
	return b
}

func (b *StreamBuilder) Build() *Stream {
	return b.stream
}
