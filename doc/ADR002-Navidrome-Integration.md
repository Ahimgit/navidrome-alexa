# ADR001 Navidrome Integration

### Goal

Provide a UI interface for selecting an Alexa/Echo playback device and controlling playback in the Navidrome UI.

### Context
A proper integration requires:
- Talking to other people 
- Saving the play queue/position (waiting on https://github.com/navidrome/navidrome/issues/245)
- UI changes in Navidrome to support selecting a playback device (potential collab with Jukebox UI)
- A plugin system/or a defined API? to send playback controls from Navidrome to a playback plugin
- Investigate if Subsonic APIs has something similar defined that can be reused

### Solution
To enable a quick prototype, a fully standalone solution has been selected without direct changes to Navidrome:
- Reverse proxy rewrite that injects a UI widget into Navidrome UI
- The widget captures the play queue from the Navidrome UI and sends it to navidrome-alexa & controls playback

### Consequences
- A short-term, potentially unattractive prototype solution
- The configuration needs to be updated with each Navidrome release
- May break at any time with UI changes in Navidrome