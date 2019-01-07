port module Session
    exposing
        ( InitError
        , Session
        , addKnownHost
        , addProject
        , branches
        , changes
        , cred
        , fromViewer
        , knownHosts
        , log
        , navKey
        , projectWithId
        , projectWithSlug
        , projects
        , updateSubscriptionStatus
        , viewer
        , SubscriptionStatus(..)
        , knownHostSubscription
        )

import Activity
import Api exposing (BaseUrl, Cred)
import Browser.Navigation as Nav
import Context exposing (Context)
import Http
import Json.Decode as Decode
import KnownHost exposing (KnownHost)
import Project exposing (Project)
import Project.Branch as Branch exposing (Branch)
import Project.Id
import Project.Slug
import Task exposing (Task)
import Viewer exposing (Viewer)
import Graphql.Http
import Graphql.Http.GraphqlError as GraphqlError
import Api.Compiled.Query as Query
import Graphql.SelectionSet as SelectionSet exposing (SelectionSet)
import Set
import Api.Compiled.Subscription as Subscription
import Graphql.Operation exposing (RootSubscription)
import Graphql.Document


-- TYPES


type Session
    = LoggedIn LoggedInInternals SubscriptionStatus
    | Guest Nav.Key SubscriptionStatus


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


type SubscriptionStatus
    = NotConnected
    | Connected
    | Reconnecting


updateSubscriptionStatus : SubscriptionStatus -> Session -> Session
updateSubscriptionStatus status session =
    case session of
        LoggedIn internals _ ->
            LoggedIn internals status

        Guest nav _ ->
            Guest nav status



-- COLLECTIONS


knownHosts : Session -> List KnownHost
knownHosts session =
    case session of
        LoggedIn internals _ ->
            internals.knownHosts

        Guest _ _ ->
            []


addKnownHost : KnownHost -> Session -> Session
addKnownHost knownHost session =
    case session of
        LoggedIn internals status ->
            LoggedIn { internals | knownHosts = KnownHost.addKnownHost internals.knownHosts knownHost } status

        Guest _ _ ->
            session


projectWithId : Project.Id.Id -> Session -> Maybe Project
projectWithId projectId session =
    case session of
        LoggedIn internals _ ->
            Project.findProjectById internals.projects projectId

        Guest _ _ ->
            Nothing


projectWithSlug : Project.Slug.Slug -> Session -> Maybe Project
projectWithSlug projectSlug session =
    case session of
        LoggedIn internals _ ->
            Project.findProjectBySlug internals.projects projectSlug

        Guest _ _ ->
            Nothing


projects : Session -> List Project
projects session =
    case session of
        LoggedIn internals _ ->
            internals.projects

        Guest _ _ ->
            []


branches : Project.Id.Id -> Session -> List Branch
branches projectId session =
    case session of
        LoggedIn internals _ ->
            Project.Id.get projectId internals.branches
                |> Maybe.withDefault []

        Guest _ _ ->
            []


addProject : Project -> Session -> Session
addProject p session =
    case session of
        LoggedIn internals status ->
            LoggedIn { internals | projects = Project.addProject p internals.projects } status

        Guest _ _ ->
            session



-- INFO


viewer : Session -> Maybe Viewer
viewer session =
    case session of
        LoggedIn internals _ ->
            Just internals.viewer

        Guest _ _ ->
            Nothing


cred : Session -> Maybe Cred
cred session =
    case session of
        LoggedIn internals _ ->
            Just (Viewer.cred internals.viewer)

        Guest _ _ ->
            Nothing


navKey : Session -> Nav.Key
navKey session =
    case session of
        LoggedIn internals _ ->
            internals.navKey

        Guest key _ ->
            key


log : Session -> Maybe Activity.Log
log session =
    case session of
        LoggedIn internals _ ->
            Just internals.log

        Guest _ _ ->
            Nothing



-- CHANGES


knownHostSubscription : SelectionSet KnownHost RootSubscription
knownHostSubscription =
    Subscription.knownHostAdded KnownHost.selectionSet


changes : (Task InitError Session -> msg) -> Context msg2 -> Session -> Sub msg
changes toMsg context session =
    Api.viewerChanges (fromViewer (navKey session) context >> toMsg) Viewer.decoder


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
                        LoggedIn
                            { navKey = key
                            , viewer = viewerVal
                            , projects = res.projects
                            , branches = Project.Id.empty
                            , knownHosts = res.knownHosts
                            , log = Activity.init
                            }
                            NotConnected
                    )
                    request

        Nothing ->
            Task.succeed (Guest key NotConnected)
