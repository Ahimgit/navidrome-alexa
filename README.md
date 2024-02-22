# navidrome-alexa

Stream your music from Navidrome UI to Amazon Echo & Alexa devices.

## How it works

Navidrome-Alexa is a combination of a UI widget, REST API and Alexa Skill that allow you to stream your music from
Navidrome Web UI to Alexa devices like Amazon Echo.
See the below diagram for details how components interact with each other.

<img alt="architecture" src="doc/architecture.svg">

### Why it works that way
See below for more details on Navidrome and Alexa integrations.

[ADR001 Alexa Interactions](doc/ADR001-Alexa-Interactions.md)    
[ADR002 Navidrome Integration](doc/ADR002-Navidrome-Integration.md)

### How it looks like



## Installation

### Prerequisites

A typical installation requires:

- Public-web accessible address for navidrome-alexa to accept Amazon Alexa Skill API requests
- Public-web accessible address for Navidrome's /rest/stream endpoint 
- Reverse proxy with SSL and rewriting support ([Caddy](https://github.com/caddyserver/caddy) is used in examples below)
- Amazon account to access [Alexa Developer Console](https://developer.amazon.com/alexa/console/ask) to configure
  self-hosted Alexa Skill

### 1. Configure skill

- 1.1. Open [Alexa Developer Console](https://developer.amazon.com/alexa/console/ask) and authenticate.
- 1.2. Click "Create New Skill" button.
- 1.3. Enter "navi stream" as skill name
- 1.4. On the "Experience, Model, Hosting Service" screen select:
    - Choose a type of experience: "Music & Audio" [(picture)](doc/install-1-4-1.png)
    - Choose a model: "Custom" [(picture)](doc/install-1-4-2.png)
    - Hosting services: "Provision your own"
    - Click "Next" button
- 1.5. On the "Template" screen select "Start from Scratch" and click "Next" button [(picture)](doc/install-1-5.png)
- 1.6. On the "Review" screen click "Create Skill" button, wait till skill is created
- 1.7. Go to "Intents" in left side menu and hit "JSON editor", copy & paste [alexa-skill.json](doc/alexa-skill.json)
  and click "Save" [(picture)](doc/install-1-7.png)
- 1.8. Go to "Endpoint", select "HTTPS" as "Service Endpoint Type"  [(picture)](doc/install-1-8.png)
    - Enter public https URL pointing to your navidrome-alexa installation ending with /skill (
      e.g. https://alexa.yourdomain.com/skill )
    - Select SSL certificate type you use (make sure to select wildcard cert type if using it for subdomains)
    - Click "Save" button
- 1.9. Go to "Interfaces" left side menu, enable "Audio Player" and press "Save" [(picture)](doc/install-1-9.png)
- 1.10. Click "Build" in top menu, click "Build skill" button, ensure skill builds successfully
- 1.11. Got to developer console root page and click "Copy Skill ID", you will need it for the next
  step.  [(picture)](doc/install-1-11.png)

### 2. Configure application

Download a pre-built [release of navidrome-alexa](https://github.com/Ahimgit/navidrome-alexa/releases) or build locally.
Run it passing configuration parameters below.

Following configuration parameters are accepted:

| Command line         | Env var | Default value | Description                                                                                          |
|----------------------|---------|---------------|------------------------------------------------------------------------------------------------------|
| amazonDomain         |         | amazon.com    | Base domain to use for Alexa API calls.                                                              |
| amazonCookiePath     |         | cookies.data  | Path to a writable file to store auth cookies.                                                       |   
| amazonUser           |         | _Required_    | Amazon account email with Alexa devices, can be left blank if auth cookies already exist.            | 
| amazonPassword       |         | _Required_    | Amazon account password, can be left blank if auth cookies already exist.                            | 
| apiKey               |         | _Required_    | Required. API key to authenticate /client calls. User provided, select arbitrary string to match 4.1 |         
| streamDomain         |         | _Required_    | Required. Navidrome public server domain URL.                                                        |         
| alexaSkillId         |         | _Required_    | Required. Skill id to authenticate calls from Alexa. Has to match copied in 1.11.                    |     
| alexaSkillName       |         | navi stream   | Skill invocation name. Has to match name configured in 1.7. JSON                                     |                           
| listenAddress        |         | :8080         | Listen address.                                                                                      |                                  
| logIncomingRequests  |         | false         | Log API and Skill requests/responses.                                                                |            
| logOutgoingRequests  |         | false         | Log outgoing (to Alexa APIs) requests/responses. Will leak sensitive data into logs.                 | 

Minimal configuration via command line example:

```shell
  na \
  -amazonUser your@email.com \
  -amazonPassword youramazonpassword \
  -apiKey yourlongenoughandsecureapikey \
  -alexaSkillId amzn1.ask.skill.xxxxx \
  -streamDomain https://navidrome.youdomain.com \ 
```

### 3. Configure proxy

### 4. Configure widget

### 3. Play

## Monitoring

### Monitoring

### Logging

## Setting up development environment

### Building locally

### Running tests

## Known issues & todo

- No re-authentication in Alexa client, if cookie token is revoked for some reason `cookies.data` needs to be deleted (
  although they have 1 year expiry)
- Authentication may be tricky and may require authing from a mobile app on the same network first to do CAPTCHA. 
- Better/more secure of configuration params
- Proper integration with Navidrome vs injected widget
- Better UI for playback controls / progress
- Volume controls
- Separate port for /health & /metrics
- Proper signature validation of incoming /skill requests
- Per-device queue / state
- More (than zero) tests