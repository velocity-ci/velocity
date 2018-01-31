module Page.Project.Commit exposing (..)

import Html exposing (..)
import Html.Attributes exposing (..)
import Data.Commit as Commit exposing (Commit)
import Data.Session as Session exposing (Session)
import Data.Project as Project exposing (Project)
import Data.Task as ProjectTask
import Data.Build as Build exposing (Build, addBuild)
import Page.Errored as Errored exposing (PageLoadError, pageLoadError)
import Page.Project.Commits as Commits
import Request.Commit
import Util exposing ((=>))
import Task exposing (Task)
import Views.Page as Page
import Http
import Route exposing (Route)
import Page.Project.Route as ProjectRoute
import Page.Project.Commit.Route as CommitRoute
import Navigation
import Views.Page as Page exposing (ActivePage)
import Page.Project.Commit.Overview as Overview
import Page.Project.Commit.Task as CommitTask
import Data.PaginatedList exposing (Paginated(..))
import Json.Encode as Encode
import Json.Decode as Decode
import Dict exposing (Dict)
import Page.Helpers exposing (sortByDatetime)
import Views.Helpers exposing (onClickPage)


-- SUB PAGES --


type SubPage
    = Blank
    | Overview Overview.Model
    | Errored PageLoadError
    | CommitTask CommitTask.Model


type SubPageState
    = Loaded SubPage
    | TransitioningFrom SubPage



-- MODEL --


type alias Model =
    { commit : Commit
    , tasks : List ProjectTask.Task
    , builds : List Build
    , subPageState : SubPageState
    }


initialSubPage : SubPage
initialSubPage =
    Blank


init : Session msg -> Project -> Commit.Hash -> Maybe CommitRoute.Route -> Task PageLoadError ( Model, Cmd Msg )
init session project hash maybeRoute =
    let
        maybeAuthToken =
            Maybe.map .token session.user

        loadCommit =
            maybeAuthToken
                |> Request.Commit.get project.id hash
                |> Http.toTask

        loadTasks =
            maybeAuthToken
                |> Request.Commit.tasks project.id hash
                |> Http.toTask

        loadBuilds =
            maybeAuthToken
                |> Request.Commit.builds project.id hash
                |> Http.toTask

        initialModel commit (Paginated tasks) (Paginated builds) =
            { commit = commit
            , tasks = tasks.results
            , builds = sortByDatetime .createdAt builds.results |> List.reverse
            , subPageState = Loaded initialSubPage
            }

        handleLoadError _ =
            pageLoadError Page.Project "Project unavailable."
    in
        Task.map3 initialModel loadCommit loadTasks loadBuilds
            |> Task.andThen
                (\successModel ->
                    case maybeRoute of
                        Just route ->
                            update project session (SetRoute maybeRoute) successModel
                                |> Task.succeed

                        Nothing ->
                            ( successModel, Cmd.none )
                                |> Task.succeed
                )
            |> Task.mapError handleLoadError



-- CHANNELS --


channelName : Project.Id -> String
channelName projectId =
    "project:" ++ (Project.idToString projectId)


mapEvents :
    (b -> c)
    -> Dict comparable (List ( a1, a -> b ))
    -> Dict comparable (List ( a1, a -> c ))
mapEvents fromMsg events =
    events
        |> Dict.map (\_ v -> List.map (Tuple.mapSecond (\msg -> msg >> fromMsg)) v)


initialEvents : Project.Id -> CommitRoute.Route -> Dict String (List ( String, Encode.Value -> Msg ))
initialEvents projectId route =
    let
        subPageEvents =
            case route of
                CommitRoute.Task taskName maybeBuildName ->
                    Dict.empty

                _ ->
                    Dict.empty

        pageEvents =
            [ ( "build:new", AddBuildEvent )
            , ( "build:delete", DeleteBuildEvent )
            , ( "build:update", UpdateBuildEvent )
            ]
    in
        Dict.singleton (channelName projectId) (pageEvents)


loadedEvents : Msg -> Model -> Dict String (List ( String, Encode.Value -> Msg ))
loadedEvents msg model =
    case msg of
        CommitTaskLoaded (Ok subModel) ->
            CommitTask.events model.builds subModel
                |> mapEvents CommitTaskMsg

        _ ->
            Dict.empty


