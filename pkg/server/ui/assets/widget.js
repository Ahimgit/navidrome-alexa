naWidgetModule = (function () {

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
                this.#statusTextTopElement.innerText = line1;
                this.#statusTextBottomElement.innerText = line2;
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

        #lastPostedQueue;

        constructor(widget, settingsAPI, playerAPI, queueAPI, pubSub) {
            this.#settingsAPI = settingsAPI;
            this.#playerAPI = playerAPI;
            this.#queueAPI = queueAPI;
            this.#pubSub = pubSub;
            this.#playButtonElement = widget.getElement('play');
            this.#stopButtonElement = widget.getElement('stop');
            this.#prevButtonElement = widget.getElement('prev');
            this.#nextButtonElement = widget.getElement('next');
        }

        #toggleControls() {
            if (!this.#settingsAPI.isDirty() && this.#settingsAPI.isApiKeySet() && this.#settingsAPI.isApiUrlSet() && this.#settingsAPI.isDeviceSelected()) {
                this.#playButtonElement.classList.remove('disabled');
                this.#stopButtonElement.classList.remove('disabled');
                this.#prevButtonElement.classList.remove('disabled');
                this.#nextButtonElement.classList.remove('disabled');
                this.#pubSub.publishStatusUpdated('Ready', `To play on ${this.#settingsAPI.getDeviceSelected().name}`, 'error');
            } else {
                this.#playButtonElement.classList.add('disabled');
                this.#stopButtonElement.classList.add('disabled');
                this.#prevButtonElement.classList.add('disabled');
                this.#nextButtonElement.classList.add('disabled');
                this.#pubSub.publishStatusUpdated('Check your settings', 'Fill in API URL, Key and select Device', 'error');
            }
        }

        bindControls() {
            this.#pubSub.subscribeSettingsUpdated(() => this.#toggleControls());
            this.#toggleControls();

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
                        transition: right 0.3s ease-in-out, bottom 0.3s ease-in-out;
                    }
                    #widget > div { margin: 8px 4px 4px 4px; }
                    #settings { display: flex; flex-direction: column; justify-content: center; }
                    #settingsButton { width: 160px; font-size: 10px; line-height: 9px; cursor: pointer; user-select: none; text-align: center; margin: 0 !important; }
                    #settingsButton:hover { color: #fff }
                    #status { display: flex; flex-direction: column; align-items: center; }
                    #status-text-top { padding: 2px 6px 2px 6px; display: inline-block; }
                    #status-text-top:hover { animation: marquee 7s linear infinite; }
                    #status-text-bottom { padding: 2px 6px 2px 6px; display: inline-block; font-size: 9px; }
                    #status-text-bottom:hover { animation: marquee 7s linear infinite; }
                    .status-marquee { position: relative; width: 180px; margin: 0 8px 0 8px; overflow: hidden; white-space: nowrap; text-align: center; mask-image: linear-gradient(to left, transparent, black 4%, black 96%, transparent); }
                    @keyframes marquee {
                        0% { transform: translateX(0); }
                        50% { transform: translateX(min(0px, calc(-100% + 180px))); }
                        100% { transform: translateX(0); }
                    }
                    
                    .form-group { position: relative; margin: 6px; padding: 2px; display: inline-flex; border: 1px solid rgba(255, 255, 255, 0.23); border-radius: 4px; }
                    .form-group:hover { border: 1px solid #fff; }
                    .form-group > label {
                        position: absolute;
                        top: -8px;
                        left: 10px;
                        padding: 2px;
                        pointer-events: none;
                        background-color: #000;
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
                    .form-row { margin:8px; display: inline-flex; justify-content: center; }
                    .button {
                        margin: 4px; padding: 3px 10px 5px 10px;
                        display: inline-block; cursor: pointer; user-select: none;
                        transition: background-color 0.1s, box-shadow 0.1s;
                        font-size: 12px; line-height: 15px; text-align: center; text-decoration: none; color: #FFF;
                        border: none; border-radius: 2px; background-color: #424242;
                    }
                    .button.disabled { cursor: default; pointer-events: none; opacity: 0.5; background-color: #616161; box-shadow: none; }
                    .button:hover, .button:active { background-color: #616161; box-shadow: 0 2px 4px rgba(0, 0, 0, 0.4); }
                    .hidden { display: none !important; }
                </style>
                <div id="widget" class="hidden">
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
                        <span id="prev" class="button disabled">&#x23EE;</span>
                        <span id="play" class="button disabled">&#x23F5;</span>
                        <span id="stop" class="button disabled">&#x23F8;︎</span>
                        <span id="next" class="button disabled">&#x23ED;</span>                    
                    </div>
                </div>
            `;
        }
    }


    class NavidromeUIIntegration {
        // here be dragons ; may cause severe eye damage ; you have been warned

        #widgetElement;

        constructor(widget) {
            this.#widgetElement = widget.getElement('widget');
        }

        #repositionWidget() {
            const naPlaylist = document.querySelector('.audio-lists-panel');
            if (!naPlaylist) {
                return;
            }
            const rect = naPlaylist.getBoundingClientRect();
            this.#widgetElement.style.right = Math.round(document.documentElement.clientWidth - rect.left) + 'px';
            this.#widgetElement.style.bottom = Math.round(document.documentElement.clientHeight - rect.bottom) + 'px';
        }

        #addWidgetPositioningListeners() {
            const naPlaylist = document.querySelector('.audio-lists-panel');
            if (!naPlaylist) {
                return;
            }
            const reposition = () => {
                this.#repositionWidget();
            }; // cap this
            naPlaylist.addEventListener('transitionend', reposition);
            naPlaylist.addEventListener('animationend', reposition);
            window.addEventListener('resize', reposition);
        }

        #createDeviceIcon() {
            const naLyricsButton = document.querySelector('.lyric-btn');
            if (!naLyricsButton) {
                return;
            }
            const deviceButton = document.createElement('span');
            deviceButton.id = 'naWToggleButton';
            deviceButton.classList.add('group');
            deviceButton.innerHTML = `
                <svg xmlns="http://www.w3.org/2000/svg"  height="1em" width="1em"  viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2">
                    <path stroke="none" d="M0 0h24v24H0z" fill="none"></path>
                    <path d="M5 3m0 2a2 2 0 0 1 2 -2h10a2 2 0 0 1 2 2v14a2 2 0 0 1 -2 2h-10a2 2 0 0 1 -2 -2z"></path>
                    <path d="M12 14m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0"></path>
                    <path d="M12 7l0 .01"></path>
                </svg>
            `;
            deviceButton.addEventListener('click', () => {
                this.#widgetElement.classList.toggle('hidden');
                this.#repositionWidget();
            });
            naLyricsButton.after(deviceButton);
        }

        attachToNavidrome() {
            const naPlayerClass = 'react-jinke-music-player-main';
            const naPlayerMobileClass = 'react-jinke-music-player-mobile';
            const naPlayerPanelClass = 'music-player-panel';
            const bind = () => {
                this.#createDeviceIcon();
                this.#repositionWidget();
                this.#addWidgetPositioningListeners();
            };
            if (document.querySelector('.' + naPlayerClass) && !document.querySelector('.' + naPlayerMobileClass)) {
                bind();
            }
            new MutationObserver(mutations => {
                for (let mutation of mutations) {
                    if (mutation.addedNodes.length) {
                        mutation.addedNodes.forEach(node => {
                            if (node.nodeType === Node.ELEMENT_NODE) {
                                if (node.classList.contains(naPlayerPanelClass) || node.classList.contains(naPlayerClass)) {
                                    bind();
                                }
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
            const widget = new Widget();
            document.addEventListener('DOMContentLoaded', () => {
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
            });
        }
    }

    return {
        Main,
    };
})();
new naWidgetModule.Main().init();