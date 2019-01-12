import './main.css';
import registerServiceWorker from './registerServiceWorker';
import * as parseGitUrl from 'git-url-parse';
import * as AbsintheSocket from "@absinthe/socket";
import { Socket as PhoenixSocket } from "phoenix";
import { Elm } from './Main.elm';


const storageKey = "store";
let notifiers = [];

/**
 * Sockets
 */
document.addEventListener("DOMContentLoaded", function() {

    const absintheSocket = AbsintheSocket.create(
        new PhoenixSocket("ws://localhost:4000/socket")
    );

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

    app.ports.subscribeTo.subscribe(function ([id, operation]) {
        console.log("subscribeTo called with", id, operation);

        const notifier = AbsintheSocket.send(absintheSocket, {
            operation,
            variables: {}
        });

        function onStart(data) {
            console.log(">>> Start", JSON.stringify(data));
            app.ports.socketStatusConnected.send(id);
        }

        function onAbort(data) {
            console.log(">>> Abort", JSON.stringify(data));
        }

        function onError(value) {
            console.log(">>> Error", JSON.stringify(value));
            app.ports.socketStatusReconnecting.send(id);
        }

        function onResult(value) {
            console.log(">>> Result", JSON.stringify(value));
            app.ports.gotSubscriptionData.send({ id, value });
        }

        AbsintheSocket.observe(absintheSocket, notifier, {
            onAbort,
            onError,
            onStart,
            onResult
        })

    });

    app.ports.parseRepository.subscribe((repository) => {
        console.log('repository', repository);
        try {
            const gitUrl = parseGitUrl(repository);
            app.ports.parsedRepository.send({repository, gitUrl});
        } catch (e) {
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


// Whenever localStorage changes in another tab, report it if necessary.
    window.addEventListener("storage", function (event) {
        if (event.storageArea === localStorage && event.key === storageKey) {
            app.ports.onStoreChange.send(event.newValue);
        }
    }, false);

});

registerServiceWorker();
