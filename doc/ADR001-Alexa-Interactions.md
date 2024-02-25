# ADR001 Alexa Integration for Music Streaming and Playback Control

### Goal
Stream music to Alexa Echo device with playback control via UI (not voice).
Similar to how Spotify "Play on device" button works.

### Context
There is no public documented API exist to send commands to Alexa skills.

Following internal APIs exist:

#### alexa.amazon.com/api/entertainment/v1/player/queue
Spotify/TuneIn use this API to start/control playback from the mobile app.
It accepts content/playlist tokens and skill id that are sent as zipped & b64 encoded string.
Unfortunately substituting skill id with a self-hosted skill produces an error. 

#### alexa.amazon.com/api/cloudplayer/queue-and-play
Amazon Music uses this API to start playback of music. 
It does not have any visible ability to pass data to third-party skills, therefore unusable.

#### alexa.amazon.com/api/behaviors/preview (Alexa.TextCommand)
Text command API used by Alexa Mobile App to send "voice" commands via chat text messages

All of the above APIs require oauth-like form/cookie authentication that is used by Alexa mobile app.

### Solution

- Current implementation will use Alexa.TextCommand and simulate form/cookie authentication.
- It will issue text commands like "Alexa tell skill_name to play" to control playback. 
- A self-hosted will be build to respond to those commands using public [AudioPlayer Alexa API](https://developer.amazon.com/en-US/docs/alexa/custom-skills/audioplayer-interface-reference.html)

### Consequences
- Authentication is quite fragile and may trigger CAPTCHA which can be mitigated by logging-in and entering it from a mobile app on the same network.
- Authentication approach / used internal APIs may break anytime without notice
- If other (better) options become available this implementation needs to be revised.