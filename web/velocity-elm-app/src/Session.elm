module Session exposing (InitError, Session, cred, fromViewer, navKey, viewer)

import Api exposing (BaseUrl, Cred)
import Browser.Navigation as Nav
import Http
import Project exposing (Project)
import Task exposing (Task)
import Viewer exposing (Viewer)



-- TYPES


type Session
    = LoggedIn Nav.Key Viewer
    | Guest Nav.Key



-- INFO


viewer : Session -> Maybe Viewer
viewer session =
    case session of
        LoggedIn _ val ->
            Just val

        Guest _ ->
            Nothing


cred : Session -> Maybe Cred
cred session =
    case session of
        LoggedIn _ val ->
            Just (Viewer.cred val)

        Guest _ ->
            Nothing


navKey : Session -> Nav.Key
navKey session =
    case session of
        LoggedIn key _ ->
            key

        Guest key ->
            key



-- CHANGES
--changes : (Session -> msg) -> Session -> Sub msg
--changes toMsg session =
--    Api.viewerChanges
--        (\maybeViewer ->
--            case maybeViewer of
--                Just viewerVal ->
--                    toMsg (LoggedIn navKey viewerVal)
--
--                Nothing ->
--                    toMsg (Guest navKey)
--        )
--        Viewer.decoder


type InitError
    = HttpError Http.Error


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
            in
            Project.list (Just credVal) baseUrl
                |> Http.toTask
                |> Task.mapError HttpError
                |> Task.map (\_ -> LoggedIn key viewerVal)

        Nothing ->
            Task.succeed (Guest key)
