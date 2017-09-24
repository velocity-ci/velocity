module Page.Project exposing (..)

import Data.Project as Project exposing (Project)
import Page.Errored as Errored exposing (PageLoadError, pageLoadError)
import Request.Project
import Task exposing (Task)
import Data.Session as Session exposing (Session)
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


-- SUB PAGES --


type SubPage
    = Blank
    | Commits Commits.Model
    | Settings Settings.Model
    | Errored PageLoadError


type SubPageState
    = Loaded SubPage
    | TransitioningFrom SubPage



-- MODEL --


type alias Model =
    { project : Project
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

        initialModel project =
            { project = project
            , subPageState = Loaded initialSubPage
            }

        handleLoadError _ =
            pageLoadError Page.Project "Project unavailable."
    in
        Task.map initialModel loadProject
            |> Task.andThen
                (\successModel ->
                    case maybeRoute of
                        Just route ->
                            Task.succeed (update session (SetRoute maybeRoute) successModel)

                        Nothing ->
                            Task.succeed (successModel => Cmd.none)
                )
            |> Task.mapError handleLoadError



-- VIEW --


type ActiveSubPage
    = OtherPage
    | CommitsPage
    | SettingsPage


view : Session -> Model -> Html Msg
view session model =
    let
        project =
            model.project

        ( subPageFrame, breadcrumb ) =
            viewSubPage session model
    in
        div []
            [ breadcrumb
            , div [ class "container-fluid" ] [ subPageFrame ]
            ]


viewSidebar : Project -> ActiveSubPage -> Html msg
viewSidebar project subPage =
    nav [ class "col-sm-3 col-md-2 d-none d-sm-block bg-light sidebar" ]
        [ ul [ class "nav nav-pills flex-column" ]
            [ sidebarLink (subPage == CommitsPage) (Route.Project ProjectRoute.Commits project.id) [ text "Commits" ]
            , sidebarLink (subPage == SettingsPage) (Route.Project ProjectRoute.Settings project.id) [ text "Settings" ]
            ]
        ]


sidebarLink : Bool -> Route -> List (Html msg) -> Html msg
sidebarLink isActive route linkContent =
    li [ class "nav-item" ]
        [ a [ class "nav-link", Route.href route, classList [ ( "active", isActive ) ] ]
            (linkContent ++ [ Util.viewIf isActive (span [ class "sr-only" ] [ text "(current)" ]) ])
        ]


viewBreadcrumb : Project -> String -> Html Msg -> Html Msg
viewBreadcrumb project activeItem extraItems =
    div [ class "d-flex justify-content-start align-items-center bg-dark breadcrumb-container", style [ ( "height", "50px" ) ] ]
        [ div [ class "p-2" ]
            [ ol [ class "breadcrumb bg-dark", style [ ( "margin", "0" ) ] ]
                [ li [ class "breadcrumb-item" ] [ a [ Route.href Route.Projects ] [ text "Projects" ] ]
                , li [ class "breadcrumb-item active" ] [ text project.name ]
                , li [ class "breadcrumb-item active" ] [ text activeItem ]
                ]
            ]
        , extraItems
        ]


viewSubPage : Session -> Model -> ( Html Msg, Html Msg )
viewSubPage session model =
    let
        page =
            case model.subPageState of
                Loaded page ->
                    page

                TransitioningFrom page ->
                    page

        breadcrumb =
            viewBreadcrumb model.project

        sidebar =
            viewSidebar model.project
    in
        case page of
            Blank ->
                Html.text ""
                    |> frame (sidebar OtherPage)
                    => breadcrumb "" (Html.text "")

            Errored subModel ->
                Html.text "Errored"
                    |> frame (sidebar OtherPage)
                    => breadcrumb "" (Html.text "")

            Commits subModel ->
                Commits.view subModel
                    |> frame (sidebar CommitsPage)
                    |> Html.map CommitsMsg
                    => breadcrumb "Commits" (Commits.viewBreadcrumbExtraItems subModel |> Html.map CommitsMsg)

            Settings subModel ->
                Settings.view subModel
                    |> frame (sidebar SettingsPage)
                    |> Html.map SettingsMsg
                    => breadcrumb "Settings" (Html.text "")


frame : Html msg -> Html msg -> Html msg
frame sidebar content =
    div [ class "row" ]
        [ sidebar
        , div [ class "col-sm-9 ml-sm-auto col-md-10 pt-3" ] [ content ]
        ]



-- UPDATE --


type Msg
    = NoOp
    | SetRoute (Maybe ProjectRoute.Route)
    | CommitsMsg Commits.Msg
    | CommitsLoaded (Result PageLoadError Commits.Model)
    | SettingsMsg Settings.Msg


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

            Just (ProjectRoute.Commits) ->
                case session.user of
                    Just user ->
                        transition CommitsLoaded (Commits.init session model.project.id)

                    Nothing ->
                        errored Page.Project "Uhoh"

            Just (ProjectRoute.Settings) ->
                case session.user of
                    Just user ->
                        { model | subPageState = Loaded (Settings (Settings.init model.project)) } => Cmd.none

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
            ( SetRoute route, _ ) ->
                setRoute session route model

            ( CommitsLoaded (Ok subModel), _ ) ->
                { model | subPageState = Loaded (Commits subModel) } => Cmd.none

            ( CommitsLoaded (Err error), _ ) ->
                { model | subPageState = Loaded (Errored error) } => Cmd.none

            ( CommitsMsg subMsg, Commits subModel ) ->
                toPage Commits CommitsMsg (Commits.update model.project session) subMsg subModel

            ( _, _ ) ->
                -- Disregard incoming messages that arrived for the wrong sub page
                (Debug.log "Fell through" model) => Cmd.none
