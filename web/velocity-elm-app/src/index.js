import './main.css';
import {Elm} from './Main.elm';
import registerServiceWorker from './registerServiceWorker';

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
