module Page.Project.Commit exposing (..)

import Html exposing (..)
import Html.Attributes exposing (..)
import Data.Commit as Commit exposing (Commit)
import Data.Session as Session exposing (Session)
import Data.Project as Project exposing (Project)
import Data.Task as ProjectTask
import Data.Build as Build exposing (Build)
import Page.Errored as Errored exposing (PageLoadError, pageLoadError)
import Page.Helpers exposing (formatDateTime, sortByDatetime)
import Page.Project.Commits as Commits
import Request.Project
import Request.Commit
import Util exposing ((=>))
import Task exposing (Task)
import Views.Page as Page
import Http
import Route exposing (Route)
import Page.Project.Route as ProjectRoute
import Page.Project.Commit.Route as CommitRoute
import Navigation
import Views.Helpers exposing (onClickPage)
import Views.Page as Page exposing (ActivePage)
import Page.Project.Commit.Overview as Overview
import Page.Project.Commit.Task as CommitTask


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


init : Session -> Project -> Commit.Hash -> Maybe CommitRoute.Route -> Task PageLoadError ( Model, Cmd Msg )
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

        initialModel commit tasks builds =
            { commit = commit
            , tasks = tasks
            , builds = builds
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



-- VIEW --


view : Project -> Model -> Html Msg
view project model =
    case getSubPage model.subPageState of
        Overview _ ->
            Overview.view project model.commit model.tasks
                |> Html.map OverviewMsg

        CommitTask subModel ->
            CommitTask.view subModel
                |> Html.map CommitTaskMsg

        _ ->
            Html.text "Nope"


breadcrumb : Project -> Commit -> List ( Route, String )
breadcrumb project commit =
    List.concat
        [ Commits.breadcrumb project
        , [ ( Route.Project project.id (ProjectRoute.Commit commit.hash CommitRoute.Overview), Commit.truncateHash commit.hash ) ]
        ]



-- UPDATE --


type Msg
    = NewUrl String
    | SetRoute (Maybe CommitRoute.Route)
    | OverviewMsg Overview.Msg
    | CommitTaskMsg CommitTask.Msg
    | CommitTaskLoaded (Result PageLoadError CommitTask.Model)


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


setRoute : Session -> Project -> Maybe CommitRoute.Route -> Model -> ( Model, Cmd Msg )
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
                        { model | subPageState = Overview.initialModel |> Overview |> Loaded } => Cmd.none

                    Nothing ->
                        errored Page.Project "Uhoh"

            Just (CommitRoute.Task name) ->
                case session.user of
                    Just user ->
                        CommitTask.init session project.id model.commit.hash name
                            |> transition CommitTaskLoaded

                    Nothing ->
                        errored Page.Project "Uhoh"

            _ ->
                { model | subPageState = Loaded Blank } => Cmd.none


update : Project -> Session -> Msg -> Model -> ( Model, Cmd Msg )
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
    in
        case ( msg, subPage ) of
            ( NewUrl url, _ ) ->
                model => Navigation.newUrl url

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
                toPage CommitTask CommitTaskMsg (CommitTask.update project model.commit session) subMsg subModel

            ( _, _ ) ->
                -- Disregard incoming messages that arrived for the wrong sub page
                (Debug.log "Fell through" model)
                    => Cmd.none
