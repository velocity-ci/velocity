module Session exposing (InitError, Session, addKnownHost, addProject, changes, cred, errorToString, fromViewer, knownHosts, navKey, projects, viewer)

import Api exposing (BaseUrl, Cred)
import Browser.Navigation as Nav
import Http
import KnownHost exposing (KnownHost)
import Project exposing (Project)
import Task exposing (Task)
import Viewer exposing (Viewer)



-- TYPES


type Session
    = LoggedIn LoggedInInternals
    | Guest Nav.Key


type alias LoggedInInternals =
    { navKey : Nav.Key
    , viewer : Viewer
    , projects : List Project
    , knownHosts : List KnownHost
    }


type InitError
    = HttpError Http.Error



-- INFO


viewer : Session -> Maybe Viewer
viewer session =
    case session of
        LoggedIn internals ->
            Just internals.viewer

        Guest _ ->
            Nothing


projects : Session -> List Project
projects session =
    case session of
        LoggedIn internals ->
            internals.projects

        Guest _ ->
            []


knownHosts : Session -> List KnownHost
knownHosts session =
    case session of
        LoggedIn internals ->
            internals.knownHosts

        Guest _ ->
            []


cred : Session -> Maybe Cred
cred session =
    case session of
        LoggedIn internals ->
            Just (Viewer.cred internals.viewer)

        Guest _ ->
            Nothing


navKey : Session -> Nav.Key
navKey session =
    case session of
        LoggedIn internals ->
            internals.navKey

        Guest key ->
            key


errorToString : InitError -> String
errorToString (HttpError httpError) =
    case httpError of
        Http.BadUrl error ->
            "Bad URL: " ++ error

        Http.NetworkError ->
            "Network Error"

        Http.BadStatus _ ->
            "Bad Status"

        Http.BadPayload payload _ ->
            "Bad Payload: " ++ payload

        Http.Timeout ->
            "Timeout"



-- CHANGES


addKnownHost : KnownHost -> Session -> Session
addKnownHost knownHost session =
    case session of
        LoggedIn internals ->
            LoggedIn { internals | knownHosts = KnownHost.addKnownHost internals.knownHosts knownHost }

        Guest _ ->
            session


addProject : Project -> Session -> Session
addProject project session =
    case session of
        LoggedIn internals ->
            LoggedIn { internals | projects = Project.addProject internals.projects project }

        Guest _ ->
            session


changes : (Task InitError Session -> msg) -> BaseUrl -> Session -> Sub msg
changes toMsg baseUrl session =
    Api.viewerChanges (fromViewer (navKey session) baseUrl >> toMsg) Viewer.decoder


fromViewer : Nav.Key -> BaseUrl -> Maybe Viewer -> Task InitError Session
fromViewer key baseUrl maybeViewer =
    -- It's stored in localStorage as a JSON String;
    -- first decode the Value as a String, then
    -- decode that String as JSON.
    -- If the person is logged in we will attempt to get a
    case maybeViewer of
        Just viewerVal ->
            let
                credVal =
                    Viewer.cred viewerVal

                projectsRequest =
                    Project.list credVal baseUrl
                        |> Http.toTask
                        |> Task.mapError HttpError

                knownHostsRequest =
                    KnownHost.list credVal baseUrl
                        |> Http.toTask
                        |> Task.mapError HttpError
            in
            Task.map2
                (\projects_ knownHosts_ ->
                    LoggedIn <|
                        { navKey = key
                        , viewer = viewerVal
                        , projects = projects_
                        , knownHosts = knownHosts_
                        }
                )
                projectsRequest
                knownHostsRequest

        Nothing ->
            Task.succeed (Guest key)
