module Page.Project.Commit exposing (..)

import Context exposing (Context)
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


init : Context -> Session msg -> Project -> Commit.Hash -> Maybe CommitRoute.Route -> Task PageLoadError ( Model, Cmd Msg )
init context session project hash maybeRoute =
    let
        maybeAuthToken =
            Maybe.map .token session.user

        loadCommit =
            maybeAuthToken
                |> Request.Commit.get context project.slug hash

        loadTasks =
            maybeAuthToken
                |> Request.Commit.tasks context project.slug hash

        loadBuilds =
            maybeAuthToken
                |> Request.Commit.builds context project.slug hash

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
            |> Task.map
                (\successModel ->
                    case maybeRoute of
                        Just route ->
                            update context project session (SetRoute maybeRoute) successModel

                        Nothing ->
                            ( successModel, Cmd.none )
                )
            |> Task.mapError handleLoadError



-- CHANNELS --


channelName : Project.Slug -> String
channelName projectSlug =
    "project:" ++ (Project.slugToString projectSlug)


mapEvents :
    (b -> c)
    -> Dict comparable (List ( a1, a -> b ))
    -> Dict comparable (List ( a1, a -> c ))
mapEvents fromMsg events =
    events
        |> Dict.map (\_ v -> List.map (Tuple.mapSecond (\msg -> msg >> fromMsg)) v)


initialEvents : Project.Slug -> CommitRoute.Route -> Dict String (List ( String, Encode.Value -> Msg ))
initialEvents projectSlug route =
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
        Dict.singleton (channelName projectSlug) (pageEvents)


loadedEvents : Msg -> Model -> Dict String (List ( String, Encode.Value -> Msg ))
loadedEvents msg model =
    case msg of
        CommitTaskLoaded (Ok subModel) ->
            CommitTask.events subModel
                |> mapEvents CommitTaskMsg

        _ ->
            Dict.empty


leaveChannels : Model -> Maybe CommitRoute.Route -> List String
leaveChannels model maybeCommitRoute =
    case ( getSubPage model.subPageState, maybeCommitRoute ) of
        ( CommitTask subModel, Just (CommitRoute.Task _ buildName) ) ->
            CommitTask.leaveChannels subModel

        ( CommitTask subModel, _ ) ->
            CommitTask.leaveChannels subModel

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
    { b | slug : Project.Slug }
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
            Route.Project project.slug <| ProjectRoute.Commit commit.hash CommitRoute.Overview

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
            , [ ( CommitRoute.Overview |> ProjectRoute.Commit commit.hash |> Route.Project project.slug
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
                        t.id == b.task.id

                    _ ->
                        False
            )


setRoute : Context -> Session msg -> Project -> Maybe CommitRoute.Route -> Model -> ( Model, Cmd Msg )
setRoute context session project maybeRoute model =
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
                                        |> CommitTask.init context session project.id model.commit.hash task maybeTab
                                        |> transition CommitTaskLoaded

                                Nothing ->
                                    errored Page.Project "Could not find task"

                    Nothing ->
                        errored Page.Project "Uhoh"

            _ ->
                { model | subPageState = Loaded Blank }
                    => Cmd.none


update : Context -> Project -> Session msg -> Msg -> Model -> ( Model, Cmd Msg )
update context project session msg model =
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
                setRoute context session project route model

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
                    builds =
                        List.filter (\b -> b.task.id == subModel.task.id) model.builds

                    ( ( newModel, newCmd ), externalMsg ) =
                        CommitTask.update context project model.commit builds session subMsg subModel

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
                        buildJson
                            |> Decode.decodeValue Build.decoder
                            |> Result.toMaybe
                            |> Maybe.map (addBuild model.builds)
                            |> Maybe.withDefault model.builds
                            |> sortByDatetime .createdAt
                            |> List.reverse
                in
                    { model | builds = builds }
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
