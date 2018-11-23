module Session exposing (InitError, Session, SocketUpdate, addKnownHost, addProject, changes, cred, errorToString, fromViewer, joinChannels, joinProjectChannel, knownHosts, navKey, projects, socketUpdate, viewer)

import Api exposing (BaseUrl, Cred)
import Browser.Navigation as Nav
import Context exposing (Context)
import Http
import Json.Decode as Decode
import KnownHost exposing (KnownHost)
import Phoenix.Channel as Channel exposing (Channel)
import Phoenix.Socket as Socket exposing (Socket)
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


type SocketUpdate
    = ProjectUpdated Project
    | ProjectAdded Project
    | NoOp



-- COLLECTIONS


knownHosts : Session -> List KnownHost
knownHosts session =
    case session of
        LoggedIn internals ->
            internals.knownHosts

        Guest _ ->
            []


addKnownHost : KnownHost -> Session -> Session
addKnownHost knownHost session =
    case session of
        LoggedIn internals ->
            LoggedIn { internals | knownHosts = KnownHost.addKnownHost internals.knownHosts knownHost }

        Guest _ ->
            session


projects : Session -> List Project
projects session =
    case session of
        LoggedIn internals ->
            internals.projects

        Guest _ ->
            []


addProject : Project -> Session -> Session
addProject project session =
    case session of
        LoggedIn internals ->
            LoggedIn { internals | projects = Project.addProject project internals.projects }

        Guest _ ->
            session



-- INFO


viewer : Session -> Maybe Viewer
viewer session =
    case session of
        LoggedIn internals ->
            Just internals.viewer

        Guest _ ->
            Nothing


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


changes : (Task InitError Session -> msg) -> Context msg2 -> Session -> Sub msg
changes toMsg context session =
    Api.viewerChanges (fromViewer (navKey session) context >> toMsg) Viewer.decoder


joinChannels : Session -> (SocketUpdate -> msg) -> Context msg -> ( Context msg, Cmd (Socket.Msg msg) )
joinChannels session toMsg context =
    case session of
        Guest _ ->
            ( context
            , Cmd.none
            )

        LoggedIn internals ->
            let
                cred_ =
                    Viewer.cred internals.viewer

                ( projectJoinedContext, projectsChannelCmd ) =
                    joinProjectsChannel { cred_ = cred_, toMsg = toMsg, context_ = context }
            in
            internals.projects
                |> List.foldl
                    (\project ( context_, cmd_ ) ->
                        let
                            ( updatedContext, newCmd ) =
                                joinProjectChannel { cred_ = cred_, toMsg = toMsg, context_ = context_ } project
                        in
                        ( updatedContext
                        , Cmd.batch [ cmd_, newCmd ]
                        )
                    )
                    ( projectJoinedContext, projectsChannelCmd )


joinProjectsChannel :
    { cred_ : Cred, toMsg : SocketUpdate -> msg, context_ : Context msg }
    -> ( Context msg, Cmd (Socket.Msg msg) )
joinProjectsChannel { cred_, toMsg, context_ } =
    let
        channelName =
            "projects"

        decoder encodedValue =
            Decode.decodeValue Project.decoder encodedValue
                |> Result.toMaybe
                |> Maybe.map (ProjectAdded >> toMsg)
                |> Maybe.withDefault (toMsg NoOp)
    in
    Context.on "project:new" channelName decoder context_
        |> Context.joinChannel (Channel.init channelName) cred_


joinProjectChannel :
    { cred_ : Cred, toMsg : SocketUpdate -> msg, context_ : Context msg }
    -> Project
    -> ( Context msg, Cmd (Socket.Msg msg) )
joinProjectChannel { cred_, toMsg, context_ } project =
    let
        channelName =
            Project.channelName project

        decoder encodedValue =
            Decode.decodeValue Project.decoder encodedValue
                |> Result.toMaybe
                |> Maybe.map (ProjectUpdated >> toMsg)
                |> Maybe.withDefault (toMsg NoOp)
    in
    Context.on "project:update" channelName decoder context_
        |> Context.joinChannel (Project.channel project) cred_


socketUpdate : SocketUpdate -> (SocketUpdate -> msg) -> Context msg -> Session -> ( Session, Context msg, Cmd (Socket.Msg msg) )
socketUpdate update toMsg context session =
    case session of
        LoggedIn internals ->
            case update of
                ProjectUpdated project ->
                    ( LoggedIn { internals | projects = Project.updateProject project internals.projects }
                    , context
                    , Cmd.none
                    )

                ProjectAdded project ->
                    let
                        credVal =
                            Viewer.cred internals.viewer

                        ( updatedContext, joinCmd ) =
                            joinProjectChannel { cred_ = credVal, toMsg = toMsg, context_ = context } project
                    in
                    ( LoggedIn { internals | projects = Project.addProject project internals.projects }
                    , updatedContext
                    , joinCmd
                    )

                NoOp ->
                    ( session, context, Cmd.none )

        Guest _ ->
            ( session, context, Cmd.none )


fromViewer : Nav.Key -> Context msg2 -> Maybe Viewer -> Task InitError Session
fromViewer key context maybeViewer =
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
                    Project.list credVal (Context.baseUrl context)
                        |> Http.toTask
                        |> Task.mapError HttpError

                knownHostsRequest =
                    KnownHost.list credVal (Context.baseUrl context)
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
