import './main.css';
import {Elm} from './Main.elm';
import registerServiceWorker from './registerServiceWorker';
import * as parseGitUrl from 'git-url-parse';

const storageKey = "store";

const app = Elm.Main.init({
    node: document.getElementById('root'),
    flags: {
        viewer: localStorage.getItem(storageKey),
        baseUrl: process.env.ARCHITECT_ADDRESS,
        width: window.innerWidth,
        height: window.innerHeight
    }
});

console.log(Object.keys(app.ports));

app.ports.parseGitUrl.subscribe(([gitUrl, configuring]) => {
    console.log('gitUrl', gitUrl, 'configuring', configuring);
    try {
        const parsed = parseGitUrl(gitUrl);
        console.log('parsed', parsed);
        app.ports.onGitUrlParsed.send({gitUrl, parsed, configuring});
    }
    catch (e) {
        console.warn('Could not parse git URL', e.message);
        app.ports.onGitUrlParsed.send({gitUrl, parsed: null, configuring: false});
    }

});


app.ports.storeCache.subscribe(function (val) {
    if (val === null) {
        localStorage.removeItem(storageKey);
    } else {
        localStorage.setItem(storageKey, JSON.stringify(val));
    }
    // Report that the new session was stored succesfully.
    setTimeout(function () {
        app.ports.onStoreChange.send(val);
    }, 0);
});
// Whenever localStorage changes in another tab, report it if necessary.
window.addEventListener("storage", function (event) {
    if (event.storageArea === localStorage && event.key === storageKey) {
        app.ports.onStoreChange.send(event.newValue);
    }
}, false);


registerServiceWorker();
