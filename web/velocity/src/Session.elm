port module Session exposing
    ( AuthenticatedInternals
    , InitError(..)
    , Session
    , SubscriptionDataMsg
    , addKnownHost
    , addProject
    , changes
    , cred
    , fromViewer
    , knownHosts
    , log
    , navKey
    , projectWithId
    , projectWithSlug
    , projects
    , subscribe
    , subscriptionDataUpdate
    , subscriptions
    , viewer
    )

import Api exposing (BaseUrl, Cred)
import Api.Compiled.Query as Query
import Api.Subscriptions as Subscriptions
import Browser.Navigation as Nav
import Connection exposing (Connection)
import Context exposing (Context)
import Edge
import Event exposing (Event, Log)
import Graphql.Http
import Graphql.OptionalArgument exposing (OptionalArgument(..))
import Graphql.SelectionSet as SelectionSet exposing (SelectionSet)
import KnownHost exposing (KnownHost)
import Project exposing (Project)
import Project.Id
import Project.Slug
import Task exposing (Task)
import Viewer exposing (Viewer)



-- TYPES


type Session msg
    = LoggedIn (AuthenticatedInternals msg)


type AuthenticatedInternals msg
    = AuthenticatedInternals (LoggedInInternals msg)


type alias LoggedInInternals msg =
    { navKey : Nav.Key
    , viewer : Viewer
    , projects : List Project
    , knownHosts : List KnownHost
    , log : Event.Log
    , subscriptions : Subscriptions.State msg
    }


type InitError
    = HttpError (Graphql.Http.Error StartupResponse)
    | Unauthenticated


type SubscriptionDataMsg
    = KnownHostAdded KnownHost
    | ProjectAdded Project
    | EventAdded Event


type SubscriptionStatusMsg
    = SubscriptionStatus Subscriptions.StatusMsg



-- COLLECTIONS


knownHosts : Session msg -> List KnownHost
knownHosts session =
    case session of
        LoggedIn (AuthenticatedInternals internals) ->
            internals.knownHosts


addKnownHost : KnownHost -> Session msg -> Session msg
addKnownHost knownHost session =
    case session of
        LoggedIn (AuthenticatedInternals internals) ->
            LoggedIn (AuthenticatedInternals { internals | knownHosts = KnownHost.addKnownHost internals.knownHosts knownHost })


projectWithId : Project.Id.Id -> Session msg -> Maybe Project
projectWithId projectId session =
    case session of
        LoggedIn (AuthenticatedInternals internals) ->
            Project.findProjectById internals.projects projectId


projectWithSlug : Project.Slug.Slug -> Session msg -> Maybe Project
projectWithSlug projectSlug session =
    case session of
        LoggedIn (AuthenticatedInternals internals) ->
            Project.findProjectBySlug internals.projects projectSlug


projects : Session msg -> List Project
projects session =
    case session of
        LoggedIn (AuthenticatedInternals internals) ->
            internals.projects


addProject : Project -> Session msg -> Session msg
addProject p session =
    case session of
        LoggedIn (AuthenticatedInternals internals) ->
            LoggedIn (AuthenticatedInternals { internals | projects = Project.addProject p internals.projects })


addEvent : Event -> Session msg -> Session msg
addEvent event session =
    case session of
        LoggedIn (AuthenticatedInternals internals) ->
            LoggedIn (AuthenticatedInternals { internals | log = Event.addEvent event internals.log })



-- INFO


viewer : Session msg -> Viewer
viewer (LoggedIn (AuthenticatedInternals internals)) =
    internals.viewer


cred : Session msg -> Cred
cred (LoggedIn (AuthenticatedInternals internals)) =
    Viewer.cred internals.viewer


navKey : Session msg -> Nav.Key
navKey (LoggedIn (AuthenticatedInternals internals)) =
    internals.navKey


log : Session msg -> Event.Log
log (LoggedIn (AuthenticatedInternals internals)) =
    internals.log



-- SUBSCRIPTIONS


subscriptions : (Session msg -> msg) -> Session msg -> Sub msg
subscriptions toMsg session =
    case session of
        LoggedIn (AuthenticatedInternals internals) ->
            Subscriptions.subscriptions
                (\state ->
                    LoggedIn (AuthenticatedInternals { internals | subscriptions = state })
                        |> toMsg
                )
                internals.subscriptions


