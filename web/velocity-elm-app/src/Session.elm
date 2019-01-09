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
        , viewer
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
import Api.Subscriptions as Subscriptions


-- TYPES


type Session msg
    = LoggedIn (LoggedInInternals msg)
    | Guest Nav.Key


type alias LoggedInInternals msg =
    { navKey : Nav.Key
    , viewer : Viewer
    , projects : List Project
    , branches : Project.Id.Dict (List Branch)
    , knownHosts : List KnownHost
    , log : Activity.Log
    , subscriptions : Subscriptions.State msg
    }


type InitError
    = HttpError (Graphql.Http.Error StartupResponse)



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
            Project.findProjectById internals.projects projectId

        Guest _ ->
            Nothing


projectWithSlug : Project.Slug.Slug -> Session msg -> Maybe Project
projectWithSlug projectSlug session =
    case session of
        LoggedIn internals ->
            Project.findProjectBySlug internals.projects projectSlug

        Guest _ ->
            Nothing


projects : Session msg -> List Project
projects session =
    case session of
        LoggedIn internals ->
            internals.projects

        Guest _ ->
            []


branches : Project.Id.Id -> Session msg -> List Branch
branches projectId session =
    case session of
        LoggedIn internals ->
            Project.Id.get projectId internals.branches
                |> Maybe.withDefault []

        Guest _ ->
            []


addProject : Project -> Session msg -> Session msg
addProject p session =
    case session of
        LoggedIn internals ->
            LoggedIn { internals | projects = Project.addProject p internals.projects }

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



-- CHANGES


changes : (Task InitError (Session msg) -> msg2) -> Context msg -> Session msg -> Sub msg2
changes toMsg context session =
    Api.viewerChanges (fromViewer (navKey session) context >> toMsg) Viewer.decoder


type alias StartupResponse =
    { projects : List Project
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
                            , subscriptions = Subscriptions.init
                            }
                    )
                    request

        Nothing ->
            Task.succeed (Guest key)
