module Page.Project exposing (..)

import Page.Errored as Errored exposing (PageLoadError, pageLoadError)
import Request.Project
import Task exposing (Task)
import Data.Session as Session exposing (Session)
import Data.Project as Project exposing (Project)
import Data.Branch as Branch exposing (Branch)
import Html exposing (..)
import Html.Attributes exposing (..)
import Views.Page as Page
import Util exposing ((=>), viewIf)
import Http
import Route exposing (Route)
import Views.Page as Page exposing (ActivePage)
import Page.Project.Route as ProjectRoute
import Page.Project.Commits as Commits
import Page.Project.Settings as Settings
import Page.Project.Commit as Commit
import Page.Project.Overview as Overview
import Page.Project.Task as CommitTask
import Views.Helpers exposing (onClickPage)
import Navigation exposing (newUrl)


--import Navigation exposing (newUrl)
-- SUB PAGES --


type SubPage
    = Blank
    | Overview
    | Commits Commits.Model
    | Commit Commit.Model
    | CommitTask CommitTask.Model
    | Settings Settings.Model
    | Errored PageLoadError


type SubPageState
    = Loaded SubPage
    | TransitioningFrom SubPage



-- MODEL --


type alias Model =
    { project : Project
    , branches : List Branch
    , subPageState : SubPageState
    }


initialSubPage : SubPage
initialSubPage =
    Blank


init : Session -> Project.Id -> Maybe ProjectRoute.Route -> Task PageLoadError ( Model, Cmd Msg )
init session id maybeRoute =
    let
        maybeAuthToken =
            Maybe.map .token session.user

        loadProject =
            maybeAuthToken
                |> Request.Project.get id
                |> Http.toTask

        loadBranches =
            maybeAuthToken
                |> Request.Project.branches id
                |> Http.toTask

        initialModel project branches =
            { project = project
            , branches = branches
            , subPageState = Loaded initialSubPage
            }

        handleLoadError _ =
            pageLoadError Page.Project "Project unavailable."
    in
        Task.map2 initialModel loadProject loadBranches
            |> Task.andThen
                (\successModel ->
                    case maybeRoute of
                        Just route ->
                            update session (SetRoute maybeRoute) successModel
                                |> Task.succeed

                        Nothing ->
                            ( successModel, Cmd.none )
                                |> Task.succeed
                )
            |> Task.mapError handleLoadError



-- VIEW --


type ActiveSubPage
    = OtherPage
    | OverviewPage
    | CommitsPage
    | SettingsPage


view : Session -> Model -> Html Msg
view session model =
    let
        ( subPageFrame, breadcrumb ) =
            viewSubPage session model
    in
        div []
            [ breadcrumb
            , div [ class "container-fluid" ] [ subPageFrame ]
            ]


viewSidebar : Project -> ActiveSubPage -> Html Msg
viewSidebar project subPage =
    nav [ class "col-sm-3 col-md-2 d-none d-sm-block bg-light sidebar" ]
        [ ul [ class "nav nav-pills flex-column" ]
            [ sidebarLink (subPage == OverviewPage)
                (Route.Project project.id ProjectRoute.Overview)
                [ i [ attribute "aria-hidden" "true", class "fa fa-home" ] [], text " Overview" ]
            , sidebarLink (subPage == CommitsPage)
                (Route.Project project.id (ProjectRoute.Commits Nothing Nothing))
                [ i [ attribute "aria-hidden" "true", class "fa fa-list" ] [], text " Commits" ]
            , sidebarLink (subPage == SettingsPage)
                (Route.Project project.id ProjectRoute.Settings)
                [ i [ attribute "aria-hidden" "true", class "fa fa-cog" ] [], text " Settings" ]
            ]
        ]


sidebarLink : Bool -> Route -> List (Html Msg) -> Html Msg
sidebarLink isActive route linkContent =
    li [ class "nav-item" ]
        [ a
            [ class "nav-link"
            , Route.href route
            , classList [ ( "active", isActive ) ]
            , onClickPage NewUrl route
            ]
            (linkContent ++ [ Util.viewIf isActive (span [ class "sr-only" ] [ text "(current)" ]) ])
        ]


