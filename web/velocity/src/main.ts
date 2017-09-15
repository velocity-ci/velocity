import './styles.scss';

const app = require('./app/Main.elm').Main.fullscreen(localStorage.session || null);

app.ports.storeSession.subscribe(session => localStorage.session = session);

window.addEventListener('storage', event => {
  if (event.storageArea === localStorage && event.key === 'session') {
    app.ports.onSessionChange.send(event.newValue);
  }
}, false);
