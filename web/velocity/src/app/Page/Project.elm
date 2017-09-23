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
import Route
import Page.Project.Route as ProjectRoute
import Page.Project.Commits as Commits
import Views.Page as Page exposing (ActivePage)


type SubPage
    = Blank
    | Commits Commits.Model
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

        loadCommits =
            Commits.init session id
                |> Task.map Commits

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


view : Session -> Model -> Html Msg
view session model =
    let
        project =
            model.project
    in
        div []
            [ viewBreadcrumb project
            , div [ class "container-fluid" ] [ viewSubPage session model ]
            ]


viewSidebar : Project -> ActiveSubPage -> Html msg
viewSidebar project subPage =
    nav [ class "col-sm-3 col-md-2 d-none d-sm-block bg-light sidebar" ]
        [ ul [ class "nav nav-pills flex-column" ]
            [ sidebarLink (subPage == CommitsPage) (ProjectRoute.Commits project.id) [ text "Commits " ]
            ]
        ]


sidebarLink : Bool -> ProjectRoute.Route -> List (Html msg) -> Html msg
sidebarLink isActive route linkContent =
    li [ class "nav-item" ]
        [ a [ class "nav-link", ProjectRoute.href route, classList [ ( "active", isActive ) ] ]
            (linkContent ++ [ Util.viewIf isActive (span [ class "sr-only" ] [ text "(current)" ]) ])
        ]


viewBreadcrumb : Project -> Html Msg
viewBreadcrumb project =
    div [ class "d-flex justify-content-start align-items-center bg-dark breadcrumb-container", style [ ( "height", "50px" ) ] ]
        [ div [ class "p-2" ]
            [ ol [ class "breadcrumb bg-dark", style [ ( "margin", "0" ) ] ]
                [ li [ class "breadcrumb-item" ] [ a [ Route.href Route.Projects ] [ text "Projects" ] ]
                , li [ class "breadcrumb-item active" ] [ text project.name ]
                ]
            ]
        , div [ class "ml-auto p-2" ] []
          --            [ button
          --                [ class "ml-auto btn btn-dark btn-outline-dark", type_ "button", onClick SubmitSync, disabled synchronizing ]
          --                [ i [ class "fa fa-refresh" ] [], text " Refresh " ]
          --            ]
        ]


viewSubPage : Session -> Model -> Html Msg
viewSubPage session model =
    let
        page =
            case model.subPageState of
                Loaded page ->
                    page

                TransitioningFrom page ->
                    page

        sidebar =
            viewSidebar model.project
    in
        case page of
            Blank ->
                Html.text ""
                    |> frame (sidebar OtherPage)

            Errored subModel ->
                Html.text "Errored"
                    |> frame (sidebar OtherPage)

            Commits subModel ->
                Commits.view subModel
                    |> frame (sidebar CommitsPage)
                    |> Html.map CommitsMsg


frame : Html msg -> Html msg -> Html msg
frame sidebar content =
    div [ class "row" ]
        [ sidebar
        , div [ class "col-sm-9 ml-sm-auto col-md-10 pt-3" ] [ content ]
        ]



-- SUBSCRIPTIONS --


getSubPage : SubPageState -> SubPage
getSubPage subPageState =
    case subPageState of
        Loaded subPage ->
            subPage

        TransitioningFrom subPage ->
            subPage



-- UPDATE --


type Msg
    = NoOp
    | SetRoute (Maybe ProjectRoute.Route)
    | CommitsMsg Commits.Msg
    | CommitsLoaded (Result PageLoadError Commits.Model)


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

            Just (ProjectRoute.Commits id) ->
                case session.user of
                    Just user ->
                        transition CommitsLoaded (Commits.init session id)

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
                toPage Commits CommitsMsg (Commits.update session) subMsg subModel

            ( _, _ ) ->
                -- Disregard incoming messages that arrived for the wrong sub page
                (Debug.log "Fell through" model) => Cmd.none
