module Session
    exposing
        ( InitError
        , Session
        , SocketUpdate
        , addKnownHost
        , addProject
        , branches
        , changes
        , cred
        , fromViewer
        , joinChannels
        , joinProjectChannel
        , knownHosts
        , log
        , navKey
        , projectWithId
        , projectWithSlug
        , projects
        , socketUpdate
        , viewer
        )

import Activity
import Api exposing (BaseUrl, Cred)
import Browser.Navigation as Nav
import Context exposing (Context)
import Http
import Json.Decode as Decode
import KnownHost exposing (KnownHost)
import Phoenix.Channel as Channel exposing (Channel)
import Phoenix.Socket as Socket exposing (Socket)
import Project exposing (Project)
import Project.Branch as Branch exposing (Branch)
import Project.Id
import Project.Slug
import Task exposing (Task)
import Viewer exposing (Viewer)
import Graphql.Http
import Graphql.Http.GraphqlError as GraphqlError
import Api.Compiled.Query as Query
import Graphql.SelectionSet as SelectionSet
import Set


-- TYPES


type Session
    = LoggedIn LoggedInInternals
    | Guest Nav.Key


type alias LoggedInInternals =
    { navKey : Nav.Key
    , viewer : Viewer
    , projects : List Project
    , branches : Project.Id.Dict (List Branch)
    , knownHosts : List KnownHost
    , log : Activity.Log
    }


type InitError
    = HttpError (Graphql.Http.Error StartupResponse)


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


projectWithId : Project.Id.Id -> Session -> Maybe Project
projectWithId projectId session =
    case session of
        LoggedIn internals ->
            Project.findProjectById internals.projects projectId

        Guest _ ->
            Nothing


projectWithSlug : Project.Slug.Slug -> Session -> Maybe Project
projectWithSlug projectSlug session =
    case session of
        LoggedIn internals ->
            Project.findProjectBySlug internals.projects projectSlug

        Guest _ ->
            Nothing


projects : Session -> List Project
projects session =
    case session of
        LoggedIn internals ->
            internals.projects

        Guest _ ->
            []


branches : Project.Id.Id -> Session -> List Branch
branches projectId session =
    case session of
        LoggedIn internals ->
            Project.Id.get projectId internals.branches
                |> Maybe.withDefault []

        Guest _ ->
            []


addProject : Project -> Session -> Session
addProject p session =
    case session of
        LoggedIn internals ->
            LoggedIn { internals | projects = Project.addProject p internals.projects }

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


log : Session -> Maybe Activity.Log
log session =
    case session of
        LoggedIn internals ->
            Just internals.log

        Guest _ ->
            Nothing



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
                        (\p ( context_, cmd_ ) ->
                            let
                                ( updatedContext, newCmd ) =
                                    joinProjectChannel { cred_ = cred_, toMsg = toMsg, context_ = context_ } p
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
joinProjectChannel { cred_, toMsg, context_ } p =
    let
        channelName =
            Project.channelName p

        decoder encodedValue =
            Decode.decodeValue Project.decoder encodedValue
                |> Result.toMaybe
                |> Maybe.map (ProjectUpdated >> toMsg)
                |> Maybe.withDefault (toMsg NoOp)
    in
        Context.on "project:update" channelName decoder context_
            |> Context.joinChannel (Project.channel p) cred_


socketUpdate : SocketUpdate -> (SocketUpdate -> msg) -> Context msg -> Session -> ( Session, Context msg, Cmd (Socket.Msg msg) )
socketUpdate update toMsg context session =
    case session of
        LoggedIn internals ->
            case update of
                ProjectUpdated p ->
                    ( LoggedIn { internals | projects = Project.updateProject p internals.projects }
                    , context
                    , Cmd.none
                    )

                ProjectAdded p ->
                    let
                        credVal =
                            Viewer.cred internals.viewer

                        ( updatedContext, joinCmd ) =
                            joinProjectChannel { cred_ = credVal, toMsg = toMsg, context_ = context } p
                    in
                        ( LoggedIn
                            { internals
                                | projects = Project.addProject p internals.projects
                                , log = Activity.projectAdded (Project.id p) internals.log
                            }
                        , updatedContext
                        , joinCmd
                        )

                NoOp ->
                    ( session, context, Cmd.none )

        Guest _ ->
            ( session, context, Cmd.none )


type alias StartupResponse =
    { projects : List Project
    , knownHosts : List KnownHost
    }


fromViewer : Nav.Key -> Context msg2 -> Maybe Viewer -> Task InitError Session
fromViewer key context maybeViewer =
    -- It's stored in localStorage as a JSON String;
    -- first decode the Value as a String, then
    -- decode that String as JSON.
    -- If the person is logged in we will attempt to get a
    case maybeViewer of
        Just viewerVal ->
            let
                baseUrl =
                    Context.baseUrl context

                credVal =
                    Viewer.cred viewerVal

                projectsSet =
                    Query.projects Project.selectionSet
                        |> SelectionSet.nonNullOrFail
                        |> SelectionSet.nonNullElementsOrFail

                knownHostSet =
                    Query.knownHosts KnownHost.selectionSet
                        |> SelectionSet.nonNullOrFail
                        |> SelectionSet.nonNullElementsOrFail

                request =
                    SelectionSet.map2 StartupResponse projectsSet knownHostSet
                        |> Graphql.Http.queryRequest "http://localhost:4000/v2"
                        |> Graphql.Http.toTask
                        |> Task.mapError HttpError
            in
                Task.map
                    (\res ->
                        LoggedIn <|
                            { navKey = key
                            , viewer = viewerVal
                            , projects = res.projects
                            , branches = Project.Id.empty
                            , knownHosts = res.knownHosts
                            , log = Activity.init
                            }
                    )
                    request

        Nothing ->
            Task.succeed (Guest key)
