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
import Views.Helpers exposing (onClickPage)
import Navigation exposing (newUrl)
import Socket.Channel as Channel exposing (Channel)
import Socket.Socket as Socket exposing (Socket)


-- SUB PAGES --


type SubPage
    = Blank
    | Overview
    | Commits Commits.Model
    | Commit Commit.Model
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


channels : Model -> List (Channel Msg)
channels { subPageState } =
    case subPageState of
        Loaded (Commit subModel) ->
            List.map (Channel.map CommitMsg) (Commit.channels subModel)

        TransitioningFrom (Commit subModel) ->
            List.map (Channel.map CommitMsg) (Commit.channels subModel)

        _ ->
            []


init : Session msg -> Project.Id -> Maybe ProjectRoute.Route -> Task PageLoadError ( ( Model, Cmd Msg ), ExternalMsg )
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

        --
        --        newSocket =
        --            List.map (Channel.map CommitMsg) (Commit.channels subModel)
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
                                => NoOp
                                |> Task.succeed
                )
            |> Task.mapError handleLoadError



-- VIEW --


type ActiveSubPage
    = OtherPage
    | OverviewPage
    | CommitsPage
    | SettingsPage


view : Session msg -> Model -> Html Msg
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


viewSubPage : Session msg -> Model -> ( Html Msg, Html Msg )
viewSubPage session model =
    let
        page =
            getSubPage model.subPageState

        project =
            model.project

        branches =
            model.branches

        breadcrumb =
            viewBreadcrumb project

        sidebar =
            viewSidebar project

        pageFrame =
            frame project
    in
        case page of
            Blank ->
                let
                    content =
                        Html.text ""
                            |> pageFrame (sidebar OtherPage)
                in
                    ( content, breadcrumb (text "") [] )

            Errored subModel ->
                let
                    content =
                        Html.text "Errored"
                            |> pageFrame (sidebar OtherPage)
                in
                    ( content, breadcrumb (text "") [] )

            Overview ->
                let
                    content =
                        Overview.view project
                            |> pageFrame (sidebar OverviewPage)
                in
                    ( content, breadcrumb (text "") [] )

            Commits subModel ->
                let
                    content =
                        Commits.view project branches subModel
                            |> Html.map CommitsMsg
                            |> pageFrame (sidebar CommitsPage)

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
                            |> pageFrame (sidebar CommitsPage)

                    crumb =
                        Commit.breadcrumb project subModel.commit subModel.subPageState
                            |> breadcrumb (text "")
                in
                    ( content, crumb )

            Settings subModel ->
                let
                    content =
                        Settings.view project subModel
                            |> Html.map SettingsMsg
                            |> pageFrame (sidebar SettingsPage)

                    crumb =
                        Settings.breadcrumb project
                            |> breadcrumb (text "")
                in
                    ( content, crumb )


frame : Project -> Html msg -> Html msg -> Html msg
frame project sidebar content =
    div [ class "row" ]
        [ sidebar
        , div [ class "col-sm-9 ml-sm-auto col-md-10 pt-3 project-content-container " ]
            [ h1 [ class "display-6" ] [ text project.name ]
            , content
            ]
        ]



-- UPDATE --


type Msg
    = NewUrl String
    | SetRoute (Maybe ProjectRoute.Route)
    | CommitsMsg Commits.Msg
    | CommitsLoaded (Result PageLoadError Commits.Model)
    | CommitMsg Commit.Msg
    | CommitLoaded (Result PageLoadError ( Commit.Model, Cmd Commit.Msg ))
    | SettingsMsg Settings.Msg


type ExternalMsg
    = SetSocket ( Socket Msg, Cmd (Socket.Msg Msg) )
    | NoOp


getSubPage : SubPageState -> SubPage
getSubPage subPageState =
    case subPageState of
        Loaded subPage ->
            subPage

        TransitioningFrom subPage ->
            subPage


setRoute : Session msg -> Maybe ProjectRoute.Route -> Model -> ( Model, Cmd Msg )
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

            Just (ProjectRoute.Commit hash maybeRoute) ->
                let
                    loadFreshPage =
                        Just maybeRoute
                            |> Commit.init session model.project hash
                            |> Task.andThen (Tuple.first >> Task.succeed)
                            |> transition CommitLoaded

                    transitionSubPage subModel =
                        let
                            ( ( newModel, newMsg ), externalMsg ) =
                                subModel
                                    |> Commit.update model.project session (Commit.SetRoute (Just maybeRoute))
                        in
                            { model | subPageState = Loaded (Commit newModel) }
                                => Cmd.map CommitMsg newMsg
                in
                    case ( session.user, model.subPageState ) of
                        ( Just _, Loaded page ) ->
                            case page of
                                -- If we're on the product page for the same product as the new route just load sub-page
                                -- Otherwise load the project page fresh
                                Commit subModel ->
                                    if hash == subModel.commit.hash then
                                        transitionSubPage subModel
                                    else
                                        loadFreshPage

                                _ ->
                                    loadFreshPage

                        ( Just _, TransitioningFrom _ ) ->
                            loadFreshPage

                        ( Nothing, _ ) ->
                            errored Page.Project "Error loading commit"

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


update : Session msg -> Msg -> Model -> ( ( Model, Cmd Msg ), ExternalMsg )
update session msg model =
    updateSubPage session (getSubPage model.subPageState) msg model


updateSubPage : Session msg -> SubPage -> Msg -> Model -> ( ( Model, Cmd Msg ), ExternalMsg )
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
                model
                    => newUrl url
                    => NoOp

            ( SetRoute route, _ ) ->
                setRoute session route model
                    => NoOp

            ( SettingsMsg subMsg, Settings subModel ) ->
                toPage Settings SettingsMsg (Settings.update model.project session) subMsg subModel
                    => NoOp

            ( CommitsLoaded (Ok subModel), _ ) ->
                { model | subPageState = Loaded (Commits subModel) }
                    => Cmd.none
                    => NoOp

            ( CommitsLoaded (Err error), _ ) ->
                { model | subPageState = Loaded (Errored error) }
                    => Cmd.none
                    => NoOp

            ( CommitsMsg subMsg, Commits subModel ) ->
                toPage Commits CommitsMsg (Commits.update model.project session) subMsg subModel
                    => NoOp

            ( CommitLoaded (Ok ( subModel, subMsg )), _ ) ->
                { model | subPageState = Loaded (Commit subModel) }
                    => Cmd.map CommitMsg subMsg
                    => NoOp

            ( CommitLoaded (Err error), _ ) ->
                { model | subPageState = Loaded (Errored error) }
                    => Cmd.none
                    => NoOp

            ( CommitMsg subMsg, Commit subModel ) ->
                let
                    ( ( newSubModel, newCmd ), externalMsg ) =
                        Commit.update model.project session subMsg subModel

                    newExternalMsg =
                        case externalMsg of
                            Commit.NoOp ->
                                NoOp

                            Commit.SetSocket socket ->
                                SetSocket ( (Socket.map CommitMsg socket), Cmd.none )

                    --                                SetSocket ( (Socket.map CommitMsg socket), Cmd.none )
                in
                    { model | subPageState = Loaded (Commit newSubModel) }
                        ! [ Cmd.map CommitMsg newCmd ]
                        => newExternalMsg

            ( _, _ ) ->
                -- Disregard incoming messages that arrived for the wrong sub page
                (Debug.log "Fell through" model)
                    => Cmd.none
                    => NoOp
