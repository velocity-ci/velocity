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

    app.ports.createSubscriptions.subscribe(function(subscription) {
        console.log("createSubscriptions called with", [subscription]);
        // Remove existing notifiers
        notifiers.map(notifier => AbsintheSocket.cancel(absintheSocket, notifier));

        // Create new notifiers for each subscription sent
        notifiers = [subscription].map(operation =>
            AbsintheSocket.send(absintheSocket, {
                operation,
                variables: {}
            })
        );

        function onStart(data) {
            console.log(">>> Start", JSON.stringify(data));
            app.ports.socketStatusConnected.send(null);
        }

        function onAbort(data) {
            console.log(">>> Abort", JSON.stringify(data));
        }

        function onCancel(data) {
            console.log(">>> Cancel", JSON.stringify(data));
        }

        function onError(data) {
            console.log(">>> Error", JSON.stringify(data));
            app.ports.socketStatusReconnecting.send(null);
        }

        function onResult(res) {
            console.log(">>> Result", JSON.stringify(res));
            app.ports.gotSubscriptionData.send(res);
        }

        notifiers.map(notifier =>
            AbsintheSocket.observe(absintheSocket, notifier, {
                onAbort,
                onError,
                onCancel,
                onStart,
                onResult
            })
        );
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