leaveChannels : Model -> Maybe CommitRoute.Route -> List String
leaveChannels model maybeCommitRoute =
    case ( getSubPage model.subPageState, maybeCommitRoute ) of
        ( CommitTask subModel, Just (CommitRoute.Task _ buildName) ) ->
            CommitTask.leaveChannels model.builds subModel buildName

        ( CommitTask subModel, _ ) ->
            CommitTask.leaveChannels model.builds subModel Nothing

        _ ->
            []



-- VIEW --


view : Project -> Model -> Html Msg
view project model =
    case getSubPage model.subPageState of
        Overview _ ->
            Overview.view project model.commit model.tasks model.builds
                |> frame project model.commit OverviewMsg

        CommitTask subModel ->
            taskBuilds model.builds (Just subModel.task)
                |> CommitTask.view project model.commit subModel
                |> frame project model.commit CommitTaskMsg

        _ ->
            Html.text "Nope"


frame :
    { b | id : Project.Id }
    -> { c | hash : Commit.Hash }
    -> (a -> Msg)
    -> Html a
    -> Html Msg
frame project commit toMsg content =
    let
        commitTitle =
            commit.hash
                |> Commit.truncateHash
                |> String.append "Commit "
                |> text

        route =
            Route.Project project.id <| ProjectRoute.Commit commit.hash CommitRoute.Overview

        link =
            a [ Route.href route, onClickPage NewUrl route ] [ commitTitle ]

        commitTitleStyle =
            [ ( "position", "absolute" )
            , ( "top", "2rem" )
            , ( "right", "1rem" )
            ]
    in
        div []
            [ h2 [ style commitTitleStyle, class "display-7" ] [ link ]
            , Html.map toMsg content
            ]


breadcrumb : Project -> Commit -> SubPageState -> List ( Route, String )
breadcrumb project commit subPageState =
    let
        subPage =
            getSubPage subPageState

        subPageCrumb =
            case subPage of
                CommitTask subModel ->
                    CommitTask.breadcrumb project commit subModel.task

                _ ->
                    []
    in
        List.concat
            [ Commits.breadcrumb project
            , [ ( CommitRoute.Overview |> ProjectRoute.Commit commit.hash |> Route.Project project.id
                , Commit.truncateHash commit.hash
                )
              ]
            , subPageCrumb
            ]



-- UPDATE --


type Msg
    = NewUrl String
    | SetRoute (Maybe CommitRoute.Route)
    | OverviewMsg Overview.Msg
    | CommitTaskMsg CommitTask.Msg
    | CommitTaskLoaded (Result PageLoadError CommitTask.Model)
    | AddBuildEvent Encode.Value
    | UpdateBuildEvent Encode.Value
    | DeleteBuildEvent Encode.Value


getSubPage : SubPageState -> SubPage
getSubPage subPageState =
    case subPageState of
        Loaded subPage ->
            subPage

        TransitioningFrom subPage ->
            subPage


pageErrored : Model -> ActivePage -> String -> ( Model, Cmd msg )
pageErrored model activePage errorMessage =
    let
        error =
            Errored.pageLoadError activePage errorMessage
    in
        { model | subPageState = Loaded (Errored error) } => Cmd.none


taskBuilds : List Build -> Maybe ProjectTask.Task -> List Build
taskBuilds builds maybeTask =
    builds
        |> List.filter
            (\b ->
                case maybeTask of
                    Just t ->
                        t.id == b.taskId

                    _ ->
                        False
            )