viewBreadcrumb : Project -> Html Msg -> List ( Route, String ) -> Html Msg
viewBreadcrumb project additionalElements items =
    let
        fixedItems =
            [ ( Route.Projects, "Projects" )
            , ( Route.Project project.id ProjectRoute.Overview, project.name )
            ]

        allItems =
            fixedItems ++ items

        allItemLength =
            List.length allItems

        breadcrumbItem i item =
            item |> viewBreadcrumbItem (i == (allItemLength - 1))

        itemElements =
            List.indexedMap breadcrumbItem allItems
    in
        div [ class "d-flex justify-content-start align-items-center bg-dark breadcrumb-container" ]
            [ div [ class "p-2" ]
                [ ol [ class "breadcrumb bg-dark", style [ ( "margin", "0" ) ] ] itemElements ]
            , additionalElements
            ]


viewBreadcrumbItem : Bool -> ( Route, String ) -> Html Msg
viewBreadcrumbItem active ( route, name ) =
    let
        children =
            if active then
                text name
            else
                a [ Route.href route ] [ text name ]
    in
        li
            [ Route.href route
            , onClickPage NewUrl route
            , class "breadcrumb-item"
            , classList [ ( "active", active ) ]
            ]
            [ children ]


viewSubPage : Session -> Model -> ( Html Msg, Html Msg )
viewSubPage session model =
    let
        page =
            case model.subPageState of
                Loaded page ->
                    page

                TransitioningFrom page ->
                    page

        project =
            model.project

        branches =
            model.branches

        breadcrumb =
            viewBreadcrumb project

        sidebar =
            viewSidebar project
    in
        case page of
            Blank ->
                let
                    content =
                        Html.text ""
                            |> frame (sidebar OtherPage)
                in
                    ( content, breadcrumb (text "") [] )

            Errored subModel ->
                let
                    content =
                        Html.text "Errored"
                            |> frame (sidebar OtherPage)
                in
                    ( content, breadcrumb (text "") [] )

            Overview ->
                let
                    content =
                        Overview.view project
                            |> frame (sidebar OverviewPage)
                in
                    ( content, breadcrumb (text "") [] )

            Commits subModel ->
                let
                    content =
                        Commits.view project branches subModel
                            |> Html.map CommitsMsg
                            |> frame (sidebar CommitsPage)

                    crumbExtraItems =
                        Commits.viewBreadcrumbExtraItems subModel
                            |> Html.map CommitsMsg

                    crumb =
                        Commits.breadcrumb project
                            |> breadcrumb crumbExtraItems
                in
                    ( content, crumb )

            Commit subModel ->
                let
                    content =
                        Commit.view project subModel
                            |> Html.map CommitMsg
                            |> frame (sidebar CommitsPage)

                    crumb =
                        Commit.breadcrumb project subModel.commit
                            |> breadcrumb (text "")
                in
                    ( content, crumb )

            CommitTask subModel ->
                let
                    content =
                        CommitTask.view subModel
                            |> Html.map CommitTaskMsg
                            |> frame (sidebar CommitsPage)

                    crumb =
                        CommitTask.breadcrumb project subModel.commit subModel.task
                            |> breadcrumb (text "")
                in
                    ( content, crumb )

            Settings subModel ->
                let
                    content =
                        Settings.view project subModel
                            |> Html.map SettingsMsg
                            |> frame (sidebar SettingsPage)

                    crumb =
                        Settings.breadcrumb project
                            |> breadcrumb (text "")
                in
                    ( content, crumb )


frame : Html msg -> Html msg -> Html msg
frame sidebar content =
    div [ class "row" ]
        [ sidebar
        , div [ class "col-sm-9 ml-sm-auto col-md-10 pt-3 project-content-container" ] [ content ]
        ]



-- UPDATE --


type Msg
    = NewUrl String
    | SetRoute (Maybe ProjectRoute.Route)
    | CommitsMsg Commits.Msg
    | CommitsLoaded (Result PageLoadError Commits.Model)
    | CommitMsg Commit.Msg
    | CommitLoaded (Result PageLoadError Commit.Model)
    | CommitTaskMsg CommitTask.Msg
    | CommitTaskLoaded (Result PageLoadError CommitTask.Model)
    | SettingsMsg Settings.Msg
    | NoOp


getSubPage : SubPageState -> SubPage
getSubPage subPageState =
    case subPageState of
        Loaded subPage ->
            subPage

        TransitioningFrom subPage ->
            subPage


