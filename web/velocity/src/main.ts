import './styles.scss';

import { parseGitUrl } from 'git-url-parse';


const flags = {
  session: localStorage.session || null,
  apiUrlBase: process.env.ARCHITECT_ADDRESS
};

const app = require("./app/Main.elm").Main.fullscreen(flags);

/*
Session port.

Stores session to local storage
 */

app.ports.storeSession.subscribe(session => localStorage.session = session);

window.addEventListener('storage', event => {
  if (event.storageArea === localStorage && event.key === 'session') {
    app.ports.onSessionChange.send(event.newValue);
  }
}, false);


/*
 Bottom of page port.

 Used to stick the build output to the bottom if scrolled to it
 */
window.onscroll = () => {
  const scrolledToBottom = (window.innerHeight + window.pageYOffset) >= document.body.offsetHeight - 2;
  app.ports.onScrolledToBottom.send(scrolledToBottom);
};


/*
Git URL parse port.

Parses and sends git parsed git urls
 */
app.ports.parseGitUrl.subscribe(gitUrl => {
  const parsed = parseGitUrl(gitUrl);
  app.ports.onGitUrlParsed.send(parsed);
});