setRoute : Session msg -> Project -> Maybe CommitRoute.Route -> Model -> ( Model, Cmd Msg )
setRoute session project maybeRoute model =
    let
        transition toMsg task =
            { model | subPageState = TransitioningFrom (getSubPage model.subPageState) }
                => Task.attempt toMsg task

        errored =
            pageErrored model
    in
        case maybeRoute of
            Just (CommitRoute.Overview) ->
                case session.user of
                    Just user ->
                        { model | subPageState = Overview.initialModel |> Overview |> Loaded }
                            => Cmd.none

                    Nothing ->
                        errored Page.Project "Uhoh"

            Just (CommitRoute.Task name maybeTab) ->
                case session.user of
                    Just user ->
                        let
                            maybeTask =
                                model.tasks
                                    |> List.filter (\t -> t.name == name)
                                    |> List.head
                        in
                            case maybeTask of
                                Just task ->
                                    taskBuilds model.builds (Just task)
                                        |> CommitTask.init session project.id model.commit.hash task maybeTab
                                        |> transition CommitTaskLoaded

                                Nothing ->
                                    errored Page.Project "Could not find task"

                    Nothing ->
                        errored Page.Project "Uhoh"

            _ ->
                { model | subPageState = Loaded Blank }
                    => Cmd.none


update : Project -> Session msg -> Msg -> Model -> ( Model, Cmd Msg )
update project session msg model =
    let
        toPage toModel toMsg subUpdate subMsg subModel =
            let
                ( newModel, newCmd ) =
                    subUpdate subMsg subModel
            in
                ( { model | subPageState = Loaded (toModel newModel) }, Cmd.map toMsg newCmd )

        subPage =
            getSubPage model.subPageState

        errored =
            pageErrored model

        findBuild b =
            List.filter (\a -> a.id == b.id) model.builds
                |> List.head
    in
        case ( msg, subPage ) of
            ( NewUrl url, _ ) ->
                model
                    => Navigation.newUrl url

            ( SetRoute route, _ ) ->
                setRoute session project route model

            ( OverviewMsg subMsg, Overview subModel ) ->
                toPage Overview OverviewMsg (Overview.update project session) subMsg subModel

            ( CommitTaskLoaded (Ok subModel), _ ) ->
                { model | subPageState = Loaded (CommitTask subModel) }
                    => Cmd.none

            ( CommitTaskLoaded (Err error), _ ) ->
                { model | subPageState = Loaded (Errored error) }
                    => Cmd.none

            ( CommitTaskMsg subMsg, CommitTask subModel ) ->
                let
                    ( ( newModel, newCmd ), externalMsg ) =
                        CommitTask.update project model.commit model.builds session subMsg subModel

                    model_ =
                        case externalMsg of
                            CommitTask.AddBuild b ->
                                { model | builds = addBuild model.builds b }

                            CommitTask.UpdateBuild b ->
                                let
                                    builds =
                                        List.map
                                            (\c ->
                                                if c.id == b.id then
                                                    b
                                                else
                                                    c
                                            )
                                            model.builds
                                in
                                    { model | builds = builds }

                            CommitTask.NoOp ->
                                model
                in
                    { model_ | subPageState = Loaded (CommitTask newModel) }
                        ! [ Cmd.map CommitTaskMsg newCmd ]

            ( AddBuildEvent buildJson, _ ) ->
                let
                    builds =
                        Decode.decodeValue Build.decoder buildJson
                            |> Result.toMaybe
                            |> Maybe.map (addBuild model.builds)
                            |> Maybe.withDefault model.builds
                in
                    { model | builds = sortByDatetime .createdAt builds |> List.reverse }
                        => Cmd.none

            ( DeleteBuildEvent buildJson, _ ) ->
                let
                    builds =
                        Decode.decodeValue Build.decoder buildJson
                            |> Result.toMaybe
                            |> Maybe.map (\b -> List.filter (\a -> b.id /= a.id) model.builds)
                            |> Maybe.withDefault model.builds
                in
                    { model | builds = builds }
                        => Cmd.none

            ( UpdateBuildEvent buildJson, _ ) ->
                let
                    builds =
                        Decode.decodeValue Build.decoder buildJson
                            |> Result.toMaybe
                            |> Maybe.map
                                (\b ->
                                    List.map
                                        (\a ->
                                            if b.id == a.id then
                                                b
                                            else
                                                a
                                        )
                                        model.builds
                                )
                            |> Maybe.withDefault model.builds
                in
                    { model | builds = builds }
                        => Cmd.none

            ( _, _ ) ->
                -- Disregard incoming messages that arrived for the wrong sub page
                (Debug.log "Fell through" model)
                    => Cmd.none