setRoute : Session -> Maybe ProjectRoute.Route -> Model -> ( Model, Cmd Msg )
setRoute session maybeRoute model =
    let
        transition toMsg task =
            { model | subPageState = TransitioningFrom (getSubPage model.subPageState) }
                => Task.attempt toMsg task

        errored =
            pageErrored model
    in
        case maybeRoute of
            Nothing ->
                { model | subPageState = Loaded Blank } => Cmd.none

            Just (ProjectRoute.Overview) ->
                case session.user of
                    Just user ->
                        { model | subPageState = Loaded Overview } => Cmd.none

                    Nothing ->
                        errored Page.Project "Uhoh"

            Just (ProjectRoute.Commits maybeBranch maybePage) ->
                case session.user of
                    Just user ->
                        Commits.init session model.project.id maybeBranch maybePage
                            |> transition CommitsLoaded

                    Nothing ->
                        errored Page.Project "Uhoh"

            Just (ProjectRoute.Commit hash) ->
                case session.user of
                    Just user ->
                        transition CommitLoaded (Commit.init session model.project.id hash)

                    Nothing ->
                        errored Page.Project "Uhoh"

            Just (ProjectRoute.Task hash name) ->
                case session.user of
                    Just user ->
                        transition CommitTaskLoaded (CommitTask.init session model.project.id hash name)

                    Nothing ->
                        errored Page.Project "Uhoh"

            Just (ProjectRoute.Settings) ->
                case session.user of
                    Just user ->
                        { model | subPageState = Loaded (Settings (Settings.initialModel)) } => Cmd.none

                    Nothing ->
                        errored Page.Project "Uhoh"


pageErrored : Model -> ActivePage -> String -> ( Model, Cmd msg )
pageErrored model activePage errorMessage =
    let
        error =
            Errored.pageLoadError activePage errorMessage
    in
        { model | subPageState = Loaded (Errored error) } => Cmd.none


update : Session -> Msg -> Model -> ( Model, Cmd Msg )
update session msg model =
    updateSubPage session (getSubPage model.subPageState) msg model


updateSubPage : Session -> SubPage -> Msg -> Model -> ( Model, Cmd Msg )
updateSubPage session subPage msg model =
    let
        toPage toModel toMsg subUpdate subMsg subModel =
            let
                ( newModel, newCmd ) =
                    subUpdate subMsg subModel
            in
                ( { model | subPageState = Loaded (toModel newModel) }, Cmd.map toMsg newCmd )

        errored =
            pageErrored model
    in
        case ( msg, subPage ) of
            ( NewUrl url, _ ) ->
                model => newUrl url

            ( SetRoute route, _ ) ->
                setRoute session route model

            ( SettingsMsg subMsg, Settings subModel ) ->
                toPage Settings SettingsMsg (Settings.update model.project session) subMsg subModel

            ( CommitsLoaded (Ok subModel), _ ) ->
                { model | subPageState = Loaded (Commits subModel) }
                    => Cmd.none

            ( CommitsLoaded (Err error), _ ) ->
                { model | subPageState = Loaded (Errored error) }
                    => Cmd.none

            ( CommitsMsg subMsg, Commits subModel ) ->
                toPage Commits CommitsMsg (Commits.update model.project session) subMsg subModel

            ( CommitLoaded (Ok subModel), _ ) ->
                { model | subPageState = Loaded (Commit subModel) }
                    => Cmd.none

            ( CommitLoaded (Err error), _ ) ->
                { model | subPageState = Loaded (Errored error) }
                    => Cmd.none

            ( CommitMsg subMsg, Commit subModel ) ->
                toPage Commit CommitMsg (Commit.update model.project session) subMsg subModel

            ( CommitTaskLoaded (Ok subModel), _ ) ->
                { model | subPageState = Loaded (CommitTask subModel) }
                    => Cmd.none

            ( CommitTaskLoaded (Err error), _ ) ->
                { model | subPageState = Loaded (Errored error) }
                    => Cmd.none

            ( CommitTaskMsg subMsg, CommitTask subModel ) ->
                toPage CommitTask CommitTaskMsg (CommitTask.update model.project session) subMsg subModel

            ( _, _ ) ->
                -- Disregard incoming messages that arrived for the wrong sub page
                (Debug.log "Fell through" model)
                    => Cmd.none
