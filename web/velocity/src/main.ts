import './styles.scss';

const flags = {
  session: localStorage.session || null,
  apiUrlBase: process.env.ARCHITECT_ADDRESS
};

const app = require("./app/Main.elm").Main.fullscreen(flags);

app.ports.storeSession.subscribe(session => localStorage.session = session);

window.addEventListener('storage', event => {
  if (event.storageArea === localStorage && event.key === 'session') {
    app.ports.onSessionChange.send(event.newValue);
  }
}, false);
