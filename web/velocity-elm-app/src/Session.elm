module Session exposing (InitError, Session, changes, cred, errorToString, fromViewer, navKey, projects, viewer)

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
                    Project.list (Just credVal) baseUrl
                        |> Http.toTask
                        |> Task.mapError HttpError

                knownHostsRequest =
                    KnownHost.list (Just credVal) baseUrl
                        |> Http.toTask
                        |> Task.mapError HttpError
            in
            Task.map2
                (\projects_ knownHosts ->
                    LoggedIn <|
                        { navKey = key
                        , viewer = viewerVal
                        , projects = projects_
                        , knownHosts = knownHosts
                        }
                )
                projectsRequest
                knownHostsRequest

        Nothing ->
            Task.succeed (Guest key)
