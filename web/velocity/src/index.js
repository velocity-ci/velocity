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
        new PhoenixSocket(process.env.ARCHITECT_CHANNEL_WS_ADDRESS)
    );

    const app = Elm.Main.init({
        node: document.getElementById('root'),
        flags: {
            viewer: localStorage.getItem(storageKey),
            baseUrl: process.env.ARCHITECT_GRAPHQL_HTTP_ADDRESS,
            width: window.innerWidth,
            height: window.innerHeight
        }
    });


    app.ports.subscribeTo.subscribe(function ([id, operation]) {
        console.log("subscribeTo called with", id, operation);

        const notifier = AbsintheSocket.send(absintheSocket, {
            operation,
            variables: {}
        });

        function onStart(data) {
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
        try {
            const gitUrl = parseGitUrl(repository);
            app.ports.parsedRepository.send({repository, gitUrl});
        } catch (e) {
            console.warn('Could not parse git URL', e.message);
            app.ports.parsedRepository.send({repository, gitUrl: null});
        }

    });


    app.ports.storeCache.subscribe(function (val) {

        console.log('STORE CACHE', val);
        if (val === null) {
            localStorage.removeItem(storageKey);
        } else {
            localStorage.setItem(storageKey, JSON.stringify(val));
        }
        // Report that the new session was stored succesfully.
        setTimeout(function () {
            console.log('port', 'onStoreChange NEW', val);

            app.ports.onStoreChange.send(val);
        }, 0);
    });


// Whenever localStorage changes in another tab, report it if necessary.
    window.addEventListener("storage", function (event) {
        console.log('port', 'onStoreChange', event.newValue);


        if (event.storageArea === localStorage && event.key === storageKey) {
            app.ports.onStoreChange.send(event.newValue);
        }
    }, false);

});

registerServiceWorker();
