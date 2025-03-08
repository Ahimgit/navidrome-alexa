const naWidgetModule = (function () {

    class PubSub {
        #statusUpdatedListeners;
        #settingsUpdatedListeners;

        constructor() {
            this.#statusUpdatedListeners = [];
            this.#settingsUpdatedListeners = [];
        }

        publishStatusUpdated(line1, line2, style) {
            this.#statusUpdatedListeners.forEach(listener => listener(line1, line2, style));
        }

        publishSettingsUpdated() {
            this.#settingsUpdatedListeners.forEach(listener => listener());
        }

        subscribeStatusUpdated(callback) {
            this.#statusUpdatedListeners.push(callback);
        }

        subscribeSettingsUpdated(callback) {
            this.#settingsUpdatedListeners.push(callback);
        }
    }

    class QueueLocalStorageAPI {
        #storage;

        constructor(storage) {
            this.#storage = storage;
        }

        getQueue() {
            const savedQueue = JSON.parse(this.#storage.getItem('state'));
            return savedQueue ? this.#mapQueue(savedQueue) : null;
        }

        #mapQueue(savedQueue) {
            let queuePosition = savedQueue.player.savedPlayIndex === -1 ? 0 : savedQueue.player.savedPlayIndex;
            return {
                trackPosition: 0, // todo cap from ui
                queuePosition: queuePosition,
                queue: savedQueue.player.queue.map(item => {
                    return {
                        id: item.trackId,
                        stream: item.musicSrc,
                        cover: item.cover,
                        name: item.song.title,
                        album: item.song.album,
                        artist: item.song.artist,
                        duration: Math.round((item.song.duration || 0) * 1000),
                    };
                })
            };
        }
    }

    class SettingsLocalStorageAPI {
        #storage;
        #data;
        #dirty;

        constructor(storage) {
            this.#storage = storage;
            this.#data = {};
            this.#dirty = true;
        }

        load() {
            this.#data = JSON.parse(this.#storage.getItem('naWidgetSettings')) || {};
            this.#dirty = false;
        }

        save() {
            this.#storage.setItem('naWidgetSettings', JSON.stringify(this.#data));
            this.#dirty = false;
        }

        setDirty() {
            this.#dirty = true;
        }

        setApiUrl(apiUrl) {
            this.#data.apiUrl = apiUrl;
            this.#dirty = true;
        }

        setApiKey(apiKey) {
            this.#data.apiKey = apiKey;
            this.#dirty = true;
        }

        setDevices(devices) {
            this.#data.devices = devices;
            this.#dirty = true;
        }

        setDeviceSelected(deviceSelected) {
            this.#data.deviceSelected = deviceSelected;
            this.#dirty = true;
        }

        getApiUrl() {
            return this.#data.apiUrl;
        }

        getApiKey() {
            return this.#data.apiKey;
        }

        getDevices() {
            return this.#data.devices;
        }

        getDeviceSelected() {
            return this.#data.deviceSelected;
        }

        isApiUrlSet() {
            return this.#data.apiUrl && this.#data.apiUrl.trim().length > 0;
        }

        isApiKeySet() {
            return this.#data.apiKey && this.#data.apiKey.trim().length > 0;
        }

        isDevicesSet() {
            return !!(this.#data.devices);
        }

        isDeviceSelected() {
            return !!(this.#data.deviceSelected);
        }

        isDirty() {
            return this.#dirty;
        }
    }

    class PlayerAPI {
        #settingsAPI;

        constructor(settingsAPI) {
            this.#settingsAPI = settingsAPI;
        }

        getDevices() {
            return this.#callAPI('GET', '/api/devices');
        }

        getQueue() {
            return this.#callAPI('GET', '/api/queue');
        }

        postQueue(queue) {
            return this.#callAPI('POST', '/api/queue', queue);
        }

        postPlay(device) {
            return this.#callAPI('POST', '/api/play', device);
        }

        postStop(device) {
            return this.#callAPI('POST', '/api/stop', device);
        }

        postNext(device) {
            return this.#callAPI('POST', '/api/next', device);
        }

        postPrev(device) {
            return this.#callAPI('POST', '/api/prev', device);
        }

        getPlaying() {
            return this.#callAPI('GET', '/api/playing');
        }

        getVolume() {
            return this.#callAPI('GET', '/api/volume');
        }

        postVolume(deviceVolume) {
            return this.#callAPI('POST', '/api/volume', deviceVolume);
        }

        async #callAPI(method, path, requestBody) {
            try {
                let headers = new Headers();
                headers.append('Authorization', `Bearer ${this.#settingsAPI.getApiKey()}`);
                let request = {
                    method: method,
                    headers: headers,
                };
                if (requestBody) {
                    headers.append('Content-Type', 'application/json');
                    request.body = JSON.stringify(requestBody);
                }
                const response = await fetch(this.#settingsAPI.getApiUrl() + path, request);
                const responseBody = await response.json();
                if (!response.ok) {
                    console.log('naW', 'got error back from the api', method, path, requestBody, responseBody);
                    return {error: responseBody.message};
                }
                return responseBody;
            } catch (error) {
                console.log('naW', 'error calling api', method, path, requestBody);
                return {error: error.message};
            }
        }
    }

    class SettingsController {
        #settingsAPI;
        #playerAPI;
        #pubSub;

        #apiUrlElement;
        #apiKeyElement;
        #deviceElement;
        #checkElement;
        #settingsElement;
        #settingsButtonElement;

        constructor(widget, settingsAPI, playerAPI, pubSub) {
            this.#settingsAPI = settingsAPI;
            this.#playerAPI = playerAPI;
            this.#pubSub = pubSub;
            this.#apiUrlElement = widget.getElement('apiUrl');
            this.#apiKeyElement = widget.getElement('apiKey');
            this.#deviceElement = widget.getElement('device');
            this.#checkElement = widget.getElement('check');
            this.#settingsElement = widget.getElement('settings');
            this.#settingsButtonElement = widget.getElement('settingsButton');
        }

        #loadValuesFromSettings() {
            this.#settingsAPI.load();
            if (this.#settingsAPI.isApiUrlSet()) {
                this.#apiUrlElement.value = this.#settingsAPI.getApiUrl();
            }
            if (this.#settingsAPI.isApiKeySet()) {
                this.#apiKeyElement.value = this.#settingsAPI.getApiKey();
            }
            if (this.#settingsAPI.isDevicesSet()) {
                while (this.#deviceElement.children.length > 1) {
                    this.#deviceElement.removeChild(this.#deviceElement.lastChild);
                }
                const deviceSelected = this.#settingsAPI.getDeviceSelected();
                this.#settingsAPI.getDevices().forEach(device => {
                    const option = document.createElement('option');
                    option.value = JSON.stringify(device);
                    option.textContent = device.name;
                    if (deviceSelected) {
                        if (deviceSelected.serialNumber === device.serialNumber) {
                            option.selected = true;
                        }
                    } else {
                        this.#settingsAPI.setDeviceSelected(device);
                        option.selected = true;
                    }
                    this.#deviceElement.appendChild(option);
                });
            }
        }

        bindControls() {
            this.#loadValuesFromSettings();

            this.#settingsButtonElement.addEventListener('click', () => {
                this.#settingsElement.classList.toggle('hidden');
            });

            this.#apiUrlElement.addEventListener('change', () => {
                this.#settingsAPI.setApiUrl(this.#apiUrlElement.value);
                this.#pubSub.publishSettingsUpdated();
            });

            this.#apiKeyElement.addEventListener('change', () => {
                this.#settingsAPI.setApiKey(this.#apiKeyElement.value);
                this.#pubSub.publishSettingsUpdated();
            });

            this.#deviceElement.addEventListener('change', () => {
                if (this.#deviceElement.value !== 'select') {
                    const isDirty = this.#settingsAPI.isDirty();
                    this.#settingsAPI.setDeviceSelected(JSON.parse(this.#deviceElement.value));
                    if (!isDirty) { // selecting new device should not trigger reloading devices
                        this.#settingsAPI.save();
                    }
                } else {
                    this.#settingsAPI.setDeviceSelected(null);
                }
                this.#pubSub.publishSettingsUpdated();
            });

            this.#checkElement.addEventListener('click', async () => {
                this.#checkElement.classList.add('disabled');
                try {
                    if (this.#apiKeyElement.value && this.#apiUrlElement.value) {
                        const devices = await this.#playerAPI.getDevices();
                        if (devices.devices) {
                            this.#settingsAPI.setDevices(devices.devices);
                            this.#settingsAPI.save();
                            this.#loadValuesFromSettings();
                        } else {
                            this.#pubSub.publishStatusUpdated('Error getting devices', devices.error, 'error');
                        }
                    }
                } finally {
                    this.#pubSub.publishSettingsUpdated();
                    this.#checkElement.classList.remove('disabled');
                }
            });
        }
    }

    class StatusController {
        #pubSub;

        #statusTextTopElement;
        #statusTextBottomElement;
        #settingsAPI;
        #playerAPI;

        constructor(widget, settingsAPI, playerAPI, pubSub) {
            this.#settingsAPI = settingsAPI;
            this.#playerAPI = playerAPI;
            this.#pubSub = pubSub;
            this.#statusTextTopElement = widget.getElement('status-text-top');
            this.#statusTextBottomElement = widget.getElement('status-text-bottom');
        }

        bindControls() {
            this.#pubSub.subscribeStatusUpdated((line1, line2) => {
                if (line1) {
                    this.#statusTextTopElement.innerText = line1;
                }
                if (line2) {
                    this.#statusTextBottomElement.innerText = line2;
                }
            });

            setInterval(async () => {
                if (this.#settingsAPI.isDirty()
                    || !this.#settingsAPI.isApiKeySet()
                    || !this.#settingsAPI.isApiUrlSet()
                    || !this.#settingsAPI.isDeviceSelected()) {
                    return;
                }
                const style = window.getComputedStyle(this.#statusTextTopElement);
                if (style.display === 'none' || document.hidden) {
                    return;
                }
                const playing = await this.#playerAPI.getPlaying();
                if (playing.error) {
                    this.#settingsAPI.setDirty();
                    this.#pubSub.publishSettingsUpdated();
                    this.#pubSub.publishStatusUpdated('Check your settings', 'Correct API URL, Key and select Device', 'error');
                } else {
                    if (playing.state === 'PLAYING') {
                        this.#pubSub.publishStatusUpdated(playing.song.name, `${playing.song.album} - ${playing.song.artist}`, 'normal');
                    } else if (playing.song) {
                        this.#pubSub.publishStatusUpdated(playing.song && playing.song.name || '', 'Stopped', 'normal');
                    } else {
                        this.#pubSub.publishStatusUpdated('Ready', `To play on ${this.#settingsAPI.getDweviceSelected().name}`, 'normal');
                    }
                }
            }, 1500);
        }
    }

    class PlayerController {
        #settingsAPI;
        #playerAPI;
        #queueAPI;
        #pubSub;

        #playButtonElement;
        #stopButtonElement;
        #prevButtonElement;
        #nextButtonElement;
        #volumeSliderElement;

        #lastPostedQueue;
        #lastPostedVolume;
        #volumeTimer;

        constructor(widget, settingsAPI, playerAPI, queueAPI, pubSub) {
            this.#settingsAPI = settingsAPI;
            this.#playerAPI = playerAPI;
            this.#queueAPI = queueAPI;
            this.#pubSub = pubSub;
            this.#playButtonElement = widget.getElement('play');
            this.#stopButtonElement = widget.getElement('stop');
            this.#prevButtonElement = widget.getElement('prev');
            this.#nextButtonElement = widget.getElement('next');
            this.#volumeSliderElement = widget.getElement('volume');
        }

        #convertSliderToVolume(slider) {
            return Math.ceil(Math.pow(slider, 2) / 100);
        }

        #convertVolumeToSliderValue(volume) {
            return Math.round(Math.sqrt(volume) * 10);
        }

        async #getCurrentVolume() {
            if (this.#settingsAPI.isApiKeySet() && this.#settingsAPI.isApiUrlSet() && this.#settingsAPI.isDeviceSelected()) {
                const volumeRS = await this.#playerAPI.getVolume();
                if (volumeRS.error || !volumeRS.volumes) {
                    this.#pubSub.publishStatusUpdated('Error getting volume', volumeRS.error, 'error');
                    return;
                }
                const device = this.#settingsAPI.getDeviceSelected();
                const volume = volumeRS.volumes.find(volume => volume.deviceSerialNumber === device.serialNumber);
                if (volume) {
                    this.#volumeSliderElement.value = this.#convertVolumeToSliderValue(volume.volume)
                }
            }
            this.#volumeSliderElement.style.setProperty('--progress', this.#volumeSliderElement.value + '%');
        }

        #setVolume() {
            let volume = this.#convertSliderToVolume(this.#volumeSliderElement.value);
            clearTimeout(this.#volumeTimer);
            this.#volumeTimer = setTimeout(async () => {
                const rs = await this.#playerAPI.postVolume({
                    device: this.#settingsAPI.getDeviceSelected(),
                    volume: volume
                });
                if (rs.error) {
                    this.#pubSub.publishStatusUpdated('Error sending volume', rs.error, 'error');
                    return;
                }
                this.#lastPostedVolume = volume;
            }, 1000);
        }

        #toggleControls() {
            if (!this.#settingsAPI.isDirty() && this.#settingsAPI.isApiKeySet() && this.#settingsAPI.isApiUrlSet() && this.#settingsAPI.isDeviceSelected()) {
                this.#playButtonElement.classList.remove('disabled');
                this.#stopButtonElement.classList.remove('disabled');
                this.#prevButtonElement.classList.remove('disabled');
                this.#nextButtonElement.classList.remove('disabled');
                this.#volumeSliderElement.disabled = false;
                this.#pubSub.publishStatusUpdated('Ready', `To play on ${this.#settingsAPI.getDeviceSelected().name}`, 'normal');
            } else {
                this.#playButtonElement.classList.add('disabled');
                this.#stopButtonElement.classList.add('disabled');
                this.#prevButtonElement.classList.add('disabled');
                this.#nextButtonElement.classList.add('disabled');
                this.#volumeSliderElement.disabled = true;
                this.#pubSub.publishStatusUpdated('Check your settings', 'Fill in API URL, Key and select Device', 'error');
            }
        }

        bindControls() {
            this.#pubSub.subscribeSettingsUpdated(() => this.#toggleControls());
            this.#toggleControls();
            this.#getCurrentVolume()

            const click = (element, callback) => {
                element.addEventListener('click', async () => {
                    element.classList.add('disabled');
                    try {
                        await callback();
                    } finally {
                        element.classList.remove('disabled');
                    }
                });
            };

            this.#volumeSliderElement.addEventListener('input', () => {
                const volume = this.#convertSliderToVolume(this.#volumeSliderElement.value)
                this.#pubSub.publishStatusUpdated('', `Volume: ${volume}%`, 'normal');
                this.#volumeSliderElement.style.setProperty('--progress', this.#volumeSliderElement.value + '%');
            });

            this.#volumeSliderElement.addEventListener('change', () => {
                this.#setVolume()
            });

            click(this.#playButtonElement, async () => {
                const queue = this.#queueAPI.getQueue();
                if (!queue || queue.queue.length === 0) {
                    this.#pubSub.publishStatusUpdated('Queue is empty', 'Select more items to play', 'warn');
                    return;
                }

                const queueString = JSON.stringify(queue);
                if (this.#lastPostedQueue !== queueString) {
                    const queueRS = await this.#playerAPI.postQueue(queue);
                    if (queueRS.error) {
                        this.#pubSub.publishStatusUpdated('Error sending queue', queueRS.error, 'error');
                        return;
                    }
                    this.#lastPostedQueue = queueString;
                }

                const playRS = await this.#playerAPI.postPlay(this.#settingsAPI.getDeviceSelected());
                if (playRS.error) {
                    this.#pubSub.publishStatusUpdated('Error sending play', playRS.error, 'error');
                }
            });

            click(this.#stopButtonElement, async () => {
                const rs = await this.#playerAPI.postStop(this.#settingsAPI.getDeviceSelected());
                if (rs.error) {
                    this.#pubSub.publishStatusUpdated('Error sending stop', rs.error, 'error');
                }
            });

            click(this.#prevButtonElement, async () => {
                const rs = await this.#playerAPI.postPrev(this.#settingsAPI.getDeviceSelected());
                if (rs.error) {
                    this.#pubSub.publishStatusUpdated('Error sending prev', rs.error, 'error');
                }
            });

            click(this.#nextButtonElement, async () => {
                const rs = await this.#playerAPI.postNext(this.#settingsAPI.getDeviceSelected());
                if (rs.error) {
                    this.#pubSub.publishStatusUpdated('Error sending next', rs.error, 'error');
                }
            });
        }
    }

    class Widget extends HTMLElement {

        constructor() {
            super();
            this.attachShadow({mode: 'open'});
        }

        getElement(id) {
            return this.shadowRoot.getElementById(id);
        }

        #bindControls() {
            this.getElement('closeButton').addEventListener('click', () => {
                this.getElement('widget').classList.add('hidden');
            });
        }

        // @Overrides
        connectedCallback() {
            this.shadowRoot.innerHTML = `
                <style>
                    #widget {
                        z-index: 9000;
                        position: fixed; width: 200px;
                        display: flex; align-items: center; flex-direction: column; flex-wrap: wrap; justify-content: center;
                        font-family: -apple-system, BlinkMacSystemFont, Segoe UI, Roboto, Oxygen, Ubuntu, Cantarell, Fira Sans, Droid Sans, Helvetica Neue, sans-serif; font-size: 10px; color: #b3b3b3;
                        border-radius: 4px 4px 0 0; 
                        background: rgba(0, 0, 0, 0.75);
                        transition: left 0.3s ease-in-out, bottom 0.3s ease-in-out;
                        padding: 4px;
                    }
                    #widget > div { margin-top: 8px; }
                    #closeButton { display: none; } 
                    #settings { display: flex; flex-direction: column; justify-content: center; }
                    #settingsButton { width: 160px; font-size: 10px; line-height: 9px; cursor: pointer; user-select: none; text-align: center; margin: 0 !important;}
                    #settingsButton:hover { color: #fff }
                    #status { display: flex; flex-direction: column; align-items: center; }
                    #status-text-top { padding: 2px 6px 2px 6px; display: inline-block; }
                    #status-text-top:hover { animation: marquee 7s linear infinite; }
                    #status-text-bottom { padding: 2px 6px 2px 6px; display: inline-block; font-size: 9px; }
                    #status-text-bottom:hover { animation: marquee 7s linear infinite; }
                    .status-marquee { position: relative; width: 160px; margin: 0 8px 0 8px; overflow: hidden; white-space: nowrap; text-align: center; mask-image: linear-gradient(to left, transparent, black 4%, black 96%, transparent); }
                    @keyframes marquee {
                        0% { transform: translateX(0); }
                        50% { transform: translateX(min(0px, calc(-100% + 180px))); }
                        100% { transform: translateX(0); }
                    }                    
                    .form-group { position: relative; margin: 10px 4px 4px 4px; padding: 2px; display: inline-flex; border: 1px solid rgba(255, 255, 255, 0.23); border-radius: 4px; }
                    .form-group:hover { border: 1px solid #fff; }
                    .form-group > label {
                        position: absolute;
                        top: -9px;
                        left: 8px;
                        padding: 2px 4px 2px 4px;
                        pointer-events: none;
                        background-color: #000;
                        border-radius: 2px;
                    }
                    .form-group > input, .form-group > select {
                        width: 160px;
                        padding: 6px 4px 4px 6px;
                        box-sizing: content-box;
                        color: currentColor;
                        border: none;
                        outline: none;
                        background: none;
                    }
                    .form-group > select > option { color: #fff; background: rgba(0, 0, 0, 0.3); text-shadow: 0 1px 0 rgba(0, 0, 0, 0.4); }
                    .form-row { margin: 6px 4px 4px 4px; display: inline-flex; justify-content: center; }
                    .button {
                        margin: 4px; padding: 3px 10px 5px 10px;
                        display: inline-block; cursor: pointer; user-select: none;
                        transition: background-color 0.1s, box-shadow 0.1s;
                        font-size: 12px; line-height: 15px; text-align: center; text-decoration: none; color: #fff;
                        border: none; border-radius: 2px; background-color: rgba(0, 0, 0, 0.45);
                    }
                    #controls { display: flex; flex-direction: column; justify-content: center;  }
                    #controls .button { font-size: 22px; padding: 3px;  margin: 1px;}
                    .button.disabled { cursor: default; pointer-events: none; opacity: 0.5; background-color: #616161; box-shadow: none; }
                    .button:hover, .button:active { background-color: #616161; box-shadow: 0 2px 4px rgba(0, 0, 0, 0.4); }
                    input[type="range"] { -webkit-appearance: none; appearance: none; background: transparent; cursor: pointer; height: 1rem; width: 120px; }
                    input[type="range"]:focus { outline: none; }                                  
                    input[type="range"]::-webkit-slider-runnable-track { border-radius: 0.5rem; height: 0.3rem; background: #5f5fc4 linear-gradient(to right, #5f5fc4 0%, #5f5fc4 var(--progress,0%), #fff var(--progress, 0%), #fff 100%); }
                    input[type="range"]::-webkit-slider-thumb { -webkit-appearance: none; appearance: none; background-color: #5f5fc4; outline: 2px solid #fff; border: none; margin-top: -0.1rem; border-radius: 0.5rem; height: 0.5rem; width: 0.5rem; } 
                    input[type="range"]::-moz-range-track { border-radius: 0.5rem; height: 0.3rem; background: #5f5fc4 linear-gradient(to right, #5f5fc4 0%, #5f5fc4 var(--progress,0%), #fff var(--progress, 0%), #fff 100%); }
                    input[type="range"]::-moz-range-thumb { background-color: #5f5fc4; outline: 2px solid #fff; border: none; border-radius: 0.5rem; height: 0.5rem; width: 0.5rem; }             
                    input[type="range"]:disabled::-webkit-slider-runnable-track { background: #616161; }                    
                    input[type="range"]:disabled::-webkit-slider-thumb { background: #616161; }         
                    input[type="range"]:disabled::-moz-range-track { background: #616161; }                    
                    input[type="range"]:disabled::-moz-range-thumb { background: #616161; }      
                    
                    .hidden { display: none !important; }
                   
                    #widget.mobile { top:0 !important; left:0 !important; height: 100vh; width: 100vw; overflow: hidden;  box-sizing: border-box; }                          
                    #widget.mobile #closeButton { display: block; position: absolute; right: 20px; top: 20px; font-size: 24px; }
                    #widget.mobile #controls .button { font-size: 34px; } 
                </style>
                <div id="widget" class="hidden">
                    <span id="closeButton">
                        <svg class="icon cross"  xmlns="http://www.w3.org/2000/svg" height="1em" width="1em" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <line x1="19" y1="5" x2="5" y2="19" stroke="currentColor"/>
                            <line x1="5" y1="5" x2="19" y2="19" stroke="currentColor"/>
                        </svg>    
                    </span>
                    <div class="hidden" id="settings">
                        <div class="form-group">
                            <label for="apiUrl">API URL</label>
                            <input id="apiUrl" type="text">
                        </div>
                        <div class="form-group">
                            <label for="apiKey">API Key</label>
                            <input id="apiKey" type="password">
                        </div>
                        <div class="form-group">
                            <label for="device">Device </label>
                            <select id="device">
                                <option value="select">- select device -</option>
                            </select>
                        </div>
                        <div class="form-row">
                            <span id="check" class="button">Save settings &#x1F408;</span>
                        </div>                   
                    </div>
                    <div id="settingsButton">&#8230;</div>
                    <div id="status">
                        <div class="status-marquee">
                            <span id="status-text-top"></span>
                        </div>       
                        <div class="status-marquee">
                            <span id="status-text-bottom"></span>
                        </div>
                    </div>
                    <div id="controls">
                        <div class="form-row">
                            <span id="prev" class="button disabled">
                                <svg class="icon prev"  xmlns="http://www.w3.org/2000/svg" height="1em" width="1em" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                    <polygon points="18,16 11.6,12 18,8" stroke="currentColor" fill="currentColor"/>
                                    <line x1="6.6" y1="6" x2="6.6" y2="18" stroke="currentColor" stroke-width="3"/>
                                </svg>      
                            </span>
                            <span id="play" class="button disabled">
                                <svg class="icon play" xmlns="http://www.w3.org/2000/svg" height="1em" width="1em" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                    <circle cx="12" cy="12" r="11" stroke="currentColor"/>
                                    <polygon points="9.6,16 16,12 9.6,8" stroke="currentColor" fill="currentColor"/>
                                </svg>
                            </span>
                            <span id="stop" class="button disabled">
                                <svg class="icon stop"  xmlns="http://www.w3.org/2000/svg" height="1em" width="1em" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                    <circle cx="12" cy="12" r="11" stroke="currentColor"/>
                                    <line x1="9"  y1="8" x2="9" y2="16" stroke="currentColor" stroke-width="3"/>
                                    <line x1="15" y1="8" x2="15" y2="16" stroke="currentColor" stroke-width="3"/>
                                </svg>
                            </span>
                            <span id="next" class="button disabled">
                                <svg class="icon next" xmlns="http://www.w3.org/2000/svg" height="1em" width="1em" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                    <polygon points="6.6,16 13,12 6.6,8" stroke="currentColor" fill="currentColor"/>
                                    <line x1="18" y1="6" x2="18" y2="18" stroke="currentColor" stroke-width="3"/>
                                </svg>                     
                            </span>
                        </div>
                        <div class="form-row">
                            <input id="volume" disabled type="range" min="0" max="100" value="0" step="1">                  
                        </div>
                    </div>    
                </div>
            `;

            this.#bindControls();
        }
    }


    class NavidromeUIIntegration {
        // here be dragons ; may cause severe eye damage ; you have been warned

        static svgDevice = `
            <svg class="iconDevice" xmlns="http://www.w3.org/2000/svg" height="24" width="24" viewBox="0 0 24 24" stroke-linecap="round">
                <rect x="4" y="2" width="16" height="20" rx="2" ry="2" fill="none" stroke="currentColor" stroke-width="2"/>
                <circle cx="12" cy="14" r="3" stroke="currentColor" fill="none"/>
                <line x1="12" y1="7" x2="12.1" y2="7" stroke="currentColor" stroke-width="2"/>
            </svg>`;

        static classPlayer = 'music-player-panel';
        static classPlayerMobile = 'react-jinke-music-player-mobile';
        static classPlaylist = 'audio-lists-panel';
        static classLyricsButton = 'lyric-btn';

        #widgetElement;
        #playButtonElement;

        constructor(widget) {
            this.#widgetElement = widget.getElement('widget');
            this.#playButtonElement = widget.getElement('play');
        }

        #getElementByClass(className) {
            return document.querySelector('.' + className);
        }

        #repositionWidget() {
            const naPlayer = this.#getElementByClass(NavidromeUIIntegration.classPlayer);
            const naPlaylist = this.#getElementByClass(NavidromeUIIntegration.classPlaylist);
            const naPlayerMobile = this.#getElementByClass(NavidromeUIIntegration.classPlayerMobile);
            if (naPlayer) {
                const widgetRect = this.#widgetElement.getBoundingClientRect();
                const playlistRect = naPlaylist.getBoundingClientRect();
                const playerRect = naPlayer.getBoundingClientRect();
                this.#widgetElement.style.bottom = `${playerRect.height}px`;
                this.#widgetElement.style.left = `${playlistRect.left - widgetRect.width}px`;
                this.#widgetElement.classList.remove('mobile');
            } else if (naPlayerMobile) {
                this.#widgetElement.classList.add('mobile');
            }
        }

        #addWidgetPositioningListeners() {
            const naPlaylist = this.#getElementByClass(NavidromeUIIntegration.classPlaylist);
            const reposition = () => { // captures this.
                this.#repositionWidget();
            };
            if (naPlaylist) {
                naPlaylist.addEventListener('transitionend', reposition);
                naPlaylist.addEventListener('animationend', reposition);
            }
            window.addEventListener('resize', reposition);
        }

        #createDeviceIcon() {
            const naLyricsButton = this.#getElementByClass(NavidromeUIIntegration.classLyricsButton);
            const naPlayerMobile = this.#getElementByClass(NavidromeUIIntegration.classPlayerMobile);

            const deviceButton = document.createElement('span');
            deviceButton.id = 'naWToggleButton';
            deviceButton.classList.add('group');
            deviceButton.innerHTML = NavidromeUIIntegration.svgDevice;
            deviceButton.addEventListener('click', () => {
                this.#widgetElement.classList.toggle('hidden');
                this.#repositionWidget();
            });
            if (naPlayerMobile) {
                const li = document.createElement('li');
                li.classList.add('item');
                li.appendChild(deviceButton);
                naLyricsButton.parentElement.after(li);
            } else {
                naLyricsButton.after(deviceButton);
            }
        }

        attachToNavidrome() {
            this.#playButtonElement.addEventListener('click', () => {
                document.querySelectorAll('audio').forEach(element => {
                    element.pause(); // stop browser playback
                });
            });

            const bind = () => {
                if (document.getElementById('naWToggleButton')) {
                    return;
                }
                this.#createDeviceIcon();
                this.#repositionWidget();
                this.#addWidgetPositioningListeners();
            };

            setTimeout(() => {
                if (this.#getElementByClass(NavidromeUIIntegration.classLyricsButton)) {
                    bind(); // direct element bind
                }
            }, 200);
            new MutationObserver(mutations => {
                for (let mutation of mutations) {
                    if (mutation.addedNodes.length) {
                        mutation.addedNodes.forEach(node => {
                            if (node.nodeType === Node.ELEMENT_NODE && (
                                node.classList.contains(NavidromeUIIntegration.classPlayer) ||
                                node.classList.contains(NavidromeUIIntegration.classPlayerMobile))) {
                                bind(); // layout change bind
                            }
                        });
                    }
                }
            }).observe(document.body, {childList: true, subtree: true});
        }
    }

    class Main {
        init() {
            customElements.define('na-widget', Widget);
            window.addEventListener('DOMContentLoaded', () => {
                try {
                    const widget = new Widget();
                    document.body.appendChild(widget);
                    const settingsAPI = new SettingsLocalStorageAPI(localStorage);
                    const queueAPI = new QueueLocalStorageAPI(localStorage);
                    const playerAPI = new PlayerAPI(settingsAPI);
                    const pubSub = new PubSub();
                    const statusController = new StatusController(widget, settingsAPI, playerAPI, pubSub);
                    const settingsController = new SettingsController(widget, settingsAPI, playerAPI, pubSub);
                    const playerController = new PlayerController(widget, settingsAPI, playerAPI, queueAPI, pubSub);
                    const navi = new NavidromeUIIntegration(widget);
                    statusController.bindControls();
                    settingsController.bindControls();
                    playerController.bindControls();
                    navi.attachToNavidrome();
                } catch (error) {
                    console.log('naW', 'error attaching widget', error);
                }
            });
        }
    }

    return {
        Main,
    };
})();
new naWidgetModule.Main().init();