subscribe : (SubscriptionDataMsg -> msg) -> Session msg -> ( Session msg, Cmd msg )
subscribe toMsg session =
    case session of
        LoggedIn (AuthenticatedInternals internals) ->
            let
                ( subs, cmd ) =
                    ( internals.subscriptions, Cmd.none )
                        |> subscribeToKnownHostAdded toMsg
                        |> subscribeToProjectAdded toMsg
                        |> subscribeToEventAdded toMsg
            in
            ( LoggedIn (AuthenticatedInternals { internals | subscriptions = subs })
            , cmd
            )


subscribeToEventAdded :
    (SubscriptionDataMsg -> msg)
    -> ( Subscriptions.State msg, Cmd msg )
    -> ( Subscriptions.State msg, Cmd msg )
subscribeToEventAdded toMsg ( subs, cmd ) =
    let
        ( subscribed, subCmd ) =
            Subscriptions.subscribeToEventAdded (EventAdded >> toMsg) subs
    in
    ( subscribed
    , Cmd.batch [ cmd, subCmd ]
    )


subscribeToKnownHostAdded :
    (SubscriptionDataMsg -> msg)
    -> ( Subscriptions.State msg, Cmd msg )
    -> ( Subscriptions.State msg, Cmd msg )
subscribeToKnownHostAdded toMsg ( subs, cmd ) =
    let
        ( subscribed, subCmd ) =
            Subscriptions.subscribeToKnownHostAdded (KnownHostAdded >> toMsg) subs
    in
    ( subscribed
    , Cmd.batch [ cmd, subCmd ]
    )


subscribeToProjectAdded :
    (SubscriptionDataMsg -> msg)
    -> ( Subscriptions.State msg, Cmd msg )
    -> ( Subscriptions.State msg, Cmd msg )
subscribeToProjectAdded toMsg ( subs, cmd ) =
    let
        ( subscribed, subCmd ) =
            Subscriptions.subscribeToProjectAdded (ProjectAdded >> toMsg) subs
    in
    ( subscribed
    , Cmd.batch [ cmd, subCmd ]
    )


subscriptionDataUpdate : SubscriptionDataMsg -> Session msg -> Session msg
subscriptionDataUpdate subMsg session =
    case subMsg of
        KnownHostAdded knownHost ->
            addKnownHost knownHost session

        ProjectAdded project ->
            addProject project session

        EventAdded event ->
            addEvent event session



-- CHANGES


changes : (Task InitError (Session msg) -> msg2) -> Context msg -> Nav.Key -> Sub msg2
changes toMsg context navKey_ =
    Api.viewerChanges
        (\maybeViewer ->
            let
                _ =
                    Debug.log "MAYBE VIEWER" maybeViewer

                newSession =
                    fromViewer navKey_ context maybeViewer
            in
            toMsg newSession
        )
        Viewer.decoder


type alias StartupResponse =
    { projects : Connection Project
    , knownHosts : List KnownHost
    , log : Event.Log
    }


eventsArgs : Query.EventsOptionalArguments
eventsArgs =
    { after = Absent
    , before = Absent
    , first = Present 200
    , last = Absent
    }


fromViewer : Nav.Key -> Context msg2 -> Maybe Viewer -> Task InitError (Session msg)
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

                baseUrl =
                    Context.baseUrl context

                projectsSet =
                    Query.projects Project.projectListArgs Project.connectionSelectionSet

                knownHostSet =
                    Query.listKnownHosts KnownHost.selectionSet

                logSet =
                    Query.events (always eventsArgs) Event.selectionSet

                request =
                    SelectionSet.map3 StartupResponse (SelectionSet.nonNullOrFail projectsSet) knownHostSet (SelectionSet.nonNullOrFail logSet)
                        |> Api.authedQueryRequest baseUrl credVal
                        |> Graphql.Http.toTask
                        |> Task.mapError HttpError
            in
            Task.map
                (\res ->
                    LoggedIn
                        (AuthenticatedInternals
                            { navKey = key
                            , viewer = viewerVal
                            , projects = List.map Edge.node res.projects.edges
                            , knownHosts = res.knownHosts
                            , log = res.log
                            , subscriptions = Subscriptions.init
                            }
                        )
                )
                request

        Nothing ->
            Task.fail Unauthenticated
