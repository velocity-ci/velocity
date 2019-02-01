port module Session exposing
    ( InitError
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

import Activity
import Api exposing (BaseUrl, Cred)
import Api.Compiled.Query as Query
import Api.Subscriptions as Subscriptions
import Browser.Navigation as Nav
import Connection exposing (Connection)
import Context exposing (Context)
import Edge
import Graphql.Http
import Graphql.SelectionSet as SelectionSet exposing (SelectionSet)
import KnownHost exposing (KnownHost)
import PageInfo exposing (PageInfo)
import Project exposing (Project)
import Project.Branch exposing (Branch)
import Project.Id
import Project.Slug
import Task exposing (Task)
import Viewer exposing (Viewer)



-- TYPES


type Session msg
    = LoggedIn (LoggedInInternals msg)
    | Guest Nav.Key


type alias LoggedInInternals msg =
    { navKey : Nav.Key
    , viewer : Viewer
    , projects : Connection Project
    , knownHosts : List KnownHost
    , log : Activity.Log
    , subscriptions : Subscriptions.State msg
    }


type InitError
    = HttpError (Graphql.Http.Error StartupResponse)


type SubscriptionDataMsg
    = KnownHostAdded KnownHost
    | ProjectAdded Project


type SubscriptionStatusMsg
    = SubscriptionStatus Subscriptions.StatusMsg



-- COLLECTIONS


knownHosts : Session msg -> List KnownHost
knownHosts session =
    case session of
        LoggedIn internals ->
            internals.knownHosts

        Guest _ ->
            []


addKnownHost : KnownHost -> Session msg -> Session msg
addKnownHost knownHost session =
    case session of
        LoggedIn internals ->
            LoggedIn { internals | knownHosts = KnownHost.addKnownHost internals.knownHosts knownHost }

        Guest _ ->
            session


projectWithId : Project.Id.Id -> Session msg -> Maybe Project
projectWithId projectId session =
    case session of
        LoggedIn internals ->
            Project.findProjectById (List.map Edge.node internals.projects.edges) projectId

        Guest _ ->
            Nothing


projectWithSlug : Project.Slug.Slug -> Session msg -> Maybe Project
projectWithSlug projectSlug session =
    case session of
        LoggedIn internals ->
            Project.findProjectBySlug (List.map Edge.node internals.projects.edges) projectSlug

        Guest _ ->
            Nothing


projects : Session msg -> List Project
projects session =
    case session of
        LoggedIn internals ->
            internals.projects.edges
                |> List.map Edge.node

        Guest _ ->
            []


addProject : Project -> Session msg -> Session msg
addProject p session =
    case session of
        LoggedIn internals ->
            LoggedIn internals

        --            LoggedIn { internals | projects = Project.addProject p internals.projects }
        Guest _ ->
            session



-- INFO


viewer : Session msg -> Maybe Viewer
viewer session =
    case session of
        LoggedIn internals ->
            Just internals.viewer

        Guest _ ->
            Nothing


cred : Session msg -> Maybe Cred
cred session =
    case session of
        LoggedIn internals ->
            Just (Viewer.cred internals.viewer)

        Guest _ ->
            Nothing


navKey : Session msg -> Nav.Key
navKey session =
    case session of
        LoggedIn internals ->
            internals.navKey

        Guest key ->
            key


log : Session msg -> Maybe Activity.Log
log session =
    case session of
        LoggedIn internals ->
            Just internals.log

        Guest _ ->
            Nothing



-- SUBSCRIPTIONS


subscriptions : (Session msg -> msg) -> Session msg -> Sub msg
subscriptions toMsg session =
    case session of
        LoggedIn internals ->
            Subscriptions.subscriptions
                (\state ->
                    LoggedIn { internals | subscriptions = state }
                        |> toMsg
                )
                internals.subscriptions

        Guest _ ->
            Sub.none


subscribe : (SubscriptionDataMsg -> msg) -> Session msg -> ( Session msg, Cmd msg )
subscribe toMsg session =
    case session of
        LoggedIn internals ->
            let
                ( subs, cmd ) =
                    ( internals.subscriptions, Cmd.none )
                        |> subscribeToKnownHostAdded toMsg
                        |> subscribeToProjectAdded toMsg
            in
            ( LoggedIn { internals | subscriptions = subs }
            , cmd
            )

        Guest _ ->
            ( session, Cmd.none )


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



-- CHANGES


changes : (Task InitError (Session msg) -> msg2) -> Context msg -> Session msg -> Sub msg2
changes toMsg context session =
    Api.viewerChanges
        (\maybeViewer ->
            let
                newSession =
                    fromViewer (navKey session) context maybeViewer
            in
            toMsg newSession
        )
        Viewer.decoder


type alias StartupResponse =
    { projects : Connection Project
    , knownHosts : List KnownHost
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

                request =
                    SelectionSet.map2 StartupResponse (SelectionSet.nonNullOrFail projectsSet) knownHostSet
                        |> Api.authedQueryRequest baseUrl credVal
                        |> Graphql.Http.toTask
                        |> Task.mapError HttpError
            in
            Task.map
                (\res ->
                    LoggedIn
                        { navKey = key
                        , viewer = viewerVal
                        , projects = res.projects
                        , knownHosts = res.knownHosts
                        , log = Activity.init
                        , subscriptions = Subscriptions.init
                        }
                )
                request

        Nothing ->
            Task.succeed (Guest key)
