import './main.css';
import {Elm} from './Main.elm';
import registerServiceWorker from './registerServiceWorker';
import * as parseGitUrl from 'git-url-parse';
import Sockette from 'sockette';
const storageKey = "store";




/**
 * Sockets
 */

let app;

app = Elm.Main.init({
    node: document.getElementById('root'),
    flags: {
        viewer: localStorage.getItem(storageKey),
        baseUrl: process.env.ARCHITECT_ADDRESS,
        width: window.innerWidth,
        height: window.innerHeight
    }
});

console.log(Object.keys(app.ports));

app.ports.parseRepository.subscribe((repository) => {
    console.log('repository', repository);
    try {
        const gitUrl = parseGitUrl(repository);
        app.ports.parsedRepository.send({repository, gitUrl});
    }
    catch (e) {
        console.warn('Could not parse git URL', e.message);
        app.ports.parsedRepository.send({repository, gitUrl: null});
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

// app.ports.send_.subscribe((msg) => {
//     console.log('PORT: send message', msg);
//     ws.send(msg);
// });


// const ws = new Sockette('ws://localhost:4000/socket/websocket ', {
//     timeout: 5e3,
//     maxAttempts: 10,
//     onopen: (e) => {
//
//
//
//     },
//     onmessage: (msg) => {
//         console.log('PORT: on message', msg.data);
//         app.ports.onMessage.send(msg.data);
//     },
//     onreconnect: e => console.log('Reconnecting...', e),
//     onmaximum: e => console.log('Stop Attempting!', e),
//     onclose: e => console.log('Closed!', e),
//     onerror: e => console.log('Error:', e)
// });


//

// ws.send('Hello, world!');
// ws.json({type: 'ping'});
// ws.close();


// graceful shutdown

// Reconnect 10s later

// const socket = new WebSocket('ws://localhost/ws');

// socket.addEventListener('open', (event) => {
//     console.info('Websocket opened', event);
// });

// app.ports.open_.subscribe(function (url) {
//
//
//
//
//
//
// });

// Whenever localStorage changes in another tab, report it if necessary.
window.addEventListener("storage", function (event) {
    if (event.storageArea === localStorage && event.key === storageKey) {
        app.ports.onStoreChange.send(event.newValue);
    }
}, false);


registerServiceWorker();
