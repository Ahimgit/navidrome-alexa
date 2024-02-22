package response

type ResponseEnvelope struct {
	Version           string            `json:"version,omitempty"`
	SessionAttributes map[string]string `json:"sessionAttributes,omitempty"`
	UserAgent         string            `json:"userAgent,omitempty"`
	Response          *Response         `json:"response,omitempty"`
}

type Response struct {
	OutputSpeech     *OutputSpeech     `json:"outputSpeech,omitempty"`
	Card             *Card             `json:"card,omitempty"`
	Reprompt         *Reprompt         `json:"reprompt,omitempty"`
	Directives       []interface{}     `json:"directives,omitempty"`
	ShouldEndSession bool              `json:"shouldEndSession,omitempty"`
	CanFulfillIntent *CanFulfillIntent `json:"canFulfillIntent,omitempty"`
}

type Card struct {
	Type  string `json:"type,omitempty"`
	Title string `json:"title,omitempty"`
	Text  string `json:"text,omitempty"`
	Image *Image `json:"image,omitempty"`
}

type Image struct {
	SmallImageUrl string `json:"smallImageUrl,omitempty"`
	LargeImageUrl string `json:"largeImageUrl,omitempty"`
}

type OutputSpeech struct {
	Type         string `json:"type,omitempty"`
	PlayBehavior string `json:"playBehavior,omitempty"`
	Text         string `json:"text,omitempty"`
}

type Reprompt struct {
	OutputSpeech *OutputSpeech `json:"outputSpeech,omitempty"`
}

type Directive struct {
	Type string `json:"type,omitempty"`
}

type CanFulfillIntent struct {
	CanFulfill string          `json:"canFulfill,omitempty"`
	Slots      map[string]Slot `json:"slots,omitempty"`
}

type Slot struct {
	CanUnderstand string `json:"canUnderstand,omitempty"`
	CanFulfill    string `json:"canFulfill,omitempty"`
}

// builder

type ResponseBuilder struct {
	responseEnvelope *ResponseEnvelope
}

func NewResponseBuilder() *ResponseBuilder {
	return &ResponseBuilder{
		responseEnvelope: &ResponseEnvelope{
			Version:   "1.0",
			UserAgent: "na/1.0",
			Response: &Response{
				Directives: make([]interface{}, 0),
			},
		},
	}
}

func (b *ResponseBuilder) WithCanFulfillIntentYES() *ResponseBuilder {
	b.responseEnvelope.Response.CanFulfillIntent = &CanFulfillIntent{CanFulfill: "YES"}
	return b
}

func (b *ResponseBuilder) WithShouldEndSession(shouldEndSession bool) *ResponseBuilder {
	b.responseEnvelope.Response.ShouldEndSession = shouldEndSession
	return b
}

func (b *ResponseBuilder) AddAudioPlayerPlayDirective(directive *AudioPlayerPlayDirective) *ResponseBuilder {
	b.responseEnvelope.Response.Directives = append(b.responseEnvelope.Response.Directives, directive)
	return b
}

func (b *ResponseBuilder) AddAudioPlayerStopDirective() *ResponseBuilder {
	b.responseEnvelope.Response.Directives = append(b.responseEnvelope.Response.Directives, NewAudioPlayerStopDirective())
	return b
}

func (b *ResponseBuilder) AddAudioPlayerClearQueueDirective() *ResponseBuilder {
	b.responseEnvelope.Response.Directives = append(b.responseEnvelope.Response.Directives, NewAudioPlayerClearQueueDirective())
	return b
}

func (b *ResponseBuilder) Build() *ResponseEnvelope {
	return b.responseEnvelope
}
