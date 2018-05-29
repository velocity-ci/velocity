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
import Views.Helpers exposing (onClickPage)
import Page.Project.Commit.Overview as Overview
import Page.Project.Commit.Task as CommitTask
import Data.PaginatedList exposing (Paginated(..))
import Json.Encode as Encode
import Dict exposing (Dict)
import Page.Helpers exposing (sortByDatetime, formatDateTime)
import Views.Spinner exposing (spinner)
import Component.DropdownFilter as DropdownFilter
import Dom


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
    , taskFilterTerm : String
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
            , taskFilterTerm = ""
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



-- SUBSCRIPTIONS --


subscriptions : Model -> Sub Msg
subscriptions model =
    case (getSubPage model.subPageState) of
        CommitTask subModel ->
            CommitTask.subscriptions subModel
                |> Sub.map CommitTaskMsg

        _ ->
            Sub.none



-- CHANNELS --


mapEvents :
    (b -> c)
    -> Dict comparable (List ( a1, a -> b ))
    -> Dict comparable (List ( a1, a -> c ))
mapEvents fromMsg events =
    events
        |> Dict.map (\_ v -> List.map (Tuple.mapSecond (\msg -> msg >> fromMsg)) v)


initialEvents : String -> CommitRoute.Route -> Dict String (List ( String, Encode.Value -> Msg ))
initialEvents channelName route =
    let
        subPageEvents =
            case route of
                CommitRoute.Task taskName maybeBuildName ->
                    Dict.empty

                _ ->
                    Dict.empty

        pageEvents =
            []
    in
        Dict.singleton channelName pageEvents


loadedEvents : Msg -> Model -> Dict String (List ( String, Encode.Value -> Msg ))
loadedEvents msg model =
    case msg of
        CommitTaskLoaded (Ok subModel) ->
            CommitTask.events subModel
                |> mapEvents CommitTaskMsg

        _ ->
            Dict.empty


leaveChannels : Model -> ProjectRoute.Route -> List String
leaveChannels model route =
    let
        leave =
            leaveSubPageChannels (getSubPage model.subPageState)
    in
        case route of
            ProjectRoute.Commit _ subRoute ->
                leave (Just subRoute)

            -- Not a commit route
            _ ->
                leave Nothing


leaveSubPageChannels : SubPage -> Maybe CommitRoute.Route -> List String
leaveSubPageChannels subPage subRoute =
    case subPage of
        CommitTask subModel ->
            CommitTask.leaveChannels subModel subRoute

        _ ->
            []



-- VIEW --


view : Project -> Model -> Html Msg
view project model =
    div []
        [ viewTaskTabs project model.commit model.tasks model.subPageState
        , viewSubPage project model
        ]


viewSubPage : Project -> Model -> Html Msg
viewSubPage project model =
    case getSubPage model.subPageState of
        Overview _ ->
            Overview.view project model.commit model.tasks model.builds
                |> frame project model OverviewMsg

        CommitTask subModel ->
            taskBuilds model.builds (Just subModel.task)
                |> CommitTask.view project model.commit subModel
                |> frame project model CommitTaskMsg

        _ ->
            div [ class "d-flex justify-content-center" ] [ spinner ]


viewTaskTabs : Project -> Commit -> List ProjectTask.Task -> SubPageState -> Html Msg
viewTaskTabs project commit tasks subPage =
    let
        toRoute =
            taskRoute project commit
    in
        ul [ class "nav nav-tabs" ] <| List.map (viewTaskTab toRoute subPage) tasks


taskRoute : Project -> Commit -> ProjectTask.Task -> Route
taskRoute project commit task =
    CommitRoute.Task task.name Nothing
        |> ProjectRoute.Commit commit.hash
        |> Route.Project project.slug


viewTaskTab : (ProjectTask.Task -> Route) -> SubPageState -> ProjectTask.Task -> Html Msg
viewTaskTab toRoute subPage task =
    let
        route =
            toRoute task

        active =
            case subPage of
                Loaded (CommitTask subModel) ->
                    subModel.task.id == task.id

                _ ->
                    False

        classes =
            [ "active" => active ]
    in
        li [ class "nav-item" ]
            [ a
                [ class "nav-link"
                , classList classes
                , Route.href route
                , onClickPage NewUrl route
                ]
                [ text <| ProjectTask.nameToString task.name ]
            ]


frame :
    Project
    -> Model
    -> (a -> Msg)
    -> Html a
    -> Html Msg
frame project { commit, tasks } toMsg content =
    Html.map toMsg content


viewCommitDetailsIcon : Commit -> String -> (Commit -> String) -> Html Msg
viewCommitDetailsIcon commit iconClass fn =
    li []
        [ i
            [ attribute "aria-hidden" "true"
            , classList
                [ ( "fa", True )
                , ( iconClass, True )
                ]
            ]
            []
        , text " "
        , fn commit |> text
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
    | AddBuild Build
    | UpdateBuild Build
    | NoOp


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

            Just (CommitRoute.Task name maybeBuildId) ->
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
                                        |> CommitTask.init context session project.id model.commit.hash task maybeBuildId
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

            ( AddBuild build, _ ) ->
                let
                    builds =
                        --                        if Commit.compare model.commit build.task.commit then
                        build
                            |> addBuild model.builds
                            |> sortByDatetime .createdAt
                            |> List.reverse

                    --                        else
                    --                            model.builds
                in
                    { model | builds = builds }
                        => Cmd.none

            ( UpdateBuild build, _ ) ->
                let
                    builds =
                        model.builds
                            |> List.map
                                (\a ->
                                    if build.id == a.id then
                                        build
                                    else
                                        a
                                )
                in
                    { model | builds = builds }
                        => Cmd.none

            ( _, _ ) ->
                -- Disregard incoming messages that arrived for the wrong sub page
                (Debug.log "Fell through (commit page)" model)
                    => Cmd.none
